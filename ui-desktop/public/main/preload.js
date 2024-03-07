'use strict';

const { ipcRenderer, clipboard, shell, contextBridge } = require('electron');
const remote = require('@electron/remote');
// electron-is-dev can't be used in preload script
// const isDev = require('electron-is-dev')
const isDev = !remote.app.isPackaged;

// @see http://electronjs.org/docs/tutorial/security#2-disable-nodejs-integration-for-remote-content

const copyToClipboard = function (text) {
  return clipboard.writeText(text);
};

const getAppVersion = function () {
  return remote.app.getVersion();
};

const openLink = function (url) {
  return shell.openExternal(url);
};

contextBridge.exposeInMainWorld('ipcRenderer', {
  send (eventName, payload) {
    return ipcRenderer.send(eventName, payload);
  },
  on (eventName, listener) {
    // For some reason the listener passed into this function doesn't work
    // if you want to use it to unsubscribe later (likely due to chrome/node connection). 
    // So we wrap it in a function and provide an unsubscribe function both to event handler
    // and as a returned value
    function unsubscribe(){
      ipcRenderer.removeListener(eventName, subscription)
    }

    function subscription(event, payload) {
      listener(event, payload, unsubscribe);
    }

    ipcRenderer.on(eventName, subscription);

    return unsubscribe;
  }
});

contextBridge.exposeInMainWorld('openLink', openLink);
contextBridge.exposeInMainWorld('getAppVersion', getAppVersion);
contextBridge.exposeInMainWorld('copyToClipboard', copyToClipboard);
contextBridge.exposeInMainWorld('isDev', isDev);
