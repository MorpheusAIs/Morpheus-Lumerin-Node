import { resolve } from 'path'
import { defineConfig, externalizeDepsPlugin } from 'electron-vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig({
  main: {
    build: {
      rollupOptions: {
        output: {
          format: 'es'
        }
      }
    },
    plugins: [externalizeDepsPlugin()]
  },
  preload: {
    build: {
      rollupOptions: {
        output: {
          format: 'es'
        }
      }
    },
    plugins: [externalizeDepsPlugin()]
  },
  renderer: {
    assetsInclude: ['**/*.png', '**/*.svg', '**/*.md'],
    resolve: {
      alias: {
        '@renderer': resolve('src/renderer/src'),
        '@tabler/icons': resolve('node_modules/@tabler/icons-react/dist/cjs/tabler-icons-react.cjs')
      }
    },
    plugins: [react(), svgr(), nodePolyfills()]
  }
})
