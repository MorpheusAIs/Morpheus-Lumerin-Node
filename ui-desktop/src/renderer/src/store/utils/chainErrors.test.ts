import { describe, expect, it } from 'vitest';
import { explainChainError, formatChainError } from './chainErrors';

// The strings below are verbatim from a real failing session-open, including
// the proxy-router's layered wrapping.
const READ_ONLY_RPC =
  'failed to send transaction: open session failed: failed to send transaction: method is not allowed on this endpoint';
const INSUFFICIENT_GAS =
  'failed to send transaction: open session failed: insufficient funds for gas * price + value: have 3150354941278 want 8624200817020';

describe('explainChainError', () => {
  describe('read-only RPC endpoint', () => {
    it('identifies the cause through the nested wrapping', () => {
      const { message, hint } = explainChainError(READ_ONLY_RPC);
      expect(message).toMatch(/refuses to broadcast transactions/i);
      expect(hint).toMatch(/ETH_NODE_ADDRESS/);
    });

    it('does not blame the user for having no funds', () => {
      expect(explainChainError(READ_ONLY_RPC).message).not.toMatch(/balance|funds|top up/i);
    });

    it('recognises the common variants', () => {
      for (const msg of [
        'method not allowed',
        'method not found',
        'unsupported method: eth_sendRawTransaction',
        'the method does not exist (-32601)',
      ]) {
        expect(explainChainError(msg).message).toMatch(/refuses to broadcast/i);
      }
    });
  });

  describe('insufficient gas', () => {
    it('converts wei to readable ETH', () => {
      const { message } = explainChainError(INSUFFICIENT_GAS);
      // 3150354941278 wei and 8624200817020 wei
      expect(message).toContain('0.00000315');
      expect(message).toContain('0.00000862');
    });

    it('says this is ETH for gas, not MOR', () => {
      const { message, hint } = explainChainError(INSUFFICIENT_GAS);
      expect(message).toMatch(/network fee/i);
      expect(hint).toMatch(/separate from your MOR/i);
    });

    it('still explains itself when the amounts cannot be parsed', () => {
      const { message } = explainChainError('insufficient funds');
      expect(message).toMatch(/not enough eth/i);
    });
  });

  it('distinguishes the two failure modes, which need different fixes', () => {
    const rpc = explainChainError(READ_ONLY_RPC);
    const gas = explainChainError(INSUFFICIENT_GAS);
    expect(rpc.message).not.toBe(gas.message);
    expect(rpc.hint).not.toBe(gas.hint);
  });

  it.each([
    ['rate limit', '429 too many requests', /rate-limiting/i],
    ['nonce', 'nonce too low', /still pending/i],
    ['revert', 'execution reverted', /contract rejected/i],
    ['allowance', 'insufficient allowance', /not enough MOR/i],
    ['offline', 'Cannot reach the local proxy-router.', /cannot reach/i],
  ])('classifies %s', (_label, input, expected) => {
    expect(explainChainError(input).message).toMatch(expected);
  });

  it('preserves the original text for logs', () => {
    expect(explainChainError(READ_ONLY_RPC).raw).toBe(READ_ONLY_RPC);
  });

  it('strips the repetitive wrapper from unrecognised errors', () => {
    const { message } = explainChainError(
      'failed to send transaction: open session failed: something entirely new',
    );
    expect(message).toBe('something entirely new');
    expect(message).not.toMatch(/failed to send transaction/);
  });

  it('accepts an Error instance as well as a string', () => {
    expect(explainChainError(new Error(READ_ONLY_RPC)).message).toMatch(
      /refuses to broadcast/i,
    );
  });

  it.each([[null], [undefined], ['']])('does not throw on %s', (input) => {
    expect(() => explainChainError(input)).not.toThrow();
  });
});

// Verbatim provider refusal, several layers of JSON deep.
const VISION_REJECTION =
  'provider request failed: provider error: upstream error 400: {"details":{"_errors":[],"messages":{"_errors":["Image content is not supported by this model. Please use a model that supports vision."]}},"error":"Invalid request parameters","issues":[{"code":"custom","message":"Image content is not supported by this model. Please use a model that supports vision.","path":["messages"]}]}';

describe('model rejected an image', () => {
  it('extracts the actionable point from the JSON wrapping', () => {
    const { message, hint } = explainChainError(VISION_REJECTION);
    expect(message).toBe('This model cannot read images.');
    expect(hint).toMatch(/vision-capable model/);
  });

  // The user's instinct on a failure is that the whole message was lost.
  // Saying what still worked prevents a pointless retry from scratch.
  it('says the text and documents were fine', () => {
    expect(explainChainError(VISION_REJECTION).hint).toMatch(/documents were fine/i);
  });

  it('does not leak the raw JSON into the message', () => {
    const { message } = explainChainError(VISION_REJECTION);
    expect(message).not.toContain('{');
    expect(message).not.toContain('upstream error');
  });

  it('is distinct from the read-only RPC and gas cases', () => {
    const vision = explainChainError(VISION_REJECTION).message;
    expect(vision).not.toBe(explainChainError(READ_ONLY_RPC).message);
    expect(vision).not.toBe(explainChainError(INSUFFICIENT_GAS).message);
  });
});

describe('formatChainError', () => {
  it('joins message and hint into one line', () => {
    const out = formatChainError(READ_ONLY_RPC);
    expect(out).toMatch(/refuses to broadcast/i);
    expect(out).toMatch(/ETH_NODE_ADDRESS/);
  });
});
