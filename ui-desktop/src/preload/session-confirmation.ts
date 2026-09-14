import { ipcRenderer } from 'electron'

// This preload belongs only to the main-owned confirmation surface. It
// exposes no bridge/API to either renderer and never imports the app preload.
const CHANNEL = 'session-confirmation:respond'
const requestId = process.argv
  .find((argument) => argument.startsWith('--session-confirmation-id='))
  ?.slice('--session-confirmation-id='.length)
const identified = !!requestId && /^[a-f0-9-]{36}$/iu.test(requestId)
let decided = false

function decide(approved: boolean): void {
  if (decided || !identified) return
  decided = true
  for (const button of document.querySelectorAll<HTMLButtonElement>('button')) {
    button.disabled = true
  }
  ipcRenderer.send(CHANNEL, { requestId, approved })
}

window.addEventListener('DOMContentLoaded', () => {
  document.getElementById('session-cancel')?.focus()
  // Report that the controls are bound. Until this existed, a preload that
  // never ran was indistinguishable from a user taking their time: the window
  // sat there, the buttons did nothing, and the only outcome was the 60 second
  // timeout arriving as a bare "Session opening cancelled."
  if (identified) ipcRenderer.send(CHANNEL, { requestId, ready: true })
})

document.addEventListener('click', (event) => {
  if (!event.isTrusted || !(event.target instanceof Element)) return
  const button = event.target.closest<HTMLButtonElement>('button[data-decision]')
  if (!button || button.disabled) return
  const decision = button.dataset.decision
  if (decision === 'approve' || decision === 'cancel') decide(decision === 'approve')
})

document.addEventListener('keydown', (event) => {
  if (!event.isTrusted) return
  if (event.key === 'Escape') {
    event.preventDefault()
    decide(false)
  } else if (event.key === 'Tab') {
    const buttons = Array.from(
      document.querySelectorAll<HTMLButtonElement>('button:not(:disabled)')
    )
    const first = buttons[0]
    const last = buttons[buttons.length - 1]
    if (
      event.shiftKey &&
      (document.activeElement === first ||
        !buttons.includes(document.activeElement as HTMLButtonElement))
    ) {
      event.preventDefault()
      last?.focus()
    } else if (
      !event.shiftKey &&
      (document.activeElement === last ||
        !buttons.includes(document.activeElement as HTMLButtonElement))
    ) {
      event.preventDefault()
      first?.focus()
    }
  }
})
