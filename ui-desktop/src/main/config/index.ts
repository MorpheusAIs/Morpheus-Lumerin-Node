// When adding/removing variables check 'electron.vite.config.ts' to update the list of 'envsToInject'

const config = {
  chain: {
    bypassAuth: process.env.BYPASS_AUTH,
    chainId: process.env.CHAIN_ID,
    defaultSellerCurrency: process.env.DEFAULT_SELLER_CURRENCY || 'BTC',
    diamondAddress: process.env.DIAMOND_ADDRESS,
    displayName: process.env.CHAIN_NAME,
    explorerUrl: process.env.EXPLORER_URL,
    localProxyRouterUrl: `http://localhost:${process.env.SERVICE_PROXY_API_PORT}`,
    mainTokenAddress: process.env.TOKEN_ADDRESS,
    symbol: 'MOR',
    symbolEth: 'ETH'
  },
  dbAutocompactionInterval: 30000,
  debug: process.env.DEBUG && !process.env.IGNORE_DEBUG_LOGS,
  devTools: process.env.DEV_TOOLS,
  sentryDsn: process.env.SENTRY_DSN,
  statePersistanceDebounce: 2000,
  trackingId: process.env.TRACKING_ID,
  isFailoverEnabled: process.env.FAILOVER_ENABLED
}
console.log('🚀 ~ config:', config)

export default config
