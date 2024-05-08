import "dotenv/config";
import "@nomicfoundation/hardhat-verify";
import "@nomicfoundation/hardhat-toolbox-viem";
import "@nomicfoundation/hardhat-ignition-viem";
import "@solarity/hardhat-gobind";
import "./tasks/upgrade";
import { HardhatUserConfig } from "hardhat/config";

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  mocha: {
    reporter: "nyan",
  },
  gasReporter: {
    enabled: process.env.REPORT_GAS ? true : false,
    outputJSON: true,
    outputJSONFile: "gas.json",
    coinmarketcap: process.env.COINMARKETCAP_API_KEY,
    darkMode: true,
    currency: "USD",
    L2Etherscan: process.env.ETHERSCAN_API_KEY,
    L2: "arbitrum",
    L1: "ethereum",
  },
  gobind: {
    outdir: "./bindings/go",
    onlyFiles: ["./contracts"],
    skipFiles: [
      "contracts/AppStorage.sol",
      "contracts/libraries",
      "contracts/diamond/libraries",
      "contracts/diamond/interfaces",
    ],
  },
};

export default config;
