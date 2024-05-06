import { PublicClient } from "@nomicfoundation/hardhat-viem/types";
import {
  Abi,
  BaseError,
  ContractFunctionRevertedError,
  UnknownRpcError,
} from "viem";
import {
  DecodeErrorResultReturnType,
  decodeErrorResult,
  padHex,
} from "viem/utils";
import crypto from "crypto";

export async function getTxTimestamp(
  client: PublicClient,
  txHash: `0x${string}`,
): Promise<bigint> {
  const receipt = await client.waitForTransactionReceipt({
    hash: txHash,
    timeout: 1000,
  });
  const block = await client.getBlock({ blockNumber: receipt.blockNumber });
  return block.timestamp;
}

/** helper function to catch errors and check if the error is the expected one
 * @example
 * await catchError(abi, "ErrorName", async () => {
 *   await contract.method();
 * });
**/
export async function catchError<const TAbi extends Abi | readonly unknown[]>(
  abi: TAbi | undefined,
  error: DecodeErrorResultReturnType<TAbi>["errorName"],
  cb: () => Promise<unknown>,
) {
  try {
    await cb();
    throw new Error(`No error was thrown, expected error "${error}"`);
  } catch (err) {
    expectError(err, abi, error);
  }
}

export function expectError<const TAbi extends Abi | readonly unknown[]>(
  err: any,
  abi: TAbi | undefined,
  error: DecodeErrorResultReturnType<TAbi>["errorName"],
) {
  if (!isErr(err, abi, error)) {
    console.error(err);
    throw new Error(
      `Expected blockchain custom error "${error}" was not thrown\n\n${err}`,
      {
        cause: err,
      },
    );
  }
}

export function isErr<const TAbi extends Abi | readonly unknown[]>(
  err: any,
  abi: TAbi | undefined,
  error: DecodeErrorResultReturnType<TAbi>["errorName"],
): boolean {
  if (err instanceof BaseError) {
    const revertError = err.walk(
      (err) =>
        err instanceof ContractFunctionRevertedError ||
        err instanceof UnknownRpcError,
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

export const now = (): bigint => {
  return BigInt(Math.floor(Date.now() / 1000));
};