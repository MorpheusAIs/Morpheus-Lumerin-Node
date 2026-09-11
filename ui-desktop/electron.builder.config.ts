import { Configuration, type AfterPackContext } from 'electron-builder'
import {
  restoreNativeModuleBinaries,
  stageNativeModuleBinaries,
  verifyPackagedNativeBinaries
} from './src/main/orchestrator/native-prebuilds'
import {
  prepareProxyRouterBundle,
  ProxyRouterBundleDirectoryName,
  verifyPackagedProxyRouterBundle
} from './src/main/orchestrator/proxy-router-bundle'

const config: Configuration = {
  appId: 'com.electron.morpheus-ui',
  productName: 'MorpheusUI',
  directories: {
    buildResources: 'buildResources'
  },
  files: [
    '!**/.vscode/*',
    '!src/*',
    '!electron.vite.config.{js,ts,mjs,cjs}',
    '!{.eslintignore,.eslintrc.cjs,.prettierignore,.prettierrc.yaml,dev-app-update.yml,CHANGELOG.md,README.md}',
    '!{.env,.env.*,.npmrc,pnpm-lock.yaml}',
    '!{tsconfig.json,tsconfig.node.json,tsconfig.web.json}',
    '!services/*',
    '!scripts/ui-preview/**',
    // Keep previous installers out when a test build uses a nested output folder.
    '!dist/**'
  ],
  // `read_document` reaches these two through a runtime `import()`, which
  // Electron's asar layer cannot resolve. Unpacking them puts the files on
  // disk where the ordinary Node resolver can load them.
  asarUnpack: [
    'resources/**',
    'pkg-scripts/**',
    'node_modules/pdfjs-dist/**',
    'node_modules/mammoth/**',
    // Named explicitly so the packaged native binary sits at a known path and
    // can be checked against the target platform before the build succeeds.
    'node_modules/keytar/**'
  ],
  extraResources: [
    {
      from: 'buildResources/.generated/proxy-router/${os}-${arch}',
      to: ProxyRouterBundleDirectoryName,
      filter: ['bundled-proxy-router', 'manifest.json']
    }
  ],
  beforePack: async (context: AfterPackContext) => {
    await prepareProxyRouterBundle(context)
    await stageNativeModuleBinaries(context)
  },
  afterPack: async (context: AfterPackContext) => {
    // Restore first, unconditionally: a failed verification must still leave the
    // checkout with binaries this machine can run.
    await restoreNativeModuleBinaries(context)
    await verifyPackagedProxyRouterBundle(context)
    await verifyPackagedNativeBinaries(context)
  },
  win: {
    executableName: 'morpheus-ui',
    target: ['portable']
  },
  portable: {
    artifactName: '${os}-${arch}-${name}-${version}.${ext}'
  },
  mac: {
    executableName: 'MorpheusUI',
    entitlements: 'buildResources/entitlements.mac.plist',
    extendInfo: {
      NSCameraUsageDescription: "Application requests access to the device's camera.",
      NSMicrophoneUsageDescription: "Application requests access to the device's microphone.",
      NSDocumentsFolderUsageDescription:
        "Application requests access to the user's Documents folder.",
      NSDownloadsFolderUsageDescription:
        "Application requests access to the user's Downloads folder."
    },
    target: ['dmg'],
    notarize: false,
    artifactName: '${os}-${arch}-${name}-${version}.${ext}'
  },
  linux: {
    target: ['AppImage'],
    maintainer: 'mor.org',
    category: 'Utility',
    executableName: 'MorpheusUI',
    artifactName: '${os}-${arch}-${name}-${version}.${ext}'
  },
  npmRebuild: false,
  publish: {
    provider: 'generic',
    url: 'https://example.com/auto-updates'
  },
  electronDownload: {
    mirror: 'https://npmmirror.com/mirrors/electron/'
  }
}

export default config
