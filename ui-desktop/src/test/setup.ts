import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, vi } from 'vitest'

// Unmount React trees between tests so a leaked component can't affect the next.
afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

// The preload script exposes these on `window` at runtime via contextBridge.
// Tests that care about them should override with their own spies.
Object.defineProperty(window, 'ipcRenderer', {
  writable: true,
  value: {
    send: vi.fn(),
    on: vi.fn(() => () => {})
  }
})

Object.defineProperty(window, 'openLink', { writable: true, value: vi.fn() })
Object.defineProperty(window, 'copyToClipboard', {
  writable: true,
  value: vi.fn(() => Promise.resolve())
})
Object.defineProperty(window, 'getAppVersion', {
  writable: true,
  value: vi.fn(() => '0.0.0-test')
})
