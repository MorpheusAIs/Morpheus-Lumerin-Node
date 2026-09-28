import { createHash } from 'node:crypto'
import { constants, promises as fs } from 'node:fs'
import path from 'node:path'
import { TextDecoder } from 'node:util'
import type {
  CoworkConnectorCatalogEntry,
  CoworkExtensionCapability,
  CoworkExtensionCatalog,
  CoworkExtensionIssue,
  CoworkExtensionIssueCode,
  CoworkRemoteMcpAuthentication,
  CoworkRemoteMcpTransport,
  CoworkSkillCatalogEntry,
  DiscoverCoworkExtensionCatalogInput
} from './cowork-extension-catalog.types'

const PROJECT_INSTRUCTIONS_PATH = path.join('.morpheus', 'cowork', 'instructions.md')
const PROJECT_SKILLS_PATH = path.join('.morpheus', 'cowork', 'skills')
const CONNECTORS_FILE = 'connectors.json'

export const COWORK_EXTENSION_LIMITS = Object.freeze({
  manifestBytes: 16 * 1024,
  instructionsBytes: 64 * 1024,
  connectorConfigBytes: 128 * 1024,
  totalInstructionBytes: 512 * 1024,
  skills: 64,
  connectors: 32,
  idCharacters: 64,
  nameCharacters: 128,
  descriptionCharacters: 1024,
  relativePathCharacters: 256,
  endpointCharacters: 2048
})

const SKILL_CAPABILITIES = new Set<CoworkExtensionCapability>([
  'project-read',
  'project-write',
  'artifact-create',
  'process-execution',
  'network-access'
])

const CONNECTOR_CAPABILITIES = new Set<CoworkExtensionCapability>(['remote-read', 'remote-write'])

const ID_PATTERN = /^[a-z0-9](?:[a-z0-9._-]{0,62}[a-z0-9])?$/
const WINDOWS_DEVICE_NAME = /^(?:con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\..*)?$/i
const textDecoder = new TextDecoder('utf-8', { fatal: true })

class CatalogSourceError extends Error {
  constructor(
    readonly code: CoworkExtensionIssueCode,
    message: string
  ) {
    super(message)
    this.name = 'CatalogSourceError'
  }
}

export class CoworkExtensionCatalogError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'CoworkExtensionCatalogError'
  }
}

interface BoundedTextFile {
  content: string
  bytes: number
}

interface SkillManifest {
  schemaVersion: 1
  id: string
  name: string
  description: string
  instructionsFile: string
  capabilities: CoworkExtensionCapability[]
  enabled: boolean
}

interface ConnectorManifest {
  schemaVersion: 1
  id: string
  name: string
  description: string
  transport: CoworkRemoteMcpTransport
  url: string
  authentication: CoworkRemoteMcpAuthentication
  capabilities: CoworkExtensionCapability[]
  enabled: boolean
}

const sourcePath = (scope: 'project' | 'user', relativePath: string): string => {
  const safeRelativePath = relativePath
    .split(path.sep)
    .join('/')
    .replace(/[\u0000-\u001f\u007f]/g, '�')
    .slice(0, 1024)
  return `${scope}:${safeRelativePath}`
}

const hashText = (content: string): string =>
  createHash('sha256').update(content, 'utf8').digest('hex')

const errorMessage = (error: unknown): string => {
  if (error instanceof CatalogSourceError || error instanceof CoworkExtensionCatalogError) {
    return error.message
  }
  return 'Catalog source could not be read safely'
}

const issueFromError = (
  source: string,
  error: unknown,
  fallbackCode: CoworkExtensionIssueCode = 'read-failed'
): CoworkExtensionIssue => ({
  severity: 'error',
  code: error instanceof CatalogSourceError ? error.code : fallbackCode,
  source,
  message: errorMessage(error)
})

const isMissing = (error: unknown): boolean =>
  typeof error === 'object' && error !== null && 'code' in error && error.code === 'ENOENT'

const isPlainObject = (value: unknown): value is Record<string, unknown> => {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) return false
  const prototype = Object.getPrototypeOf(value)
  return prototype === Object.prototype || prototype === null
}

