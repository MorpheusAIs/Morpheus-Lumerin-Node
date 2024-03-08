import { resolve } from 'path'
import { defineConfig, externalizeDepsPlugin } from 'electron-vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'

export default defineConfig({
  main: {
    plugins: [externalizeDepsPlugin()]
  },
  preload: {
    plugins: [externalizeDepsPlugin()]
  },
  renderer: {
    assetsInclude: ['**/*.png', '**/*.svg', '**/*.md'],
    resolve: {
      alias: {
        '@renderer': resolve('src/renderer/src'),
        '@tabler/icons': resolve('node_modules/@tabler/icons/dist/es/tabler-icons.js')
      }
    },
    plugins: [react(), svgr()]
  }
})
