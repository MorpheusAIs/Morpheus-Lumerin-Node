import styled from 'styled-components';
import { LoadingBar, AltLayout, Flex } from './common';
import { FC, useEffect, useState } from 'react';
import { withClient } from '../store/hocs/clientContext';
import type { Client } from 'src/main/src/client/subscriptions';
import withServicesState from '../store/hocs/withServicesState';
import {
  DownloadItem,
  StartupItem,
  LoadingState,
} from 'src/main/orchestrator.types';
import {
  IconCheck,
  IconX,
  IconLoader2,
  IconChevronDown,
  IconChevronUp,
  IconClock,
  IconPlayerStop,
} from '@tabler/icons-react';

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

const Entry = styled.div`
  margin: 0.5rem 0;
  padding: 1rem;
  background: ${(p) => p.theme.colors.morLight};
  border-radius: 4px;
  min-height: 3.5rem;
  display: flex;
  flex-direction: column;
  justify-content: center;
`;

const EntryHeader = styled(Flex.Row)`
  justify-content: space-between;
  align-items: center;
  margin-bottom: ${(props) => (props.hasProgress ? '0.5rem' : '0')};
`;

const Name = styled.div`
  font-weight: 500;
  font-size: 0.875em;
`;

const StatusWrapper = styled(Flex.Row)`
  align-items: center;
  font-size: 0.875em;
`;

const StatusIcon = styled.span<{ status: string }>`
  display: inline-flex;
  align-items: center;
  margin-right: 0.5rem;
  color: ${(p) =>
    p.status === 'error' || p.status === 'stopped'
      ? p.theme.colors.danger
      : p.status === 'success' || p.status === 'running'
        ? p.theme.colors.success
        : p.status === 'pending'
          ? p.theme.colors.warning
          : p.status === 'starting'
            ? p.theme.colors.primaryLight
            : p.status === 'downloading'
              ? p.theme.colors.warning
              : p.status === 'unzipping'
                ? p.theme.colors.primary
                : p.theme.colors.dark};

  svg {
    animation: ${(props) =>
      props.status === 'starting' ||
      props.status === 'downloading' ||
      props.status === 'unzipping'
        ? 'spin 1s linear infinite'
        : 'none'};
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
`;

const Status = styled.span<{ status: string }>`
  color: ${(p) =>
    p.status === 'error' || p.status === 'stopped'
      ? p.theme.colors.danger
      : p.status === 'success' || p.status === 'running'
        ? p.theme.colors.success
        : p.status === 'pending'
          ? p.theme.colors.warning
          : p.status === 'starting'
            ? p.theme.colors.primaryLight
            : p.status === 'downloading'
              ? p.theme.colors.warning
              : p.status === 'unzipping'
                ? p.theme.colors.primary
                : p.theme.colors.dark};
  font-weight: 500;
`;

const ProgressBarContainer = styled.div`
  background: ${(p) => p.theme.colors.morMain};
  border-radius: 4px;
  padding: 0;
  border: 2px solid ${(p) => p.theme.colors.morMain};
`;

const ProgressBar = styled.div<{ progress: number }>`
  background: ${(p) => p.theme.colors.primary};
  height: 3px;
  width: ${(p) => p.progress * 100}%;
  border-radius: 4px;
  margin: 0;
`;

const Error = styled.div`
  color: ${(p) => p.theme.colors.danger};
  margin-top: 0.5rem;
  font-size: 0.875em;
`;

const StderrButton = styled.button`
  background: none;
  border: none;
  color: ${(p) => p.theme.colors.dark};
  font-size: 0.875em;
  padding: 0.5rem 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  opacity: 0.7;
  transition: opacity 0.2s;

  &:hover {
    opacity: 1;
  }
`;

const ProcessLogs = styled.pre`
  background: ${(p) => p.theme.colors.copy};
  padding: 1rem;
  border-radius: 4px;
  margin-top: 0.5rem;
  font-size: 0.875em;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 200px;
  overflow-y: auto;
`;

type LoadingProps = {
  services: LoadingState;
  client: Client;
};

const getStatusIcon = (status: string) => {
  const iconProps = { size: 16, stroke: 1.5 };

  switch (status) {
    case 'success':
      return <IconCheck {...iconProps} />;
    case 'error':
      return <IconX {...iconProps} />;
    case 'running':
      return <IconCheck {...iconProps} />;
    case 'pending':
      return <IconClock {...iconProps} />;
    case 'starting':
      return <IconLoader2 {...iconProps} />;
    case 'stopped':
      return <IconPlayerStop {...iconProps} />;
    case 'downloading':
    case 'unzipping':
      return <IconLoader2 {...iconProps} />;
    default:
      return null;
  }
};

const DownloadItemComponent: FC<{ item: DownloadItem }> = ({ item }) => {
  return (
    <Entry>
      <EntryHeader hasProgress>
        <Name>{item.name}</Name>
        <StatusWrapper>
          <StatusIcon status={item.status}>
            {getStatusIcon(item.status)}
          </StatusIcon>
          <Status status={item.status}>{item.status}</Status>
        </StatusWrapper>
      </EntryHeader>
      <ProgressBarContainer>
        <ProgressBar progress={item.progress} />
      </ProgressBarContainer>
      {item.error && <Error>{item.error}</Error>}
    </Entry>
  );
};

const StartupItemComponent: FC<{ item: StartupItem }> = ({ item }) => {
  const [showStderr, setShowStderr] = useState(false);

  return (
    <Entry>
      <EntryHeader>
        <Name>{item.name}</Name>
        <StatusWrapper>
          <StatusIcon status={item.status}>
            {getStatusIcon(item.status)}
          </StatusIcon>
          <Status status={item.status}>{item.status}</Status>
        </StatusWrapper>
      </EntryHeader>
      {item.error && <Error>{item.error}</Error>}
      {item.stderrOutput && (
        <>
          <StderrButton onClick={() => setShowStderr(!showStderr)}>
            {showStderr ? (
              <IconChevronUp size={16} />
            ) : (
              <IconChevronDown size={16} />
            )}
            {showStderr ? 'Hide details' : 'Show details'}
          </StderrButton>
          {showStderr && <ProcessLogs>{item.stderrOutput}</ProcessLogs>}
        </>
      )}
    </Entry>
  );
};

const Loading: FC<LoadingProps> = ({ services, client }) => {
  useEffect(() => {
    client.startServices({});
  }, [client]);

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
          <StartupItemComponent key={item.name} item={item} />
        ))}
      </EntryGroup>
    </AltLayout>
  );
};

export default withServicesState(withClient(Loading));
