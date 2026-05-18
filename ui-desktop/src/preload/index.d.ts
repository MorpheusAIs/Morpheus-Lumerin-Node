import { ElectronAPI } from '@electron-toolkit/preload'

type IpcUnsubscribe = () => void

type IpcRendererBridge = {
  send: (eventName: string, payload?: unknown) => void
  on: (
    eventName: string,
    listener: (event: unknown, payload: any, unsubscribe: IpcUnsubscribe) => void
  ) => IpcUnsubscribe
}

declare global {
  interface Window {
    electron: ElectronAPI
    api: unknown
    ipcRenderer: IpcRendererBridge
    openLink: (url: string) => Promise<void>
    getAppVersion: () => string
    copyToClipboard: (text: string) => Promise<void>
    isDev: boolean
  }
}
