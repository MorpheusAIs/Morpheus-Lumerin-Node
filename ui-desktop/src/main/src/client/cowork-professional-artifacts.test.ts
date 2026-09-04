import { Buffer } from 'buffer'
import ExcelJS from 'exceljs'
import { PDFDocument } from 'pdf-lib'
import { describe, expect, it } from 'vitest'
import * as yauzl from 'yauzl'
import {
  PROFESSIONAL_ARTIFACT_LIMITS,
  generateDocxArtifact,
  generatePdfArtifact,
  generatePptxArtifact,
  generateProfessionalArtifact,
  generateXlsxArtifact,
  validateProfessionalArtifactRequest
} from './cowork-professional-artifacts'

function readZip(buffer: Buffer): Promise<Map<string, Buffer>> {
  return new Promise((resolve, reject) => {
    yauzl.fromBuffer(buffer, { lazyEntries: true }, (openError, archive) => {
      if (openError || !archive) {
        reject(openError ?? new Error('Unable to open archive.'))
        return
      }
      const entries = new Map<string, Buffer>()
      archive.on('error', reject)
      archive.on('end', () => resolve(entries))
      archive.on('entry', (entry) => {
        if (entry.fileName.endsWith('/')) {
          archive.readEntry()
          return
        }
        archive.openReadStream(entry, (streamError, stream) => {
          if (streamError || !stream) {
            reject(streamError ?? new Error(`Unable to read ${entry.fileName}.`))
            return
          }
          const chunks: Buffer[] = []
          stream.on('data', (chunk: Buffer) => chunks.push(Buffer.from(chunk)))
          stream.on('error', reject)
          stream.on('end', () => {
            entries.set(entry.fileName, Buffer.concat(chunks))
            archive.readEntry()
          })
        })
      })
      archive.readEntry()
    })
  })
}

function archiveRelationships(entries: Map<string, Buffer>): string {
  return [...entries.entries()]
    .filter(([name]) => name.endsWith('.rels'))
    .map(([, value]) => value.toString('utf8'))
    .join('\n')
}

