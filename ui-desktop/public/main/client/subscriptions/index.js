'use strict';

const { subscribeSingleCore, unsubscribeSingleCore } = require('./single-core');
const { subscribeWithoutCore, unsubscribeWithoutCore } = require('./no-core');

function subscribe (core) {
  subscribeSingleCore(core);
  subscribeWithoutCore();
}

function unsubscribe (core) {
  unsubscribeSingleCore(core);
  unsubscribeWithoutCore();
}

module.exports = { subscribe, unsubscribe };
