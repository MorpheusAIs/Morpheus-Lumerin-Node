import "@nomicfoundation/hardhat-chai-matchers";
import "@nomicfoundation/hardhat-ethers";
import "@nomicfoundation/hardhat-ignition-viem";
import "@nomicfoundation/hardhat-toolbox-viem";
import "@nomicfoundation/hardhat-verify";
import "@solarity/hardhat-gobind";
import "@typechain/hardhat";
import "dotenv/config";
import { HardhatUserConfig } from "hardhat/config";
import "./tasks/upgrade";

function typechainTarget() {
  const target = process.env.TYPECHAIN_TARGET;

  return target === "" || target === undefined ? "ethers-v6" : target;
}

function forceTypechain() {
  return process.env.TYPECHAIN_FORCE === "false";
}

const config: HardhatUserConfig = {
  networks: {
    hardhat: {
      initialDate: "2024-07-16T01:00:00.000Z",
      gas: "auto", // required for tests where two transactions should be mined in the same block
      // loggingEnabled: true,
    },
  },
  solidity: {
    version: "0.8.24",
    settings: {
      optimizer: {
        enabled: true,
        runs: 1000,
        details: {
          yul: true,
          constantOptimizer: true,
        },
      },
      // viaIR: true,
    },
  },
  mocha: {
    // reporter: "",
  },
  gasReporter: {
    enabled: process.env.REPORT_GAS ? true : false,
    // outputJSON: true,
    // outputJSONFile: "gas.json",
    coinmarketcap: process.env.COINMARKETCAP_API_KEY,
    // reportPureAndViewMethods: true,
    // darkMode: true,
    currency: "USD",
    // offline: true,
    // L2Etherscan: process.env.ETHERSCAN_API_KEY,
    // L2: "arbitrum",
    // L1Etherscan: process.env.ETHERSCAN_API_KEY,
    // L1: "ethereum",
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
  typechain: {
    outDir: `generated-types/${typechainTarget().split("-")[0]}`,
    target: typechainTarget(),
    alwaysGenerateOverloads: true,
    discriminateTypes: true,
    dontOverrideCompile: forceTypechain(),
  },
};

export default config;
