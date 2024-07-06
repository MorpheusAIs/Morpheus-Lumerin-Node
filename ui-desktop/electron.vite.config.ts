import { resolve } from 'path'
import { defineConfig, externalizeDepsPlugin, loadEnv } from 'electron-vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

const envsToInject = [
  'BYPASS_AUTH',
  'CHAIN_ID',
  'DEBUG',
  'DEFAULT_SELLER_CURRENCY',
  'DEV_TOOLS',
  'DIAMOND_ADDRESS',
  'DISPLAY_NAME',
  'EXPLORER_URL',
  'IGNORE_DEBUG_LOGS',
  'PROXY_WEB_DEFAULT_PORT',
  'SENTRY_DSN',
  'SYMBOL_ETH',
  'SYMBOL_LMR',
  'TOKEN_ADDRESS',
  'TRACKING_ID'
] as const

export default defineConfig(({ /*command,*/ mode }) => {
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
