import styled from 'styled-components';
import { IconX, IconRefresh } from "@tabler/icons-react";
import { abbreviateAddress } from '../../utils';
import { isClosed } from './utils';

const Container = styled.div`
    text-align: center;
`

const Title = styled.div`
    text-align: center;
    margin-bottom: 2.4rem;
`
const HistoryItem = styled.div`
    color: ${p => p.theme.colors.morMain}
    display: flex;
    align-items: center;
    justify-content: space-evenly;
    cursor: pointer;
    padding: 5px 0;
`

interface ChatHistoryProps {
    onCloseSession: (string) => void;
    onSelectSession: (string) => void;
    refreshSessions: () => void;
    sessions: any[];
}

export const ChatHistory = (props: ChatHistoryProps) => {
    const sessions = props.sessions;
    
    return (
        <Container>
            <Title>
                Sessions <span><IconRefresh width={"14px"}></IconRefresh></span>
            </Title>
            {
                sessions?.length && (
                    sessions.map(a => (
                    <HistoryItem key={a.Id} onClick={() => props.onSelectSession(a.Id)}>
                        <div>{abbreviateAddress(a.Id, 3)}</div>
                        <div>{(a.EndsAt - a.OpenedAt) / 60} min</div>
                        {
                            !isClosed(a) ? (<IconX onClick={() => props.onCloseSession(a.Id)}></IconX>) : <div>CLOSED</div>
                        }
                    </HistoryItem>)) 
                )
            }
        </Container>
    )
}