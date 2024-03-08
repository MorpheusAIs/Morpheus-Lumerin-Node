'use strict';

const { merge, union } = require('lodash');
const debug = require('debug')('lmr-wallet:core');
const EventEmitter = require('events');

const pluginCreators = [
  require('./plugins/rates'),
  require('./plugins/eth'),
  require('./plugins/explorer'),
  require('./plugins/wallet'),
  require('./plugins/token'),
  require('./plugins/lumerin'),
  require('./plugins/proxy-router'),
  require('./plugins/contracts'),
  require('./plugins/devices'),
];

function createCore () {
  let eventBus;
  let initialized = false;
  let plugins;

  function start (givenConfig) {
    if (initialized) {
      throw new Error('Wallet Core already initialized');
    }

    const config = merge({}, givenConfig);

    eventBus = new EventEmitter();

    debug.enabled = config.debug;
    if (config.debug) {
      const emit = eventBus.emit.bind(eventBus);
      eventBus.emit = function (eventName, ...args) {
        debug('<<--', eventName, ...args);
        emit(eventName, ...args);
      }
    }

    debug('Wallet core starting', config);

    let coreEvents = [];
    const pluginsApi = {};

    plugins = pluginCreators.map(create => create());

    plugins.forEach(function (plugin) {
      const params = { config, eventBus, plugins: pluginsApi };
      const { api, events, name } = plugin.start(params);

      if (api && name) {
        pluginsApi[name] = api;
      }

      if (events) {
        coreEvents = union(coreEvents, events);
      }
    });

    debug('Exposed events', coreEvents);

    initialized = true;

    return {
      api: pluginsApi,
      emitter: eventBus,
      events: coreEvents
    };
  }

  function stop () {
    if (!initialized) {
      throw new Error('Wallet Core not initialized');
    }

    plugins.reverse().forEach(function (plugin) {
      plugin.stop();
    });

    plugins = null;

    eventBus.removeAllListeners();
    eventBus = null;

    initialized = false;

    debug('Wallet core stopped');
  }

  return {
    start,
    stop
  };
}

module.exports = createCore;
