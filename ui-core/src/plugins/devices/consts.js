const EVENT_DEVICES_DEVICE_UPDATED = 'devices-device-updated'
const EVENT_DEVICES_STATE_UPDATED = 'devices-state-updated'

const Status = {
  Success: 'S',
  Error: 'E',
}

/** @type {{Miner: 'miner', Unknown:'unknown'}} */
const DeviceType = {
  Miner: 'miner',
  Unknown: 'unknown',
}

const MINER_API_PORT = 4028

module.exports = {
  EVENT_DEVICES_DEVICE_UPDATED,
  EVENT_DEVICES_STATE_UPDATED,
  MINER_API_PORT,
  Status,
  DeviceType,
}
