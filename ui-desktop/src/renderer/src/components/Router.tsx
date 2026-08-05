import { useEffect } from 'react';
import { HashRouter, Routes, Route, Navigate, useLocation } from 'react-router';
import { useSelector } from 'react-redux';
import { useQueryClient } from '@tanstack/react-query';
import styled, { keyframes } from 'styled-components';
import OfflineWarning from './OfflineWarning';
// import ChangePassword from './ChangePassword'
import Dashboard from './dashboard/Dashboard';
import Sidebar from './sidebar/Sidebar';
import Chat from './chat/Chat';
import Models from './models/Models';
import Agents from './agents/Agents';
import Settings from './settings/Settings';
import 'bootstrap/dist/css/bootstrap.min.css';
import Providers from './providers/Providers';
import { withClient } from '../store/hocs/clientContext';
import selectors from '../store/selectors';
import { queryKeys } from '../store/queries';
import { getSessionsByUser } from '../store/utils/apiCallsHelper';
import ErrorBoundary from './common/ErrorBoundary';

const fadeIn = keyframes`
  from {
    transform: scale(1.025);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
`;

const Container = styled.div`
  display: flex;
  height: 100vh;
  padding-left: 64px;
  animation: ${fadeIn} 0.3s linear;

  @media (min-width: 800px) {
    left: 200px;
    padding-left: 0;
  }
`;

const Main = styled.div`
  flex-grow: 1;
  overflow-x: hidden;
  overflow-y: hidden;
  min-height: 100vh;
  position: relative;
`;

// Warms the shared session cache as soon as the main app shell mounts, so the
// first visit to the Chat or Wallet tab finds sessions already loaded (the
// heaviest, paginated, cross-tab call). Subsequent visits hit the cache via the
// stale-while-revalidate config. Failures are non-fatal — the tabs refetch.
const SessionPrefetcher = withClient(({ client }: any) => {
  const queryClient = useQueryClient();
  const address = useSelector((state: any) => selectors.getWalletAddress(state));
  const url = useSelector((state: any) =>
    selectors.getLocalProxyRouterUrl(state),
  );

  useEffect(() => {
    if (!address || !url) {
      return;
    }
    queryClient
      .prefetchQuery({
        queryKey: queryKeys.sessions(address),
        queryFn: async () => {
          const authHeaders = await client.getAuthHeaders();
          return (await getSessionsByUser(url, address, authHeaders)) || [];
        },
      })
      .catch((e) => console.warn('Session prefetch failed', e));
  }, [address, url, queryClient, client]);

  return null;
});

export const Layout = () => {
  // Keyed on pathname so a crash on one tab is cleared when the user navigates
  // elsewhere, instead of persisting for the rest of the session. The boundary
  // wraps only the route outlet — the sidebar stays usable, so a broken screen
  // never traps the user.
  const location = useLocation();

  return (
    <Container data-testid="router-container">
      <Sidebar />
      <Main
        data-scrollelement // Required by react-virtualized implementation in Dashboard/TxList
      >
        <ErrorBoundary resetKey={location.pathname} label={location.pathname}>
          <Routes>
            <Route path="/wallet" element={<Dashboard />} />
            <Route path="/chat" element={<Chat />} />
            <Route path="/agents" element={<Agents />} />
            <Route path="/models" element={<Models />} />
            <Route path="/providers" element={<Providers />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="*" element={<Navigate replace to="/wallet" />} />
          </Routes>
        </ErrorBoundary>
      </Main>
      {/* <AutoPriceAdjuster /> */}
      <SessionPrefetcher />
      <OfflineWarning />
    </Container>
  );
};

export default function Router() {
  return (
    <HashRouter>
      <Layout />
    </HashRouter>
  );
}
