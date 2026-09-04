type IpcUnsubscribe = () => void

type IpcRendererBridge = {
  send: (eventName: string, payload?: unknown) => void
  on: (
    eventName: string,
    listener: (payload: any, unsubscribe: IpcUnsubscribe) => void
  ) => IpcUnsubscribe
}

declare global {
  interface Window {
    ipcRenderer: IpcRendererBridge
    openLink: (url: string) => Promise<void>
    getAppVersion: () => string
    copyToClipboard: (text: string) => Promise<void>
  }
}

export {}
