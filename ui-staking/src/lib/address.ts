/** Returns short representation of an address for UI e.g. 0x123...789 */
export const shortAddress = (address: `0x${string}`): string => {
  return `${address.slice(0, 5)}...${address.slice(-3)}`;
};
