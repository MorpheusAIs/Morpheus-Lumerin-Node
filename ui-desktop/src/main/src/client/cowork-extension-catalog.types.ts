export type CoworkExtensionCapability =
  | 'project-read'
  | 'project-write'
  | 'artifact-create'
  | 'process-execution'
  | 'network-access'
  | 'remote-read'
  | 'remote-write'

export type CoworkExtensionIssueCode =
  | 'invalid-schema'
  | 'invalid-path'
  | 'path-changed'
  | 'symlink-blocked'
  | 'hardlink-blocked'
  | 'not-a-file'
  | 'not-a-directory'
  | 'file-too-large'
  | 'catalog-too-large'
  | 'count-limit'
  | 'duplicate-id'
  | 'invalid-encoding'
  | 'unsafe-endpoint'
  | 'read-failed'

export interface CoworkExtensionIssue {
  severity: 'warning' | 'error'
  code: CoworkExtensionIssueCode
  source: string
  message: string
}

export interface CoworkProjectInstructions {
  kind: 'project-instructions'
  content: string
  contentHash: string
  source: string
  bytes: number
  enabled: false
  trust: 'untrusted-project-content'
}

export interface CoworkSkillCatalogEntry {
  kind: 'skill'
  schemaVersion: 1
  id: string
  name: string
  description: string
  instructions: string
  instructionsHash: string
  source: string
  declaredCapabilities: CoworkExtensionCapability[]
  activationRequested: boolean
  enabled: false
  executionMode: 'instructions-only'
  trust: 'untrusted-project-content'
}

export type CoworkRemoteMcpTransport = 'streamable-http' | 'sse'
export type CoworkRemoteMcpAuthentication = 'none' | 'oauth2'

export interface CoworkConnectorCatalogEntry {
  kind: 'remote-mcp-connector'
  schemaVersion: 1
  id: string
  name: string
  description: string
  transport: CoworkRemoteMcpTransport
  url: string
  authentication: CoworkRemoteMcpAuthentication
  declaredCapabilities: CoworkExtensionCapability[]
  activationRequested: boolean
  enabled: false
  connectionState: 'not-connected'
  requiresEndpointVerification: true
  trust: 'unverified-user-configuration'
  source: string
}

export interface CoworkExtensionCatalog {
  schemaVersion: 1
  projectInstructions?: CoworkProjectInstructions
  skills: CoworkSkillCatalogEntry[]
  connectors: CoworkConnectorCatalogEntry[]
  issues: CoworkExtensionIssue[]
  executionAvailable: false
  networkAccessPerformed: false
}

export interface DiscoverCoworkExtensionCatalogInput {
  /** Canonical or user-selected project folder. Discovery never reads outside it. */
  projectRoot: string
  /**
   * Optional directory containing a fixed `connectors.json` file. Passing a directory rather
   * than an arbitrary file keeps connector discovery scoped to application-owned storage.
   */
  userConfigRoot?: string
}
