// Browser stub for Electron IPC - allows renderer to run as a web app without Electron.
// Original Electron logic is untouched; this stub is only installed when window.ipcRenderer
// is not already provided (i.e. running in a plain browser).

if (typeof window !== 'undefined' && !(window as any).__electronIPCInstalled) {
  (window as any).__electronIPCInstalled = true;

  // Listener registry: eventName -> array of { listener, unsubscribe }
  const registry: Record<string, Array<{ listener: Function; unsubscribe: () => void }>> = {};

  /** Dispatch an event to all registered listeners for a given channel */
  function fireChannel(eventName: string, payload: unknown) {
    const entries = registry[eventName] || [];
    // iterate over a copy so unsubscribes inside handlers are safe
    [...entries].forEach(({ listener, unsubscribe }) => {
      try {
        listener(null, payload, unsubscribe);
      } catch (err) {
        console.warn('[IPC stub] listener error on', eventName, err);
      }
    });
  }

  /** Mock data returned for each IPC call from renderer → main */
  const WEB_CONFIG = {
    chain: {
      bypassAuth: true,
      chainId: 8453,
      defaultSellerCurrency: 'BTC',
      diamondAddress: '0x6aBE1d282f72B474E54527D93b979A4f64d3030a',
      displayName: 'Base',
      explorerUrl: 'https://basescan.org/tx/{{hash}}',
      localProxyRouterUrl: 'http://localhost:8082',
      mainTokenAddress: '0x7431ada8a591c955a994a21710752ef9b882b8e3',
      symbol: 'MOR',
      symbolEth: 'ETH',
    },
    statePersistanceDebounce: 2000,
    sentryDsn: null,
  };

  const MOCK_RESPONSES: Record<string, (data?: any) => any> = {
    // Startup / lifecycle
    'ui-ready': () => ({ onboardingComplete: true, persistedState: {}, config: WEB_CONFIG }),
    'ui-unload': () => null,
    'onboarding-completed': () => null,
    'login-submit': () => null,
    'logout': () => null,
    'quit-app': () => null,
    'clear-cache': () => null,
    'persist-state': () => null,
    'handle-client-error': () => null,

    // Settings / config
    'get-default-currency-settings': () => 'BTC',
    'set-default-currency-settings': () => null,
    'get-custom-env-values': () => ({}),
    'set-custom-env-values': () => null,
    'get-proxy-router-settings': () => ({}),
    'save-proxy-router-settings': () => null,
    'get-profit-settings': () => ({}),
    'set-profit-settings': () => null,
    'get-auto-adjust-price': () => ({}),
    'set-auto-adjust-price': () => null,
    'get-contract-hashrate': () => 0,
    'get-marketplace-fee': () => '0',
    'get-local-ip': () => '127.0.0.1',
    'get-pool-address': () => '',
    'get-token-gas-limit': () => '21000',
    'get-lmr-transfer-gas-limit': () => '21000',
    'open-select-folder-dialog': () => null,

    // Auth
    'validate-password': () => true,
    'change-password': () => null,
    'get-private-key': () => '',
    'suggest-addresses': () => [],

    // Wallet / transactions
    'get-transactions': () => [],
    'refresh-all-transactions': () => [],
    'refresh-all-contracts': () => [],
    'get-past-transactions': () => [],
    'send-lmr': () => null,
    'send-eth': () => null,

    // Balances / rates
    'get-balances': () => ({ eth: '0', mor: '0' }),
    'get-rates': () => ({ ETH: 0, MOR: 0 }),
    'get-todays-budget': () => 0,
    'get-supply': () => '0',

    // Models / chat
    'get-all-models': () => [],
    'get-auth-headers': () => ({ 'Content-Type': 'application/json' }),
    'get-chat-history-titles': () => [],
    'get-chat-history': () => ({ messages: [], title: '', modelId: '', sessionId: '' }),
    'delete-chat-history': () => null,
    'update-chat-history-title': () => null,

    // Failover
    'get-failover-setting': () => ({ isEnabled: false }),
    'set-failover-setting': () => null,
    'check-provider-connectivity': () => false,

    // IPFS
    'get-ipfs-version': () => '0.0.0',
    'get-ipfs-file': () => null,
    'pin-ipfs-file': () => null,
    'unpin-ipfs-file': () => null,
    'add-file-to-ipfs': () => null,
    'get-ipfs-pinned-files': () => [],

    // Agents
    'get-agent-users': () => ({ agents: [] }),
    'confirm-decline-agent-user': () => null,
    'remove-agent-user': () => null,
    'get-agent-txs': () => ({ txHashes: [], nextCursor: '' }),
    'revoke-agent-allowance': () => null,
    'get-agent-allowance-requests': () => ({ requests: [] }),
    'confirm-decline-agent-allowance-request': () => null,

    // Services (startup orchestrator)
    'start-services': () => null,
    'restart-service': () => null,
    'ping-service': () => true,
  };

  const ipcRendererStub = {
    send(eventName: string, payload: any) {
      const { id, data } = payload || {};
      const mockFn = MOCK_RESPONSES[eventName];
      if (mockFn !== undefined) {
        const responseData = mockFn(data);
        // Respond asynchronously (next microtask) so listeners are always registered first
        Promise.resolve().then(() => {
          fireChannel(eventName, { id, data: responseData });
        });
      } else {
        console.debug('[IPC stub] unhandled send:', eventName, payload);
        Promise.resolve().then(() => {
          fireChannel(eventName, { id, data: null });
        });
      }
    },

    on(eventName: string, listener: Function) {
      if (!registry[eventName]) registry[eventName] = [];
      const entry: { listener: Function; unsubscribe: () => void } = {
        listener,
        unsubscribe: () => {
          const arr = registry[eventName];
          if (arr) {
            const idx = arr.indexOf(entry);
            if (idx !== -1) arr.splice(idx, 1);
          }
        },
      };
      registry[eventName].push(entry);
      return entry.unsubscribe;
    },
  };

  // Install globals expected by the renderer
  (window as any).ipcRenderer = ipcRendererStub;
  (window as any).openLink = (url: string) => window.open(url, '_blank');
  (window as any).getAppVersion = () => '1.0.0-web';
  (window as any).copyToClipboard = (text: string) => {
    navigator.clipboard?.writeText(text).catch(() => {});
  };
  (window as any).isDev = true;
  (window as any).electron = {
    process: {
      versions: {
        electron: '28.0.0',
        chrome: '120.0.0',
        node: '20.0.0',
      },
    },
  };
  (window as any).api = {};

  // Drive the startup flow by firing IPC events once the Redux store and
  // subscriptions are wired up. This sidesteps the React prop timing issue
  // (this.props.isAuthBypassed hasn't updated yet when Root checks it in
  // componentDidMount's .then() chain, so we push session-started ourselves).
  //
  //  100ms → services-state          → startupComplete = true
  //  200ms → session-started         → isSessionActive = true (skips LoginComponent)
  //  350ms → required-data-gathered  → hasEnoughData   = true → RouterComponent
  setTimeout(() => {
    fireChannel('services-state', {
      orchestratorStatus: 'ready',
      download: [],
      startup: [],
    });
  }, 100);

  setTimeout(() => {
    fireChannel('session-started', {});
  }, 200);

  setTimeout(() => {
    fireChannel('required-data-gathered', {});
  }, 350);
}