describe('professional Cowork artifacts', () => {
  it('strictly validates JSON shape, filenames, counts, and total text', () => {
    expect(() =>
      validateProfessionalArtifactRequest({
        format: 'docx',
        title: 'Report',
        blocks: [{ type: 'paragraph', text: 'Body' }],
        unexpected: true
      })
    ).toThrow(/unsupported field/)

    expect(() =>
      validateProfessionalArtifactRequest({
        format: 'xlsx',
        title: 'Workbook',
        filename: '../escape',
        sheets: [{ name: 'Data', headers: ['Value'], rows: [] }]
      })
    ).toThrow(/safe filename/)

    expect(() =>
      validateProfessionalArtifactRequest({
        format: 'xlsx',
        title: 'Workbook',
        sheets: [
          { name: 'Data', headers: ['Value'], rows: [] },
          { name: 'data', headers: ['Other'], rows: [] }
        ]
      })
    ).toThrow(/duplicates/)

    expect(() =>
      validateProfessionalArtifactRequest({
        format: 'pptx',
        title: 'Deck',
        slides: Array.from(
          { length: PROFESSIONAL_ARTIFACT_LIMITS.presentationSlides + 1 },
          (_, index) => ({
            title: `Slide ${index + 1}`
          })
        )
      })
    ).toThrow(/between 1 and 40/)

    expect(() =>
      validateProfessionalArtifactRequest({
        format: 'pptx',
        title: 'Dense deck',
        slides: [
          { title: 'Dense slide', body: Array.from({ length: 21 }, () => 'line').join('\n') }
        ]
      })
    ).toThrow(/too much text/)

    expect(() =>
      validateProfessionalArtifactRequest({
        format: 'pdf',
        title: 'Large report',
        blocks: Array.from({ length: 13 }, () => ({ type: 'paragraph', text: 'x'.repeat(20_000) }))
      })
    ).toThrow(/total-character limit/)

    expect(() =>
      validateProfessionalArtifactRequest({
        format: 'docx',
        title: 'Not JSON',
        blocks: [{ type: 'paragraph', text: new Date() }]
      })
    ).toThrow(/plain JSON object/)
  })

  it('creates a bounded DOCX with native structure and no executable or external relationship', async () => {
    const artifact = await generateDocxArtifact({
      format: 'docx',
      filename: 'quarterly-brief.docx',
      title: 'Quarterly Brief',
      subtitle: 'Prepared locally by Morpheus',
      accentColor: '#2563EB',
      blocks: [
        { type: 'heading', level: 1, text: 'Executive Summary' },
        { type: 'paragraph', text: 'Revenue grew while operating risk remained controlled.' },
        { type: 'bulletList', items: ['Customer retention improved', 'Latency declined'] },
        { type: 'numberedList', items: ['Validate inputs', 'Publish results'] },
        {
          type: 'table',
          headers: ['Metric', 'Value'],
          rows: [
            ['Revenue', 125000],
            ['On track', true]
          ]
        }
      ]
    })

    expect(artifact).toMatchObject({
      format: 'docx',
      extension: '.docx',
      suggestedFilename: 'quarterly-brief.docx',
      mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    })
    expect(artifact.sizeBytes).toBe(artifact.buffer.length)
    expect(artifact.buffer.subarray(0, 2).toString()).toBe('PK')

    const entries = await readZip(artifact.buffer)
    expect(entries.has('[Content_Types].xml')).toBe(true)
    expect(entries.has('word/document.xml')).toBe(true)
    const documentXml = entries.get('word/document.xml')!.toString('utf8')
    expect(documentXml).toContain('Quarterly Brief')
    expect(documentXml).toContain('Executive Summary')
    expect(documentXml).toContain('Revenue')
    expect(documentXml).toContain('<w:tbl')
    expect([...entries.keys()].some((name) => /vbaProject|externalLinks/iu.test(name))).toBe(false)
    expect(archiveRelationships(entries)).not.toMatch(/TargetMode=["']External["']/iu)
  }, 20_000)

  it('creates a styled XLSX and keeps formula-looking input as inert text', async () => {
    const formulaLookingText = '=WEBSERVICE("https://example.invalid")'
    const artifact = await generateXlsxArtifact({
      format: 'xlsx',
      title: 'Regional Results',
      accentColor: '#7C3AED',
      sheets: [
        {
          name: 'Results',
          headers: ['Region', 'Amount', 'Note'],
          rows: [
            ['APAC', 1200, formulaLookingText],
            ['EMEA', 900, null]
          ],
          columnWidths: [16, 12, 42]
        }
      ]
    })

    expect(artifact.buffer.subarray(0, 2).toString()).toBe('PK')
    expect(artifact).toMatchObject({
      format: 'xlsx',
      extension: '.xlsx',
      suggestedFilename: 'Regional-Results.xlsx',
      mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    })
    const workbook = new ExcelJS.Workbook()
    await workbook.xlsx.load(artifact.buffer as unknown as ExcelJS.Buffer)
    const worksheet = workbook.getWorksheet('Results')
    expect(worksheet?.getCell('A2').value).toBe('APAC')
    expect(worksheet?.getCell('B2').value).toBe(1200)
    expect(worksheet?.getCell('C2').value).toBe(formulaLookingText)
    expect(worksheet?.getCell('C2').type).toBe(ExcelJS.ValueType.String)
    expect(worksheet?.views[0]).toMatchObject({ state: 'frozen', ySplit: 1 })

    const entries = await readZip(artifact.buffer)
    expect(entries.has('xl/workbook.xml')).toBe(true)
    expect(entries.has('xl/worksheets/sheet1.xml')).toBe(true)
    expect([...entries.keys()].some((name) => /vbaProject|externalLinks/iu.test(name))).toBe(false)
    expect(archiveRelationships(entries)).not.toMatch(/TargetMode=["']External["']/iu)
  }, 20_000)

  it('creates a widescreen PPTX with title, native bullets, and native table content', async () => {
    const artifact = await generatePptxArtifact({
      format: 'pptx',
      filename: 'launch-review',
      title: 'Launch Review',
      slides: [
        { title: 'Launch Review', subtitle: 'Readiness and next actions' },
        {
          title: 'What changed',
          body: 'The release is ready for staged validation.',
          bullets: ['Run smoke tests', 'Review telemetry']
        },
        {
          title: 'Owners',
          table: {
            headers: ['Workstream', 'Owner'],
            rows: [
              ['Desktop', 'Sam'],
              ['Router', 'Lee']
            ]
          }
        }
      ]
    })

    expect(artifact.buffer.subarray(0, 2).toString()).toBe('PK')
    expect(artifact).toMatchObject({
      format: 'pptx',
      extension: '.pptx',
      suggestedFilename: 'launch-review.pptx',
      mimeType: 'application/vnd.openxmlformats-officedocument.presentationml.presentation'
    })
    const entries = await readZip(artifact.buffer)
    expect(entries.has('ppt/presentation.xml')).toBe(true)
    expect(entries.has('ppt/slides/slide1.xml')).toBe(true)
    expect(entries.has('ppt/slides/slide3.xml')).toBe(true)
    const slideText = [...entries.entries()]
      .filter(([name]) => /^ppt\/slides\/slide\d+\.xml$/u.test(name))
      .map(([, value]) => value.toString('utf8'))
      .join('\n')
    expect(slideText).toContain('Launch Review')
    expect(slideText).toContain('Run smoke tests')
    expect(slideText).toContain('Workstream')
    expect(slideText).toContain('Desktop')
    expect(
      [...entries.keys()].some((name) => /vbaProject|externalLinks|embeddings/iu.test(name))
    ).toBe(false)
    expect(archiveRelationships(entries)).not.toMatch(/TargetMode=["']External["']/iu)
  }, 20_000)

  it('creates a deterministic-metadata PDF with pagination and no active content', async () => {
    const artifact = await generatePdfArtifact({
      format: 'pdf',
      title: 'Operating Review',
      subtitle: 'A concise local report',
      blocks: [
        { type: 'heading', level: 1, text: 'Overview' },
        { type: 'paragraph', text: 'The team completed the planned milestones.' },
        { type: 'table', headers: ['Milestone', 'Status'], rows: [['Desktop', 'Complete']] },
        { type: 'pageBreak' },
        { type: 'heading', level: 1, text: 'Next steps' },
        { type: 'numberedList', items: ['Verify the release', 'Monitor adoption'] }
      ]
    })

    expect(artifact.buffer.subarray(0, 5).toString('ascii')).toBe('%PDF-')
    expect(artifact).toMatchObject({
      format: 'pdf',
      extension: '.pdf',
      suggestedFilename: 'Operating-Review.pdf',
      mimeType: 'application/pdf'
    })
    const document = await PDFDocument.load(Uint8Array.from(artifact.buffer))
    expect(document.getTitle()).toBe('Operating Review')
    expect(document.getPageCount()).toBe(2)
    const source = artifact.buffer.toString('latin1')
    expect(source).not.toMatch(/\/JavaScript|\/EmbeddedFile|\/Launch|\/URI/iu)
  }, 20_000)

  it('rejects non-encodable PDF text and format mismatches instead of emitting partial files', async () => {
    await expect(
      generatePdfArtifact({
        format: 'pdf',
        title: 'Unicode check',
        blocks: [{ type: 'paragraph', text: 'Unsupported: 漢' }]
      })
    ).rejects.toThrow(/cannot encode/)

    await expect(
      generateDocxArtifact({
        format: 'pdf',
        title: 'Wrong wrapper',
        blocks: [{ type: 'paragraph', text: 'Body' }]
      })
    ).rejects.toThrow(/must be docx/)
  })

  it('paginates long PDF prose within the page cap', async () => {
    const artifact = await generatePdfArtifact({
      format: 'pdf',
      title: 'Long Report',
      blocks: [{ type: 'paragraph', text: 'A bounded sentence for pagination. '.repeat(350) }]
    })
    const document = await PDFDocument.load(Uint8Array.from(artifact.buffer))
    expect(document.getPageCount()).toBeGreaterThan(1)
    expect(document.getPageCount()).toBeLessThanOrEqual(PROFESSIONAL_ARTIFACT_LIMITS.pdfPages)
  }, 20_000)

  it('dispatches through the common generator without mutating caller input', async () => {
    const request = {
      format: 'xlsx' as const,
      title: 'Immutable Input',
      sheets: [{ name: 'Data', headers: ['Value'], rows: [[1], [2]] }]
    }
    const snapshot = JSON.stringify(request)
    const artifact = await generateProfessionalArtifact(request)
    expect(artifact.format).toBe('xlsx')
    expect(JSON.stringify(request)).toBe(snapshot)
  }, 20_000)
})
