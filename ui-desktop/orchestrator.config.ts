import os from 'node:os'
import { OrchestratorConfig } from './src/main/orchestrator/orchestrator.types'
import {
  buildLocalModelsConfig,
  buildLocalRatingConfig
} from './src/main/orchestrator/proxy-config'

const configMacArm = {
  proxyRouter: {
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_MAC_ARM64,
    fileName: './services/proxy-router/proxy-router' as string,
    runPath: './services/proxy-router/proxy-router' as string,
    ports: [process.env.SERVICE_PROXY_PORT, process.env.SERVICE_PROXY_API_PORT],
    env: {
      DIAMOND_CONTRACT_ADDRESS: process.env.DIAMOND_ADDRESS,
      MOR_TOKEN_ADDRESS: process.env.TOKEN_ADDRESS,
      BLOCKSCOUT_API_URL: process.env.BLOCKSCOUT_API_URL,
      ETH_NODE_CHAIN_ID: String(process.env.CHAIN_ID),
      ENVIRONMENT: process.env.NODE_ENV,
      AUTH_CONFIG_FILE_PATH: './proxy.conf',
      COOKIE_FILE_PATH: './.cookie',
      PROXY_ADDRESS: `0.0.0.0:${process.env.SERVICE_PROXY_PORT}`,
      WEB_ADDRESS: `0.0.0.0:${process.env.SERVICE_PROXY_API_PORT}`,
      WEB_PUBLIC_URL: `http://localhost:${process.env.SERVICE_PROXY_API_PORT}`,
      MODELS_CONFIG_PATH: './models-config.json',
      RATING_CONFIG_PATH: './rating-config.json',
      ETH_NODE_USE_SUBSCRIPTIONS: 'false',
      ETH_NODE_ADDRESS: '',
      PROXY_STORE_CHAT_CONTEXT: 'true',
      PROXY_STORAGE_PATH: './data/',
      LOG_COLOR: 'false',
      LOG_FOLDER_PATH: './logs/',
      IPFS_MULTADDR: `/ip4/127.0.0.1/tcp/${process.env.SERVICE_IPFS_API_PORT}`,
      DOCKER_HOST: 'unix:///var/run/docker.sock' as string
    },
    modelsConfig: JSON.stringify(
      buildLocalModelsConfig(
        'tiny-llama-1.1B-chat',
        'openai',
        `http://localhost:${process.env.SERVICE_AI_API_PORT}/v1`
      )
    ),
    ratingConfig: JSON.stringify(buildLocalRatingConfig()),
    probe: {
      url: `http://localhost:${process.env.SERVICE_PROXY_API_PORT}/healthcheck`
    }
  },
  aiRuntime: {
    //original b4406
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b5002/llama-b5002-bin-macos-arm64.zip' as string,
    fileName: './services/llama.zip' as string,
    extractPath: './services/ai-runtime',
    runPath: './services/ai-runtime/build/bin/llama-server' as string,
    ports: [process.env.SERVICE_AI_API_PORT],
    runArgs: [
      '--no-webui',
      '--model',
      '../../../ai-model.gguf',
      '--port',
      `${process.env.SERVICE_AI_API_PORT}`,
      '--log-file',
      './llama.log'
    ] as string[],
    probe: {
      url: `http://127.0.0.1:${process.env.SERVICE_AI_API_PORT}/health`
    }
  },
  aiModel: {
    downloadUrl:
      'https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q2_K.gguf' as string,
    fileName: './services/ai-model.gguf' as string
  },
  ipfs: {
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.1/kubo_v0.34.1_darwin-arm64.tar.gz' as string,
    fileName: './services/ipfs.tar.gz' as string,
    extractPath: './services/ipfs',
    runPath: './services/ipfs/kubo/ipfs' as string,
    ports: [process.env.SERVICE_IPFS_API_PORT],
    runArgs: [
      'daemon',
      '--init',
      `--api=/ip4/127.0.0.1/tcp/${process.env.SERVICE_IPFS_API_PORT}`,
      `--repo-dir=../data`
    ],
    probe: {
      url: `http://127.0.0.1:${process.env.SERVICE_IPFS_API_PORT}/api/v0/version`,
      method: 'POST',
      timeout: 20000
    }
  },
  containerRuntime: {
    downloadUrl: 'https://desktop.docker.com/mac/main/arm64/Docker.dmg' as string,
    probe: {
      url: 'unix:///var/run/docker.sock:/version' as string
    }
  }
} as const satisfies OrchestratorConfig

