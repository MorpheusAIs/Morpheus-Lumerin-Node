import { afterEach, describe, expect, it, vi } from 'vitest'

const runtime = vi.hoisted(() => ({ dev: true }))

vi.mock('@electron-toolkit/utils', () => ({
  is: {
    get dev() {
      return runtime.dev
    }
  }
}))

import {
  assertTrustedRendererEvent,
  getTrustedRendererUrl,
  isTrustedRendererEvent,
  isTrustedRendererUrl
} from './rendererTrust'

const rendererEvent = (frameUrl: string, parent: unknown = null, senderUrl = frameUrl) =>
  ({
    senderFrame: { url: frameUrl, parent },
    sender: { getURL: vi.fn(() => senderUrl) }
  }) as any

describe.sequential('renderer IPC trust boundary', () => {
  afterEach(() => {
    runtime.dev = true
    vi.unstubAllEnvs()
  })

  it('accepts only the configured development origin', () => {
    vi.stubEnv('ELECTRON_RENDERER_URL', 'http://127.0.0.1:5173/app/index.html')

    expect(isTrustedRendererUrl('http://127.0.0.1:5173/another/route?view=settings')).toBe(true)
    expect(isTrustedRendererUrl('https://127.0.0.1:5173/app/index.html')).toBe(false)
    expect(isTrustedRendererUrl('http://127.0.0.1.evil.example:5173/app/index.html')).toBe(false)
    expect(isTrustedRendererUrl('http://127.0.0.1:5174/app/index.html')).toBe(false)
    expect(isTrustedRendererUrl('not a URL')).toBe(false)
  })

  it('requires the exact packaged renderer file while allowing harmless query state', () => {
    runtime.dev = false
    const trusted = getTrustedRendererUrl()
    const withQuery = new URL(trusted)
    withQuery.searchParams.set('view', 'settings')
    const sibling = new URL('other.html', trusted)

    expect(isTrustedRendererUrl(withQuery.href)).toBe(true)
    expect(isTrustedRendererUrl(sibling.href)).toBe(false)
    expect(isTrustedRendererUrl('https://example.com/index.html')).toBe(false)
  })

  it('accepts the trusted main frame and rejects same-origin subframes', () => {
    vi.stubEnv('ELECTRON_RENDERER_URL', 'http://localhost:5173')
    const mainFrame = rendererEvent('http://localhost:5173/settings')
    const subframe = rendererEvent('http://localhost:5173/settings', { routingId: 2 })

    expect(isTrustedRendererEvent(mainFrame)).toBe(true)
    expect(isTrustedRendererEvent(subframe)).toBe(false)
    expect(() => assertTrustedRendererEvent(mainFrame)).not.toThrow()
    expect(() => assertTrustedRendererEvent(subframe)).toThrow(
      'Rejected IPC from an untrusted renderer.'
    )
  })

  it('uses the sender URL only when no sender-frame URL is available', () => {
    vi.stubEnv('ELECTRON_RENDERER_URL', 'http://localhost:5173')
    const trustedFallback = rendererEvent('', null, 'http://localhost:5173/settings')
    const untrustedFallback = rendererEvent('', null, 'https://attacker.example/settings')

    expect(isTrustedRendererEvent(trustedFallback)).toBe(true)
    expect(isTrustedRendererEvent(untrustedFallback)).toBe(false)
  })
})
