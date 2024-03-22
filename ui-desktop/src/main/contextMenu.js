'use strict'

import electronContextMenu from 'electron-context-menu'

// TODO pass i18n labels for context menu items
// @see https://github.com/sindresorhus/electron-context-menu#api
const config = {}

export default function () {
  electronContextMenu(config)
}
