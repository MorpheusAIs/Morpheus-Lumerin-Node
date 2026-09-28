import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  CUSTOM_INSTRUCTIONS_KEY,
  CUSTOM_INSTRUCTIONS_MAX_LENGTH,
  customInstructionsMessage,
  loadCustomInstructions,
  saveCustomInstructions,
  withCustomInstructions,
} from './customInstructions';

const userMessage = { role: 'user', content: 'hello' };

// Restoring globally matters: one test below replaces Storage.prototype, and a
// mock that leaked into the next describe would silently disable storage there.
beforeEach(() => {
  vi.restoreAllMocks();
  window.localStorage.clear();
});

describe('customInstructions storage', () => {
  it('reads an empty string when nothing has been saved', () => {
    expect(loadCustomInstructions()).toBe('');
  });

  it('round-trips a saved instruction', () => {
    saveCustomInstructions('Answer in British English.');
    expect(window.localStorage.getItem(CUSTOM_INSTRUCTIONS_KEY)).toBe(
      'Answer in British English.',
    );
    expect(loadCustomInstructions()).toBe('Answer in British English.');
  });

  it('trims on the way in so padding never becomes an instruction', () => {
    saveCustomInstructions('   be terse   ');
    expect(loadCustomInstructions()).toBe('be terse');
  });

  it('removes the key when saving blank, so the feature can be turned off', () => {
    saveCustomInstructions('be terse');
    saveCustomInstructions('   \n  ');
    expect(window.localStorage.getItem(CUSTOM_INSTRUCTIONS_KEY)).toBeNull();
    expect(loadCustomInstructions()).toBe('');
  });

  it('survives storage that throws', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('denied');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('denied');
    });
    expect(loadCustomInstructions()).toBe('');
    expect(() => saveCustomInstructions('anything')).not.toThrow();
  });
});

describe('customInstructionsMessage', () => {
  it('is null when there are no instructions', () => {
    expect(customInstructionsMessage('')).toBeNull();
    expect(customInstructionsMessage('   \n\t ')).toBeNull();
  });

  it('is null when storage is empty and no argument is given', () => {
    expect(customInstructionsMessage()).toBeNull();
  });

  it('builds a system message from the stored value', () => {
    saveCustomInstructions('Use metric units.');
    expect(customInstructionsMessage()).toEqual({
      role: 'system',
      content: 'Use metric units.',
    });
  });

  it('clamps to the maximum length', () => {
    const long = 'x'.repeat(CUSTOM_INSTRUCTIONS_MAX_LENGTH + 500);
    const message = customInstructionsMessage(long);
    expect(message?.content).toHaveLength(CUSTOM_INSTRUCTIONS_MAX_LENGTH);
  });
});

describe('withCustomInstructions', () => {
  it('is completely inert when the user has written nothing', () => {
    const messages = [userMessage];
    const result = withCustomInstructions(messages);
    expect(result).toBe(messages);
    expect(result).toHaveLength(1);
  });

  it('prepends exactly one system message', () => {
    saveCustomInstructions('Always show code first.');
    const result = withCustomInstructions([userMessage]);
    expect(result).toEqual([
      { role: 'system', content: 'Always show code first.' },
      userMessage,
    ]);
    expect(result.filter((m: any) => m.role === 'system')).toHaveLength(1);
  });

  it('prefers an explicitly passed instruction over storage', () => {
    saveCustomInstructions('stored');
    const result = withCustomInstructions([userMessage], 'explicit');
    expect((result[0] as any).content).toBe('explicit');
  });

  it('leaves the original array untouched', () => {
    saveCustomInstructions('be terse');
    const messages = [userMessage];
    withCustomInstructions(messages);
    expect(messages).toEqual([userMessage]);
  });
});
