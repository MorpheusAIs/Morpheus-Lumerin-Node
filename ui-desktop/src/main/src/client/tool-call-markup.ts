/**
 * Detects tool-call markup that a model wrote into its reply text instead of
 * using the API's tool_calls field.
 *
 * Models whose serving stack has no tool parser (or that were given a tool
 * manifest they cannot honour) print their chat template's tool syntax as
 * plain content: `<tool_calls>[{"name":"list_files",...}]</tool_calls>`,
 * `<tool_call>{...}</tool_call>`, `[TOOL_CALLS]`, `<update_plan_step>...`, or a
 * bare `{"name":"list_files","arguments":{...}}`. None of that is an answer.
 * It must never execute (only the structured tool_calls field or the strict
 * Workspace envelope can authorise an action) and it must never be shown as
 * the assistant's reply.
 *
 * Dependency-free on purpose: the main-process runner and both renderers use
 * the same verdict.
 */

/** Workspace tools plus the plan-update spellings models invent for them. */
export const KNOWN_TOOL_CALL_NAMES: readonly string[] = [
  'set_plan',
  'update_plan_step',
  'update_plan',
  'plan_update',
  'list_files',
  'inspect_file',
  'read_file',
  'read_document',
  'read_image',
  'search_files',
  'analyze_csv',
  'fetch_web_page',
  'write_file',
  'edit_file',
  'create_docx',
  'create_xlsx',
  'create_pptx',
  'create_pdf',
  'make_directory',
  'copy_file',
  'move_file',
  'delete_file',
  'delegate_analysis',
  'finish_task'
]

export type ToolCallMarkupVerdict = {
  /** Tool-call markup appeared outside ordinary code the reply meant to show. */
  found: boolean
  /** Nothing but markup (and whitespace or reasoning blocks) is in the reply. */
  markupOnly: boolean
  /** Tool names the markup referenced, in order of first appearance. */
  toolNames: string[]
  /** The reply with every markup segment removed. */
  cleaned: string
}

/** Wrapper tags chat templates use around tool calls. */
const WRAPPER_TAGS = [
  'tool_calls',
  'tool_call',
  'tool_use',
  'tool_code',
  'tool_request',
  'function_calls',
  'function_call',
  'invoke'
]

const REASONING_BLOCK =
  /<(think|thinking|thought|reasoning|reflection)(?:\s[^>]*)?>[\s\S]*?(?:<\/\1\s*>|$)/gi

type Span = { start: number; end: number }

const escapeRegExp = (value: string): string => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const isPlainObject = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === 'object' && !Array.isArray(value)

/** Scans one balanced JSON value starting at `start`; returns its end or -1. */
function balancedJsonEnd(text: string, start: number): number {
  const open = text[start]
  if (open !== '{' && open !== '[') return -1
  const stack: string[] = []
  let inString = false
  for (let index = start; index < text.length; index++) {
    const char = text[index]
    if (inString) {
      if (char === '\\') index++
      else if (char === '"') inString = false
      continue
    }
    if (char === '"') inString = true
    else if (char === '{' || char === '[') stack.push(char)
    else if (char === '}' || char === ']') {
      const expected = char === '}' ? '{' : '['
      if (stack.pop() !== expected) return -1
      if (!stack.length) return index + 1
    }
  }
  return -1
}

function toolCallNames(value: unknown, known: ReadonlySet<string>, names: string[]): boolean {
  if (Array.isArray(value)) {
    if (!value.length) return false
    return value.every((item) => toolCallNames(item, known, names))
  }
  if (!isPlainObject(value)) return false
  if (Array.isArray(value.tool_calls) && value.tool_calls.length) {
    return toolCallNames(value.tool_calls, known, names)
  }
  const fn = isPlainObject(value.function) ? value.function : null
  const name =
    typeof value.name === 'string'
      ? value.name
      : typeof value.tool === 'string'
        ? value.tool
        : typeof fn?.name === 'string'
          ? fn.name
          : ''
  if (!name) return false
  const hasArguments = ['arguments', 'parameters', 'args', 'input'].some(
    (key) => key in value || (fn !== null && key in fn)
  )
  const typed = value.type === 'function' || value.type === 'tool_call' || value.type === 'tool_use'
  // An unknown name only counts when the object is unmistakably a call.
  if (!known.has(name) && !(fn && hasArguments) && !typed) return false
  if (!hasArguments && !typed) return false
  if (!names.includes(name)) names.push(name)
  return true
}

