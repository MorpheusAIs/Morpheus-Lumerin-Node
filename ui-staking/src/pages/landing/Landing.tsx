import { useAccount, useConnect, useDisconnect } from "wagmi";
import { Container } from "../../components/Container.tsx";
import { LumerinLogo } from "../../icons/Lumerin.tsx";
import homeElement from "../../images/home-element.png";
import { useNavigate } from "react-router-dom";
import { useEffect } from "react";
import { Button } from "../../components/Button.tsx";
// import { ConnectKitButton } from "connectkit";

export const Landing = () => {
  const { address, isConnected } = useAccount();
  const { connectors, connect } = useConnect();
  const { connectors: connectedConnectors, disconnect } = useDisconnect();
  const navigate = useNavigate();

  //TODO: do not redirect to pool 0 if user was already connected
  useEffect(() => {
    if (address) {
      navigate("/pool/0");
    }
  }, [address, navigate]);

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
        <div className="cta-button">
          {isConnected ? <Button onClick={() => navigate("/pool/0")}>Stake LMR</Button> : <w3m-connect-button />}
        </div>
      </Container>
    </>
  );
};
