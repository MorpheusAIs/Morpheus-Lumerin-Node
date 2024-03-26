'use strict';

const crypto = require('crypto');

const sha256 = require('./sha256');

function encrypt (key, text) {
  const iv = crypto.randomBytes(16).toString('hex'); // 128 bits

  const cipher = crypto.createCipheriv(
    'aes-256-cbc',
    Buffer.from(sha256.hash(key), 'hex'), // 256 bits
    Buffer.from(iv, 'hex')
  );

  return iv + cipher.update(text, 'utf-8', 'hex') + cipher.final('hex');
}

function decrypt (key, text) {
  const iv = text.substr(0, 32);
  const encrypted = text.substr(32);

  const decipher = crypto.createDecipheriv(
    'aes-256-cbc',
    Buffer.from(sha256.hash(key), 'hex'), // 256 bits
    Buffer.from(iv, 'hex')
  );

  return decipher.update(encrypted, 'hex', 'utf-8') + decipher.final('utf-8');
}

module.exports = { decrypt, encrypt };
