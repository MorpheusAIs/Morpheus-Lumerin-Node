import styled from 'styled-components';
import { IconX, IconRefresh } from "@tabler/icons-react";
import { abbreviateAddress } from '../../utils';
import { isClosed } from './utils';

const Container = styled.div`

    .history-scroll-block {
        overflow-y: auto;
        height: calc(100vh - 100px);
    }
`

const Title = styled.div`
    text-align: center;
    margin-bottom: 2.4rem;

    span {
        cursor: pointer;
    }
`
const HistoryItem = styled.div`
    color: ${p => p.theme.colors.morMain}
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 5px 0 0 0;
`
const HistoryEntryContainer = styled.div`
    background: rgba(255,255,255, 0.04);
    border-width: 1px;
    border: 1px solid rgba(255, 255, 255, 0.04);
    color: white;
    margin-bottom: 15px;
    cursor: pointer;
    padding: 10px;
`

const HistoryEntryTitle = styled.div`
    text-align: justify;
    text-overflow: ellipsis;
    overflow: hidden;
    white-space: nowrap;
`

const ModelName = styled.div`
    text-overflow: ellipsis;
    width: 181px;
    height: 24px;
    overflow: auto;
`

const Duration = styled.div`
    color: white;
`

interface ChatHistoryProps {
    onCloseSession: (string) => void;
    onSelectSession: (string) => void;
    refreshSessions: () => void;
    sessions: any[];
    models: any[];
    sessionTitles: {sessionId: string, title: string }[]
}

export const ChatHistory = (props: ChatHistoryProps) => {
    const sessions = props.sessions;

    return (
        <Container>
            <Title>
                Sessions <span onClick={props.refreshSessions}><IconRefresh width={"14px"}></IconRefresh></span>
            </Title>
            <div className='history-scroll-block'>
                {
                    sessions?.length ? (
                        sessions.map(a => {
                            const titleObj = props.sessionTitles.find(t => t.sessionId.toLowerCase() == String(a.Id).toLowerCase());
                            const title = titleObj?.title || "";
                            const model = props.models.find(x => x.Id == a.ModelAgentId);
                            return (
                                <HistoryEntryContainer key={a.Id} onClick={() => props.onSelectSession(a.Id)}>
                                    {title ? <HistoryEntryTitle data-rh={title} data-rh-negative>{title}</HistoryEntryTitle> : null}
                                    <HistoryItem>
                                        <ModelName data-rh={abbreviateAddress(a.Id, 3)} data-rh-negative>{model?.Name}</ModelName>
                                        <Duration>{((a.EndsAt - a.OpenedAt) / 60).toFixed(0)} min</Duration>
                                        {
                                            !isClosed(a) ? (<IconX onClick={() => props.onCloseSession(a.Id)}></IconX>) : <div>CLOSED</div>
                                        }
                                    </HistoryItem>
                                </HistoryEntryContainer>
                            )
                        }
                        ))
                        : <div>You have not any sessions</div>
                }
            </div>
        </Container>
    )
}