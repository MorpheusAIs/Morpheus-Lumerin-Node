import * as bip39 from 'bip39';

const createMnemonic = () => Promise.resolve(bip39.generateMnemonic());

const isValidMnemonic = mnemonic => bip39.validateMnemonic(mnemonic);

const mnemonicToSeedHex = mnemonic =>
  bip39.mnemonicToSeed(mnemonic).toString('hex');

const mnemonicToEntropy = mnemonic =>
  bip39.mnemonicToEntropy(mnemonic).toString('hex');

const entropyToMnemonic = entropy => {
  const buffer = Buffer.from(entropy, 'hex');
  return bip39.entropyToMnemonic(buffer);
};

const keys = {
  createMnemonic,
  isValidMnemonic,
  mnemonicToSeedHex,
  mnemonicToEntropy,
  entropyToMnemonic
};

export default keys;
