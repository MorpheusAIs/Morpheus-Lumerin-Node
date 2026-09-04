import {
  AlignmentType,
  BorderStyle,
  Document,
  Footer,
  HeadingLevel,
  LevelFormat,
  Packer,
  PageBreak,
  PageNumber,
  Paragraph,
  ShadingType,
  Table,
  TableCell,
  TableLayoutType,
  TableRow,
  TextRun,
  VerticalAlign,
  WidthType
} from 'docx'
import ExcelJS from 'exceljs'
import { PDFDocument, PDFFont, PDFPage, StandardFonts, rgb } from 'pdf-lib'
import PptxGenJS from 'pptxgenjs'

export type ProfessionalArtifactFormat = 'docx' | 'xlsx' | 'pptx' | 'pdf'
export type ProfessionalArtifactScalar = string | number | boolean | null

export interface ProfessionalArtifactTable {
  headers: string[]
  rows: ProfessionalArtifactScalar[][]
}

export type ProfessionalDocumentBlock =
  | { type: 'heading'; level: 1 | 2 | 3; text: string }
  | { type: 'paragraph'; text: string }
  | { type: 'bulletList'; items: string[] }
  | { type: 'numberedList'; items: string[] }
  | ({ type: 'table' } & ProfessionalArtifactTable)
  | { type: 'pageBreak' }

interface ProfessionalArtifactBaseRequest {
  filename?: string
  title: string
  accentColor?: string
}

export interface ProfessionalDocxRequest extends ProfessionalArtifactBaseRequest {
  format: 'docx'
  subtitle?: string
  pageSize?: 'letter' | 'a4'
  blocks: ProfessionalDocumentBlock[]
}

export interface ProfessionalPdfRequest extends ProfessionalArtifactBaseRequest {
  format: 'pdf'
  subtitle?: string
  pageSize?: 'letter' | 'a4'
  blocks: ProfessionalDocumentBlock[]
}

export interface ProfessionalWorkbookSheet {
  name: string
  headers: string[]
  rows: ProfessionalArtifactScalar[][]
  columnWidths?: number[]
  freezeHeader?: boolean
}

export interface ProfessionalXlsxRequest extends ProfessionalArtifactBaseRequest {
  format: 'xlsx'
  sheets: ProfessionalWorkbookSheet[]
}

export interface ProfessionalPresentationSlide {
  title: string
  subtitle?: string
  body?: string
  bullets?: string[]
  table?: ProfessionalArtifactTable
}

export interface ProfessionalPptxRequest extends ProfessionalArtifactBaseRequest {
  format: 'pptx'
  slides: ProfessionalPresentationSlide[]
}

export type ProfessionalArtifactRequest =
  | ProfessionalDocxRequest
  | ProfessionalXlsxRequest
  | ProfessionalPptxRequest
  | ProfessionalPdfRequest

export interface GeneratedProfessionalArtifact {
  format: ProfessionalArtifactFormat
  extension: '.docx' | '.xlsx' | '.pptx' | '.pdf'
  mimeType: string
  suggestedFilename: string
  sizeBytes: number
  buffer: Buffer
}

export const PROFESSIONAL_ARTIFACT_LIMITS = Object.freeze({
  inputBytes: 1024 * 1024,
  outputBytes: 32 * 1024 * 1024,
  totalTextCharacters: 250_000,
  documentBlocks: 250,
  documentTableRows: 500,
  documentTableColumns: 20,
  workbookSheets: 12,
  workbookRowsPerSheet: 5_000,
  workbookColumns: 100,
  totalCells: 100_000,
  presentationSlides: 40,
  presentationBulletsPerSlide: 12,
  presentationTableRows: 12,
  presentationTableColumns: 8,
  pdfPages: 100
})

type JsonObject = Record<string, unknown>

const FORMAT_METADATA: Record<
  ProfessionalArtifactFormat,
  { extension: GeneratedProfessionalArtifact['extension']; mimeType: string }
> = {
  docx: {
    extension: '.docx',
    mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
  },
  xlsx: {
    extension: '.xlsx',
    mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  },
  pptx: {
    extension: '.pptx',
    mimeType: 'application/vnd.openxmlformats-officedocument.presentationml.presentation'
  },
  pdf: { extension: '.pdf', mimeType: 'application/pdf' }
}

const DEFAULT_ACCENT = '0F766E'
const CONTROL_CHARACTERS = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f]/u
const WINDOWS_RESERVED_NAME = /^(con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\.|$)/iu
const SAFE_FILENAME = /^[\p{L}\p{N}][\p{L}\p{N} ._-]*$/u

class ValidationBudget {
  private textCharacters = 0
  private cells = 0

  addText(text: string, path: string): void {
    this.textCharacters += text.length
    if (this.textCharacters > PROFESSIONAL_ARTIFACT_LIMITS.totalTextCharacters) {
      throw new Error(
        `${path} exceeds the ${PROFESSIONAL_ARTIFACT_LIMITS.totalTextCharacters.toLocaleString()} total-character limit.`
      )
    }
  }

  addCells(count: number, path: string): void {
    this.cells += count
    if (this.cells > PROFESSIONAL_ARTIFACT_LIMITS.totalCells) {
      throw new Error(
        `${path} exceeds the ${PROFESSIONAL_ARTIFACT_LIMITS.totalCells.toLocaleString()} total-cell limit.`
      )
    }
  }
}