const requireObject = (value: unknown, label: string): Record<string, unknown> => {
  if (!isPlainObject(value))
    throw new CatalogSourceError('invalid-schema', `${label} must be an object`)
  return value
}

const requireExactKeys = (
  object: Record<string, unknown>,
  allowed: readonly string[],
  required: readonly string[],
  label: string
): void => {
  const allowedSet = new Set(allowed)
  const unknownKeys = Object.keys(object).filter((key) => !allowedSet.has(key))
  if (unknownKeys.length > 0) {
    throw new CatalogSourceError(
      'invalid-schema',
      `${label} contains unsupported field${unknownKeys.length === 1 ? '' : 's'}: ${unknownKeys.join(', ')}`
    )
  }

  const missingKeys = required.filter((key) => !(key in object))
  if (missingKeys.length > 0) {
    throw new CatalogSourceError(
      'invalid-schema',
      `${label} is missing required field${missingKeys.length === 1 ? '' : 's'}: ${missingKeys.join(', ')}`
    )
  }
}

const requireSchemaVersion = (value: unknown, label: string): 1 => {
  if (value !== 1)
    throw new CatalogSourceError('invalid-schema', `${label} schemaVersion must be 1`)
  return 1
}

const requireString = (
  value: unknown,
  label: string,
  maximumCharacters: number,
  options: { allowEmpty?: boolean } = {}
): string => {
  if (typeof value !== 'string') {
    throw new CatalogSourceError('invalid-schema', `${label} must be a string`)
  }
  const normalized = value.trim()
  if (!options.allowEmpty && normalized.length === 0) {
    throw new CatalogSourceError('invalid-schema', `${label} cannot be empty`)
  }
  if (normalized.length > maximumCharacters) {
    throw new CatalogSourceError(
      'invalid-schema',
      `${label} exceeds the ${maximumCharacters}-character limit`
    )
  }
  if (/\p{Cc}/u.test(normalized)) {
    throw new CatalogSourceError('invalid-schema', `${label} cannot contain control characters`)
  }
  return normalized
}

const requireId = (value: unknown, label: string): string => {
  const id = requireString(value, label, COWORK_EXTENSION_LIMITS.idCharacters)
  if (!ID_PATTERN.test(id)) {
    throw new CatalogSourceError(
      'invalid-schema',
      `${label} must use lowercase letters, numbers, dots, underscores, or hyphens`
    )
  }
  if (WINDOWS_DEVICE_NAME.test(id)) {
    throw new CatalogSourceError('invalid-schema', `${label} uses a reserved device name`)
  }
  return id
}

const optionalBoolean = (value: unknown, label: string): boolean => {
  if (value === undefined) return false
  if (typeof value !== 'boolean') {
    throw new CatalogSourceError('invalid-schema', `${label} must be a boolean`)
  }
  return value
}

const requireCapabilities = (
  value: unknown,
  label: string,
  allowed: ReadonlySet<CoworkExtensionCapability>
): CoworkExtensionCapability[] => {
  if (value === undefined) return []
  if (!Array.isArray(value)) {
    throw new CatalogSourceError('invalid-schema', `${label} must be an array`)
  }
  if (value.length > allowed.size) {
    throw new CatalogSourceError('invalid-schema', `${label} contains too many entries`)
  }
  const result: CoworkExtensionCapability[] = []
  for (const item of value) {
    if (typeof item !== 'string' || !allowed.has(item as CoworkExtensionCapability)) {
      throw new CatalogSourceError('invalid-schema', `${label} contains an unsupported capability`)
    }
    const capability = item as CoworkExtensionCapability
    if (result.includes(capability)) {
      throw new CatalogSourceError(
        'invalid-schema',
        `${label} contains duplicate capability ${capability}`
      )
    }
    result.push(capability)
  }
  return result
}

