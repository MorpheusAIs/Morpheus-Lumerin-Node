const { create: createAxios } = require('axios')
import { getDb } from './database'
import logger from '../../logger'

export const startMonitoringHashrate = (url, period) => {
  const interval = setInterval(async () => {
    try {
      const items = (await createAxios({ baseURL: url })('/contracts')).data
      persistData(items)
    } catch (e) {
      logger.debug(e.message, 'Failed to poll hashrate')
      persistData()
    }
  }, period)
  return interval
}

export const persistData = (data) => {
  const db = getDb()
  const collection = db.collection('hashrate')

  if (!data) {
    return
  }

  data.forEach((item) => {
    const id = item.ID
    const currentHashrate = item.ResourceEstimatesActual['ema--5m']
    collection.insert({
      id,
      hashrate: currentHashrate,
      timestamp: new Date().getTime()
    })
  })
}

export default { startMonitoringHashrate }
