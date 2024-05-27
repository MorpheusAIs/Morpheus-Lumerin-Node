import React from "react";
import styled from 'styled-components';
import { IconX } from "@tabler/icons-react";

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
`

interface ChatHistoryProps {
    history: { id: string, title: string}[],
    onCloseSession: (string) => void;
}

export const ChatHistory = (props: ChatHistoryProps) => {
    return (
        <Container>
            <Title>
                Sessions
            </Title>
            {
                props.history?.length && (
                    props.history.map(a => (
                    <HistoryItem key={a.id}>
                        {a.title}
                        <IconX onClick={() => props.onCloseSession(a.id)}></IconX>
                    </HistoryItem>)) 
                )
            }
        </Container>
    )
}