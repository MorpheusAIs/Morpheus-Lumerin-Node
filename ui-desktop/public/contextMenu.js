'use strict';

const electronContextMenu = require('electron-context-menu');

// TODO pass i18n labels for context menu items
// @see https://github.com/sindresorhus/electron-context-menu#api
const config = {};

module.exports = function () {
  electronContextMenu(config);
}
