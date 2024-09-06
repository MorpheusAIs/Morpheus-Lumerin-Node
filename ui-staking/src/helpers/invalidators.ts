import type { Query, QueryKey } from "@tanstack/react-query";

export function filterPoolQuery(poolId: bigint) {
  return (q: Query<unknown, Error, unknown, QueryKey>) => {
    const params = q.queryKey?.[1];
    if (!params) {
      return false;
    }
    if (params?.functionName === "pools" && params?.args?.[0] === BigInt(poolId)) {
      console.log("invalidating pools", params);

      return true;
    }
    return false;
  };
}

export function filterStakeQuery(poolId: bigint) {
  return (q: Query<unknown, Error, unknown, QueryKey>) => {
    const params = q.queryKey?.[1];
    if (!params) {
      return false;
    }
    if (params?.functionName === "getStakes" && params?.args?.[1] === BigInt(poolId)) {
      console.log("invalidating getStakes", params);
      return true;
    }
    return false;
  };
}

export function filterUserBalanceQuery(address: `0x${string}`) {
  return (q: Query<unknown, Error, unknown, QueryKey>) => {
    const params = q.queryKey?.[1];
    if (!params) {
      return false;
    }
    if (
      params?.functionName === "balanceOf" &&
      params?.address === process.env.REACT_APP_LMR_ADDR &&
      params?.args?.[0] === address
    ) {
      return true;
    }
    return false;
  };
}
