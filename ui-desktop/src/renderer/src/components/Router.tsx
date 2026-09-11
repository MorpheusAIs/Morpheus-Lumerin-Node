import { lazy, Suspense, useCallback, useEffect } from 'react';
import { HashRouter, Routes, Route, Navigate, useLocation } from 'react-router';
import { useSelector } from 'react-redux';
import { useQueryClient } from '@tanstack/react-query';
import styled, { keyframes } from 'styled-components';
import OfflineWarning from './OfflineWarning';
import QuickStartGuide from './onboarding/QuickStartGuide';
// import ChangePassword from './ChangePassword'
import Sidebar from './sidebar/Sidebar';
import 'bootstrap/dist/css/bootstrap.min.css';
import { withClient } from '../store/hocs/clientContext';
import selectors from '../store/selectors';
import { queryKeys } from '../store/queries';
import ErrorBoundary from './common/ErrorBoundary';
import { AppRoutePath, preloadRoute, routeModules } from './routeModules';
import { modelPagesQueryOptions } from '../store/modelQueries';

const Dashboard = lazy(routeModules['/wallet']);
const Chat = lazy(routeModules['/chat']);
const Cowork = lazy(routeModules['/workspace']);
const Agents = lazy(routeModules['/agents']);
const Models = lazy(routeModules['/models']);
const Providers = lazy(routeModules['/providers']);
const Settings = lazy(routeModules['/settings']);

const LegacyCoworkRedirect = () => {
  const location = useLocation();
  return <Navigate replace to={`/workspace${location.search}`} />;
};

const fadeIn = keyframes`
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
`;

const Container = styled.div`
  --sidebar-width: 220px;
  display: flex;
  height: 100vh;
  padding-left: 64px;
  animation: ${fadeIn} 160ms ease-out;

  @media (min-width: 800px) {
    left: 200px;
    padding-left: 0;
  }
`;

const Main = styled.div`
  container: route-space / inline-size;
  flex-grow: 1;
  min-width: 0;
  overflow-x: hidden;
  overflow-y: hidden;
  min-height: 100vh;
  position: relative;
`;

const RouteLoading = styled.div`
  align-items: center;
  background: #04130d;
  color: rgba(255, 255, 255, 0.62);
  display: flex;
  font-family: var(--font-ui);
  font-size: 1.2rem;
  height: 100vh;
  justify-content: center;
`;

// Warms the shared session cache as soon as the main app shell mounts, so the
// first visit to the Chat or Wallet tab finds sessions already loaded (the
// heaviest, paginated, cross-tab call). Subsequent visits hit the cache via the
// stale-while-revalidate config. Failures are non-fatal — the tabs refetch.
const SessionPrefetcher = withClient(({ client }: any) => {
  const queryClient = useQueryClient();
  const address = useSelector((state: any) =>
    selectors.getWalletAddress(state),
  );
  useEffect(() => {
    if (!address) {
      return;
    }
    queryClient
      .prefetchQuery({
        queryKey: queryKeys.sessions(address),
        queryFn: async () => {
          return (await client.getSessionsByUser({ user: address })) || [];
        },
      })
      .catch((e) => console.warn('Session prefetch failed', e));
  }, [address, queryClient, client]);

  return null;
});

// Navigation intent warms both the code chunk and the first useful response.
// This keeps expensive chain reads out of initial wallet startup while still
// beginning them before the click that opens Models or Provider Hub.
const RoutePrefetchSidebar = withClient(({ client }: any) => {
  const queryClient = useQueryClient();
  const address = useSelector((state: any) =>
    selectors.getWalletAddress(state),
  );
  const onRouteIntent = useCallback(
    (path: AppRoutePath) => {
      const requests: Promise<unknown>[] = [preloadRoute(path)];
      if (path === '/models') {
        requests.push(
          queryClient.prefetchInfiniteQuery(
            modelPagesQueryOptions(client.getModelsPage),
          ),
        );
      }
      if (path === '/providers' && address) {
        requests.push(
          queryClient.prefetchQuery({
            queryKey: queryKeys.providerSessions(address),
            queryFn: () => client.getSessionsByProvider({ provider: address }),
          }),
        );
      }
      return Promise.all(requests);
    },
    [address, client, queryClient],
  );

  return <Sidebar onRouteIntent={onRouteIntent} />;
});

export const Layout = () => {
  // Keyed on pathname so a crash on one tab is cleared when the user navigates
  // elsewhere, instead of persisting for the rest of the session. The boundary
  // wraps only the route outlet — the sidebar stays usable, so a broken screen
  // never traps the user.
  const location = useLocation();

  return (
    <Container data-testid="router-container">
      <RoutePrefetchSidebar />
      <Main
        data-scrollelement // Required by react-virtualized implementation in Dashboard/TxList
      >
        <ErrorBoundary resetKey={location.pathname} label={location.pathname}>
          <Suspense
            fallback={
              <RouteLoading role="status" aria-live="polite">
                Loading screen…
              </RouteLoading>
            }
          >
            <Routes>
              <Route path="/wallet" element={<Dashboard />} />
              <Route path="/chat" element={<Chat />} />
              <Route path="/workspace" element={<Cowork />} />
              <Route path="/cowork" element={<LegacyCoworkRedirect />} />
              <Route path="/agents" element={<Agents />} />
              <Route path="/models" element={<Models />} />
              <Route path="/providers" element={<Providers />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<Navigate replace to="/wallet" />} />
            </Routes>
          </Suspense>
        </ErrorBoundary>
      </Main>
      {/* <AutoPriceAdjuster /> */}
      <SessionPrefetcher />
      <QuickStartGuide />
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
