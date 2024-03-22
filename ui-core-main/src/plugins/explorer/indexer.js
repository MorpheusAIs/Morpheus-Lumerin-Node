'use strict';

const { CookieJar } = require('tough-cookie');
const { create: createAxios } = require('axios');
const { default: axiosCookieJarSupport } = require('axios-cookiejar-support');
const { isArrayLike } = require('lodash');
const blockscout = require('./blockscout');
const debug = require('debug')('lmr-wallet:core:explorer:indexer');
const EventEmitter = require('events');
const io = require('socket.io-client');
const pRetry = require('p-retry');

/**
 * Create an object to interact with the Lumerin indexer.
 *
 * @param {object} config The configuration object.
 * @param {object} eventBus The corss-plugin event bus.
 * @returns {object} The exposed indexer API.
 */
function createIndexer (config, eventBus) {
  const { chainId, debug: enableDebug, indexerUrl, useNativeCookieJar, wsIndexerUrl } = config;

  debug.enabled = enableDebug;

  let axios;
  let jar;
  let socket;

  if (useNativeCookieJar) {
    axios = createAxios({
      baseURL: indexerUrl,
      params: {
        apikey: '4VPHZ7SNPRRWKE23RBMX1MFUHZYDCAM9A4',
      }
    });
  } else {
    jar = new CookieJar();
    axios = axiosCookieJarSupport(createAxios(({
      baseURL: indexerUrl,
      withCredentials: true
    })));
    axios.defaults.jar = jar;
  }

  const getBestBlock = () =>
    axios('/blocks/best')
      .then(res => res.data)
      .then(best =>
        best && best.number && best.hash
          ? best
          : new Error('Indexer\' response is invalid for best block')
      );

  const getTransactions = (from, to, address) =>
    chainId === 61 // Ethereum Classic Mainnet chain ID
      ? blockscout.getTransactions(address, from, to)
      : axios(`/addresses/${address}/transactions`, { params: { from, to } })
        .then(res => res.data)
        .then(transactions =>
          isArrayLike(transactions)
            ? transactions
            : new Error(`Indexer response is invalid for ${address}`)
        );

  const getCookiePromise = useNativeCookieJar
    ? Promise.resolve()
    : pRetry(
      () =>
        getBestBlock()
          .then(function () {
            debug('Got indexer cookie')
          }),
      {
        forever: true,
        maxTimeout: 5000,
        onFailedAttempt (err) {
          debug('Failed to get indexer cookie', err.message)
        }
      }
    );

  const getSocket = () =>
    io(wsIndexerUrl || indexerUrl, {
      autoConnect: false,
      extraHeaders: jar
        ? { Cookie: jar.getCookiesSync(wsIndexerUrl || indexerUrl).join(';') }
        : {}
    });

  /**
   * Create a stream that will emit an event each time a transaction for the
   * specified address is indexed.
   *
   * The stream will emit `data` for each transaction. If the connection is lost
   * or an error occurs, an `error` event will be emitted. In addition, when the
   * connection is restablished, a `resync` will be emitted.
   *
   * @param {string} address The address.
   * @returns {object} The event emitter.
   */
  function getTransactionStream (address) {
    const stream = new EventEmitter();

    getCookiePromise
      .then(function () {
        socket = getSocket();

        socket.on('connect', function () {
          debug('Indexer connected');
          eventBus.emit('indexer-connection-status-changed', {
            connected: true
          });
          // TODO: Find out why this 'subscribe' event emitter is even here
          socket.emit('subscribe', { type: 'txs', addresses: [address] },
            function (err) {
              if (err) {
                stream.emit('error', err)
              }
            }
          )
        });

        socket.on('tx', function (data) {
          if (!data) {
            stream.emit('error', new Error('Indexer sent no tx event data'));
            return;
          }

          const { type, txid } = data;

          if (typeof txid !== 'string' || txid.length !== 66) {
            stream.emit('error', new Error('Indexer sent bad tx event data'));
            return;
          }

          stream.emit('data', txid);
        });

        socket.on('disconnect', function (reason) {
          debug('Indexer disconnected');
          eventBus.emit('indexer-connection-status-changed', {
            connected: false
          });
          stream.emit('error', new Error(`Indexer disconnected with ${reason}`));
        })

        socket.on('reconnect', function () {
          stream.emit('resync');
        });

        socket.on('error', function (err) {
          stream.emit('error', err);
        });

        socket.open();
      })
      .catch(function (err) {
        stream.emit('error', err);
      });

    return stream;
  }

  /**
   * Disconnects from the indexer.
   */
  function disconnect () {
    if (socket) {
      socket.close();
    }
  }

  return {
    disconnect,
    getBestBlock,
    getTransactions,
    getTransactionStream
  };
}

module.exports = createIndexer;
