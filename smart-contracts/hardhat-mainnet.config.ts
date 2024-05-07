import baseConfig from "./hardhat.config";
import { HardhatUserConfig } from "hardhat/types";

if (!process.env.ETH_NODE_ADDRESS) {
  throw new Error("ETH_NODE_ADDRESS env variable is not set");
}

if (!process.env.OWNER_PRIVATE_KEY) {
  throw new Error("OWNER_PRIVATE_KEY env variable is not set");
}

if (!process.env.ETHERSCAN_API_KEY) {
  throw new Error("ETHERSCAN_API_KEY env variable is not set");
}

const config: HardhatUserConfig = {
  ...baseConfig,
  networks: {
    ...baseConfig.networks,
    default: {
      chainId: 421614,
      url: process.env.ETH_NODE_ADDRESS,
      accounts: [process.env.OWNER_PRIVATE_KEY],
    },
  },
  etherscan: {
    apiKey: {
      arbitrumSepolia: process.env.ETHERSCAN_API_KEY,
    },
  },
  sourcify: {
    enabled: false,
  },
};

export default config;
