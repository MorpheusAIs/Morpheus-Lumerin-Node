import "dotenv/config";
import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox-viem";
import "@solarity/hardhat-gobind";

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  mocha: {
    reporter: "nyan",
  },
  gasReporter: {
    outputJSON: true,
    coinmarketcap: process.env.COINMARKETCAP_API_KEY,
    darkMode: true,
    currency: "USD",
    L2: "arbitrum",
    L1: "ethereum",
  },
  gobind: {
    outdir: "./bindings/go",
    onlyFiles: ["./contracts"],
    skipFiles: ["./contracts/KeySet.sol"],
  },
};

export default config;
