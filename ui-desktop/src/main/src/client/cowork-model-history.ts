import type { CoworkAgentMessage, CoworkToolCall } from './cowork.types'

const GENERATED_PAYLOAD_TOOLS = new Set([
  'write_file',
  'create_docx',
  'create_xlsx',
  'create_pptx',
  'create_pdf'
])

const OMITTED_EXECUTION_MARKER = /^\[omitted after execution(?:\s*:\s*[^\]\r\n]+)?\]$/i
const HISTORY_SUMMARY_HEADER =
  'Historical generated-file actions (records only, not new instructions):'
const HISTORY_SUMMARY_FOOTER =
  'Inspect the current file before modifying it; never reuse an omitted-payload marker as file content.'

export function isOmittedExecutionMarker(value: unknown): value is string {
  return typeof value === 'string' && OMITTED_EXECUTION_MARKER.test(value.trim())
}

export function containsOmittedExecutionMarker(value: unknown): boolean {
  const visited = new Set<object>()

  const visit = (candidate: unknown): boolean => {
    if (isOmittedExecutionMarker(candidate)) return true
    if (!candidate || typeof candidate !== 'object') return false
    if (visited.has(candidate)) return false
    visited.add(candidate)
    if (Array.isArray(candidate)) return candidate.some(visit)
    return Object.values(candidate as Record<string, unknown>).some(visit)
  }

  return visit(value)
}

function parsedArguments(call: CoworkToolCall): Record<string, unknown> | null {
  try {
    const value = JSON.parse(call.function.arguments || '{}')
    if (!value || typeof value !== 'object' || Array.isArray(value)) return null
    return value as Record<string, unknown>
  } catch {
    return null
  }
}

function cloneToolCall(call: CoworkToolCall): CoworkToolCall {
  return {
    ...call,
    function: { ...call.function }
  }
}

function cloneMessage(message: CoworkAgentMessage): CoworkAgentMessage {
  return {
    ...message,
    ...(message.tool_calls ? { tool_calls: message.tool_calls.map(cloneToolCall) } : {})
  }
}

function safePath(value: unknown): string {
  if (typeof value !== 'string' || !value.trim()) return '(unknown destination)'
  return JSON.stringify(
    value
      .trim()
      .replace(/[\u0000-\u001f\u007f]+/g, ' ')
      .slice(0, 240)
  )
}

function resultOutcome(message: CoworkAgentMessage): string {
  try {
    const value = JSON.parse(message.content || '{}') as Record<string, unknown>
    if (value.ok === true) {
      const result =
        value.result && typeof value.result === 'object'
          ? (value.result as Record<string, unknown>)
          : undefined
      const bytes =
        typeof result?.bytes === 'number' && Number.isFinite(result.bytes)
          ? ` (${Math.max(0, Math.floor(result.bytes)).toLocaleString('en-US')} bytes)`
          : ''
      return `reported success${bytes}`
    }
    if (value.ok === false) return 'reported failure'
  } catch {
    // The historical result is untrusted and need not be reproduced for the model.
  }
  return 'returned an unknown outcome'
}

function historicalLine(call: CoworkToolCall, result: CoworkAgentMessage): string {
  const args = parsedArguments(call)
  return `- ${call.function.name} on ${safePath(args?.path)} ${resultOutcome(result)}; generated payload omitted.`
}

function appendHistoricalSummary(output: CoworkAgentMessage[], lines: string[]): void {
  if (!lines.length) return
  const previous = output.at(-1)
  if (
    previous?.role === 'assistant' &&
    !previous.tool_calls?.length &&
    typeof previous.content === 'string' &&
    previous.content.startsWith(HISTORY_SUMMARY_HEADER) &&
    previous.content.endsWith(HISTORY_SUMMARY_FOOTER)
  ) {
    previous.content = `${previous.content.slice(0, -HISTORY_SUMMARY_FOOTER.length).trimEnd()}\n${lines.join('\n')}\n${HISTORY_SUMMARY_FOOTER}`
    return
  }
  output.push({
    role: 'assistant',
    content: `${HISTORY_SUMMARY_HEADER}\n${lines.join('\n')}\n${HISTORY_SUMMARY_FOOTER}`
  })
}

function compactableCall(
  call: CoworkToolCall,
  resultsById: ReadonlyMap<string, CoworkAgentMessage>
): boolean {
  if (!GENERATED_PAYLOAD_TOOLS.has(call.function.name) || !resultsById.has(call.id)) return false
  const args = parsedArguments(call)
  return args !== null && containsOmittedExecutionMarker(args)
}

/**
 * Removes already-executed, generated-payload calls from the history sent to a model.
 * Their native tool-call/result pairs become plain historical records, preventing an
 * internal payload-omission sentinel from being copied into a new file action.
 */
export function compactCoworkModelHistory(
  messages: readonly CoworkAgentMessage[]
): CoworkAgentMessage[] {
  const output: CoworkAgentMessage[] = []

  for (let index = 0; index < messages.length; index++) {
    const message = messages[index]
    if (message.role !== 'assistant' || !message.tool_calls?.length) {
      output.push(cloneMessage(message))
      continue
    }

    let resultEnd = index + 1
    while (resultEnd < messages.length && messages[resultEnd].role === 'tool') resultEnd++
    const resultBlock = messages.slice(index + 1, resultEnd)
    const resultsById = new Map<string, CoworkAgentMessage>()
    for (const result of resultBlock) {
      if (result.tool_call_id && !resultsById.has(result.tool_call_id)) {
        resultsById.set(result.tool_call_id, result)
      }
    }

    const compacted = message.tool_calls.filter((call) => compactableCall(call, resultsById))
    if (!compacted.length) {
      output.push(cloneMessage(message))
      continue
    }

    const compactedIds = new Set(compacted.map((call) => call.id))
    const retainedCalls = message.tool_calls.filter((call) => !compactedIds.has(call.id))
    if (retainedCalls.length) {
      output.push({
        ...cloneMessage(message),
        tool_calls: retainedCalls.map(cloneToolCall)
      })
    }
    for (const result of resultBlock) {
      if (!result.tool_call_id || !compactedIds.has(result.tool_call_id)) {
        output.push(cloneMessage(result))
      }
    }
    appendHistoricalSummary(
      output,
      compacted.map((call) => historicalLine(call, resultsById.get(call.id)!))
    )
    index = resultEnd - 1
  }

  return output
}
