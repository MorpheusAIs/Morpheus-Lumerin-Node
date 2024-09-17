import {
  type Abi,
  BaseError,
  type ContractErrorName,
  ContractFunctionRevertedError,
  type DecodeErrorResultReturnType,
} from "viem";

export function isErr<const TAbi extends Abi | readonly unknown[]>(
  err: unknown,
  // abi: TAbi,
  errName: DecodeErrorResultReturnType<TAbi>["errorName"]
): boolean {
  if (err instanceof BaseError) {
    const revertError = err.walk((err) => {
      return err instanceof ContractFunctionRevertedError;
    });

    if (revertError instanceof ContractFunctionRevertedError) {
      const errorName = revertError.data?.errorName ?? "";
      if (errorName === errName) {
        return true;
      }
    }
  }

  return false;
}

export function getErr<
  const TAbi extends Abi | readonly unknown[],
  allErrorNames extends ContractErrorName<TAbi> = ContractErrorName<TAbi>
>(
  err: unknown,
  // abi: TAbi,
  errName: allErrorNames
): DecodeErrorResultReturnType<TAbi, typeof errName> | undefined {
  if (err instanceof BaseError) {
    const revertError = err.walk((err) => {
      return err instanceof ContractFunctionRevertedError;
    });

    if (revertError instanceof ContractFunctionRevertedError) {
      const errorName = revertError.data?.errorName ?? "";
      if (errorName === errName) {
        return revertError.data;
      }
    }
  }

  return undefined;
}

export function isFunctionRevertedError(err: unknown): boolean {
  if (err instanceof BaseError) {
    const revertError = err.walk((err) => {
      return err instanceof ContractFunctionRevertedError;
    });

    if (revertError) {
      return true;
    }
  }
  return false;
}

export function errorToPOJO(error: any) {
  const ret = {} as Record<string, unknown>;
  for (const properyName of Object.getOwnPropertyNames(error)) {
    ret[properyName] = error[properyName];
  }
  return ret;
}
