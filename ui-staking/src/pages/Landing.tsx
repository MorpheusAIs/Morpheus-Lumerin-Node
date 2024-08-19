import { Button } from "../components/Button.tsx";
import { Container } from "../components/Container.tsx";
import { LumerinLogo } from "../icons/Lumerin.tsx";
import homeElement from "../images/home-element.png";

export const Landing = () => {
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
					<Button>Connect Wallet</Button>
				</div>
			</Container>
		</>
	);
};
