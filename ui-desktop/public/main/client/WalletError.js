'use strict';

function WalletError (message, data) {
  this.name = 'WalletError';
  this.message = message || 'Unknown error';

  if (data) {
    this.data = data;
  }
  if (data && data instanceof Error) {
    this.stack = data.stack;
    this.inner = data.message;
  }
}

module.exports = WalletError;
