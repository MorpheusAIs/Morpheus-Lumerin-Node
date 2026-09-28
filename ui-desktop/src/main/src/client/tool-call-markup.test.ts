import { describe, expect, it } from 'vitest'
import { analyzeToolCallMarkup } from './tool-call-markup'
import {
  LEAKED_BARE_JSON_CALL,
  LEAKED_PLACEHOLDER_REPLY,
  LEAKED_PLAN_UPDATE,
  LEAKED_SINGLE_TOOL_CALL,
  LEAKED_TOOL_CALLS_BLOCK
} from './tool-call-markup.fixtures'

describe('analyzeToolCallMarkup', () => {
  it('flags the reported <tool_calls> list_files placeholder as markup only', () => {
    const verdict = analyzeToolCallMarkup(LEAKED_TOOL_CALLS_BLOCK)
    expect(verdict).toMatchObject({ found: true, markupOnly: true, cleaned: '' })
    expect(verdict.toolNames).toContain('list_files')
  })

  it('flags the reported plan-update markup', () => {
    const verdict = analyzeToolCallMarkup(LEAKED_PLAN_UPDATE)
    expect(verdict).toMatchObject({ found: true, markupOnly: true })
    expect(verdict.toolNames).toContain('update_plan_step')
  })

  it('strips the markup from a reply that wraps it in a sentence', () => {
    const verdict = analyzeToolCallMarkup(LEAKED_PLACEHOLDER_REPLY)
    expect(verdict.found).toBe(true)
    expect(verdict.markupOnly).toBe(false)
    expect(verdict.cleaned).toBe("I'll start by looking at the project.")
    expect(verdict.toolNames).toEqual(expect.arrayContaining(['list_files', 'update_plan_step']))
  })

  it.each([
    ['single <tool_call>', LEAKED_SINGLE_TOOL_CALL],
    ['bare JSON call', LEAKED_BARE_JSON_CALL],
    ['fenced JSON call', '```json\n' + LEAKED_BARE_JSON_CALL + '\n```'],
    ['special token', '[TOOL_CALLS] [{"name":"list_files","arguments":{"path":"."}}]'],
    ['pipe token', '<|tool_call|>{"name":"list_files","arguments":{}}'],
    ['function tag', '<function=list_files>{"path":"."}</function>'],
    ['self-closing tool tag', '<list_files path="." />'],
    ['unterminated while streaming', '<tool_calls>\n[{"name": "list_fi'],
    ['after reasoning', '<think>Need files.</think>\n' + LEAKED_TOOL_CALLS_BLOCK]
  ])('flags %s', (_label, text) => {
    expect(analyzeToolCallMarkup(text)).toMatchObject({ found: true, markupOnly: true })
  })

  it.each([
    ['plain answer', 'The project has three files: a.md, b.md and c.md.'],
    ['inline code mention', 'Models emit `<tool_calls>` blocks when a parser is missing.'],
    ['ordinary JSON answer', '{"name": "Alice", "age": 30}'],
    ['mid-sentence tag mention', 'Wrap it in <tool_call> tags, then send it.']
  ])('leaves a %s alone', (_label, text) => {
    const verdict = analyzeToolCallMarkup(text)
    expect(verdict).toMatchObject({ found: false, markupOnly: false, cleaned: text })
  })

  it('keeps an explanation around a fence and drops only a fence that is just a call', () => {
    const text =
      'Here is what a call looks like:\n\n```json\n' +
      LEAKED_BARE_JSON_CALL +
      '\n```\n\n```xml\n<tool_calls>\n</tool_calls>\nnot a call\n```'
    const verdict = analyzeToolCallMarkup(text)
    expect(verdict.markupOnly).toBe(false)
    expect(verdict.cleaned).toContain('Here is what a call looks like')
    expect(verdict.cleaned).toContain('not a call')
  })

  it('treats empty input as no markup', () => {
    expect(analyzeToolCallMarkup(null)).toMatchObject({ found: false, markupOnly: false })
  })
})
