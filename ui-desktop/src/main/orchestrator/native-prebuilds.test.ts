import { describe, expect, it } from 'vitest'
import { nativeBinaryFormat } from './native-prebuilds'

const header = (...bytes: number[]): Buffer => Buffer.from([...bytes, 0, 0, 0, 0])

describe('nativeBinaryFormat', () => {
  it('recognises a Windows PE object', () => {
    expect(nativeBinaryFormat(header(0x4d, 0x5a, 0x90, 0x00))).toBe('pe')
  })

  it('recognises an ELF object', () => {
    expect(nativeBinaryFormat(header(0x7f, 0x45, 0x4c, 0x46))).toBe('elf')
  })

  it('recognises 64-bit Mach-O in either byte order', () => {
    expect(nativeBinaryFormat(header(0xcf, 0xfa, 0xed, 0xfe))).toBe('mach-o')
    expect(nativeBinaryFormat(header(0xfe, 0xed, 0xfa, 0xcf))).toBe('mach-o')
  })

  it('recognises the fat wrapper a universal build produces', () => {
    expect(nativeBinaryFormat(header(0xca, 0xfe, 0xba, 0xbe))).toBe('mach-o')
  })

  it('reports nothing for content that is not an object file', () => {
    expect(nativeBinaryFormat(Buffer.from('not a binary at all', 'utf8'))).toBeNull()
    expect(nativeBinaryFormat(Buffer.from([0x00, 0x01]))).toBeNull()
  })

  /**
   * The exact case that shipped: a macOS binary packaged for Windows. The
   * formats must not be confusable, or the guard would pass and the app would
   * die on launch with "is not a valid Win32 application".
   */
  it('does not mistake a macOS binary for a Windows one', () => {
    expect(nativeBinaryFormat(header(0xcf, 0xfa, 0xed, 0xfe))).not.toBe('pe')
  })
})
