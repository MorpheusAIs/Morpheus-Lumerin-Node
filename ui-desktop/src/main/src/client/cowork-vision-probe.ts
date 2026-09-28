/**
 * Whether a model can actually see an image, established by asking it.
 *
 * A model's name and tags are a guess. "gpt-4o" in a name means nothing about
 * what a particular provider is serving behind it, and a task that hands an
 * image to a text-only endpoint gets either a hard rejection or, far worse, a
 * confident description of a picture the model never received. So Workspace
 * sends one tiny image with a question only a model that saw it can answer,
 * and remembers the answer per model.
 *
 * The probe is deliberately cheap: a handful of bytes and a one-word reply.
 */

import { createHash } from 'node:crypto'

/** Colours far enough apart that no reasonable describer confuses two of them. */
const PROBE_COLOURS: ReadonlyArray<{ name: string; rgb: [number, number, number] }> = [
  { name: 'red', rgb: [220, 20, 20] },
  { name: 'green', rgb: [20, 170, 60] },
  { name: 'blue', rgb: [30, 60, 220] },
  { name: 'yellow', rgb: [240, 220, 30] },
  { name: 'purple', rgb: [130, 40, 190] },
  { name: 'orange', rgb: [240, 130, 20] }
]

/** Words a model reaching for a plausible answer without looking tends to use. */
const COLOUR_SYNONYMS: Readonly<Record<string, readonly string[]>> = {
  red: ['red', 'crimson', 'scarlet'],
  green: ['green', 'emerald'],
  blue: ['blue', 'azure', 'navy'],
  yellow: ['yellow', 'gold', 'golden'],
  purple: ['purple', 'violet', 'magenta'],
  orange: ['orange', 'amber']
}

const crc32 = (bytes: Buffer): number => {
  let crc = 0xffffffff
  for (const byte of bytes) {
    crc ^= byte
    for (let bit = 0; bit < 8; bit++) crc = crc & 1 ? (crc >>> 1) ^ 0xedb88320 : crc >>> 1
  }
  return (crc ^ 0xffffffff) >>> 0
}

const pngChunk = (type: string, body: Buffer): Buffer => {
  const head = Buffer.alloc(8)
  head.writeUInt32BE(body.length, 0)
  head.write(type, 4, 'latin1')
  const tail = Buffer.alloc(4)
  tail.writeUInt32BE(crc32(Buffer.concat([Buffer.from(type, 'latin1'), body])), 0)
  return Buffer.concat([head, body, tail])
}

/**
 * Deflate's stored-block mode, so the probe needs no compression library and
 * produces byte-identical output on every platform.
 */
function storedZlib(payload: Buffer): Buffer {
  const blocks: Buffer[] = [Buffer.from([0x78, 0x01])]
  for (let offset = 0; offset < payload.length; offset += 0xffff) {
    const slice = payload.subarray(offset, offset + 0xffff)
    const header = Buffer.alloc(5)
    header[0] = offset + slice.length >= payload.length ? 1 : 0
    header.writeUInt16LE(slice.length, 1)
    header.writeUInt16LE(~slice.length & 0xffff, 3)
    blocks.push(header, slice)
  }
  // Adler-32 over the uncompressed bytes, as the zlib trailer requires.
  let low = 1
  let high = 0
  for (const byte of payload) {
    low = (low + byte) % 65521
    high = (high + low) % 65521
  }
  const adler = Buffer.alloc(4)
  adler.writeUInt32BE(((high << 16) | low) >>> 0, 0)
  return Buffer.concat([...blocks, adler])
}

/** A solid square of one colour, small enough that no provider will resize it. */
export function probeImagePng(rgb: readonly [number, number, number], size = 32): Buffer {
  const rows: Buffer[] = []
  for (let row = 0; row < size; row++) {
    const line = Buffer.alloc(1 + size * 3)
    for (let column = 0; column < size; column++) {
      line[1 + column * 3] = rgb[0]
      line[2 + column * 3] = rgb[1]
      line[3 + column * 3] = rgb[2]
    }
    rows.push(line)
  }
  const ihdr = Buffer.alloc(13)
  ihdr.writeUInt32BE(size, 0)
  ihdr.writeUInt32BE(size, 4)
  ihdr[8] = 8 // bit depth
  ihdr[9] = 2 // truecolour
  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    pngChunk('IHDR', ihdr),
    pngChunk('IDAT', storedZlib(Buffer.concat(rows))),
    pngChunk('IEND', Buffer.alloc(0))
  ])
}

export const VISION_PROBE_QUESTION =
  'Reply with exactly one word: the colour that fills this image. ' +
  'If you cannot see an image, reply with the single word none.'

/**
 * Picks the colour from the model id, so re-probing the same model asks the
 * same question and two models rarely get the same one. A model that answers
 * from a guess rather than from pixels is right one time in six.
 */
export function probeColourFor(modelId: string): { name: string; rgb: [number, number, number] } {
  const digest = createHash('sha256').update(modelId).digest()
  return PROBE_COLOURS[digest[0] % PROBE_COLOURS.length]
}

/**
 * Grades a reply. Only a clear naming of the right colour counts as sight:
 * silence, a refusal, an error, or the wrong colour all mean the model did not
 * see the picture, and Workspace would rather withhold the image tool from a
 * model that can use it than hand it to one that cannot.
 */
export function visionProbePassed(answer: string, colour: string): boolean {
  const text = answer.toLowerCase()
  const accepted = COLOUR_SYNONYMS[colour] ?? [colour]
  if (!accepted.some((word) => new RegExp(`\\b${word}\\b`).test(text))) return false
  // A reply that names several colours has described a palette, not this image.
  const named = Object.entries(COLOUR_SYNONYMS).filter(([, words]) =>
    words.some((word) => new RegExp(`\\b${word}\\b`).test(text))
  )
  return named.length === 1
}

export interface CoworkVisionProbeResult {
  modelId: string
  sees: boolean
  probedAt: number
  /** Kept so a surprising verdict can be explained to the user, never rendered raw. */
  answer: string
}

/** The request body a probe sends, in the shape the completions endpoint expects. */
export function visionProbeBody(modelId: string): Record<string, unknown> {
  const colour = probeColourFor(modelId)
  return {
    model: modelId,
    stream: false,
    temperature: 0,
    max_tokens: 16,
    messages: [
      {
        role: 'user',
        content: [
          { type: 'text', text: VISION_PROBE_QUESTION },
          {
            type: 'image_url',
            image_url: {
              url: `data:image/png;base64,${probeImagePng(colour.rgb).toString('base64')}`
            }
          }
        ]
      }
    ]
  }
}

/**
 * Runs one probe. Any failure at all — a rejected request, a malformed body, a
 * timeout — is a negative result rather than a thrown error, because the caller
 * is deciding whether to offer a tool, not whether the task can continue.
 */
export async function runCoworkVisionProbe(
  modelId: string,
  send: (body: Record<string, unknown>) => Promise<string>
): Promise<CoworkVisionProbeResult> {
  const colour = probeColourFor(modelId)
  let answer = ''
  try {
    const raw = await send(visionProbeBody(modelId))
    const parsed = JSON.parse(raw)
    const content = parsed?.choices?.[0]?.message?.content
    answer = typeof content === 'string' ? content : ''
  } catch (error) {
    answer = error instanceof Error ? `probe failed: ${error.message}` : 'probe failed'
  }
  return {
    modelId,
    sees: visionProbePassed(answer, colour.name),
    probedAt: Date.now(),
    answer: answer.slice(0, 200)
  }
}
