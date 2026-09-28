import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { COWORK_EXTENSION_LIMITS, discoverCoworkExtensionCatalog } from './cowork-extension-catalog'

let suiteDirectory: string
let projectRoot: string
let userConfigRoot: string

const writeJson = async (filename: string, value: unknown): Promise<void> => {
  await fs.mkdir(path.dirname(filename), { recursive: true })
  await fs.writeFile(filename, JSON.stringify(value), 'utf8')
}

const writeSkill = async (
  id: string,
  overrides: Record<string, unknown> = {},
  instructions = 'Follow the approved research checklist.'
): Promise<void> => {
  const skillDirectory = path.join(projectRoot, '.morpheus', 'cowork', 'skills', id)
  await fs.mkdir(skillDirectory, { recursive: true })
  await writeJson(path.join(skillDirectory, 'skill.json'), {
    schemaVersion: 1,
    id,
    name: 'Research checklist',
    description: 'Creates a source-grounded research memo.',
    instructionsFile: 'SKILL.md',
    capabilities: ['project-read', 'artifact-create'],
    ...overrides
  })
  await fs.writeFile(path.join(skillDirectory, 'SKILL.md'), instructions, 'utf8')
}

const connector = (
  id: string,
  overrides: Record<string, unknown> = {}
): Record<string, unknown> => ({
  schemaVersion: 1,
  id,
  name: `Connector ${id}`,
  description: 'A user-configured remote MCP service.',
  transport: 'streamable-http',
  url: `https://${id}.example.com/mcp`,
  authentication: 'oauth2',
  capabilities: ['remote-read'],
  ...overrides
})

beforeAll(async () => {
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-cowork-extensions-'))
})

beforeEach(async () => {
  projectRoot = await fs.mkdtemp(path.join(suiteDirectory, 'project-'))
  userConfigRoot = await fs.mkdtemp(path.join(suiteDirectory, 'config-'))
})

