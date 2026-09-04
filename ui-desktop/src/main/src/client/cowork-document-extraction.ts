import ExcelJS from 'exceljs'
import * as yauzl from 'yauzl'

export type CoworkDocumentFormat = 'pdf' | 'docx' | 'xlsx' | 'pptx'
export type CoworkDocumentSectionKind = 'page' | 'document' | 'sheet' | 'slide'

export interface CoworkDocumentSection {
  kind: CoworkDocumentSectionKind
  index: number
  title: string
  text: string
  rowCount?: number
  cellCount?: number
}

export interface CoworkDocumentExtraction {
  format: CoworkDocumentFormat
  sourceName: string
  text: string
  sections: CoworkDocumentSection[]
  metadata: {
    inputBytes: number
    extractedCharacters: number
    extractedTextBytes: number
    pageCount?: number
    sheetCount?: number
    slideCount?: number
    rowCount?: number
    cellCount?: number
  }
  empty: boolean
  truncated: boolean
  warnings: string[]
}

export type CoworkDocumentExtractionErrorCode =
  | 'invalid-input'
  | 'unsupported-format'
  | 'input-too-large'
  | 'encrypted'
  | 'unsafe-content'
  | 'limit-exceeded'
  | 'malformed'

export class CoworkDocumentExtractionError extends Error {
  readonly code: CoworkDocumentExtractionErrorCode

  constructor(code: CoworkDocumentExtractionErrorCode, message: string) {
    super(message)
    this.name = 'CoworkDocumentExtractionError'
    this.code = code
  }
}

/**
 * Hard limits for local document extraction. These are intentionally not
 * caller-configurable: model-authored tool arguments must not be able to turn
 * a preview into an unbounded archive or document parser.
 */
export const COWORK_DOCUMENT_EXTRACTION_LIMITS = Object.freeze({
  inputBytes: 20 * 1024 * 1024,
  outputBytes: 1024 * 1024,
  textCharacters: 120_000,
  textBytes: 200_000,
  pdfPages: 100,
  workbookSheets: 20,
  workbookRowsPerSheet: 5_000,
  workbookRowsTotal: 20_000,
  workbookColumnsPerSheet: 100,
  workbookCellsTotal: 100_000,
  presentationSlides: 100,
  archiveEntries: 2_048,
  archiveEntryBytes: 16 * 1024 * 1024,
  archiveInflatedBytes: 64 * 1024 * 1024,
  archiveRelationshipBytes: 1024 * 1024
})

const SUPPORTED_FORMATS = new Set<CoworkDocumentFormat>(['pdf', 'docx', 'xlsx', 'pptx'])
const CONTROL_CHARACTERS = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f]/gu
const ACTIVE_OOXML_PART =
  /(?:^|\/)(?:vbaProject\.bin|activeX(?:\/|$)|embeddings(?:\/|$)|externalLinks(?:\/|$)|customUI(?:\/|$))/iu
