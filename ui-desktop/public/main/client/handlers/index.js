'use strict';

const singleCore = require('./single-core');
const noCore = require('./no-core');

module.exports = Object.assign({}, singleCore, noCore);
