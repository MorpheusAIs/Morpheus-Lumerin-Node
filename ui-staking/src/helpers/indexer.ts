import type { Chain } from "viem";

export function getTxURL(
  txhash: `0x${string}` | null | undefined,
  chain?: Chain
): string | undefined {
  if (!chain?.blockExplorers) {
    return undefined;
  }
  return `${chain.blockExplorers?.default.url}/tx/${txhash}`;
}
