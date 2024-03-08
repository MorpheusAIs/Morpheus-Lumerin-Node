'use strict';

const axios = require('axios').default;

/**
 *
 * @param {string} address The address.
 * @param {number} [startblock] The starting block.
 * @param {number} [endblock] The ending block.
 * @returns {Promise<string[]>} The list of transaction ids.
 */
function getTransactions (address, startblock, endblock) {
  return axios({
    // baseURL: 'https://api-ropsten.etherscan.io/api',
    // url: '/',
    // params: {
    //   apikey: '4VPHZ7SNPRRWKE23RBMX1MFUHZYDCAM9A4',
    //   module: 'account',
    //   action: 'txlist',
    //   address,
    //   sort: 'desc',
    //   startblock,
    //   endblock
    // }
    baseURL: 'https://blockscout.com/etc/mainnet/api',
    url: '/',
    params: {
      module: 'account',
      action: 'txlist',
      address,
      sort: 'desc',
      startblock,
      endblock
    }
  }).then(function ({ data }) {
    
    if (data.status !== '1' && data.message !== 'No transactions found') {
      return Promise.reject(new Error(data.message));
    }
    return data.result.map(t => t.hash);
  });
}

module.exports = { getTransactions };
