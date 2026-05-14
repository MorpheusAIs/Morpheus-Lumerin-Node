import { resolve } from 'path'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const envsToInject = [
    'BLOCKSCOUT_API_URL',
    'BYPASS_AUTH',
    'CHAIN_ID',
    'CHAIN_NAME',
    'DEBUG',
    'DEFAULT_SELLER_CURRENCY',
    'DEV_TOOLS',
    'DIAMOND_ADDRESS',
    'EXPLORER_URL',
    'FAILOVER_ENABLED',
    'NODE_ENV',
    'LOG_LEVEL',
    'SENTRY_DSN',
    'TOKEN_ADDRESS',
    'TRACKING_ID',
    'SERVICE_PROXY_DOWNLOAD_URL_MAC_ARM64',
    'SERVICE_PROXY_DOWNLOAD_URL_MAC_X64',
    'SERVICE_PROXY_DOWNLOAD_URL_LINUX_X64',
    'SERVICE_PROXY_DOWNLOAD_URL_LINUX_ARM64',
    'SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_X64',
    'SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_ARM64',
    'SERVICE_PROXY_API_PORT',
    'SERVICE_PROXY_PORT',
    'SERVICE_IPFS_API_PORT',
    'SERVICE_AI_API_PORT',
  ]

  const processEnvDefineMap: Record<string, string> = {}
  for (const key of envsToInject) {
    const val = env[key]
    if (val !== undefined) {
      processEnvDefineMap[`process.env.${key}`] = JSON.stringify(val)
    }
  }

  return {
    root: 'src/renderer',
    assetsInclude: ['**/*.png', '**/*.svg', '**/*.md'],
    resolve: {
      alias: {
        '@renderer': resolve('src/renderer/src'),
        'src/main': resolve('src/main'),
      }
    },
    plugins: [
      react({ babel: { plugins: ['styled-components'], babelrc: false, configFile: false } }),
      svgr(),
      nodePolyfills(),
    ],
    define: {
      ...processEnvDefineMap,
      'process.env.NODE_ENV': JSON.stringify(env.NODE_ENV || 'development'),
    },
    server: {
      host: '0.0.0.0',
      port: 5000,
      strictPort: true,
      allowedHosts: 'all',
    },
  }
})
