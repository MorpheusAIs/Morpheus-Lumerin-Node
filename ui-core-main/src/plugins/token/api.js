'use strict'

const {
  utils: { isAddress, toChecksumAddress },
} = require('web3')
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

    events.getEventDataCreators(checksumAddress).forEach(explorer.registerEvent)

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

module.exports = {
  registerToken,
  getTokenBalance,
  getTokenGasLimit,
}
