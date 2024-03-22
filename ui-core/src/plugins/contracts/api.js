// const debug = require('debug')('lmr-wallet:core:contracts:api')
const logger = require('../../logger')
const { encrypt } = require('ecies-geth')
const { Implementation } = require('contracts-js')
const { remove0xPrefix, add65BytesPrefix } = require('./helpers')
const { ContractEventsListener } = require('./events-listener')
const ethereumWallet = require('ethereumjs-wallet').default

/**
 * @param {import('web3').default} web3
 * @param {string} implementationAddress
 * @param {string} [walletAddress]
 */
async function _loadContractInstance(
  web3,
  implementationAddress,
  walletAddress
) {
  try {
    const implementationContract = Implementation(web3, implementationAddress)
    const contract = await implementationContract.methods
      .getPublicVariablesV2()
      .call()
    const stats = await implementationContract.methods.getStats().call()

    const history = await implementationContract.methods
      .getHistory('0', '100')
      .call()
    const buyerHistory = history
      .filter((h) => {
        return h[6] === walletAddress
      })
      .map((h) => ({
        ...h,
        id: implementationAddress,
      }))

    const { _successCount: successCount, _failCount: failCount } = stats

    const {
      _state: state,
      _terms: {
        _price: price, // cost to purchase the contract
        _limit: limit, // max th provided
        _speed: speed, // th/s of contract
        _length: length, // duration of the contract in seconds
        _version: version,
        _profitTarget: profitTarget
      }, 
      _startingBlockTimestamp: timestamp, // timestamp of the block at moment of purchase
      _buyer: buyer, // wallet address of the purchasing party
      _seller: seller, // wallet address of the selling party
      _encryptedPoolData: encryptedPoolData, // encrypted data for pool target info,
      _isDeleted: isDead, // check if contract is dead
      _balance: balance,
      _hasFutureTerms: hasFutureTerms,
    } = contract

    let futureTerms = null
    if (walletAddress && hasFutureTerms && seller === walletAddress) {
      const data = await implementationContract.methods.futureTerms().call()
      futureTerms = {
        price: data._price,
        speed: data._speed,
        length: data._length,
        limit: data._limit,
        version: data._version,
        profitTarget: data._profitTarget
      }
    }

    return {
      data: {
        id: implementationAddress,
        price,
        speed,
        length,
        buyer,
        seller,
        timestamp,
        state,
        encryptedPoolData,
        limit,
        isDead,
        balance,
        stats: {
          successCount,
          failCount,
        },
        hasFutureTerms,
        futureTerms,
        history: buyerHistory,
        version,
        profitTarget
      },
    }
  } catch (err) {
    logger.error(err)
    logger.error(
      'Error when trying to load Contracts by address in the Implementation contract: ',
      err
    )
    throw err
  }
}

/**
 * @param {import('web3').default} web3
 * @param {import('web3').default} web3Subscriptionable
 * @param {import('contracts-js').LumerinContext} lumerin
 * @param {import('contracts-js').CloneFactoryContext} cloneFactory
 * @param {string[]} addresses
 * @param {string} walletAddress
 */
async function getContracts(
  web3,
  web3Subscriptionable,
  lumerin,
  cloneFactory,
  addresses,
  walletAddress,
  eventBus
) {
  const chunkSize = 5
  const result = []
  for (let i = 0; i < addresses.length; i += chunkSize) {
    const contracts = await Promise.all(
      addresses
        .slice(i, i + chunkSize)
        .map((address) =>
          getContract(
            web3,
            web3Subscriptionable,
            lumerin,
            cloneFactory,
            address,
            walletAddress
          )
        )
    )
    eventBus.emit('contract-updated', {
      actives: contracts,
    })
    result.push(...contracts)
  }
  return result
}

/**
 * @param {import('web3').default} web3
 * @param {import('web3').default} web3Subscriptionable
 * @param {import('contracts-js').LumerinContext} lumerin
 * @param {string} contractId
 * @param {string} walletAddress
 */
