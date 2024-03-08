//@ts-check
const { expect } = require('chai')
const { Server } = require('net')
const { detectMiner, detectMinersByRange } = require('./device-finder')
const { AbortController } = require('@azure/abort-controller')
const EventEmitter = require('events')

describe('detectMiner tests', () => {
  it('should reject if no connection', (done) => {
    const host = '255.255.255.0'
    detectMiner({ host, abort: AbortController.timeout(500) }, (data) => {
      expect(data.host).to.eq(host)
      expect(data.isHostUp).to.be.false
      expect(data.isDone).to.be.true
      done()
    }).catch((err) => {
      // ignoring promise rejection, as we track progress in callback
    })
  })

  it('should set isHostUp if there is a connection', (done) => {
    const port = 4028
    const host = '127.0.0.1'
    const testServer = new Server((socket) => {
      socket.write('Hello world')
      socket.end()
    })
    testServer.listen(port)

    const abort = new AbortController()

    detectMiner({ host, abort: abort.signal }, async (data) => {
      expect(data.host).to.eq(host)
      expect(data.isHostUp).to.be.true
      abort.abort()
      await new Promise((res) => testServer.close(res))
      done()
    })
  })

  it('should abort on signal', (done) => {
    const host = '127.0.0.1'
    const requestTimeout = 1 * 1000
    const connectTimeout = 3 * 1000

    class TestSocket extends EventEmitter {
      connect() {
        this.timeout = setTimeout(() => {
          this.emit('connect')
        }, connectTimeout)
      }

      destroy() {
        clearTimeout(this.timeout)
      }
    }

    const SocketFactory = () => {
      return new TestSocket()
    }

    const abort = new AbortController()
    setTimeout(() => abort.abort(), requestTimeout)

    detectMiner(
      {
        host,
        abort: abort.signal,
        //@ts-ignore
        SocketFactory,
      },
      (data) => {
        expect(data.host).to.be.eq(host)
        expect(data.isHostUp).to.be.false
        expect(data.isDone).to.be.true
        done()
      }
    )
  }).timeout(5000)
})

/**
 * Can be used to run detection without involving a client
 * Use describe and run tests
 */
describe.skip('It should detect miners in  network', () => {
  it('detect', async () => {
    await detectMinersByRange(
      ['192.168.1.1', '192.168.255.255'],
      AbortController.timeout(100000),
      (data) => {
        console.log(
          new Date(),
          data.host,
          'HOST UP',
          data.isHostUp,
          'API',
          data.hashRateGHS
        )
      }
    )
  }).timeout(40000)
})
