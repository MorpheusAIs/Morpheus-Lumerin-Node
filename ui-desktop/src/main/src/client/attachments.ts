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

async function parsePdf(buf: Buffer): Promise<ParsedAttachment> {
  // Legacy build: the modern one assumes a browser environment.
  const pdfjs = await import('pdfjs-dist/legacy/build/pdf.mjs')

  const doc = await pdfjs.getDocument({
    data: new Uint8Array(buf),
    // Fonts and images are irrelevant when all we want is the text layer.
    disableFontFace: true,
    isEvalSupported: false
  }).promise

  const pages: string[] = []
  for (let i = 1; i <= doc.numPages; i++) {
    const page = await doc.getPage(i)
    const content = await page.getTextContent()
    const pageText = content.items
      .map((it: any) => ('str' in it ? it.str : ''))
      .join(' ')
      .replace(/\s+/g, ' ')
      .trim()
    if (pageText) {
      pages.push(`--- page ${i} ---\n${pageText}`)
    }
  }

  const joined = pages.join('\n\n')
  if (!joined.trim()) {
    // Almost always a scanned document: images of text, no text layer.
    return {
      text: '',
      empty: true,
      note: `${doc.numPages} page(s), no selectable text — this looks like a scanned PDF`
    }
  }

  const { text, note } = truncate(joined)
  return { text, note: note ?? `${doc.numPages} page(s)` }
}

async function parseDocx(buf: Buffer): Promise<ParsedAttachment> {
  const mammoth = await import('mammoth')
  const result = await mammoth.extractRawText({ buffer: buf })
  const { text, note } = truncate(result.value ?? '')
  return { text, note, empty: !text.trim() }
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
 * `data` is base64 — the renderer reads the file uniformly whether it came from
 * the picker, a drop, or a paste (a pasted image has no path), so bytes rather
 * than a path is the one shape that covers all three.
 */
export const parseAttachment = async (params: {
  name: string
  mime: string
  data: string
}): Promise<ParsedAttachment> => {
  const { name, mime } = params
  const buf = Buffer.from(params.data, 'base64')
  const ext = (name.split('.').pop() ?? '').toLowerCase()

  try {
    if (mime === 'application/pdf' || ext === 'pdf') {
      return await parsePdf(buf)
    }
    if (
      mime === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' ||
      ext === 'docx'
    ) {
      return await parseDocx(buf)
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
