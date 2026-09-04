import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { CoworkApprovalMode, CoworkProject } from './cowork.types'

const electron = vi.hoisted(() => ({
  userData: '',
  trashItem: vi.fn(async (_target: string) => undefined),
  showItemInFolder: vi.fn((_target: string) => undefined)
}))

vi.mock('electron', () => ({
  app: {
    getPath: (name: string) => {
      if (name !== 'userData') throw new Error(`Unexpected Electron path request: ${name}`)
      return electron.userData
    }
  },
  shell: {
    trashItem: electron.trashItem,
    showItemInFolder: electron.showItemInFolder
  }
}))

const {
  approvalRequirement,
  executeCoworkTool,
  isCoworkMutationTool,
  resolveProjectPath,
  toolArguments,
  withCoworkMutationLock
} = await import('./cowork-tools')

let suiteDirectory: string
let projectRoot: string
let outsideRoot: string

const project = (approvalMode: CoworkApprovalMode = 'manual'): CoworkProject => ({
  schemaVersion: 1,
  id: 'project-under-test',
  name: 'Test project',
  rootPath: projectRoot,
  instructions: '',
  approvalMode,
  createdAt: 1,
  updatedAt: 1
})

beforeAll(async () => {
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-cowork-tools-'))
  electron.userData = path.join(suiteDirectory, 'user-data')
})

beforeEach(async () => {
  projectRoot = await fs.mkdtemp(path.join(suiteDirectory, 'project-'))
  outsideRoot = await fs.mkdtemp(path.join(suiteDirectory, 'outside-'))
  electron.trashItem.mockClear()
  electron.showItemInFolder.mockClear()
})

