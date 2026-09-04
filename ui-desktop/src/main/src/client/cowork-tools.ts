import { app, shell } from 'electron'
import { randomUUID } from 'node:crypto'
import { constants, promises as fs } from 'node:fs'
import path from 'node:path'
import { CoworkApprovalMode, CoworkArtifact, CoworkProject } from './cowork.types'
import { analyzeCsvText } from './cowork-data-analysis'
import type { ProfessionalArtifactFormat } from './cowork-professional-artifacts'

const MAX_READ_BYTES = 512 * 1024
const MAX_DOCUMENT_READ_BYTES = 20 * 1024 * 1024
const MAX_WRITE_BYTES = 2 * 1024 * 1024
const MAX_LIST_ENTRIES = 800
const MAX_SEARCH_FILES = 400
const MAX_SEARCH_MATCHES = 120
const MAX_BACKUPS_PER_PROJECT = 100
const MAX_BACKUP_FILE_BYTES = 64 * 1024 * 1024
const MAX_BACKUP_TOTAL_BYTES = 512 * 1024 * 1024
const MAX_COPY_BYTES = 64 * 1024 * 1024
const mutationTails = new Map<string, Promise<void>>()
const SENSITIVE_NAMES = new Set([
  '.env',
  '.npmrc',
  '.pypirc',
  '.netrc',
  '.git-credentials',
  '.git',
  '.ssh',
  '.gnupg',
  '.aws',
  '.azure',
  '.kube',
  '.docker',
  '.secrets',
  'credentials',
  'credentials.json',
  'auth.json',
  'tokens.json',
  'secrets.json',
  '.cookie',
  'models-config.json',
  'wallets.json',
  '.password-store',
  'login.keychain-db',
  'mnemonic',
  'mnemonic.txt',
  'seed',
  'seed.txt',
  'seedphrase',
  'seed-phrase',
  'privatekey',
  'private-key',
  'private_key',
  'id_rsa',
  'id_ed25519',
  'keystore',
  'wallet.dat'
])

const WINDOWS_RESERVED_NAME = /^(?:con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\..*)?$/i

function portableRelativePath(requestedPath: string): string {
  if (/[\0-\x1f\x7f]/.test(requestedPath))
    throw new Error('Paths cannot contain control characters.')
  if (requestedPath.includes(':')) throw new Error('Paths cannot contain colon characters.')
  const portable = requestedPath.replace(/[\\/]+/g, path.sep)
  for (const part of portable.split(path.sep).filter(Boolean)) {
    if (part !== '.' && part !== '..' && /[. ]$/.test(part)) {
      throw new Error('Path segments cannot end with a dot or space.')
    }
    if (WINDOWS_RESERVED_NAME.test(part)) throw new Error('This path uses a reserved system name.')
  }
  return portable
}

const sensitivePath = (relative: string): boolean => {
  const parts = relative
    .split(/[\\/]/)
    .filter(Boolean)
    .map((part) => part.toLowerCase())
  return parts.some(
    (part) =>
      SENSITIVE_NAMES.has(part) ||
      part.startsWith('.env.') ||
      part.endsWith('.pem') ||
      part.endsWith('.key') ||
      part.endsWith('.p12') ||
      part.endsWith('.pfx') ||
      part.endsWith('.jks') ||
      part.endsWith('.keystore')
  )
}

const coworkConfigurationPath = (relative: string): boolean => {
  const parts = relative
    .split(/[\\/]/)
    .filter(Boolean)
    .map((part) => part.toLowerCase())
  return parts[0] === '.morpheus' && parts[1] === 'cowork'
}

function assertAllowedRelativePath(relative: string): void {
  if (coworkConfigurationPath(relative)) {
    throw new Error(
      'Workspace extension configuration can only be changed by the user-facing project controls.'
    )
  }
  if (sensitivePath(relative)) {
    throw new Error('Credential and key-material paths are blocked from Workspace tasks.')
  }
}

const pathInside = (root: string, candidate: string): boolean =>
  candidate === root || candidate.startsWith(root + path.sep)

