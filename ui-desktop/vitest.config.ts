import { resolve } from 'path'
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

// Standalone vitest config.
//
// Deliberately NOT derived from electron.vite.config.ts: that config validates
// the full runtime env schema at load time and would force every contributor
// (and CI) to provide a complete .env just to run unit tests. Tests should be
// runnable on a fresh clone with nothing but `npm ci`.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@renderer': resolve(__dirname, 'src/renderer/src'),
      src: resolve(__dirname, 'src')
    }
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{js,jsx,ts,tsx}'],
    // Electron main-process code imports `electron`, which has no meaning
    // outside a running Electron process. Anything needing it must mock it.
    coverage: {
      reporter: ['text', 'lcov'],
      include: ['src/renderer/src/store/**', 'src/renderer/src/client/**', 'src/main/orchestrator/**']
    }
  }
})