const requireConfigRelativePath = (value: unknown, label: string): string => {
  const configuredPath = requireString(value, label, COWORK_EXTENSION_LIMITS.relativePathCharacters)
  if (
    configuredPath.includes('\\') ||
    configuredPath.includes('\0') ||
    path.posix.isAbsolute(configuredPath) ||
    /^[a-zA-Z]:/.test(configuredPath)
  ) {
    throw new CatalogSourceError('invalid-path', `${label} must be a portable relative path`)
  }
  const segments = configuredPath.split('/')
  if (
    segments.some(
      (segment) =>
        segment.length === 0 ||
        segment === '.' ||
        segment === '..' ||
        segment.includes(':') ||
        /[. ]$/.test(segment) ||
        WINDOWS_DEVICE_NAME.test(segment)
    )
  ) {
    throw new CatalogSourceError(
      'invalid-path',
      `${label} contains a non-portable or reserved path segment`
    )
  }
  return segments.join(path.sep)
}

const isWithin = (root: string, target: string): boolean => {
  const relative = path.relative(root, target)
  return (
    relative === '' ||
    (!relative.startsWith(`..${path.sep}`) && relative !== '..' && !path.isAbsolute(relative))
  )
}

const canonicalDirectory = async (directory: string, label: string): Promise<string> => {
  if (typeof directory !== 'string' || directory.trim().length === 0 || directory.includes('\0')) {
    throw new CoworkExtensionCatalogError(`${label} must be a valid directory path`)
  }
  try {
    const canonical = await fs.realpath(directory)
    const stat = await fs.stat(canonical)
    if (!stat.isDirectory()) throw new CoworkExtensionCatalogError(`${label} must be a directory`)
    return canonical
  } catch (error) {
    if (error instanceof CoworkExtensionCatalogError) throw error
    throw new CoworkExtensionCatalogError(`${label} is not an accessible directory`)
  }
}

const verifyNoSymlinkPath = async (
  root: string,
  relativePath: string,
  expected: 'file' | 'directory',
  optional: boolean
): Promise<string | undefined> => {
  const candidate = path.resolve(root, relativePath)
  if (!isWithin(root, candidate)) {
    throw new CatalogSourceError('invalid-path', 'Catalog path escapes its configured root')
  }

  const relative = path.relative(root, candidate)
  let current = root
  for (const segment of relative.split(path.sep).filter(Boolean)) {
    current = path.join(current, segment)
    let stat
    try {
      stat = await fs.lstat(current)
    } catch (error) {
      if (optional && isMissing(error)) return undefined
      throw error
    }
    if (stat.isSymbolicLink()) {
      throw new CatalogSourceError(
        'symlink-blocked',
        'Symbolic links are not allowed in extension paths'
      )
    }
  }

  const canonical = await fs.realpath(candidate)
  if (!isWithin(root, canonical)) {
    throw new CatalogSourceError(
      'invalid-path',
      'Catalog path resolves outside its configured root'
    )
  }
  const finalStat = await fs.stat(canonical)
  if (expected === 'file' && !finalStat.isFile()) {
    throw new CatalogSourceError('not-a-file', 'Catalog source must be a regular file')
  }
  if (expected === 'directory' && !finalStat.isDirectory()) {
    throw new CatalogSourceError('not-a-directory', 'Catalog source must be a directory')
  }
  return canonical
}

