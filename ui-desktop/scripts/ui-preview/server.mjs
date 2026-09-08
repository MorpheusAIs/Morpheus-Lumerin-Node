import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import { createServer } from 'vite'
import react from '@vitejs/plugin-react'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

const root = dirname(fileURLToPath(import.meta.url))
const desktop = resolve(root, '../..')
const server = await createServer({
  configFile: false,
  envFile: false,
  root,
  plugins: [nodePolyfills(), react()],
  resolve: {
    alias: {
      '@renderer': resolve(desktop, 'src/renderer/src'),
      src: resolve(desktop, 'src')
    }
  },
  server: {
    host: '127.0.0.1',
    port: 5188,
    strictPort: true,
    fs: { allow: [desktop] }
  }
})
await server.listen()
console.log('Synthetic UI preview: http://127.0.0.1:5188/')
console.log('No Electron bridge, wallet, proxy-router, or real project files are used.')
for (const signal of ['SIGINT', 'SIGTERM']) {
  process.once(signal, async () => {
    await server.close()
    process.exit(0)
  })
}
