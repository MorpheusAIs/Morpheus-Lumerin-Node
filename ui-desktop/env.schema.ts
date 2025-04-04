import { NumberOptions, Type, type Static } from '@sinclair/typebox'

// Eth address field validator
const TypeEthAddress = Type.String({ pattern: '^0x[a-fA-F0-9]{40}$' })
const TypePort = (opt?: NumberOptions) => Type.Number({ minimum: 1, maximum: 65535, ...opt })

// Environment variables schema
export const EnvSchema = Type.Object({
  BLOCKSCOUT_API_URL: Type.String({ format: 'uri' }),
  BYPASS_AUTH: Type.Boolean({ default: false }),
  CHAIN_ID: Type.Number(),
  CHAIN_NAME: Type.String(),
  DEBUG: Type.Boolean({ default: false }),
  DEFAULT_SELLER_CURRENCY: Type.String(),
  DEV_TOOLS: Type.Boolean({ default: false }),
  DIAMOND_ADDRESS: TypeEthAddress,
  EXPLORER_URL: Type.String(),
  FAILOVER_ENABLED: Type.Boolean({ default: true }),
  NODE_ENV: Type.Union([Type.Literal('development'), Type.Literal('production')]),
  LOG_LEVEL: Type.Union(
    [
      Type.Literal('error'),
      Type.Literal('warn'),
      Type.Literal('info'),
      Type.Literal('verbose'),
      Type.Literal('debug'),
      Type.Literal('silly')
    ],
    { default: 'info' }
  ),
  SENTRY_DSN: Type.Optional(Type.String({ format: 'uri' })),
  TOKEN_ADDRESS: TypeEthAddress,
  TRACKING_ID: Type.String({ default: '' }),
  SERVICE_PROXY_DOWNLOAD_URL_MAC_ARM64: Type.String({ default: '' }),
  SERVICE_PROXY_DOWNLOAD_URL_MAC_X64: Type.String({ default: '' }),
  SERVICE_PROXY_DOWNLOAD_URL_LINUX_X64: Type.String({ default: '' }),
  SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_X64: Type.String({ default: '' }),
  SERVICE_PROXY_API_PORT: TypePort({ default: 8082 }),
  SERVICE_PROXY_PORT: TypePort({ default: 3333 }),
  SERVICE_IPFS_API_PORT: TypePort({ default: 5001 }),
  SERVICE_AI_API_PORT: TypePort({ default: 3434 })
})

// Inferred type of environment variables
export type Env = Static<typeof EnvSchema>
