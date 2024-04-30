import { PublicClient } from "@nomicfoundation/hardhat-viem/types";
import {
  Abi,
  AbiFunction,
  AbiItem,
  BaseError,
  ContractFunctionRevertedError,
  UnknownRpcError,
} from "viem";
import { DecodeErrorResultReturnType, decodeErrorResult, padHex, toFunctionHash } from "viem/utils";
import crypto from "crypto";

export async function getTxTimestamp(client: PublicClient, txHash: `0x${string}`): Promise<bigint> {
  const receipt = await client.waitForTransactionReceipt({ hash: txHash, timeout: 1000 });
  const block = await client.getBlock({ blockNumber: receipt.blockNumber });
  return block.timestamp;
}

export function expectError<const TAbi extends Abi | readonly unknown[]>(
  err: any,
  abi: TAbi | undefined,
  error: DecodeErrorResultReturnType<TAbi>["errorName"]
) {
  if (!catchErr(err, abi, error)) {
    console.error(err);
    throw new Error(`Expected blockchain custom error "${error}" was not thrown\n\n${err}`, {
      cause: err,
    });
  }
}

export function catchErr<const TAbi extends Abi | readonly unknown[]>(
  err: any,
  abi: TAbi | undefined,
  error: DecodeErrorResultReturnType<TAbi>["errorName"]
): boolean {
  if (err instanceof BaseError) {
    const revertError = err.walk(
      (err) => err instanceof ContractFunctionRevertedError || err instanceof UnknownRpcError
    );

    // support for regular provider
    if (revertError instanceof ContractFunctionRevertedError) {
      const errorName = revertError.data?.errorName ?? "";
      if (errorName === error) {
        return true;
      }
    }

    // support for hardhat node
    if (revertError instanceof UnknownRpcError) {
      const cause = revertError.cause as any;
      if (cause.data) {
        try {
          const decodedError = decodeErrorResult({ abi, data: cause.data });
          if (decodedError.errorName === error) {
            return true;
          }
        } catch (e) {
          console.error(e);
          return false;
        }
      }
    }
  }
  return false;
}

export const getHex = (buffer: Buffer, padding = 32): `0x${string}` => {
  return padHex(`0x${buffer.toString("hex")}`, { size: padding });
};

export const randomBytes32 = (): `0x${string}` => {
  return getHex(crypto.randomBytes(32));
};

export const randomAddress = (): `0x${string}` => {
  return getHex(crypto.randomBytes(20), 20);
};

export function getSelectors(abi: Abi) {
  return abi.filter(isFunctionExceptInitAbi).map((item) => {
    const hash = toFunctionHash(item);
    // return "0x" + 4 bytes of the hash
    return hash.slice(0, 2 + 8);
  });
}

export function isFunctionExceptInitAbi(abi: AbiItem): abi is AbiFunction {
  return abi.type === "function" && abi.name !== "init";
}