async function getContract(
  web3,
  web3Subscriptionable,
  lumerin,
  cloneFactory,
  contractId,
  walletAddress
) {
  const contractEventsListener = ContractEventsListener.getInstance()
  const contractInfo = await _loadContractInstance(
    web3,
    contractId,
    walletAddress
  )

  contractEventsListener.addContract(
    contractInfo.data.id,
    Implementation(web3Subscriptionable, contractId),
    walletAddress
  )
  return contractInfo.data
}

/**
 * @param {import('contracts-js').CloneFactoryContext} cloneFactory
 */
const getMarketplaceFee = (cloneFactory) => async () => {
  return await cloneFactory.methods.marketplaceFee().call()
}

/**
 * @param {import('web3').default} web3
 * @param {import('contracts-js').CloneFactoryContext} cloneFactory
 */
function createContract(web3, cloneFactory) {
  if (!web3) {
    logger.error('Not a valid Web3 instance')
    return
  }

  return async function (params) {
    let {
      price,
      limit = 0,
      speed,
      duration,
      sellerAddress,
      profit = 0,
      validatorAddress = '0x0000000000000000000000000000000000000000',
      privateKey,
    } = params

    const isWhitelisted = await cloneFactory.methods
      .checkWhitelist(sellerAddress)
      .call()
    if (!isWhitelisted) {
      throw new Error('seller is not whitelisted')
    }

    const tempWallet = ethereumWallet.fromPrivateKey(
      Buffer.from(remove0xPrefix(privateKey), 'hex')
    )
    const pubKey = tempWallet.getPublicKey()

    const account = web3.eth.accounts.privateKeyToAccount(privateKey)
    web3.eth.accounts.wallet.create(0).add(account)
    const marketplaceFee = await cloneFactory.methods.marketplaceFee().call()

    const gas = await cloneFactory.methods
      .setCreateNewRentalContractV2(
        price,
        limit,
        speed,
        duration,
        +profit,
        validatorAddress,
        pubKey.toString('hex')
      )
      .estimateGas({
        from: sellerAddress,
        value: marketplaceFee,
      })

    return cloneFactory.methods
      .setCreateNewRentalContractV2(
        price,
        limit,
        speed,
        duration,
        +profit,
        validatorAddress,
        pubKey.toString('hex')
      )
      .send({ from: sellerAddress, gas, value: marketplaceFee })
  }
}

/**
 * @param {import('web3').default} web3
 */
function cancelContract(web3, cloneFactory) {
  if (!web3) {
    logger.error('Not a valid Web3 instance')
    return
  }

  return async function (params) {
    const {
      walletAddress,
      contractId,
      privateKey,
      closeOutType,
    } = params

    const account = web3.eth.accounts.privateKeyToAccount(privateKey)
    web3.eth.accounts.wallet.create(0).add(account)

    const marketplaceFee = await cloneFactory.methods.marketplaceFee().call()

    const gas = await Implementation(web3, contractId)
      .methods.setContractCloseOut(closeOutType)
      .estimateGas({
        from: walletAddress,
        value: marketplaceFee,
      })

    return await Implementation(web3, contractId)
      .methods.setContractCloseOut(closeOutType)
      .send({
        from: walletAddress,
        gas,
        value: marketplaceFee,
      })
  }
}

/**
 * @param {import('web3').default} web3
 * @param {import('contracts-js').CloneFactoryContext} cloneFactory
 */
