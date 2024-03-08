'use strict';

const { app, shell, Menu } = require('electron');
const APP_NAME = 'Morpheus';

const template = [
  {
    label: 'Edit',
    submenu: [
      { role: 'undo' },
      { role: 'redo' },
      { type: 'separator' },
      { role: 'cut' },
      { role: 'copy' },
      { role: 'paste' },
      { role: 'selectall' }
    ]
  },

  {
    label: 'View',
    submenu: [
      {
        label: 'Toggle Full Screen',
        accelerator: (() => process.platform === 'darwin' ? 'Ctrl+Command+F' : 'F11')(),
        click (item, focusedWindow) {
          focusedWindow && focusedWindow.setFullScreen(!focusedWindow.isFullScreen());
        }
      },

      { type: 'separator' },
      { label: 'Reload', accelerator: 'CmdOrCtrl+R', role: 'reload' },

      {
        label: 'Force Reload',
        accelerator: 'Shift+CmdOrCtrl+R',
        role: 'forcereload'
      },

      {
        label: 'Toggle Developer Tools',
        accelerator: 'Alt+CmdOrCtrl+I',
        role: 'toggledevtools'
      }
    ]
  },

  {
    role: 'window',
    submenu: [{ role: 'minimize' }, { role: 'close' }]
  },

  {
    role: 'help',
    submenu: [
      {
        label: 'Learn More',
        click () {
          return shell.openExternal('https://lumerin.io')
        }
      }
    ]
  }
];

if (process.platform === 'darwin') {
  template.unshift({
    label: APP_NAME,
    submenu: [
      { role: 'about', label: `About ${APP_NAME}` },
      { type: 'separator' },
      { role: 'services' },
      { type: 'separator' },
      { role: 'hide', label: `Hide ${APP_NAME}` },
      { role: 'hideothers' },
      { role: 'unhide' },
      { type: 'separator' },
      { role: 'quit', label: `Quit ${APP_NAME}` }
    ]
  });

  const windowMenu = template.find(m => m.role === 'window');

  if (windowMenu) {
    windowMenu.submenu.push({ type: 'separator' }, { role: 'front' });
  }
} else {
  template.unshift({
    label: 'File',
    submenu: [{ role: 'quit', label: `Quit ${APP_NAME}` }]
  });

  const help = template.find(t => t.role === 'help');

  if (help) {
    help.submenu.unshift({ type: 'separator' });
    help.submenu.unshift({ role: 'about', label: `${APP_NAME} v${app.getVersion()}`, enabled: false });
  }
}

module.exports = function () {
  Menu.setApplicationMenu(Menu.buildFromTemplate(template));
};
