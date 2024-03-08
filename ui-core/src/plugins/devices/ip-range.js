//@ts-check
const ip = require('ip')
const os = require('os')

/**
 * Returns all IPs withing local networks
 * @returns {String[]}
 */
const getHostsInLocalSubnets = () => {
  const interfaces = os.networkInterfaces()
  const result = []

  for (const key in interfaces) {
    const addresses = interfaces[key]
    if (!addresses) {
      continue
    }
    for (let i = addresses.length; i--; ) {
      const address = addresses[i]
      if (address.family === 'IPv4' && !address.internal) {
        const subnet = ip.subnet(address.address, address.netmask)
        const ips = getIPsWithinRange(subnet.firstAddress, subnet.lastAddress)
        result.push(...ips)
      }
    }
  }

  return result
}

/**
 * Returns all IP addresses within range
 * @param {String} fromIp
 * @param {String} toIp
 * @returns {String[]}
 */
const getIPsWithinRange = (fromIp, toIp) => {
  const from = ip.toLong(fromIp)
  const to = ip.toLong(toIp)
  const IPs = []
  for (let i = from; i <= to; i++) {
    IPs.push(ip.fromLong(i))
  }
  return IPs
}

module.exports = {
  getIPsWithinRange,
  getHostsInLocalSubnets,
}