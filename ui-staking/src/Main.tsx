import React from "react";
import { useAccount, useBlock, useConnect, useDisconnect, useReadContract } from "wagmi";
import { StakeList } from "./StakeList.tsx";
import { erc20Abi } from "viem";
import { StakeAdd } from "./StakeAdd.tsx";
import { stakingMasterChefAbi } from "./blockchain/abi.ts";
import { formatDate } from "./lib/date.ts";
import { useQueryClient } from "@tanstack/react-query";
import { BalanceLMR, BalanceMOR } from "./balance.tsx";

export const Main = () => {
  const { address, isConnected } = useAccount();

  const { connectors, connect } = useConnect();
  const { connectors: connectedConnectors, disconnect } = useDisconnect();

  return (
    <>
      <h2>Connectors</h2>
      {connectors.map((connector) => (
        <button type="button" key={connector.uid} onClick={() => connect({ connector })}>
          {connector.name}
        </button>
      ))}
      <h2>Connected Connectors</h2>
      {isConnected &&
        connectedConnectors.map((connector) => (
          <button type="button" key={connector.uid} onClick={() => disconnect()}>
            Disconnect {connector.name}
          </button>
        ))}

      {address ? <Connected address={address} /> : <p>Connect to see your staking information</p>}
    </>
  );
};

export const Connected = (props: { address: `0x${string}` }) => {
  const queryClient = useQueryClient();

  const precision = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "PRECISION",
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
    },
  });

  const balanceMor = useReadContract({
    abi: erc20Abi,
    functionName: "balanceOf",
    address: process.env.REACT_APP_MOR_ADDR as `0x${string}`,
    args: [props.address],
  });

  const balanceLMR = useReadContract({
    abi: erc20Abi,
    functionName: "balanceOf",
    address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
    args: [props.address],
  });

  const block = useBlock({
    blockTag: "latest",
    // watch: {
    // 	enabled: true,
    // 	poll: true,
    // 	pollingInterval: 10_000,
    // 	syncConnectedChain: true,
    // },
  });

  const [poolId, setPoolId] = React.useState(0n);

  const poolData = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "pools",
    args: [poolId],
  });

  const poolBalanceLMR = useReadContract({
    abi: erc20Abi,
    address: process.env.REACT_APP_LMR_ADDR as `0x${string}`,
    functionName: "balanceOf",
    args: [process.env.REACT_APP_STAKING_ADDR as `0x${string}`],
  });

  const poolBalanceMOR = useReadContract({
    abi: erc20Abi,
    address: process.env.REACT_APP_MOR_ADDR as `0x${string}`,
    functionName: "balanceOf",
    args: [process.env.REACT_APP_STAKING_ADDR as `0x${string}`],
  });

  const poolDataObj = poolData.data
    ? {
        rewardPerSecondScaled: poolData.data[0],
        lastRewardTime: poolData.data[1],
        accRewardPerShareScaled: poolData.data[2],
        totalShares: poolData.data[3],
        startTime: poolData.data[4],
        endTime: poolData.data[5],
        balanceMOR: poolBalanceMOR.data,
        balanceLMR: poolBalanceLMR.data,
      }
    : undefined;

  const stakes = useReadContract({
    abi: stakingMasterChefAbi,
    address: process.env.REACT_APP_STAKING_ADDR as `0x${string}`,
    functionName: "getStakes",
    args: [props.address, poolId],
    query: {
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchOnReconnect: false,
      retry: true,
    },
  });

  return (
    <>
      <h1>Staking contract</h1>

      <h2>Account</h2>
      <p>{props.address}</p>
      <p>
        {balanceMor.isSuccess ? (
          <>
            Balance: <BalanceMOR value={balanceMor.data} />
          </>
        ) : (
          <>Loading balance...</>
        )}
      </p>

      <p>
        {balanceLMR.isSuccess ? (
          <>
            Balance: <BalanceLMR value={balanceLMR.data} />
          </>
        ) : (
          <>Loading balance...</>
        )}
      </p>

      <h2>Pool</h2>
      <input type="number" value={poolId.toString()} onChange={(e) => setPoolId(BigInt(e.target.value))} />
      {poolData.data && <p>Pool {poolId.toString()}</p>}
      {poolDataObj && (
        <p>
          rewardPerSecondScaled <BalanceMOR value={poolDataObj.accRewardPerShareScaled} />
          <br />
          lastRewardTime {formatDate(poolDataObj.lastRewardTime)},
          <br />
          accRewardPerShareScaled {<BalanceMOR value={poolDataObj.accRewardPerShareScaled} />},
          <br />
          totalShares {poolDataObj.totalShares.toString()},
          <br />
          startTime {formatDate(poolDataObj.startTime)},
          <br />
          endTime {formatDate(poolDataObj.endTime)},
          <br />
          balanceLMR {poolDataObj.balanceLMR ? <BalanceLMR value={poolDataObj.balanceLMR} /> : "Loading..."}
          <br />
          balanceMOR {poolDataObj.balanceMOR ? <BalanceMOR value={poolDataObj.balanceMOR} /> : "Loading..."}
        </p>
      )}

      {block.data && poolDataObj && stakes.data && precision.data && (
        <StakeList
          userAddr={props.address}
          poolId={0n}
          blockTimestamp={block.data?.timestamp}
          poolData={poolDataObj}
          stakes={stakes.data}
          precision={precision.data}
          onUpdate={() => {
            queryClient.invalidateQueries({ queryKey: block.queryKey });
            queryClient.invalidateQueries({ queryKey: stakes.queryKey });
            queryClient.invalidateQueries({ queryKey: balanceMor.queryKey });
            queryClient.invalidateQueries({ queryKey: balanceLMR.queryKey });
            queryClient.invalidateQueries({ queryKey: poolData.queryKey });
            queryClient.invalidateQueries({
              queryKey: poolBalanceMOR.queryKey,
            });
            queryClient.invalidateQueries({
              queryKey: poolBalanceLMR.queryKey,
            });
          }}
        />
      )}

      {balanceMor.isSuccess && precision.data ? (
        <StakeAdd
          poolId={0n}
          userBalance={balanceMor.data}
          precision={precision.data}
          onAdd={() => {
            queryClient.invalidateQueries({ queryKey: block.queryKey });
            queryClient.invalidateQueries({ queryKey: stakes.queryKey });
            queryClient.invalidateQueries({ queryKey: balanceMor.queryKey });
            queryClient.invalidateQueries({ queryKey: balanceLMR.queryKey });
            queryClient.invalidateQueries({ queryKey: poolData.queryKey });
            queryClient.invalidateQueries({
              queryKey: poolBalanceMOR.queryKey,
            });
            queryClient.invalidateQueries({
              queryKey: poolBalanceLMR.queryKey,
            });
          }}
        />
      ) : (
        <p>Loading...</p>
      )}
    </>
  );
};