async function canonicalRoot(project: CoworkProject): Promise<string> {
  const root = await fs.realpath(project.rootPath)
  const stat = await fs.stat(root)
  if (!stat.isDirectory()) throw new Error('The connected project folder is unavailable.')
  return root
}

async function nearestExistingParent(candidate: string): Promise<string> {
  let current = candidate
  while (true) {
    try {
      return await fs.realpath(current)
    } catch (error: any) {
      if (error?.code !== 'ENOENT') throw error
      const parent = path.dirname(current)
      if (parent === current) throw error
      current = parent
    }
  }
}

async function assertNoSymlinkComponents(root: string, absolute: string): Promise<void> {
  let current = root
  for (const part of path.relative(root, absolute).split(path.sep).filter(Boolean)) {
    current = path.join(current, part)
    const stat = await fs.lstat(current).catch((error: any) => {
      if (error?.code === 'ENOENT') return null
      throw error
    })
    if (!stat) return
    if (stat.isSymbolicLink())
      throw new Error('Symbolic-link paths are blocked from Workspace tasks.')
  }
}

export async function resolveProjectPath(
  project: CoworkProject,
  requestedPath = '.',
  allowMissing = false
): Promise<{ root: string; absolute: string; relative: string }> {
  const root = await canonicalRoot(project)
  const portablePath = portableRelativePath(requestedPath)
  if (path.isAbsolute(portablePath)) {
    throw new Error('Use a path relative to the connected project folder.')
  }
  const absolute = path.resolve(root, portablePath || '.')
  if (!pathInside(root, absolute)) throw new Error('Path escapes the connected project folder.')
  const relative = path.relative(root, absolute) || '.'
  if (relative !== '.') assertAllowedRelativePath(relative)
  if (allowMissing) {
    const existing = await nearestExistingParent(absolute)
    if (!pathInside(root, existing)) throw new Error('Path resolves outside the connected folder.')
    const canonicalRelative = path.relative(root, existing) || '.'
    if (canonicalRelative !== '.') assertAllowedRelativePath(canonicalRelative)
    const stat = await fs.stat(existing)
    if (stat.isFile() && stat.nlink > 1) {
      throw new Error('Hard-linked files are blocked from Workspace tasks.')
    }
  } else {
    const real = await fs.realpath(absolute)
    if (!pathInside(root, real)) throw new Error('Path resolves outside the connected folder.')
    const canonicalRelative = path.relative(root, real) || '.'
    if (canonicalRelative !== '.') assertAllowedRelativePath(canonicalRelative)
    const stat = await fs.stat(real)
    if (stat.isFile() && stat.nlink > 1) {
      throw new Error('Hard-linked files are blocked from Workspace tasks.')
    }
  }
  await assertNoSymlinkComponents(root, absolute)

  return { root, absolute, relative }
}

const ignoredDirectory = (name: string) => name === '.git' || name === 'node_modules'
const ignoredEntry = (relative: string) =>
  sensitivePath(relative) || coworkConfigurationPath(relative)

async function walk(
  root: string,
  current: string,
  depth: number,
  maxDepth: number,
  entries: Array<{ path: string; type: 'file' | 'directory'; size?: number }>
): Promise<void> {
  if (entries.length >= MAX_LIST_ENTRIES) return
  const dirents = await fs.readdir(current, { withFileTypes: true })
  dirents.sort((a, b) => a.name.localeCompare(b.name))
  for (const entry of dirents) {
    if (entries.length >= MAX_LIST_ENTRIES) return
    if (entry.isSymbolicLink()) continue
    const absolute = path.join(current, entry.name)
    const relative = path.relative(root, absolute)
    if (ignoredEntry(relative)) continue
    if (entry.isDirectory()) {
      entries.push({ path: relative, type: 'directory' })
      if (depth < maxDepth && !ignoredDirectory(entry.name)) {
        await walk(root, absolute, depth + 1, maxDepth, entries)
      }
    } else if (entry.isFile()) {
      const stat = await fs.stat(absolute)
      if (stat.nlink > 1) continue
      entries.push({ path: relative, type: 'file', size: stat.size })
    }
  }
}

