import { describe, expect, it } from 'vitest';
import {
  Attachment,
  buildDocumentContext,
  buildUserMessage,
  classify,
  estimateTokens,
  formatBytes,
  getVisionCapability,
  looksVisionCapable,
  MAX_ATTACHMENTS,
  MAX_IMAGE_BYTES,
  validateFile,
} from './attachments';

const doc = (over: Partial<Attachment> = {}): Attachment => ({
  id: 'd1',
  name: 'spec.pdf',
  mime: 'application/pdf',
  size: 1000,
  kind: 'document',
  text: 'hello world',
  status: 'ready',
  ...over,
});

const img = (over: Partial<Attachment> = {}): Attachment => ({
  id: 'i1',
  name: 'shot.png',
  mime: 'image/png',
  size: 1000,
  kind: 'image',
  dataUrl: 'data:image/png;base64,AAAA',
  status: 'ready',
  ...over,
});

describe('classify', () => {
  it.each([
    ['photo.png', 'image/png'],
    ['photo.JPG', 'image/jpeg'],
    ['anim.gif', 'image/gif'],
  ])('treats %s as an image', (name, type) => {
    expect(classify({ name, type })).toBe('image');
  });

  // SVG is an image MIME but is XML underneath. Models read the source far
  // better than a rasterised version, and routing it to text means it also
  // works on non-vision models.
  it.each([
    ['icon.svg', 'image/svg+xml'],
    ['icon.SVG', ''],
  ])('routes %s to the text path, not the image path', (name, type) => {
    expect(classify({ name, type })).toBe('document');
  });

  it.each([
    ['report.pdf', 'application/pdf'],
    ['notes.md', 'text/markdown'],
    ['main.go', ''],
    ['unknown.bin', ''],
  ])('treats %s as a document', (name, type) => {
    expect(classify({ name, type })).toBe('document');
  });
});

describe('validateFile', () => {
  it('accepts a normal file', () => {
    expect(validateFile({ name: 'a.pdf', type: '', size: 1000 }, 0)).toBeNull();
  });

  it('rejects an oversized image with a readable message', () => {
    const err = validateFile(
      { name: 'huge.png', type: 'image/png', size: MAX_IMAGE_BYTES + 1 },
      0,
    );
    expect(err).toMatch(/huge\.png/);
    expect(err).toMatch(/limit for images/);
  });

  it('rejects an empty file', () => {
    expect(validateFile({ name: 'a.txt', type: '', size: 0 }, 0)).toMatch(
      /empty/,
    );
  });

  it('enforces the attachment count cap', () => {
    expect(
      validateFile({ name: 'a.txt', type: '', size: 10 }, MAX_ATTACHMENTS),
    ).toMatch(/at most/);
  });

  // Documents get a higher cap than images: a PDF shrinks enormously once
  // extracted to text, whereas an image goes into the prompt whole as base64.
  it('allows a document larger than the image cap', () => {
    expect(
      validateFile(
        { name: 'big.pdf', type: 'application/pdf', size: MAX_IMAGE_BYTES + 1 },
        0,
      ),
    ).toBeNull();
  });
});

describe('buildUserMessage', () => {
  // The text-only path must keep the exact shape it had before attachments
  // existed — a parts array can break models that only accept a string.
  it('returns a plain string when there are no images', () => {
    const msg = buildUserMessage('hi', []);
    expect(msg).toEqual({ role: 'user', content: 'hi' });
    expect(typeof msg.content).toBe('string');
  });

  it('keeps content a string when only documents are attached', () => {
    const msg = buildUserMessage('summarise', [doc()]);
    expect(typeof msg.content).toBe('string');
    expect(msg.content).toContain('summarise');
    expect(msg.content).toContain('hello world');
  });

  it('switches to a parts array when an image is attached', () => {
    const msg = buildUserMessage('what is this', [img()]);
    expect(Array.isArray(msg.content)).toBe(true);
    const parts = msg.content as any[];
    expect(parts[0]).toEqual({ type: 'text', text: 'what is this' });
    expect(parts[1]).toEqual({
      type: 'image_url',
      image_url: { url: 'data:image/png;base64,AAAA' },
    });
  });

  it('puts document context and images in one message', () => {
    const parts = buildUserMessage('compare', [doc(), img()]).content as any[];
    expect(parts[0].text).toContain('hello world');
    expect(parts[0].text).toContain('compare');
    expect(parts.filter((p) => p.type === 'image_url')).toHaveLength(1);
  });

  it('ignores attachments that are not ready', () => {
    const msg = buildUserMessage('hi', [
      doc({ status: 'parsing' }),
      doc({ id: 'd2', status: 'error' }),
      img({ status: 'parsing' }),
    ]);
    expect(msg.content).toBe('hi');
  });

  it('sends attachments even with an empty prompt', () => {
    const parts = buildUserMessage('', [img()]).content as any[];
    expect(parts.filter((p) => p.type === 'image_url')).toHaveLength(1);
  });
});

