import React from "react";
import "./main2.css";
import { Router } from "./Router.tsx";
import { WagmiProvider } from "wagmi";
import { config } from "./wagmi.ts";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
// import { ConnectKitProvider } from "connectkit";
import { createWeb3Modal } from "@web3modal/wagmi/react";

const queryClient = new QueryClient();

createWeb3Modal({
	wagmiConfig: config,
	projectId: process.env.REACT_APP_WALLET_CONNECT_PROJECT_ID,
	enableAnalytics: true, // Optional - defaults to your Cloud configuration
	enableOnramp: true, // Optional - false as default
});

export const App2 = () => {
	return (
		<>
			<WagmiProvider config={config}>
				<QueryClientProvider client={queryClient}>
					{/* <ConnectKitProvider> */}
					<Router />
					{/* </ConnectKitProvider> */}
				</QueryClientProvider>
			</WagmiProvider>
		</>
	);
};
