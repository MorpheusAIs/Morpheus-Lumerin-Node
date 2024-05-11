'use strict'

const { Lumerin } = require('contracts-js')

const addAccount = (web3, privateKey) =>
  web3.eth.accounts.wallet
    .create(0)
    .add(web3.eth.accounts.privateKeyToAccount(privateKey))

/**
 * 
 * @param {*} web3 
 * @param {import("contracts-js").LumerinContext} lumerin 
 * @param {*} logTransaction 
 * @param {*} metaParsers 
 * @returns 
 */
const sendLmr = (web3, lumerin, logTransaction, metaParsers) => {
  return (privateKey, { gasPrice, gas, from, to, value }) => {
    addAccount(web3, privateKey)
    const lmrValue = parseFloat(value);
    const lmrUnits = Math.floor(Number(lmrValue * 10 ** 18)).toString();
console.log("mor value", lmrValue);
console.log("mor units", lmrUnits);
    return logTransaction(
        lumerin.methods.transfer(to, lmrUnits).send({ from, gas }),
        from,
        metaParsers.transfer({
          address: lumerin.options.address,
          returnValues: { _from: from, _to: to, _value: lmrUnits },
        })
      )
  }
}

const estimateGasTransfer = (lumerin) => {
  return ({ from, to, value }) => {
    return lumerin.methods.transfer(to, value).estimateGas({ from }).then((gasLimit) => ({ gasLimit }))
  }
}

// {
//   [1]   gasPrice: '1000000000',
//   [1]   gas: '999999',
//   [1]   from: '0x7525960Bb65713E0A0e226EF93A19a1440f1116d',
//   [1]   to: '0x146590438A9Ab7F186d9758629Af476b2B962A37',
//   [1]   value: '632911392405063'
//   [1] }
// Approves claimant contract to transfer LMR tokens on the Lumerin Contract's behalf
const increaseAllowance = (
  web3,
  chain,
  claimantAddress,
  lmrAmount,
  walletAddress,
  gasLimit = 1000000
) => {
  return Lumerin(web3, chain).methods
    .increaseAllowance(claimantAddress, lmrAmount)
    .send({ from: walletAddress, gas: gasLimit })
}

module.exports = {
  increaseAllowance,
  sendLmr,
  estimateGasTransfer,
}
