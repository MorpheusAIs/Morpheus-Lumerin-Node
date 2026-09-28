import { inflateSync } from 'node:zlib'
import { describe, expect, it } from 'vitest'
import { readCoworkImageDimensions } from './cowork-images'
import {
  probeColourFor,
  probeImagePng,
  runCoworkVisionProbe,
  visionProbeBody,
  visionProbePassed
} from './cowork-vision-probe'

/** Walks the chunk list so a malformed length or CRC cannot pass unnoticed. */
function pngChunks(png: Buffer): Array<{ type: string; body: Buffer }> {
  const chunks: Array<{ type: string; body: Buffer }> = []
  let offset = 8
  while (offset < png.length) {
    const length = png.readUInt32BE(offset)
    const type = png.toString('latin1', offset + 4, offset + 8)
    chunks.push({ type, body: png.subarray(offset + 8, offset + 8 + length) })
    offset += 12 + length
  }
  return chunks
}

describe('vision probe image', () => {
  it('produces a PNG a decoder actually accepts', () => {
    const png = probeImagePng([220, 20, 20], 8)

    expect(readCoworkImageDimensions(png, 'image/png')).toEqual({ width: 8, height: 8 })
    const chunks = pngChunks(png)
    expect(chunks.map((chunk) => chunk.type)).toEqual(['IHDR', 'IDAT', 'IEND'])
    // Every scanline is a zero filter byte followed by the same solid colour.
    const raw = inflateSync(chunks[1].body)
    expect(raw).toHaveLength(8 * (1 + 8 * 3))
    for (let row = 0; row < 8; row++) {
      const line = raw.subarray(row * 25, row * 25 + 25)
      expect(line[0]).toBe(0)
      for (let column = 0; column < 8; column++) {
        expect([...line.subarray(1 + column * 3, 4 + column * 3)]).toEqual([220, 20, 20])
      }
    }
  })

  it('asks the same model the same question every time, and different models different ones', () => {
    expect(probeColourFor('model-a')).toEqual(probeColourFor('model-a'))
    const colours = new Set(
      Array.from({ length: 40 }, (_, index) => probeColourFor(`model-${index}`).name)
    )
    expect(colours.size).toBeGreaterThan(1)
  })

  it('sends the image on a user message in the parts form a provider expects', () => {
    const body = visionProbeBody('model-a') as any

    expect(body.messages[0].role).toBe('user')
    expect(body.messages[0].content[1].image_url.url).toMatch(/^data:image\/png;base64,/)
    expect(body.stream).toBe(false)
  })
})

describe('grading a vision probe answer', () => {
  it('accepts only an unambiguous naming of the colour that was sent', () => {
    expect(visionProbePassed('Red', 'red')).toBe(true)
    expect(visionProbePassed('The image is a solid crimson square.', 'red')).toBe(true)
    expect(visionProbePassed('blue', 'red')).toBe(false)
    expect(visionProbePassed('none', 'red')).toBe(false)
    expect(visionProbePassed('', 'red')).toBe(false)
  })

  it('rejects a hedged reply that lists colours instead of reporting one', () => {
    expect(visionProbePassed('It could be red, green, or blue.', 'red')).toBe(false)
    expect(
      visionProbePassed('I cannot see images, but common answers are red or blue.', 'red')
    ).toBe(false)
  })

  it('does not match a colour embedded in another word', () => {
    expect(visionProbePassed('This is a redirect error.', 'red')).toBe(false)
  })
})

describe('running a vision probe', () => {
  it('records a pass when the model names the colour it was sent', async () => {
    const colour = probeColourFor('seeing-model')
    const result = await runCoworkVisionProbe('seeing-model', async () =>
      JSON.stringify({ choices: [{ message: { content: colour.name } }] })
    )

    expect(result).toMatchObject({ modelId: 'seeing-model', sees: true })
    expect(result.probedAt).toBeGreaterThan(0)
  })

  it('treats a rejected request as no vision rather than letting it throw', async () => {
    const result = await runCoworkVisionProbe('blind-model', async () => {
      throw new Error('tools is not supported by this model')
    })

    expect(result.sees).toBe(false)
    expect(result.answer).toContain('probe failed')
  })

  it('treats an unparseable or empty completion as no vision', async () => {
    await expect(runCoworkVisionProbe('a', async () => 'not json')).resolves.toMatchObject({
      sees: false
    })
    await expect(
      runCoworkVisionProbe('b', async () => JSON.stringify({ choices: [] }))
    ).resolves.toMatchObject({ sees: false })
  })
})
