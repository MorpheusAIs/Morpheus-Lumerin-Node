import { Configuration } from 'electron-builder'

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
    '!services/*'
  ],
  asarUnpack: ['resources/**', 'pkg-scripts/**'],
  win: {
    executableName: 'morpheus-ui',
    target: ['portable']
  },
  portable: {
    artifactName: '${name}-${version}-${arch}-${os}.${ext}'
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
    artifactName: '${name}-${version}-${arch}-${os}.${ext}'
  },
  dmg: {
    format: 'UDZO',
    contents: [
      {
        x: 50,
        y: 220,
        type: 'file'
      },
      {
        x: 200,
        y: 220,
        type: 'link',
        path: '/Applications'
      },
      {
        x: 410,
        y: 220,
        type: 'file',
        path: './pkg-scripts/postinstall',
        name: 'EnablePermissions'
      }
    ]
  },
  pkg: {
    artifactName: '${name}-${version}-${arch}-${os}.${ext}',
    installLocation: '/Applications',
    isRelocatable: false
    // allowAnywhere: false,
    // allowCurrentUserHome: false,
    // allowRootDirectory: false,
    // scripts: './pkg-scripts'
  },
  linux: {
    target: ['AppImage'],
    maintainer: 'mor.org',
    category: 'Utility',
    executableName: 'MorpheusUI',
    artifactName: '${name}-${version}-${arch}-${os}.${ext}'
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
