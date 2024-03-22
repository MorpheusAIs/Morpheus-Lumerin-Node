'use strict';

const abi = require('./erc20-abi.json');

const transferMetaParser = ({ returnValues }) => ({
  token: {
    event: 'Transfer',
    from: returnValues.from,
    to: returnValues.to,
    value: returnValues.value,
    processing: false
  }
});

const approvalMetaParser = ({ returnValues }) => ({
  token: {
    event: 'Approval',
    from: returnValues.owner,
    to: returnValues.spender,
    value: returnValues.value,
    processing: false
  }
});

const getEventDataCreators = contractAddress => [
  address => ({
    contractAddress,
    abi,
    eventName: 'Transfer',
    filter: { _from: address },
    metaParser: transferMetaParser
  }),
  address => ({
    contractAddress,
    abi,
    eventName: 'Transfer',
    filter: { _to: address },
    metaParser: transferMetaParser
  }),
  address => ({
    contractAddress,
    abi,
    eventName: 'Approval',
    filter: { _owner: address },
    metaParser: approvalMetaParser
  })
];

module.exports = {
  getEventDataCreators,
  approvalMetaParser,
  transferMetaParser
};
