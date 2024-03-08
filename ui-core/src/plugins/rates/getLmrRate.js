const axios = require('axios').default

const getLmrRate = () =>
  axios
    .get(`https://api.coingecko.com/api/v3/simple/price`, {
      params: {
        ids: 'lumerin',
        vs_currencies: 'usd',
      },
    })
    .then((response) =>
      response.data &&
      response.data.lumerin &&
      typeof response.data.lumerin.usd === 'number'
        ? Number.parseFloat(response.data.lumerin.usd)
        : null
    )

module.exports = { getLmrRate }
