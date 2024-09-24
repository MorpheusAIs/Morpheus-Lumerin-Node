import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { AbiCoder, getBytes, keccak256 } from 'ethers';

import { getChainId, getCurrentBlockTime } from './block-helper';

export const getProviderApproval = async (provider: SignerWithAddress, user: string, bidId: string, chainId = 0n) => {
  chainId = chainId || (await getChainId());
  const timestamp = await getCurrentBlockTime();
  const msg = AbiCoder.defaultAbiCoder().encode(
    ['bytes32', 'uint256', 'address', 'uint128'],
    [bidId, chainId, user, timestamp],
  );
  const signature = await provider.signMessage(getBytes(keccak256(msg)));

  return {
    msg,
    signature,
  };
};

export const getReport = async (reporter: SignerWithAddress, sessionId: string, tps: number, ttftMs: number) => {
  const timestamp = await getCurrentBlockTime();
  const msg = AbiCoder.defaultAbiCoder().encode(
    ['bytes32', 'uint256', 'uint128', 'uint32', 'uint32'],
    [sessionId, await getChainId(), timestamp, tps * 1000, ttftMs],
  );
  const signature = await reporter.signMessage(getBytes(keccak256(msg)));

  return {
    msg,
    signature,
  };
};