function assertJsonFriendly(value: unknown): void {
  const ancestors = new WeakSet<object>()
  let visitedNodes = 0

  const visit = (candidate: unknown, path: string, depth: number): void => {
    visitedNodes++
    if (visitedNodes > 200_000) throw new Error('request contains too many JSON values.')
    if (depth > 12) throw new Error(`${path} exceeds the maximum nesting depth.`)
    if (candidate === null || typeof candidate === 'string' || typeof candidate === 'boolean')
      return
    if (typeof candidate === 'number') {
      if (!Number.isFinite(candidate)) throw new Error(`${path} must be a finite number.`)
      return
    }
    if (typeof candidate !== 'object')
      throw new Error(`${path} must contain JSON-compatible values only.`)
    if (ancestors.has(candidate)) throw new Error(`${path} contains a circular reference.`)

    ancestors.add(candidate)
    if (Array.isArray(candidate)) {
      const keys = Reflect.ownKeys(candidate)
      const expectedKeys = new Set([...candidate.keys()].map(String).concat('length'))
      if (keys.some((key) => typeof key !== 'string' || !expectedKeys.has(key))) {
        throw new Error(`${path} must be a plain JSON array without custom properties.`)
      }
      for (let index = 0; index < candidate.length; index++) {
        if (!Object.prototype.hasOwnProperty.call(candidate, index)) {
          throw new Error(`${path}[${index}] must not be a sparse array item.`)
        }
        visit(candidate[index], `${path}[${index}]`, depth + 1)
      }
    } else {
      const prototype = Object.getPrototypeOf(candidate)
      if (prototype !== Object.prototype && prototype !== null) {
        throw new Error(`${path} must be a plain JSON object.`)
      }
      for (const key of Reflect.ownKeys(candidate)) {
        if (typeof key !== 'string') throw new Error(`${path} must not contain symbol keys.`)
        const descriptor = Object.getOwnPropertyDescriptor(candidate, key)
        if (!descriptor?.enumerable || !('value' in descriptor)) {
          throw new Error(`${path}.${key} must be an enumerable JSON data property.`)
        }
        visit(descriptor.value, `${path}.${key}`, depth + 1)
      }
    }
    ancestors.delete(candidate)
  }

  visit(value, 'request', 0)
  const serialized = JSON.stringify(value)
  if (Buffer.byteLength(serialized, 'utf8') > PROFESSIONAL_ARTIFACT_LIMITS.inputBytes) {
    throw new Error(
      `request exceeds the ${PROFESSIONAL_ARTIFACT_LIMITS.inputBytes / 1024} KiB input limit.`
    )
  }
}

function plainObject(value: unknown, path: string): JsonObject {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    throw new Error(`${path} must be an object.`)
  return value as JsonObject
}

function assertKeys(object: JsonObject, allowed: readonly string[], path: string): void {
  const allowedSet = new Set(allowed)
  const unexpected = Object.keys(object).filter((key) => !allowedSet.has(key))
  if (unexpected.length) throw new Error(`${path} contains unsupported field "${unexpected[0]}".`)
}

function arrayValue(value: unknown, path: string, minimum: number, maximum: number): unknown[] {
  if (!Array.isArray(value)) throw new Error(`${path} must be an array.`)
  if (value.length < minimum || value.length > maximum) {
    throw new Error(`${path} must contain between ${minimum} and ${maximum} items.`)
  }
  return value
}

function textValue(
  value: unknown,
  path: string,
  budget: ValidationBudget,
  options: { maximum: number; allowEmpty?: boolean; trim?: boolean }
): string {
  if (typeof value !== 'string') throw new Error(`${path} must be a string.`)
  const result = options.trim ? value.trim() : value
  if (!options.allowEmpty && !result.trim()) throw new Error(`${path} must not be empty.`)
  if (result.length > options.maximum)
    throw new Error(`${path} exceeds the ${options.maximum}-character limit.`)
  if (CONTROL_CHARACTERS.test(result))
    throw new Error(`${path} contains an unsupported control character.`)
  budget.addText(result, path)
  return result
}

function optionalText(
  value: unknown,
  path: string,
  budget: ValidationBudget,
  options: { maximum: number; trim?: boolean }
): string | undefined {
  if (value === undefined) return undefined
  return textValue(value, path, budget, options)
}

function booleanValue(value: unknown, path: string, fallback: boolean): boolean {
  if (value === undefined) return fallback
  if (typeof value !== 'boolean') throw new Error(`${path} must be a boolean.`)
  return value
}

function enumValue<T extends string>(value: unknown, allowed: readonly T[], path: string): T {
  if (typeof value !== 'string' || !allowed.includes(value as T)) {
    throw new Error(`${path} must be one of: ${allowed.join(', ')}.`)
  }
  return value as T
}

function accentValue(value: unknown): string | undefined {
  if (value === undefined) return undefined
  if (typeof value !== 'string' || !/^#[0-9a-f]{6}$/iu.test(value)) {
    throw new Error('request.accentColor must be a six-digit hex color such as #0F766E.')
  }
  return value.toUpperCase()
}

function filenameValue(
  value: unknown,
  format: ProfessionalArtifactFormat,
  budget: ValidationBudget
): string | undefined {
  if (value === undefined) return undefined
  let filename = textValue(value, 'request.filename', budget, { maximum: 100, trim: true })
  const extension = FORMAT_METADATA[format].extension
  if (filename.toLowerCase().endsWith(extension)) filename = filename.slice(0, -extension.length)
  if (
    !SAFE_FILENAME.test(filename) ||
    filename.endsWith('.') ||
    filename.endsWith(' ') ||
    WINDOWS_RESERVED_NAME.test(filename)
  ) {
    throw new Error(
      'request.filename must be a safe filename without a path, reserved name, or special characters.'
    )
  }
  return filename
}

function scalarValue(
  value: unknown,
  path: string,
  budget: ValidationBudget,
  maximumText: number
): ProfessionalArtifactScalar {
  if (value === null || typeof value === 'boolean') return value
  if (typeof value === 'number') {
    if (!Number.isFinite(value) || Math.abs(value) > 1_000_000_000_000_000) {
      throw new Error(`${path} must be a finite number between -1e15 and 1e15.`)
    }
    return value
  }
  return textValue(value, path, budget, { maximum: maximumText, allowEmpty: true })
}

function validateTable(
  value: unknown,
  path: string,
  budget: ValidationBudget,
  limits: { rows: number; columns: number; cellCharacters: number }
): ProfessionalArtifactTable {
  const object = plainObject(value, path)
  assertKeys(object, ['headers', 'rows'], path)
  const rawHeaders = arrayValue(object.headers, `${path}.headers`, 1, limits.columns)
  const headers = rawHeaders.map((header, index) =>
    textValue(header, `${path}.headers[${index}]`, budget, { maximum: 250, trim: true })
  )
  const rawRows = arrayValue(object.rows, `${path}.rows`, 0, limits.rows)
  budget.addCells(headers.length * (rawRows.length + 1), path)
  const rows = rawRows.map((row, rowIndex) => {
    const values = arrayValue(row, `${path}.rows[${rowIndex}]`, headers.length, headers.length)
    return values.map((cell, columnIndex) =>
      scalarValue(cell, `${path}.rows[${rowIndex}][${columnIndex}]`, budget, limits.cellCharacters)
    )
  })
  return { headers, rows }
}

