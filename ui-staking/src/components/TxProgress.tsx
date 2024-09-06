import type { Chain } from "wagmi/chains";
import { getTxURL } from "../helpers/indexer.ts";
import { Check } from "../icons/Check.tsx";
import { Spinner } from "../icons/Spinner.tsx";
import { ErrorIcon } from "../icons/Error.tsx";

interface Props {
  isTransacting: boolean;
  txHash?: `0x${string}` | null;
  error?: string | null | unknown;
}

export function TxProgress(props: Props, chain: Chain) {
  const { isTransacting, error, txHash } = props;
  const isSuccess = !isTransacting && !error && !!txHash;
  const isError = !!error;

  return (
    <>
      {isTransacting && (
        <>
          <Spinner className="tx-icon" />
          Please confirm transaction in your wallet
        </>
      )}
      {isSuccess && (
        <>
          <Check fill="#fff" className="tx-icon" />
          <a href={getTxURL(txHash, chain)}>Transaction succesfull</a>
        </>
      )}
      {isError && (
        <>
          <ErrorIcon fill="#cc1111" className="tx-icon" />
          Error: {String(error)}
        </>
      )}
    </>
  );
}
