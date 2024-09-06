import React from "react";
import "./App.css";
import { Router } from "./Router.tsx";
import { WagmiProvider } from "wagmi";
import { config } from "./config/wagmi.ts";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createWeb3Modal } from "@web3modal/wagmi/react";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

const queryClient = new QueryClient();

createWeb3Modal({
  tokens: {},
  wagmiConfig: config,
  projectId: process.env.REACT_APP_WALLET_CONNECT_PROJECT_ID,
  enableAnalytics: true, // Optional - defaults to your Cloud configuration
  enableOnramp: false, // Optional - false as default
  enableSwaps: false,
  themeVariables: {
    "--w3m-border-radius-master": "12px",
    "--w3m-font-family": "APK Protocol, sans-serif",
    "--w3m-accent": "#1976d2",
    // "--w3m-color-mix": "#fff",
    // "--w3m-color-mix-strength": 1,
    // "--w3m-font-size-master": "14px",
  },
});

export const App = () => {
  return (
    <>
      <WagmiProvider config={config}>
        <QueryClientProvider client={queryClient}>
          <ReactQueryDevtools position="right" />
          <Router />
          <svg width="0" height="0">
            <title>SVG gradients</title>
            <defs>
              <linearGradient
                id="cl1"
                gradientUnits="objectBoundingBox"
                x1="0"
                y1="0.5"
                x2="1"
                y2="0.5"
              >
                <stop stopColor="#A855F7" />
                <stop offset="100%" stopColor="#3B82F6" />
              </linearGradient>
            </defs>
          </svg>
        </QueryClientProvider>
      </WagmiProvider>
    </>
  );
};
