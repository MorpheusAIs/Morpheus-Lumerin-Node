'use strict';

const checkChain = (web3, chainId) =>
  web3.eth.getChainId()
    .then(id => {
      return id === +chainId || Promise.reject(new Error('Wrong chain'))
    });

module.exports = checkChain;
