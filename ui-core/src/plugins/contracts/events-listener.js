//@ts-check
// const debug = require('debug')('lmr-wallet:core:contracts:event-listener')
const logger = require('../../logger');

class ContractEventsListener {
  /**
   * @param {import('contracts-js').CloneFactoryContext} cloneFactory
   */
  constructor(cloneFactory) {
    this.cloneFactory = cloneFactory
    this.cloneFactoryListener = null
    this.contracts = {}
    this.walletAddress = null;
  }

  /**
   * @param {(contractId?: string, walletAddress?: string) => void} onUpdate
   */
  setOnUpdate(onUpdate) {
    this.onUpdate = onUpdate
  }

  /**
   *
   * @param {string} id
   * @param {import('contracts-js').ImplementationContext} instance
   * @param {string} walletAddress
   */
  addContract(id, instance, walletAddress) {
    if (!this.contracts[id]) {
      this.contracts[id] = instance.events.allEvents()
      this.contracts[id]
        .on('connected', () => {
          logger.debug(`Start listen contract (${id}) events`)
        })
        .on('data', async () => {
          logger.debug(`Contract (${id}) updated`)
          if (this.onUpdate){
            await new Promise((resolve) => setTimeout(resolve, 1000))
            this.onUpdate(id, this.walletAddress || walletAddress)
          }
        })
    }
  }

  listenCloneFactory() {
    if (!this.cloneFactoryListener) {
      this.cloneFactoryListener = this.cloneFactory.events.contractCreated()
      this.cloneFactoryListener
        .on('connected', () => {
          logger.debug('Start listen clone factory events')
        })
        .on('data', async (event) => {
          const contractId = event.returnValues._address
          logger.debug('New contract created', contractId)
          await new Promise((resolve) => setTimeout(resolve, 1000))
          this.onUpdate(contractId, this.walletAddress)
        })
    }
  }

  /**
   * @static
   * @param {import('contracts-js').CloneFactoryContext} cloneFactory
   * @param {boolean} [debugEnabled=false]
   * @returns {ContractEventsListener}
   */
  static create(cloneFactory, debugEnabled = false) {
    if (ContractEventsListener.instance) {
      return ContractEventsListener.instance
    }

    const instance = new ContractEventsListener(cloneFactory)
    ContractEventsListener.instance = instance
    instance.listenCloneFactory()
    return instance
  }

  /**
   * @returns {ContractEventsListener}
   */
  static getInstance() {
    if (!ContractEventsListener.instance) {
      throw new Error("ContractEventsListener instance not created")
    }
    return ContractEventsListener.instance
  }

  /**
   * @static
   * @param {(contractId?: string) => void} onUpdate
  */
  static setOnUpdate(onUpdate) {    
    ContractEventsListener.getInstance().onUpdate = onUpdate
  }
}

module.exports = { ContractEventsListener }
