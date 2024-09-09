import { http } from "wagmi";
import { hardhat, type Chain, sepolia, arbitrumSepolia, arbitrum } from "wagmi/chains";
import { defaultWagmiConfig } from "@web3modal/wagmi/react/config";

const supportedChains: Record<number, Chain> = {
  [hardhat.id]: hardhat,
  [sepolia.id]: sepolia,
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
    name: "Lumerin Morpheus Staking",
    description: "Stake your LMR tokens to earn rewards in MOR",
    url: process.env.REACT_APP_URL,
    icons: ["https://avatars.githubusercontent.com/u/37784886"],
  },
});
