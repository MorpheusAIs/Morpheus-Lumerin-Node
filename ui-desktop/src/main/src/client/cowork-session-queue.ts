/**
 * One model request at a time per marketplace session.
 *
 * The provider holds a semaphore keyed by session id and will only process one
 * prompt per session at a time. Anything else that arrives on that session sits
 * in the provider's queue with nothing to show for it, which is what two
 * Workspace projects running against the same model looked like: the second one
 * appeared to stop doing anything, because it had.
 *
 * Worse, a queued request that gives up takes the provider's in-flight call down
 * with it. Aborting the HTTP request closes the connection, the provider's
 * request context cancels, and the call it had already made to the model dies as
 * `context canceled`, which surfaces to the user as the model having failed. The
 * background vision probe was the usual culprit, since it fires on the same
 * session as the turn it was launched from and carries its own abort timeout.
 *
 * Queueing here rather than at the provider costs nothing that was not already
 * being paid, and it is the difference between waiting and hanging: the caller
 * knows it is waiting, so the UI can say so, and no request is ever cancelled
 * out from under a provider that is mid-answer.
 */

type SessionSlot = {
  /** Resolves when the request currently holding the session has finished. */
  tail: Promise<void>
  /** Callers waiting behind the one in flight. */
  waiting: number
  busy: boolean
}

const slots = new Map<string, SessionSlot>()

/**
 * The key a request contends on, or null when nothing contends.
 *
 * A marketplace session is the unit the provider serialises against. A local
 * model is served straight off the machine with no session and no semaphore, so
 * queueing it would only throw away concurrency the user already has.
 */
export function coworkSessionKey(model: {
  isLocal?: boolean
  modelId: string
  sessionId?: string
}): string | null {
  if (model.isLocal || !model.sessionId) return null
  return `session:${model.sessionId}`
}

/** Whether a request is in flight on this session right now. */
export function coworkSessionBusy(key: string | null): boolean {
  return key !== null && slots.get(key)?.busy === true
}

/** How many callers are queued behind the one in flight. */
export function coworkSessionWaiting(key: string | null): number {
  return key === null ? 0 : (slots.get(key)?.waiting ?? 0)
}

function slotFor(key: string): SessionSlot {
  const existing = slots.get(key)
  if (existing) return existing
  const created: SessionSlot = { tail: Promise.resolve(), waiting: 0, busy: false }
  slots.set(key, created)
  return created
}

function forget(key: string, slot: SessionSlot): void {
  // Only drop the slot when nothing is using it, so a later caller cannot
  // create a second one and run alongside the first.
  if (!slot.busy && slot.waiting === 0 && slots.get(key) === slot) slots.delete(key)
}

/**
 * Runs `operation` once the session is free. A local model never contends, but
 * takes the same path so there is one ordering rule rather than two.
 *
 * `onWait` is called only when the caller actually has to queue, so a run can
 * say it is waiting for the model without saying so on every uncontended turn.
 */
export async function withCoworkSessionTurn<T>(
  key: string | null,
  operation: () => Promise<T>,
  onWait?: () => void
): Promise<T> {
  if (key === null) return operation()
  const slot = slotFor(key)
  const previous = slot.tail
  let release!: () => void
  const gate = new Promise<void>((resolve) => {
    release = resolve
  })
  slot.tail = previous.then(() => gate)

  const contended = slot.busy || slot.waiting > 0
  slot.waiting += 1
  try {
    if (contended) onWait?.()
    await previous
  } finally {
    slot.waiting -= 1
  }

  slot.busy = true
  try {
    return await operation()
  } finally {
    slot.busy = false
    release()
    forget(key, slot)
  }
}

/**
 * Runs `operation` only if the session is idle, and reports whether it ran.
 *
 * For work that is worth doing but never worth delaying a real turn for. The
 * vision probe is the case: skipping it costs one run the verified answer, where
 * queueing it would put a whole turn behind a diagnostic.
 */
export async function withIdleCoworkSession<T>(
  key: string | null,
  operation: () => Promise<T>
): Promise<{ ran: false } | { ran: true; result: T }> {
  // A null key means nothing contends, so there is no idleness to wait for.
  if (key === null) return { ran: true, result: await operation() }
  const slot = slotFor(key)
  if (slot.busy || slot.waiting > 0) return { ran: false }
  const result = await withCoworkSessionTurn(key, operation)
  return { ran: true, result }
}

/** Test seam. Never call this with work outstanding. */
export function resetCoworkSessionQueue(): void {
  slots.clear()
}
