import React from "react";
import { WagmiProvider } from "wagmi";
import { config } from "./wagmi.ts";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Main } from "./Main.tsx";

const queryClient = new QueryClient();

export const App = () => {
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <Main />
      </QueryClientProvider>
    </WagmiProvider>
  );
};