function validateDocumentBlock(
  value: unknown,
  path: string,
  budget: ValidationBudget,
  format: 'docx' | 'pdf'
): ProfessionalDocumentBlock {
  const object = plainObject(value, path)
  const type = enumValue(
    object.type,
    ['heading', 'paragraph', 'bulletList', 'numberedList', 'table', 'pageBreak'],
    `${path}.type`
  )
  if (type === 'heading') {
    assertKeys(object, ['type', 'level', 'text'], path)
    if (object.level !== 1 && object.level !== 2 && object.level !== 3) {
      throw new Error(`${path}.level must be one of: 1, 2, 3.`)
    }
    const level = object.level
    return {
      type,
      level,
      text: textValue(object.text, `${path}.text`, budget, { maximum: 500, trim: true })
    }
  }
  if (type === 'paragraph') {
    assertKeys(object, ['type', 'text'], path)
    return { type, text: textValue(object.text, `${path}.text`, budget, { maximum: 20_000 }) }
  }
  if (type === 'bulletList' || type === 'numberedList') {
    assertKeys(object, ['type', 'items'], path)
    const items = arrayValue(object.items, `${path}.items`, 1, 200).map((item, index) =>
      textValue(item, `${path}.items[${index}]`, budget, { maximum: 2_000 })
    )
    return { type, items }
  }
  if (type === 'table') {
    assertKeys(object, ['type', 'headers', 'rows'], path)
    const table = validateTable(
      { headers: object.headers, rows: object.rows },
      path,
      budget,
      format === 'pdf'
        ? { rows: 100, columns: 8, cellCharacters: 300 }
        : {
            rows: PROFESSIONAL_ARTIFACT_LIMITS.documentTableRows,
            columns: PROFESSIONAL_ARTIFACT_LIMITS.documentTableColumns,
            cellCharacters: 2_000
          }
    )
    return { type, ...table }
  }
  assertKeys(object, ['type'], path)
  return { type: 'pageBreak' }
}

function validateDocumentRequest(
  object: JsonObject,
  format: 'docx' | 'pdf',
  budget: ValidationBudget
): ProfessionalDocxRequest | ProfessionalPdfRequest {
  assertKeys(
    object,
    ['format', 'filename', 'title', 'subtitle', 'accentColor', 'pageSize', 'blocks'],
    'request'
  )
  const title = textValue(object.title, 'request.title', budget, { maximum: 250, trim: true })
  const filename = filenameValue(object.filename, format, budget)
  const subtitle = optionalText(object.subtitle, 'request.subtitle', budget, {
    maximum: 500,
    trim: true
  })
  const accentColor = accentValue(object.accentColor)
  const pageSize =
    object.pageSize === undefined
      ? undefined
      : enumValue(object.pageSize, ['letter', 'a4'], 'request.pageSize')
  const rawBlocks = arrayValue(
    object.blocks,
    'request.blocks',
    1,
    PROFESSIONAL_ARTIFACT_LIMITS.documentBlocks
  )
  const blocks = rawBlocks.map((block, index) =>
    validateDocumentBlock(block, `request.blocks[${index}]`, budget, format)
  )
  const pageBreaks = blocks.filter((block) => block.type === 'pageBreak').length
  if (format === 'pdf' && pageBreaks >= PROFESSIONAL_ARTIFACT_LIMITS.pdfPages) {
    throw new Error(
      `request.blocks cannot request ${PROFESSIONAL_ARTIFACT_LIMITS.pdfPages} or more PDF page breaks.`
    )
  }
  return { format, filename, title, subtitle, accentColor, pageSize, blocks }
}

function validateWorkbookRequest(
  object: JsonObject,
  budget: ValidationBudget
): ProfessionalXlsxRequest {
  assertKeys(object, ['format', 'filename', 'title', 'accentColor', 'sheets'], 'request')
  const format = 'xlsx' as const
  const title = textValue(object.title, 'request.title', budget, { maximum: 250, trim: true })
  const filename = filenameValue(object.filename, format, budget)
  const accentColor = accentValue(object.accentColor)
  const rawSheets = arrayValue(
    object.sheets,
    'request.sheets',
    1,
    PROFESSIONAL_ARTIFACT_LIMITS.workbookSheets
  )
  const usedNames = new Set<string>()
  const sheets = rawSheets.map((sheet, sheetIndex): ProfessionalWorkbookSheet => {
    const path = `request.sheets[${sheetIndex}]`
    const item = plainObject(sheet, path)
    assertKeys(item, ['name', 'headers', 'rows', 'columnWidths', 'freezeHeader'], path)
    const name = textValue(item.name, `${path}.name`, budget, { maximum: 31, trim: true })
    if (/[\[\]:*?/\\]/u.test(name) || name.startsWith("'") || name.endsWith("'")) {
      throw new Error(`${path}.name contains a character Excel does not allow in sheet names.`)
    }
    const normalizedName = name.toLocaleLowerCase('en-US')
    if (usedNames.has(normalizedName))
      throw new Error(`${path}.name duplicates another sheet name.`)
    usedNames.add(normalizedName)
    const table = validateTable({ headers: item.headers, rows: item.rows }, path, budget, {
      rows: PROFESSIONAL_ARTIFACT_LIMITS.workbookRowsPerSheet,
      columns: PROFESSIONAL_ARTIFACT_LIMITS.workbookColumns,
      cellCharacters: 10_000
    })
    let columnWidths: number[] | undefined
    if (item.columnWidths !== undefined) {
      const rawWidths = arrayValue(
        item.columnWidths,
        `${path}.columnWidths`,
        table.headers.length,
        table.headers.length
      )
      columnWidths = rawWidths.map((width, index) => {
        if (!Number.isInteger(width) || (width as number) < 6 || (width as number) > 60) {
          throw new Error(`${path}.columnWidths[${index}] must be an integer from 6 through 60.`)
        }
        return width as number
      })
    }
    return {
      name,
      headers: table.headers,
      rows: table.rows,
      columnWidths,
      freezeHeader: booleanValue(item.freezeHeader, `${path}.freezeHeader`, true)
    }
  })
  return { format, filename, title, accentColor, sheets }
}

