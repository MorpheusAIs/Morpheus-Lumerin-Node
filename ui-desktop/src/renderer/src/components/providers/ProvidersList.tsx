import { useContext, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import withProvidersState from '../../store/hocs/withProvidersState';
import styled from 'styled-components';
import Accordion from 'react-bootstrap/Accordion';

import { abbreviateAddress } from '../../utils';
import Table from 'react-bootstrap/Table';
import Button from 'react-bootstrap/Button';
import { ToastsContext } from '../toasts';
import { queryKeys } from '../../store/queries';
import './Providers.css';

const BidTable = styled(Table)`
  text-align: center !important;
  border: 0.5px solid#21dc8f !important;

  th {
    background: #244a47 !important;
    color: #21dc8f !important;
  }

  td {
    background: #244a47 !important;
    color: #21dc8f !important;
    padding: 12px 0 !important;
  }
`;

const StartBtn = styled(Button)`
  background: rgba(0, 0, 0, 0.9) !important;
  border-radius: 0 !important;
  border: 1px solid #21dc8f !important;
`;

const Container = styled.div`
  height: 75vh;
  overflow-y: auto;
`;

function renderTable({ onClaim, claiming, sessions }) {
  return (
    <BidTable striped bordered hover size="sm">
      <thead>
        <tr>
          <th>Session</th>
          <th>Bid</th>
          <th>Status</th>
          <th>Balance</th>
          <th></th>
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
                  <td>{b.Balance / 10 ** 18} MOR</td>
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

function ProvidersList({ data, claimFunds, providerId }) {
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
      await queryClient.invalidateQueries({
        queryKey: queryKeys.providerData(providerId),
      });
    } catch (e: any) {
      context.toast('error', e?.message || 'Failed to claim funds');
    } finally {
      setClaiming(null);
    }
  };

  return (
    <Container>
      {data?.modelsNames &&
        Object.keys(data?.modelsNames).map((model) => {
          const modelSessions = data.results.filter(
            (r) => r.ModelAgentId.toLowerCase() == model.toLowerCase(),
          );

          return (
            <Accordion alwaysOpen key={model}>
              <Accordion.Item eventKey={model}>
                <Accordion.Header className="model-header">
                  {data?.modelsNames[model]}
                </Accordion.Header>
                <Accordion.Body>
                  {renderTable({
                    onClaim: handleClaim,
                    claiming,
                    sessions: modelSessions,
                  })}
                </Accordion.Body>
              </Accordion.Item>
            </Accordion>
          );
        })}
    </Container>
  );
}

export default withProvidersState(ProvidersList);
