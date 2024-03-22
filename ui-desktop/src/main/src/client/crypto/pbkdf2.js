'use strict';

const credential = require('credential-plus');
const pbkdf2 = require('credential-plus-pbkdf2');

credential.install(pbkdf2);

const config = {
  digest: 'sha512',
  func: 'pbkdf2',
  iterations: 19997,
  keylen: 128
};

const hash = password => credential.hash(password, config);

// eslint-disable-next-line no-shadow
const verify = (hash, password) => credential.verify(hash, password);

export default { hash, verify };