function validatePresentationRequest(
  object: JsonObject,
  budget: ValidationBudget
): ProfessionalPptxRequest {
  assertKeys(object, ['format', 'filename', 'title', 'accentColor', 'slides'], 'request')
  const format = 'pptx' as const
  const title = textValue(object.title, 'request.title', budget, { maximum: 250, trim: true })
  const filename = filenameValue(object.filename, format, budget)
  const accentColor = accentValue(object.accentColor)
  const rawSlides = arrayValue(
    object.slides,
    'request.slides',
    1,
    PROFESSIONAL_ARTIFACT_LIMITS.presentationSlides
  )
  const slides = rawSlides.map((slide, slideIndex): ProfessionalPresentationSlide => {
    const path = `request.slides[${slideIndex}]`
    const item = plainObject(slide, path)
    assertKeys(item, ['title', 'subtitle', 'body', 'bullets', 'table'], path)
    const slideTitle = textValue(item.title, `${path}.title`, budget, { maximum: 100, trim: true })
    const subtitle = optionalText(item.subtitle, `${path}.subtitle`, budget, {
      maximum: 220,
      trim: true
    })
    const body = optionalText(item.body, `${path}.body`, budget, { maximum: 1_800 })
    let bullets: string[] | undefined
    if (item.bullets !== undefined) {
      bullets = arrayValue(
        item.bullets,
        `${path}.bullets`,
        1,
        PROFESSIONAL_ARTIFACT_LIMITS.presentationBulletsPerSlide
      ).map((bullet, bulletIndex) =>
        textValue(bullet, `${path}.bullets[${bulletIndex}]`, budget, { maximum: 240 })
      )
    }
    const table =
      item.table === undefined
        ? undefined
        : validateTable(item.table, `${path}.table`, budget, {
            rows: PROFESSIONAL_ARTIFACT_LIMITS.presentationTableRows,
            columns: PROFESSIONAL_ARTIFACT_LIMITS.presentationTableColumns,
            cellCharacters: 80
          })
    if (table && (body || bullets)) {
      throw new Error(
        `${path} cannot combine a table with body or bullet content; use a separate slide.`
      )
    }
    if (!table) {
      const bodyLines = body
        ? body
            .split('\n')
            .reduce((count, line) => count + Math.max(1, Math.ceil(line.length / 90)), 0)
        : 0
      const bulletLines =
        bullets?.reduce((count, bullet) => count + Math.max(1, Math.ceil(bullet.length / 82)), 0) ??
        0
      if (bodyLines + bulletLines > 20) {
        throw new Error(
          `${path} contains too much text for a readable 16:9 slide; split it into multiple slides.`
        )
      }
    }
    return { title: slideTitle, subtitle, body, bullets, table }
  })
  return { format, filename, title, accentColor, slides }
}

export function validateProfessionalArtifactRequest(input: unknown): ProfessionalArtifactRequest {
  assertJsonFriendly(input)
  const object = plainObject(input, 'request')
  const format = enumValue(object.format, ['docx', 'xlsx', 'pptx', 'pdf'], 'request.format')
  const budget = new ValidationBudget()
  if (format === 'docx' || format === 'pdf') return validateDocumentRequest(object, format, budget)
  if (format === 'xlsx') return validateWorkbookRequest(object, budget)
  return validatePresentationRequest(object, budget)
}

function accentHex(request: ProfessionalArtifactRequest): string {
  return request.accentColor?.slice(1) ?? DEFAULT_ACCENT
}

function scalarText(value: ProfessionalArtifactScalar): string {
  if (value === null) return ''
  if (typeof value === 'boolean') return value ? 'Yes' : 'No'
  return String(value)
}

function titleFilename(title: string): string {
  const normalized = title
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/gu, '')
    .replace(/[^a-zA-Z0-9]+/gu, '-')
    .replace(/^-+|-+$/gu, '')
    .slice(0, 80)
    .replace(/-+$/gu, '')
  return normalized || 'artifact'
}

function finishArtifact(
  request: ProfessionalArtifactRequest,
  binary: Buffer | Uint8Array | ArrayBuffer
): GeneratedProfessionalArtifact {
  const buffer = Buffer.isBuffer(binary)
    ? Buffer.from(binary)
    : binary instanceof ArrayBuffer
      ? Buffer.from(new Uint8Array(binary))
      : Buffer.from(binary)
  if (!buffer.length)
    throw new Error(`Generated ${request.format.toUpperCase()} artifact is empty.`)
  if (buffer.length > PROFESSIONAL_ARTIFACT_LIMITS.outputBytes) {
    throw new Error(
      `Generated ${request.format.toUpperCase()} artifact exceeds the ${PROFESSIONAL_ARTIFACT_LIMITS.outputBytes / (1024 * 1024)} MiB output limit.`
    )
  }
  const metadata = FORMAT_METADATA[request.format]
  const baseName = request.filename ?? titleFilename(request.title)
  return {
    format: request.format,
    extension: metadata.extension,
    mimeType: metadata.mimeType,
    suggestedFilename: `${baseName}${metadata.extension}`,
    sizeBytes: buffer.length,
    buffer
  }
}

function wordTextRuns(
  text: string,
  options: { bold?: boolean; color?: string; size?: number } = {}
): TextRun[] {
  return text.split('\n').map(
    (line, index) =>
      new TextRun({
        text: line,
        break: index === 0 ? undefined : 1,
        bold: options.bold,
        color: options.color,
        size: options.size,
        font: 'Arial'
      })
  )
}

function wordTable(table: ProfessionalArtifactTable, accent: string): Table {
  const usableWidth = 9_360
  const baseWidth = Math.floor(usableWidth / table.headers.length)
  const columnWidths = table.headers.map((_, index) =>
    index === table.headers.length - 1 ? usableWidth - baseWidth * index : baseWidth
  )
  const border = { style: BorderStyle.SINGLE, size: 4, color: 'CBD5E1' }
  const makeCell = (value: string, index: number, header: boolean): TableCell =>
    new TableCell({
      width: { size: columnWidths[index], type: WidthType.DXA },
      margins: { top: 120, right: 120, bottom: 120, left: 120 },
      verticalAlign: VerticalAlign.CENTER,
      shading: header ? { type: ShadingType.CLEAR, fill: accent, color: 'auto' } : undefined,
      children: [
        new Paragraph({
          spacing: { before: 0, after: 0, line: 240 },
          children: wordTextRuns(value, {
            bold: header,
            color: header ? 'FFFFFF' : '1F2937',
            size: 19
          })
        })
      ]
    })
  const rows = [
    new TableRow({
      tableHeader: true,
      cantSplit: true,
      children: table.headers.map((header, index) => makeCell(header, index, true))
    }),
    ...table.rows.map(
      (row) =>
        new TableRow({
          cantSplit: true,
          children: row.map((value, index) => makeCell(scalarText(value), index, false))
        })
    )
  ]
  return new Table({
    rows,
    width: { size: usableWidth, type: WidthType.DXA },
    columnWidths,
    layout: TableLayoutType.FIXED,
    margins: { top: 120, right: 120, bottom: 120, left: 120 },
    borders: {
      top: border,
      bottom: border,
      left: border,
      right: border,
      insideHorizontal: border,
      insideVertical: border
    }
  })
}

