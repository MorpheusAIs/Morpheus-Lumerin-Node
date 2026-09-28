import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  coworkSessionBusy,
  coworkSessionKey,
  coworkSessionWaiting,
  resetCoworkSessionQueue,
  withCoworkSessionTurn,
  withIdleCoworkSession
} from './cowork-session-queue'

afterEach(() => resetCoworkSessionQueue())

/** A promise plus the handle that settles it, so a test can hold a turn open. */
function deferred<T = void>() {
  let resolve!: (value: T) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

const tick = () => new Promise((resolve) => setTimeout(resolve, 0))

describe('session key', () => {
  it('keys a marketplace model on its session', () => {
    expect(coworkSessionKey({ modelId: 'm', sessionId: '0xabc' })).toBe('session:0xabc')
  })

  it('gives two sessions on one model separate keys', () => {
    // Two projects can hold separate sessions against the same model, and the
    // provider serialises per session rather than per model.
    expect(coworkSessionKey({ modelId: 'm', sessionId: 'a' })).not.toBe(
      coworkSessionKey({ modelId: 'm', sessionId: 'b' })
    )
  })

  it('does not key a local model', () => {
    // A local model is served off the machine with no session and no provider
    // semaphore, so queueing it would only throw away concurrency.
    expect(coworkSessionKey({ isLocal: true, modelId: 'm', sessionId: 'a' })).toBeNull()
  })

  it('does not key a model with no session yet', () => {
    expect(coworkSessionKey({ modelId: 'm' })).toBeNull()
  })
})

describe('taking a session turn', () => {
  it('runs an uncontended turn without waiting', async () => {
    await expect(withCoworkSessionTurn('session:a', async () => 'done')).resolves.toBe('done')
  })

  it('never overlaps two turns on one session', async () => {
    const first = deferred()
    const order: string[] = []

    const a = withCoworkSessionTurn('session:a', async () => {
      order.push('a:start')
      await first.promise
      order.push('a:end')
    })
    const b = withCoworkSessionTurn('session:a', async () => {
      order.push('b:start')
    })

    await tick()
    expect(order).toEqual(['a:start'])
    first.resolve()
    await Promise.all([a, b])
    expect(order).toEqual(['a:start', 'a:end', 'b:start'])
  })

  it('lets two different sessions run at once', async () => {
    const held = deferred()
    const started: string[] = []

    const a = withCoworkSessionTurn('session:a', async () => {
      started.push('a')
      await held.promise
    })
    const b = withCoworkSessionTurn('session:b', async () => {
      started.push('b')
    })

    await tick()
    // This is the whole point of keying on the session rather than a global
    // lock: unrelated work must not be slowed down by it.
    expect(started).toEqual(['a', 'b'])
    held.resolve()
    await Promise.all([a, b])
  })

  it('runs a null key straight through with no serialisation', async () => {
    const held = deferred()
    const started: string[] = []

    const a = withCoworkSessionTurn(null, async () => {
      started.push('a')
      await held.promise
    })
    const b = withCoworkSessionTurn(null, async () => {
      started.push('b')
    })

    await tick()
    expect(started).toEqual(['a', 'b'])
    held.resolve()
    await Promise.all([a, b])
  })

  it('preserves the order callers arrived in', async () => {
    const held = deferred()
    const order: number[] = []
    const runs = [
      withCoworkSessionTurn('session:a', async () => {
        await held.promise
        order.push(0)
      }),
      withCoworkSessionTurn('session:a', async () => {
        order.push(1)
      }),
      withCoworkSessionTurn('session:a', async () => {
        order.push(2)
      })
    ]

    held.resolve()
    await Promise.all(runs)
    expect(order).toEqual([0, 1, 2])
  })

  it('tells a caller it is waiting, and only when it is', async () => {
    const held = deferred()
    const firstWait = vi.fn()
    const secondWait = vi.fn()

    const a = withCoworkSessionTurn('session:a', () => held.promise, firstWait)
    await tick()
    const b = withCoworkSessionTurn('session:a', async () => undefined, secondWait)
    await tick()

    // The turn that went straight through has nothing to report. Saying
    // "waiting for the model" on every turn would make the word meaningless.
    expect(firstWait).not.toHaveBeenCalled()
    expect(secondWait).toHaveBeenCalledTimes(1)

    held.resolve()
    await Promise.all([a, b])
  })

  it('reports who is busy and how many are queued', async () => {
    const held = deferred()
    expect(coworkSessionBusy('session:a')).toBe(false)

    const a = withCoworkSessionTurn('session:a', () => held.promise)
    await tick()
    expect(coworkSessionBusy('session:a')).toBe(true)
    expect(coworkSessionWaiting('session:a')).toBe(0)

    const b = withCoworkSessionTurn('session:a', async () => undefined)
    await tick()
    expect(coworkSessionWaiting('session:a')).toBe(1)

    held.resolve()
    await Promise.all([a, b])
    expect(coworkSessionBusy('session:a')).toBe(false)
    expect(coworkSessionWaiting('session:a')).toBe(0)
  })

  it('reports a null key as idle', () => {
    expect(coworkSessionBusy(null)).toBe(false)
    expect(coworkSessionWaiting(null)).toBe(0)
  })

  it('releases the session when a turn throws', async () => {
    // A failed request must not wedge the session. This is the case that would
    // turn one bad turn into a project that never runs again.
    await expect(
      withCoworkSessionTurn('session:a', async () => {
        throw new Error('upstream refused')
      })
    ).rejects.toThrow('upstream refused')

    expect(coworkSessionBusy('session:a')).toBe(false)
    await expect(withCoworkSessionTurn('session:a', async () => 'next')).resolves.toBe('next')
  })

  it('runs the next turn even though the one before it failed', async () => {
    const failing = withCoworkSessionTurn('session:a', async () => {
      await tick()
      throw new Error('upstream refused')
    })
    const following = withCoworkSessionTurn('session:a', async () => 'ran')

    await expect(failing).rejects.toThrow('upstream refused')
    await expect(following).resolves.toBe('ran')
  })
})

describe('taking a session turn only when it is free', () => {
  it('runs when nothing holds the session', async () => {
    await expect(withIdleCoworkSession('session:a', async () => 7)).resolves.toEqual({
      ran: true,
      result: 7
    })
  })

  it('skips rather than queues when a turn is in flight', async () => {
    const held = deferred()
    const a = withCoworkSessionTurn('session:a', () => held.promise)
    await tick()

    // The vision probe is the caller here. Queueing it would put a whole turn
    // behind a diagnostic, which is worse than not running the diagnostic.
    const probe = vi.fn()
    await expect(withIdleCoworkSession('session:a', probe)).resolves.toEqual({ ran: false })
    expect(probe).not.toHaveBeenCalled()

    held.resolve()
    await a
  })

  it('skips when someone is already queued', async () => {
    const held = deferred()
    const a = withCoworkSessionTurn('session:a', () => held.promise)
    await tick()
    const b = withCoworkSessionTurn('session:a', async () => undefined)
    await tick()

    await expect(withIdleCoworkSession('session:a', async () => 1)).resolves.toEqual({
      ran: false
    })

    held.resolve()
    await Promise.all([a, b])
  })

  it('holds the session for as long as it runs', async () => {
    const held = deferred()
    const idle = withIdleCoworkSession('session:a', () => held.promise)
    await tick()
    // It is not fire and forget any more. A real turn arriving mid-probe waits
    // rather than racing it, which is what used to cancel the provider.
    expect(coworkSessionBusy('session:a')).toBe(true)
    held.resolve()
    await idle
  })

  it('runs a null key with nothing to check', async () => {
    await expect(withIdleCoworkSession(null, async () => 'local')).resolves.toEqual({
      ran: true,
      result: 'local'
    })
  })
})
