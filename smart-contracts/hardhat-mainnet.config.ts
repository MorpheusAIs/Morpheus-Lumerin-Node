import baseConfig from "./hardhat.config";
import type { HardhatUserConfig } from "hardhat/types";

if (!process.env.ETH_NODE_ADDRESS) {
  throw new Error("ETH_NODE_ADDRESS env variable is not set");
}

if (!process.env.OWNER_PRIVATE_KEY) {
  throw new Error("OWNER_PRIVATE_KEY env variable is not set");
}

if (!process.env.ETHERSCAN_API_KEY) {
  throw new Error("ETHERSCAN_API_KEY env variable is not set");
}

if (!process.env.CHAIN_ID) {
  throw new Error("CHAIN_ID env variable is not set");
}

const config: HardhatUserConfig = {
  ...baseConfig,
  networks: {
    ...baseConfig.networks,
    default: {
      chainId: Number(process.env.CHAIN_ID),
      url: process.env.ETH_NODE_ADDRESS,
      accounts: [process.env.OWNER_PRIVATE_KEY],
    },
  },
  etherscan: {
    apiKey: {
      sepolia: process.env.ETHERSCAN_API_KEY,
      arbitrumSepolia: process.env.ETHERSCAN_API_KEY,
      arbitrumOne: process.env.ETHERSCAN_API_KEY,
    },
  },
  sourcify: {
    enabled: true,
  },
};

export default config;
