import { execFile } from 'node:child_process'
import { copyFileSync, existsSync } from 'node:fs'
import { copyFile, mkdir, readFile, stat } from 'node:fs/promises'
import path from 'node:path'
import { promisify } from 'node:util'
import { Arch, type AfterPackContext } from 'electron-builder'

const execFileAsync = promisify(execFile)

/**
 * A native module compiles to one binary per platform, and electron-builder
 * packages whatever `npm install` happened to leave in node_modules. Packaging
 * Windows from a Mac therefore shipped a Mach-O `keytar.node`, and the app died
 * on launch with "is not a valid Win32 application" before a single window
 * appeared. The binary matching the pack target is fetched from the module's
 * published prebuilds before packing, and the machine's own binary is put back
 * afterwards so the checkout stays runnable.
 */
interface NativeModuleBinary {
  moduleName: string
  /** Path of the compiled binary inside the module, as its loader requires it. */
  relativePath: string
}

const NATIVE_MODULE_BINARIES: readonly NativeModuleBinary[] = [
  { moduleName: 'keytar', relativePath: path.join('build', 'Release', 'keytar.node') }
]

type NativeBinaryFormat = 'pe' | 'mach-o' | 'elf'

const FORMAT_BY_PLATFORM: Readonly<Record<string, NativeBinaryFormat>> = {
  win32: 'pe',
  darwin: 'mach-o',
  linux: 'elf'
}

/** electron-builder names some architectures differently from Node. */
const PREBUILD_ARCH: Readonly<Record<string, string>> = {
  ia32: 'ia32',
  x64: 'x64',
  arm64: 'arm64',
  armv7l: 'arm'
}

const MACH_O_MAGICS = new Set([
  0xfeedface, 0xfeedfacf, 0xcefaedfe, 0xcffaedfe,
  // The fat wrapper a universal build produces, in both byte orders.
  0xcafebabe, 0xbebafeca
])

/**
 * Identifies an object file from its first four bytes. Every format states what
 * it is in its header, so this needs no toolchain and works for a target the
 * build machine cannot execute.
 */
export function nativeBinaryFormat(header: Buffer): NativeBinaryFormat | null {
  if (header.length < 4) return null
  if (header[0] === 0x4d && header[1] === 0x5a) return 'pe'
  if (header[0] === 0x7f && header.toString('latin1', 1, 4) === 'ELF') return 'elf'
  if (MACH_O_MAGICS.has(header.readUInt32BE(0))) return 'mach-o'
  return null
}

const readBinaryFormat = async (file: string): Promise<NativeBinaryFormat | null> =>
  nativeBinaryFormat(await readFile(file))

const exists = (file: string): Promise<boolean> =>
  stat(file).then(
    () => true,
    () => false
  )

const hostKey = `${process.platform}-${process.arch}`

const cacheDirectory = (projectDir: string, moduleName: string, key: string): string =>
  path.join(projectDir, 'node_modules', '.cache', 'morpheus-native-binaries', moduleName, key)

const cachedBinary = (projectDir: string, binary: NativeModuleBinary, key: string): string =>
  path.join(cacheDirectory(projectDir, binary.moduleName, key), path.basename(binary.relativePath))

const liveBinary = (projectDir: string, binary: NativeModuleBinary): string =>
  path.join(projectDir, 'node_modules', binary.moduleName, binary.relativePath)

/**
 * Restores every staged binary the moment the build process ends, however it
 * ends. Without this, an interrupted package would leave a foreign binary in
 * node_modules and the next `npm run dev` would fail for a reason with nothing
 * to do with the change being tested.
 */
const stagedRestores = new Map<string, string>()
let restoreHookInstalled = false

function installRestoreHook(): void {
  if (restoreHookInstalled) return
  restoreHookInstalled = true
  process.once('exit', () => {
    for (const [live, hostCopy] of stagedRestores) {
      try {
        if (existsSync(hostCopy)) copyFileSync(hostCopy, live)
      } catch {
        // Nothing useful can be done during exit; the next build re-restores.
      }
    }
  })
}

async function ensureHostCopy(projectDir: string, binary: NativeModuleBinary): Promise<string> {
  const hostCopy = cachedBinary(projectDir, binary, hostKey)
  if (await exists(hostCopy)) return hostCopy
  const live = liveBinary(projectDir, binary)
  const format = await readBinaryFormat(live)
  if (format !== FORMAT_BY_PLATFORM[process.platform]) {
    throw new Error(
      `node_modules/${binary.moduleName} holds a ${format ?? 'unrecognised'} binary rather than one this machine can run. Reinstall dependencies before packaging.`
    )
  }
  await mkdir(path.dirname(hostCopy), { recursive: true })
  await copyFile(live, hostCopy)
  return hostCopy
}