const EXTERNAL_RELATIONSHIP = /\bTargetMode\s*=\s*["']External["']/iu
const OOXML_ACTIVE_CONTENT_TYPE = /(?:macroEnabled|vbaProject|activeX)/iu
const OLE_COMPOUND_HEADER = Buffer.from([0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1])

interface ArchiveInspection {
  entries: Map<string, Buffer>
}

class TextBudget {
  private characters = 0
  private bytes = 0
  truncated = false

  take(raw: string): string {
    const normalized = normalizeExtractedText(raw)
    if (!normalized) return ''

    const remainingCharacters = COWORK_DOCUMENT_EXTRACTION_LIMITS.textCharacters - this.characters
    const remainingBytes = COWORK_DOCUMENT_EXTRACTION_LIMITS.textBytes - this.bytes
    if (remainingCharacters <= 0 || remainingBytes <= 0) {
      this.truncated = true
      return ''
    }

    let value = normalized.slice(0, remainingCharacters)
    if (Buffer.byteLength(value, 'utf8') > remainingBytes) {
      value = truncateUtf8(value, remainingBytes)
    }
    if (value.length < normalized.length) this.truncated = true

    this.characters += value.length
    this.bytes += Buffer.byteLength(value, 'utf8')
    return value
  }

  get characterCount(): number {
    return this.characters
  }

  get byteCount(): number {
    return this.bytes
  }
}

function truncateUtf8(value: string, maximumBytes: number): string {
  if (maximumBytes <= 0) return ''
  if (Buffer.byteLength(value, 'utf8') <= maximumBytes) return value

  let low = 0
  let high = value.length
  while (low < high) {
    const middle = Math.ceil((low + high) / 2)
    if (Buffer.byteLength(value.slice(0, middle), 'utf8') <= maximumBytes) low = middle
    else high = middle - 1
  }

  // Do not leave half of a UTF-16 surrogate pair at the boundary.
  let end = low
  if (end > 0 && end < value.length) {
    const last = value.charCodeAt(end - 1)
    if (last >= 0xd800 && last <= 0xdbff) end--
  }
  return value.slice(0, end)
}

function normalizeExtractedText(value: string): string {
  return value
    .replace(/\r\n?/gu, '\n')
    .replace(CONTROL_CHARACTERS, '')
    .replace(/[ \t]+\n/gu, '\n')
    .replace(/\n{4,}/gu, '\n\n\n')
    .trim()
}

function extractionError(
  code: CoworkDocumentExtractionErrorCode,
  message: string
): CoworkDocumentExtractionError {
  return new CoworkDocumentExtractionError(code, message)
}

function resolveSource(nameOrExtension: string): {
  format: CoworkDocumentFormat
  sourceName: string
} {
  if (typeof nameOrExtension !== 'string') {
    throw extractionError('invalid-input', 'A document filename or extension is required.')
  }
  const value = nameOrExtension.trim()
  if (!value || value.length > 255 || /[\u0000-\u001f\u007f]/u.test(value)) {
    throw extractionError('invalid-input', 'The document filename or extension is invalid.')
  }

  const finalComponent = value.split(/[\\/]/u).pop() ?? value
  const extensionOnly = /^\.?([a-z0-9]+)$/iu.exec(finalComponent)
  const extension = (
    finalComponent.includes('.')
      ? finalComponent.slice(finalComponent.lastIndexOf('.') + 1)
      : extensionOnly?.[1]
  )?.toLowerCase()
  if (!extension || !SUPPORTED_FORMATS.has(extension as CoworkDocumentFormat)) {
    throw extractionError(
      'unsupported-format',
      'Unsupported document format. Workspace can extract PDF, DOCX, XLSX, and PPTX files.'
    )
  }

  const format = extension as CoworkDocumentFormat
  return {
    format,
    sourceName:
      /^\.?[a-z0-9]+$/iu.test(finalComponent) && !finalComponent.includes('.')
        ? `document.${format}`
        : finalComponent.startsWith('.') && finalComponent.indexOf('.', 1) === -1
          ? `document.${format}`
          : finalComponent
  }
}

function assertInput(buffer: Buffer): void {
  if (!Buffer.isBuffer(buffer) || buffer.length === 0) {
    throw extractionError('invalid-input', 'The document must be a non-empty Buffer.')
  }
  if (buffer.length > COWORK_DOCUMENT_EXTRACTION_LIMITS.inputBytes) {
    throw extractionError(
      'input-too-large',
      `The document exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.inputBytes / (1024 * 1024)} MiB input limit.`
    )
  }
}

function looksLikeOleCompoundFile(buffer: Buffer): boolean {
  return (
    buffer.length >= OLE_COMPOUND_HEADER.length && buffer.subarray(0, 8).equals(OLE_COMPOUND_HEADER)
  )
}

function assertOoxmlContainer(buffer: Buffer, format: CoworkDocumentFormat): void {
  if (looksLikeOleCompoundFile(buffer)) {
    throw extractionError(
      'encrypted',
      `The ${format.toUpperCase()} file is encrypted or password protected and cannot be extracted.`
    )
  }
  if (buffer.length < 4 || buffer[0] !== 0x50 || buffer[1] !== 0x4b) {
    throw extractionError(
      'malformed',
      `The ${format.toUpperCase()} file is not a valid OOXML archive.`
    )
  }
}

function inspectOoxmlArchive(
  buffer: Buffer,
  selectedEntry: (name: string) => boolean
): Promise<ArchiveInspection> {
  return new Promise((resolve, reject) => {
    let settled = false
    let entryCount = 0
    let inflatedBytes = 0
    const entries = new Map<string, Buffer>()

    const fail = (error: CoworkDocumentExtractionError): void => {
      if (settled) return
      settled = true
      reject(error)
    }

    yauzl.fromBuffer(
      buffer,
      {
        lazyEntries: true,
        autoClose: true,
        decodeStrings: true,
        validateEntrySizes: true,
        strictFileNames: true
      },
      (openError, archive) => {
        if (openError || !archive) {
          fail(extractionError('malformed', 'The document is not a valid OOXML archive.'))
          return
        }

        archive.on('error', () => {
          fail(extractionError('malformed', 'The document contains a malformed OOXML archive.'))
        })
        archive.on('end', () => {
          if (settled) return
          settled = true
          resolve({ entries })
        })
        archive.on('entry', (entry) => {
          if (settled) return
          entryCount++
          if (entryCount > COWORK_DOCUMENT_EXTRACTION_LIMITS.archiveEntries) {
            fail(extractionError('limit-exceeded', 'The document contains too many archive parts.'))
            archive.close()
            return
          }

          const name = entry.fileName
          if (
            !name ||
            name.includes('\\') ||
            name.startsWith('/') ||
            name.split('/').some((part) => part === '..')
          ) {
            fail(extractionError('unsafe-content', 'The document contains an unsafe archive path.'))
            archive.close()
            return
          }
          if ((entry.generalPurposeBitFlag & 0x1) !== 0) {
            fail(extractionError('encrypted', 'The document contains encrypted archive parts.'))
            archive.close()
            return
          }
          if (!name.endsWith('/') && ACTIVE_OOXML_PART.test(name)) {
            fail(
              extractionError('unsafe-content', 'The document contains active or embedded content.')
            )
            archive.close()
            return
          }
          if (entry.uncompressedSize > COWORK_DOCUMENT_EXTRACTION_LIMITS.archiveEntryBytes) {
            fail(extractionError('limit-exceeded', 'A document archive part is too large.'))
            archive.close()
            return
          }
          if (
            entry.uncompressedSize > 1024 * 1024 &&
            entry.uncompressedSize > Math.max(1, entry.compressedSize) * 500
          ) {
            fail(
              extractionError('limit-exceeded', 'The document contains a suspicious archive part.')
            )
            archive.close()
            return
          }
          inflatedBytes += entry.uncompressedSize
          if (inflatedBytes > COWORK_DOCUMENT_EXTRACTION_LIMITS.archiveInflatedBytes) {
            fail(
              extractionError(
                'limit-exceeded',
                'The expanded document archive is too large to inspect.'
              )
            )
            archive.close()
            return
          }

          const relationship = name.endsWith('.rels')
          const contentTypes = name === '[Content_Types].xml'
          const shouldRead = relationship || contentTypes || selectedEntry(name)
          if (!shouldRead || name.endsWith('/')) {
            archive.readEntry()
            return
          }
          if (
            relationship &&
            entry.uncompressedSize > COWORK_DOCUMENT_EXTRACTION_LIMITS.archiveRelationshipBytes
          ) {
            fail(extractionError('limit-exceeded', 'A document relationship part is too large.'))
            archive.close()
            return
          }

          archive.openReadStream(entry, (streamError, stream) => {
            if (settled) return
            if (streamError || !stream) {
              fail(extractionError('malformed', 'A document archive part could not be read.'))
              archive.close()
              return
            }
            const chunks: Buffer[] = []
            let bytesRead = 0
            stream.on('data', (chunk: Buffer) => {
              bytesRead += chunk.length
              if (bytesRead <= entry.uncompressedSize) chunks.push(Buffer.from(chunk))
            })
            stream.on('error', () => {
              fail(extractionError('malformed', 'A document archive part is malformed.'))
              archive.close()
            })
            stream.on('end', () => {
              if (settled) return
              if (bytesRead !== entry.uncompressedSize) {
                fail(extractionError('malformed', 'A document archive part has an invalid size.'))
                archive.close()
                return
              }
              const value = Buffer.concat(chunks, bytesRead)
              const xml = value.toString('utf8')
              if (relationship && EXTERNAL_RELATIONSHIP.test(xml)) {
                fail(
                  extractionError(
                    'unsafe-content',
                    'The document contains an external relationship, which Workspace will not resolve.'
                  )
                )
                archive.close()
                return
              }
              if (contentTypes && OOXML_ACTIVE_CONTENT_TYPE.test(xml)) {
                fail(extractionError('unsafe-content', 'The document declares active content.'))
                archive.close()
                return
              }
              if (selectedEntry(name) || contentTypes) entries.set(name, value)
              archive.readEntry()
            })
          })
        })

        archive.readEntry()
      }
    )
  })
}

function ensureOoxmlParts(
  inspection: ArchiveInspection,
  format: CoworkDocumentFormat,
  requiredParts: string[]
): void {
  if (!inspection.entries.has('[Content_Types].xml')) {
    throw extractionError(
      'malformed',
      `The ${format.toUpperCase()} file has no content-types part.`
    )
  }
  for (const part of requiredParts) {
    if (!inspection.entries.has(part)) {
      throw extractionError(
        'malformed',
        `The ${format.toUpperCase()} file is missing a required document part.`
      )
    }
  }
}

async function extractPdf(
  buffer: Buffer,
  budget: TextBudget
): Promise<{ sections: CoworkDocumentSection[]; pageCount: number }> {
  if (!buffer.subarray(0, Math.min(buffer.length, 1024)).includes(Buffer.from('%PDF-'))) {
    throw extractionError('malformed', 'The PDF file has an invalid header.')
  }
  if (buffer.includes(Buffer.from('/Encrypt'))) {
    throw extractionError('encrypted', 'Encrypted or password-protected PDFs are not supported.')
  }

  let loadingTask: { destroy: () => Promise<void> } | undefined
  let document:
    | {
        numPages: number
        getPage: (pageNumber: number) => Promise<unknown>
        destroy: () => Promise<void>
      }
    | undefined
  try {
    const pdfjs = await import('pdfjs-dist/legacy/build/pdf.mjs')
    const task = pdfjs.getDocument({
      data: new Uint8Array(buffer),
      disableFontFace: true,
      isEvalSupported: false,
      useWorkerFetch: false,
      disableAutoFetch: true,
      disableStream: true,
      stopAtErrors: true
    })
    loadingTask = task
    document = (await task.promise) as typeof document
    if (!document) throw new Error('Missing PDF document.')
    if (document.numPages > COWORK_DOCUMENT_EXTRACTION_LIMITS.pdfPages) {
      throw extractionError(
        'limit-exceeded',
        `The PDF exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.pdfPages}-page limit.`
      )
    }

    const sections: CoworkDocumentSection[] = []
    for (let pageNumber = 1; pageNumber <= document.numPages; pageNumber++) {
      const page = (await document.getPage(pageNumber)) as {
        getTextContent: () => Promise<{ items: Array<{ str?: unknown; hasEOL?: boolean }> }>
        cleanup?: () => void
      }
      try {
        const content = await page.getTextContent()
        const fragments: string[] = []
        for (const item of content.items) {
          if (typeof item.str !== 'string') continue
          fragments.push(item.str)
          fragments.push(item.hasEOL ? '\n' : ' ')
        }
        sections.push({
          kind: 'page',
          index: pageNumber,
          title: `Page ${pageNumber}`,
          text: budget.take(fragments.join(''))
        })
      } finally {
        page.cleanup?.()
      }
    }
    return { sections, pageCount: document.numPages }
  } catch (error) {
    if (error instanceof CoworkDocumentExtractionError) throw error
    const message = error instanceof Error ? `${error.name} ${error.message}`.toLowerCase() : ''
    if (/password|encrypted/u.test(message)) {
      throw extractionError('encrypted', 'Encrypted or password-protected PDFs are not supported.')
    }
    throw extractionError('malformed', 'The PDF is malformed or could not be extracted safely.')
  } finally {
    try {
      await document?.destroy()
    } catch {
      // Cleanup failure must not replace the bounded extraction result/error.
    }
    try {
      await loadingTask?.destroy()
    } catch {
      // Cleanup failure must not replace the bounded extraction result/error.
    }
  }
}

async function extractDocx(
  buffer: Buffer,
  budget: TextBudget
): Promise<{ sections: CoworkDocumentSection[]; warnings: string[] }> {
  assertOoxmlContainer(buffer, 'docx')
  const archive = await inspectOoxmlArchive(
    buffer,
    (name) => name === 'word/document.xml' || name === 'word/styles.xml'
  )
  ensureOoxmlParts(archive, 'docx', ['word/document.xml'])

  try {
    const mammoth = await import('mammoth')
    const result = await mammoth.extractRawText({ buffer })
    const warnings = result.messages.length
      ? ['The DOCX parser reported recoverable document-structure warnings.']
      : []
    return {
      sections: [
        {
          kind: 'document',
          index: 1,
          title: 'Document',
          text: budget.take(result.value ?? '')
        }
      ],
      warnings
    }
  } catch (error) {
    if (error instanceof CoworkDocumentExtractionError) throw error
    const message = error instanceof Error ? error.message.toLowerCase() : ''
    if (/password|encrypted/u.test(message)) {
      throw extractionError(
        'encrypted',
        'Encrypted or password-protected DOCX files are not supported.'
      )
    }
    throw extractionError('malformed', 'The DOCX is malformed or could not be extracted safely.')
  }
}

function excelCellText(value: ExcelJS.CellValue): { text: string; formula: boolean } {
  if (value === null || value === undefined) return { text: '', formula: false }
  if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
    return { text: String(value), formula: false }
  }
  if (value instanceof Date) return { text: value.toISOString(), formula: false }
  if (typeof value !== 'object') return { text: String(value), formula: false }

  if ('formula' in value && typeof value.formula === 'string') {
    return { text: `[formula not evaluated] ${value.formula}`, formula: true }
  }
  if ('sharedFormula' in value && typeof value.sharedFormula === 'string') {
    return { text: `[shared formula not evaluated] ${value.sharedFormula}`, formula: true }
  }
  if ('richText' in value && Array.isArray(value.richText)) {
    return {
      text: value.richText
        .map((run) => (run && typeof run.text === 'string' ? run.text : ''))
        .join(''),
      formula: false
    }
  }
  if ('text' in value && typeof value.text === 'string') {
    // Hyperlink targets are deliberately not returned or resolved.
    return { text: value.text, formula: false }
  }
  if ('error' in value && typeof value.error === 'string') {
    return { text: `[spreadsheet error] ${value.error}`, formula: false }
  }
  return { text: '[unsupported cell value]', formula: false }
}

