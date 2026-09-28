import { describe, expect, it } from 'vitest'
import { analyzeCsvText } from './cowork-data-analysis'

describe('analyzeCsvText', () => {
  it('summarizes numeric and text columns without floating-point coercion of text', () => {
    const result = analyzeCsvText(
      'region,amount,note\nAPAC,10,"hello, world"\nEMEA,20,hello\nAPAC,,hello'
    )
    expect(result).toMatchObject({ rowCount: 3, columnCount: 3, truncated: false })
    expect(result.columns[0]).toMatchObject({ name: 'region', kind: 'text', uniqueCount: 2 })
    expect(result.columns[1]).toMatchObject({
      kind: 'number',
      nonEmpty: 2,
      missing: 1,
      sum: 30,
      mean: 15,
      min: 10,
      max: 20
    })
    expect(result.columns[2].topValues[0]).toEqual({ value: 'hello', count: 2 })
  })

  it('supports tabs, escaped quotes, embedded newlines, and duplicate headers', () => {
    const result = analyzeCsvText('name\tname\n"A"\t"line 1\nline 2"\n"B"\t"said ""yes"""', '\\t')
    expect(result.columns.map((column) => column.name)).toEqual(['name', 'name_2'])
    expect(result.sample[0].name_2).toBe('line 1\nline 2')
    expect(result.sample[1].name_2).toBe('said "yes"')
  })

  it('rejects unsafe sizes, delimiters, and malformed quoted input', () => {
    expect(() => analyzeCsvText('a\nb', ':')).toThrow(/delimiter/)
    expect(() => analyzeCsvText('a\n"unfinished')).toThrow(/unterminated/)
    expect(() => analyzeCsvText('a'.repeat(512 * 1024 + 1))).toThrow(/analysis limit/)
  })
})
