'use strict';

const { setSessionPassword, getSessionPassword } = require('./secure.js');
const logger = require('../../logger.js');
const {
  getPasswordHash,
  setPasswordHash
} = require('./settings');
const {
  pbkdf2: { hash, verify },
  sha256
} = require('./crypto');

async function setPassword(password) {
  await setSessionPassword(password);

  const passwordHash = getPasswordHash();
  if (!passwordHash) {
    logger.info('No password set, using current as default');
  }
  await hash(password).then(setPasswordHash);
}

function isValidPassword(password) {
  const passwordHash = getPasswordHash();

  return verify(passwordHash, password)
    .then(async function(isValid) {
      if (isValid) {
        await setSessionPassword(password);
        logger.verbose('Supplied password is valid');
      } else {
        logger.warn('Supplied password is invalid');
      }
      return isValid;
    })
    .catch(function(err) {
      logger.warn('Could not verify password', err);


      return false;
    });
}

async function getHashedPassword() {
  const hashString = await getPasswordHash();

  const hash = JSON.parse(hashString, (key, val) =>
    key == 'hash' ? JSON.parse(val) : val
  );

  return hash.hash.secret;
}

module.exports = {
  isValidPassword,
  setPassword,
  setSessionPassword,
  getSessionPassword,
  getHashedPassword
};