describe('buildDocumentContext', () => {
  it('delimits each document so the model can tell them apart', () => {
    const ctx = buildDocumentContext([
      doc({ name: 'a.pdf', text: 'AAA' }),
      doc({ id: 'd2', name: 'b.md', text: 'BBB' }),
    ]);
    expect(ctx).toContain('<attachment name="a.pdf"');
    expect(ctx).toContain('<attachment name="b.md"');
    expect(ctx).toContain('AAA');
    expect(ctx).toContain('BBB');
    expect(ctx).toContain('2 files');
    expect(ctx).toContain('untrusted reference data');
  });

  it('is empty when nothing has usable text', () => {
    expect(buildDocumentContext([])).toBe('');
    expect(buildDocumentContext([doc({ text: '   ' })])).toBe('');
    expect(buildDocumentContext([img()])).toBe('');
  });
});

describe('estimateTokens', () => {
  it('counts the prompt', () => {
    expect(estimateTokens([], 'a'.repeat(400))).toBe(100);
  });

  it('counts document text', () => {
    expect(estimateTokens([doc({ text: 'a'.repeat(4000) })], '')).toBe(1000);
  });

  it('charges a flat estimate per image', () => {
    expect(estimateTokens([img()], '')).toBe(850);
  });

  it('ignores documents with no extracted text', () => {
    expect(estimateTokens([doc({ text: '' })], '')).toBe(0);
  });
});

describe('looksVisionCapable', () => {
  it.each([
    'llava-1.5-7b',
    'gpt-4o-mini',
    'Claude-3-Opus',
    'gemini-1.5-pro',
    'Qwen2-VL-7B',
    'pixtral-12b',
  ])('recognises %s', (Name) => {
    expect(looksVisionCapable({ Name })).toBe(true);
  });

  it.each(['llama-3-8b', 'mistral-7b', 'tinyllama', 'deepseek-coder'])(
    'does not claim %s is vision-capable',
    (Name) => {
      expect(looksVisionCapable({ Name })).toBe(false);
    },
  );

  // No on-chain tag exists today, but honour one if a provider sets it —
  // otherwise a genuinely new vision model stays unrecognised forever.
  it('honours an explicit tag', () => {
    expect(looksVisionCapable({ Name: 'custom-model', Tags: ['vision'] })).toBe(
      true,
    );
    expect(
      looksVisionCapable({ Name: 'custom-model', Tags: ['multimodal'] }),
    ).toBe(true);
    expect(looksVisionCapable({ Name: 'custom-model', Tags: ['llm'] })).toBe(
      false,
    );
  });

  it('distinguishes declared support from name-based detection', () => {
    expect(getVisionCapability({ Name: 'custom', Tags: ['vision'] })).toBe(
      'declared',
    );
    expect(getVisionCapability({ Name: 'Qwen2-VL-7B', Tags: ['llm'] })).toBe(
      'detected',
    );
    expect(getVisionCapability({ Name: 'llama-3-8b', Tags: ['llm'] })).toBe(
      'none',
    );
  });

  it('does not throw on a missing model', () => {
    expect(looksVisionCapable(undefined)).toBe(false);
    expect(looksVisionCapable({})).toBe(false);
  });

  it('handles null, non-array, and mixed tag metadata safely', () => {
    expect(getVisionCapability({ Name: 'custom', Tags: null })).toBe('none');
    expect(
      getVisionCapability({ Name: 'custom', Tags: { bad: 'shape' } }),
    ).toBe('none');
    expect(getVisionCapability({ Name: 'custom', Tags: 'llm, vision' })).toBe(
      'declared',
    );
    expect(
      getVisionCapability({
        Name: 'custom',
        Tags: [null, { nested: true }, 'multimodal'],
      }),
    ).toBe('declared');
  });
});

describe('formatBytes', () => {
  it.each([
    [512, '512 B'],
    [2048, '2 KB'],
    [5 * 1024 * 1024, '5.0 MB'],
  ])('formats %i', (n, expected) => {
    expect(formatBytes(n)).toBe(expected);
  });
});