const parseArguments = (value: string): Record<string, any> => {
  try {
    const parsed = JSON.parse(value || '{}')
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      throw new Error('Tool arguments must be an object.')
    }
    return parsed
  } catch (error: any) {
    throw new Error(`Invalid tool arguments: ${error.message}`)
  }
}

export const toolArguments = parseArguments

export const isCoworkMutationTool = (toolName: string): boolean =>
  [
    'write_file',
    'make_directory',
    'copy_file',
    'move_file',
    'delete_file',
    'create_docx',
    'create_xlsx',
    'create_pptx',
    'create_pdf'
  ].includes(toolName)

/** Serializes mutation classification and execution for every task sharing a project. */
export async function withCoworkMutationLock<T>(
  projectId: string,
  operation: () => Promise<T>
): Promise<T> {
  const previous = mutationTails.get(projectId) ?? Promise.resolve()
  let release!: () => void
  const gate = new Promise<void>((resolve) => {
    release = resolve
  })
  const tail = previous.catch(() => undefined).then(() => gate)
  mutationTails.set(projectId, tail)
  await previous.catch(() => undefined)
  try {
    return await operation()
  } finally {
    release()
    if (mutationTails.get(projectId) === tail) mutationTails.delete(projectId)
  }
}

export async function approvalRequirement(
  project: CoworkProject,
  toolName: string,
  input: Record<string, any>
): Promise<{ reason: string; risk: 'write' | 'overwrite' | 'delete' } | null> {
  if (toolName === 'delete_file') {
    const target = await resolveProjectPath(project, String(input.path ?? ''), true)
    if (target.absolute === target.root)
      throw new Error('The connected project folder itself cannot be deleted.')
    return { reason: `Move “${target.relative}” to the system trash.`, risk: 'delete' }
  }
  const mutation = [
    'write_file',
    'make_directory',
    'copy_file',
    'move_file',
    'create_docx',
    'create_xlsx',
    'create_pptx',
    'create_pdf'
  ].includes(toolName)
  if (!mutation) return null

  const destination = String(input.path ?? input.destination ?? '').trim()
  if (!destination || destination === '.')
    throw new Error(`${toolName} requires a path below the project root.`)
  const resolved = await resolveProjectPath(project, destination, true)
  if (toolName === 'copy_file' || toolName === 'move_file') {
    const source = String(input.source ?? '').trim()
    if (!source || source === '.')
      throw new Error(`${toolName} requires a source below the project root.`)
    const resolvedSource = await resolveProjectPath(project, source, true)
    if (resolvedSource.absolute === resolved.absolute) {
      throw new Error(`${toolName} source and destination must be different paths.`)
    }
  }
  // Skip mode is explicit project-scoped consent for file writes. Destructive
  // deletion remains gated above in every mode.
  if (project.approvalMode === 'skip') return null

  const destinationStat = await fs.lstat(resolved.absolute).catch((error: any) => {
    if (error?.code === 'ENOENT') return null
    throw error
  })
  if (destinationStat?.isSymbolicLink())
    throw new Error('Symbolic-link paths are blocked from Workspace tasks.')
  if (
    destinationStat &&
    [
      'write_file',
      'move_file',
      'copy_file',
      'create_docx',
      'create_xlsx',
      'create_pptx',
      'create_pdf'
    ].includes(toolName)
  ) {
    return {
      reason: `Allow replacing the existing path “${destination}”. A backup will be kept.`,
      risk: 'overwrite'
    }
  }
  if (
    project.approvalMode === 'manual' ||
    (project.approvalMode === 'auto' && toolName === 'move_file')
  ) {
    return {
      reason: `Allow ${toolName.replaceAll('_', ' ')} for “${destination}” in the connected folder.`,
      risk: 'write'
    }
  }
  return null
}

function sameFile(
  a: { dev: number | bigint; ino: number | bigint },
  b: { dev: number | bigint; ino: number | bigint }
): boolean {
  return String(a.dev) === String(b.dev) && String(a.ino) === String(b.ino)
}

