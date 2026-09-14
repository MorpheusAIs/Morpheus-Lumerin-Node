import { execFile } from 'node:child_process'
import { createHash } from 'node:crypto'
import { chmod, mkdir, readFile, rm, stat, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { Arch, type AfterPackContext } from 'electron-builder'

export const ProxyRouterBundleDirectoryName = 'proxy-router-bundle'
export const ProxyRouterBundleExecutableName = 'bundled-proxy-router'
export const ProxyRouterBundleManifestName = 'manifest.json'

interface ProxyRouterBundleTarget {
  platform: 'darwin' | 'linux' | 'win32'
  arch: 'x64' | 'arm64'
  goos: 'darwin' | 'linux' | 'windows'
  goarch: 'amd64' | 'arm64'
  builderOs: 'mac' | 'linux' | 'win'
  downloadUrlEnvironmentVariable: string
}

export interface ProxyRouterBundleManifest {
  schemaVersion: 1
  platform: ProxyRouterBundleTarget['platform']
  arch: ProxyRouterBundleTarget['arch']
  size: number
  sha256: string
  buildVersion: string
  commit: string
  dirty: boolean
}

export type BundleEnvironment = Record<string, string | undefined>

interface CommandOptions {
  cwd: string
  env?: BundleEnvironment
}

export type BundleCommandRunner = (
  executable: string,
  args: string[],
  options: CommandOptions
) => Promise<{ stdout: string; stderr: string }>

export interface ProxyRouterBundleDependencies {
  runCommand: BundleCommandRunner
  environment: BundleEnvironment
}

const defaultRunCommand: BundleCommandRunner = (executable, args, options) =>
  new Promise((resolve, reject) => {
    execFile(
      executable,
      args,
      {
        cwd: options.cwd,
        env: options.env as NodeJS.ProcessEnv,
        encoding: 'utf8',
        maxBuffer: 10 * 1024 * 1024
      },
      (error, stdout, stderr) => {
        if (error) {
          reject(
            new Error(`${executable} ${args[0] ?? ''} failed: ${stderr.trim() || error.message}`)
          )
          return
        }
        resolve({ stdout, stderr })
      }
    )
  })

const defaultDependencies: ProxyRouterBundleDependencies = {
  runCommand: defaultRunCommand,
  environment: process.env as unknown as BundleEnvironment
}

export function resolveProxyRouterBundleTarget(
  platform: string,
  architecture: string
): ProxyRouterBundleTarget {
  if (architecture !== 'x64' && architecture !== 'arm64') {
    throw new Error(`Bundled proxy-router does not support architecture ${architecture}`)
  }

  if (platform === 'darwin') {
    return {
      platform,
      arch: architecture,
      goos: 'darwin',
      goarch: architecture === 'x64' ? 'amd64' : 'arm64',
      builderOs: 'mac',
      downloadUrlEnvironmentVariable:
        architecture === 'x64'
          ? 'SERVICE_PROXY_DOWNLOAD_URL_MAC_X64'
          : 'SERVICE_PROXY_DOWNLOAD_URL_MAC_ARM64'
    }
  }

  if (platform === 'linux') {
    return {
      platform,
      arch: architecture,
      goos: 'linux',
      goarch: architecture === 'x64' ? 'amd64' : 'arm64',
      builderOs: 'linux',
      downloadUrlEnvironmentVariable:
        architecture === 'x64'
          ? 'SERVICE_PROXY_DOWNLOAD_URL_LINUX_X64'
          : 'SERVICE_PROXY_DOWNLOAD_URL_LINUX_ARM64'
    }
  }

  if (platform === 'win32') {
    // The router's go-ole dependency does not compile for windows/arm64; the
    // repository's release matrix excludes it for the same reason. Fail the
    // package explicitly instead of shipping an Electron app with no router.
    if (architecture === 'arm64') {
      throw new Error('Bundled proxy-router does not support platform win32/arm64')
    }
    return {
      platform,
      arch: architecture,
      goos: 'windows',
      goarch: architecture === 'x64' ? 'amd64' : 'arm64',
      builderOs: 'win',
      downloadUrlEnvironmentVariable:
        architecture === 'x64'
          ? 'SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_X64'
          : 'SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_ARM64'
    }
  }

  throw new Error(`Bundled proxy-router does not support platform ${platform}`)
}

const sha256File = async (filePath: string) => {
  return createHash('sha256')
    .update(await readFile(filePath))
    .digest('hex')
}

const getTargetFromContext = (context: AfterPackContext) => {
  const architecture = Arch[context.arch]
  return resolveProxyRouterBundleTarget(context.electronPlatformName, architecture)
}

const getStageDirectory = (context: AfterPackContext, target: ProxyRouterBundleTarget) =>
  path.join(
    context.packager.info.projectDir,
    'buildResources',
    '.generated',
    'proxy-router',
    `${target.builderOs}-${target.arch}`
  )

const shouldBuildLocalBundle = (target: ProxyRouterBundleTarget, environment: BundleEnvironment) =>
  environment.FORCE_BUNDLED_PROXY_ROUTER === '1' ||
  !environment[target.downloadUrlEnvironmentVariable]?.trim()

const getPackagedResourcesDirectory = (context: AfterPackContext) => {
  if (context.electronPlatformName === 'darwin') {
    return path.join(
      context.appOutDir,
      `${context.packager.appInfo.productFilename}.app`,
      'Contents',
      'Resources'
    )
  }
  return path.join(context.appOutDir, 'resources')
}

const readManifest = async (manifestPath: string): Promise<ProxyRouterBundleManifest> => {
  const parsed = JSON.parse(await readFile(manifestPath, 'utf8')) as ProxyRouterBundleManifest
  if (
    parsed.schemaVersion !== 1 ||
    typeof parsed.platform !== 'string' ||
    typeof parsed.arch !== 'string' ||
    !Number.isSafeInteger(parsed.size) ||
    parsed.size <= 0 ||
    !/^[a-f0-9]{64}$/.test(parsed.sha256) ||
    typeof parsed.buildVersion !== 'string' ||
    !parsed.buildVersion ||
    !/^[a-f0-9]{40}$/.test(parsed.commit) ||
    typeof parsed.dirty !== 'boolean'
  ) {
    throw new Error(`Invalid bundled proxy-router manifest at ${manifestPath}`)
  }
  return parsed
}

/** Build the current proxy source for local packages; release CI keeps its injected URL. */
export async function prepareProxyRouterBundle(
  context: AfterPackContext,
  dependencies: ProxyRouterBundleDependencies = defaultDependencies
): Promise<void> {
  const target = getTargetFromContext(context)
  const stageDirectory = getStageDirectory(context, target)

  // Clean only this exact target so stale binaries cannot leak into a package.
  await rm(stageDirectory, { recursive: true, force: true })
  await mkdir(stageDirectory, { recursive: true })

  if (!shouldBuildLocalBundle(target, dependencies.environment)) {
    console.info(
      `Using ${target.downloadUrlEnvironmentVariable}; no local proxy-router will be bundled`
    )
    return
  }

  const projectDirectory = context.packager.info.projectDir
  const repositoryDirectory = path.resolve(projectDirectory, '..')
  const proxyRouterDirectory = path.join(repositoryDirectory, 'proxy-router')
  const commit = (
    await dependencies.runCommand('git', ['rev-parse', 'HEAD'], {
      cwd: repositoryDirectory
    })
  ).stdout.trim()
  const shortCommit = (
    await dependencies.runCommand('git', ['rev-parse', '--short=12', 'HEAD'], {
      cwd: repositoryDirectory
    })
  ).stdout.trim()
  const dirty = Boolean(
    (
      await dependencies.runCommand('git', ['status', '--porcelain'], {
        cwd: repositoryDirectory
      })
    ).stdout.trim()
  )
  const buildVersion = `dev-${shortCommit}`
  const outputPath = path.join(stageDirectory, ProxyRouterBundleExecutableName)
  const linkerFlags = [
    '-s',
    '-w',
    '-X',
    `github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.BuildVersion=${buildVersion}`,
    '-X',
    `github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.Commit=${commit}`
  ].join(' ')

  console.info(`Building bundled proxy-router for ${target.platform}/${target.arch}`)
  await dependencies.runCommand(
    'go',
    [
      'build',
      '-trimpath',
      '-tags',
      'docker',
      '-ldflags',
      linkerFlags,
      '-o',
      outputPath,
      'cmd/main.go'
    ],
    {
      cwd: proxyRouterDirectory,
      env: {
        ...dependencies.environment,
        GOOS: target.goos,
        GOARCH: target.goarch,
        CGO_ENABLED: '0'
      }
    }
  )

  if (target.platform !== 'win32') {
    await chmod(outputPath, 0o755)
  }
  const outputInfo = await stat(outputPath)
  if (!outputInfo.isFile() || outputInfo.size === 0) {
    throw new Error(`Go produced an empty bundled proxy-router at ${outputPath}`)
  }

  const manifest: ProxyRouterBundleManifest = {
    schemaVersion: 1,
    platform: target.platform,
    arch: target.arch,
    size: outputInfo.size,
    sha256: await sha256File(outputPath),
    buildVersion,
    commit,
    dirty
  }
  await writeFile(
    path.join(stageDirectory, ProxyRouterBundleManifestName),
    `${JSON.stringify(manifest, null, 2)}\n`,
    { mode: 0o644 }
  )
}

/** Fail packaging if the bundle copied by electron-builder is missing or inconsistent. */
export async function verifyPackagedProxyRouterBundle(
  context: AfterPackContext,
  dependencies: Pick<ProxyRouterBundleDependencies, 'environment'> = defaultDependencies
): Promise<void> {
  const target = getTargetFromContext(context)
  const bundleDirectory = path.join(
    getPackagedResourcesDirectory(context),
    ProxyRouterBundleDirectoryName
  )
  const executablePath = path.join(bundleDirectory, ProxyRouterBundleExecutableName)
  const manifestPath = path.join(bundleDirectory, ProxyRouterBundleManifestName)

  if (!shouldBuildLocalBundle(target, dependencies.environment)) {
    const unexpectedlyBundled = await stat(executablePath)
      .then(() => true)
      .catch((error: NodeJS.ErrnoException) => {
        if (error.code === 'ENOENT') return false
        throw error
      })
    if (unexpectedlyBundled) {
      throw new Error('Release URL build unexpectedly contains a local proxy-router bundle')
    }
    return
  }

  const manifest = await readManifest(manifestPath)
  if (manifest.platform !== target.platform || manifest.arch !== target.arch) {
    throw new Error(
      `Bundled proxy-router target ${manifest.platform}/${manifest.arch} does not match ${target.platform}/${target.arch}`
    )
  }
  const executableInfo = await stat(executablePath)
  const executableHash = await sha256File(executablePath)
  if (executableInfo.size !== manifest.size || executableHash !== manifest.sha256) {
    throw new Error('Packaged proxy-router does not match its build manifest')
  }
}
