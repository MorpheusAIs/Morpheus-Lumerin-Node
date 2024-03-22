'use strict'

const estimateGas = (web3) => {
  return ({ from, to, value }) =>
    web3.eth.estimateGas({ from, to, value }).then((gasLimit) => ({ gasLimit }))
}

const getGasPrice = (web3) => () => {
  return web3.eth.getGasPrice().then((gasPrice) => ({ gasPrice }))
}

function addAccount(web3, privateKey) {
  web3.eth.accounts.wallet
    .create(0)
    .add(web3.eth.accounts.privateKeyToAccount(privateKey))
}

const ensureAccount = function (web3) {
  return (privateKey) => {
    addAccount(web3, privateKey)
  }
}

const getNextNonce = (web3, from) => web3.eth.getTransactionCount(from, 'pending')

const sendSignedTransaction = (web3, logTransaction) =>
  function (privateKey, { from, to, value, gas, gasPrice }) {
    addAccount(web3, privateKey)
    const units = Math.floor(Number(value * 10 ** 18)).toString()
    return getNextNonce(web3, from)
      .then((nonce) =>
        logTransaction(
          web3.eth.sendTransaction({ from, to, value: units, gas, nonce }),
          from
        )
      )
  }

module.exports = {
  estimateGas,
  getGasPrice,
  sendSignedTransaction,
  ensureAccount,
}
