const { app, BrowserWindow, Notification, dialog } = require("electron");
const { autoUpdater } = require("electron-updater");
const isDev = require("electron-is-dev");
const path = require("path");
const windowStateKeeper = require("electron-window-state");

const logger = require("../logger");
const analytics = require("../analytics");
const restart = require("./client/electron-restart");

let mainWindow;

// Disable electron security warnings since local content is served via http
if (isDev) {
  process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = true;
}

function showUpdateNotification(info = {}) {
  if (!Notification.isSupported()) {
    return;
  }

  const versionLabel = info.label
    ? `Version ${info.version}`
    : "The latest version";

  const notification = new Notification({
    title: `${versionLabel} was installed`,
    body: "Morpheus will be automatically updated after restart.",
  });

  notification.show();
}

function initAutoUpdate() {
  if (isDev) {
    return;
  }
  autoUpdater.on("checking-for-update", () =>
    logger.info("Checking for update...")
  );
  autoUpdater.on("update-available", () => logger.info("Update available."));
  autoUpdater.on("download-progress", function(progressObj) {
    let msg = `Download speed: ${progressObj.bytesPerSecond}`;
    msg += ` - Downloaded ${progressObj.percent}%`;
    msg += ` (${progressObj.transferred}/${progressObj.total})`;
    logger.info(msg);
  });

  autoUpdater.on("update-not-available", () =>
    logger.info("Update not available.")
  );
  autoUpdater.on("error", (err) =>
    logger.error(`Error in auto-updater. ${err}`)
  );

  autoUpdater
    .checkForUpdatesAndNotify()
    .then((res) => {
      logger.info(`Checked for the updates: ${res}`);
    })
    .catch(function(err) {
      logger.warn("Could not find updates", err.message);
    });
}

function loadWindow(config) {
  // Ensure the app is ready before creating the main window
  if (!app.isReady()) {
    logger.warn("Tried to load main window while app not ready. Reloading...");
    restart(1);
    return;
  }

  if (mainWindow) {
    return;
  }

  const mainWindowState = windowStateKeeper({
    // defaultWidth: 660,
    defaultWidth: 1200,
    defaultHeight: 800,
  });

  // TODO this should be read from config
  mainWindow = new BrowserWindow({
    show: false,
    width: mainWindowState.width,
    height: mainWindowState.height,
    // maxWidth: 660,
    // maxHeight: 700,
    minWidth: 660,
    minHeight: 800,
    backgroundColor: "#323232",
    webPreferences: {
      enableRemoteModule: true,
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, "preload.js"),
      devTools: config.devTools || !app.isPackaged,
    },
    x: mainWindowState.x,
    y: mainWindowState.y,
  });
  mainWindow.center();

  require("@electron/remote/main").enable(mainWindow.webContents);

  mainWindowState.manage(mainWindow);

  analytics.init(mainWindow.webContents.getUserAgent());

  const appUrl = isDev
    ? process.env.ELECTRON_START_URL
    : `file://${path.join(__dirname, "../index.html")}`;

  logger.info("Roading renderer from URL:", appUrl);

  mainWindow.loadURL(appUrl);

  mainWindow.webContents.on("crashed", function(ev, killed) {
    logger.error("Crashed", ev.sender.id, killed);
  });

  mainWindow.on("unresponsive", function(ev) {
    logger.error("Unresponsive", ev.sender.id);
  });

  mainWindow.on("closed", function() {
    mainWindow = null;
  });

  mainWindow.once("ready-to-show", function() {
    initAutoUpdate();
    mainWindow.show();
  });

  mainWindow.on("close", (event) => {
    event.preventDefault();
    if (app.quitting || process.platform !== "darwin") {
      const choice = dialog.showMessageBoxSync(mainWindow, {
        type: "question",
        buttons: ["Yes", "No"],
        title: "Confirm",
        message: "Are you sure you want to quit?",
      });
      if (choice === 1) {
        return;
      } else {
        mainWindow.destroy();
        mainWindow = null;
        app.quit();
      }
    } else {
      mainWindow.hide();
    }
  });

  app.on("activate", () => {
    mainWindow.show();
  });

  app.on("before-quit", () => {
    app.quitting = true;
  });
}

function createWindow(config) {
  app.on("fullscreen", function() {
    mainWindow.isFullScreenable
      ? mainWindow.setFullScreen(true)
      : mainWindow.setFullScreen(false);
  });

  const load = loadWindow.bind(null, config);

  app.on("ready", load);
  app.on("activate", load);
}

module.exports = { createWindow };
