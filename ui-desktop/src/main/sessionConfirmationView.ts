import interfaceStyles from '../renderer/src/ui/interface.css?raw'
import confirmationStyles from './sessionConfirmation.css?raw'

export interface SessionConfirmationDetails {
  readonly modelId: string
  readonly duration: number
  readonly directPayment: boolean
  readonly failover: boolean
}

const escapeHtml = (value: string): string =>
  value.replace(
    /[&<>"']/g,
    (character) =>
      ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[character]!
  )

/** Only main-validated transaction fields enter this isolated confirmation view. */
export function renderSessionConfirmationHtml(
  details: SessionConfirmationDetails,
  nonce: string
): string {
  const safeNonce = escapeHtml(nonce)
  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'none'; style-src 'nonce-${safeNonce}'; connect-src 'none'; img-src 'none'; frame-src 'none'; object-src 'none'; base-uri 'none'; form-action 'none'">
  <title>Open session · Morpheus</title>
  <style nonce="${safeNonce}">${interfaceStyles}\n${confirmationStyles}</style>
</head>
<body>
  <section class="session-confirmation" role="dialog" aria-modal="true" aria-labelledby="session-title" aria-describedby="session-description">
    <header class="session-confirmation__header">
      <h1 id="session-title">Open session</h1>
      <button class="session-confirmation__close" type="button" data-decision="cancel" aria-label="Cancel session opening">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18"/></svg>
      </button>
    </header>
    <div class="session-confirmation__body">
      <p id="session-description">Review the session details before submitting a blockchain transaction.</p>
      <dl class="session-confirmation__details">
        <div class="session-confirmation__model"><dt>Model ID</dt><dd><code>${escapeHtml(details.modelId)}</code></dd></div>
        <div><dt>Contract duration input</dt><dd>${escapeHtml(details.duration.toLocaleString('en-US'))} seconds</dd></div>
        <div><dt>Payment method</dt><dd>${details.directPayment ? 'Direct MOR payment' : 'Stake MOR · escrow'}</dd></div>
        <div><dt>Provider failover</dt><dd>${details.failover ? 'Enabled' : 'Disabled'}</dd></div>
      </dl>
      <p class="session-confirmation__payment">${
        details.directPayment
          ? 'This session uses direct MOR payment. Confirm only if you want to proceed with this payment mode.'
          : 'Your MOR is escrowed for the session. Unused stake returns when the session closes.'
      }</p>
      <p class="session-confirmation__note">Nothing is submitted until you choose Open session. This confirmation expires after 60 seconds.</p>
    </div>
    <footer class="session-confirmation__actions">
      <button id="session-cancel" class="session-confirmation__cancel" type="button" data-decision="cancel" autofocus>Cancel</button>
      <button class="session-confirmation__confirm" type="button" data-decision="approve">
        <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14m-6-6 6 6-6 6"/></svg>
        <span>Open session</span>
      </button>
    </footer>
  </section>
</body>
</html>`
}
