import '@nomicfoundation/hardhat-chai-matchers';
import '@nomicfoundation/hardhat-ethers';
import '@solarity/hardhat-gobind';
import '@solarity/hardhat-markup';
import '@solarity/hardhat-migrate';
import '@typechain/hardhat';
import * as dotenv from 'dotenv';
import { HardhatUserConfig } from 'hardhat/types';
import 'solidity-coverage';
import 'tsconfig-paths/register';

dotenv.config();

function typechainTarget() {
  const target = process.env.TYPECHAIN_TARGET;

  return target === '' || target === undefined ? 'ethers-v6' : target;
}

function forceTypechain() {
  return process.env.TYPECHAIN_FORCE === 'false';
}

const config: HardhatUserConfig = {
  networks: {
    hardhat: {
      initialDate: '2024-07-15T01:00:00.000Z',
      gas: 'auto', // required for tests where two transactions should be mined in the same block
      // loggingEnabled: true,
      // mining: {
      //   auto: true,
      //   interval: 10_000,
      // },
    },
    localhost: {
      url: 'http://127.0.0.1:8545',
      initialDate: '1970-01-01T00:00:00Z',
      gasMultiplier: 1.2,
      timeout: 1000000000000000,
    },
  },
  solidity: {
    version: '0.8.24',
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
      evmVersion: 'paris',
    },
  },
  gobind: {
    outdir: './bindings/go',
    onlyFiles: ['./contracts'],
    skipFiles: [
      'contracts/AppStorage.sol',
      'contracts/libraries',
      'contracts/diamond/libraries',
      'contracts/diamond/interfaces',
    ],
  },
  typechain: {
    outDir: `generated-types/${typechainTarget().split('-')[0]}`,
    target: typechainTarget(),
    alwaysGenerateOverloads: true,
    discriminateTypes: true,
    dontOverrideCompile: forceTypechain(),
  },
};

export default config;
