import styled from 'styled-components';
import { LoadingBar, AltLayout, Flex, Btn } from './common';
import { FC, useContext, useEffect, useState } from 'react';
import { withClient } from '../store/hocs/clientContext';
import type { Client } from 'src/main/src/client/subscriptions';
import withServicesState from '../store/hocs/withServicesState';
import { LoadingState } from 'src/main/orchestrator/orchestrator.types';
import {
  DownloadItemComponent,
  StartupItemComponent,
} from '@renderer/components/StartupItem';
import { ToastsContext } from '@renderer/components/toasts';

const Title = styled.div`
  font-weight: 600;
  font-size: 1em;
  margin-bottom: 1rem;
  color: ${(p) => p.theme.colors.dark};
`;

const Subtitle = styled.div`
  font-size: 0.8em;
  color: ${(p) => p.theme.colors.dark};
  margin-bottom: 1rem;
`;

const EntryGroup = styled(Flex.Column)`
  margin: 2rem 0;
  width: 100%;
  max-width: 600px;
`;

type LoadingProps = {
  services: LoadingState;
  client: Client;
};

const Loading: FC<LoadingProps> = ({ services, client }) => {
  const toast = useContext(ToastsContext);
  const [isError, setIsError] = useState(false);

  useEffect(() => {
    startServices();
  }, [client]);

  const onRetry = () => {
    startServices();
  };

  const startServices = async () => {
    try {
      setIsError(false);
      await client.startServices({});
    } catch (err) {
      toast.toast('error', 'Failed to start services');
      console.error(err);
      setIsError(true);
    }
  };

  const onExit = () => {
    client.quitApp();
  };

  return (
    <AltLayout title="Starting services...">
      <LoadingBar />

      <EntryGroup>
        <Title>Downloading</Title>
        <Subtitle>
          This will happen only on the first startup or after updating the app.
        </Subtitle>
        {services.download.map((item) => (
          <DownloadItemComponent key={item.name} item={item} />
        ))}
      </EntryGroup>

      <EntryGroup>
        <Title>Startup</Title>
        {services.startup.map((item) => (
          <StartupItemComponent
            key={item.name}
            item={item}
            onRestart={async () => {
              await client.restartService({ service: item.id });
            }}
            onPing={async () => {
              const res = await client.pingService({ service: item.id });
              res
                ? toast.toast('success', 'Ping successful')
                : toast.toast('error', 'Ping failed');
            }}
          />
        ))}
      </EntryGroup>

      <Actions>
        <RetryBtn onClick={onRetry} disabled={!isError}>
          Retry
        </RetryBtn>
        <ExitBtn onClick={onExit}>Exit</ExitBtn>
      </Actions>
    </AltLayout>
  );
};

const Actions = styled(Flex.Row)`
  gap: 1rem;
  justify-content: center;
`;

const RetryBtn = styled(Btn)`
  padding: 0.5rem 1rem;
  &:disabled {
    background-color: ${(p) => p.theme.colors.weak};
    color: ${(p) => p.theme.colors.primary};
  }
`;

const ExitBtn = styled(Btn)`
  padding: 0.5rem 1rem;
`;

export default withServicesState(withClient(Loading));
