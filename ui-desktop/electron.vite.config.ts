import { resolve } from 'path'
import { defineConfig, externalizeDepsPlugin, loadEnv } from 'electron-vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

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
      define: {
        'process.env': JSON.stringify(env)
      }
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
      define: {
        'process.env': JSON.stringify(env)
      }
    },
    renderer: {
      assetsInclude: ['**/*.png', '**/*.svg', '**/*.md'],
      resolve: {
        alias: {
          '@renderer': resolve('src/renderer/src')
        }
      },
      plugins: [react(), svgr(), nodePolyfills()],
      define: {
        'process.env': JSON.stringify(env)
      }
    }
  }
})