async function openStableFile(absolute: string, writable = false) {
  const before = await fs.lstat(absolute)
  if (before.isSymbolicLink())
    throw new Error('Symbolic-link paths are blocked from Workspace tasks.')
  if (!before.isFile()) throw new Error('This action requires a regular file.')
  if (before.nlink > 1) throw new Error('Hard-linked files are blocked from Workspace tasks.')
  const flags = (writable ? constants.O_RDWR : constants.O_RDONLY) | (constants.O_NOFOLLOW ?? 0)
  const handle = await fs.open(absolute, flags)
  try {
    const after = await handle.stat()
    if (!sameFile(before, after) || !after.isFile() || after.nlink > 1) {
      throw new Error('The file changed while Workspace was opening it. Try the action again.')
    }
    return { handle, stat: after }
  } catch (error) {
    await handle.close().catch(() => undefined)
    throw error
  }
}

async function readStableFile(
  absolute: string,
  maxBytes: number
): Promise<{ buffer: Buffer; stat: Awaited<ReturnType<typeof fs.stat>> }> {
  const opened = await openStableFile(absolute)
  try {
    if (opened.stat.size > maxBytes) {
      throw new Error(`File exceeds the ${Math.floor(maxBytes / 1024 / 1024)} MB copy limit.`)
    }
    return { buffer: await opened.handle.readFile(), stat: opened.stat }
  } finally {
    await opened.handle.close().catch(() => undefined)
  }
}

async function storeBackup(
  project: CoworkProject,
  absolute: string,
  buffer: Buffer
): Promise<void> {
  if (buffer.byteLength > MAX_BACKUP_FILE_BYTES) {
    throw new Error(
      'The existing file is too large to back up safely, so Workspace will not overwrite it.'
    )
  }
  const backupDir = path.join(app.getPath('userData'), 'CoworkBackups', project.id)
  await fs.mkdir(backupDir, { recursive: true, mode: 0o700 })
  await fs.chmod(backupDir, 0o700)
  const safeName = path.basename(absolute).replace(/[^a-zA-Z0-9._-]/g, '_')
  const records = await Promise.all(
    (await fs.readdir(backupDir)).map(async (name) => {
      const target = path.join(backupDir, name)
      const stat = await fs.stat(target).catch(() => null)
      return stat?.isFile() ? { target, size: stat.size, mtimeMs: stat.mtimeMs } : null
    })
  )
  const existing = records
    .filter((record): record is NonNullable<typeof record> => Boolean(record))
    .sort((a, b) => a.mtimeMs - b.mtimeMs)
  let total = existing.reduce((sum, record) => sum + record.size, 0)
  while (
    existing.length >= MAX_BACKUPS_PER_PROJECT ||
    total + buffer.byteLength > MAX_BACKUP_TOTAL_BYTES
  ) {
    const oldest = existing.shift()
    if (!oldest) break
    await fs.unlink(oldest.target).catch(() => undefined)
    total -= oldest.size
  }
  const backupPath = path.join(backupDir, `${Date.now()}-${randomUUID()}-${safeName}`)
  await fs.writeFile(backupPath, buffer, { flag: 'wx', mode: 0o600 })
}

async function writeDestination(
  project: CoworkProject,
  absolute: string,
  content: string | Buffer,
  allowOverwrite: boolean
): Promise<void> {
  const existing = await fs.lstat(absolute).catch((error: any) => {
    if (error?.code === 'ENOENT') return null
    throw error
  })
  if (!existing) {
    try {
      await fs.writeFile(absolute, content, { flag: 'wx', mode: 0o600 })
      return
    } catch (error: any) {
      if (error?.code === 'EEXIST') {
        throw new Error(
          'The destination was created by another process. Retry so Workspace can request overwrite approval.'
        )
      }
      throw error
    }
  }
  if (!allowOverwrite) {
    throw new Error(
      'The destination now exists. Retry so Workspace can request overwrite approval.'
    )
  }
  const opened = await openStableFile(absolute, true)
  try {
    const original = await opened.handle.readFile()
    await storeBackup(project, absolute, original)
    const replacement = Buffer.isBuffer(content) ? content : Buffer.from(content, 'utf8')
    await opened.handle.truncate(0)
    await opened.handle.write(replacement, 0, replacement.byteLength, 0)
    await opened.handle.truncate(replacement.byteLength)
    await opened.handle.sync()
  } finally {
    await opened.handle.close().catch(() => undefined)
  }
}

