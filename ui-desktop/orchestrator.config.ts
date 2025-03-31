import os from 'node:os'
import { OrchestratorConfig } from './src/main/orchestrator/orchestrator.types'

const configMac = {
  proxyRouter: {
    downloadUrl: null,
    fileName: './services/proxy-router' as string,
    runPath: './services/proxy-router' as string,
    runArgs: [
      `--diamond-address=0xDE819AaEE474626E3f34Ef0263373357e5a6C71b`,
      `--mor-token-address=0x092bAaDB7DEf4C3981454dD9c0A0D7FF07bCFc86`,
      `--blockscout-api-url=https://arbitrum.blockscout.com/api/v2`,
      `--eth-node-chain-id=42161`,
      `--environment=production`,
      `--auth-config-file-path=./proxy.conf`,
      `--cookie-content=admin:admin`,
      `--cookie-file-path=./.cookie`,
      `--proxy-address=0.0.0.0:3333`,
      `--web-address=0.0.0.0:8082`,
      `--web-public-url=http://localhost:8082`
    ],
    probe: {
      url: 'http://localhost:8082/healthcheck'
    }
  },
  aiRuntime: {
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b4406/llama-b4406-bin-macos-arm64.zip' as string,
    fileName: './services/llama.zip' as string,
    extractPath: './services/ai-runtime',
    runPath: './services/ai-runtime/build/bin/llama-server' as string,
    runArgs: ['--no-webui', '--model', '../../../ai-model.gguf', '--port', '3434'],
    probe: {
      url: 'http://127.0.0.1:3434/health'
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
    runArgs: ['daemon', '--init', '--api=/ip4/127.0.0.1/tcp/5001'],
    probe: {
      url: 'http://127.0.0.1:5001/api/v0/version',
      method: 'POST'
    }
  }
} as const satisfies OrchestratorConfig

const configLinux: typeof configMac = {
  proxyRouter: {
    ...configMac.proxyRouter
  },
  aiRuntime: {
    ...configMac.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b4406/llama-b4406-bin-ubuntu-x64.zip'
  },
  aiModel: {
    ...configMac.aiModel
  },
  ipfs: {
    ...configMac.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.0/kubo_v0.34.0_linux-amd64.tar.gz'
  }
}

const configWin: typeof configMac = {
  proxyRouter: {
    ...configMac.proxyRouter,
    runPath: './proxy-router.exe'
  },
  aiRuntime: {
    ...configMac.aiRuntime,
    downloadUrl:
      'https://github.com/ggml-org/llama.cpp/releases/download/b5002/llama-b5002-bin-win-avx2-x64.zip',
    runPath: './services/ai-runtime/llama-server.exe'
  },
  aiModel: {
    ...configMac.aiModel
  },
  ipfs: {
    ...configMac.ipfs,
    downloadUrl:
      'https://github.com/ipfs/kubo/releases/download/v0.34.0/kubo_v0.34.0_windows-amd64.zip',
    fileName: './services/ipfs.zip',
    runPath: './services/ipfs/kubo/ipfs.exe'
  }
}

const cfg = {
  darwin: {
    arm64: configMac
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
