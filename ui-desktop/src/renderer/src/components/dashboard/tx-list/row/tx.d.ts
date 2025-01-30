export interface Tx {
  hash: string;
  from: string;
  to: string;
  contract: {
    contractAddress: string;
    contractName: string;
    methodName: string;
    decodedInput: { type: string; key: string; value: unknown }[];
  };
  transfers: Transfer[];
  timestamp: string;
}

export interface Transfer {
  from: string;
  to: string;
  value: string;
  tokenAddress: string;
  tokenSymbol: string;
  tokenName: string;
  tokenIcon: string;
  tokenDecimals: number;
}
