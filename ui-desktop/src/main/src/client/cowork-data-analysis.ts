const MAX_CSV_BYTES = 512 * 1024
const MAX_CSV_ROWS = 5_000
const MAX_CSV_COLUMNS = 100
const MAX_CELL_CHARACTERS = 50_000
const DELIMITERS = new Set([',', '\t', ';', '|'])

export interface CsvColumnSummary {
  name: string
  kind: 'number' | 'text' | 'empty'
  nonEmpty: number
  missing: number
  uniqueCount: number
  topValues: Array<{ value: string; count: number }>
  sum?: number
  mean?: number
  min?: number
  max?: number
}

export interface CsvAnalysis {
  rowCount: number
  columnCount: number
  delimiter: string
  columns: CsvColumnSummary[]
  sample: Record<string, string>[]
  truncated: boolean
}

function parseRows(text: string, delimiter: string): { rows: string[][]; truncated: boolean } {
  const rows: string[][] = []
  let row: string[] = []
  let cell = ''
  let quoted = false
  let truncated = false

  const pushCell = () => {
    if (cell.length > MAX_CELL_CHARACTERS) {
      cell = cell.slice(0, MAX_CELL_CHARACTERS)
      truncated = true
    }
    row.push(cell)
    cell = ''
    if (row.length > MAX_CSV_COLUMNS)
      throw new Error(`CSV exceeds the ${MAX_CSV_COLUMNS}-column limit.`)
  }
  const pushRow = () => {
    pushCell()
    rows.push(row)
    row = []
    if (rows.length > MAX_CSV_ROWS + 1) {
      rows.pop()
      truncated = true
    }
  }

  for (let index = 0; index < text.length; index++) {
    const character = text[index]
    if (quoted) {
      if (character === '"') {
        if (text[index + 1] === '"') {
          cell += '"'
          index++
        } else {
          quoted = false
        }
      } else {
        cell += character
      }
      continue
    }
    if (character === '"' && cell.length === 0) {
      quoted = true
    } else if (character === delimiter) {
      pushCell()
    } else if (character === '\n') {
      if (rows.length <= MAX_CSV_ROWS) pushRow()
      else truncated = true
    } else if (character !== '\r') {
      cell += character
    }
  }
  if (quoted) throw new Error('CSV contains an unterminated quoted field.')
  if (row.length || cell.length) pushRow()
  if (rows.at(-1)?.every((value) => value === '')) rows.pop()
  return { rows, truncated }
}

function uniqueHeaders(raw: string[]): string[] {
  const used = new Map<string, number>()
  return raw.map((value, index) => {
    const base = value.trim().slice(0, 200) || `column_${index + 1}`
    const count = (used.get(base) ?? 0) + 1
    used.set(base, count)
    return count === 1 ? base : `${base}_${count}`
  })
}

export function analyzeCsvText(text: string, requestedDelimiter = ','): CsvAnalysis {
  if (Buffer.byteLength(text, 'utf8') > MAX_CSV_BYTES) {
    throw new Error(`CSV exceeds the ${MAX_CSV_BYTES / 1024} KiB analysis limit.`)
  }
  const delimiter = requestedDelimiter === '\\t' ? '\t' : requestedDelimiter
  if (!DELIMITERS.has(delimiter))
    throw new Error('CSV delimiter must be comma, tab, semicolon, or pipe.')
  const { rows, truncated } = parseRows(text.replace(/^\uFEFF/, ''), delimiter)
  if (!rows.length) throw new Error('CSV is empty.')
  const headers = uniqueHeaders(rows[0])
  const data = rows.slice(1).map((candidate) => headers.map((_, index) => candidate[index] ?? ''))
  const columns = headers.map((name, columnIndex): CsvColumnSummary => {
    const values = data.map((row) => row[columnIndex].trim())
    const populated = values.filter(Boolean)
    const frequencies = new Map<string, number>()
    populated.forEach((value) => frequencies.set(value, (frequencies.get(value) ?? 0) + 1))
    const topValues = [...frequencies.entries()]
      .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))
      .slice(0, 10)
      .map(([value, count]) => ({ value: value.slice(0, 500), count }))
    const numeric = populated.map(Number)
    const allNumeric = populated.length > 0 && numeric.every(Number.isFinite)
    const common = {
      name,
      nonEmpty: populated.length,
      missing: data.length - populated.length,
      uniqueCount: frequencies.size,
      topValues
    }
    if (!populated.length) return { ...common, kind: 'empty' }
    if (!allNumeric) return { ...common, kind: 'text' }
    const sum = numeric.reduce((total, value) => total + value, 0)
    return {
      ...common,
      kind: 'number',
      sum,
      mean: sum / numeric.length,
      min: Math.min(...numeric),
      max: Math.max(...numeric)
    }
  })
  return {
    rowCount: data.length,
    columnCount: headers.length,
    delimiter: delimiter === '\t' ? '\\t' : delimiter,
    columns,
    sample: data
      .slice(0, 5)
      .map((row) =>
        Object.fromEntries(headers.map((header, index) => [header, row[index].slice(0, 1_000)]))
      ),
    truncated
  }
}
