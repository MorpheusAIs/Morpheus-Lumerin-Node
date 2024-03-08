'use strict'

const { utils: { hexToUtf8 } } = require('web3')
// const LumerinContracts = require('metronome-contracts')
// const LumerinContracts = require('@lumerin/contracts')

const exportMetaParser = ({ returnValues }) => ({
  lumerin: {
    export: {
      blockTimestamp: returnValues.blockTimestamp,
      burnSequence: returnValues.burnSequence,
      currentBurnHash: returnValues.currentBurnHash,
      currentTick: returnValues.currentTick,
      dailyAuctionStartTime: returnValues.dailyAuctionStartTime,
      dailyMintable: returnValues.dailyMintable,
      destinationChain: hexToUtf8(returnValues.destinationChain),
      extraData: returnValues.extraData,
      fee: returnValues.fee,
      genesisTime: returnValues.genesisTime,
      previousBurnHash: returnValues.prevBurnHash,
      supply: returnValues.supplyOnChain,
      to: returnValues.destinationRecipientAddr,
      value: returnValues.amountToBurn
    }
  }
})

const importRequestMetaParser = ({ returnValues }) => ({
  lumerin: {
    importRequest: {
      currentBurnHash: returnValues.currentBurnHash,
      fee: returnValues.fee,
      originChain: hexToUtf8(returnValues.originChain),
      to: returnValues.destinationRecipientAddr,
      value: returnValues.amountToImport
    }
  }
})

const importMetaParser = ({ returnValues }) => ({
  lumerin: {
    import: {
      currentBurnHash: returnValues.currentHash,
      fee: returnValues.fee,
      originChain: hexToUtf8(returnValues.originChain),
      to: returnValues.destinationRecipientAddr,
      value: returnValues.amountImported
    }
  }
})

// function getEventDataCreator (chain) {
//   const {
//     abi,
//     address: contractAddress,
//     birthblock: minBlock
//   } = LumerinContracts[chain].TokenPorter

//   return [
//     address => ({
//       contractAddress,
//       abi,
//       eventName: 'LogExportReceipt',
//       filter: { exporter: address },
//       metaParser: exportMetaParser,
//       minBlock
//     }),
//     address => ({
//       contractAddress,
//       abi,
//       eventName: 'LogExportReceipt',
//       filter: { destinationRecipientAddr: address },
//       metaParser: exportMetaParser,
//       minBlock
//     }),
//     address => ({
//       contractAddress,
//       abi,
//       eventName: 'LogImportRequest',
//       filter: { destinationRecipientAddr: address },
//       metaParser: importRequestMetaParser,
//       minBlock
//     }),
//     address => ({
//       contractAddress,
//       abi,
//       eventName: 'LogImport',
//       filter: { destinationRecipientAddr: address },
//       metaParser: importMetaParser,
//       minBlock
//     })
//   ]
// }

module.exports = {
  // getEventDataCreator,
  exportMetaParser,
  importMetaParser,
  importRequestMetaParser
}