afterAll(async () => {
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

describe('discoverCoworkExtensionCatalog', () => {
  it('discovers bounded declarative instructions, skills, and remote MCP descriptors without activating them', async () => {
    const projectInstructions = 'Preserve source files and cite every factual claim.'
    await fs.mkdir(path.join(projectRoot, '.morpheus', 'cowork'), { recursive: true })
    await fs.writeFile(
      path.join(projectRoot, '.morpheus', 'cowork', 'instructions.md'),
      projectInstructions,
      'utf8'
    )
    await writeSkill('research-memo', { enabled: true })
    await writeJson(path.join(userConfigRoot, 'connectors.json'), {
      schemaVersion: 1,
      connectors: [connector('knowledge', { enabled: true })]
    })
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)

    const catalog = await discoverCoworkExtensionCatalog({ projectRoot, userConfigRoot })

    expect(catalog).toMatchObject({
      schemaVersion: 1,
      executionAvailable: false,
      networkAccessPerformed: false,
      projectInstructions: {
        content: projectInstructions,
        source: 'project:.morpheus/cowork/instructions.md',
        enabled: false,
        trust: 'untrusted-project-content'
      }
    })
    expect(catalog.projectInstructions?.contentHash).toMatch(/^[a-f0-9]{64}$/)
    expect(catalog.skills).toEqual([
      expect.objectContaining({
        id: 'research-memo',
        declaredCapabilities: ['project-read', 'artifact-create'],
        activationRequested: true,
        enabled: false,
        executionMode: 'instructions-only',
        trust: 'untrusted-project-content'
      })
    ])
    expect(catalog.connectors).toEqual([
      expect.objectContaining({
        id: 'knowledge',
        url: 'https://knowledge.example.com/mcp',
        authentication: 'oauth2',
        activationRequested: true,
        enabled: false,
        connectionState: 'not-connected',
        requiresEndpointVerification: true,
        trust: 'unverified-user-configuration'
      })
    ])
    expect(catalog.issues).toEqual([])
    expect(fetchMock).not.toHaveBeenCalled()
    vi.unstubAllGlobals()
  })

  it('strictly rejects unknown fields, invalid capability declarations, and credential-shaped connector fields', async () => {
    await writeSkill('unsafe-skill', {
      capabilities: ['project-read', 'shell-anything'],
      command: 'run-me'
    })
    await writeJson(path.join(userConfigRoot, 'connectors.json'), {
      schemaVersion: 1,
      connectors: [
        connector('unsafe', {
          headers: { Authorization: 'Bearer should-not-be-stored-here' }
        })
      ]
    })

    const catalog = await discoverCoworkExtensionCatalog({ projectRoot, userConfigRoot })

    expect(catalog.skills).toEqual([])
    expect(catalog.connectors).toEqual([])
    expect(catalog.issues).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          code: 'invalid-schema',
          source: expect.stringContaining('skill.json')
        }),
        expect.objectContaining({ code: 'invalid-schema', source: 'user:connectors.json#0' })
      ])
    )
    expect(JSON.stringify(catalog)).not.toContain('should-not-be-stored-here')
  })

  it.each(['nested/file.md:stream', 'nested/NUL.txt', 'nested/trailing.'])(
    'rejects non-portable Windows instruction path %s on every platform',
    async (instructionsFile) => {
      await writeSkill('portable-path', { instructionsFile })

      const catalog = await discoverCoworkExtensionCatalog({ projectRoot })

      expect(catalog.skills).toEqual([])
      expect(catalog.issues).toEqual([
        expect.objectContaining({
          code: 'invalid-path',
          source: expect.stringContaining('skill.json')
        })
      ])
    }
  )

  it('rejects parent traversal and symbolic-link escapes from skill instruction paths', async () => {
    await writeSkill('traversal', { instructionsFile: '../../../../outside.md' })

    const outsideFile = path.join(suiteDirectory, 'outside-instructions.md')
    await fs.writeFile(outsideFile, 'Do something outside the connected project.', 'utf8')
    const linkedSkillDirectory = path.join(projectRoot, '.morpheus', 'cowork', 'skills', 'linked')
    await fs.mkdir(linkedSkillDirectory, { recursive: true })
    await writeJson(path.join(linkedSkillDirectory, 'skill.json'), {
      schemaVersion: 1,
      id: 'linked',
      name: 'Linked instructions',
      description: '',
      instructionsFile: 'SKILL.md'
    })
    try {
      await fs.symlink(outsideFile, path.join(linkedSkillDirectory, 'SKILL.md'), 'file')
    } catch (error: any) {
      if (process.platform !== 'win32' || error?.code !== 'EPERM') throw error
    }

    const catalog = await discoverCoworkExtensionCatalog({ projectRoot })

    expect(catalog.skills).toEqual([])
    expect(catalog.issues).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: 'invalid-path' }),
        ...(process.platform === 'win32'
          ? []
          : [expect.objectContaining({ code: 'symlink-blocked' })])
      ])
    )
    expect(JSON.stringify(catalog)).not.toContain('Do something outside')
  })

  it('does not read a hard-linked catalog file that could alias data outside the project', async () => {
    const outsideFile = path.join(suiteDirectory, 'outside-hardlink.md')
    await fs.writeFile(outsideFile, 'Sensitive content outside the connected project.', 'utf8')
    const coworkDirectory = path.join(projectRoot, '.morpheus', 'cowork')
    await fs.mkdir(coworkDirectory, { recursive: true })
    await fs.link(outsideFile, path.join(coworkDirectory, 'instructions.md'))

    const catalog = await discoverCoworkExtensionCatalog({ projectRoot })

    expect(catalog.projectInstructions).toBeUndefined()
    expect(catalog.issues).toEqual([
      expect.objectContaining({
        code: 'hardlink-blocked',
        source: 'project:.morpheus/cowork/instructions.md'
      })
    ])
    expect(JSON.stringify(catalog)).not.toContain('Sensitive content outside')
  })

  it('fails closed when catalog size or entry-count limits are exceeded', async () => {
    await writeSkill('oversized', {}, 'x'.repeat(COWORK_EXTENSION_LIMITS.instructionsBytes + 1))
    await writeJson(path.join(userConfigRoot, 'connectors.json'), {
      schemaVersion: 1,
      connectors: Array.from({ length: COWORK_EXTENSION_LIMITS.connectors + 1 }, (_, index) =>
        connector(`connector-${index}`)
      )
    })

    const catalog = await discoverCoworkExtensionCatalog({ projectRoot, userConfigRoot })

    expect(catalog.skills).toEqual([])
    expect(catalog.connectors).toEqual([])
    expect(catalog.issues).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: 'file-too-large' }),
        expect.objectContaining({ code: 'count-limit', source: 'user:connectors.json' })
      ])
    )
  })

  it('does not partially load an over-limit skill directory', async () => {
    const skillsRoot = path.join(projectRoot, '.morpheus', 'cowork', 'skills')
    await Promise.all(
      Array.from({ length: COWORK_EXTENSION_LIMITS.skills + 1 }, (_, index) =>
        fs.mkdir(path.join(skillsRoot, `skill-${index}`), { recursive: true })
      )
    )

    const catalog = await discoverCoworkExtensionCatalog({ projectRoot })

    expect(catalog.skills).toEqual([])
    expect(catalog.issues).toEqual([
      expect.objectContaining({
        code: 'count-limit',
        source: 'project:.morpheus/cowork/skills'
      })
    ])
  })

  it.each([
    ['plain HTTP', { url: 'http://mcp.example.com/mcp' }],
    ['loopback', { url: 'https://127.0.0.1/mcp' }],
    ['IPv6 literal', { url: 'https://[2606:4700:4700::1111]/mcp' }],
    ['trailing-dot localhost', { url: 'https://localhost./mcp' }],
    ['private network', { url: 'https://192.168.10.20/mcp' }],
    ['local hostname', { url: 'https://service.internal/mcp' }],
    ['embedded credentials', { url: 'https://user:password@mcp.example.com/mcp' }],
    ['query credentials', { url: 'https://mcp.example.com/mcp?token=secret' }]
  ])('does not catalog an unsafe remote endpoint: %s', async (_label, overrides) => {
    await writeJson(path.join(userConfigRoot, 'connectors.json'), {
      schemaVersion: 1,
      connectors: [connector('unsafe-endpoint', overrides)]
    })

    const catalog = await discoverCoworkExtensionCatalog({ projectRoot, userConfigRoot })

    expect(catalog.connectors).toEqual([])
    expect(catalog.issues).toEqual([
      expect.objectContaining({ code: 'unsafe-endpoint', source: 'user:connectors.json#0' })
    ])
  })

  it('returns an empty catalog when the optional declarative files do not exist', async () => {
    await expect(discoverCoworkExtensionCatalog({ projectRoot, userConfigRoot })).resolves.toEqual({
      schemaVersion: 1,
      skills: [],
      connectors: [],
      issues: [],
      executionAvailable: false,
      networkAccessPerformed: false
    })
  })

  it('rejects an invalid project root before reading any catalog data', async () => {
    await expect(
      discoverCoworkExtensionCatalog({ projectRoot: path.join(suiteDirectory, 'missing-project') })
    ).rejects.toThrow(/not an accessible directory/)
  })
})
