'use strict'

const { debounce, groupBy, merge, noop, reduce } = require('lodash')
const logger = require('../../logger');
const getTransactionStatus = require('./transaction-status')
const promiseAllProps = require('promise-all-props')
const abiDecoder = require('abi-decoder')
const { Lumerin, CloneFactory } = require('contracts-js')

function createQueue(config, eventBus, web3) {
  // debug.enabled = config.debug

  const lumerin = Lumerin(web3, config.lmrTokenAddress)
  const cloneFactory = CloneFactory(web3, config.cloneFactoryAddress)
  abiDecoder.addABI(lumerin.options.jsonInterface)
  abiDecoder.addABI(cloneFactory.options.jsonInterface)

  const metasCache = {}

  let pendingEvents = []

  function mergeEvents(hash, events) {
    const metas = events.map(({ event, metaParser }) => {
      return metaParser(event)}
    )

    metas.unshift(metasCache[hash] || {})

    metasCache[hash] = reduce(metas, merge)

    return metasCache[hash]
  }

  const mergeDones = (events) => events.map((event) => event.done || noop)

  function fillInStatus({ transaction, receipt, meta }) {
    if (receipt && meta) {
      meta.contractCallFailed = !getTransactionStatus(transaction, receipt)
    }
    return { transaction, receipt, meta }
  }

  const decodeInput = (address) =>
    ({ transaction, receipt, meta }) => {
      try {

        // for lmr token transactions retrieved by rpc call
        if (receipt.events){
          if (receipt.events.Transfer){
            const {from, to, value} = receipt.events.Transfer.returnValues
            transaction.input = {
              to,
              from,
              amount: value,
            }
            return { transaction, receipt, meta }
          }
        }

        // for lmr token transactions retrieved by etherscan api
        if (receipt.tokenSymbol === 'LMR'){
          transaction.input = {
            to: receipt.to,
            from: receipt.from,
            amount: receipt.value,
          }
          return { transaction, receipt, meta }
        }

        if (
          typeof transaction.input === 'string' &&
          transaction.input !== '0x'
        ) {
          if (!receipt.logs){
            return { transaction, receipt, meta }
          }


          const logs = abiDecoder.decodeLogs(receipt.logs)
          if (!logs) {
            return null
          }

          const transfer = logs.find(
            (l) =>
              l.name === 'Transfer' &&
              l.events.find(
                (e) =>
                  (e.name === 'to' && e.value === address) ||
                  (e.name === 'from' && e.value === address)
              )
          )
          if (!transfer) {
            return null
          }
          const { events } = transfer

          const valueParam = events.find((p) => p.name === 'value')
          if (valueParam === undefined) {
            return null
          }

          const toParam = events.find((p) => p.name === 'to')
          if (toParam === undefined) {
            return null
          }

          const fromParam = events.find((p) => p.name === 'from')
          if (fromParam === undefined) {
            return null
          }

          transaction.input = {
            to: toParam.value,
            amount: valueParam.value,
            from: fromParam.value,
          }
          return { transaction, receipt, meta }
        }
        return { transaction, receipt, meta }
      } catch (err) {
        return null
      }
    }

  function emitTransactions(address, transactions) {
    eventBus.emit('token-transactions-changed', {
      transactions: transactions
        .filter((data) => !!data.transaction)
        .map(fillInStatus)
        .map(decodeInput(address.toLowerCase()))
        .filter((i) => Number(i.transaction.value) !== 0) // filters out eth transactions that correspond to token transfers
        .filter((i) => !!i)
    })

    eventBus.emit('eth-tx');
    eventBus.emit('lmr-tx');
  }

  function tryEmitTransactions(address, transactions) {
    try {
      emitTransactions(address, transactions)
      return null
    } catch (err) {
      return err
    }
  }

  //TODO: accept transaction and reciept to avoid api calls
  //TODO: if transaction/reciept are not available, use block number to get all transactions and receipts in one api call
  function emitPendingEvents(address) {
    logger.debug('About to emit pending events')

    const eventsToEmit = pendingEvents.filter((e) => e.address === address)
    const eventsToKeep = pendingEvents.filter((e) => e.address !== address)
    pendingEvents = eventsToKeep

    const grouped = groupBy(eventsToEmit, 'event.transactionHash')

    Promise.all(
      Object.keys(grouped).map((hash) =>
        promiseAllProps({
          transaction: web3.eth.getTransaction(hash),
          receipt: web3.eth.getTransactionReceipt(hash),
          meta: mergeEvents(hash, grouped[hash]),
          done: mergeDones(grouped[hash]),
        })
      )
    )
      .then(function (transactions) {
        const err = tryEmitTransactions(address, transactions)
        return Promise.all(
          transactions.map((transaction) =>
            Promise.all(transaction.done.map((done) => done(err)))
          )
        )
      })
      .catch(function (err) {        
        eventBus.emit('wallet-error', {
          inner: err,
          message: 'Could not emit event transaction',
          meta: { plugin: 'explorer' },
        })
        eventsToEmit.forEach(function (event) {
          event.done(err)
        })
      })
  }

  /**
   * 
   * @param {string} address 
   * @param {any[]} eventsData 
   * @returns {void}
   */
  function emitPendingEventsV2(address, eventsData) {
    const transactionItems = eventsData.map((eventData) => ({
      transaction: eventData.event.transaction,
      receipt: eventData.event.receipt,
      meta: mergeEvents(eventData.event.receipt.transactionHash, [eventData]),
      done: eventData.done
    }))

    const err = tryEmitTransactions(address, transactionItems)

    transactionItems.forEach(async function (transaction) {
      transaction.done(err)
    })

    if (err){
      eventBus.emit('wallet-error', {
        inner: err,
        message: 'Could not emit event transaction',
        meta: { plugin: 'explorer' },
      })
    }
    
    return
  }

  const debouncedEmitPendingEvents = debounce(
    emitPendingEvents,
    config.explorerDebounce
  )

  const addTransaction = (address, meta) =>
    function (hash) {
      return new Promise(function (resolve, reject) {
        const event = {
          address,
          event: { transactionHash: hash },
          metaParser: () => meta || {},
          done: (err) => (err ? reject(err) : resolve()),
        }
        pendingEvents.push(event)

        debouncedEmitPendingEvents(address)
      })
    }

  const addTx = (address, metaParser) => (txAndReceipt) => {
    return new Promise(function (resolve, reject) {
      const event = {
        address,
        event: txAndReceipt,
        metaParser: () => metaParser || {},
        done: (err) => (err ? reject(err) : resolve()),
      }

      emitPendingEventsV2(address, [event])
    })
  }

  const addTxs = (address, metaParser) => (txAndReceipts) => {
    return new Promise(function (resolve, reject) {
      const events = txAndReceipts.map((txAndReceipt) => ({
        address,
        event: txAndReceipt,
        metaParser: () => metaParser || {},
        done: (err) => (err ? reject(err) : resolve()),
      }))

      emitPendingEventsV2(address, events)
    })
  }

  const addEvent = (address, metaParser) =>
    function (event) {
      logger.debug('Queueing event', event.event)
      return new Promise(function (resolve, reject) {
        pendingEvents.push({
          address,
          event,
          metaParser,
          done: (err) => (err ? reject(err) : resolve()),
        })
        debouncedEmitPendingEvents(address)
      })
    }

  return {
    addEvent,
    addTransaction,
    addTx,
    addTxs,
  }
}

module.exports = createQueue
