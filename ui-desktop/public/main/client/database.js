'use strict';

const { app } = require('electron');
const { promisify } = require('util');
const { sum, upperFirst } = require('lodash');
const Datastore = require('nedb');
const path = require('path');

const logger = require('../../logger');
const config = require('../../config');

const promisifyMethods = methods =>
  function (obj) {
    methods.forEach(function (method) {
      obj[`${method}Async`] = promisify(obj[method].bind(obj))
    });
    return obj;
  }

const promisifyCollection = promisifyMethods([
  'find',
  'insert',
  'findOne',
  'remove',
  'update'
]);

const collections = {};

function addCollection (name) {
  logger.verbose(`Creating database collection ${name}`);

  const newCollection = promisifyCollection(
    new Datastore({
      filename: path.join(
        app.getPath('userData'),
        `Database/${upperFirst(name)}`
      ),
      autoload: true
    })
  );

  collections[name] = newCollection;

  if (config.dbAutocompactionInterval) {
    newCollection.persistence.setAutocompactionInterval(config.dbAutocompactionInterval);
  }

  return newCollection;
}

function collection (name) {
  return collections[name] || addCollection(name);
}

function dropDatabase () {
  logger.verbose('Dropping database');

  return Promise.all(
    Object.keys(collections).map(name =>
      collections[name].removeAsync({}, { multi: true })
    )
  )
    .then(sum)
    .then(function (count) {
      logger.verbose(`Database drop - removed ${count} documents`);

      return count;
    })
    .catch(function (err) {
      logger.error('Database drop failed', err);

      throw err;
    });
}

const db = {
  collection,
  dropDatabase
};

const dbManager = {
  getDb: () => db
};

module.exports = dbManager;