async function extractXlsx(
  buffer: Buffer,
  budget: TextBudget
): Promise<{
  sections: CoworkDocumentSection[]
  warnings: string[]
  sheetCount: number
  rowCount: number
  cellCount: number
}> {
  assertOoxmlContainer(buffer, 'xlsx')
  const archive = await inspectOoxmlArchive(
    buffer,
    (name) =>
      name === 'xl/workbook.xml' ||
      name === 'xl/styles.xml' ||
      name === 'xl/sharedStrings.xml' ||
      /^xl\/worksheets\/sheet\d+\.xml$/u.test(name)
  )
  ensureOoxmlParts(archive, 'xlsx', ['xl/workbook.xml'])

  try {
    const workbook = new ExcelJS.Workbook()
    await workbook.xlsx.load(buffer as unknown as ExcelJS.Buffer, {
      ignoreNodes: [
        'dataValidations',
        'extLst',
        'headerFooter',
        'legacyDrawing',
        'pageMargins',
        'pageSetup',
        'picture',
        'printOptions',
        'sheetPr',
        'sheetProtection',
        'tableParts'
      ]
    })
    if (workbook.worksheets.length > COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookSheets) {
      throw extractionError(
        'limit-exceeded',
        `The workbook exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookSheets}-sheet limit.`
      )
    }

    let totalRows = 0
    let totalCells = 0
    let sawFormula = false
    const sections: CoworkDocumentSection[] = []

    workbook.worksheets.forEach((worksheet, sheetIndex) => {
      let sheetRows = 0
      let sheetCells = 0
      const lines: string[] = []
      worksheet.eachRow({ includeEmpty: false }, (row, rowNumber) => {
        sheetRows++
        totalRows++
        if (
          rowNumber > COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookRowsPerSheet ||
          sheetRows > COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookRowsPerSheet
        ) {
          throw extractionError(
            'limit-exceeded',
            `A worksheet exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookRowsPerSheet.toLocaleString()}-row limit.`
          )
        }
        if (totalRows > COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookRowsTotal) {
          throw extractionError('limit-exceeded', 'The workbook contains too many rows.')
        }
        if (row.cellCount > COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookColumnsPerSheet) {
          throw extractionError(
            'limit-exceeded',
            `A worksheet exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookColumnsPerSheet}-column limit.`
          )
        }

        let maximumColumn = 0
        row.eachCell({ includeEmpty: false }, (cell, columnNumber) => {
          maximumColumn = Math.max(maximumColumn, columnNumber)
          sheetCells++
          totalCells++
          if (columnNumber > COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookColumnsPerSheet) {
            throw extractionError(
              'limit-exceeded',
              `A worksheet exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookColumnsPerSheet}-column limit.`
            )
          }
          if (totalCells > COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookCellsTotal) {
            throw extractionError(
              'limit-exceeded',
              'The workbook contains too many populated cells.'
            )
          }
          if (excelCellText(cell.value).formula) sawFormula = true
        })

        const values: string[] = []
        for (let column = 1; column <= maximumColumn; column++) {
          values.push(excelCellText(row.getCell(column).value).text)
        }
        lines.push(values.join('\t').trimEnd())
      })

      sections.push({
        kind: 'sheet',
        index: sheetIndex + 1,
        title: worksheet.name || `Sheet ${sheetIndex + 1}`,
        text: budget.take(lines.join('\n')),
        rowCount: sheetRows,
        cellCount: sheetCells
      })
    })

    return {
      sections,
      warnings: sawFormula
        ? ['Spreadsheet formulas were returned as inert text and were not evaluated.']
        : [],
      sheetCount: workbook.worksheets.length,
      rowCount: totalRows,
      cellCount: totalCells
    }
  } catch (error) {
    if (error instanceof CoworkDocumentExtractionError) throw error
    const message = error instanceof Error ? error.message.toLowerCase() : ''
    if (/password|encrypted/u.test(message)) {
      throw extractionError(
        'encrypted',
        'Encrypted or password-protected XLSX files are not supported.'
      )
    }
    throw extractionError('malformed', 'The XLSX is malformed or could not be extracted safely.')
  }
}

