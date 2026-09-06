/**
 * Image handling for Workspace tasks.
 *
 * Images are never stored as base64 in the durable transcript. A task records
 * only a reference to the file on disk, and the bytes are read and encoded at
 * request time. That keeps task databases the size they were before images
 * existed, and it means a model always sees the current contents of a file
 * rather than a snapshot taken many turns ago.
 */

/** Formats every vision-capable endpoint in practice accepts. */
const MEDIA_TYPES_BY_EXTENSION: Record<string, string> = {
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.webp': 'image/webp',
  '.gif': 'image/gif',
  '.bmp': 'image/bmp'
}

/**
 * A single image a provider will accept. Larger files are refused with their
 * size named, so the model can downscale rather than retry the same call.
 */
export const MAX_IMAGE_BYTES = 8 * 1024 * 1024
/** Total encoded image payload allowed in one completion request. */
export const MAX_REQUEST_IMAGE_BYTES = 16 * 1024 * 1024
/** Newest images are sent as pixels; older ones become a one-line record. */
export const MAX_IMAGES_PER_REQUEST = 4

export const imageMediaTypeForPath = (relativePath: string): string | null => {
  const match = /\.[A-Za-z0-9]+$/.exec(relativePath.trim())
  if (!match) return null
  return MEDIA_TYPES_BY_EXTENSION[match[0].toLowerCase()] ?? null
}

export const isCoworkImagePath = (relativePath: string): boolean =>
  imageMediaTypeForPath(relativePath) !== null

export const supportedCoworkImageExtensions = (): string[] => Object.keys(MEDIA_TYPES_BY_EXTENSION)

export interface CoworkImageDimensions {
  width: number
  height: number
}

const readUInt24LE = (buffer: Buffer, offset: number): number =>
  buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16)

function pngDimensions(buffer: Buffer): CoworkImageDimensions | null {
  // Signature, then a length/type pair, then IHDR's width and height.
  if (buffer.length < 24) return null
  if (buffer.readUInt32BE(0) !== 0x89504e47 || buffer.readUInt32BE(4) !== 0x0d0a1a0a) return null
  if (buffer.toString('latin1', 12, 16) !== 'IHDR') return null
  return { width: buffer.readUInt32BE(16), height: buffer.readUInt32BE(20) }
}

function gifDimensions(buffer: Buffer): CoworkImageDimensions | null {
  if (buffer.length < 10) return null
  const header = buffer.toString('latin1', 0, 6)
  if (header !== 'GIF87a' && header !== 'GIF89a') return null
  return { width: buffer.readUInt16LE(6), height: buffer.readUInt16LE(8) }
}

function bmpDimensions(buffer: Buffer): CoworkImageDimensions | null {
  if (buffer.length < 26 || buffer.toString('latin1', 0, 2) !== 'BM') return null
  // Height is signed: a negative value means the rows are stored top-down.
  return { width: buffer.readInt32LE(18), height: Math.abs(buffer.readInt32LE(22)) }
}

/** Start-of-frame markers, which are the only ones carrying the frame size. */
const JPEG_FRAME_MARKERS = new Set([
  0xc0, 0xc1, 0xc2, 0xc3, 0xc5, 0xc6, 0xc7, 0xc9, 0xca, 0xcb, 0xcd, 0xce, 0xcf
])

function jpegDimensions(buffer: Buffer): CoworkImageDimensions | null {
  if (buffer.length < 4 || buffer.readUInt16BE(0) !== 0xffd8) return null
  let offset = 2
  while (offset + 3 < buffer.length) {
    if (buffer[offset] !== 0xff) {
      offset += 1
      continue
    }
    const marker = buffer[offset + 1]
    // Padding and the standalone markers carry no length field.
    if (marker === 0xff) {
      offset += 1
      continue
    }
    if (marker === 0xd8 || (marker >= 0xd0 && marker <= 0xd9)) {
      offset += 2
      continue
    }
    const length = buffer.readUInt16BE(offset + 2)
    if (length < 2) return null
    if (JPEG_FRAME_MARKERS.has(marker)) {
      if (offset + 9 > buffer.length) return null
      return { height: buffer.readUInt16BE(offset + 5), width: buffer.readUInt16BE(offset + 7) }
    }
    offset += 2 + length
  }
  return null
}

function webpDimensions(buffer: Buffer): CoworkImageDimensions | null {
  if (buffer.length < 30) return null
  if (buffer.toString('latin1', 0, 4) !== 'RIFF') return null
  if (buffer.toString('latin1', 8, 12) !== 'WEBP') return null
  const chunk = buffer.toString('latin1', 12, 16)
  if (chunk === 'VP8X') {
    return { width: readUInt24LE(buffer, 24) + 1, height: readUInt24LE(buffer, 27) + 1 }
  }
  if (chunk === 'VP8 ') {
    // The keyframe header repeats the start code before the 14-bit dimensions.
    return {
      width: buffer.readUInt16LE(26) & 0x3fff,
      height: buffer.readUInt16LE(28) & 0x3fff
    }
  }
  if (chunk === 'VP8L') {
    if (buffer[20] !== 0x2f) return null
    const bits = buffer.readUInt32LE(21)
    return { width: (bits & 0x3fff) + 1, height: ((bits >> 14) & 0x3fff) + 1 }
  }
  return null
}

/**
 * Reads the frame size straight out of the container header. Decoding the
 * pixels would mean a native image dependency in the main process for a value
 * that every one of these formats states in its first few dozen bytes.
 */
export function readCoworkImageDimensions(
  buffer: Buffer,
  mediaType: string
): CoworkImageDimensions | null {
  const dimensions =
    mediaType === 'image/png'
      ? pngDimensions(buffer)
      : mediaType === 'image/jpeg'
        ? jpegDimensions(buffer)
        : mediaType === 'image/gif'
          ? gifDimensions(buffer)
          : mediaType === 'image/bmp'
            ? bmpDimensions(buffer)
            : mediaType === 'image/webp'
              ? webpDimensions(buffer)
              : null
  if (!dimensions) return null
  const { width, height } = dimensions
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) return null
  return { width, height }
}

/**
 * Confirms the bytes really are the format the extension claims. A renamed
 * file would otherwise be sent to a provider as an image and rejected there,
 * where the error says nothing a person can act on.
 */
export function assertCoworkImageBytes(buffer: Buffer, mediaType: string, path: string): void {
  if (readCoworkImageDimensions(buffer, mediaType) === null) {
    throw new Error(
      `“${path}” is not a readable ${mediaType.replace('image/', '').toUpperCase()} image, whatever its extension says.`
    )
  }
}

export const coworkImageDataUrl = (buffer: Buffer, mediaType: string): string =>
  `data:${mediaType};base64,${buffer.toString('base64')}`

/** Base64 costs four characters for every three bytes, rounded up to a block. */
export const encodedImageBytes = (bytes: number): number => Math.ceil(bytes / 3) * 4