function wordBlocks(blocks: ProfessionalDocumentBlock[], accent: string): Array<Paragraph | Table> {
  const children: Array<Paragraph | Table> = []
  for (const block of blocks) {
    if (block.type === 'heading') {
      const headings = [
        HeadingLevel.HEADING_1,
        HeadingLevel.HEADING_2,
        HeadingLevel.HEADING_3
      ] as const
      children.push(
        new Paragraph({
          heading: headings[block.level - 1],
          spacing: { before: block.level === 1 ? 360 : 260, after: 120 },
          keepNext: true,
          children: wordTextRuns(block.text, {
            bold: true,
            color: accent,
            size: block.level === 1 ? 30 : block.level === 2 ? 26 : 23
          })
        })
      )
    } else if (block.type === 'paragraph') {
      children.push(
        new Paragraph({
          spacing: { after: 180, line: 276 },
          children: wordTextRuns(block.text, { color: '1F2937', size: 22 })
        })
      )
    } else if (block.type === 'bulletList') {
      for (const item of block.items) {
        children.push(
          new Paragraph({
            bullet: { level: 0 },
            spacing: { after: 90, line: 264 },
            children: wordTextRuns(item, { color: '1F2937', size: 22 })
          })
        )
      }
      children.push(new Paragraph({ spacing: { after: 80 }, children: [] }))
    } else if (block.type === 'numberedList') {
      for (const item of block.items) {
        children.push(
          new Paragraph({
            numbering: { reference: 'professional-numbering', level: 0 },
            spacing: { after: 90, line: 264 },
            children: wordTextRuns(item, { color: '1F2937', size: 22 })
          })
        )
      }
      children.push(new Paragraph({ spacing: { after: 80 }, children: [] }))
    } else if (block.type === 'table') {
      children.push(
        wordTable(block, accent),
        new Paragraph({ spacing: { after: 180 }, children: [] })
      )
    } else {
      children.push(new Paragraph({ children: [new PageBreak()] }))
    }
  }
  return children
}

async function buildDocx(request: ProfessionalDocxRequest): Promise<Buffer> {
  const accent = accentHex(request)
  const page =
    request.pageSize === 'a4'
      ? { width: 11_906, height: 16_838 }
      : { width: 12_240, height: 15_840 }
  const children: Array<Paragraph | Table> = [
    new Paragraph({
      spacing: { before: 0, after: request.subtitle ? 100 : 280 },
      children: wordTextRuns(request.title, { bold: true, color: accent, size: 44 })
    })
  ]
  if (request.subtitle) {
    children.push(
      new Paragraph({
        spacing: { after: 300, line: 264 },
        children: wordTextRuns(request.subtitle, { color: '475569', size: 24 })
      })
    )
  }
  children.push(...wordBlocks(request.blocks, accent))
  const document = new Document({
    title: request.title,
    subject: 'Morpheus Cowork artifact',
    creator: 'Morpheus',
    lastModifiedBy: 'Morpheus',
    styles: {
      default: {
        document: {
          run: { font: 'Arial', size: 22, color: '1F2937' },
          paragraph: { spacing: { after: 180, line: 276 } }
        }
      },
      paragraphStyles: [
        {
          id: 'Heading1',
          name: 'Heading 1',
          basedOn: 'Normal',
          next: 'Normal',
          quickFormat: true,
          run: { font: 'Arial', size: 30, bold: true, color: accent },
          paragraph: { spacing: { before: 360, after: 120 }, keepNext: true, outlineLevel: 0 }
        },
        {
          id: 'Heading2',
          name: 'Heading 2',
          basedOn: 'Normal',
          next: 'Normal',
          quickFormat: true,
          run: { font: 'Arial', size: 26, bold: true, color: accent },
          paragraph: { spacing: { before: 260, after: 100 }, keepNext: true, outlineLevel: 1 }
        },
        {
          id: 'Heading3',
          name: 'Heading 3',
          basedOn: 'Normal',
          next: 'Normal',
          quickFormat: true,
          run: { font: 'Arial', size: 23, bold: true, color: accent },
          paragraph: { spacing: { before: 220, after: 80 }, keepNext: true, outlineLevel: 2 }
        }
      ]
    },
    numbering: {
      config: [
        {
          reference: 'professional-numbering',
          levels: [
            {
              level: 0,
              format: LevelFormat.DECIMAL,
              text: '%1.',
              alignment: AlignmentType.START,
              style: { paragraph: { indent: { left: 720, hanging: 360 } } }
            }
          ]
        }
      ]
    },
    sections: [
      {
        properties: {
          page: {
            size: page,
            margin: { top: 1_080, right: 1_440, bottom: 1_080, left: 1_440 }
          }
        },
        footers: {
          default: new Footer({
            children: [
              new Paragraph({
                alignment: AlignmentType.RIGHT,
                children: [
                  new TextRun({ text: 'Page ', color: '64748B', size: 18 }),
                  new TextRun({ children: [PageNumber.CURRENT], color: '64748B', size: 18 })
                ]
              })
            ]
          })
        },
        children
      }
    ]
  })
  return Packer.toBuffer(document)
}

function excelColumnWidth(sheet: ProfessionalWorkbookSheet, columnIndex: number): number {
  if (sheet.columnWidths) return sheet.columnWidths[columnIndex]
  const values = [
    sheet.headers[columnIndex],
    ...sheet.rows.slice(0, 500).map((row) => scalarText(row[columnIndex]))
  ]
  const longest = Math.max(...values.map((value) => Math.min(value.length, 50)))
  return Math.max(10, Math.min(42, longest + 2))
}

