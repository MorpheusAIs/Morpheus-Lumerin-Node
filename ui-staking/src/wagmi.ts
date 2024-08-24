import { http, createConfig } from "wagmi";
import { hardhat, arbitrumSepolia, arbitrum, type Chain } from "wagmi/chains";
import { defaultWagmiConfig } from "@web3modal/wagmi/react/config";
import { createWeb3Modal } from "@web3modal/wagmi/react";

// import { injected, walletConnect } from "wagmi/connectors";
// import { getDefaultConfig } from "connectkit";
const supportedChains: Record<number, Chain> = {
  [hardhat.id]: hardhat,
  [arbitrumSepolia.id]: arbitrumSepolia,
  [arbitrum.id]: arbitrum,
};

const chain = supportedChains[process.env.REACT_APP_CHAIN_ID];
if (!chain) {
  throw new Error(`Unsupported chain ID: ${process.env.REACT_APP_CHAIN_ID}`);
}

export const config = defaultWagmiConfig({
  projectId: process.env.REACT_APP_WALLET_CONNECT_PROJECT_ID,
  chains: [chain],
  transports: {
    [process.env.REACT_APP_CHAIN_ID]: http(process.env.REACT_APP_ETH_NODE_URL),
  },
  metadata: {
    name: "WAGMI",
    description: "WAGMI",
    url: "https://wagmi.io",
    icons: ["https://avatars.githubusercontent.com/u/37784886"],
  },
});

//)
// );
