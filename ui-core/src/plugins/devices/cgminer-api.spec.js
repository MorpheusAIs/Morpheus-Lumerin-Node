//@ts-check

const { expect } = require('chai')
const { CGMinerApi } = require('./cgminer-api')

describe('parseJSON test', () => {
  it('should drop zero bytes', () => {
    const buffer = Buffer.from(`${JSON.stringify({ a: 1 })}\0`)
    const res = new CGMinerApi().parseJSON(buffer)
    expect(res.a).eq(1)
  })

  it('should normalize invalid json for BMMiner=2.0.0, API=3.1, Miner=30.0.1.3, Type=Antminer S9i', () => {
    const buffer = Buffer.from(
      `{"STATS":[{"BMMiner":"2.0.0","Miner":"30.0.1.3","Type":"Antminer S9i"}{"STATS":0,"ID":"BC50"}]}`
    )
    const res = new CGMinerApi().parseJSON(buffer)
    expect(res.STATS[0].STATS).eq(0)
    expect(res.STATS[0].ID).eq('BC50')
  })
})
