import { BrowserWindow, ipcMain, type IpcMainEvent } from 'electron'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'
import {
  renderSessionConfirmationHtml,
  type SessionConfirmationDetails
} from './sessionConfirmationView'

export const SESSION_CONFIRMATION_CHANNEL = 'session-confirmation:respond'
export const SESSION_CONFIRMATION_TIMEOUT_MS = 60_000
/**
 * How long the prompt has to report that its controls are live.
 *
 * The preload pings once the document is parsed. If that ping never arrives the
 * window cannot be answered, so waiting out the full minute only delays a
 * failure we already know about.
 */
export const SESSION_CONFIRMATION_ATTACH_TIMEOUT_MS = 5_000
let confirmationActive = false
let lastDismissal = ''

/**
 * Why the most recent prompt ended without the user deciding, once.
 *
 * The caller turns a `false` into one sentence for the user, and until now that
 * sentence was the same whether they pressed Cancel or the prompt died on its
 * own. Reading this lets it say which, so a failure in the field is diagnosable
 * from the toast instead of from a terminal the user has to be walked into.
 * Reading clears it: a stale reason on a later cancel is worse than none.
 */
export function consumeSessionConfirmationDismissal(): string {
  const reason = lastDismissal
  lastDismissal = ''
  return reason
}

/**
 * Main-owned, isolated web contents above the app, not an approval boolean
 * supplied by the app renderer. The underlying renderer cannot approve itself.
 */
