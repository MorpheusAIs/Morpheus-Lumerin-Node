import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { chmod, mkdir, readFile, rename, rm, stat, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url))
const projectDirectory = path.resolve(scriptDirectory, '..')
const repositoryDirectory = path.resolve(projectDirectory, '..')
const proxyRouterDirectory = path.join(repositoryDirectory, 'proxy-router')

const targetForHost = () => {
  if (process.arch !== 'x64' && process.arch !== 'arm64') {
    throw new Error(`The local proxy-router does not support ${process.arch}`)
  }
  if (process.platform === 'win32' && process.arch === 'arm64') {
    throw new Error('The local proxy-router does not support win32/arm64')
  }
  if (process.platform === 'darwin') {
    return { builderOs: 'mac', goos: 'darwin', goarch: process.arch === 'x64' ? 'amd64' : 'arm64' }
  }
  if (process.platform === 'linux') {
    return { builderOs: 'linux', goos: 'linux', goarch: process.arch === 'x64' ? 'amd64' : 'arm64' }
  }
  if (process.platform === 'win32') {
    return { builderOs: 'win', goos: 'windows', goarch: 'amd64' }
  }
  throw new Error(`The local proxy-router does not support ${process.platform}`)
}

const git = (...args) =>
  execFileSync('git', args, { cwd: repositoryDirectory, encoding: 'utf8' }).trim()

const target = targetForHost()
const commit = git('rev-parse', 'HEAD')
const shortCommit = git('rev-parse', '--short=12', 'HEAD')
const dirty = Boolean(git('status', '--porcelain'))
const buildVersion = `dev-${shortCommit}`
const stageDirectory = path.join(
  projectDirectory,
  'buildResources',
  '.generated',
  'proxy-router',
  `${target.builderOs}-${process.arch}`
)
const temporaryDirectory = `${stageDirectory}.${process.pid}.tmp`
const outputPath = path.join(temporaryDirectory, 'bundled-proxy-router')

await rm(temporaryDirectory, { recursive: true, force: true })
await mkdir(temporaryDirectory, { recursive: true })

try {
  const linkerFlags = [
    '-s',
    '-w',
    '-X',
    `github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.BuildVersion=${buildVersion}`,
    '-X',
    `github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config.Commit=${commit}`
  ].join(' ')

  console.info(
    `Building current proxy-router for local development (${process.platform}/${process.arch})`
  )
  execFileSync(
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
        ...process.env,
        GOOS: target.goos,
        GOARCH: target.goarch,
        CGO_ENABLED: '0'
      },
      stdio: 'inherit'
    }
  )

  if (process.platform !== 'win32') await chmod(outputPath, 0o755)
  const outputInfo = await stat(outputPath)
  if (!outputInfo.isFile() || outputInfo.size === 0) {
    throw new Error('Go produced an empty local proxy-router')
  }
  const sha256 = createHash('sha256')
    .update(await readFile(outputPath))
    .digest('hex')
  await writeFile(
    path.join(temporaryDirectory, 'manifest.json'),
    `${JSON.stringify(
      {
        schemaVersion: 1,
        platform: process.platform,
        arch: process.arch,
        size: outputInfo.size,
        sha256,
        buildVersion,
        commit,
        dirty
      },
      null,
      2
    )}\n`
  )

  await rm(stageDirectory, { recursive: true, force: true })
  await mkdir(path.dirname(stageDirectory), { recursive: true })
  await rename(temporaryDirectory, stageDirectory)
} finally {
  await rm(temporaryDirectory, { recursive: true, force: true })
}
