import { resolve } from 'path'
import { existsSync } from 'fs'
import { defineConfig, externalizeDepsPlugin, loadEnv } from 'electron-vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import { nodePolyfills } from 'vite-plugin-node-polyfills'
import type { Plugin } from 'vite'
import { EnvSchema } from './env.schema'
import { newAjv } from './validator'

const envsToInject = Object.keys(EnvSchema.properties)

function httpsOrigin(value: string | undefined): string | undefined {
  if (!value) return undefined
  try {
    const url = new URL(value)
    return url.protocol === 'https:' ? url.origin : undefined
  } catch {
    return undefined
  }
}

function contentSecurityPolicyPlugin(isDevelopment: boolean, sentryDsn?: string): Plugin {
  const scriptSource = isDevelopment
    ? "script-src 'self' 'unsafe-inline' 'unsafe-eval'"
    : "script-src 'self'"
  const sentry = httpsOrigin(sentryDsn)
  const connectSource = [
    "connect-src 'self'",
    ...(isDevelopment
      ? ['http://127.0.0.1:*', 'http://localhost:*', 'ws://127.0.0.1:*', 'ws://localhost:*']
      : []),
    ...(sentry ? [sentry] : [])
  ].join(' ')
  const policy = [
    "default-src 'self'",
    "base-uri 'none'",
    "object-src 'none'",
    scriptSource,
    "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
    "font-src 'self' data: https://fonts.gstatic.com",
    "img-src 'self' data: blob:",
    "media-src 'self' data: blob:",
    connectSource,
    "worker-src 'self' blob:",
    "frame-src 'none'",
    "form-action 'self'"
  ].join('; ')

  return {
    name: 'morpheus-content-security-policy',
    transformIndexHtml() {
      return [
        {
          tag: 'meta',
          attrs: {
            'http-equiv': 'Content-Security-Policy',
            content: policy
          },
          injectTo: 'head-prepend'
        }
      ]
    }
  }
}

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const ajv = newAjv()
  const validate = ajv.compile(EnvSchema)

  if (!validate(env)) {
    // A missing .env is by far the most common cause here, and the raw ajv
    // output ("ENV must have required property 'NODE_ENV'") does not make that
    // obvious to someone running the app for the first time. Say it plainly.
    const envFileExists = existsSync(resolve(__dirname, '.env'))
    const details = ajv.errorsText(validate.errors, { dataVar: 'ENV', separator: '\n  - ' })

    throw new Error(
      [
        '',
        'Invalid environment configuration:',
        `  - ${details}`,
        '',
        envFileExists
          ? 'ui-desktop/.env exists but is incomplete. Compare it against .env.example.'
          : 'ui-desktop/.env does not exist. Create it with:\n\n    cp .env.example .env\n',
        'Note: an empty value is not the same as an absent one. SENTRY_DSN= fails',
        'the uri format check and FAILOVER_ENABLED= fails the boolean check —',
        'comment those lines out instead of blanking them.',
        ''
      ].join('\n')
    )
  }

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
          // CJS for the preload script. Electron 28's ESM preload support is
          // patchy (notably, the bundled sandbox bootstrap fails with
          // "object null is not iterable"). CJS sidesteps the whole CJS↔ESM
          // interop issue around `electron` named imports and just works.
          output: {
            format: 'cjs',
            entryFileNames: '[name].js'
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
      plugins: [
        contentSecurityPolicyPlugin(command === 'serve', env.SENTRY_DSN),
        react({ babel: { plugins: ['styled-components'], babelrc: false, configFile: false } }),
        svgr(),
        nodePolyfills()
      ],
      define: processEnvDefineMap
    }
  }
})
