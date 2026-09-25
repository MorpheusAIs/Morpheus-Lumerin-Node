// Standalone UI-only Electron smoke test. Never bootstraps the real wallet app.
import { build } from 'esbuild'
import { mkdtemp, mkdir, copyFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRequire } from 'node:module'
import { spawn } from 'node:child_process'

const require = createRequire(import.meta.url)
const here = dirname(fileURLToPath(import.meta.url))
const desktop = resolve(here, '../..')
const temporary = await mkdtemp(join(tmpdir(), 'morpheus-session-ui-'))
await mkdir(join(temporary, 'preload'))
await build({
  entryPoints: [join(desktop, 'src/main/sessionConfirmation.ts')],
  outfile: join(temporary, 'main/sessionConfirmation.cjs'),
  bundle: true,
  platform: 'node',
  format: 'cjs',
  external: ['electron'],
  loader: { '.css': 'text' }
})
await copyFile(
  join(desktop, 'out/preload/session-confirmation.js'),
  join(temporary, 'preload/session-confirmation.js')
)
const environment = { ...process.env }
delete environment.ELECTRON_RUN_AS_NODE
const child = spawn(
  require('electron'),
  [join(here, 'session-confirmation-smoke.cjs'), temporary],
  { env: environment, stdio: 'inherit' }
)
console.log(`UI-only session confirmation test data: ${temporary}`)
child.on('exit', (code) => process.exit(code ?? 1))
for (const signal of ['SIGINT', 'SIGTERM']) {
  process.once(signal, () => child.kill(signal))
}
