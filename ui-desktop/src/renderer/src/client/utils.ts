import * as utils from 'web3-utils';
import cuid from 'cuid';
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

export function forwardToMainProcess<T>(eventName: string, timeout = 10000) {
  return function (data: T) {
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
  data: T,
  timeout = 10000,
  ipcRenderer = window.ipcRenderer,
): Promise<K> {
  const id = cuid();

  const deferred = new Deferred();
  let timeoutId;

  function listener(_, { id: _id, data: _data, error }, unsubscribe) {
    if (timeoutId) {
      window.clearTimeout(timeoutId);
    }
    if (_id !== id) {
      return;
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
