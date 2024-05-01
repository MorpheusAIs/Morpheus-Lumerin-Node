import React from "react";
import styled from 'styled-components';

const Container = styled.div`
    text-align: center;
`

const Title = styled.div`
    text-align: center;
    margin-bottom: 2.4rem;
`
const HistoryItem = styled.div`
    color: ${p => p.theme.colors.morMain}
`
export const ChatHistory = (props: { history: string[]}) => {
    return (
        <Container>
            <Title>
                History
            </Title>
            {
                props.history?.length && (
                    props.history.map(a => (<HistoryItem>{a}</HistoryItem>)) 
                )
            }
        </Container>
    )
}