afterAll(async () => {
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

describe('resolveProjectPath', () => {
  it('resolves ordinary relative paths inside the canonical project root', async () => {
    await fs.mkdir(path.join(projectRoot, 'notes'))
    await fs.writeFile(path.join(projectRoot, 'notes', 'todo.md'), 'ship it')

    const result = await resolveProjectPath(project(), 'notes/todo.md')

    expect(result.root).toBe(await fs.realpath(projectRoot))
    expect(result.absolute).toBe(path.join(await fs.realpath(projectRoot), 'notes', 'todo.md'))
    expect(result.relative).toBe(path.join('notes', 'todo.md'))
  })

  it('rejects absolute paths and parent traversal', async () => {
    await expect(
      resolveProjectPath(project(), path.join(outsideRoot, 'secret.txt'), true)
    ).rejects.toThrow(/relative to the connected project folder/)
    await expect(resolveProjectPath(project(), '../outside/secret.txt', true)).rejects.toThrow(
      /escapes the connected project folder/
    )
  })

  it.each<[string, RegExp]>([
    ['CON', /reserved system name/],
    ['nested/LPT9.txt', /reserved system name/],
    ['nested/report.', /end with a dot or space/],
    ['nested/report /draft.txt', /end with a dot or space/],
    ['wallet.key:$DATA', /colon characters/],
    ['..\\outside\\secret.txt', /escapes the connected project folder/]
  ])('rejects the non-portable path alias %s', async (requestedPath, errorPattern) => {
    await expect(resolveProjectPath(project(), requestedPath, true)).rejects.toThrow(errorPattern)
  })

  it('rejects a symlink escape, including when the final path does not exist yet', async () => {
    const linkedDirectory = path.join(projectRoot, 'linked-outside')
    try {
      await fs.symlink(
        outsideRoot,
        linkedDirectory,
        process.platform === 'win32' ? 'junction' : 'dir'
      )
    } catch (error: any) {
      // Creating symlinks can be disabled for unprivileged Windows CI users.
      if (process.platform === 'win32' && error?.code === 'EPERM') return
      throw error
    }

    await expect(
      resolveProjectPath(project(), 'linked-outside/new-file.txt', true)
    ).rejects.toThrow(/outside the connected folder/)
  })

  it('does not allow a harmless alias to reveal a sensitive in-project path', async () => {
    const secret = path.join(projectRoot, '.env')
    await fs.writeFile(secret, 'TOKEN=secret')
    try {
      await fs.symlink(secret, path.join(projectRoot, 'safe-looking.txt'), 'file')
    } catch (error: any) {
      if (process.platform === 'win32' && error?.code === 'EPERM') return
      throw error
    }

    await expect(resolveProjectPath(project(), 'safe-looking.txt')).rejects.toThrow(
      /Credential and key-material paths are blocked/
    )
  })

  it('blocks hard-linked aliases that could hide sensitive content', async () => {
    const secret = path.join(projectRoot, '.env')
    await fs.writeFile(secret, 'TOKEN=secret')
    await fs.link(secret, path.join(projectRoot, 'safe-looking.txt'))

    await expect(resolveProjectPath(project(), 'safe-looking.txt')).rejects.toThrow(
      /Hard-linked files are blocked/
    )
  })

  it.each([
    '.env',
    'nested/.env.production',
    '.git/config',
    'nested/CREDENTIALS.JSON',
    'nested/.cookie',
    'config/models-config.json',
    'config/wallets.json',
    'secrets/.password-store/item',
    'keychains/login.keychain-db',
    'keys/private_key',
    'keys/seed.txt',
    'keys/id_rsa',
    'keys/wallet.pem',
    'keys/wallet.key',
    'keys/wallet.p12',
    'keys/wallet.pfx'
  ])('blocks credential or key-material path %s', async (requestedPath) => {
    await expect(resolveProjectPath(project(), requestedPath, true)).rejects.toThrow(
      /Credential and key-material paths are blocked/
    )
  })

  it('keeps extension guidance outside the model file-tool boundary', async () => {
    await expect(
      resolveProjectPath(project(), '.morpheus/cowork/instructions.md', true)
    ).rejects.toThrow(/user-facing project controls/)
  })
})

describe('approvalRequirement', () => {
  it('requires approval for every mutation in manual mode', async () => {
    await expect(
      approvalRequirement(project('manual'), 'write_file', { path: 'new.txt' })
    ).resolves.toMatchObject({
      risk: 'write'
    })
    await expect(
      approvalRequirement(project('manual'), 'make_directory', { path: 'new-dir' })
    ).resolves.toMatchObject({
      risk: 'write'
    })
    await expect(
      approvalRequirement(project('manual'), 'create_pdf', { path: 'report.pdf' })
    ).resolves.toMatchObject({ risk: 'write' })
  })

  it('allows new writes in auto mode but gates overwrites and moves', async () => {
    await fs.writeFile(path.join(projectRoot, 'existing.txt'), 'before')

    await expect(
      approvalRequirement(project('auto'), 'write_file', { path: 'new.txt' })
    ).resolves.toBeNull()
    await expect(
      approvalRequirement(project('auto'), 'write_file', { path: 'existing.txt' })
    ).resolves.toMatchObject({ risk: 'overwrite' })
    await expect(
      approvalRequirement(project('auto'), 'move_file', { source: 'a.txt', destination: 'b.txt' })
    ).resolves.toMatchObject({ risk: 'write' })
  })

  it('does not prompt for writes or overwrites in skip mode', async () => {
    await fs.writeFile(path.join(projectRoot, 'existing.txt'), 'before')

    await expect(
      approvalRequirement(project('skip'), 'write_file', { path: 'new.txt' })
    ).resolves.toBeNull()
    await expect(
      approvalRequirement(project('skip'), 'write_file', { path: 'existing.txt' })
    ).resolves.toBeNull()
  })

  it.each<CoworkApprovalMode>(['manual', 'auto', 'skip'])(
    'always gates deletion in %s mode',
    async (mode) => {
      await expect(
        approvalRequirement(project(mode), 'delete_file', { path: 'old.txt' })
      ).resolves.toMatchObject({
        risk: 'delete'
      })
    }
  )

  it('never gates read-only tools', async () => {
    await expect(
      approvalRequirement(project(), 'read_file', { path: 'notes.txt' })
    ).resolves.toBeNull()
    await expect(
      approvalRequirement(project(), 'search_files', { query: 'needle' })
    ).resolves.toBeNull()
  })
})

describe('executeCoworkTool', () => {
  it('writes only within the project and returns a relative artifact', async () => {
    const output = await executeCoworkTool(project(), 'write_file', {
      path: 'reports/result.md',
      content: '# Result\nDone.'
    })

    await expect(fs.readFile(path.join(projectRoot, 'reports', 'result.md'), 'utf8')).resolves.toBe(
      '# Result\nDone.'
    )
    expect(output.result).toEqual({ path: path.join('reports', 'result.md'), bytes: 14 })
    expect(output.artifact).toMatchObject({
      path: path.join('reports', 'result.md'),
      name: 'result.md',
      kind: 'file'
    })
  })

  it('backs up an existing file before replacing it', async () => {
    const target = path.join(projectRoot, 'result.txt')
    await fs.writeFile(target, 'old value')

    await executeCoworkTool(
      project(),
      'write_file',
      { path: 'result.txt', content: 'new value' },
      { allowOverwrite: true }
    )

    await expect(fs.readFile(target, 'utf8')).resolves.toBe('new value')
    const backupDirectory = path.join(electron.userData, 'CoworkBackups', project().id)
    const backups = await fs.readdir(backupDirectory)
    expect(backups).toHaveLength(1)
    await expect(fs.readFile(path.join(backupDirectory, backups[0]), 'utf8')).resolves.toBe(
      'old value'
    )
  })

  it('creates a bounded professional binary artifact inside the project', async () => {
    const output = await executeCoworkTool(project(), 'create_pdf', {
      path: 'reports/operating-review.pdf',
      title: 'Operating Review',
      blocks: [
        { type: 'heading', level: 1, text: 'Summary' },
        { type: 'paragraph', text: 'The requested report was generated locally.' }
      ]
    })

    const bytes = await fs.readFile(path.join(projectRoot, 'reports', 'operating-review.pdf'))
    expect(bytes.subarray(0, 5).toString('ascii')).toBe('%PDF-')
    expect(output.result).toMatchObject({
      path: path.join('reports', 'operating-review.pdf'),
      format: 'pdf',
      mimeType: 'application/pdf'
    })
    expect(output.artifact).toMatchObject({
      path: path.join('reports', 'operating-review.pdf'),
      kind: 'file'
    })

    const extracted = await executeCoworkTool(project(), 'read_document', {
      path: 'reports/operating-review.pdf'
    })
    expect(extracted.result).toMatchObject({
      path: path.join('reports', 'operating-review.pdf'),
      format: 'pdf',
      empty: false,
      metadata: { pageCount: 1 }
    })
    expect((extracted.result as any).text).toContain('Operating Review')
  })

  it('rejects a professional artifact path with the wrong extension', async () => {
    await expect(
      executeCoworkTool(project(), 'create_docx', {
        path: 'report.pdf',
        title: 'Wrong extension',
        blocks: [{ type: 'paragraph', text: 'Body' }]
      })
    ).rejects.toThrow(/requires a \.docx output path/)
  })

  it('does not silently overwrite a destination created after approval classification', async () => {
    const autoProject = project('auto')
    const input = { path: 'race-result.txt', content: 'agent content' }

    await expect(approvalRequirement(autoProject, 'write_file', input)).resolves.toBeNull()
    await fs.writeFile(path.join(projectRoot, input.path), 'other process content')

    await expect(
      executeCoworkTool(autoProject, 'write_file', input, { allowOverwrite: false })
    ).rejects.toThrow(/destination now exists/i)
    await expect(fs.readFile(path.join(projectRoot, input.path), 'utf8')).resolves.toBe(
      'other process content'
    )
  })

  it('serializes same-project mutation classification and execution', async () => {
    const autoProject = project('auto')
    const input = { path: 'serialized.txt', content: 'first writer' }
    let releaseFirst!: () => void
    let markFirstEntered!: () => void
    const firstMayFinish = new Promise<void>((resolve) => {
      releaseFirst = resolve
    })
    const firstEntered = new Promise<void>((resolve) => {
      markFirstEntered = resolve
    })

    const first = withCoworkMutationLock(autoProject.id, async () => {
      await expect(approvalRequirement(autoProject, 'write_file', input)).resolves.toBeNull()
      markFirstEntered()
      await firstMayFinish
      await executeCoworkTool(autoProject, 'write_file', input, { allowOverwrite: false })
    })
    await firstEntered

    let secondEntered = false
    const second = withCoworkMutationLock(autoProject.id, async () => {
      secondEntered = true
      return approvalRequirement(autoProject, 'write_file', {
        path: input.path,
        content: 'second writer'
      })
    })

    await Promise.resolve()
    try {
      expect(secondEntered).toBe(false)
    } finally {
      releaseFirst()
    }

    await first
    await expect(second).resolves.toMatchObject({ risk: 'overwrite' })
    await expect(fs.readFile(path.join(projectRoot, input.path), 'utf8')).resolves.toBe(
      'first writer'
    )
  })

  it('does not disclose sensitive files in directory listings', async () => {
    await fs.writeFile(path.join(projectRoot, 'README.md'), 'public')
    await fs.writeFile(path.join(projectRoot, '.env'), 'TOKEN=secret')
    await fs.mkdir(path.join(projectRoot, '.git'))
    await fs.writeFile(path.join(projectRoot, '.git', 'config'), 'credential=secret')
    await fs.mkdir(path.join(projectRoot, 'nested'))
    await fs.writeFile(path.join(projectRoot, 'nested', 'wallet.key'), 'secret')

    const output = await executeCoworkTool(project(), 'list_files', { path: '.', maxDepth: 3 })
    const entries = (output.result as { entries: Array<{ path: string }> }).entries.map(
      (entry) => entry.path
    )

    expect(entries).toContain('README.md')
    expect(entries).not.toContain('.env')
    expect(entries).not.toContain('.git')
    expect(entries).not.toContain(path.join('nested', 'wallet.key'))
  })

  it('never allows the connected project root to be deleted', async () => {
    await expect(executeCoworkTool(project(), 'delete_file', { path: '.' })).rejects.toThrow(
      /project folder itself cannot be deleted/
    )
    expect(electron.trashItem).not.toHaveBeenCalled()
  })

  it('never copies or moves a file onto itself', async () => {
    await fs.writeFile(path.join(projectRoot, 'same.txt'), 'keep me')
    await expect(
      executeCoworkTool(
        project('skip'),
        'copy_file',
        {
          source: 'same.txt',
          destination: 'same.txt'
        },
        { allowOverwrite: true }
      )
    ).rejects.toThrow(/different paths/)
    await expect(
      executeCoworkTool(
        project('skip'),
        'move_file',
        {
          source: 'same.txt',
          destination: 'same.txt'
        },
        { allowOverwrite: true }
      )
    ).rejects.toThrow(/different paths/)
    await expect(fs.readFile(path.join(projectRoot, 'same.txt'), 'utf8')).resolves.toBe('keep me')
  })

  it('rejects malformed tool arguments', () => {
    expect(() => toolArguments('[]')).toThrow(/must be an object/)
    expect(() => toolArguments('{not json')).toThrow(/Invalid tool arguments/)
  })

  it('classifies every filesystem mutation for project-level serialization', () => {
    for (const name of [
      'write_file',
      'make_directory',
      'copy_file',
      'move_file',
      'delete_file',
      'create_docx',
      'create_xlsx',
      'create_pptx',
      'create_pdf'
    ]) {
      expect(isCoworkMutationTool(name)).toBe(true)
    }
    for (const name of [
      'list_files',
      'inspect_file',
      'read_file',
      'read_document',
      'search_files'
    ]) {
      expect(isCoworkMutationTool(name)).toBe(false)
    }
  })
})
