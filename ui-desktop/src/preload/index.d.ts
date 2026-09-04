type IpcUnsubscribe = () => void

type IpcRendererBridge = {
  send: (eventName: string, payload?: unknown) => void
  on: (
    eventName: string,
    listener: (payload: any, unsubscribe: IpcUnsubscribe) => void
  ) => IpcUnsubscribe
}

type ChatStreamEvent =
  | { requestId: string; kind: 'chunk'; dataBase64: string }
  | { requestId: string; kind: 'end' }
  | { requestId: string; kind: 'error'; message: string }

type ChatStreamApi = {
  start: (
    requestId: string,
    payload: unknown
  ) => Promise<{ ok: boolean; status: number; contentType: string }>
  cancel: (requestId: string) => void
  onEvent: (requestId: string, listener: (event: ChatStreamEvent) => void) => IpcUnsubscribe
}

type IpfsDownloadProgress = {
  status: 'downloading' | 'completed'
  downloaded: number
  total: number
  percentage: number
  timeUpdated: number
}

type IpfsDownloadApi = {
  selectFolder: () => Promise<{
    canceled: boolean
    folderToken?: string
  }>
  start: (requestId: string, folderToken: string, cidHash: string) => Promise<{ accepted: true }>
  cancel: (requestId: string) => void
  onEvent: (
    requestId: string,
    listener: (
      event:
        | { requestId: string; kind: 'progress'; progress: IpfsDownloadProgress }
        | { requestId: string; kind: 'error'; message: string }
    ) => void
  ) => IpcUnsubscribe
}

declare global {
  interface Window {
    ipcRenderer: IpcRendererBridge
    openLink: (url: string) => Promise<void>
    getAppVersion: () => string
    copyToClipboard: (text: string) => Promise<void>
    chatStream: ChatStreamApi
    ipfsDownload: IpfsDownloadApi
  }
}

export {}