function jsonToolCall(text: string, known: ReadonlySet<string>, names: string[]): boolean {
  const trimmed = text.trim()
  if (!trimmed) return false
  try {
    return toolCallNames(JSON.parse(trimmed), known, names)
  } catch {
    return false
  }
}

function nameFromMarkupBody(body: string, known: ReadonlySet<string>, names: string[]): void {
  if (jsonToolCall(body, known, names)) return
  const attribute = body.match(/\bname\s*=\s*["']?([A-Za-z0-9_.-]+)/)
  const jsonName = body.match(/"name"\s*:\s*"([A-Za-z0-9_.-]+)"/)
  const functionTag = body.match(/<function=([A-Za-z0-9_.-]+)>/)
  for (const candidate of [attribute?.[1], jsonName?.[1], functionTag?.[1]]) {
    if (candidate && !names.includes(candidate)) names.push(candidate)
  }
}

/**
 * Finds markup spans in text that holds no code the reply meant to show.
 * Unterminated wrappers run to the end, which also covers a reply still
 * streaming in.
 */
function markupSpans(text: string, known: ReadonlySet<string>, names: string[]): Span[] {
  const spans: Span[] = []
  const add = (start: number, end: number): void => {
    spans.push({ start, end })
  }

  // Tag wrappers: <tool_calls>…</tool_calls>, tags named after a tool, and
  // <function=name>…</function>. Only at a line start, so prose that merely
  // mentions a tag mid-sentence is left alone.
  const tagNames = [...WRAPPER_TAGS, ...known].map(escapeRegExp).join('|')
  const openTag = new RegExp(`(^|\\n)[ \\t]*<(${tagNames})(\\s[^>]*)?(/?)>`, 'gi')
  for (let match = openTag.exec(text); match; match = openTag.exec(text)) {
    const start = match.index + match[1].length
    const tag = match[2]
    const bodyStart = match.index + match[0].length
    if (known.has(tag) && !names.includes(tag)) names.push(tag)
    if (match[4] === '/') {
      add(start, bodyStart)
      continue
    }
    const close = new RegExp(`</${escapeRegExp(tag)}\\s*>`, 'i')
    const rest = text.slice(bodyStart)
    const closing = rest.match(close)
    const end =
      closing?.index !== undefined ? bodyStart + closing.index + closing[0].length : text.length
    nameFromMarkupBody(
      text.slice(bodyStart, closing?.index !== undefined ? bodyStart + closing.index : text.length),
      known,
      names
    )
    add(start, end)
    openTag.lastIndex = end
  }

  const functionTag = /(^|\n)[ \t]*<function=([A-Za-z0-9_.-]+)>/g
  for (let match = functionTag.exec(text); match; match = functionTag.exec(text)) {
    const start = match.index + match[1].length
    const rest = text.slice(match.index + match[0].length)
    const closing = rest.match(/<\/function\s*>/)
    const end =
      closing?.index !== undefined
        ? match.index + match[0].length + closing.index + closing[0].length
        : text.length
    if (!names.includes(match[2])) names.push(match[2])
    add(start, end)
    functionTag.lastIndex = end
  }

  // Special-token forms: <|tool_call|>, <|python_tag|>, <｜tool▁calls▁begin｜>,
  // [TOOL_CALLS]. These never occur in a real answer, so they match anywhere
  // and consume the rest of the reply.
  const token =
    /<[|｜](?:tool[_▁ ]?calls?(?:[_▁ ]section)?(?:[_▁ ]begin)?|python_tag|start_of_tool_call)[|｜]>|\[TOOL_CALLS\]/i
  const tokenMatch = text.match(token)
  if (tokenMatch?.index !== undefined) {
    nameFromMarkupBody(text.slice(tokenMatch.index + tokenMatch[0].length), known, names)
    add(tokenMatch.index, text.length)
  }

  // Bare JSON calls at a line start: {"name":"list_files","arguments":{…}}.
  const jsonStart = /(^|\n)[ \t]*([{[])/g
  for (let match = jsonStart.exec(text); match; match = jsonStart.exec(text)) {
    const start = match.index + match[0].length - 1
    const end = balancedJsonEnd(text, start)
    if (end < 0) continue
    if (jsonToolCall(text.slice(start, end), known, names)) {
      add(start, end)
      jsonStart.lastIndex = end
    }
  }
  return spans
}

function removeSpans(text: string, spans: Span[]): string {
  if (!spans.length) return text
  const sorted = [...spans].sort((a, b) => a.start - b.start)
  let output = ''
  let cursor = 0
  for (const span of sorted) {
    if (span.end <= cursor) continue
    output += text.slice(cursor, Math.max(cursor, span.start))
    cursor = Math.max(cursor, span.end)
  }
  return output + text.slice(cursor)
}

/**
 * Splits the reply around fenced code. A fence whose whole body is a tool call
 * is markup; any other fence is content the reply meant to show and is never
 * inspected, so an answer explaining tool-call syntax renders untouched.
 */
function splitFences(text: string): Array<{ code: boolean; text: string }> {
  const parts: Array<{ code: boolean; text: string }> = []
  const fence = /(^|\n)([ \t]*)(```|~~~)[^\n]*\n?/g
  let cursor = 0
  for (let match = fence.exec(text); match; match = fence.exec(text)) {
    const openStart = match.index + match[1].length
    const marker = match[3]
    const bodyStart = match.index + match[0].length
    const closeRe = new RegExp(`\\n[ \\t]*${marker}[ \\t]*(?=\\n|$)`, 'g')
    closeRe.lastIndex = Math.max(bodyStart - 1, 0)
    const close = closeRe.exec(text)
    const end = close ? close.index + close[0].length : text.length
    parts.push({ code: false, text: text.slice(cursor, openStart) })
    parts.push({ code: true, text: text.slice(openStart, end) })
    cursor = end
    fence.lastIndex = end
  }
  parts.push({ code: false, text: text.slice(cursor) })
  return parts
}

function fenceBody(block: string): string {
  return block.replace(/^[ \t]*(```|~~~)[^\n]*\n?/, '').replace(/\n?[ \t]*(```|~~~)[ \t]*$/, '')
}

/** Replaces inline code spans with same-length filler so they are never matched. */
const maskInlineCode = (text: string): string =>
  text.replace(/`[^`\n]*`/g, (span) => ' '.repeat(span.length))

export function analyzeToolCallMarkup(
  text: string | null | undefined,
  knownToolNames: Iterable<string> = KNOWN_TOOL_CALL_NAMES
): ToolCallMarkupVerdict {
  const source = typeof text === 'string' ? text : ''
  const known = new Set(knownToolNames)
  for (const name of KNOWN_TOOL_CALL_NAMES) known.add(name)
  const names: string[] = []
  let found = false
  let cleaned = ''

  for (const part of splitFences(source)) {
    if (part.code) {
      const body = fenceBody(part.text)
      const bodyNames: string[] = []
      const bodySpans = markupSpans(maskInlineCode(body), known, bodyNames)
      const residue = removeSpans(body, bodySpans).trim()
      if (bodySpans.length && !residue) {
        found = true
        for (const name of bodyNames) if (!names.includes(name)) names.push(name)
      } else {
        cleaned += part.text
      }
      continue
    }
    const spans = markupSpans(maskInlineCode(part.text), known, names)
    if (spans.length) found = true
    cleaned += removeSpans(part.text, spans)
  }

  cleaned = cleaned.replace(/\n{3,}/g, '\n\n').trim()
  const visible = cleaned.replace(REASONING_BLOCK, '').trim()
  return {
    found,
    markupOnly: found && !/[\p{L}\p{N}]/u.test(visible),
    toolNames: found ? names : [],
    cleaned: found ? cleaned : source
  }
}

/** Plain-language notice shown in place of a reply that was only markup. */
export const TOOL_MARKUP_REPLY_NOTICE =
  'The model replied with a raw tool call instead of an answer, so there is nothing to show. ' +
  'Retry the message, or choose a model that supports tools.'
