// import { ConnectKitButton } from "connectkit";
import { Link } from "react-router-dom";
import { LumerinLogo } from "../icons/Lumerin.tsx";
import { shortAddress } from "../lib/address.ts";
import { Container } from "./Container.tsx";

type HeaderProps = {
  address?: `0x${string}`;
};

export const Header: React.FC<HeaderProps> = (props) => {
  return (
    <header>
      <Container className="header">
        <Link to="/">
          <LumerinLogo className="header-logo" />
        </Link>
        <w3m-account-button />
      </Container>
    </header>
  );
};
