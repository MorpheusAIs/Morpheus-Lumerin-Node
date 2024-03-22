'use strict'

const debug = require('debug')('lmr-wallet:core:contracts:api')
const { CloneFactory, Implementation, Lumerin } = require('contracts-js')

/**
 * @param {CloneFactory} cloneFactory
 */
async function _getContractAddresses(cloneFactory) {
  return await cloneFactory.methods
    .getContractList()
    .call()
    .catch((error) => {
      debug(
        'Error when trying get list of contract addresses from CloneFactory contract: ',
        error
      )
    })
}

/**
 * @param {web3} web3
 * @param {string} implementationAddress
 */
async function _loadContractInstance(web3, implementationAddress) {
  try {
    const implementationContract = Implementation(web3, implementationAddress)
    const contract = await implementationContract.methods
      .getPublicVariables()
      .call()

    const {
      0: state,
      1: price, // cost to purchase the contract
      2: limit, // max th provided
      3: speed, // th/s of contract
      4: length, // duration of the contract in seconds
      5: timestamp, // timestamp of the block at moment of purchase
      6: buyer, // wallet address of the purchasing party
      7: seller, // wallet address of the selling party
      8: encryptedPoolData, // encrypted data for pool target info
    } = contract

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
      },
      instance: implementationContract,
    }
  } catch (err) {
    debug(
      'Error when trying to load Contracts by address in the Implementation contract: ',
      err
    )
  }
}

/**
 * @param {web3} web3
 * @param {Lumerin} lumerin
 * @param {CloneFactory} cloneFactory
 */
async function getActiveContracts(web3, lumerin, cloneFactory) {
  if (!web3) {
    debug('Not a valid Web3 instance')
    return
  }
  const addresses = await _getContractAddresses(cloneFactory)

  return Promise.all(addresses.map(async a => {
    const contract = await _loadContractInstance(web3, a)
    const balance = await lumerin.methods.balanceOf(contract.data.id).call();
    return {
      ...contract.data,
      balance,
    };
  }));
}
        
/**
 * @param {web3} web3
 * @param {CloneFactory} cloneFactory
 */
function createContract(web3, cloneFactory, plugins) {
  if (!web3) {
    debug('Not a valid Web3 instance')
    return
  }

  return async function (params) {
    // const { gasPrice } = await plugins.wallet.getGasPrice()
    let {
      price,
      limit = 0,
      speed,
      duration,
      sellerAddress,
      validatorAddress = '0x0000000000000000000000000000000000000000',
      password,
      privateKey,
    } = params

    const account = web3.eth.accounts.privateKeyToAccount(privateKey)

    web3.eth.accounts.wallet.create(0).add(account)


    return web3.eth
      .getTransactionCount(sellerAddress, 'pending')
      .then((nonce) =>
        plugins.explorer.logTransaction(
          cloneFactory.methods
            .setCreateNewRentalContract(
              price,
              limit,
              speed,
              duration,
              validatorAddress,
              ''
            )
            .send(
              {
                from: sellerAddress,
                gas: 500000,
              },
              function (data, err) {
                console.log('error: ', err)
                console.log('data: ', data)
              }
            ),
          sellerAddress
        )
      )
  }
}

// function updateContract(web3, chain) {
//   if(!web3) {
//     debug('Not a valid Web3 instance');
//     return;
//   }

//   return function(params) {
//     // const { Implementation } = LumerinContracts(web3, chain)
//     //   .createContract(LumerinContracts[chain].Implementation.abi, address);
//     const implementationContract = _loadContractInstance(web3, chain, address);
//     const isRunning = implementationContract.contractState() === 'Running';

//     if(isRunning) {
//       debug("Contract is currently in the 'Running' state");
//       return;
//     }

//     implementationContract.methods.setUpdatePurchaseInformation()
//   }
// }

function cancelContract(web3) {
  if (!web3) {
    debug('Not a valid Web3 instance')
    return
  }

  return async function (params) {
    const {
      walletAddress,
      gasLimit = 1000000,
      contractId,
      privateKey,
      closeOutType,
    } = params

    const account = web3.eth.accounts.privateKeyToAccount(privateKey)
    web3.eth.accounts.wallet.create(0).add(account)

    const implementationContract = await _loadContractInstance(
      web3,
      contractId
    )
    const isRunning = implementationContract.data.state === '1'

    if (isRunning && closeOutType !== 1) {
      debug("Contract is currently in the 'Running' state")
      return
    }

    return implementationContract.instance.methods
      .setContractCloseOut(closeOutType)
      .send({
        from: walletAddress,
        gas: gasLimit,
      })
  }
}

function purchaseContract(web3, cloneFactory, lumerin) {
  return async (params) => {
    const { walletId, contractId, url, privateKey, price } = params;
    const sendOptions = { from: walletId, gas: 1_000_000}

    const account = web3.eth.accounts.privateKeyToAccount(privateKey);
    web3.eth.accounts.wallet.create(0).add(account);
    
    await lumerin.methods
      .increaseAllowance(cloneFactory.options.address, price)
      .send(sendOptions);

    const purchaseResult = await cloneFactory.methods
      .setPurchaseRentalContract(contractId, url)
      .send(sendOptions);

    debug(`Finished puchase transaction`, purchaseResult);
  }
}

module.exports = {
  getActiveContracts,
  createContract,
  cancelContract,
  purchaseContract
}
