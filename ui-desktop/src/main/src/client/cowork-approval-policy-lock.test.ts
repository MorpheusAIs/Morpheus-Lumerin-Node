import { describe, expect, it, vi } from 'vitest'
import {
  withCoworkApprovalPolicyReadLock,
  withCoworkApprovalPolicyWriteLock
} from './cowork-approval-policy-lock'

describe('Workspace approval policy transition lock', () => {
  it('lets readers overlap while a writer waits for the old policy boundary', async () => {
    const order: string[] = []
    let releaseReaders!: () => void
    const readersBlocked = new Promise<void>((resolve) => {
      releaseReaders = resolve
    })
    const first = withCoworkApprovalPolicyReadLock(async () => {
      order.push('read-1-start')
      await readersBlocked
      order.push('read-1-end')
    })
    const second = withCoworkApprovalPolicyReadLock(async () => {
      order.push('read-2-start')
      await readersBlocked
      order.push('read-2-end')
    })
    await vi.waitFor(() => expect(order).toEqual(['read-1-start', 'read-2-start']))

    const writer = withCoworkApprovalPolicyWriteLock(async () => {
      order.push('write')
    })
    const lateReader = withCoworkApprovalPolicyReadLock(async () => {
      order.push('late-read')
    })
    await Promise.resolve()
    expect(order).toEqual(['read-1-start', 'read-2-start'])

    releaseReaders()
    await Promise.all([first, second, writer, lateReader])

    expect(order.slice(0, 2)).toEqual(['read-1-start', 'read-2-start'])
    expect(order.indexOf('write')).toBeGreaterThan(order.indexOf('read-1-end'))
    expect(order.indexOf('write')).toBeGreaterThan(order.indexOf('read-2-end'))
    expect(order.indexOf('late-read')).toBeGreaterThan(order.indexOf('write'))
  })
})
