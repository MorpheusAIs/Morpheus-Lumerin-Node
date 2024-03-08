<h1>
  <img src="logo.svg" alt="Morpheus Desktop Application" width="20%">
</h1>

💻💰 Morpheus Application for desktop computers

<!--
[![JavaScript Style Guide](https://img.shields.io/badge/code_style-standard-brightgreen.svg)](https://standardjs.com)

![Lumerin Desktop Wallet](https://lumerin.io/images/lumerin-apps-demo@2x.png) 
-->

## Development

Create a local `.env` file with the content of .devSepoliaArbitrum.

You'll also need to supply a websocket endpoint to your web3 node.

### Requirements

- [Node.js](https://nodejs.org) LTS (v14 minimum recommended)

### Launch

```sh
# Install dependencies
npm i

# Run dev mode
npm run dev
```

For mac arm, you will need to run `npm run postinstallMacDist` after `npm i` to fix the electron-builder issue.

#### Troubleshooting

- For errors related to `node-gyp` when installing the dependencies, try using `sudo` to postinstall the dependencies.
- For Windows, installing `windows-build-tools` may be required. To do so, run:

```sh
npm i --global --production windows-build-tools
```

### Logs

The log output is in the next directories:

- **Linux:** `~/.config/<app name>/logs/{process-type}.log`
- **macOS:** `~/Library/Logs/<app name>/logs/{process-type}.log`
- **Windows:** `%USERPROFILE%\AppData\Roaming\<app name>\logs\{process-type}.log`

`process-type` being equal to `main`, `renderer` or `worker`

More info [github.com/megahertz/electron-log](https://github.com/megahertz/electron-log).

### Settings

- **Linux**: `~/.config/morpheus-ui-desktop/Settings`
- **macOS**: `~/Library/Application Support/morpheus-ui-desktop/Settings`
- **Windows**: `%APPDATA%\\morpheus-ui-desktop\\Settings`

To completely remove the application and start over, remove the settings file too.

## License

MIT
