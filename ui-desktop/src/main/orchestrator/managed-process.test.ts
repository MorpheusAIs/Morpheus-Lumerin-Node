import net from 'node:net'
import http from 'node:http'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ManagedProcess } from './managed-process'
import { ProcessFactory } from './process-factory'

const logger = {
  info: vi.fn(),
  warn: vi.fn(),
  error: vi.fn()
}

const servers: net.Server[] = []

afterEach(async () => {
  await Promise.all(
    servers.splice(0).map(
      (server) =>
        new Promise<void>((resolve, reject) => {
          server.close((error) => (error ? reject(error) : resolve()))
        })
    )
  )
  vi.clearAllMocks()
})

async function listenOnEphemeralPort(): Promise<number> {
  const server = net.createServer()
  servers.push(server)

  await new Promise<void>((resolve, reject) => {
    server.once('error', reject)
    server.listen(0, '127.0.0.1', resolve)
  })

  const address = server.address()
  if (!address || typeof address === 'string') {
    throw new Error('Test server did not expose a TCP port')
  }
  return address.port
}

describe('ManagedProcess listener ownership', () => {
  it('fails closed when an unowned healthy listener cannot be trusted', async () => {
    const port = await listenOnEphemeralPort()
    const process = new ManagedProcess({
      command: '/bin/true',
      args: [],
      log: logger as any,
      ports: [port],
      pinger: { ping: vi.fn().mockResolvedValue(undefined) },
      adoptIfDetected: false
    })

    await expect(process.start()).rejects.toThrow(
      `Port ${port} is already serving a process, but the app cannot verify that it owns that process`
    )
    expect(process.getState()).toBe('stopped')
    expect(process.getError()).toContain('Close it, then restart this service from Settings')
  })

  it('preserves opt-in adoption for services that permit it', async () => {
    const port = await listenOnEphemeralPort()
    const process = new ManagedProcess({
      command: '/bin/true',
      args: [],
      log: logger as any,
      ports: [port],
      pinger: { ping: vi.fn().mockResolvedValue(undefined) },
      adoptIfDetected: true
    })

    await expect(process.start()).resolves.toBeUndefined()
    expect(process.getState()).toBe('running')
    expect(logger.info).toHaveBeenCalledWith(
      `port(s) ${port} already serving a healthy instance; adopting it`
    )
  })

  it('does not let a direct ping promote an unowned listener to running', async () => {
    const pinger = { ping: vi.fn().mockResolvedValue(undefined) }
    const process = new ManagedProcess({
      command: '/bin/true',
      args: [],
      log: logger as any,
      pinger,
      adoptIfDetected: false
    })

    await expect(process.ping()).rejects.toThrow('no app-owned service process is running')
    expect(pinger.ping).not.toHaveBeenCalled()
    expect(process.getState()).toBe('stopped')
  })

  it('rotates the child-bound health identity on every spawn', async () => {
    const port = await findEphemeralPort()
    const serverScript = [
      "const http = require('node:http')",
      "const token = process.env.MORPHEUS_DESKTOP_INSTANCE_TOKEN || ''",
      'http.createServer((_req, res) => {',
      "res.setHeader('X-Morpheus-Instance-Token', token)",
      "res.end('ok')",
      `}).listen(${port}, '127.0.0.1')`
    ].join(';')

    const managedProcess = await ProcessFactory({
      command: globalThis.process.execPath,
      args: ['-e', serverScript],
      log: logger as any,
      ports: [port],
      probe: { url: `http://127.0.0.1:${port}`, interval: 10, timeout: 3000 },
      reclaimIfDetected: true,
      adoptIfDetected: false
    })

    try {
      await expect(managedProcess.start()).resolves.toBeUndefined()
      expect(managedProcess.getState()).toBe('running')
      await expect(managedProcess.ping(500)).resolves.toBeUndefined()
      const firstIdentity = await readProcessIdentity(port)

      await managedProcess.stop()
      await managedProcess.start()
      const secondIdentity = await readProcessIdentity(port)

      expect(firstIdentity).toBeTruthy()
      expect(secondIdentity).toBeTruthy()
      expect(secondIdentity).not.toBe(firstIdentity)
    } finally {
      await managedProcess.stop()
    }
  })
})

async function findEphemeralPort(): Promise<number> {
  const server = net.createServer()
  await new Promise<void>((resolve, reject) => {
    server.once('error', reject)
    server.listen(0, '127.0.0.1', resolve)
  })
  const address = server.address()
  if (!address || typeof address === 'string') {
    throw new Error('Test server did not expose a TCP port')
  }
  await new Promise<void>((resolve, reject) => {
    server.close((error) => (error ? reject(error) : resolve()))
  })
  return address.port
}

async function readProcessIdentity(port: number): Promise<string | undefined> {
  return new Promise((resolve, reject) => {
    const request = http.get(`http://127.0.0.1:${port}`, (response) => {
      response.resume()
      response.once('end', () => {
        const value = response.headers['x-morpheus-instance-token']
        resolve(Array.isArray(value) ? value[0] : value)
      })
    })
    request.once('error', reject)
  })
}
