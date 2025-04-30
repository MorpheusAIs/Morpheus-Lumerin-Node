import { Btn, Flex } from '@renderer/components/common';
import styled from 'styled-components';
import { FC, ForwardRefExoticComponent, useState } from 'react';
import {
  IconCheck,
  IconX,
  IconLoader2,
  IconChevronDown,
  IconChevronUp,
  IconClock,
  IconPlayerStop,
  IconRefresh,
  IconProps,
  IconBell,
} from '@tabler/icons-react';
import {
  DownloadItem,
  DownloadStatus,
  StartupItem,
  StartupStatus,
} from 'src/main/orchestrator/orchestrator.types';
import theme from '@renderer/ui/theme';

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

const EntryHeader = styled(Flex.Row)<{ hasProgress?: boolean }>`
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
  gap: 1rem;
`;

const SpinningLoaderIcon = styled(IconLoader2)`
  animation: spin 1s linear infinite;
  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
` as ForwardRefExoticComponent<IconProps>;

const statusIconColorMap: Record<
  StartupStatus | DownloadStatus,
  [ForwardRefExoticComponent<IconProps>, string]
> = {
  success: [IconCheck, theme.colors.success],
  running: [IconCheck, theme.colors.success],
  error: [IconX, theme.colors.danger],
  stopped: [IconPlayerStop, theme.colors.danger],
  pending: [IconClock, theme.colors.warning],
  starting: [SpinningLoaderIcon, theme.colors.primaryLight],
  downloading: [SpinningLoaderIcon, theme.colors.warning],
  unzipping: [SpinningLoaderIcon, theme.colors.warning],
};

const StatusIcon = styled.span`
  display: inline-flex;
  align-items: center;
`;

const Status = styled.span<{ color: string }>`
  color: ${(p) => p.color};
  font-weight: 500;
`;

export const StatusIconText = (props: {
  status: DownloadStatus | StartupStatus;
}) => {
  const [Icon, color] = statusIconColorMap[props.status];

  return (
    <IconText
      icon={<Icon size="16" stroke="1.5" color={color} />}
      text={props.status}
      color={color}
    />
  );
};

const IconText = (p: {
  icon: React.ReactNode;
  text: string;
  color: string;
}) => {
  return (
    <Flex.Row align="center" gap="0.5rem">
      <StatusIcon>{p.icon}</StatusIcon>
      <Status color={p.color}>{p.text}</Status>
    </Flex.Row>
  );
};

const Port = styled.span`
  background: rgba(105, 105, 105, 0.4);
  color: ${(p) => p.theme.colors.dark};

  opacity: 0.8;
  font-size: 0.7em;
  padding: 0rem 0.3rem;
  height: 2.3rem;
  display: flex;
  align-items: center;
  border-radius: 7px;
`;

const ProgressBarContainer = styled.div`
  background: ${(p) => p.theme.colors.morMain};
  border-radius: 4px;
  padding: 0;
  border: 2px solid ${(p) => p.theme.colors.morMain};
`;

const ProgressBar = styled.div.attrs<{ progress: number }>({
  style: ({ progress }) => {
    return {
      width: `${progress * 100}%`,
    };
  },
})`
  background: ${(p) => p.theme.colors.primary};
  height: 3px;
  border-radius: 4px;
  margin: 0;
`;

const Error = styled.div`
  color: ${(p) => p.theme.colors.danger};
  margin-top: 0.5rem;
  font-size: 0.875em;
`;

const LogsButton = styled.button`
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

const PortsWrapper = styled(Flex.Row)`
  gap: 0.5rem;
`;

export const DownloadItemComponent: FC<{ item: DownloadItem }> = ({ item }) => {
  return (
    <Entry>
      <EntryHeader hasProgress={!!item.progress}>
        <Name>{item.name}</Name>
        <StatusWrapper>
          <StatusIconText status={item.status} />
        </StatusWrapper>
      </EntryHeader>
      <ProgressBarContainer>
        <ProgressBar progress={item.progress} />
      </ProgressBarContainer>
      {item.error && <Error>{item.error}</Error>}
    </Entry>
  );
};

const RestartBtn = styled(Btn)`
  padding: 0rem 0.5rem;
  font-size: 1.3rem;
  background-color: ${(p) => p.theme.colors.warning};
`;

const PingBtn = styled(Btn)`
  padding: 0rem 0.5rem;
  font-size: 1.3rem;
  background-color: ${(p) => p.theme.colors.morMain};
`;

export const StartupItemComponent: FC<{
  item: StartupItem;
  onRestart: () => void;
  onPing: () => Promise<void>;
  alwaysShowPingRestart?: boolean;
}> = (props) => {
  const [showLogs, setShowLogs] = useState(false);
  const [isPinging, setIsPinging] = useState(false);

  const handleRestart = () => {
    props.onRestart?.();
  };

  const handlePing = async () => {
    setIsPinging(true);
    await props.onPing?.();
    setIsPinging(false);
  };

  const isManagedProcess = props.item.isExternal === false;

  return (
    <Entry>
      <EntryHeader>
        <Flex.Row gap="0.5rem">
          <Name>{props.item.name}</Name>
          <PortsWrapper>
            {props.item.ports?.map((port) => <Port key={port}>:{port}</Port>)}
          </PortsWrapper>
        </Flex.Row>
        <StatusWrapper>
          <StatusIconText status={props.item.status} />
        </StatusWrapper>
      </EntryHeader>
      {props.item.error && <Error>{props.item.error}</Error>}
      <Flex.Row justify="space-between" align="center">
        {isManagedProcess && (
          <LogsButton onClick={() => setShowLogs(!showLogs)}>
            {showLogs ? (
              <IconChevronUp size={16} />
            ) : (
              <IconChevronDown size={16} />
            )}
            {showLogs ? 'Hide logs' : 'Show logs'}
          </LogsButton>
        )}
        {props.alwaysShowPingRestart || props.item.status === 'stopped' ? (
          <Flex.Row gap="0.5rem" justify="flex-end" grow="1">
            <PingBtn onClick={handlePing}>
              <IconText
                icon={
                  isPinging ? (
                    <SpinningLoaderIcon
                      size={14}
                      color={theme.colors.morLight}
                    />
                  ) : (
                    <IconBell size={14} color={theme.colors.morLight} />
                  )
                }
                text="Ping"
                color={theme.colors.morLight}
              />
            </PingBtn>
            {isManagedProcess && (
              <RestartBtn onClick={handleRestart}>
                <IconText
                  icon={<IconRefresh size={14} color={theme.colors.morLight} />}
                  text="Restart"
                  color={theme.colors.morLight}
                />
              </RestartBtn>
            )}
          </Flex.Row>
        ) : null}
      </Flex.Row>
      {showLogs && <ProcessLogs>{props.item.stderrOutput}</ProcessLogs>}
    </Entry>
  );
};