function setContractDeleteStatus(web3, cloneFactory, onUpdate) {
  if (!web3) {
    logger.error('Not a valid Web3 instance')
    return
  }

  return async function (params) {
    const {
      walletAddress,
      contractId,
      privateKey,
      deleteContract,
    } = params

    const account = web3.eth.accounts.privateKeyToAccount(privateKey)
    web3.eth.accounts.wallet.create(0).add(account)

    const {
      data: { isDead },
    } = await _loadContractInstance(web3, contractId)
    if (Boolean(isDead) === Boolean(deleteContract)) {
      return true
    }

    const gas = await cloneFactory.methods
      .setContractDeleted(contractId, deleteContract)
      .estimateGas({
        from: walletAddress,
      })

    const result = await cloneFactory.methods
      .setContractDeleted(contractId, deleteContract)
      .send({
        from: walletAddress,
        gas,
      })
    onUpdate(contractId, walletAddress).catch((err) =>
      logger.error(`Failed to refresh after setContractDeadStatus: ${err}`)
    )
    return result
  }
}

/**
 *
 * @param {import('web3').default} web3
 * @param {import('contracts-js').CloneFactoryContext} cloneFactory
 * @param {import('contracts-js').LumerinContext} lumerin
 * @returns
 */
function purchaseContract(web3, cloneFactory, lumerin) {
  return async (params) => {
    const { walletId, contractId, url, privateKey, price, version } = params
    const sendOptions = { from: walletId }

    //getting pubkey from contract to be purchased
    const implementationContract = Implementation(web3, contractId)

    const pubKey = await implementationContract.methods.pubKey().call()

    //encrypting plaintext url parameter
    const ciphertext = await encrypt(
      Buffer.from(add65BytesPrefix(pubKey), 'hex'),
      Buffer.from(url)
    )

    const account = web3.eth.accounts.privateKeyToAccount(privateKey)
    web3.eth.accounts.wallet.create(0).add(account)

    const {
      data: { isDead, price: p },
    } = await _loadContractInstance(web3, contractId)
    if (isDead) {
      throw new Error('Contract is deleted already')
    }

    const increaseAllowanceEstimate = await lumerin.methods
      .increaseAllowance(cloneFactory.options.address, price)
      .estimateGas({
        from: walletId,
      })

    await lumerin.methods
      .increaseAllowance(cloneFactory.options.address, price)
      .send({
        from: walletId,
        gas: increaseAllowanceEstimate,
      })

    const marketplaceFee = await cloneFactory.methods.marketplaceFee().call()

    const purchaseGas = await cloneFactory.methods
      .setPurchaseRentalContract(
        contractId,
        ciphertext.toString('hex'),
        version
      )
      .estimateGas({
        from: sendOptions.from,
        value: marketplaceFee,
      })

    const purchaseResult = await cloneFactory.methods
      .setPurchaseRentalContract(
        contractId,
        ciphertext.toString('hex'),
        version
      )
      .send({
        ...sendOptions,
        gas: purchaseGas,
        value: marketplaceFee,
      })

    logger.debug('Finished puchase transaction', purchaseResult)
  }
}

/**
 *
 * @param {import('web3').default} web3
 * @param {import('contracts-js').CloneFactoryContext} cloneFactory
 * @param {import('contracts-js').LumerinContext} lumerin
 * @returns
 */
function editContract(web3, cloneFactory, lumerin) {
  return async (params) => {
    const {
      walletId,
      contractId,
      privateKey,
      price,
      limit = 0,
      speed,
      duration,
      profit = 0
    } = params
    const sendOptions = { from: walletId }

    const account = web3.eth.accounts.privateKeyToAccount(privateKey)
    web3.eth.accounts.wallet.create(0).add(account)

    const editGas = await cloneFactory.methods
      .setUpdateContractInformationV2(contractId, price, limit, speed, duration, +profit)
      .estimateGas({
        from: sendOptions.from,
      });
      
    const editResult = await cloneFactory.methods
      .setUpdateContractInformationV2(contractId, price, limit, speed, duration, +profit)
      .send({
        ...sendOptions,
        gas: editGas,
      })

    logger.debug('Finished edit contract transaction', editResult)
  }
}

module.exports = {
  getContracts,
  getContract,
  createContract,
  cancelContract,
  purchaseContract,
  setContractDeleteStatus,
  editContract,
  getMarketplaceFee,
}
