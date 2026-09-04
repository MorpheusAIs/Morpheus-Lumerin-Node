import * as utils from 'web3-utils';
import { createId as cuid } from '@paralleldrive/cuid2';
import Deferred from '../lib/Deferred';
import BN from 'bn.js';

export const fromWei = (str: string, unit: utils.Unit = 'ether') =>
  utils.fromWei(str, unit);
export const toWei = (bn: BN, unit: utils.Unit = 'ether') =>
  utils.toWei(bn, unit);

export const isAddress = (str: string): str is `0x${string}` =>
  utils.isAddress(str);

export const toBN = (str: string) => utils.toBN(str);
export const toHex = (bn: BN) => utils.toHex(bn);

export function forwardToMainProcess<T>(eventName: string, timeout: number | null = 10000) {
  return function (data?: T) {
    return sendToMainProcess<T>(eventName, data, timeout);
  };
}

/**
 * Sends a message to Main Process and returns a Promise.
 *
 * This makes it easier to handle IPC inside components
 * without the need of manual (un)subscriptions.
 */
export function sendToMainProcess<T = any, K = unknown>(
  eventName: string,
  data?: T,
  timeout: number | null = 10000,
  ipcRenderer = window.ipcRenderer,
): Promise<K> {
  const id = cuid();

  const deferred = new Deferred();
  let timeoutId;

  function listener({ id: _id, data: _data, error }, unsubscribe) {
    // IMPORTANT: check the correlation id BEFORE cancelling our timeout.
    //
    // Every in-flight call on the same channel registers its own listener, and
    // each listener sees *every* response on that channel. The previous version
    // cleared the timeout first and only then compared ids — so a response
    // belonging to request A would cancel request B's timeout and then bail
    // out, leaving B with no timer and no resolution. B's promise hung forever.
    // That is why rapid navigation (which fires many overlapping IPC calls)
    // left the UI with permanently dead buttons.
    if (_id !== id) {
      return;
    }

    if (timeoutId) {
      window.clearTimeout(timeoutId);
    }

    const responseError = error || (_data && _data.error);

    if (responseError) {
      deferred.reject(responseError);
      ipcRenderer.send('handle-client-error', {
        id: cuid(),
        data: responseError,
      });
    } else {
      deferred.resolve(_data);
    }

    return unsubscribe();
  }

  const unsubscribe = ipcRenderer.on(eventName, listener);
  ipcRenderer.send(eventName, { id, data });

  if (timeout) {
    timeoutId = setTimeout(() => {
      console.warn(`Event "${eventName}" timed out after ${timeout}ms.`);
      deferred.reject(
        new Error('Operation timed out. Please try again later.'),
      );
      unsubscribe();
    }, timeout);
  }

  return deferred.promise as Promise<K>;
}
