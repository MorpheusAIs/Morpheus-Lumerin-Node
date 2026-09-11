import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { Arch } from 'electron-builder'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  prepareProxyRouterBundle,
  resolveProxyRouterBundleTarget,
  verifyPackagedProxyRouterBundle,
  type BundleEnvironment,
  type BundleCommandRunner
} from './proxy-router-bundle'

let suiteDirectory: string
let projectDirectory: string

const createContext = (platform = 'darwin', arch = Arch.arm64, appOutDir?: string) =>
  ({
    electronPlatformName: platform,
    arch,
    appOutDir: appOutDir ?? path.join(suiteDirectory, 'out'),
    packager: {
      info: { projectDir: projectDirectory },
      appInfo: { productFilename: 'MorpheusUI' }
    }
  }) as any

beforeEach(async () => {
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-router-bundle-'))
  projectDirectory = path.join(suiteDirectory, 'ui-desktop')
  await fs.mkdir(projectDirectory, { recursive: true })
})

afterEach(async () => {
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

describe('proxy-router package bundle', () => {
  it.each([
    ['darwin', 'x64', 'darwin', 'amd64', 'SERVICE_PROXY_DOWNLOAD_URL_MAC_X64'],
    ['darwin', 'arm64', 'darwin', 'arm64', 'SERVICE_PROXY_DOWNLOAD_URL_MAC_ARM64'],
    ['linux', 'x64', 'linux', 'amd64', 'SERVICE_PROXY_DOWNLOAD_URL_LINUX_X64'],
    ['linux', 'arm64', 'linux', 'arm64', 'SERVICE_PROXY_DOWNLOAD_URL_LINUX_ARM64'],
    ['win32', 'x64', 'windows', 'amd64', 'SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_X64']
  ])('maps %s/%s to the matching Go and release targets', (platform, arch, goos, goarch, env) => {
    const target = resolveProxyRouterBundleTarget(platform, arch)
    expect(target).toMatchObject({
      goos,
      goarch,
      downloadUrlEnvironmentVariable: env
    })
  })

  it('builds current source with production tags and writes a verifiable manifest', async () => {
    const invocations: Array<{ executable: string; args: string[]; env?: BundleEnvironment }> = []
    const commit = '1'.repeat(40)
    const runCommand: BundleCommandRunner = vi.fn(async (executable, args, options) => {
      invocations.push({ executable, args, env: options.env })
      if (executable === 'git' && args.includes('--short=12'))
        return { stdout: '123456789abc\n', stderr: '' }
      if (executable === 'git' && args.includes('rev-parse'))
        return { stdout: `${commit}\n`, stderr: '' }
      if (executable === 'git') return { stdout: ' M proxy-router/file.go\n', stderr: '' }

      const outputIndex = args.indexOf('-o') + 1
      await fs.writeFile(args[outputIndex], 'target-specific-router')
      return { stdout: '', stderr: '' }
    })
    const context = createContext()

    await prepareProxyRouterBundle(context, { runCommand, environment: {} })

    const goCall = invocations.find((call) => call.executable === 'go')!
    expect(goCall.args).toEqual(
      expect.arrayContaining(['-trimpath', '-tags', 'docker', '-ldflags', 'cmd/main.go'])
    )
    expect(goCall.args.join(' ')).toContain('BuildVersion=dev-123456789abc')
    expect(goCall.args.join(' ')).toContain(`Commit=${commit}`)
    expect(goCall.env).toMatchObject({ GOOS: 'darwin', GOARCH: 'arm64', CGO_ENABLED: '0' })

    const stageDirectory = path.join(
      projectDirectory,
      'buildResources/.generated/proxy-router/mac-arm64'
    )
    const manifest = JSON.parse(
      await fs.readFile(path.join(stageDirectory, 'manifest.json'), 'utf8')
    )
    expect(manifest).toMatchObject({
      schemaVersion: 1,
      platform: 'darwin',
      arch: 'arm64',
      size: Buffer.byteLength('target-specific-router'),
      buildVersion: 'dev-123456789abc',
      commit,
      dirty: true
    })

    const packagedBundle = path.join(
      context.appOutDir,
      'MorpheusUI.app/Contents/Resources/proxy-router-bundle'
    )
    await fs.mkdir(packagedBundle, { recursive: true })
    await fs.copyFile(
      path.join(stageDirectory, 'bundled-proxy-router'),
      path.join(packagedBundle, 'bundled-proxy-router')
    )
    await fs.copyFile(
      path.join(stageDirectory, 'manifest.json'),
      path.join(packagedBundle, 'manifest.json')
    )
    await expect(
      verifyPackagedProxyRouterBundle(context, { environment: {} })
    ).resolves.toBeUndefined()
  })

  it('uses an injected release URL and removes a stale local bundle', async () => {
    const stageDirectory = path.join(
      projectDirectory,
      'buildResources/.generated/proxy-router/mac-arm64'
    )
    await fs.mkdir(stageDirectory, { recursive: true })
    await fs.writeFile(path.join(stageDirectory, 'bundled-proxy-router'), 'stale')
    const runCommand: BundleCommandRunner = vi.fn(async () => {
      throw new Error('Go and Git must not run in release URL mode')
    })

    await prepareProxyRouterBundle(createContext(), {
      runCommand,
      environment: {
        SERVICE_PROXY_DOWNLOAD_URL_MAC_ARM64: 'https://release.test/proxy-router'
      }
    })

    expect(runCommand).not.toHaveBeenCalled()
    expect(await fs.readdir(stageDirectory)).toEqual([])
  })

  it('fails clearly for unsupported package architectures', () => {
    expect(() => resolveProxyRouterBundleTarget('darwin', 'universal')).toThrow(
      'does not support architecture universal'
    )
    expect(() => resolveProxyRouterBundleTarget('win32', 'arm64')).toThrow(
      'does not support platform win32/arm64'
    )
  })
})
