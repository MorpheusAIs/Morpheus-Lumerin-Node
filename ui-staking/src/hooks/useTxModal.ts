import { useState } from "react";
import type { WriteContractErrorType } from "viem";
import { usePublicClient } from "wagmi";

export interface StartProps {
  approveCall?: () => Promise<`0x${string}`>;
  txCall: () => Promise<`0x${string}`>;
  onSuccess?: () => Promise<void>;
}

type Stage = "inactive" | "approving" | "approve-error" | "transacting" | "tx-error" | "done";

export const useTxModal = () => {
  const [stage, setStage] = useState<Stage>("inactive");
  const [approveTxHash, setApproveTxHash] = useState<`0x${string}` | null>(null);
  const [txHash, setTxHash] = useState<`0x${string}` | null>(null);
  const [approveError, setApproveError] = useState<WriteContractErrorType | null>(null);
  const [txError, setTxError] = useState<WriteContractErrorType | null>(null);
  const pc = usePublicClient();

  async function start(props: StartProps) {
    setStage("inactive");
    setApproveTxHash(null);
    setApproveError(null);

    if (props.approveCall) {
      setStage("approving");
      try {
        const approveTxHash = await props.approveCall();
        setApproveTxHash(approveTxHash);
        await pc?.waitForTransactionReceipt({ hash: approveTxHash });
      } catch (e) {
        setStage("approve-error");
        setApproveError(e as WriteContractErrorType);
        return;
      }
    }

    setStage("transacting");
    try {
      const txHash = await props.txCall();
      setTxHash(txHash);
      await pc?.waitForTransactionReceipt({ hash: txHash });
    } catch (e) {
      setStage("tx-error");
      setTxError(e as WriteContractErrorType);
    }

    setStage("done");
    await props.onSuccess?.();
  }

  function reset() {
    setStage("inactive");
    setApproveTxHash(null);
    setTxHash(null);
    setApproveError(null);
    setTxError(null);
  }

  return {
    stage,
    approveTxHash,
    txHash,
    approveError,
    txError,
    start,
    reset,
    isVisible: stage !== "inactive",
    isApproving: stage === "approving",
    isApproveError: stage === "approve-error",
    isApproveSuccess: stage === "transacting" || stage === "tx-error" || stage === "done",
    isTransacting: stage === "transacting",
    isTransactionSuccess: stage === "done",
    isTransactionError: stage === "tx-error",
  };
};