async function buildXlsx(request: ProfessionalXlsxRequest): Promise<Buffer> {
  const workbook = new ExcelJS.Workbook()
  const accent = accentHex(request)
  workbook.creator = 'Morpheus'
  workbook.lastModifiedBy = 'Morpheus'
  workbook.created = new Date(0)
  workbook.modified = new Date(0)
  workbook.subject = request.title
  workbook.title = request.title
  workbook.company = 'Morpheus'
  workbook.calcProperties.fullCalcOnLoad = false

  for (const sheet of request.sheets) {
    const worksheet = workbook.addWorksheet(sheet.name, {
      properties: { defaultRowHeight: 19 },
      pageSetup: {
        orientation: sheet.headers.length > 8 ? 'landscape' : 'portrait',
        fitToPage: true,
        fitToWidth: 1
      }
    })
    worksheet.views = sheet.freezeHeader ? [{ state: 'frozen', ySplit: 1, activeCell: 'A2' }] : []
    worksheet.addRow(sheet.headers)
    for (const row of sheet.rows) worksheet.addRow(row)
    worksheet.autoFilter = {
      from: { row: 1, column: 1 },
      to: { row: Math.max(1, sheet.rows.length + 1), column: sheet.headers.length }
    }
    worksheet.pageSetup.printTitlesRow = '1:1'

    const headerRow = worksheet.getRow(1)
    headerRow.height = 24
    headerRow.eachCell((cell) => {
      cell.font = { name: 'Arial', size: 11, bold: true, color: { argb: 'FFFFFFFF' } }
      cell.alignment = { vertical: 'middle', horizontal: 'left', wrapText: true }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: `FF${accent}` } }
      cell.border = { bottom: { style: 'thin', color: { argb: `FF${accent}` } } }
    })

    for (let rowIndex = 2; rowIndex <= sheet.rows.length + 1; rowIndex++) {
      const row = worksheet.getRow(rowIndex)
      row.eachCell({ includeEmpty: true }, (cell) => {
        cell.font = { name: 'Arial', size: 10, color: { argb: 'FF1F2937' } }
        cell.alignment = {
          vertical: 'middle',
          horizontal: typeof cell.value === 'number' ? 'right' : 'left',
          wrapText: true
        }
        cell.border = { bottom: { style: 'hair', color: { argb: 'FFE2E8F0' } } }
        if (rowIndex % 2 === 1) {
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF8FAFC' } }
        }
      })
    }
    sheet.headers.forEach((_, columnIndex) => {
      worksheet.getColumn(columnIndex + 1).width = excelColumnWidth(sheet, columnIndex)
    })
  }
  const binary = await workbook.xlsx.writeBuffer({ useStyles: true, useSharedStrings: true })
  return Buffer.from(binary as unknown as Uint8Array)
}

function pptScalar(value: ProfessionalArtifactScalar): string {
  return scalarText(value)
}

async function buildPptx(request: ProfessionalPptxRequest): Promise<Buffer> {
  const presentation = new PptxGenJS()
  const accent = accentHex(request)
  presentation.layout = 'LAYOUT_WIDE'
  presentation.author = 'Morpheus'
  presentation.company = 'Morpheus'
  presentation.subject = 'Morpheus Cowork artifact'
  presentation.title = request.title
  presentation.theme = {
    headFontFace: 'Arial',
    bodyFontFace: 'Arial'
  }

  request.slides.forEach((content, slideIndex) => {
    const slide = presentation.addSlide()
    slide.background = { color: 'F8FAFC' }
    const titleOnly = !content.body && !content.bullets && !content.table
    if (titleOnly) {
      slide.addText(content.title, {
        x: 0.9,
        y: content.subtitle ? 2.55 : 2.8,
        w: 11.53,
        h: 0.8,
        margin: 0,
        fontFace: 'Arial',
        fontSize: 50,
        bold: true,
        color: accent,
        align: 'center',
        valign: 'middle',
        breakLine: false
      })
      if (content.subtitle) {
        slide.addText(content.subtitle, {
          x: 1.35,
          y: 3.55,
          w: 10.63,
          h: 0.8,
          margin: 0,
          fontFace: 'Arial',
          fontSize: 24,
          color: '475569',
          align: 'center',
          valign: 'top',
          breakLine: false
        })
      }
    } else {
      slide.addText(content.title, {
        x: 0.72,
        y: 0.45,
        w: 11.9,
        h: 0.62,
        margin: 0,
        fontFace: 'Arial',
        fontSize: 35,
        bold: true,
        color: accent,
        valign: 'middle',
        breakLine: false
      })
      let y = 1.22
      if (content.subtitle) {
        slide.addText(content.subtitle, {
          x: 0.74,
          y,
          w: 11.85,
          h: 0.48,
          margin: 0,
          fontFace: 'Arial',
          fontSize: 20,
          color: '475569',
          breakLine: false
        })
        y += 0.65
      }
      if (content.table) {
        const headerRow: PptxGenJS.TableRow = content.table.headers.map((header) => ({
          text: header,
          options: {
            bold: true,
            color: 'FFFFFF',
            fill: { color: accent },
            valign: 'middle',
            margin: 0.08
          }
        }))
        const bodyRows: PptxGenJS.TableRow[] = content.table.rows.map((row, rowIndex) =>
          row.map((value) => ({
            text: pptScalar(value),
            options: {
              color: '1F2937',
              fill: { color: rowIndex % 2 === 0 ? 'FFFFFF' : 'F1F5F9' },
              valign: 'middle',
              margin: 0.08
            }
          }))
        )
        slide.addTable([headerRow, ...bodyRows], {
          x: 0.74,
          y,
          w: 11.85,
          h: Math.min(5.15, 0.45 * (content.table.rows.length + 1)),
          border: { type: 'solid', color: 'CBD5E1', pt: 0.6 },
          fontFace: 'Arial',
          fontSize: 16,
          color: '1F2937',
          margin: 0.08,
          valign: 'middle'
        })
      } else {
        if (content.body) {
          const bodyHeight = content.bullets ? 2.25 : 4.85
          slide.addText(content.body, {
            x: 0.82,
            y,
            w: 11.65,
            h: bodyHeight,
            margin: 0,
            fontFace: 'Arial',
            fontSize: 18,
            color: '1F2937',
            valign: 'top',
            breakLine: false,
            paraSpaceAfter: 10
          })
          y += bodyHeight + 0.18
        }
        if (content.bullets) {
          const runs: PptxGenJS.TextProps[] = content.bullets.map((bullet, index) => ({
            text: bullet,
            options: {
              bullet: { indent: 18 },
              breakLine: index < content.bullets!.length - 1,
              paraSpaceAfter: 10
            }
          }))
          slide.addText(runs, {
            x: 0.9,
            y,
            w: 11.45,
            h: content.body ? 2.4 : 4.9,
            margin: 0,
            fontFace: 'Arial',
            fontSize: 18,
            color: '1F2937',
            valign: 'top',
            breakLine: false
          })
        }
      }
    }
    slide.addText(String(slideIndex + 1), {
      x: 12.05,
      y: 7.08,
      w: 0.55,
      h: 0.2,
      margin: 0,
      fontFace: 'Arial',
      fontSize: 10,
      color: '64748B',
      align: 'right'
    })
  })
  const binary = await presentation.write({ outputType: 'uint8array', compression: true })
  if (typeof binary === 'string' || binary instanceof Blob)
    throw new Error('PPTX generator returned an unexpected output type.')
  return binary instanceof Uint8Array ? Buffer.from(binary) : Buffer.from(new Uint8Array(binary))
}

