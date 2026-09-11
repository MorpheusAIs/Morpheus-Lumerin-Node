export type ExistingSessionConflict = {
  existingSessionID: string
}

/**
 * Extract only the one conflict shape the desktop can safely recover from.
 * Every other proxy error must keep propagating through the normal error path.
 */
export function parseExistingSessionConflict(error: unknown): ExistingSessionConflict | null {
  if (!error || typeof error !== 'object') return null

  const candidate = error as { status?: unknown; responseBody?: unknown }
  if (
    candidate.status !== 409 ||
    !candidate.responseBody ||
    typeof candidate.responseBody !== 'object'
  ) {
    return null
  }

  const existingSessionID = (candidate.responseBody as Record<string, unknown>).existingSessionID
  if (typeof existingSessionID !== 'string' || !/^0x[0-9a-fA-F]{64}$/u.test(existingSessionID)) {
    return null
  }

  return { existingSessionID }
}
