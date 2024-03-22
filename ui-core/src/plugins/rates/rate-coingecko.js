//@ts-check
const axios = require('axios').default

/**
 * Returns ETH and LMR prices in USD from coingecko api
 * @returns {Promise<{ LMR: number, ETH: number, BTC: number}>}
 */
const getRateCoingecko = async () => {
  const baseUrl = 'https://api.coingecko.com/api'
  const res = await axios.get(`${baseUrl}/v3/simple/price`, {
    params: {
      ids: 'ethereum,lumerin,bitcoin',
      vs_currencies: 'usd',
    },
  })

  const LMR = res?.data?.lumerin?.usd
  const ETH = res?.data?.ethereum?.usd
  const BTC = res?.data?.bitcoin?.usd;

  if (!LMR || !ETH || !BTC) {
    throw new Error(`invalid price response from coingecko: ${res.data}`)
  }
  return { LMR, ETH, BTC }
}

module.exports = { getRateCoingecko }
