export interface ActiveCoworkMarketplaceSession {
  session: Record<string, any>
  model: Record<string, any>
  endsAt: number
}

export function normalizeCoworkWalletAddress(value: unknown): string {
  return typeof value === 'string' ? value.trim().toLowerCase() : ''
}

export async function loadForStableCoworkWallet<T>(
  expectedWalletAddress: string,
  load: () => Promise<T>,
  readActiveWalletAddress: () => Promise<unknown>
): Promise<T> {
  const expectedWallet = normalizeCoworkWalletAddress(expectedWalletAddress)
  const value = await load()
  const currentWallet = normalizeCoworkWalletAddress(await readActiveWalletAddress())
  if (!expectedWallet || currentWallet !== expectedWallet) {
    throw new Error(
      'The active wallet changed while Workspace sessions were loading. Refresh and try again.'
    )
  }
  return value
}

function isChatModel(model: Record<string, any>): boolean {
  if (model.IsDeleted) return false
  const type = String(model.ModelType ?? '')
    .trim()
    .toLowerCase()
  const tags = Array.isArray(model.Tags)
    ? model.Tags.map((tag: unknown) => String(tag).toLowerCase().trim())
    : []
  // Canonical model metadata is authoritative. In particular, an explicit
  // speech or embedding type must not become Cowork-eligible because of a
  // stale/conflicting free-form tag.
  if (type === 'llm') return true
  if (type && type !== 'unknown') return false
  if (tags.includes('llm') || tags.includes('chat')) return true
  const nonChatTags = new Set([
    'tts',
    'text-to-speech',
    'text2speech',
    't2s',
    'stt',
    'transcribe',
    's2t',
    'speech',
    'speech-to-text',
    'speech2text',
    'embedding',
    'embeddings'
  ])
  return !tags.some((tag: string) => nonChatTags.has(tag))
}

/**
 * Resolves only currently-open, on-chain marketplace sessions to Cowork models.
 * Local/configured endpoints deliberately never enter this function: Cowork
 * requires the user to choose a marketplace model and open a funded session
 * before project data or tools can be used.
 */
export function activeCoworkMarketplaceSessions(
  sessions: unknown,
  models: unknown,
  now = Date.now(),
  expectedWalletAddress?: string
): ActiveCoworkMarketplaceSession[] {
  if (!Array.isArray(sessions) || !Array.isArray(models)) return []
  const expectedWallet = normalizeCoworkWalletAddress(expectedWalletAddress)

  return sessions.flatMap((candidate): ActiveCoworkMarketplaceSession[] => {
    if (!candidate || typeof candidate !== 'object' || Array.isArray(candidate)) return []
    const session = candidate as Record<string, any>
    if (
      expectedWallet &&
      session.User !== undefined &&
      normalizeCoworkWalletAddress(session.User) !== expectedWallet
    ) {
      return []
    }
    const sessionId = String(session.Id ?? '').trim()
    const modelId = String(session.ModelAgentId ?? '').trim()
    const endsAt = Number(session.EndsAt) * 1000
    const closedAt = Number(session.ClosedAt ?? 0)
    const isClosed = Number.isFinite(closedAt) ? closedAt > 0 : Boolean(session.ClosedAt)
    if (!sessionId || !modelId || isClosed || !Number.isFinite(endsAt) || endsAt <= now) {
      return []
    }

    const model = models.find(
      (value) =>
        value &&
        typeof value === 'object' &&
        !Array.isArray(value) &&
        String((value as Record<string, any>).Id) === modelId
    ) as Record<string, any> | undefined
    if (!model || !isChatModel(model)) return []
    return [{ session, model, endsAt }]
  })
}