interface PdfWriterState {
  document: PDFDocument
  regular: PDFFont
  bold: PDFFont
  accent: ReturnType<typeof rgb>
  pageWidth: number
  pageHeight: number
  margin: number
  pages: PDFPage[]
  page: PDFPage
  cursorY: number
}

function pdfNewPage(state: PdfWriterState): void {
  if (state.pages.length >= PROFESSIONAL_ARTIFACT_LIMITS.pdfPages) {
    throw new Error(
      `Generated PDF exceeds the ${PROFESSIONAL_ARTIFACT_LIMITS.pdfPages}-page limit.`
    )
  }
  state.page = state.document.addPage([state.pageWidth, state.pageHeight])
  state.pages.push(state.page)
  state.cursorY = state.pageHeight - state.margin
}

function pdfEnsureSpace(state: PdfWriterState, height: number): void {
  if (height > state.pageHeight - state.margin * 2 - 24)
    throw new Error('PDF content item is too tall to fit on one page.')
  if (state.cursorY - height < state.margin + 24) pdfNewPage(state)
}

function pdfBreakLongWord(
  word: string,
  font: PDFFont,
  fontSize: number,
  maximumWidth: number
): string[] {
  const fragments: string[] = []
  let fragment = ''
  for (const character of word) {
    const candidate = `${fragment}${character}`
    if (fragment && font.widthOfTextAtSize(candidate, fontSize) > maximumWidth) {
      fragments.push(fragment)
      fragment = character
    } else {
      fragment = candidate
    }
  }
  if (fragment || !fragments.length) fragments.push(fragment)
  return fragments
}

function pdfWrap(text: string, font: PDFFont, fontSize: number, maximumWidth: number): string[] {
  const lines: string[] = []
  for (const paragraph of text.split('\n')) {
    if (!paragraph) {
      lines.push('')
      continue
    }
    let line = ''
    for (const rawWord of paragraph.split(/\s+/u)) {
      const words =
        font.widthOfTextAtSize(rawWord, fontSize) > maximumWidth
          ? pdfBreakLongWord(rawWord, font, fontSize, maximumWidth)
          : [rawWord]
      for (const word of words) {
        const candidate = line ? `${line} ${word}` : word
        if (line && font.widthOfTextAtSize(candidate, fontSize) > maximumWidth) {
          lines.push(line)
          line = word
        } else {
          line = candidate
        }
      }
    }
    lines.push(line)
  }
  return lines
}

function pdfAssertEncodable(text: string, font: PDFFont, path: string): void {
  try {
    font.encodeText(text)
  } catch {
    throw new Error(`${path} contains a character that the safe embedded PDF font cannot encode.`)
  }
}

function pdfDrawLines(
  state: PdfWriterState,
  lines: string[],
  options: {
    x: number
    font: PDFFont
    fontSize: number
    lineHeight: number
    color: ReturnType<typeof rgb>
  }
): void {
  for (const line of lines) {
    state.page.drawText(line, {
      x: options.x,
      y: state.cursorY - options.fontSize,
      size: options.fontSize,
      font: options.font,
      color: options.color
    })
    state.cursorY -= options.lineHeight
  }
}

function pdfDrawTextBlock(
  state: PdfWriterState,
  text: string,
  options: {
    font: PDFFont
    fontSize: number
    lineHeight: number
    before: number
    after: number
    color: ReturnType<typeof rgb>
    indent?: number
  }
): void {
  const indent = options.indent ?? 0
  const maximumWidth = state.pageWidth - state.margin * 2 - indent
  const lines = pdfWrap(text, options.font, options.fontSize, maximumWidth)
  if (state.cursorY - options.before - options.lineHeight < state.margin + 24) pdfNewPage(state)
  state.cursorY -= options.before
  for (const line of lines) {
    if (state.cursorY - options.lineHeight < state.margin + 24) pdfNewPage(state)
    pdfDrawLines(state, [line], {
      x: state.margin + indent,
      font: options.font,
      fontSize: options.fontSize,
      lineHeight: options.lineHeight,
      color: options.color
    })
  }
  state.cursorY -= options.after
}

function pdfTableLines(text: string, font: PDFFont, width: number): string[] {
  return pdfWrap(text, font, 9, width - 12)
}

function pdfDrawTableRow(
  state: PdfWriterState,
  values: string[],
  options: { header: boolean; widths: number[] }
): void {
  const font = options.header ? state.bold : state.regular
  const lines = values.map((value, index) => pdfTableLines(value, font, options.widths[index]))
  const rowHeight = Math.max(24, Math.max(...lines.map((cellLines) => cellLines.length)) * 11 + 12)
  pdfEnsureSpace(state, rowHeight)
  let x = state.margin
  for (let index = 0; index < values.length; index++) {
    state.page.drawRectangle({
      x,
      y: state.cursorY - rowHeight,
      width: options.widths[index],
      height: rowHeight,
      color: options.header ? state.accent : rgb(1, 1, 1),
      borderColor: rgb(0.8, 0.84, 0.88),
      borderWidth: 0.6
    })
    lines[index].forEach((line, lineIndex) => {
      state.page.drawText(line, {
        x: x + 6,
        y: state.cursorY - 11 - lineIndex * 11,
        size: 9,
        font,
        color: options.header ? rgb(1, 1, 1) : rgb(0.12, 0.16, 0.22)
      })
    })
    x += options.widths[index]
  }
  state.cursorY -= rowHeight
}

