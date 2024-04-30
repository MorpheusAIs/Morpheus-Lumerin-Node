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
    enabled: process.env.REPORT_GAS ? true : false,
    outputJSON: true,
    outputJSONFile: "gas.json",
    coinmarketcap: process.env.COINMARKETCAP_API_KEY,
    darkMode: true,
    currency: "USD",
    L2Etherscan: "E6UST5HFK6DNNTVUV1YTTRTN3BX727G8SU",
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