export async function showSessionConfirmation(
  owner: BrowserWindow | null | undefined,
  details: SessionConfirmationDetails
): Promise<boolean> {
  if (confirmationActive) throw new Error('Another security confirmation is already open.')
  lastDismissal = ''
  if (!owner || owner.isDestroyed() || owner.webContents.isDestroyed()) {
    // The one refusal that used to be completely silent: no window, no log, no
    // reason, and an instant "Session opening cancelled." with nothing to go on.
    lastDismissal = 'there was no app window to attach the prompt to'
    console.warn(`Session confirmation not shown: ${lastDismissal}`)
    return false
  }
  const ownerContents = owner.webContents
  confirmationActive = true
  try {
    const requestId = randomUUID()
    /**
     * The nonce is this prompt's identity inside the document itself.
     *
     * It is hex, so `encodeURIComponent` leaves it byte-for-byte intact in the
     * `data:` URL, and it appears nowhere else. That makes it the one part of
     * the URL we can compare against later without betting on Chromium handing
     * `getURL()` back exactly as `loadURL()` received it.
     */
    const nonce = requestId.replace(/-/g, '')
    const html = renderSessionConfirmationHtml({ ...details }, nonce)
    const url = `data:text/html;charset=utf-8,${encodeURIComponent(html)}`
    const view = new BrowserWindow({
      parent: owner,
      modal: true,
      frame: false,
      transparent: true,
      show: false,
      resizable: false,
      movable: false,
      minimizable: false,
      maximizable: false,
      fullscreenable: false,
      skipTaskbar: true,
      hasShadow: false,
      webPreferences: {
        preload: join(__dirname, '../preload/session-confirmation.js'),
        additionalArguments: [`--session-confirmation-id=${requestId}`],
        partition: 'session-confirmation',
        contextIsolation: true,
        sandbox: true,
        nodeIntegration: false,
        webviewTag: false,
        webSecurity: true,
        devTools: false
      }
    })
    const contents = view.webContents

    return await new Promise<boolean>((resolve) => {
      let settled = false
      /**
       * Which teardown path dismissed the prompt, or '' when the user decided.
       *
       * Nine different events resolve this promise with `false`, and the caller
       * turns every one of them into the same "Session opening cancelled."
       * Without this, a prompt that dies on its own is indistinguishable from
       * the user pressing Cancel, which is exactly the state a report of
       * "it cancels every time" leaves you in.
       */
      let dismissedBecause = ''
      /** Set once the preload reports the prompt's controls are live. */
      let attached = false
      let attachTimer: ReturnType<typeof setTimeout> | undefined
      const finish = (approved = false) => {
        if (settled) return
        settled = true
        clearTimeout(timeout)
        clearTimeout(attachTimer)
        ipcMain.removeListener(SESSION_CONFIRMATION_CHANNEL, onDecision)
        owner.removeListener('closed', cancel)
        owner.removeListener('resize', resize)
        owner.removeListener('move', resize)
        ownerContents.removeListener('destroyed', cancel)
        ownerContents.removeListener('did-start-navigation', onOwnerNavigation)
        ownerContents.removeListener('render-process-gone', cancel)
        contents.removeListener('destroyed', cancel)
        contents.removeListener('render-process-gone', cancel)
        contents.removeListener('did-fail-load', onLoadFailed)
        contents.removeListener('will-navigate', denyNavigation)
        contents.removeListener('will-redirect', denyNavigation)
        // Native objects may disappear between events. Teardown must never
        // leave the caller waiting or turn a failed cleanup into permission.
        for (const cleanup of [
          // Close before destroying. On Windows a modal child disables its
          // parent through EnableWindow, and `destroy()` deliberately skips
          // the close path that would undo that, leaving a main window that
          // still paints but ignores every click and keystroke until the app
          // is restarted.
          //
          // This used to call `setParentWindow(null)`, which looks like the
          // direct way to sever the link. Electron throws
          // "Can not be called for modal window" for precisely the windows
          // that need it, so the call failed on every approval, on every
          // platform, and the catch below turned the user's Open session into
          // "Session opening cancelled." `close()` runs the real close path
          // and is legal here; `destroy()` below still guarantees teardown.
          () => {
            if (!view.isDestroyed()) view.close()
          },
          () => {
            if (!view.isDestroyed()) view.destroy()
          },
          // Belt and braces: if the parent was left disabled regardless, undo
          // it directly rather than trusting the platform to have done so.
          () => {
            if (!owner.isDestroyed() && !owner.isEnabled()) owner.setEnabled(true)
          },
          () => {
            if (!owner.isDestroyed() && !ownerContents.isDestroyed()) ownerContents.focus()
          }
        ]) {
          try {
            cleanup()
          } catch (error) {
            // This is the last silent path, and it is the worst one: it turns
            // an Approve the user just gave into a cancel. Nothing else here
            // can produce a reasonless `false`, so a bare "Session opening
            // cancelled." after a visible prompt means teardown threw. Fail
            // closed, but never again without saying which call did it.
            approved = false
            dismissedBecause ||= `the prompt could not be torn down (${
              error instanceof Error ? error.message : 'unknown teardown error'
            })`
          }
        }
        if (!approved && dismissedBecause) {
          lastDismissal = dismissedBecause
          console.warn(`Session confirmation dismissed without a decision: ${dismissedBecause}`)
        }
        resolve(approved)
      }
      const cancelBecause =
        (reason: string) =>
        (): void => {
          dismissedBecause = reason
          finish(false)
        }
      const cancel = cancelBecause('the app window or the prompt was torn down')
      const onTimeout = cancelBecause('no decision arrived before the timeout')
      const onOwnerReplaced = cancelBecause('the app renderer loaded a different document')
      const onPromptInert = cancelBecause('the prompt opened but its controls never came up')
      /**
       * A failed load is the one dismissal that knows exactly what went wrong,
       * and it used to throw that away. Chromium's error code is the difference
       * between "the document was rejected" and "something tore the window
       * down", so it travels with the reason all the way to the user.
       */
      const onLoadFailed = (
        _event: Electron.Event,
        errorCode: number,
        errorDescription: string
      ): void => {
        dismissedBecause = `the prompt document failed to load (${
          errorDescription || 'no description'
        } ${errorCode})`
        finish(false)
      }
      const onDecision = (event: IpcMainEvent, payload: unknown) => {
        if (settled) return
        if (owner.isDestroyed() || ownerContents.isDestroyed() || contents.isDestroyed())
          return cancel()
        /**
         * Assert the prompt is still showing *this* document, not that its URL
         * round-tripped unchanged.
         *
         * This used to be `contents.getURL() !== url`. Chromium is free to
         * re-canonicalise a `data:` URL — percent-encoding of characters that
         * `encodeURIComponent` leaves alone is the obvious way that differs —
         * and this document is a whole rendered HTML page inlined into the URL,
         * so it is a long string with plenty of surface for that to happen on.
         * Any drift at all silently dropped the user's Approve, and the caller
         * reported the resulting timeout as "Session opening cancelled."
         *
         * Scheme plus nonce keeps what the check was for. The scheme pins it to
         * the inline document rather than anything navigated to (`about:blank`
         * after a teardown being the case that matters), and the nonce pins it
         * to this request. Navigation is blocked, new windows are denied, and
         * the sender, sender-frame and requestId checks around this one carry
         * the rest of the binding.
         */
        const shown = contents.getURL()
        if (
          event.sender !== contents ||
          event.senderFrame !== contents.mainFrame ||
          !shown.startsWith('data:text/html') ||
          !shown.includes(nonce) ||
          !payload ||
          typeof payload !== 'object' ||
          Array.isArray(payload)
        )
          return
        const decision = payload as { requestId?: unknown; approved?: unknown; ready?: unknown }
        if (decision.requestId !== requestId) return
        /**
         * The preload says its handlers are bound before it says anything else.
         *
         * A prompt whose preload never ran looks exactly like a prompt the user
         * is still reading: the window is up, and nothing happens. The only
         * signal was the full 60 second timeout, reported as a plain cancel.
         * One ping distinguishes the two, so an unanswerable prompt fails in
         * seconds and says that is what happened.
         */
        if (decision.ready === true) {
          attached = true
          clearTimeout(attachTimer)
          return
        }
        if (typeof decision.approved !== 'boolean') return
        // The decision the user actually made, before anything downstream can
        // reinterpret it. Without this line an approval that teardown downgrades
        // and a genuine Cancel are the same single word in the log.
        console.warn(
          `Session confirmation decision: ${decision.approved ? 'approved' : 'cancelled by the user'}`
        )
        finish(decision.approved)
      }
      const resize = () => {
        if (owner.isDestroyed() || contents.isDestroyed()) return cancel()
        const { x, y, width, height } = owner.getContentBounds()
        view.setBounds({ x, y, width: Math.max(1, width), height: Math.max(1, height) })
      }
      /**
       * Only a real document swap invalidates the prompt.
       *
       * This listener used to cancel on any main-frame navigation, ignoring
       * `isInPlace`. The app renders under a HashRouter, so every in-app route
       * change, every `?query` update and every history.replaceState is a
       * same-document main-frame navigation and arrived here as a cancel — and
       * React flushes the state updates queued by the open button *during* the
       * await that is showing this prompt, so a route change landing while the
       * user is still reading it is ordinary, not exceptional. The document the
       * prompt belongs to is still there in that case, so there is nothing to
       * invalidate.
       */
      const onOwnerNavigation = (
        _event: Electron.Event,
        _url: string,
        isInPlace: boolean,
        isMainFrame: boolean
      ) => {
        if (isMainFrame && !isInPlace) onOwnerReplaced()
      }
      const denyNavigation = (event: Electron.Event) => {
        event.preventDefault()
        dismissedBecause = 'the prompt attempted to navigate'
        finish(false)
      }
      const timeout = setTimeout(onTimeout, SESSION_CONFIRMATION_TIMEOUT_MS)
      ipcMain.on(SESSION_CONFIRMATION_CHANNEL, onDecision)
      owner.once('closed', cancel)
      owner.on('resize', resize)
      owner.on('move', resize)
      ownerContents.once('destroyed', cancel)
      ownerContents.on('did-start-navigation', onOwnerNavigation)
      ownerContents.once('render-process-gone', cancel)
      contents.once('destroyed', cancel)
      contents.once('render-process-gone', cancel)
      contents.once('did-fail-load', onLoadFailed)
      contents.on('will-navigate', denyNavigation)
      contents.on('will-redirect', denyNavigation)
      try {
        contents.session.setPermissionRequestHandler((_contents, _permission, callback) =>
          callback(false)
        )
        contents.session.setPermissionCheckHandler(() => false)
        contents.session.webRequest.onBeforeRequest(
          { urls: ['http://*/*', 'https://*/*', 'ws://*/*', 'wss://*/*', 'file://*/*'] },
          (_details, callback) => callback({ cancel: true })
        )
        contents.setWindowOpenHandler(() => ({ action: 'deny' }))
        // `setMenu` is Linux and Windows only. On macOS it is not a function,
        // so calling it threw a TypeError here — inside the setup try/catch
        // that fails closed — and the prompt was cancelled before `loadURL`
        // ever ran. No window appeared, and the caller reported the instant
        // rejection as "Session opening cancelled." every single time.
        // There is no window menu bar to strip on macOS in the first place.
        view.setMenu?.(null)
        resize()
        void contents
          .loadURL(url)
          .then(() => {
            if (settled || contents.isDestroyed()) return
            resize()
            view.show()
            contents.focus()
            // Whether the window is up, and at what size. A prompt nobody can
            // see is the difference between "it cancelled itself" and "I was
            // never asked", and the bounds are how an off-screen or zero-sized
            // one tells you which.
            console.warn(
              `Session confirmation prompt shown: ${JSON.stringify(view.getBounds?.() ?? 'bounds unavailable')}`
            )
            // The ping can beat this, since DOMContentLoaded runs before the
            // load settles. Only start the watchdog if it has not.
            if (!attached) {
              attachTimer = setTimeout(onPromptInert, SESSION_CONFIRMATION_ATTACH_TIMEOUT_MS)
            }
          })
          .catch((error: unknown) => {
            dismissedBecause = `the prompt document could not be loaded (${
              error instanceof Error ? error.message : 'unknown load error'
            })`
            finish(false)
          })
      } catch (error) {
        dismissedBecause = `the prompt could not be set up (${
          error instanceof Error ? error.message : 'unknown display error'
        })`
        console.error(
          'Unable to display session confirmation:',
          error instanceof Error ? error.message : 'Unknown display error'
        )
        cancel()
      }
    })
  } finally {
    confirmationActive = false
  }
}
