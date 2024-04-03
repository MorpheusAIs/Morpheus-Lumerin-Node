'use strict';

const { hdkey } = require('ethereumjs-wallet');

const getWalletFromSeed = (seed, index = 0) =>
  hdkey.fromMasterSeed(Buffer.from(seed, 'hex'))
    .derivePath(`m/44'/60'/0'/0/${index}`)
    .getWallet();

const getAddress = (seed, index) =>
  getWalletFromSeed(seed, index)
    .getChecksumAddressString();

const getPrivateKey = (seed, index) =>
  getWalletFromSeed(seed, index)
    .getPrivateKey();

const getPrivateKeyStringPrefixed = (seed, index) =>
    getWalletFromSeed(seed, index)
      .getPrivateKeyString();

const getAddressAndPrivateKey = (seed, index) => ({
  address: getAddress(seed, index),
  privateKey: getPrivateKey(seed, index).toString('hex')
});

module.exports = {
  getAddress,
  getPrivateKey,
  getPrivateKeyStringPrefixed,
  getAddressAndPrivateKey
};
