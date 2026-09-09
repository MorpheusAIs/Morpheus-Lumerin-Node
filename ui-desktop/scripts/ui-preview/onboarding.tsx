import React, { useState } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import { createStore } from 'redux'
import { ThemeProvider } from 'styled-components'
import 'bootstrap/dist/css/bootstrap.min.css'
import '../../src/renderer/src/ui/interface.css'
import './preview.css'
import theme from '../../src/renderer/src/ui/theme'
import { Root } from '../../src/renderer/src/components/common/Root'
import Login from '../../src/renderer/src/components/Login'
import Onboarding from '../../src/renderer/src/components/onboarding/Onboarding'
import { Provider as ClientProvider } from '../../src/renderer/src/store/hocs/clientContext'
import { ToastsContext } from '../../src/renderer/src/components/toasts'

// Synthetic data only. No App bootstrap, preload, real settings, router, keys,
// wallet files or network API. Never enter real secrets in this harness.
const phrase =
  'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about'
const address = '0x1111111111111111111111111111111111111111'
const params = new URLSearchParams(location.search)
let failuresRemaining = params.has('fail-setup') ? 1 : 0
const config = { chain: { localProxyRouterUrl: 'http://127.0.0.1:1' } }
const store = createStore((state = { config }) => state)
const client = {
  onInit: async () => ({ onboardingComplete: params.has('existing'), persistedState: {}, config }),
  getDefaultCurrencySetting: async () => 'MOR',
  createMnemonic: async () => phrase,
  isValidMnemonic: (value) => value === phrase,
  suggestAddresses: async () => [address],
  onTermsLinkClick: () => undefined,
  onOnboardingCompleted: async () => {
    if (failuresRemaining-- > 0) throw new Error('Simulated connection failure. Try again.')
  },
  onLoginSubmit: async () =>
    params.has('recover') ? { requiresOnboarding: true } : { address, isActive: true },
  logout: async () => {
    throw new Error('Wallet erasure is disabled in this preview.')
  }
}
window.fetch = async () => {
  throw new Error('Application network disabled in preview')
}
XMLHttpRequest.prototype.open = function () {
  throw new Error('Application network disabled in preview')
}

function Fixture() {
  const [isSessionActive, setSessionActive] = useState(false)
  const [notice, setNotice] = useState('')
  return (
    <ThemeProvider theme={theme}>
      <Provider store={store}>
        <ClientProvider value={client}>
          <ToastsContext.Provider value={{ toast: (_kind, message) => setNotice(String(message)) }}>
            <aside
              style={{ position: 'fixed', top: 0, zIndex: 1, padding: '8px 16px', fontSize: 12 }}
            >
              Synthetic wallet setup preview. Do not enter real secrets. {notice}
            </aside>
            <Root
              config={config}
              client={client}
              dispatch={(action) => {
                if (action.type === 'session-started') setSessionActive(true)
              }}
              isSessionActive={isSessionActive}
              isAuthBypassed={false}
              sellerDefaultCurrency="MOR"
              servicesState={{ orchestratorStatus: 'ready' } as any}
              StartupComponent={() => <p>Preparing preview</p>}
              OnboardingComponent={Onboarding as any}
              LoginComponent={Login as any}
              RouterComponent={() => <h1>Synthetic wallet setup completed</h1>}
            />
          </ToastsContext.Provider>
        </ClientProvider>
      </Provider>
    </ThemeProvider>
  )
}
createRoot(document.getElementById('root')!).render(<Fixture />)
