import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  getModelModality,
  isClosed,
  isCoworkCandidate,
  isSecureModel,
  scheduleSessionExpiry,
} from './utils';

const nowSeconds = () => Math.floor(Date.now() / 1000);

// `isClosed` decides whether the Chat tab shows "you have an open session" or
// the "Select payment method" screen. Getting it wrong in the permissive
// direction is what let users stake a second time on top of a live session.
describe('isClosed', () => {
  it('treats a session with ClosedAt as closed', () => {
    expect(
      isClosed({ ClosedAt: 1700000000, EndsAt: nowSeconds() + 3600 }),
    ).toBeTruthy();
  });

  it('treats an expired session as closed', () => {
    expect(isClosed({ ClosedAt: 0, EndsAt: nowSeconds() - 1 })).toBeTruthy();
  });

  it('treats a live session as open', () => {
    expect(isClosed({ ClosedAt: 0, EndsAt: nowSeconds() + 3600 })).toBeFalsy();
  });

  it('does not mark a session closed merely because it ends soon', () => {
    expect(isClosed({ ClosedAt: 0, EndsAt: nowSeconds() + 5 })).toBeFalsy();
  });

  it('treats string zero as open and positive numeric strings as closed', () => {
    const now = 2_000_000;
    expect(isClosed({ ClosedAt: '0', EndsAt: 2_100 }, now)).toBe(false);
    expect(isClosed({ ClosedAt: '1', EndsAt: 2_100 }, now)).toBe(true);
  });

  it('fails closed for missing or invalid expiry values', () => {
    const now = 2_000_000;
    expect(isClosed({ ClosedAt: 0 }, now)).toBe(true);
    expect(isClosed({ ClosedAt: 0, EndsAt: 'not-a-time' }, now)).toBe(true);
  });

  it('treats the exact expiry boundary as closed', () => {
    expect(isClosed({ ClosedAt: 0, EndsAt: 2_000 }, 2_000_000)).toBe(true);
  });
});

describe('scheduleSessionExpiry', () => {
  afterEach(() => vi.useRealTimers());

  it('notifies Chat when a live session reaches its expiry', () => {
    vi.useFakeTimers();
    vi.setSystemTime(2_000_000);
    const onExpiry = vi.fn();
    const cleanup = scheduleSessionExpiry(
      { ClosedAt: '0', EndsAt: 2_005 },
      onExpiry,
    );

    vi.advanceTimersByTime(4_999);
    expect(onExpiry).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(onExpiry).toHaveBeenCalledTimes(1);
    cleanup();
  });

  it('notifies immediately for an already closed session', () => {
    vi.useFakeTimers();
    vi.setSystemTime(2_000_000);
    const onExpiry = vi.fn();

    scheduleSessionExpiry({ ClosedAt: 1, EndsAt: 2_005 }, onExpiry);

    expect(onExpiry).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
  });
});

describe('isCoworkCandidate', () => {
  it('accepts LLM and chat-tagged legacy marketplace models', () => {
    expect(isCoworkCandidate({ ModelType: 'llm' })).toBe(true);
    expect(isCoworkCandidate({ ModelType: ' LLM ' })).toBe(true);
    expect(isCoworkCandidate({ ModelType: 'UNKNOWN', Tags: ['chat'] })).toBe(
      true,
    );
    expect(isCoworkCandidate({ ModelType: ' UNKNOWN ', Tags: ['chat'] })).toBe(
      true,
    );
    expect(isCoworkCandidate({ Tags: ['chat'] })).toBe(true);
    expect(isCoworkCandidate({ ModelType: 'UNKNOWN' })).toBe(true);
  });

  it('rejects audio and embedding-only models', () => {
    expect(isCoworkCandidate(undefined)).toBe(false);
    expect(isCoworkCandidate({ ModelType: 'llm', isLocal: true })).toBe(false);
    expect(isCoworkCandidate({ ModelType: 'llm', IsDeleted: true })).toBe(
      false,
    );
    expect(isCoworkCandidate({ ModelType: 'tts', Tags: ['tts'] })).toBe(false);
    expect(isCoworkCandidate({ Tags: ['tts'] })).toBe(false);
    expect(
      isCoworkCandidate({ ModelType: 'embedding', Tags: ['embedding'] }),
    ).toBe(false);
  });

  it('lets an explicit non-LLM type override contradictory tags', () => {
    expect(isCoworkCandidate({ ModelType: 'tts', Tags: ['llm'] })).toBe(false);
    expect(isCoworkCandidate({ ModelType: 'embedding', Tags: ['chat'] })).toBe(
      false,
    );
    expect(isCoworkCandidate({ ModelType: 'agent', Tags: ['chat'] })).toBe(
      false,
    );
  });

  it('handles malformed tag metadata without throwing or broadening audio models', () => {
    expect(isCoworkCandidate({ ModelType: 'UNKNOWN', Tags: null })).toBe(true);
    expect(
      isCoworkCandidate({ ModelType: 'UNKNOWN', Tags: { bad: 'shape' } }),
    ).toBe(true);
    expect(isCoworkCandidate({ Tags: 'tts' })).toBe(false);
    expect(isCoworkCandidate({ Tags: ['tts', null, { nested: true }] })).toBe(
      false,
    );
  });
});

describe('model tag helpers', () => {
  it('detects secure models from supported tag shapes', () => {
    expect(isSecureModel({ Tags: null })).toBe(false);
    expect(isSecureModel({ Tags: { bad: 'shape' } })).toBe(false);
    expect(isSecureModel({ Tags: 'llm, tee' })).toBe(true);
    expect(isSecureModel({ Tags: [null, {}, 'TEE'] })).toBe(true);
  });

  it('derives modality safely from malformed and legacy tag shapes', () => {
    expect(getModelModality({ Tags: null })).toBe('llm');
    expect(getModelModality({ Tags: { bad: 'shape' } })).toBe('llm');
    expect(getModelModality({ Tags: 'llm, tts' })).toBe('tts');
    expect(getModelModality({ Tags: [null, {}, 'embedding'] })).toBe(
      'embedding',
    );
  });

  it('prefers the canonical model type over contradictory legacy tags', () => {
    expect(getModelModality({ ModelType: 'TTS', Tags: ['llm'] })).toBe('tts');
    expect(getModelModality({ ModelType: 'EMBEDDING', Tags: ['speech'] })).toBe(
      'embedding',
    );
    expect(getModelModality({ ModelType: 'UNKNOWN', Tags: ['s2t'] })).toBe(
      'stt',
    );
  });
});
