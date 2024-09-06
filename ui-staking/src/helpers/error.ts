import {
  ContractFunctionRevertedError,
  TransactionExecutionError,
  UserRejectedRequestError,
} from "viem";
import type { WriteContractErrorType } from "wagmi/actions";
import { errorToPOJO } from "../lib/error.ts";

export function getDisplayErrorMessage(err: WriteContractErrorType | null): string | null {
  if (err === null) {
    return null;
  }

  console.error(errorToPOJO(err));

  if (err.cause instanceof ContractFunctionRevertedError) {
    if (err.cause.data) {
      return err.cause.data.errorName;
    }
  }

  if (err.cause instanceof TransactionExecutionError) {
    if (err.cause.cause instanceof UserRejectedRequestError) {
      return "Transaction was rejected by the user";
    }
  }

  return String(err.cause);
}
