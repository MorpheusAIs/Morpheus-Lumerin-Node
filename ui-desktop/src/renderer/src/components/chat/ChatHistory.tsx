import React, { useEffect, useState } from "react";
import styled from 'styled-components';
import { IconX } from "@tabler/icons-react";
import { abbreviateAddress } from '../../utils';

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
    sessions: any[];
}

export const ChatHistory = (props: ChatHistoryProps) => {
    const sessions = props.sessions;
    
    return (
        <Container>
            <Title>
                Sessions
            </Title>
            {
                sessions?.length && (
                    sessions.map(a => (
                    <HistoryItem key={a.Id}>
                        <div>{abbreviateAddress(a.Id, 5)}</div>
                        <div>{new Date(a.OpenedAt * 1000).toLocaleString()}</div>
                        {
                            !a.ClosedAt ? (<IconX onClick={() => props.onCloseSession(a.Id)}></IconX>) : <div>CLOSED</div>
                        }
                    </HistoryItem>)) 
                )
            }
        </Container>
    )
}