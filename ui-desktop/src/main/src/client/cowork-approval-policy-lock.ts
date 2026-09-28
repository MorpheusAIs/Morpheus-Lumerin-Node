type Waiter = {
  kind: 'read' | 'write'
  enter: () => void
}

const queue: Waiter[] = []
let activeReaders = 0
let activeWriter = false

function drain(): void {
  if (activeWriter || queue.length === 0) return
  if (queue[0].kind === 'write') {
    if (activeReaders > 0) return
    activeWriter = true
    queue.shift()!.enter()
    return
  }

  // Admit the contiguous reader group ahead of the first queued writer. New
  // readers queue behind that writer, so frequent file work cannot starve a
  // policy change.
  while (queue[0]?.kind === 'read' && !activeWriter) {
    activeReaders += 1
    queue.shift()!.enter()
  }
}

function acquire(kind: Waiter['kind']): Promise<() => void> {
  return new Promise((resolve) => {
    queue.push({
      kind,
      enter: () => {
        let released = false
        resolve(() => {
          if (released) return
          released = true
          if (kind === 'read') activeReaders -= 1
          else activeWriter = false
          drain()
        })
      }
    })
    drain()
  })
}

async function withLock<T>(kind: Waiter['kind'], operation: () => Promise<T>): Promise<T> {
  const release = await acquire(kind)
  try {
    return await operation()
  } finally {
    release()
  }
}

/** Keeps one policy snapshot stable from file-action classification through execution. */
export const withCoworkApprovalPolicyReadLock = <T>(operation: () => Promise<T>): Promise<T> =>
  withLock('read', operation)

/** Waits for old-policy file actions before publishing a new global policy. */
export const withCoworkApprovalPolicyWriteLock = <T>(operation: () => Promise<T>): Promise<T> =>
  withLock('write', operation)
