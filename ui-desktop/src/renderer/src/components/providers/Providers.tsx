import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import ProvidersList from './ProvidersList';

import { BtnAccent } from '../dashboard/BalanceBlock.styles';

import withProvidersState from '../../store/hocs/withProvidersState';
import { queryKeys } from '../../store/queries';
import QueryError from '../common/QueryError';
import { pooledMapSettled, withTimeout } from '../../store/utils/concurrency';

const BALANCE_CONCURRENCY = 6;
const BALANCE_TIMEOUT_MS = 8_000;

export const loadProviderBalances = async (
  sessions: any[],
  getBalanceBySession: (sessionId: string) => Promise<unknown>,
) => {
  const entries = await pooledMapSettled(
    sessions,
    async (session) =>
      [
        session.Id,
        await withTimeout(getBalanceBySession(session.Id), BALANCE_TIMEOUT_MS),
      ] as const,
    (session) => [session.Id, null] as const,
    BALANCE_CONCURRENCY,
  );
  return Object.fromEntries(entries);
};

export const Providers = ({
  getAllModels,
  getSessionsByProvider,
  getBalanceBySession,
  claimFunds,
  providerId,
}) => {
  // Cached, stale-while-revalidate: revisiting the Providers tab renders the
  // last result instantly and revalidates in the background instead of
  // re-running the expensive per-session balance fetch on every mount.
  const sessionsQuery = useQuery({
    queryKey: queryKeys.providerSessions(providerId),
    queryFn: () => getSessionsByProvider(providerId),
    enabled: !!providerId,
  });
  const sessions = (sessionsQuery.data ?? []) as any[];
  const modelsQuery = useQuery({
    queryKey: queryKeys.allModels,
    queryFn: getAllModels,
    // Model names are only decoration for the session groups. Let the
    // session request own the cold path; cached names remain available
    // immediately, and an uncached registry scan starts after rows arrive.
    enabled: sessions.length > 0,
  });
  const openSessions = useMemo(
    () => sessions.filter((session) => !session.ClosedAt),
    [sessions],
  );
  const openSessionIds = useMemo(
    () => openSessions.map((session) => String(session.Id)),
    [openSessions],
  );
  const balancesQuery = useQuery({
    queryKey: queryKeys.providerBalances(providerId, openSessionIds),
    queryFn: () => loadProviderBalances(openSessions, getBalanceBySession),
    enabled: !!providerId && openSessions.length > 0,
  });
  const modelNames = useMemo(() => {
    const names: Record<string, string> = {};
    for (const model of (modelsQuery.data ?? []) as any[]) {
      names[String(model.Id).toLowerCase()] = model.Name;
    }
    return names;
  }, [modelsQuery.data]);

  return (
    <View data-testid="providers-container">
      <QueryError
        error={sessionsQuery.error}
        what="provider sessions"
        onRetry={() => sessionsQuery.refetch()}
      />
      <QueryError
        error={modelsQuery.error}
        what="model names"
        onRetry={() => modelsQuery.refetch()}
      />
      <LayoutHeader title="Providers">
        <BtnAccent style={{ padding: '1.5rem' }} disabled>
          Add provider
        </BtnAccent>
      </LayoutHeader>
      <ProvidersList
        sessions={sessions}
        sessionsLoading={sessionsQuery.isPending && !!providerId}
        modelNames={modelNames}
        balances={balancesQuery.data ?? {}}
        balancesLoading={balancesQuery.isPending && openSessions.length > 0}
        claimFunds={claimFunds}
        providerId={providerId}
      />
    </View>
  );
};

export default withProvidersState(Providers);
