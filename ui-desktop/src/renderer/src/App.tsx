import React from 'react'
import Modal from 'react-modal'
import ReactDOM from 'react-dom'
import { ThemeProvider } from 'styled-components'

import theme from './ui/theme'
import Root from './components/common/Root'
import { Provider as ClientProvider } from './store/hocs/clientContext'
import { Provider, createStore } from './store/store'

import createClient from './client'
import { subscribeToMainProcessMessages } from './subscriptions'

import Web3ConnectionNotifier from './components/Web3ConnectionNotifier'
import { ToastsProvider } from './components/toasts'
import { GlobalTooltips } from './components/common'
import Onboarding from './components/onboarding/Onboarding'
import Loading from './components/Loading'
import Router from './components/Router'
import Login from './components/Login'

const client = createClient(createStore)

// Initialize all the Main Process subscriptions
subscribeToMainProcessMessages(client.store)

function App(): JSX.Element {
  const ipcHandle = (): void => window.electron.ipcRenderer.send('ping')

  return (
    <>
      <ClientProvider value={client}>
        <Provider store={client.store}>
          <ThemeProvider theme={theme}>
            <ToastsProvider>
              <Root
                OnboardingComponent={Onboarding}
                LoadingComponent={Loading}
                RouterComponent={Router}
                LoginComponent={Login}
              />
              <GlobalTooltips />
              <Web3ConnectionNotifier />
            </ToastsProvider>
          </ThemeProvider>
        </Provider>
      </ClientProvider>
    </>
  )
}

export default App
