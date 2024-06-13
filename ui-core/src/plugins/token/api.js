'use strict'

const {
  utils: { isAddress, toChecksumAddress },
} = require('web3')
const axios = require('axios').default
const debug = require('debug')('lmr-wallet:core:debug')

const registerToken = ({ explorer }) =>
  function (contractAddress) {
    debug('Registering token', contractAddress)

    if (!isAddress(contractAddress)) {
      return false
    }

    const checksumAddress = toChecksumAddress(contractAddress)

    if (contractAddress === checksumAddress) {
      return false
    }

    return true
  }

function getTokenBalance(lumerin, walletAddress) {
  return lumerin.methods.balanceOf(walletAddress).call()
}

function getTokenGasLimit(lumerin) {
  return function ({ to, from, value }) {
    return lumerin.methods
      .transfer(to, value)
      .estimateGas({ from })
      .then((gasLimit) => ({ gasLimit }))
  }
}

// const claimFaucetDirectly = async (web3, privateKey, walletId) => {
//   const claimFunction = web3.utils.keccak256('claim()').substring(0, 10)
//   const gasLimit = 300_000
//   const txObject = {
//     from: walletId,
//     to: faucetAddress,
//     gas: gasLimit,
//     data: claimFunction,
//   }

//   const signedTx = await web3.eth.accounts.signTransaction(txObject, privateKey)
//   const result = await web3.eth.sendSignedTransaction(signedTx.rawTransaction)
//   return result
// }

const claimFaucetHttp = async (token, walletAddr, faucetUrl) => {
  const baseURL = `${faucetUrl}/${walletAddr}`
  const api = axios.create({ baseURL });
  const result = await api.post(
    '',
    {},
    {
      headers: {
        'x-captcha-token': token,
      },
    }
  );
  return result.data
}

function claimFaucet(web3, faucetAddress, faucetUrl) {
  return async (params) => {
    const { walletId, privateKey, token } = params
    return claimFaucetHttp(token, walletId, faucetUrl);
  }
}

module.exports = {
  registerToken,
  getTokenBalance,
  getTokenGasLimit,
  claimFaucet,
}