function decodeXmlText(value: string): string {
  return value.replace(/&(?:#(x[0-9a-f]+|\d+)|amp|apos|gt|lt|quot);/giu, (entity, numeric) => {
    if (numeric) {
      const hexadecimal = numeric[0]?.toLowerCase() === 'x'
      const codePoint = Number.parseInt(
        hexadecimal ? numeric.slice(1) : numeric,
        hexadecimal ? 16 : 10
      )
      if (!Number.isSafeInteger(codePoint) || codePoint < 0 || codePoint > 0x10ffff) return ''
      try {
        return String.fromCodePoint(codePoint)
      } catch {
        return ''
      }
    }
    switch (entity.toLowerCase()) {
      case '&amp;':
        return '&'
      case '&apos;':
        return "'"
      case '&gt;':
        return '>'
      case '&lt;':
        return '<'
      case '&quot;':
        return '"'
      default:
        return ''
    }
  })
}

function pptxSlideText(xml: string): string {
  const values: string[] = []
  const textNode = /<(?:[a-z_][\w.-]*:)?t(?:\s[^>]*)?>([\s\S]*?)<\/(?:[a-z_][\w.-]*:)?t>/giu
  for (const match of xml.matchAll(textNode)) values.push(decodeXmlText(match[1]))
  return values.join(' ')
}

