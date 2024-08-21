import React from "react";
import "./main2.css";
import { Router } from "./Router.tsx";
import { WagmiProvider } from "wagmi";
import { config } from "./wagmi.ts";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ConnectKitProvider } from "connectkit";

const queryClient = new QueryClient();

export const App2 = () => {
	return (
		<>
			<WagmiProvider config={config}>
				<QueryClientProvider client={queryClient}>
					<ConnectKitProvider>
						<Router />
					</ConnectKitProvider>
				</QueryClientProvider>
			</WagmiProvider>
		</>
	);
};
