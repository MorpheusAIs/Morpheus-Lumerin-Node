import { describe, expect, it } from 'vitest'
import {
  activeCoworkMarketplaceSessions,
  loadForStableCoworkWallet
} from './cowork-marketplace-sessions'

const NOW = 2_000_000
const futureSeconds = (offset = 60): number => Math.floor((NOW + offset * 1_000) / 1_000)

describe('activeCoworkMarketplaceSessions', () => {
  it('returns only active marketplace chat sessions', () => {
    const sessions = [
      { Id: 'active', ModelAgentId: 'llm', ClosedAt: '0', EndsAt: futureSeconds() },
      { Id: 'closed', ModelAgentId: 'llm', ClosedAt: 1, EndsAt: futureSeconds() },
      { Id: 'expired', ModelAgentId: 'llm', EndsAt: Math.floor((NOW - 1) / 1_000) },
      { Id: 'deleted-model', ModelAgentId: 'deleted', EndsAt: futureSeconds() },
      { Id: 'audio', ModelAgentId: 'audio', EndsAt: futureSeconds() },
      { Id: 'tagged-audio', ModelAgentId: 'tagged-audio', EndsAt: futureSeconds() },
      { Id: 'missing-model', ModelAgentId: 'missing', EndsAt: futureSeconds() }
    ]
    const models = [
      { Id: 'llm', Name: 'Marketplace LLM', ModelType: 'llm' },
      { Id: 'deleted', Name: 'Deleted', ModelType: 'llm', IsDeleted: true },
      { Id: 'audio', Name: 'Speech', ModelType: 'tts', Tags: ['tts'] },
      { Id: 'tagged-audio', Name: 'Speech tags only', Tags: ['tts'] }
    ]

    expect(activeCoworkMarketplaceSessions(sessions, models, NOW)).toEqual([
      {
        session: sessions[0],
        model: models[0],
        endsAt: futureSeconds() * 1_000
      }
    ])
  })

  it('accepts models whose metadata explicitly declares chat capability', () => {
    const session = { Id: 'session-1', ModelAgentId: 'model-1', EndsAt: futureSeconds() }
    const model = { Id: 'model-1', Name: 'Agent', ModelType: 'UNKNOWN', Tags: ['chat'] }

    expect(activeCoworkMarketplaceSessions([session], [model], NOW)).toEqual([
      { session, model, endsAt: futureSeconds() * 1_000 }
    ])
  })

  it.each(['tts', 'STT', 'embedding', 'embeddings', 'agent'])(
    'rejects explicit %s models even when free-form tags claim chat',
    (modelType) => {
      const session = { Id: 'session-1', ModelAgentId: 'model-1', EndsAt: futureSeconds() }
      const model = {
        Id: 'model-1',
        Name: 'Conflicting metadata',
        ModelType: modelType,
        Tags: ['chat', 'llm']
      }

      expect(activeCoworkMarketplaceSessions([session], [model], NOW)).toEqual([])
    }
  )

  it('accepts an explicit LLM despite conflicting free-form tags', () => {
    const session = { Id: 'session-1', ModelAgentId: 'model-1', EndsAt: futureSeconds() }
    const model = { Id: 'model-1', Name: 'LLM', ModelType: ' LLM ', Tags: ['tts'] }

    expect(activeCoworkMarketplaceSessions([session], [model], NOW)).toEqual([
      { session, model, endsAt: futureSeconds() * 1_000 }
    ])
  })

  it('keeps only sessions owned by the expected active wallet when ownership is present', () => {
    const sessions = [
      {
        Id: 'owned',
        User: ' 0xAbC ',
        ModelAgentId: 'model-1',
        EndsAt: futureSeconds()
      },
      {
        Id: 'other-wallet',
        User: '0xdef',
        ModelAgentId: 'model-1',
        EndsAt: futureSeconds()
      },
      {
        Id: 'missing-owner-field',
        ModelAgentId: 'model-1',
        EndsAt: futureSeconds()
      },
      {
        Id: 'invalid-owner-field',
        User: null,
        ModelAgentId: 'model-1',
        EndsAt: futureSeconds()
      }
    ]
    const model = { Id: 'model-1', Name: 'LLM', ModelType: 'llm' }

    expect(activeCoworkMarketplaceSessions(sessions, [model], NOW, ' 0xABC ')).toEqual([
      { session: sessions[0], model, endsAt: futureSeconds() * 1_000 },
      { session: sessions[2], model, endsAt: futureSeconds() * 1_000 }
    ])
  })

  it('does not manufacture options from local models without a funded session', () => {
    const localModel = { Id: 'local', Name: 'Local model', isLocal: true, ModelType: 'llm' }

    expect(activeCoworkMarketplaceSessions([], [localModel], NOW)).toEqual([])
  })

  it('rejects an async result when the active wallet changes before it can be returned', async () => {
    const loaded = { models: ['wallet-a-session'] }

    await expect(
      loadForStableCoworkWallet(
        ' 0xAaA ',
        async () => loaded,
        async () => '0xbbb'
      )
    ).rejects.toThrow('active wallet changed')
  })

  it('returns an async result when the active wallet remains the same after normalization', async () => {
    const loaded = { models: ['wallet-a-session'] }

    await expect(
      loadForStableCoworkWallet(
        ' 0xAaA ',
        async () => loaded,
        async () => '0xaaa'
      )
    ).resolves.toBe(loaded)
  })
})
