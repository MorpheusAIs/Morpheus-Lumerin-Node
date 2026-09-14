export type InferenceTargetPayload = {
  modelId?: string
  sessionId?: string
  chatId?: string
}

function boundedProxyId(value: unknown, label: string): string {
  const result = String(value ?? '').trim()
  if (!result || result.length > 256 || /[\u0000-\u001f\u007f]/u.test(result)) {
    throw new Error(`${label} is invalid.`)
  }
  return result
}

/** Builds proxy-router inference headers without permitting the local demo path. */
export function sessionInferenceHeaders(target: InferenceTargetPayload): Record<string, string> {
  if (target?.modelId) {
    throw new Error(
      'An active Morpheus marketplace session is required. Local model inference is disabled.'
    )
  }
  const sessionId = boundedProxyId(target?.sessionId, 'Session ID')
  return {
    session_id: sessionId,
    ...(target?.chatId ? { chat_id: boundedProxyId(target.chatId, 'Chat ID') } : {})
  }
}
