import {
  MAX_IMAGES_PER_REQUEST,
  MAX_REQUEST_IMAGE_BYTES,
  coworkImageDataUrl,
  encodedImageBytes
} from './cowork-images'
import type {
  CoworkAgentMessage,
  CoworkContentPart,
  CoworkImageReference,
  CoworkToolCall
} from './cowork.types'

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
    ...(Array.isArray(message.content)
      ? { content: message.content.map((part) => ({ ...part })) }
      : {}),
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

/** Content is a parts array only on the image messages, which are never tool results. */
export const messageTextContent = (content: CoworkAgentMessage['content']): string =>
  typeof content === 'string' ? content : ''

function resultOutcome(message: CoworkAgentMessage): string {
  try {
    const value = JSON.parse(messageTextContent(message.content) || '{}') as Record<string, unknown>
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
    if (value.ok === false) {
      // The reason is the one part of a failed result worth keeping. Dropping it
      // left the model unable to tell a malformed request from a transient fault,
      // so it re-sent the same broken call and narrated the same non-diagnosis.
      const reason = typeof value.error === 'string' ? value.error.trim() : ''
      if (!reason) return 'reported failure'
      return `reported failure: ${reason.replace(/[\u0000-\u001f\u007f]+/g, ' ').slice(0, 300)}`
    }
  } catch {
    // The historical result is untrusted and need not be reproduced for the model.
  }
  return 'returned an unknown outcome'
}

function historicalLine(call: CoworkToolCall, result: CoworkAgentMessage): string {
  const args = parsedArguments(call)
  return `- ${call.function.name} on ${safePath(args?.path)} ${resultOutcome(result)}; generated payload omitted.`
}

function appendHistoricalSummary(
  output: CoworkAgentMessage[],
  lines: string[],
  /**
   * Opaque reasoning state of the replaced assistant turn. Thinking-mode
   * providers require their own prior reasoning back, so a record that fully
   * replaces an assistant turn inherits it rather than discarding it.
   */
  reasoning?: string | null
): void {
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
    // A coalesced record stands in for the newest turn it replaced.
    if (reasoning !== undefined) previous.reasoning_content = reasoning
    return
  }
  output.push({
    role: 'assistant',
    content: `${HISTORY_SUMMARY_HEADER}\n${lines.join('\n')}\n${HISTORY_SUMMARY_FOOTER}`,
    ...(reasoning === undefined ? {} : { reasoning_content: reasoning })
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
      compacted.map((call) => historicalLine(call, resultsById.get(call.id)!)),
      // A retained partial turn already carries the reasoning state itself.
      retainedCalls.length ? undefined : message.reasoning_content
    )
    index = resultEnd - 1
  }

  return output
}

/** An outbound message in the shape the completions endpoint expects. */
export interface CoworkOutboundMessage {
  role: 'user' | 'assistant' | 'tool'
  content?: string | null | Array<Record<string, unknown>>
  reasoning_content?: string | null
  tool_calls?: CoworkToolCall[]
  tool_call_id?: string
}

const describeImage = (reference: CoworkImageReference): string => {
  const size = `${Math.max(1, Math.round(reference.bytes / 1024)).toLocaleString('en-US')} KB`
  const frame =
    reference.width && reference.height ? `, ${reference.width}\u00d7${reference.height}` : ''
  return `${JSON.stringify(reference.path)} (${reference.mediaType}${frame}, ${size})`
}

/**
 * Chooses which images are actually sent as pixels. A long task can look at
 * many images, and resending every one on every turn would grow each request
 * without bound for no benefit: the model has already described the older ones
 * in the transcript. The newest few are sent; the rest are named in text so
 * the model still knows they exist and can ask for one again.
 */
function sendableImageParts(messages: readonly CoworkAgentMessage[]): Set<CoworkContentPart> {
  const sendable = new Set<CoworkContentPart>()
  let budget = MAX_REQUEST_IMAGE_BYTES
  for (let index = messages.length - 1; index >= 0; index--) {
    const content = messages[index].content
    if (!Array.isArray(content)) continue
    for (let part = content.length - 1; part >= 0; part--) {
      const candidate = content[part]
      if (candidate.type !== 'image') continue
      if (sendable.size >= MAX_IMAGES_PER_REQUEST) return sendable
      const cost = encodedImageBytes(candidate.image.bytes)
      if (cost > budget) return sendable
      budget -= cost
      sendable.add(candidate)
    }
  }
  return sendable
}

/**
 * Turns stored image references into the content parts a provider accepts,
 * reading the bytes only for the images actually being sent. A reference whose
 * file has since been moved or deleted degrades to a line of text rather than
 * failing the turn, because the rest of the conversation is still valid.
 */
export async function materialiseCoworkImages(
  messages: readonly CoworkAgentMessage[],
  load: (reference: CoworkImageReference) => Promise<Buffer | null>,
  options: { pixels?: boolean } = {}
): Promise<CoworkOutboundMessage[]> {
  // Set false for an endpoint that has refused image parts. Every picture then
  // degrades to a line of text and the message collapses to a plain string,
  // which is a shape no endpoint rejects.
  const pixels = options.pixels !== false
  const sendable = pixels ? sendableImageParts(messages) : new Set<CoworkContentPart>()
  const outbound: CoworkOutboundMessage[] = []

  for (const message of messages) {
    const content = message.content
    if (!Array.isArray(content)) {
      outbound.push(message as CoworkOutboundMessage)
      continue
    }
    const parts: Array<Record<string, unknown>> = []
    for (const part of content) {
      if (part.type === 'text') {
        if (part.text) parts.push({ type: 'text', text: part.text })
        continue
      }
      if (!sendable.has(part)) {
        parts.push({
          type: 'text',
          text: pixels
            ? `[earlier image ${describeImage(part.image)}]`
            : `[image ${describeImage(part.image)}; this model endpoint cannot receive pictures, so only these details are available]`
        })
        continue
      }
      const buffer = await load(part.image).catch(() => null)
      if (!buffer) {
        parts.push({
          type: 'text',
          text: `[image ${describeImage(part.image)} is no longer readable in the connected folder]`
        })
        continue
      }
      parts.push({
        type: 'image_url',
        image_url: { url: coworkImageDataUrl(buffer, part.image.mediaType) }
      })
    }
    // A message that ended up entirely textual is sent as a plain string, which
    // every endpoint accepts, including ones that reject the parts form.
    const textOnly = parts.every((part) => part.type === 'text')
    outbound.push({
      ...(message as CoworkOutboundMessage),
      content: textOnly ? parts.map((part) => String(part.text)).join('\n') : parts
    })
  }

  return outbound
}
