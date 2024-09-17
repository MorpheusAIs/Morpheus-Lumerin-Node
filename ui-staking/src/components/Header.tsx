import { Link } from "react-router-dom";
import { LumerinLogo } from "../icons/Lumerin.tsx";
import { shortAddress } from "../lib/address.ts";
import { Container } from "./Container.tsx";
import { useWeb3Modal, useWeb3ModalState, useWeb3ModalTheme } from "@web3modal/wagmi/react";
import { useAccount } from "wagmi";
import AvatarImport from "boring-avatars";

const Avatar: typeof AvatarImport = AvatarImport.default;

type HeaderProps = {
  address?: `0x${string}`;
  hideWallet?: boolean;
};

export const Header: React.FC<HeaderProps> = (props) => {
  const { open } = useWeb3Modal();
  const { address } = useAccount();

  return (
    <header>
      <Container className="header">
        <Link to="/">
          <LumerinLogo className="header-logo" />
        </Link>
        {!props.hideWallet && (
          <>
            {address ? (
              <>
                <button
                  type="button"
                  className="header-wallet"
                  onClick={() => open({ view: "Account" })}
                >
                  <Avatar
                    size="24px"
                    name={address}
                    variant="marble"
                    colors={["#1876D1", "#9A5AF7", "#CF9893", "#849483", "#4E937A"]}
                  />
                  {shortAddress(address)}
                </button>
              </>
            ) : (
              <button
                type="button"
                className="header-wallet"
                onClick={() => open({ view: "Connect" })}
              >
                Connect wallet
              </button>
            )}
          </>
        )}
      </Container>
    </header>
  );
};
