import log from '../../logger'

/**
 * Attachment text extraction.
 *
 * Runs in the main process on purpose. pdfjs and mammoth are large and pdfjs
 * wants a web worker; pulling them into the renderer bundle means worker
 * plumbing, a much bigger bundle, and CSP friction. Main is plain Node, so both
 * libraries work with no configuration, and the renderer only ever sees text.
 *
 * Why extraction at all: the chat-completions protocol has no attachment
 * concept. A message is text parts plus image_url parts — there is nowhere to
 * put a PDF. So documents are converted to text locally and injected into the
 * prompt, which has the useful property of working with every text model
 * rather than only vision ones.
 */

export type ParsedAttachment = {
  /** Extracted text. Empty when the file yielded nothing readable. */
  text: string
  /** Page count for PDFs, sheet names for spreadsheets, etc. */
  note?: string
  /** True when the file parsed but contained no text layer (e.g. a scanned PDF). */
  empty?: boolean
}

/** Rough cap on extracted text, in characters, to keep prompts sane. */
const MAX_TEXT_CHARS = 400_000

const truncate = (text: string): { text: string; note?: string } => {
  if (text.length <= MAX_TEXT_CHARS) {
    return { text }
  }
  return {
    text: text.slice(0, MAX_TEXT_CHARS),
    note: `truncated from ${text.length.toLocaleString()} characters`
  }
}

const MAX_ATTACHMENT_INPUT_BYTES = 20 * 1024 * 1024
const STRUCTURED_DOCUMENT_EXTENSIONS = new Set(['pdf', 'docx', 'xlsx', 'pptx'])

async function parseStructuredDocument(buf: Buffer, name: string): Promise<ParsedAttachment> {
  const { extractCoworkDocument } = await import('./cowork-document-extraction')
  const extracted = await extractCoworkDocument(buf, name)
  const sectionCount =
    extracted.metadata.pageCount ?? extracted.metadata.sheetCount ?? extracted.metadata.slideCount
  const sectionLabel =
    extracted.format === 'pdf'
      ? 'page(s)'
      : extracted.format === 'xlsx'
        ? 'sheet(s)'
        : extracted.format === 'pptx'
          ? 'slide(s)'
          : 'document section(s)'
  const notes = [
    sectionCount === undefined ? undefined : `${sectionCount} ${sectionLabel}`,
    extracted.truncated ? 'text truncated to the safe extraction limit' : undefined,
    ...extracted.warnings
  ].filter((value): value is string => Boolean(value))
  return {
    text: extracted.text,
    ...(notes.length ? { note: notes.join('; ').slice(0, 2_000) } : {}),
    empty: extracted.empty
  }
}

/**
 * Plain-text-ish formats: .txt .md .csv .json .svg .xml and source code.
 *
 * SVG is deliberately treated as text rather than as an image. It is XML, so
 * models read the source directly and get the structure, which is more useful
 * than a rasterised picture — and it works on non-vision models.
 */
function parseText(buf: Buffer): ParsedAttachment {
  const raw = buf.toString('utf-8')

  // A NUL byte in the first chunk means this is really binary; decoding it
  // produces replacement-character noise that wastes tokens and helps nobody.
  if (raw.slice(0, 4096).includes('\u0000')) {
    return { text: '', empty: true, note: 'binary file, not readable as text' }
  }

  const { text, note } = truncate(raw)
  return { text, note, empty: !text.trim() }
}

/**
 * Extracts text from an attachment.
 *
 * `data` is a structured-cloned binary buffer, not a path. This covers picker,
 * drop, and paste inputs without exposing a host path or amplifying a large
 * document into a base64 string on the renderer thread.
 */
export const parseAttachment = async (params: {
  name: string
  mime: string
  data: ArrayBuffer | ArrayBufferView
}): Promise<ParsedAttachment> => {
  const name = String(params?.name ?? '').trim()
  const mime = String(params?.mime ?? '').trim()
  if (!name || name.length > 255 || /[\u0000-\u001f\u007f]/u.test(name)) {
    throw new Error('The attachment filename is invalid.')
  }
  if (mime.length > 200 || /[\u0000-\u001f\u007f]/u.test(mime)) {
    throw new Error('The attachment content type is invalid.')
  }
  const data = params?.data
  if (!(data instanceof ArrayBuffer) && !ArrayBuffer.isView(data)) {
    throw new Error('The attachment data is invalid.')
  }
  const buf =
    data instanceof ArrayBuffer
      ? Buffer.from(data)
      : Buffer.from(data.buffer, data.byteOffset, data.byteLength)
  if (!buf.length || buf.length > MAX_ATTACHMENT_INPUT_BYTES) {
    throw new Error('The attachment must be between 1 byte and 20 MiB.')
  }
  const ext = (name.split('.').pop() ?? '').toLowerCase()
  const mimeExtension =
    mime === 'application/pdf'
      ? 'pdf'
      : mime === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
        ? 'docx'
        : mime === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
          ? 'xlsx'
          : mime === 'application/vnd.openxmlformats-officedocument.presentationml.presentation'
            ? 'pptx'
            : undefined
  const structuredExtension = STRUCTURED_DOCUMENT_EXTENSIONS.has(ext) ? ext : mimeExtension

  try {
    if (structuredExtension) {
      return await parseStructuredDocument(
        buf,
        STRUCTURED_DOCUMENT_EXTENSIONS.has(ext) ? name : `${name}.${structuredExtension}`
      )
    }
    // Legacy .doc is a binary OLE format; mammoth cannot read it.
    if (ext === 'doc') {
      return {
        text: '',
        empty: true,
        note: 'legacy .doc is not supported — save as .docx or PDF'
      }
    }
    return parseText(buf)
  } catch (e: any) {
    log.error(`failed to parse attachment ${name}:`, e?.message ?? e)
    throw new Error(`Could not read ${name}: ${e?.message ?? 'unsupported or corrupt file'}`)
  }
}