const configMacX64 = {
  proxyRouter: {
    ...configMacArm.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_MAC_X64
  },
  aiRuntime: {
    ...configMacArm.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b4406/llama-b4406-bin-macos-x64.zip'
  },
  aiModel: {
    ...configMacArm.aiModel
  },
  ipfs: {
    ...configMacArm.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.1/kubo_v0.34.1_darwin-amd64.tar.gz'
  },
  containerRuntime: {
    ...configMacArm.containerRuntime,
    downloadUrl: 'https://desktop.docker.com/mac/main/amd64/Docker.dmg' as string
  }
} as const satisfies OrchestratorConfig

const configLinux: typeof configMacArm = {
  proxyRouter: {
    ...configMacArm.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_LINUX_X64
  },
  // original b4406
  aiRuntime: {
    ...configMacArm.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b5002/llama-b5002-bin-ubuntu-x64.zip'
  },
  aiModel: {
    ...configMacArm.aiModel
  },
  ipfs: {
    ...configMacArm.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.1/kubo_v0.34.1_linux-amd64.tar.gz'
  },
  containerRuntime: {
    ...configMacArm.containerRuntime,
    downloadUrl: 'https://desktop.docker.com/linux/main/amd64/docker-desktop-amd64.deb' as string
  }
}

const configLinuxArm: typeof configMacArm = {
  proxyRouter: {
    ...configMacArm.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_LINUX_ARM64
  },
  aiRuntime: {
    ...configMacArm.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b5002/llama-b5002-bin-ubuntu-arm64.zip'
  },
  aiModel: {
    ...configMacArm.aiModel
  },
  ipfs: {
    ...configMacArm.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.1/kubo_v0.34.1_linux-arm64.tar.gz'
  },
  containerRuntime: {
    ...configMacArm.containerRuntime,
    downloadUrl: 'https://docs.docker.com/desktop/setup/install/linux/' as string
  }
}

const configWin: typeof configMacArm = {
  proxyRouter: {
    ...configMacArm.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_X64,
    fileName: './services/proxy-router.exe' as string,
    runPath: './services/proxy-router.exe' as string,
    env: {
      ...configMacArm.proxyRouter.env,
      DOCKER_HOST: 'npipe:////./pipe/docker_engine'
    }
  },
  aiRuntime: {
    ...configMacArm.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b5002/llama-b5002-bin-win-avx2-x64.zip',
    runPath: './services/ai-runtime/llama-server.exe' as string,
    runArgs: [
      '--no-webui',
      '--model',
      '../ai-model.gguf',
      '--port',
      `${process.env.SERVICE_AI_API_PORT}`
    ]
  },
  aiModel: {
    ...configMacArm.aiModel
  },
  ipfs: {
    ...configMacArm.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.1/kubo_v0.34.1_windows-amd64.zip',
    fileName: './services/ipfs.zip',
    runPath: './services/ipfs/kubo/ipfs.exe'
  },
  containerRuntime: {
    probe: {
      url: 'npipe:////./pipe/docker_engine:/version'
    },
    downloadUrl:
      'https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe' as string
  }
}
// *********************************************************************************
// WARNING: LLAMA.CPP DOES NOT SUPPORT ARM64 for GGUF (found one for win-llvm-arm64 so need to change model as well...no idea if it works)
// *********************************************************************************
const configWinArm: typeof configMacArm = {
  proxyRouter: {
    ...configWin.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_ARM64,
    fileName: './services/proxy-router.exe' as string,
    runPath: './services/proxy-router.exe' as string
  },
  aiRuntime: {
    ...configMacArm.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b5002/llama-b5002-bin-win-llvm-arm64.zip',
    runPath: './services/ai-runtime/llama-server.exe' as string,
    runArgs: [
      '--no-webui',
      '--model',
      '../ai-model.llvm',
      '--port',
      `${process.env.SERVICE_AI_API_PORT}`
    ]
  },
  aiModel: {
    ...configMacArm.aiModel,
    downloadUrl:
      'https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q2_K.llvm',
    fileName: './services/ai-model.llvm'
  },
  ipfs: {
    ...configMacArm.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.1/kubo_v0.34.1_windows-arm64.zip',
    fileName: './services/ipfs.zip',
    runPath: './services/ipfs/kubo/ipfs.exe'
  },
  containerRuntime: {
    ...configWin.containerRuntime,
    downloadUrl:
      'https://desktop.docker.com/win/main/arm64/Docker%20Desktop%20Installer.exe' as string
  }
}

const cfg = {
  darwin: {
    x64: configMacX64,
    arm64: configMacArm
  },
  linux: {
    x64: configLinux,
    arm64: configLinuxArm
  },
  win32: {
    x64: configWin,
    arm64: configWinArm
  }
}[os.platform()]?.[os.arch()]

if (!cfg) {
  throw new Error(`Unsupported platform: ${os.platform()} ${os.arch()}`)
}

export { cfg }