async function fetchPrebuild(
  projectDir: string,
  binary: NativeModuleBinary,
  platform: string,
  arch: string
): Promise<void> {
  const installer = path.join(projectDir, 'node_modules', 'prebuild-install', 'bin.js')
  if (!(await exists(installer))) {
    throw new Error(
      `Cannot package ${platform}/${arch}: prebuild-install is not installed, so the ${binary.moduleName} binary for that target cannot be fetched.`
    )
  }
  try {
    await execFileAsync(
      process.execPath,
      [installer, `--platform=${platform}`, `--arch=${PREBUILD_ARCH[arch] ?? arch}`, '--force'],
      { cwd: path.join(projectDir, 'node_modules', binary.moduleName) }
    )
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error)
    throw new Error(
      `No published ${binary.moduleName} binary could be fetched for ${platform}/${arch}. Build this target on a ${platform} machine instead. (${detail})`
    )
  }
}

/**
 * Puts the target platform's native binaries in place. Called from `beforePack`,
 * once per architecture electron-builder is about to package.
 */
export async function stageNativeModuleBinaries(context: AfterPackContext): Promise<void> {
  const projectDir = context.packager.info.projectDir
  const platform = context.electronPlatformName
  const arch = Arch[context.arch]
  installRestoreHook()

  for (const binary of NATIVE_MODULE_BINARIES) {
    const live = liveBinary(projectDir, binary)
    // A module that was never installed is not this hook's problem to report.
    if (!(await exists(live))) continue
    const hostCopy = await ensureHostCopy(projectDir, binary)
    stagedRestores.set(live, hostCopy)

    // A universal macOS package is assembled from the per-architecture packages,
    // each of which comes through here with a real architecture of its own.
    if (arch === 'universal' || (platform === process.platform && arch === process.arch)) {
      await copyFile(hostCopy, live)
      continue
    }

    const staged = cachedBinary(projectDir, binary, `${platform}-${arch}`)
    if (!(await exists(staged))) {
      // Fetch into the module itself, which is where its loader looks, then keep
      // a copy so a second build of the same target needs no network at all.
      await copyFile(hostCopy, live)
      await fetchPrebuild(projectDir, binary, platform, arch).catch(async (error) => {
        await copyFile(hostCopy, live)
        throw error
      })
      const fetched = await readBinaryFormat(live)
      if (fetched !== FORMAT_BY_PLATFORM[platform]) {
        await copyFile(hostCopy, live)
        throw new Error(
          `The ${binary.moduleName} binary fetched for ${platform}/${arch} is ${fetched ?? 'unrecognised'}, not ${FORMAT_BY_PLATFORM[platform]}. Refusing to package an app that cannot start.`
        )
      }
      await mkdir(path.dirname(staged), { recursive: true })
      await copyFile(live, staged)
    }
    await copyFile(staged, live)
  }
}

/** Puts the build machine's own binaries back. Called from `afterPack`. */
export async function restoreNativeModuleBinaries(context: AfterPackContext): Promise<void> {
  const projectDir = context.packager.info.projectDir
  for (const binary of NATIVE_MODULE_BINARIES) {
    const hostCopy = cachedBinary(projectDir, binary, hostKey)
    if (!(await exists(hostCopy))) continue
    const live = liveBinary(projectDir, binary)
    await copyFile(hostCopy, live)
    stagedRestores.delete(live)
  }
}

const packagedResourcesDirectory = (context: AfterPackContext): string =>
  context.electronPlatformName === 'darwin'
    ? path.join(
        context.appOutDir,
        `${context.packager.appInfo.productFilename}.app`,
        'Contents',
        'Resources'
      )
    : path.join(context.appOutDir, 'resources')

/**
 * Fails the build if a packaged native binary is not the target's own format.
 * This is the check that turns "the app crashes on a teammate's laptop" into
 * "the build stopped and said which binary was wrong".
 */
export async function verifyPackagedNativeBinaries(context: AfterPackContext): Promise<void> {
  const platform = context.electronPlatformName
  const arch = Arch[context.arch]
  if (arch === 'universal') return
  const expected = FORMAT_BY_PLATFORM[platform]
  if (!expected) return

  for (const binary of NATIVE_MODULE_BINARIES) {
    const packaged = path.join(
      packagedResourcesDirectory(context),
      'app.asar.unpacked',
      'node_modules',
      binary.moduleName,
      binary.relativePath
    )
    if (!(await exists(packaged))) {
      throw new Error(
        `The packaged app is missing ${binary.moduleName} at ${packaged}, so it would fail to start.`
      )
    }
    const format = await readBinaryFormat(packaged)
    if (format !== expected) {
      throw new Error(
        `The packaged ${binary.moduleName} binary is ${format ?? 'unrecognised'} but ${platform}/${arch} needs ${expected}. The app would not start.`
      )
    }
  }
}
