const { app, BrowserWindow, ipcMain } = require('electron')
const { join } = require('node:path')
const temporary = process.argv[2]
if (!temporary) throw new Error('Use session-confirmation-smoke.mjs')
app.setPath('userData', join(temporary, 'user-data'))
const { showSessionConfirmation } = require(join(temporary, 'main/sessionConfirmation.cjs'))
app.on('web-contents-created', (_event, contents) => {
  contents.on('did-fail-load', (_e, code, description) =>
    console.error('Preview load failure', code, description)
  )
  contents.on('render-process-gone', (_e, details) =>
    console.error('Preview renderer exited', details.reason)
  )
})

app.whenReady().then(async () => {
  const owner = new BrowserWindow({
    width: 1200,
    height: 800,
    title: 'Session confirmation preview — no transactions',
    webPreferences: {
      sandbox: true,
      contextIsolation: true,
      nodeIntegration: false,
      devTools: false
    }
  })
  owner.setMenu(null)
  owner.webContents.setWindowOpenHandler(() => ({ action: 'deny' }))
  owner.webContents.session.setPermissionRequestHandler((_sender, _permission, callback) =>
    callback(false)
  )
  await owner.loadURL('http://127.0.0.1:5188/#/chat')
  const details = Object.freeze({
    modelId: `0x${'ab'.repeat(32)}`,
    duration: 1800,
    directPayment: false,
    failover: true
  })
  let started = false
  const open = async (directPayment = false) => {
    if (started || owner.isDestroyed()) return
    started = true
    try {
      const confirmation = showSessionConfirmation(owner, { ...details, directPayment })
      const view = owner.getChildWindows().at(-1)
      view?.webContents.once('did-finish-load', () => {
        console.log(
          'Confirmation rendered; click Cancel/Open session, or press Escape. R=stake, D=direct, N=narrow, W=wide when the parent has focus.'
        )
      })
      const approved = await confirmation
      console.log(
        JSON.stringify({
          approved,
          transactionSubmitted: false,
          remainingViews: owner.getChildWindows().length,
          remainingDecisionListeners: ipcMain.listenerCount('session-confirmation:respond')
        })
      )
    } finally {
      started = false
    }
  }
  owner.webContents.on('before-input-event', (_event, input) => {
    if (input.type !== 'keyDown') return
    if (input.key.toLowerCase() === 'r') void open(false)
    if (input.key.toLowerCase() === 'd') void open(true)
    if (input.key.toLowerCase() === 'n') owner.setSize(760, 680)
    if (input.key.toLowerCase() === 'w') owner.setSize(1200, 800)
  })
  console.log('Preview ready. Press R to open the confirmation; D for direct pay, N/W to resize.')
})
app.on('window-all-closed', () => app.quit())