const readBoundedTextFile = async (
  root: string,
  relativePath: string,
  maximumBytes: number,
  optional = true
): Promise<BoundedTextFile | undefined> => {
  const filename = await verifyNoSymlinkPath(root, relativePath, 'file', optional)
  if (!filename) return undefined

  let handle
  try {
    // O_NOFOLLOW protects the final component if it is replaced between validation and open.
    const noFollowFlag = process.platform === 'win32' ? 0 : (constants.O_NOFOLLOW ?? 0)
    try {
      handle = await fs.open(filename, constants.O_RDONLY | noFollowFlag)
    } catch (error) {
      if (
        typeof error === 'object' &&
        error !== null &&
        'code' in error &&
        error.code === 'ELOOP'
      ) {
        throw new CatalogSourceError(
          'symlink-blocked',
          'Symbolic links are not allowed in extension paths'
        )
      }
      throw error
    }
    const stat = await handle.stat()
    if (!stat.isFile())
      throw new CatalogSourceError('not-a-file', 'Catalog source must be a regular file')
    if (stat.nlink > 1) {
      throw new CatalogSourceError('hardlink-blocked', 'Hard-linked extension files are not loaded')
    }

    // Re-check after opening and compare the path identity to the open handle. This catches
    // parent junction/symlink swaps on Windows and narrows the path-race window on Unix.
    const pathStat = await fs.lstat(filename)
    if (pathStat.isSymbolicLink()) {
      throw new CatalogSourceError(
        'symlink-blocked',
        'Symbolic links are not allowed in extension paths'
      )
    }
    const postOpenCanonical = await fs.realpath(filename)
    if (!isWithin(root, postOpenCanonical)) {
      throw new CatalogSourceError(
        'invalid-path',
        'Catalog path resolves outside its configured root'
      )
    }
    if (pathStat.dev !== stat.dev || pathStat.ino !== stat.ino) {
      throw new CatalogSourceError(
        'path-changed',
        'Catalog source changed while it was being inspected'
      )
    }
    if (stat.size > maximumBytes) {
      throw new CatalogSourceError(
        'file-too-large',
        `Catalog source exceeds the ${maximumBytes}-byte limit`
      )
    }
    const boundedBuffer = Buffer.alloc(maximumBytes + 1)
    let bytesRead = 0
    while (bytesRead < boundedBuffer.byteLength) {
      const result = await handle.read(
        boundedBuffer,
        bytesRead,
        boundedBuffer.byteLength - bytesRead,
        bytesRead
      )
      if (result.bytesRead === 0) break
      bytesRead += result.bytesRead
    }
    if (bytesRead > maximumBytes) {
      throw new CatalogSourceError(
        'file-too-large',
        `Catalog source exceeds the ${maximumBytes}-byte limit`
      )
    }
    const bytes = boundedBuffer.subarray(0, bytesRead)
    let content: string
    try {
      content = textDecoder.decode(bytes)
    } catch {
      throw new CatalogSourceError(
        'invalid-encoding',
        'Catalog source must contain valid UTF-8 text'
      )
    }
    if (content.includes('\0')) {
      throw new CatalogSourceError('invalid-encoding', 'Catalog source cannot contain NUL bytes')
    }
    return { content, bytes: bytes.byteLength }
  } finally {
    await handle?.close()
  }
}

const parseJson = (content: string, label: string): unknown => {
  try {
    return JSON.parse(content)
  } catch {
    throw new CatalogSourceError('invalid-schema', `${label} must contain valid JSON`)
  }
}

const parseSkillManifest = (value: unknown, directoryName: string): SkillManifest => {
  const object = requireObject(value, 'Skill manifest')
  requireExactKeys(
    object,
    ['schemaVersion', 'id', 'name', 'description', 'instructionsFile', 'capabilities', 'enabled'],
    ['schemaVersion', 'id', 'name', 'description', 'instructionsFile'],
    'Skill manifest'
  )
  const id = requireId(object.id, 'Skill id')
  if (id !== directoryName) {
    throw new CatalogSourceError(
      'invalid-schema',
      'Skill id must match its containing directory name'
    )
  }
  return {
    schemaVersion: requireSchemaVersion(object.schemaVersion, 'Skill manifest'),
    id,
    name: requireString(object.name, 'Skill name', COWORK_EXTENSION_LIMITS.nameCharacters),
    description: requireString(
      object.description,
      'Skill description',
      COWORK_EXTENSION_LIMITS.descriptionCharacters,
      { allowEmpty: true }
    ),
    instructionsFile: requireConfigRelativePath(object.instructionsFile, 'Skill instructionsFile'),
    capabilities: requireCapabilities(
      object.capabilities,
      'Skill capabilities',
      SKILL_CAPABILITIES
    ),
    enabled: optionalBoolean(object.enabled, 'Skill enabled')
  }
}

const requireEnum = <T extends string>(value: unknown, allowed: readonly T[], label: string): T => {
  if (typeof value !== 'string' || !allowed.includes(value as T)) {
    throw new CatalogSourceError('invalid-schema', `${label} has an unsupported value`)
  }
  return value as T
}

