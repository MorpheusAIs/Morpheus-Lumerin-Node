export function formatDate(num: bigint | number): string {
  return new Date(Number(num) * 1000).toLocaleString();
}