function slideNumber(name: string): number {
  return Number.parseInt(/^ppt\/slides\/slide(\d+)\.xml$/u.exec(name)?.[1] ?? '0', 10)
}

async function extractPptx(
  buffer: Buffer,
  budget: TextBudget
): Promise<{ sections: CoworkDocumentSection[]; slideCount: number }> {
  assertOoxmlContainer(buffer, 'pptx')
  const archive = await inspectOoxmlArchive(
    buffer,
    (name) =>
      name === 'ppt/presentation.xml' ||
      name === 'ppt/_rels/presentation.xml.rels' ||
      /^ppt\/slides\/slide\d+\.xml$/u.test(name)
  )
  ensureOoxmlParts(archive, 'pptx', ['ppt/presentation.xml'])

  const slideParts = [...archive.entries.keys()]
    .filter((name) => /^ppt\/slides\/slide\d+\.xml$/u.test(name))
    .sort((left, right) => slideNumber(left) - slideNumber(right))
  if (slideParts.length > COWORK_DOCUMENT_EXTRACTION_LIMITS.presentationSlides) {
    throw extractionError(
      'limit-exceeded',
      `The presentation exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.presentationSlides}-slide limit.`
    )
  }

  const sections = slideParts.map((name, index) => ({
    kind: 'slide' as const,
    index: index + 1,
    title: `Slide ${index + 1}`,
    text: budget.take(pptxSlideText(archive.entries.get(name)!.toString('utf8')))
  }))
  return { sections, slideCount: slideParts.length }
}

