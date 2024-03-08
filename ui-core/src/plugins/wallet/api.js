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

const sendSignedTransaction = (web3, logTransaction) =>
  function (privateKey, { from, to, value, gas, gasPrice }) {
    
    addAccount(web3, privateKey)
    return web3.eth
      .getTransactionCount(from, 'pending')
      .then((nonce) =>
        logTransaction(
          web3.eth.sendTransaction({ from, to, value, gas, gasPrice, nonce }),
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