const requireSafeRemoteEndpoint = (value: unknown): string => {
  const configuredUrl = requireString(
    value,
    'Connector url',
    COWORK_EXTENSION_LIMITS.endpointCharacters
  )
  let endpoint: URL
  try {
    endpoint = new URL(configuredUrl)
  } catch {
    throw new CatalogSourceError('unsafe-endpoint', 'Connector url must be a valid HTTPS URL')
  }
  if (
    endpoint.protocol !== 'https:' ||
    endpoint.username.length > 0 ||
    endpoint.password.length > 0 ||
    endpoint.search.length > 0 ||
    endpoint.hash.length > 0
  ) {
    throw new CatalogSourceError(
      'unsafe-endpoint',
      'Connector url must use HTTPS and cannot contain credentials, a query, or a fragment'
    )
  }

  const hostname = endpoint.hostname
    .toLowerCase()
    .replace(/^\[|\]$/g, '')
    .replace(/\.+$/, '')
  const localName =
    !hostname.includes('.') ||
    hostname === 'localhost' ||
    hostname.endsWith('.localhost') ||
    hostname.endsWith('.local') ||
    hostname.endsWith('.internal') ||
    hostname.endsWith('.lan') ||
    hostname.endsWith('.home.arpa')
  // Bare IP endpoints make SSRF controls and TLS identity harder to enforce. Activation can add
  // DNS resolution plus redirect checks; discovery accepts only a DNS name and still stays off.
  const ipLiteral = /^\d{1,3}(?:\.\d{1,3}){3}$/.test(hostname) || hostname.includes(':')
  if (localName || ipLiteral) {
    throw new CatalogSourceError(
      'unsafe-endpoint',
      'Connector url must identify a public remote endpoint'
    )
  }
  return endpoint.toString()
}

const parseConnectorManifest = (value: unknown): ConnectorManifest => {
  const object = requireObject(value, 'Connector descriptor')
  requireExactKeys(
    object,
    [
      'schemaVersion',
      'id',
      'name',
      'description',
      'transport',
      'url',
      'authentication',
      'capabilities',
      'enabled'
    ],
    ['schemaVersion', 'id', 'name', 'description', 'transport', 'url', 'authentication'],
    'Connector descriptor'
  )
  return {
    schemaVersion: requireSchemaVersion(object.schemaVersion, 'Connector descriptor'),
    id: requireId(object.id, 'Connector id'),
    name: requireString(object.name, 'Connector name', COWORK_EXTENSION_LIMITS.nameCharacters),
    description: requireString(
      object.description,
      'Connector description',
      COWORK_EXTENSION_LIMITS.descriptionCharacters,
      { allowEmpty: true }
    ),
    transport: requireEnum(
      object.transport,
      ['streamable-http', 'sse'] as const,
      'Connector transport'
    ),
    url: requireSafeRemoteEndpoint(object.url),
    authentication: requireEnum(
      object.authentication,
      ['none', 'oauth2'] as const,
      'Connector authentication'
    ),
    capabilities: requireCapabilities(
      object.capabilities,
      'Connector capabilities',
      CONNECTOR_CAPABILITIES
    ),
    enabled: optionalBoolean(object.enabled, 'Connector enabled')
  }
}

const collectDirectoryNames = async (
  directory: string,
  maximumEntries: number
): Promise<{ names?: string[]; exceeded: boolean; symlinks: string[] }> => {
  const names: string[] = []
  const symlinks: string[] = []
  let scannedEntries = 0
  const handle = await fs.opendir(directory)
  try {
    for await (const entry of handle) {
      scannedEntries += 1
      if (entry.isSymbolicLink()) symlinks.push(entry.name)
      else if (entry.isDirectory()) names.push(entry.name)
      if (scannedEntries > maximumEntries) {
        return { exceeded: true, symlinks }
      }
    }
  } finally {
    await handle.close().catch(() => undefined)
  }
  return { names: names.sort((a, b) => a.localeCompare(b)), exceeded: false, symlinks }
}