function composeText(sections: CoworkDocumentSection[]): string {
  return sections
    .filter((section) => section.text)
    .map((section) => `--- ${section.title} ---\n${section.text}`)
    .join('\n\n')
}

function assertBoundedOutput(result: CoworkDocumentExtraction): void {
  const bytes = Buffer.byteLength(JSON.stringify(result), 'utf8')
  if (bytes > COWORK_DOCUMENT_EXTRACTION_LIMITS.outputBytes) {
    throw extractionError(
      'limit-exceeded',
      `The extracted document exceeds the ${COWORK_DOCUMENT_EXTRACTION_LIMITS.outputBytes / 1024} KiB output limit.`
    )
  }
}

/**
 * Extracts inert text and bounded section metadata from a document Buffer.
 *
 * `nameOrExtension` accepts a filename (for example `report.pdf`) or a bare
 * extension (`pdf` or `.pdf`). Parsing stays entirely in the Electron main
 * process. OOXML archives are preflighted before their format parser runs;
 * encrypted, active-content, embedded-object, and external-relationship files
 * are rejected. Spreadsheet formulas are represented as plain text and are
 * never evaluated.
 */
export async function extractCoworkDocument(
  buffer: Buffer,
  nameOrExtension: string
): Promise<CoworkDocumentExtraction> {
  assertInput(buffer)
  const { format, sourceName } = resolveSource(nameOrExtension)
  const budget = new TextBudget()
  const warnings: string[] = []
  let sections: CoworkDocumentSection[]
  const metadata: CoworkDocumentExtraction['metadata'] = {
    inputBytes: buffer.length,
    extractedCharacters: 0,
    extractedTextBytes: 0
  }

  if (format === 'pdf') {
    const extracted = await extractPdf(buffer, budget)
    sections = extracted.sections
    metadata.pageCount = extracted.pageCount
  } else if (format === 'docx') {
    const extracted = await extractDocx(buffer, budget)
    sections = extracted.sections
    warnings.push(...extracted.warnings)
  } else if (format === 'xlsx') {
    const extracted = await extractXlsx(buffer, budget)
    sections = extracted.sections
    warnings.push(...extracted.warnings)
    metadata.sheetCount = extracted.sheetCount
    metadata.rowCount = extracted.rowCount
    metadata.cellCount = extracted.cellCount
  } else {
    const extracted = await extractPptx(buffer, budget)
    sections = extracted.sections
    metadata.slideCount = extracted.slideCount
  }

  if (budget.truncated) {
    warnings.push(
      `Extracted text was truncated at ${COWORK_DOCUMENT_EXTRACTION_LIMITS.textCharacters.toLocaleString()} characters or ${COWORK_DOCUMENT_EXTRACTION_LIMITS.textBytes.toLocaleString()} UTF-8 bytes.`
    )
  }
  metadata.extractedCharacters = budget.characterCount
  metadata.extractedTextBytes = budget.byteCount
  const text = composeText(sections)
  const result: CoworkDocumentExtraction = {
    format,
    sourceName,
    text,
    sections,
    metadata,
    empty: !text.trim(),
    truncated: budget.truncated,
    warnings
  }
  assertBoundedOutput(result)
  return result
}
