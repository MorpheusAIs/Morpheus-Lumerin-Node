/* eslint-disable max-params */
import { IsPasswordStrong } from '../lib/PasswordStrength';
import { isWeiable, isHexable, sanitize, sanitizeMnemonic } from './utils';

function validateAmount(client, amount, propName, max, errors = {}) {
  if (!amount) {
    errors[propName] = 'Amount is required';
  } else if (!isWeiable(client, amount)) {
    errors[propName] = 'Invalid amount';
  } else if (max && parseFloat(amount) > parseFloat(max)) {
    errors[propName] = 'Insufficient funds';
  } else if (parseFloat(amount) < 0) {
    errors[propName] = 'Amount must be greater than 0';
  }

  return errors;
}

export function validateCoinAmount(client, coinAmount, max, errors = {}) {
  return validateAmount(client, coinAmount, 'coinAmount', max, errors);
}

export function validateLmrAmount(client, lmrAmount, max, errors = {}) {
  return validateAmount(client, lmrAmount, 'lmrAmount', max, errors);
}

export function validateToAddress(client, toAddress, errors = {}) {
  if (!toAddress) {
    errors.toAddress = 'Address is required';
  } else if (!client.isAddress(toAddress)) {
    errors.toAddress = 'Invalid address';
  }

  return errors;
}

export function validateGasLimit(client, gasLimit, min, errors = {}) {
  const value = parseFloat(sanitize(gasLimit), 10);

  if (gasLimit === null || gasLimit === '') {
    errors.gasLimit = 'Gas limit is required';
  } else if (Number.isNaN(value)) {
    errors.gasLimit = 'Invalid value';
  } else if (Math.floor(value) !== value) {
    errors.gasLimit = 'Gas limit must be an integer';
  } else if (value <= 0) {
    errors.gasLimit = 'Gas limit must be greater than 0';
  } else if (!isHexable(client, value)) {
    errors.gasLimit = 'Invalid value';
  }

  return errors;
}

export function validateGasPrice(client, gasPrice, errors = {}) {
  const value = parseFloat(sanitize(gasPrice), 10);

  if (gasPrice === null || gasPrice === '') {
    errors.gasPrice = 'Gas price is required';
  } else if (Number.isNaN(value)) {
    errors.gasPrice = 'Invalid value';
  } else if (value <= 0) {
    errors.gasPrice = 'Gas price must be greater than 0';
  } else if (!isWeiable(client, gasPrice, 'gwei')) {
    errors.gasPrice = 'Invalid value';
  } else if (!isHexable(client, client.toWei(gasPrice, 'gwei'))) {
    errors.gasPrice = 'Invalid value';
  }

  return errors;
}

export function validateMnemonic(
  client,
  mnemonic,
  propName = 'mnemonic',
  errors = {}
) {
  if (!mnemonic) {
    errors[propName] = 'The phrase is required';
  } else if (!client.isValidMnemonic(sanitizeMnemonic(mnemonic))) {
    errors[propName] = "These words don't look like a valid recovery phrase";
  }

  return errors;
}

export function validateMnemonicAgain(
  client,
  mnemonic,
  mnemonicAgain,
  propName = 'mnemonicAgain',
  errors = {}
) {
  if (!mnemonicAgain) {
    errors[propName] = 'The phrase is required';
  } else if (!client.isValidMnemonic(sanitizeMnemonic(mnemonicAgain))) {
    errors[propName] = "These words don't look like a valid recovery phrase";
  } else if (sanitizeMnemonic(mnemonicAgain) !== mnemonic) {
    errors[propName] =
      'The text provided does not match your recovery passphrase.';
  }

  return errors;
}

export function validatePassword(password, errors = {}) {
  if (!password) {
    errors.password = 'Password is required';
  }

  return errors;
}

export function validatePasswordCreation(
  client,
  config,
  password,
  errors = {}
) {
  if (!password) {
    errors.password = 'Password is required';
  }
  // else if (!IsPasswordStrong(password)) {
  //   errors.password = 'Password is not strong enough';
  // }

  return errors;
}

export function validateUseMinimum(useMinimum, estimate, errors = {}) {
  if (useMinimum && !estimate) {
    errors.useMinimum = 'No estimated return. Try again.';
  }

  return errors;
}

export const validatePoolAddress = (address, errors = {}) => {
  const defaultPoolFormat = '{host}:{port}';
  const expectedFormat = `Expected format: ${defaultPoolFormat}`;

  if (!address) {
    errors.proxyDefaultPool = `Enter default destination. ${expectedFormat}`;
    return errors;
  }

  const pattern = /^[a-zA-Z0-9.-]+:\d+$/;
  const result = pattern.test(address);

  if (!result) {
    errors.proxyDefaultPool = `Invalid destination address. ${expectedFormat}`;
  }
  return errors;
};

export const validatePoolUsername = (username, errors = {}) => {
  if (!username || !username.trim()) {
    errors.proxyPoolUsername = 'Enter username';
    return errors;
  }
  return errors;
};
