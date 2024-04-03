import Raven from 'raven-js';
import get from 'lodash/get';

window.ipcRenderer.on('ui-ready', (ev, payload) => {
  const sentryDsn = get(payload, 'data.config.sentryDsn', null);
  if (sentryDsn) {
    try {
      Raven.config(sentryDsn, { release: window.getAppVersion() }).install();
    } catch (err) {
      // eslint-disable-next-line no-console
      console.warn(`Error while configuring Raven with Sentry. ${err.message}`);
    }
    window.addEventListener('unhandledrejection', err =>
      Raven.captureException(err.reason)
    );
  }
});