function pdfDrawTable(state: PdfWriterState, table: ProfessionalArtifactTable): void {
  const availableWidth = state.pageWidth - state.margin * 2
  const baseWidth = availableWidth / table.headers.length
  const widths = table.headers.map(() => baseWidth)
  pdfEnsureSpace(state, 48)
  pdfDrawTableRow(state, table.headers, { header: true, widths })
  for (const row of table.rows) {
    const values = row.map(scalarText)
    const font = state.regular
    const rowHeight = Math.max(
      24,
      Math.max(...values.map((value, index) => pdfTableLines(value, font, widths[index]).length)) *
        11 +
        12
    )
    if (state.cursorY - rowHeight < state.margin + 24) {
      pdfNewPage(state)
      pdfDrawTableRow(state, table.headers, { header: true, widths })
    }
    pdfDrawTableRow(state, values, { header: false, widths })
  }
  state.cursorY -= 14
}

async function buildPdf(request: ProfessionalPdfRequest): Promise<Buffer> {
  const document = await PDFDocument.create({ updateMetadata: false })
  const regular = await document.embedFont(StandardFonts.Helvetica)
  const bold = await document.embedFont(StandardFonts.HelveticaBold)
  const dimensions =
    request.pageSize === 'a4' ? { width: 595.28, height: 841.89 } : { width: 612, height: 792 }
  const accent = accentHex(request)
  const accentColor = rgb(
    Number.parseInt(accent.slice(0, 2), 16) / 255,
    Number.parseInt(accent.slice(2, 4), 16) / 255,
    Number.parseInt(accent.slice(4, 6), 16) / 255
  )
  const firstPage = document.addPage([dimensions.width, dimensions.height])
  const state: PdfWriterState = {
    document,
    regular,
    bold,
    accent: accentColor,
    pageWidth: dimensions.width,
    pageHeight: dimensions.height,
    margin: 54,
    pages: [firstPage],
    page: firstPage,
    cursorY: dimensions.height - 54
  }

  const allText = [
    request.title,
    request.subtitle ?? '',
    ...request.blocks.flatMap((block) => {
      if (block.type === 'heading' || block.type === 'paragraph') return [block.text]
      if (block.type === 'bulletList' || block.type === 'numberedList') return block.items
      if (block.type === 'table') return [...block.headers, ...block.rows.flat().map(scalarText)]
      return []
    })
  ]
  allText.forEach((text, index) =>
    pdfAssertEncodable(text, regular, `request text item ${index + 1}`)
  )

  document.setTitle(request.title, { showInWindowTitleBar: true })
  document.setAuthor('Morpheus')
  document.setCreator('Morpheus Cowork')
  document.setProducer('Morpheus Cowork')
  document.setSubject('Morpheus Cowork artifact')
  document.setCreationDate(new Date(0))
  document.setModificationDate(new Date(0))

  pdfDrawTextBlock(state, request.title, {
    font: bold,
    fontSize: 24,
    lineHeight: 29,
    before: 0,
    after: request.subtitle ? 6 : 20,
    color: accentColor
  })
  if (request.subtitle) {
    pdfDrawTextBlock(state, request.subtitle, {
      font: regular,
      fontSize: 12,
      lineHeight: 16,
      before: 0,
      after: 22,
      color: rgb(0.28, 0.35, 0.43)
    })
  }

  for (const block of request.blocks) {
    if (block.type === 'heading') {
      const size = block.level === 1 ? 17 : block.level === 2 ? 14 : 12
      pdfDrawTextBlock(state, block.text, {
        font: bold,
        fontSize: size,
        lineHeight: size + 4,
        before: block.level === 1 ? 14 : 10,
        after: 7,
        color: accentColor
      })
    } else if (block.type === 'paragraph') {
      pdfDrawTextBlock(state, block.text, {
        font: regular,
        fontSize: 10.5,
        lineHeight: 14.5,
        before: 0,
        after: 10,
        color: rgb(0.12, 0.16, 0.22)
      })
    } else if (block.type === 'bulletList' || block.type === 'numberedList') {
      block.items.forEach((item, index) => {
        pdfDrawTextBlock(state, `${block.type === 'bulletList' ? '-' : `${index + 1}.`} ${item}`, {
          font: regular,
          fontSize: 10.5,
          lineHeight: 14.5,
          before: 0,
          after: 4,
          color: rgb(0.12, 0.16, 0.22),
          indent: 12
        })
      })
      state.cursorY -= 6
    } else if (block.type === 'table') {
      pdfDrawTable(state, block)
    } else {
      pdfNewPage(state)
    }
  }

  state.pages.forEach((page, index) => {
    const label = `Page ${index + 1} of ${state.pages.length}`
    page.drawText(label, {
      x: state.pageWidth - state.margin - regular.widthOfTextAtSize(label, 8),
      y: 28,
      size: 8,
      font: regular,
      color: rgb(0.39, 0.45, 0.53)
    })
  })
  return Buffer.from(
    await document.save({
      useObjectStreams: false,
      addDefaultPage: false,
      updateFieldAppearances: false
    })
  )
}

export async function generateDocxArtifact(input: unknown): Promise<GeneratedProfessionalArtifact> {
  const request = validateProfessionalArtifactRequest(input)
  if (request.format !== 'docx') throw new Error('request.format must be docx.')
  return finishArtifact(request, await buildDocx(request))
}

export async function generateXlsxArtifact(input: unknown): Promise<GeneratedProfessionalArtifact> {
  const request = validateProfessionalArtifactRequest(input)
  if (request.format !== 'xlsx') throw new Error('request.format must be xlsx.')
  return finishArtifact(request, await buildXlsx(request))
}

export async function generatePptxArtifact(input: unknown): Promise<GeneratedProfessionalArtifact> {
  const request = validateProfessionalArtifactRequest(input)
  if (request.format !== 'pptx') throw new Error('request.format must be pptx.')
  return finishArtifact(request, await buildPptx(request))
}

export async function generatePdfArtifact(input: unknown): Promise<GeneratedProfessionalArtifact> {
  const request = validateProfessionalArtifactRequest(input)
  if (request.format !== 'pdf') throw new Error('request.format must be pdf.')
  return finishArtifact(request, await buildPdf(request))
}

export async function generateProfessionalArtifact(
  input: unknown
): Promise<GeneratedProfessionalArtifact> {
  const request = validateProfessionalArtifactRequest(input)
  if (request.format === 'docx') return finishArtifact(request, await buildDocx(request))
  if (request.format === 'xlsx') return finishArtifact(request, await buildXlsx(request))
  if (request.format === 'pptx') return finishArtifact(request, await buildPptx(request))
  return finishArtifact(request, await buildPdf(request))
}
