import { useContext, useMemo, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import styled from 'styled-components';
import Accordion from 'react-bootstrap/Accordion';

import { abbreviateAddress } from '../../utils';
import Table from 'react-bootstrap/Table';
import Button from 'react-bootstrap/Button';
import { ToastsContext } from '../toasts';
import { queryKeys } from '../../store/queries';
import './Providers.css';

const BidTable = styled(Table)`
  text-align: left;
  font-size: 1.3rem;
  --bs-table-border-color: var(--border-subtle);
  --bs-table-striped-bg: transparent;
  font-variant-numeric: tabular-nums;

  th {
    background: var(--surface-raised) !important;
    color: var(--text-muted) !important;
    padding: 1.2rem !important;
    font-weight: 550;
  }

  td {
    background: var(--surface-base) !important;
    color: var(--text-primary) !important;
    padding: 1.2rem !important;
    vertical-align: middle;
  }
`;

const StartBtn = styled(Button)`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 3.6rem;
  padding: 0.8rem 1.2rem;
  font-size: 1.3rem;
  background: var(--surface-hover) !important;
  color: var(--accent) !important;
  border-radius: 8px !important;
  border: 1px solid var(--border-strong) !important;
`;

const Container = styled.div`
  overflow: auto;
`;

const EmptyState = styled.div`
  color: var(--text-muted);
  font-size: 1.4rem;
  padding: 3.2rem 0;
`;

const formatBalance = (session, balances, balancesLoading) => {
  if (session.ClosedAt) return '0 MOR';
  const balance = balances[session.Id];
  if (balance == null) return balancesLoading ? 'Loading…' : 'Unavailable';
  const numericBalance = Number(balance);
  return Number.isFinite(numericBalance)
    ? `${numericBalance / 10 ** 18} MOR`
    : 'Unavailable';
};

export const groupProviderSessions = (sessions, modelNames) => {
  const groups = new Map();
  for (const session of sessions ?? []) {
    const modelId = String(session.ModelAgentId ?? 'Unknown model');
    const key = modelId.toLowerCase();
    const current = groups.get(key) ?? {
      id: modelId,
      name: modelNames[key] ?? abbreviateAddress(modelId, 6),
      sessions: [],
    };
    current.sessions.push(session);
    groups.set(key, current);
  }
  return [...groups.values()];
};

function renderTable({
  onClaim,
  claiming,
  sessions,
  balances,
  balancesLoading,
}) {
  return (
    <BidTable striped bordered hover size="sm">
      <thead>
        <tr>
          <th>Session</th>
          <th>Bid</th>
          <th>Status</th>
          <th>Balance</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        {sessions?.length
          ? sessions.map((b) => {
              // const provider = providers?.find(x => x.Address.toLowerCase() === b.Provider.toLowerCase());
              return (
                <tr key={b.Id}>
                  <td>{abbreviateAddress(b.Id, 5)}</td>
                  <td>{abbreviateAddress(b.BidID, 5)}</td>
                  <td>{b.ClosedAt ? 'CLOSED' : 'OPEN'}</td>
                  <td>{formatBalance(b, balances, balancesLoading)}</td>
                  <td>
                    {!b.ClosedAt && (
                      <StartBtn
                        disabled={!!claiming}
                        onClick={() => onClaim(b.Id)}
                      >
                        {claiming === b.Id ? 'Claiming…' : 'Claim'}
                      </StartBtn>
                    )}
                  </td>
                </tr>
              );
            })
          : null}
      </tbody>
    </BidTable>
  );
}

function ProvidersList({
  sessions,
  sessionsLoading,
  modelNames,
  balances,
  balancesLoading,
  claimFunds,
  providerId,
}) {
  const context = useContext(ToastsContext);
  const queryClient = useQueryClient();
  const [claiming, setClaiming] = useState<string | null>(null);

  // Claim used to fail silently (see withProvidersState.claimFunds). Now the
  // result is actually surfaced, and the table refreshes so the claimed
  // balance disappears without a manual tab switch.
  const handleClaim = async (sessionId: string) => {
    if (claiming) {
      return;
    }
    setClaiming(sessionId);
    try {
      await claimFunds(sessionId);
      context.toast('success', 'Funds claimed');
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: queryKeys.providerSessions(providerId),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.providerBalances(providerId),
        }),
      ]);
    } catch (e: any) {
      context.toast('error', e?.message || 'Failed to claim funds');
    } finally {
      setClaiming(null);
    }
  };

  const groups = useMemo(
    () => groupProviderSessions(sessions, modelNames),
    [sessions, modelNames],
  );

  if (sessionsLoading) {
    return <EmptyState role="status">Loading provider sessions…</EmptyState>;
  }

  if (!groups.length) {
    return (
      <EmptyState>
        No provider sessions yet. Sessions will appear here when your provider
        serves a request.
      </EmptyState>
    );
  }

  return (
    <Container>
      {groups.map((group) => {
        return (
          <Accordion alwaysOpen key={group.id}>
            <Accordion.Item eventKey={group.id}>
              <Accordion.Header className="model-header">
                {group.name}
              </Accordion.Header>
              <Accordion.Body>
                {renderTable({
                  onClaim: handleClaim,
                  claiming,
                  sessions: group.sessions,
                  balances,
                  balancesLoading,
                })}
              </Accordion.Body>
            </Accordion.Item>
          </Accordion>
        );
      })}
    </Container>
  );
}

export default ProvidersList;