const artifact = (relative: string, kind: CoworkArtifact['kind']): CoworkArtifact => ({
  path: relative,
  name: path.basename(relative),
  kind,
  createdAt: Date.now(),
  updatedAt: Date.now()
})

export async function executeCoworkTool(
  project: CoworkProject,
  toolName: string,
  input: Record<string, any>,
  options: { allowOverwrite?: boolean } = {}
): Promise<{ result: unknown; artifact?: CoworkArtifact }> {
  switch (toolName) {
    case 'list_files': {
      const target = await resolveProjectPath(project, String(input.path ?? '.'))
      const stat = await fs.stat(target.absolute)
      if (!stat.isDirectory()) throw new Error('list_files requires a directory path.')
      const entries: Array<{ path: string; type: 'file' | 'directory'; size?: number }> = []
      const requestedDepth = Number(input.maxDepth ?? 2)
      const maxDepth = Number.isFinite(requestedDepth)
        ? Math.min(Math.max(Math.trunc(requestedDepth), 0), 6)
        : 2
      await walk(target.root, target.absolute, 0, maxDepth, entries)
      return { result: { entries, truncated: entries.length >= MAX_LIST_ENTRIES } }
    }
    case 'inspect_file': {
      const target = await resolveProjectPath(project, String(input.path ?? ''))
      const stat = await fs.stat(target.absolute)
      return {
        result: {
          path: target.relative,
          type: stat.isDirectory() ? 'directory' : 'file',
          size: stat.size,
          modifiedAt: stat.mtime.toISOString()
        }
      }
    }
    case 'read_file': {
      const target = await resolveProjectPath(project, String(input.path ?? ''))
      const stat = await fs.stat(target.absolute)
      if (!stat.isFile()) throw new Error('read_file requires a file path.')
      if (stat.size > MAX_READ_BYTES) {
        throw new Error(
          `File is too large to read safely (${stat.size} bytes; limit ${MAX_READ_BYTES}).`
        )
      }
      const { buffer } = await readStableFile(target.absolute, MAX_READ_BYTES)
      if (buffer.includes(0)) throw new Error('Binary files cannot be read as text.')
      const lines = buffer.toString('utf8').split(/\r?\n/)
      const parsedStart = Number(input.startLine ?? 1)
      const start = Number.isFinite(parsedStart) ? Math.max(1, Math.trunc(parsedStart)) : 1
      const parsedEnd = Number(input.endLine ?? start + 399)
      const end = Number.isFinite(parsedEnd)
        ? Math.max(start, Math.min(lines.length, Math.trunc(parsedEnd)))
        : Math.min(lines.length, start + 399)
      return {
        result: {
          path: target.relative,
          startLine: start,
          endLine: end,
          totalLines: lines.length,
          content: lines
            .slice(start - 1, end)
            .join('\n')
            .slice(0, 160_000)
        }
      }
    }
    case 'read_document': {
      const target = await resolveProjectPath(project, String(input.path ?? ''))
      const { buffer } = await readStableFile(target.absolute, MAX_DOCUMENT_READ_BYTES)
      // Keep the heavyweight parsers out of startup and ordinary text-only
      // tasks. The extractor itself rejects active OOXML content, external
      // relationships, encrypted files, and oversized archives.
      const { extractCoworkDocument } = await import('./cowork-document-extraction')
      const extracted = await extractCoworkDocument(buffer, target.relative)
      return {
        result: {
          path: target.relative,
          format: extracted.format,
          text: extracted.text,
          sections: extracted.sections.map(({ text, ...section }) => ({
            ...section,
            characters: text.length
          })),
          metadata: extracted.metadata,
          empty: extracted.empty,
          truncated: extracted.truncated,
          warnings: extracted.warnings
        }
      }
    }
    case 'search_files': {
      const query = String(input.query ?? '')
      if (!query) throw new Error('search_files requires a non-empty query.')
      const target = await resolveProjectPath(project, String(input.path ?? '.'))
      const candidates: Array<{ path: string; type: 'file' | 'directory'; size?: number }> = []
      await walk(target.root, target.absolute, 0, 6, candidates)
      const matches: Array<{ path: string; line: number; text: string }> = []
      let inspected = 0
      for (const candidate of candidates) {
        if (
          candidate.type !== 'file' ||
          inspected >= MAX_SEARCH_FILES ||
          matches.length >= MAX_SEARCH_MATCHES
        )
          continue
        if ((candidate.size ?? 0) > MAX_READ_BYTES) continue
        inspected++
        const absolute = path.join(target.root, candidate.path)
        const { buffer } = await readStableFile(absolute, MAX_READ_BYTES).catch(() => ({
          buffer: null
        }))
        if (!buffer) continue
        if (buffer.includes(0)) continue
        buffer
          .toString('utf8')
          .split(/\r?\n/)
          .forEach((line, index) => {
            if (
              matches.length < MAX_SEARCH_MATCHES &&
              line.toLowerCase().includes(query.toLowerCase())
            ) {
              matches.push({ path: candidate.path, line: index + 1, text: line.slice(0, 500) })
            }
          })
      }
      return {
        result: {
          matches,
          inspectedFiles: inspected,
          truncated: matches.length >= MAX_SEARCH_MATCHES
        }
      }
    }
    case 'analyze_csv': {
      const target = await resolveProjectPath(project, String(input.path ?? ''))
      const { buffer } = await readStableFile(target.absolute, MAX_READ_BYTES)
      if (buffer.includes(0)) throw new Error('Binary files cannot be analyzed as CSV text.')
      return {
        result: {
          path: target.relative,
          ...analyzeCsvText(buffer.toString('utf8'), String(input.delimiter ?? ','))
        }
      }
    }
    case 'write_file': {
      const content = String(input.content ?? '')
      if (Buffer.byteLength(content, 'utf8') > MAX_WRITE_BYTES)
        throw new Error('File content exceeds the 2 MB write limit.')
      const requested = String(input.path ?? '').trim()
      if (!requested || requested === '.')
        throw new Error('write_file requires a file path, not the project root.')
      const target = await resolveProjectPath(project, requested, true)
      await fs.mkdir(path.dirname(target.absolute), { recursive: true })
      await resolveProjectPath(project, requested, true)
      await writeDestination(project, target.absolute, content, options.allowOverwrite === true)
      return {
        result: { path: target.relative, bytes: Buffer.byteLength(content, 'utf8') },
        artifact: artifact(target.relative, 'file')
      }
    }
    case 'create_docx':
    case 'create_xlsx':
    case 'create_pptx':
    case 'create_pdf': {
      const format = toolName.slice('create_'.length) as ProfessionalArtifactFormat
      const requested = String(input.path ?? '').trim()
      if (!requested || requested === '.') {
        throw new Error(`${toolName} requires an output path below the project root.`)
      }
      if (path.extname(requested).toLowerCase() !== `.${format}`) {
        throw new Error(`${toolName} requires a .${format} output path.`)
      }
      const target = await resolveProjectPath(project, requested, true)
      const { path: _requestedPath, ...request } = input
      // DOCX/XLSX/PPTX/PDF generators pull in several large libraries. Keep
      // them out of the normal app-start path and load them only when a task
      // actually requests a professional artifact.
      const { generateProfessionalArtifact } = await import('./cowork-professional-artifacts')
      const generated = await generateProfessionalArtifact({ ...request, format })
      await fs.mkdir(path.dirname(target.absolute), { recursive: true })
      await resolveProjectPath(project, requested, true)
      await writeDestination(
        project,
        target.absolute,
        generated.buffer,
        options.allowOverwrite === true
      )
      return {
        result: {
          path: target.relative,
          bytes: generated.sizeBytes,
          format: generated.format,
          mimeType: generated.mimeType
        },
        artifact: artifact(target.relative, 'file')
      }
    }
    case 'make_directory': {
      const requested = String(input.path ?? '').trim()
      if (!requested || requested === '.')
        throw new Error('make_directory requires a path below the project root.')
      const target = await resolveProjectPath(project, requested, true)
      await fs.mkdir(target.absolute, { recursive: true })
      return { result: { path: target.relative }, artifact: artifact(target.relative, 'folder') }
    }
    case 'copy_file': {
      const requestedSource = String(input.source ?? '').trim()
      const requestedDestination = String(input.destination ?? '').trim()
      if (
        !requestedSource ||
        requestedSource === '.' ||
        !requestedDestination ||
        requestedDestination === '.'
      ) {
        throw new Error(
          'copy_file requires source and destination file paths below the project root.'
        )
      }
      const source = await resolveProjectPath(project, requestedSource)
      if (
        source.absolute === (await resolveProjectPath(project, requestedDestination, true)).absolute
      ) {
        throw new Error('copy_file source and destination must be different paths.')
      }
      const { buffer } = await readStableFile(source.absolute, MAX_COPY_BYTES)
      const destination = await resolveProjectPath(project, requestedDestination, true)
      await fs.mkdir(path.dirname(destination.absolute), { recursive: true })
      await resolveProjectPath(project, requestedDestination, true)
      await writeDestination(project, destination.absolute, buffer, options.allowOverwrite === true)
      return {
        result: { source: source.relative, destination: destination.relative },
        artifact: artifact(destination.relative, 'file')
      }
    }
    case 'move_file': {
      const requestedSource = String(input.source ?? '').trim()
      const requestedDestination = String(input.destination ?? '').trim()
      if (
        !requestedSource ||
        requestedSource === '.' ||
        !requestedDestination ||
        requestedDestination === '.'
      ) {
        throw new Error('move_file requires source and destination paths below the project root.')
      }
      const source = await resolveProjectPath(project, requestedSource)
      if (
        source.absolute === (await resolveProjectPath(project, requestedDestination, true)).absolute
      ) {
        throw new Error('move_file source and destination must be different paths.')
      }
      const { buffer, stat: sourceStat } = await readStableFile(source.absolute, MAX_COPY_BYTES)
      const destination = await resolveProjectPath(project, requestedDestination, true)
      await fs.mkdir(path.dirname(destination.absolute), { recursive: true })
      await resolveProjectPath(project, requestedDestination, true)
      await writeDestination(project, destination.absolute, buffer, options.allowOverwrite === true)
      const currentSource = await fs.lstat(source.absolute)
      if (
        !sameFile(sourceStat, currentSource) ||
        !currentSource.isFile() ||
        currentSource.nlink > 1
      ) {
        throw new Error(
          'The source changed while it was being moved. The destination copy was kept, but the source was not deleted.'
        )
      }
      await fs.unlink(source.absolute)
      return {
        result: { source: source.relative, destination: destination.relative },
        artifact: artifact(destination.relative, 'file')
      }
    }
    case 'delete_file': {
      const target = await resolveProjectPath(project, String(input.path ?? ''))
      if (target.absolute === target.root)
        throw new Error('The connected project folder itself cannot be deleted.')
      await shell.trashItem(target.absolute)
      return { result: { path: target.relative, movedToTrash: true } }
    }
    default:
      throw new Error(`Unsupported Workspace tool: ${toolName}`)
  }
}

export async function openCoworkPath(project: CoworkProject, relativePath: string): Promise<void> {
  const target = await resolveProjectPath(project, relativePath)
  shell.showItemInFolder(target.absolute)
}

export const permissionSummary = (mode: CoworkApprovalMode): string => {
  if (mode === 'manual') return 'Reads run automatically; every file change asks first.'
  if (mode === 'auto')
    return 'Reads and new files run automatically; overwrites and deletion ask first.'
  return 'Reads and writes run automatically; deletion always asks first.'
}
