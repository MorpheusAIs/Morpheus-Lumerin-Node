import { http, createConfig } from "wagmi";
import { arbitrumSepolia, hardhat } from "wagmi/chains";
import { injected, metaMask, safe, walletConnect } from "wagmi/connectors";

const projectId = "0b6c36d2ed2244ffc6aa04915320b907";

// 2. Create wagmiConfig
// const metadata = {
//   name: "local-staking",
//   description: "AppKit Example",
//   url: "https://web3modal.com", // origin must match your domain & subdomain
//   icons: ["https://avatars.githubusercontent.com/u/37784886"],
// };

export const config = createConfig({
  chains: [hardhat /*arbitrumSepolia*/],
  connectors: [injected(), walletConnect({ projectId }), metaMask(), safe()],
  transports: {
    [hardhat.id]: http(process.env.REACT_APP_ETH_NODE_URL),
  },
});
