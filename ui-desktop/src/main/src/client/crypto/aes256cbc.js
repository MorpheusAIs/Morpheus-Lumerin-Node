'use strict';

const crypto = require('crypto');

function decrypt (password, data) {
  // eslint-disable-next-line node/no-deprecated-api
  const decipher = crypto.createDecipher('aes-256-cbc', password);

  let decrypted = decipher.update(data, 'hex', 'utf-8');
  decrypted += decipher.final('utf-8');
  return decrypted;
}

export default { decrypt };
