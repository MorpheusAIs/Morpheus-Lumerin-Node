'use strict'

const { app } = require('electron')
const { sum, upperFirst } = require('lodash')
// The original `nedb` calls util.isDate, which Node 23 removed (Electron 43
// embeds Node 24). This fork reads the same data files and has native *Async
// methods.
const Datastore = require('@seald-io/nedb')
const path = require('path')

import logger from '../../logger'
import config from '../../config'

const collections = {}

function addCollection(name) {
  logger.verbose(`Creating database collection ${name}`)

  const newCollection = new Datastore({
    filename: path.join(app.getPath('userData'), `Database/${upperFirst(name)}`),
    autoload: true
  })

  collections[name] = newCollection

  if (config.dbAutocompactionInterval) {
    newCollection.setAutocompactionInterval(config.dbAutocompactionInterval)
  }

  return newCollection
}

function collection(name) {
  return collections[name] || addCollection(name)
}

function dropDatabase() {
  logger.verbose('Dropping database')

  return Promise.all(
    Object.keys(collections).map((name) => collections[name].removeAsync({}, { multi: true }))
  )
    .then(sum)
    .then(function (count) {
      logger.verbose(`Database drop - removed ${count} documents`)

      return count
    })
    .catch(function (err) {
      logger.error('Database drop failed', err)

      throw err
    })
}

const db = {
  collection,
  dropDatabase
}
export const getDb = () => db
const dbManager = {
  getDb
}

export default dbManager
