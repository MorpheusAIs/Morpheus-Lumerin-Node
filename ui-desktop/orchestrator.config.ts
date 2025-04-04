import os from 'node:os'
import { OrchestratorConfig } from './src/main/orchestrator/orchestrator.types'
import { buildLocalModelsConfig } from './src/main/orchestrator/proxy-config'

const configMacArm = {
  proxyRouter: {
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_MAC_ARM64,
    fileName: './services/proxy-router/proxy-router' as string,
    runPath: './services/proxy-router/proxy-router' as string,
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
      LOG_COLOR: 'false'
    },
    modelsConfig: JSON.stringify(
      buildLocalModelsConfig(
        'tiny-llama-1.1B-chat',
        'openai',
        `http://localhost:${process.env.SERVICE_AI_API_PORT}/v1`
      )
    ),
    probe: {
      url: `http://localhost:${process.env.SERVICE_PROXY_API_PORT}/healthcheck`
    }
  },
  aiRuntime: {
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b4406/llama-b4406-bin-macos-arm64.zip' as string,
    fileName: './services/llama.zip' as string,
    extractPath: './services/ai-runtime',
    runPath: './services/ai-runtime/build/bin/llama-server' as string,
    runArgs: [
      '--no-webui',
      '--model',
      '../../../ai-model.gguf',
      '--port',
      `${process.env.SERVICE_AI_API_PORT}`
    ] as string[],
    probe: {
      url: `http://127.0.0.1:${process.env.SERVICE_AI_API_PORT}/health`
    }
  },
  aiModel: {
    downloadUrl:
      'https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q2_K.gguf',
    fileName: './services/ai-model.gguf'
  },
  ipfs: {
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.0/kubo_v0.34.0_darwin-arm64.tar.gz' as string,
    fileName: './services/ipfs.tar.gz' as string,
    extractPath: './services/ipfs',
    runPath: './services/ipfs/kubo/ipfs' as string,
    runArgs: ['daemon', '--init', `--api=/ip4/127.0.0.1/tcp/${process.env.SERVICE_IPFS_API_PORT}`],
    probe: {
      url: `http://127.0.0.1:${process.env.SERVICE_IPFS_API_PORT}/api/v0/version`,
      method: 'POST'
    }
  }
} as const satisfies OrchestratorConfig

const configMacX64 = {
  ...configMacArm,
  proxyRouter: {
    ...configMacArm.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_MAC_X64
  },
  aiRuntime: {
    ...configMacArm.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b4406/llama-b4406-bin-macos-x64.zip'
  },
  ipfs: {
    ...configMacArm.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.0/kubo_v0.34.0_darwin-amd64.tar.gz'
  }
} as const satisfies OrchestratorConfig

const configLinux: typeof configMacArm = {
  proxyRouter: {
    ...configMacArm.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_LINUX_X64
  },
  aiRuntime: {
    ...configMacArm.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b4406/llama-b4406-bin-ubuntu-x64.zip'
  },
  aiModel: {
    ...configMacArm.aiModel
  },
  ipfs: {
    ...configMacArm.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.0/kubo_v0.34.0_linux-amd64.tar.gz'
  }
}

const configWin: typeof configMacArm = {
  proxyRouter: {
    ...configMacArm.proxyRouter,
    downloadUrl: process.env.SERVICE_PROXY_DOWNLOAD_URL_WINDOWS_X64,
    fileName: './services/proxy-router.exe' as string,
    runPath: './services/proxy-router.exe' as string
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
      'https://github.com/ipfs/kubo/releases/download/v0.34.0/kubo_v0.34.0_windows-amd64.zip',
    fileName: './services/ipfs.zip',
    runPath: './services/ipfs/kubo/ipfs.exe'
  }
}

const cfg = {
  darwin: {
    arm64: configMacArm,
    x64: configMacX64
  },
  linux: {
    x64: configLinux
  },
  win32: {
    x64: configWin
  }
}[os.platform()]?.[os.arch()]

if (!cfg) {
  throw new Error(`Unsupported platform: ${os.platform()} ${os.arch()}`)
}

export { cfg }
