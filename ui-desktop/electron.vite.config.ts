import { resolve } from 'path'
import { defineConfig, externalizeDepsPlugin, loadEnv } from 'electron-vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

const envsToInject = [
  'AUTO_ADJUST_CONTRACT_PRICE_TIMEOUT',
  'AUTO_ADJUST_PRICE_INTERVAL',
  'BYPASS_AUTH',
  'CHAIN_ID',
  'CLONE_FACTORY_ADDRESS',
  'COIN_DEFAULT_GAS_LIMIT',
  'DEFAULT_GAS_PRICE',
  'DEFAULT_SELLER_CURRENCY',
  'DISPLAY_NAME',
  'ETH_NODE_ADDRESS',
  'ETH_NODE_ADDRESS_HTTP',
  'EXPLORER_API_URLS',
  'EXPLORER_URL',
  'FAUCET_URL',
  'IP_LOOKUP_URL',
  'LMR_DEFAULT_GAS_LIMIT',
  'LUMERIN_TOKEN_ADDRESS',
  'MAX_GAS_PRICE',
  'RECAPTCHA_SITE_KEY',
  'SENTRY_DSN',
  'SHOW_FAUCET',
  'SYMBOL_ETH',
  'SYMBOL_LMR',
  'TITAN_LIGHTNING_POOL',
  'TRACKING_ID'
] as const

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  // TODO: migrate to import.meta.env
  // Temporary hack to support process.env way to get env variables

  // Simply injecting our env variables for each occurence of process.env
  // doesn't work because it overwrites existing process.env variables
  // and screws the electron dev mode
  //
  // define: {
  //   'process.env': JSON.stringify(env) // don't do it
  // }
  //
  // so we need to define them in a way that they are merged with the existing process.env

  const processEnvDefineMap: Record<string, string> = {}

  for (const key of envsToInject) {
    processEnvDefineMap[`process.env.${key}`] = JSON.stringify(env[key])
  }

  return {
    main: {
      build: {
        rollupOptions: {
          output: {
            format: 'es'
          }
        }
      },
      plugins: [externalizeDepsPlugin()],
      define: processEnvDefineMap
    },
    preload: {
      build: {
        rollupOptions: {
          output: {
            format: 'es'
          }
        }
      },
      plugins: [externalizeDepsPlugin()],
      define: processEnvDefineMap
    },
    renderer: {
      assetsInclude: ['**/*.png', '**/*.svg', '**/*.md'],
      resolve: {
        alias: {
          '@renderer': resolve('src/renderer/src')
        }
      },
      plugins: [react(), svgr(), nodePolyfills()],
      define: processEnvDefineMap
    }
  }
})
