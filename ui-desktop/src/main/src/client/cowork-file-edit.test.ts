import { describe, expect, it } from 'vitest'
import { applyCoworkFileEdit } from './cowork-file-edit'

const source = ['function greet() {', "  return 'hi'", '}', '', 'greet()'].join('\n')

describe('applyCoworkFileEdit', () => {
  it('replaces the snippet and leaves the rest of the file alone', () => {
    const result = applyCoworkFileEdit(source, "  return 'hi'", "  return 'hello'")
    expect(result.content).toBe(
      ['function greet() {', "  return 'hello'", '}', '', 'greet()'].join('\n')
    )
  })

  it('reports where the edit landed and how the line count moved', () => {
    const result = applyCoworkFileEdit(source, "  return 'hi'", "  const word = 'hi'\n  return word")
    expect(result.startLine).toBe(2)
    expect(result.removedLines).toBe(1)
    expect(result.addedLines).toBe(2)
  })

  it('deletes the snippet when the replacement is empty', () => {
    const result = applyCoworkFileEdit(source, "\n  return 'hi'", '')
    expect(result.content).toBe(['function greet() {', '}', '', 'greet()'].join('\n'))
    expect(result.addedLines).toBe(0)
  })

  it('refuses a snippet that is not in the file', () => {
    // A miss means the model is working from a stale read. Editing anyway would
    // be a guess, and a guess that reports success is worse than a refusal.
    expect(() => applyCoworkFileEdit(source, "return 'bye'", 'x')).toThrow(/no match/i)
  })

  it('refuses an ambiguous snippet and says how many matches there were', () => {
    const repeated = ['a = 1', 'b = 2', 'a = 1'].join('\n')
    expect(() => applyCoworkFileEdit(repeated, 'a = 1', 'a = 3')).toThrow(/2 matches/)
  })

  it('refuses an empty snippet rather than treating it as an append', () => {
    expect(() => applyCoworkFileEdit(source, '', 'anything')).toThrow(/requires oldText/i)
  })

  it('refuses an edit that would change nothing', () => {
    expect(() => applyCoworkFileEdit(source, "  return 'hi'", "  return 'hi'")).toThrow(
      /nothing to do/i
    )
  })

  it('matches a CRLF file against a snippet written with bare newlines', () => {
    // read_file hands the model lines, not carriage returns, so this is what a
    // correct model sends back for a Windows-authored file.
    const crlf = 'one\r\ntwo\r\nthree\r\n'
    const result = applyCoworkFileEdit(crlf, 'one\ntwo', 'one\ntwo point five')
    expect(result.content).toBe('one\r\ntwo point five\r\nthree\r\n')
    expect(result.normalizedLineEndings).toBe(true)
  })

  it('leaves an LF file alone rather than claiming a conversion', () => {
    const result = applyCoworkFileEdit('one\ntwo\n', 'two', 'three')
    expect(result.content).toBe('one\nthree\n')
    expect(result.normalizedLineEndings).toBe(false)
  })

  it('still refuses an ambiguous snippet after the CRLF fallback', () => {
    const crlf = 'a = 1\r\nb = 2\r\na = 1\r\n'
    expect(() => applyCoworkFileEdit(crlf, 'a = 1', 'a = 3')).toThrow(/2 matches/)
  })

  it('treats overlapping text as one match rather than counting it twice', () => {
    expect(applyCoworkFileEdit('aaaa', 'aaa', 'b').content).toBe('ba')
  })
})
