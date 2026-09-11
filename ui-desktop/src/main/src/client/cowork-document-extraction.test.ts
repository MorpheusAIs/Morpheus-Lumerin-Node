import ExcelJS from 'exceljs'
import { PDFDocument } from 'pdf-lib'
import PptxGenJS from 'pptxgenjs'
import { describe, expect, it } from 'vitest'
import {
  COWORK_DOCUMENT_EXTRACTION_LIMITS,
  CoworkDocumentExtractionError,
  extractCoworkDocument
} from './cowork-document-extraction'
import {
  generateDocxArtifact,
  generatePdfArtifact,
  generatePptxArtifact,
  generateXlsxArtifact
} from './cowork-professional-artifacts'

function expectCode(code: CoworkDocumentExtractionError['code']): (error: unknown) => boolean {
  return (error) => error instanceof CoworkDocumentExtractionError && error.code === code
}

async function workbookBuffer(workbook: ExcelJS.Workbook): Promise<Buffer> {
  return Buffer.from((await workbook.xlsx.writeBuffer()) as unknown as Uint8Array)
}

describe('Cowork document extraction', () => {
  it('extracts bounded PDF text into ordered page sections', async () => {
    const artifact = await generatePdfArtifact({
      format: 'pdf',
      title: 'Local Review',
      blocks: [
        { type: 'heading', level: 1, text: 'Overview' },
        { type: 'paragraph', text: 'First-page content stays local.' },
        { type: 'pageBreak' },
        { type: 'heading', level: 1, text: 'Decision' },
        { type: 'paragraph', text: 'Ship after verification.' }
      ]
    })

    const result = await extractCoworkDocument(artifact.buffer, 'review.pdf')

    expect(result).toMatchObject({
      format: 'pdf',
      sourceName: 'review.pdf',
      empty: false,
      truncated: false,
      metadata: { pageCount: 2, inputBytes: artifact.buffer.length }
    })
    expect(result.sections.map((section) => section.title)).toEqual(['Page 1', 'Page 2'])
    expect(result.text).toContain('Overview')
    expect(result.text).toContain('Ship after verification.')
    expect(Buffer.byteLength(JSON.stringify(result))).toBeLessThanOrEqual(
      COWORK_DOCUMENT_EXTRACTION_LIMITS.outputBytes
    )
  }, 20_000)

  it('extracts DOCX text as one inert document section', async () => {
    const artifact = await generateDocxArtifact({
      format: 'docx',
      title: 'Research Brief',
      blocks: [
        { type: 'heading', level: 1, text: 'Findings' },
        { type: 'paragraph', text: 'The source was reviewed without network access.' },
        { type: 'bulletList', items: ['Bounded input', 'Plain-text output'] }
      ]
    })

    const result = await extractCoworkDocument(artifact.buffer, '.docx')

    expect(result).toMatchObject({
      format: 'docx',
      sourceName: 'document.docx',
      empty: false,
      truncated: false
    })
    expect(result.sections).toHaveLength(1)
    expect(result.sections[0]).toMatchObject({ kind: 'document', index: 1, title: 'Document' })
    expect(result.text).toContain('Research Brief')
    expect(result.text).toContain('Plain-text output')
  }, 20_000)

  it('extracts XLSX sheets and represents formulas as text without evaluating them', async () => {
    const workbook = new ExcelJS.Workbook()
    const worksheet = workbook.addWorksheet('Forecast')
    worksheet.addRow(['Region', 'Value'])
    worksheet.addRow(['APAC', 42])
    worksheet.getCell('B3').value = {
      formula: 'WEBSERVICE("https://example.invalid/private")',
      result: 'cached-value'
    }
    const buffer = await workbookBuffer(workbook)

    const result = await extractCoworkDocument(buffer, 'forecast.xlsx')

    expect(result).toMatchObject({
      format: 'xlsx',
      metadata: { sheetCount: 1, rowCount: 3, cellCount: 5 }
    })
    expect(result.sections[0]).toMatchObject({
      kind: 'sheet',
      title: 'Forecast',
      rowCount: 3,
      cellCount: 5
    })
    expect(result.text).toContain('[formula not evaluated] WEBSERVICE')
    expect(result.text).not.toContain('cached-value')
    expect(result.warnings).toContain(
      'Spreadsheet formulas were returned as inert text and were not evaluated.'
    )
  }, 20_000)

  it('extracts native PPTX text into ordered slide sections', async () => {
    const artifact = await generatePptxArtifact({
      format: 'pptx',
      title: 'Launch Plan',
      slides: [
        { title: 'Launch Plan', subtitle: 'Local-first workflow' },
        { title: 'Next steps', bullets: ['Verify build', 'Publish notes'] }
      ]
    })

    const result = await extractCoworkDocument(artifact.buffer, 'deck.pptx')

    expect(result).toMatchObject({
      format: 'pptx',
      metadata: { slideCount: 2 },
      empty: false,
      truncated: false
    })
    expect(result.sections.map((section) => section.title)).toEqual(['Slide 1', 'Slide 2'])
    expect(result.sections[0].text).toContain('Launch Plan')
    expect(result.sections[1].text).toContain('Verify build')
  }, 20_000)

  it('supports generated workbooks and preserves sheet structure', async () => {
    const artifact = await generateXlsxArtifact({
      format: 'xlsx',
      title: 'Results',
      sheets: [
        { name: 'Summary', headers: ['Metric', 'Value'], rows: [['Latency', 125]] },
        { name: 'Notes', headers: ['Note'], rows: [['Validated locally']] }
      ]
    })

    const result = await extractCoworkDocument(artifact.buffer, 'xlsx')

    expect(result.metadata).toMatchObject({ sheetCount: 2, rowCount: 4, cellCount: 6 })
    expect(result.sections.map((section) => section.title)).toEqual(['Summary', 'Notes'])
    expect(result.text).toContain('Latency\t125')
    expect(result.text).toContain('Validated locally')
  }, 20_000)

  it('truncates large extracted text within both text and serialized-output limits', async () => {
    // Sized off the limit rather than a literal, so raising the ceiling cannot
    // quietly turn this into a test of a document that never needed truncating.
    const paragraphCharacters = 20_000
    const artifact = await generateDocxArtifact({
      format: 'docx',
      title: 'Large Document',
      blocks: Array.from(
        {
          length:
            Math.ceil(COWORK_DOCUMENT_EXTRACTION_LIMITS.textCharacters / paragraphCharacters) + 2
        },
        () => ({
          type: 'paragraph' as const,
          text: '四'.repeat(paragraphCharacters)
        })
      )
    })

    const result = await extractCoworkDocument(artifact.buffer, 'large.docx')

    expect(result.truncated).toBe(true)
    expect(result.metadata.extractedCharacters).toBeLessThanOrEqual(
      COWORK_DOCUMENT_EXTRACTION_LIMITS.textCharacters
    )
    expect(result.metadata.extractedTextBytes).toBeLessThanOrEqual(
      COWORK_DOCUMENT_EXTRACTION_LIMITS.textBytes
    )
    expect(Buffer.byteLength(JSON.stringify(result))).toBeLessThanOrEqual(
      COWORK_DOCUMENT_EXTRACTION_LIMITS.outputBytes
    )
    expect(result.warnings.join(' ')).toMatch(/truncated/iu)
  }, 20_000)

  it('rejects unsupported, empty, oversized, and malformed inputs with stable error codes', async () => {
    await expect(extractCoworkDocument(Buffer.from('text'), 'notes.txt')).rejects.toSatisfy(
      expectCode('unsupported-format')
    )
    await expect(extractCoworkDocument(Buffer.alloc(0), 'empty.pdf')).rejects.toSatisfy(
      expectCode('invalid-input')
    )
    await expect(
      extractCoworkDocument(
        Buffer.alloc(COWORK_DOCUMENT_EXTRACTION_LIMITS.inputBytes + 1),
        'large.pdf'
      )
    ).rejects.toSatisfy(expectCode('input-too-large'))

    for (const extension of ['pdf', 'docx', 'xlsx', 'pptx']) {
      await expect(
        extractCoworkDocument(Buffer.from('not a real document'), extension)
      ).rejects.toSatisfy(expectCode('malformed'))
    }
  })

  it('rejects password-container signatures and encrypted PDF trailers', async () => {
    const ole = Buffer.concat([
      Buffer.from([0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1]),
      Buffer.alloc(64)
    ])
    for (const extension of ['docx', 'xlsx', 'pptx']) {
      await expect(extractCoworkDocument(ole, extension)).rejects.toSatisfy(expectCode('encrypted'))
    }
    await expect(
      extractCoworkDocument(Buffer.from('%PDF-1.7\n1 0 obj << /Encrypt 2 0 R >>'), 'locked.pdf')
    ).rejects.toSatisfy(expectCode('encrypted'))
  })

  it('rejects OOXML external relationships instead of resolving them', async () => {
    const workbook = new ExcelJS.Workbook()
    const worksheet = workbook.addWorksheet('Links')
    worksheet.getCell('A1').value = {
      text: 'Do not fetch this',
      hyperlink: 'https://example.invalid/private'
    }
    const buffer = await workbookBuffer(workbook)

    await expect(extractCoworkDocument(buffer, 'links.xlsx')).rejects.toSatisfy(
      expectCode('unsafe-content')
    )
  }, 20_000)

  it('enforces PDF page, worksheet row/column, and presentation slide limits', async () => {
    const pdf = await PDFDocument.create()
    for (let index = 0; index <= COWORK_DOCUMENT_EXTRACTION_LIMITS.pdfPages; index++) pdf.addPage()
    await expect(
      extractCoworkDocument(Buffer.from(await pdf.save()), 'pages.pdf')
    ).rejects.toSatisfy(expectCode('limit-exceeded'))

    const farRowWorkbook = new ExcelJS.Workbook()
    farRowWorkbook
      .addWorksheet('Far row')
      .getCell(COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookRowsPerSheet + 1, 1).value = 'x'
    await expect(
      extractCoworkDocument(await workbookBuffer(farRowWorkbook), 'far-row.xlsx')
    ).rejects.toSatisfy(expectCode('limit-exceeded'))

    const farColumnWorkbook = new ExcelJS.Workbook()
    farColumnWorkbook
      .addWorksheet('Far column')
      .getCell(1, COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookColumnsPerSheet + 1).value = 'x'
    await expect(
      extractCoworkDocument(await workbookBuffer(farColumnWorkbook), 'far-column.xlsx')
    ).rejects.toSatisfy(expectCode('limit-exceeded'))

    const presentation = new PptxGenJS()
    for (let index = 0; index <= COWORK_DOCUMENT_EXTRACTION_LIMITS.presentationSlides; index++) {
      presentation.addSlide().addText(`Slide ${index + 1}`, { x: 1, y: 1, w: 5, h: 1 })
    }
    const presentationBuffer = Buffer.from(
      (await presentation.write({ outputType: 'nodebuffer' })) as unknown as Uint8Array
    )
    await expect(extractCoworkDocument(presentationBuffer, 'slides.pptx')).rejects.toSatisfy(
      expectCode('limit-exceeded')
    )
  }, 30_000)

  it('enforces workbook-wide sheet and populated-cell limits', async () => {
    const manySheets = new ExcelJS.Workbook()
    for (let index = 0; index <= COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookSheets; index++) {
      manySheets.addWorksheet(`Sheet ${index + 1}`).getCell('A1').value = index
    }
    await expect(
      extractCoworkDocument(await workbookBuffer(manySheets), 'many-sheets.xlsx')
    ).rejects.toSatisfy(expectCode('limit-exceeded'))

    const manyCells = new ExcelJS.Workbook()
    const worksheet = manyCells.addWorksheet('Dense')
    const rowValues = Array.from(
      { length: COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookColumnsPerSheet },
      () => 'x'
    )
    const rowCount =
      Math.floor(
        COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookCellsTotal /
          COWORK_DOCUMENT_EXTRACTION_LIMITS.workbookColumnsPerSheet
      ) + 1
    for (let row = 0; row < rowCount; row++) worksheet.addRow(rowValues)
    await expect(
      extractCoworkDocument(await workbookBuffer(manyCells), 'many-cells.xlsx')
    ).rejects.toSatisfy(expectCode('limit-exceeded'))
  }, 30_000)
})
