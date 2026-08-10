// Turns raw on-chain / RPC error text into something a user can act on.
//
// The proxy-router wraps failures several layers deep, so the UI was showing
// things like:
//
//   Failed to open session: "failed to send transaction: open session failed:
//   failed to send transaction: method is not allowed on this endpoint"
//
// which tells the user nothing about what to do. Worse, the two most common
// causes — no ETH for gas, and an RPC endpoint that refuses to broadcast
// transactions — have completely different fixes and are indistinguishable in
// that text.

/** Formats a wei amount as ETH with enough precision to be meaningful. */
const weiToEth = (wei: string): string => {
  try {
    const v = BigInt(wei);
    const whole = v / 10n ** 18n;
    const frac = (v % 10n ** 18n).toString().padStart(18, '0').slice(0, 8);
    return `${whole}.${frac}`.replace(/0+$/, '').replace(/\.$/, '.0');
  } catch {
    return wei;
  }
};

export type FriendlyError = {
  /** Short, plain-language summary. */
  message: string;
  /** Concrete next step, when there is an unambiguous one. */
  hint?: string;
  /** The original text, for logs and the details view. */
  raw: string;
};

export function explainChainError(error: unknown): FriendlyError {
  const raw =
    typeof error === 'string'
      ? error
      : ((error as any)?.message ?? String(error ?? 'Unknown error'));
  const msg = raw.toLowerCase();

  // ---- Model rejected an image -------------------------------------------
  // Providers wrap this in several layers of JSON, e.g.
  //   provider request failed: provider error: upstream error 400:
  //   {"details":{...},"issues":[{"message":"Image content is not supported
  //   by this model. Please use a model that supports vision."}]}
  // which is unreadable, and the actionable part is one clause in the middle.
  if (
    (msg.includes('image') &&
      (msg.includes('not supported') || msg.includes('does not support'))) ||
    msg.includes('supports vision') ||
    msg.includes('vision model')
  ) {
    return {
      message: 'This model cannot read images.',
      hint:
        'Your text and any documents were fine — only the image was rejected. ' +
        'Remove it, or switch to a vision-capable model.',
      raw,
    };
  }

  // ---- RPC endpoint refuses to broadcast transactions ----------------------
  // Base's public RPC (https://mainnet.base.org) accepts reads but rejects
  // eth_sendRawTransaction. It is first in the proxy-router's public endpoint
  // list, so with no ETH_NODE_ADDRESS configured every write fails this way.
  // Keep this list in step with shouldRetryRPCError in
  // proxy-router/internal/repositories/ethclient/rpcclientmultiple.go — that
  // decides whether to rotate endpoints, this decides what to tell the user.
  if (
    msg.includes('method is not allowed') ||
    msg.includes('method not allowed') ||
    msg.includes('not allowed on this endpoint') ||
    msg.includes('method is not available') ||
    msg.includes('method not found') ||
    msg.includes('method not supported') ||
    msg.includes('unsupported method') ||
    msg.includes('-32601')
  ) {
    return {
      message:
        'Your Ethereum RPC endpoint accepts reads but refuses to broadcast transactions.',
      hint:
        'Set ETH_NODE_ADDRESS to an RPC that allows eth_sendRawTransaction ' +
        '(Alchemy, Infura, QuickNode, or your own node). The default public ' +
        'Base endpoint is read-only.',
      raw,
    };
  }

  // ---- Not enough ETH for gas ---------------------------------------------
  // geth phrasing: "insufficient funds for gas * price + value: have X want Y"
  if (msg.includes('insufficient funds')) {
    const m = raw.match(/have (\d+).*?want (\d+)/i);
    if (m) {
      const have = weiToEth(m[1]);
      const want = weiToEth(m[2]);
      return {
        message: `Not enough ETH to cover the network fee. You have ${have} ETH but this transaction needs about ${want} ETH.`,
        hint: 'Top up the ETH balance on this wallet. This is gas, separate from your MOR.',
        raw,
      };
    }
    return {
      message: 'Not enough ETH to cover the network fee.',
      hint: 'Top up the ETH balance on this wallet. This is gas, separate from your MOR.',
      raw,
    };
  }

  // ---- Token allowance / balance ------------------------------------------
  if (msg.includes('transfer amount exceeds balance') || msg.includes('insufficient allowance')) {
    return {
      message: 'Not enough MOR, or the contract is not approved to spend it.',
      hint: 'Check your MOR balance on the Wallet tab.',
      raw,
    };
  }

  // ---- Rate limiting -------------------------------------------------------
  if (
    msg.includes('429') ||
    msg.includes('rate limit') ||
    msg.includes('too many requests') ||
    msg.includes('quota exceeded')
  ) {
    return {
      message: 'The Ethereum RPC endpoint is rate-limiting this node.',
      hint: 'Wait a moment and retry, or configure a dedicated RPC via ETH_NODE_ADDRESS.',
      raw,
    };
  }

  // ---- Nonce / replacement -------------------------------------------------
  if (msg.includes('nonce too low') || msg.includes('replacement transaction underpriced')) {
    return {
      message: 'A previous transaction from this wallet is still pending.',
      hint: 'Wait for it to confirm before trying again.',
      raw,
    };
  }

  // ---- Reverts -------------------------------------------------------------
  if (msg.includes('execution reverted') || msg.includes('revert')) {
    return {
      message: 'The contract rejected this transaction.',
      hint: 'The session terms may have changed — refresh the model list and retry.',
      raw,
    };
  }

  // ---- Connectivity --------------------------------------------------------
  if (
    msg.includes('cannot reach') ||
    msg.includes('econnrefused') ||
    msg.includes('connection refused') ||
    msg.includes('failed to fetch')
  ) {
    return {
      message: 'Cannot reach the local proxy-router.',
      hint: 'Check that it is running, on the Settings tab.',
      raw,
    };
  }

  // Unknown: strip the repetitive wrapper prefixes so at least the innermost
  // cause is what the user reads first.
  const inner = raw
    .split(/:\s*/)
    .filter((p) => !/^(failed to send transaction|open session failed|error)$/i.test(p.trim()))
    .join(': ')
    .trim();

  return { message: inner || raw, raw };
}

/** Convenience: one-line string suitable for a toast. */
export function formatChainError(error: unknown): string {
  const { message, hint } = explainChainError(error);
  return hint ? `${message} ${hint}` : message;
}
