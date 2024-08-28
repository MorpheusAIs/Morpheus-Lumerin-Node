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
          <svg width="0" height="0">
            <title>SVG gradients</title>
            <defs>
              <linearGradient id="cl1" gradientUnits="objectBoundingBox" x1="0" y1="0.5" x2="1" y2="0.5">
                <stop stop-color="#A855F7" />
                <stop offset="100%" stop-color="#3B82F6" />
              </linearGradient>
            </defs>
          </svg>
          {/* </ConnectKitProvider> */}
        </QueryClientProvider>
      </WagmiProvider>
    </>
  );
};
