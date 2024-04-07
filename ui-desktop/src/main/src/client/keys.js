'use strict';

const bip39 = require('bip39');

export const createMnemonic = () => Promise.resolve(bip39.generateMnemonic());

export const isValidMnemonic = mnemonic => bip39.validateMnemonic(mnemonic);

export const mnemonicToSeedHex = mnemonic =>
  bip39.mnemonicToSeedSync(mnemonic).toString('hex');

export const mnemonicToEntropy = mnemonic => bip39.mnemonicToEntropy(mnemonic).toString('hex');

export const entropyToMnemonic = entropy => {
  const buffer = Buffer.from(entropy, 'hex');
  return bip39.entropyToMnemonic(buffer);
}

export default {
  createMnemonic,
  isValidMnemonic,
  mnemonicToSeedHex,
  mnemonicToEntropy,
  entropyToMnemonic
};
