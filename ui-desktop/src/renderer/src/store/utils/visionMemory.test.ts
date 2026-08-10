import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  forgetImageRejection,
  hasRejectedImages,
  isVisionRejection,
  rememberImageRejection,
} from './visionMemory';

// Verbatim from a real provider refusal, including the layers of wrapping the
// proxy-router adds. The actionable clause is buried in the middle.
const REAL_REJECTION =
  'provider request failed: provider error: upstream error 400: {"details":{"_errors":[],"messages":{"_errors":["Image content is not supported by this model. Please use a model that supports vision."]}},"error":"Invalid request parameters","issues":[{"code":"custom","message":"Image content is not supported by this model. Please use a model that supports vision.","path":["messages"]}]}';

beforeEach(() => {
  localStorage.clear();
});

describe('isVisionRejection', () => {
  it('detects the real provider refusal through all the JSON wrapping', () => {
    expect(isVisionRejection(REAL_REJECTION)).toBe(true);
  });

  it.each([
    'Image content is not supported by this model.',
    'Please use a model that supports vision.',
    'this model does not support image input',
    'requires a vision model',
  ])('detects the variant: %s', (msg) => {
    expect(isVisionRejection(msg)).toBe(true);
  });

  it('accepts an Error instance', () => {
    expect(isVisionRejection(new Error(REAL_REJECTION))).toBe(true);
  });

  // Misclassifying an unrelated failure would permanently mark a perfectly
  // good vision model as broken, so the false-positive direction matters more
  // than the false-negative one.
  it.each([
    'insufficient funds for gas',
    'method is not allowed on this endpoint',
    'context length exceeded',
    'rate limit exceeded',
    'the image was processed successfully',
  ])('does not fire on: %s', (msg) => {
    expect(isVisionRejection(msg)).toBe(false);
  });

  it.each([[null], [undefined], ['']])('does not throw on %s', (input) => {
    expect(() => isVisionRejection(input)).not.toThrow();
    expect(isVisionRejection(input)).toBe(false);
  });
});

describe('rejection memory', () => {
  it('remembers a rejection', () => {
    expect(hasRejectedImages('0xabc')).toBe(false);
    rememberImageRejection('0xabc');
    expect(hasRejectedImages('0xabc')).toBe(true);
  });

  it('scopes the memory to one model', () => {
    rememberImageRejection('0xabc');
    expect(hasRejectedImages('0xdef')).toBe(false);
  });

  it('does not duplicate on repeat rejections', () => {
    rememberImageRejection('0xabc');
    rememberImageRejection('0xabc');
    expect(JSON.parse(localStorage.getItem('vision-rejected-models')!)).toEqual(['0xabc']);
  });

  it('can be cleared for a retry', () => {
    rememberImageRejection('0xabc');
    forgetImageRejection('0xabc');
    expect(hasRejectedImages('0xabc')).toBe(false);
  });

  it.each([[undefined], ['']])('ignores a missing model id (%s)', (id) => {
    expect(() => rememberImageRejection(id)).not.toThrow();
    expect(hasRejectedImages(id)).toBe(false);
  });

  it('survives a corrupt store rather than throwing', () => {
    localStorage.setItem('vision-rejected-models', '{not json');
    expect(hasRejectedImages('0xabc')).toBe(false);
    expect(() => rememberImageRejection('0xabc')).not.toThrow();
  });

  it('keeps the list bounded', () => {
    for (let i = 0; i < 250; i++) {
      rememberImageRejection(`0x${i}`);
    }
    const stored = JSON.parse(localStorage.getItem('vision-rejected-models')!);
    expect(stored.length).toBeLessThanOrEqual(200);
    // The most recent entries are the ones worth keeping.
    expect(stored).toContain('0x249');
  });

  it('does not throw when localStorage is unavailable', () => {
    const spy = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('QuotaExceededError');
    });
    expect(() => rememberImageRejection('0xabc')).not.toThrow();
    spy.mockRestore();
  });
});
