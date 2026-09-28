import { describe, expect, it, vi } from 'vitest'

vi.mock('../../logger', () => ({
  default: { error: vi.fn() }
}))

import { parseAttachment } from './attachments'

const bytes = (value: string | number[]): ArrayBuffer =>
  Uint8Array.from(typeof value === 'string' ? Buffer.from(value, 'utf8') : value).buffer

describe('chat attachment parsing boundary', () => {
  it('extracts bounded plain text without requiring a MIME type', async () => {
    await expect(
      parseAttachment({
        name: 'notes.md',
        mime: '',
        data: bytes('Project notes')
      })
    ).resolves.toEqual({ text: 'Project notes', empty: false })
  })

  it('rejects malformed data and invalid filenames before parser dispatch', async () => {
    await expect(
      parseAttachment({ name: 'notes.md', mime: 'text/markdown', data: 'not binary' as any })
    ).rejects.toThrow(/attachment data is invalid/i)
    await expect(
      parseAttachment({
        name: '../bad\u0000name.pdf',
        mime: 'application/pdf',
        data: bytes('x')
      })
    ).rejects.toThrow(/filename is invalid/i)
  })

  it('keeps binary text-like files out of the prompt', async () => {
    await expect(
      parseAttachment({
        name: 'payload.txt',
        mime: 'text/plain',
        data: bytes([0x41, 0x00, 0x42])
      })
    ).resolves.toMatchObject({ text: '', empty: true })
  })
})
