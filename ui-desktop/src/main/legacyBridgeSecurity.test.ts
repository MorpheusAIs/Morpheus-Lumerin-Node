import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'

const source = (relativePath: string): string =>
  readFileSync(path.join(process.cwd(), relativePath), 'utf8')

describe('legacy renderer bridge security boundaries', () => {
  it('does not auto-open DevTools and disables them in packaged builds', () => {
    const main = source('src/main/index.ts')

    expect(main).not.toContain('.openDevTools(')
    expect(main).toContain('devTools: is.dev')
  })

  it('sandboxes the renderer and blocks untrusted navigation and permissions', () => {
    const main = source('src/main/index.ts')

    expect(main).toContain('contextIsolation: true')
    expect(main).toContain('nodeIntegration: false')
    expect(main).toContain('webviewTag: false')
    expect(main).toContain('sandbox: true')
    expect(main).toContain("webContents.on('will-navigate', guardNavigation)")
    expect(main).toContain("webContents.on('will-redirect', guardNavigation)")
    expect(main).toContain('isTrustedRendererUrl(requestingUrl)')
    expect(main).toContain("permission === 'media'")
    expect(main).toContain('isTrustedRendererEvent(event)')
    expect(main).toContain('ipcMain.handle(openExternalChannel')
  })

  it('exposes only fixed legacy channels and keeps Electron events behind the bridge', () => {
    const preload = source('src/preload/index.ts')
    const rendererClient = source('src/renderer/src/client/index.ts')
    const forwarded = [
      ...rendererClient.matchAll(/forwardToMainProcess\(\s*['"]([^'"]+)['"]/gu)
    ].map((match) => match[1])

    expect(forwarded.length).toBeGreaterThan(30)
    for (const channel of forwarded) {
      expect(preload, `${channel} is missing from the preload allowlist`).toContain(`'${channel}'`)
    }
    expect(preload).toContain('legacyIpcChannels.has(eventName)')
    expect(preload).toContain('listener(payload, unsubscribe)')
    expect(preload).not.toContain('listener(event, payload')
    expect(preload).not.toContain('listener(_event, payload')
    expect(preload).not.toContain("from '@electron-toolkit/preload'")
    expect(preload).not.toContain('shell.openExternal')
    expect(preload).not.toContain("exposeInMainWorld('electron'")
    expect(preload).not.toContain("exposeInMainWorld('api'")
  })

  it('validates trusted renderer envelopes and redacts secrets from IPC logs', () => {
    const bootstrap = source('src/main/src/client/index.js')
    const subscriptions = source('src/main/src/client/subscriptions/utils.js')

    expect(bootstrap).toContain('isTrustedRendererEvent(webContent)')
    expect(bootstrap).toContain('isTrustedRendererEvent(event)')
    expect(subscriptions).toContain('isTrustedRendererEvent(event)')
    expect(subscriptions).toContain("typeof evProps !== 'object'")
    expect(subscriptions).toContain("typeof id !== 'string'")
    expect(subscriptions).toMatch(/private\.\?key\|mnemonic\|seed\|authorization/u)
  })

  it('uses only the configured loopback proxy for wallet onboarding', () => {
    const handlers = source('src/main/src/client/subscriptions/handlers.ts')
    const start = handlers.indexOf('export const onboardingCompleted')
    const end = handlers.indexOf('\nexport const ', start + 1)
    const onboarding = handlers.slice(start, end < 0 ? handlers.length : end)

    expect(handlers).toContain('export function configuredLoopbackProxyUrl()')
    expect(onboarding).toContain('configuredLoopbackProxyUrl()')
    expect(onboarding).not.toContain('data.proxyUrl')
    expect(onboarding).not.toMatch(/\{\s*proxyUrl\s*\}\s*=\s*data/u)
  })

  it('validates and natively confirms sensitive renderer-triggered mutations', () => {
    const handlers = source('src/main/src/client/subscriptions/handlers.ts')

    for (const name of [
      'removeWallet',
      'resetWallet',
      'confirmDeclineAgentUser',
      'removeAgentUser',
      'revokeAgentAllowance',
      'confirmDeclineAgentAllowanceRequest'
    ]) {
      const start = handlers.indexOf(`export const ${name}`)
      const end = handlers.indexOf('\nexport const ', start + 1)
      expect(start, `${name} handler is missing`).toBeGreaterThan(-1)
      expect(handlers.slice(start, end < 0 ? handlers.length : end)).toContain(
        'confirmNativeAction'
      )
    }

    expect(handlers).toContain('validateAgentUsername')
    expect(handlers).toContain('validateAgentToken')
    expect(handlers).toContain('validateAgentDecision')
    expect(handlers).toContain('mutateAgentAccess')
    expect(handlers).toContain('if (!/^0x[0-9a-fA-F]{40}$/.test(to))')
    expect(handlers).toContain('if (!/^[0-9]{1,78}$/.test(amount))')
  })

  it('injects a production CSP while retaining the exact loopback API connection', () => {
    const vite = source('electron.vite.config.ts')
    const html = source('src/renderer/index.html')

    expect(vite).toContain("default-src 'self'")
    expect(vite).toContain("object-src 'none'")
    expect(vite).toContain("frame-src 'none'")
    expect(vite).toContain('http://localhost:${proxyPort}')
    expect(vite).toContain("command === 'serve'")
    expect(html).not.toContain('default-src *')
  })

  it('does not embed a wallet key or expose the admin API beyond loopback in run-user', () => {
    const makefile = source('../proxy-router/Makefile')
    const runUser = makefile.slice(makefile.indexOf('run-user:'), makefile.indexOf('\nrun-race:'))

    expect(runUser).not.toMatch(/WALLET_PRIVATE_KEY=0x[0-9a-fA-F]{64}/u)
    expect(runUser).toContain('WALLET_PRIVATE_KEY="$$WALLET_PRIVATE_KEY"')
    expect(runUser).toContain("WEB_ADDRESS='127.0.0.1:8083'")
  })
})