const discoverSkills = async (
  projectRoot: string,
  issues: CoworkExtensionIssue[],
  initialInstructionBytes: number
): Promise<CoworkSkillCatalogEntry[]> => {
  const skillsSource = sourcePath('project', PROJECT_SKILLS_PATH)
  let skillsDirectory: string | undefined
  try {
    skillsDirectory = await verifyNoSymlinkPath(projectRoot, PROJECT_SKILLS_PATH, 'directory', true)
  } catch (error) {
    issues.push(issueFromError(skillsSource, error))
    return []
  }
  if (!skillsDirectory) return []

  let directoryEntries
  try {
    directoryEntries = await collectDirectoryNames(skillsDirectory, COWORK_EXTENSION_LIMITS.skills)
  } catch (error) {
    issues.push(issueFromError(skillsSource, error))
    return []
  }
  for (const symlink of directoryEntries.symlinks) {
    issues.push({
      severity: 'error',
      code: 'symlink-blocked',
      source: sourcePath('project', path.join(PROJECT_SKILLS_PATH, symlink)),
      message: 'Symbolic-link skill directories are not loaded'
    })
  }
  if (directoryEntries.exceeded || !directoryEntries.names) {
    issues.push({
      severity: 'error',
      code: 'count-limit',
      source: skillsSource,
      message: `Skill directory exceeds the ${COWORK_EXTENSION_LIMITS.skills}-entry limit; no skills were loaded`
    })
    return []
  }

  const discovered: CoworkSkillCatalogEntry[] = []
  let totalInstructionBytes = initialInstructionBytes
  for (const directoryName of directoryEntries.names) {
    const manifestRelative = path.join(PROJECT_SKILLS_PATH, directoryName, 'skill.json')
    const manifestSource = sourcePath('project', manifestRelative)
    try {
      requireId(directoryName, 'Skill directory name')
      const manifestFile = await readBoundedTextFile(
        projectRoot,
        manifestRelative,
        COWORK_EXTENSION_LIMITS.manifestBytes,
        false
      )
      const manifest = parseSkillManifest(
        parseJson(manifestFile!.content, 'Skill manifest'),
        directoryName
      )
      const instructionRelative = path.join(
        PROJECT_SKILLS_PATH,
        directoryName,
        manifest.instructionsFile
      )
      const instructions = await readBoundedTextFile(
        projectRoot,
        instructionRelative,
        COWORK_EXTENSION_LIMITS.instructionsBytes,
        false
      )
      if (
        totalInstructionBytes + instructions!.bytes >
        COWORK_EXTENSION_LIMITS.totalInstructionBytes
      ) {
        issues.push({
          severity: 'error',
          code: 'catalog-too-large',
          source: sourcePath('project', instructionRelative),
          message: `Combined project instructions exceed the ${COWORK_EXTENSION_LIMITS.totalInstructionBytes}-byte catalog limit`
        })
        continue
      }
      totalInstructionBytes += instructions!.bytes
      discovered.push({
        kind: 'skill',
        schemaVersion: 1,
        id: manifest.id,
        name: manifest.name,
        description: manifest.description,
        instructions: instructions!.content,
        instructionsHash: hashText(instructions!.content),
        source: manifestSource,
        declaredCapabilities: manifest.capabilities,
        activationRequested: manifest.enabled,
        enabled: false,
        executionMode: 'instructions-only',
        trust: 'untrusted-project-content'
      })
    } catch (error) {
      issues.push(issueFromError(manifestSource, error))
    }
  }

  const duplicateIds = new Set(
    discovered.map((entry) => entry.id).filter((id, index, ids) => ids.indexOf(id) !== index)
  )
  for (const id of duplicateIds) {
    issues.push({
      severity: 'error',
      code: 'duplicate-id',
      source: skillsSource,
      message: `Duplicate skill id ${id} was not loaded`
    })
  }
  return discovered.filter((entry) => !duplicateIds.has(entry.id))
}

