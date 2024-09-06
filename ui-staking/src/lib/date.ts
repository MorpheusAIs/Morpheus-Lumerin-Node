import prettyMs from "pretty-ms";

export function formatDate(num: bigint | number): string {
  return new Date(Number(num) * 1000).toLocaleString();
}

export function formatDuration(seconds: bigint): string {
  return prettyMs(Number(seconds) * 1000, { compact: true });
}
