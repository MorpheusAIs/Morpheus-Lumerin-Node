import { ipcRenderer } from 'electron'

// This preload belongs only to the main-owned confirmation surface. It
// exposes no bridge/API to either renderer and never imports the app preload.
const requestId = process.argv
  .find((argument) => argument.startsWith('--session-confirmation-id='))
  ?.slice('--session-confirmation-id='.length)
let decided = false

function decide(approved: boolean): void {
  if (decided || !requestId || !/^[a-f0-9-]{36}$/i.test(requestId)) return
  decided = true
  for (const button of document.querySelectorAll<HTMLButtonElement>('button')) {
    button.disabled = true
  }
  ipcRenderer.send('session-confirmation:respond', { requestId, approved })
}

window.addEventListener('DOMContentLoaded', () => {
  document.getElementById('session-cancel')?.focus()
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
