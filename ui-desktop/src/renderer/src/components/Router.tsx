import { HashRouter, Routes, Route, Navigate } from "react-router";
import styled, { keyframes } from 'styled-components'

import OfflineWarning from './OfflineWarning'
// import ChangePassword from './ChangePassword'
import Dashboard from './dashboard/Dashboard'
import Sidebar from './sidebar/Sidebar'
import Chat from './chat/Chat';
import Models from './models/Models'
import Agents from './agents/Agents'
import Settings from './settings/Settings';
import 'bootstrap/dist/css/bootstrap.min.css';
import Providers from './providers/Providers'

const fadeIn = keyframes`
  from {
    transform: scale(1.025);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
`

const Container = styled.div`
  display: flex;
  height: 100vh;
  padding-left: 64px;
  animation: ${fadeIn} 0.3s linear;

  @media (min-width: 800px) {
    left: 200px;
    padding-left: 0;
  }
`

const Main = styled.div`
  flex-grow: 1;
  overflow-x: hidden;
  overflow-y: hidden;
  min-height: 100vh;
  padding-top: 1.4rem;
  position: relative;
`

export const Layout = () => (
  <Container data-testid="router-container">
    <Sidebar />
    <Main
      data-scrollelement // Required by react-virtualized implementation in Dashboard/TxList
    >
      <Routes>
        <Route path="/wallet" element={<Dashboard />} />
        <Route path="/chat" element={<Chat />} />
        <Route path="/agents" element={<Agents />} />
        <Route path="/models" element={<Models />} />
        <Route path="/providers" element={<Providers />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="*" element={<Navigate replace to="/wallet" />} />
      </Routes>
    </Main>
    {/* <AutoPriceAdjuster /> */}
    <OfflineWarning />
  </Container>
)

export default function Router() {
  return (
    <HashRouter>
      <Layout />
    </HashRouter>
  )
}