const discoverConnectors = async (
  userConfigRoot: string | undefined,
  issues: CoworkExtensionIssue[]
): Promise<CoworkConnectorCatalogEntry[]> => {
  if (!userConfigRoot) return []
  const configSource = sourcePath('user', CONNECTORS_FILE)
  let canonicalRoot: string
  try {
    canonicalRoot = await canonicalDirectory(
      userConfigRoot,
      'Workspace extension configuration folder'
    )
  } catch (error) {
    issues.push(issueFromError(configSource, error))
    return []
  }

  let configFile: BoundedTextFile | undefined
  try {
    configFile = await readBoundedTextFile(
      canonicalRoot,
      CONNECTORS_FILE,
      COWORK_EXTENSION_LIMITS.connectorConfigBytes,
      true
    )
  } catch (error) {
    issues.push(issueFromError(configSource, error))
    return []
  }
  if (!configFile) return []

  let connectorValues: unknown[]
  try {
    const root = requireObject(
      parseJson(configFile.content, 'Connector configuration'),
      'Connector configuration'
    )
    requireExactKeys(
      root,
      ['schemaVersion', 'connectors'],
      ['schemaVersion', 'connectors'],
      'Connector configuration'
    )
    requireSchemaVersion(root.schemaVersion, 'Connector configuration')
    if (!Array.isArray(root.connectors)) {
      throw new CatalogSourceError(
        'invalid-schema',
        'Connector configuration connectors must be an array'
      )
    }
    if (root.connectors.length > COWORK_EXTENSION_LIMITS.connectors) {
      throw new CatalogSourceError(
        'count-limit',
        `Connector configuration exceeds the ${COWORK_EXTENSION_LIMITS.connectors}-connector limit`
      )
    }
    connectorValues = root.connectors
  } catch (error) {
    issues.push(issueFromError(configSource, error))
    return []
  }

  const discovered: CoworkConnectorCatalogEntry[] = []
  connectorValues.forEach((value, index) => {
    const descriptorSource = `${configSource}#${index}`
    try {
      const connector = parseConnectorManifest(value)
      discovered.push({
        kind: 'remote-mcp-connector',
        schemaVersion: 1,
        id: connector.id,
        name: connector.name,
        description: connector.description,
        transport: connector.transport,
        url: connector.url,
        authentication: connector.authentication,
        declaredCapabilities: connector.capabilities,
        activationRequested: connector.enabled,
        enabled: false,
        connectionState: 'not-connected',
        requiresEndpointVerification: true,
        trust: 'unverified-user-configuration',
        source: descriptorSource
      })
    } catch (error) {
      issues.push(issueFromError(descriptorSource, error))
    }
  })

  const duplicateIds = new Set(
    discovered.map((entry) => entry.id).filter((id, index, ids) => ids.indexOf(id) !== index)
  )
  for (const id of duplicateIds) {
    issues.push({
      severity: 'error',
      code: 'duplicate-id',
      source: configSource,
      message: `Duplicate connector id ${id} was not loaded`
    })
  }
  return discovered.filter((entry) => !duplicateIds.has(entry.id))
}

/**
 * Reads declarative extension metadata only. The catalog deliberately has no activation,
 * process-execution, credential, IPC, or network API; every discovered capability remains off.
 */
export const discoverCoworkExtensionCatalog = async (
  input: DiscoverCoworkExtensionCatalogInput
): Promise<CoworkExtensionCatalog> => {
  const projectRoot = await canonicalDirectory(input.projectRoot, 'Workspace project folder')
  const issues: CoworkExtensionIssue[] = []

  let projectInstructions: CoworkExtensionCatalog['projectInstructions']
  try {
    const instructions = await readBoundedTextFile(
      projectRoot,
      PROJECT_INSTRUCTIONS_PATH,
      COWORK_EXTENSION_LIMITS.instructionsBytes,
      true
    )
    if (instructions) {
      projectInstructions = {
        kind: 'project-instructions',
        content: instructions.content,
        contentHash: hashText(instructions.content),
        source: sourcePath('project', PROJECT_INSTRUCTIONS_PATH),
        bytes: instructions.bytes,
        enabled: false,
        trust: 'untrusted-project-content'
      }
    }
  } catch (error) {
    issues.push(issueFromError(sourcePath('project', PROJECT_INSTRUCTIONS_PATH), error))
  }

  const [skills, connectors] = await Promise.all([
    discoverSkills(projectRoot, issues, projectInstructions?.bytes ?? 0),
    discoverConnectors(input.userConfigRoot, issues)
  ])

  issues.sort((left, right) =>
    `${left.source}\0${left.code}`.localeCompare(`${right.source}\0${right.code}`)
  )

  return {
    schemaVersion: 1,
    ...(projectInstructions ? { projectInstructions } : {}),
    skills,
    connectors,
    issues,
    executionAvailable: false,
    networkAccessPerformed: false
  }
}
