import { useAccount, useConnect, useDisconnect } from "wagmi";
import { Container } from "../../components/Container.tsx";
import { LumerinLogo } from "../../icons/Lumerin.tsx";
import homeElement from "../../images/home-element.png";
import { useNavigate } from "react-router-dom";
import { useEffect } from "react";
import { Button } from "../../components/Button.tsx";
import { useWeb3Modal, useWeb3ModalEvents } from "@web3modal/wagmi/react";
// import { ConnectKitButton } from "connectkit";

export const Landing = () => {
  const { address, isConnected } = useAccount();
  const { connectors, connect } = useConnect();
  const { connectors: connectedConnectors, disconnect } = useDisconnect();
  const navigate = useNavigate();
  const { open } = useWeb3Modal();
  const event = useWeb3ModalEvents();

  useEffect(() => {
    if (event.data.event === "CONNECT_SUCCESS") {
      navigate("/pool/0");
    }
  }, [event, navigate]);

  return (
    <>
      <img className="home-element" src={homeElement} alt="Home Element" />
      <Container>
        <div className="lens" />
        <LumerinLogo className="header-logo" />
        <h1 className="cta">
          Stake LMR,
          <br />
          Earn MOR
        </h1>
        <h2 className="sub-cta">Your Pathway to Effortless Rewards</h2>
        <div className="cta-buttons">
          {isConnected ? (
            <>
              <Button className="button-primary" onClick={() => navigate("/pool/0")}>
                Stake LMR
              </Button>
              <Button onClick={() => disconnect()}>Disconnect</Button>
            </>
          ) : (
            <Button onClick={() => open({ view: "Connect" })}>Connect Wallet</Button>
          )}
        </div>
      </Container>
    </>
  );
};
