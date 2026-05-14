// Browser stub for Electron IPC - allows renderer to run as a web app
// All IPC calls will return pending promises (no Electron main process)

const noop = () => {};

const ipcListeners: Record<string, Array<(event: any, payload: any, unsub: () => void) => void>> = {};

const ipcRendererStub = {
  send(eventName: string, payload: any) {
    console.debug('[IPC stub] send:', eventName, payload);
  },
  on(eventName: string, listener: (event: any, payload: any, unsub: () => void) => void) {
    if (!ipcListeners[eventName]) {
      ipcListeners[eventName] = [];
    }
    const subscription = listener;
    ipcListeners[eventName].push(subscription);

    const unsubscribe = () => {
      const arr = ipcListeners[eventName];
      if (arr) {
        const idx = arr.indexOf(subscription);
        if (idx !== -1) arr.splice(idx, 1);
      }
    };
    return unsubscribe;
  },
};

if (typeof window !== 'undefined') {
  (window as any).ipcRenderer = ipcRendererStub;
  (window as any).openLink = (url: string) => { window.open(url, '_blank'); };
  (window as any).getAppVersion = () => '1.0.0';
  (window as any).copyToClipboard = (text: string) => {
    navigator.clipboard?.writeText(text).catch(noop);
  };
  (window as any).isDev = true;
  (window as any).electron = {};
  (window as any).api = {};
}
