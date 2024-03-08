//@ts-check

const { expect } = require('chai')
const { getIPsWithinRange } = require('./ip-range')

describe('getIPsWithinRange tests', () => {
  it('should correctly return IP within small range', () => {
    const from = '192.168.1.1'
    const to = '192.168.1.3'

    const ips = getIPsWithinRange(from, to)

    expect(ips).includes(from)
    expect(ips).includes(to)
    expect(new Set(ips).size).eq(3)
  })

  it('should correctly return IP within /24 range', () => {
    const from = '192.168.1.0'
    const to = '192.168.1.255'

    const ips = getIPsWithinRange(from, to)

    expect(ips).includes(from)
    expect(ips).includes(to)
    expect(new Set(ips).size).eq(256)
  })

  it('should correctly return IP within /16 range', () => {
    const from = '192.168.0.0'
    const to = '192.168.255.255'

    const ips = getIPsWithinRange(from, to)

    expect(ips).includes(from)
    expect(ips).includes(to)
    expect(new Set(ips).size).eq(65536)
  })
})