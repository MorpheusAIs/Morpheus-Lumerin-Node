import { ConnectKitButton } from "connectkit";
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
				<LumerinLogo className="header-logo" />
				{/* {props.address && (
					<button className="header-wallet" type="button">
						{shortAddress(props.address)}
					</button>
				)} */}
				<ConnectKitButton />
			</Container>
		</header>
	);
};
