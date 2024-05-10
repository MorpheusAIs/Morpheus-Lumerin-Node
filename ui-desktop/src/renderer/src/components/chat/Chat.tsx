import React, { createRef, useContext, useEffect, useRef, useState } from 'react'
// import component 👇
import Drawer from 'react-modern-drawer'
import { IconHistory, IconArrowUp } from '@tabler/icons-react';
import {
    View,
    ContainerTitle,
    ChatTitleContainer,
    ChatAvatar,
    Avatar,
    Title,
    TitleRow,
    AvatarHeader,
    MessageBody,
    Container,
    ChatBlock,
    CustomTextArrea,
    Control,
    SendBtn
} from './Chat.styles';
import { withRouter } from 'react-router-dom';
import withChatState from '../../store/hocs/withChatState';
import { abbreviateAddress } from '../../utils'

import 'react-modern-drawer/dist/index.css'
import './Chat.css'
import { ChatHistory } from './ChatHistory';
import Spinner from 'react-bootstrap/Spinner';

const colors = [
    '#1899cb', '#da4d76', '#d66b38', '#d39d00', '#b46fc4', '#269c68', '#86858a'
];

const getColor = (name) => {
    return colors[(getHashCode(name) + 1) % colors.length]
}

const Chat = (props) => {
    const chatBlockRef = useRef<null | HTMLDivElement>(null);

    const [value, setValue] = useState("");

    const [chatHistory, setChatHistory] = useState<string[]>([]);
    const [isSpinning, setIsSpinning] = useState(false);

    const modelName = props?.model?.Name || "GPT";
    const providerAddress = props?.provider?.Address ? abbreviateAddress(props?.provider?.Address, 4) : null;

    useEffect(() => {
        if(!props.activeSession) {
            props.history.push("/models");
            return;
        }
    }, [])

    const [messages, setMessages] = useState<any>([]);

    const [isOpen, setIsOpen] = useState(false);
    const toggleDrawer = () => {
        setIsOpen((prevState) => !prevState)
    }

    const scrollToBottom = () => {
        chatBlockRef.current?.scrollIntoView({ behavior: "smooth", block: 'end' })
    }

    const call = async (message) => {
        setIsSpinning(true);
        const chatHistory = messages.map(m => ({ role: m.role, content: m.text }))
        const response = await fetch(`${props.config.chain.localProxyRouterUrl}/proxy/sessions/${props.activeSession.sessionId}/prompt`, {
            method: 'POST',
            body: JSON.stringify({
                prompt : { 
                    model: "llama2:latest",
                    stream: true,
                    messages: [
                        ...chatHistory,
                        {
                            role: "user",
                            content: message
                        }
                    ]
                },
                providerUrl: props.provider.Endpoint.replace("http://", ""),
                providerPublicKey: props.activeSession.signature
            })
        });
        
        function parse(decodedChunk) {
            const lines = decodedChunk.split('\n');
            const trimmedData = lines.map(line => line.replace(/^data: /, "").trim());
            const filteredData = trimmedData.filter(line => !["", "[DONE]"].includes(line));
            const parsedData = filteredData.map(line => JSON.parse(line));
            
            return parsedData;
        }

        const textDecoder = new TextDecoder();

        if (response.body != null) {
            const reader = response.body.getReader()

            let memoState = [...messages, { id: "some", user: 'Me', text: value, role: "user", icon: "M", color: "#20dc8e" }];
            
            while (true) {
                const { value, done } = await reader.read();
                if (done) {
                    break;
                }
                const decodedString = textDecoder.decode(value, { stream: true });
                
                const parts = parse(decodedString);
                parts.forEach(part => {
                    const message = memoState.find(m => m.id == part.id);
                    const otherMessages = memoState.filter(m => m.id != part.id);
                    const text = `${message?.text || ''}${part?.choices[0]?.delta?.content || ''}`;
                    const result = [...otherMessages, { id: part.id, user: modelName, role: "assistant", text: text, icon: "L", color: getColor("L") }];
                    memoState = result;
                    setMessages(result);
                    scrollToBottom();
                })
            }

            setIsSpinning(false);
        }
    }

    const handleSubmit = () => {
        if(!value) {
            return;
        }
        
        setIsSpinning(true);
        setChatHistory([...chatHistory, value]);
        setMessages([...messages, { id: "some", user: 'Me', text: value, role: "user", icon: "M", color: "#20dc8e" }]);
        call(value);
        setValue("");
    }

    return (
        <>
            <Drawer
                open={isOpen}
                onClose={toggleDrawer}
                direction='right'
                className='history-drawer'
            >
                <ChatHistory history={chatHistory} />
            </Drawer>
            <View>
                <ContainerTitle style={{ padding: '0 2.4rem' }}>
                    <TitleRow>
                        <Title>Chat</Title>
                    </TitleRow>
                </ContainerTitle>
                <ChatTitleContainer>
                    <ChatAvatar>
                        <Avatar style={{ color: 'white' }} color={getColor("L")}>
                            L
                        </Avatar>
                        <div style={{ marginLeft: '10px' }}>{modelName}</div>
                    </ChatAvatar>
                    <div>Provider: {providerAddress}</div>
                    <div>
                        <div onClick={toggleDrawer}>
                            <IconHistory size={"2.4rem"}></IconHistory>
                        </div>
                    </div>
                </ChatTitleContainer>

                <Container>
                    <ChatBlock ref={chatBlockRef}>
                        {
                            messages && messages.map(x => (
                                <Message key={makeid(6)} message={x}></Message>
                            ))
                        }
                    </ChatBlock>
                    <Control>
                        <CustomTextArrea
                            onKeyPress={(e) => {
                                if (e.key === 'Enter') {
                                    e.preventDefault();
                                    handleSubmit();
                                }
                            }}
                            value={value}
                            onChange={ev => setValue(ev.target.value)}
                            // style={{ background: 'transparent', boxSizing: 'border-box'}}
                            placeholder={"Ask me anything..."}
                            minRows={1}
                            maxRows={6} />
                        <SendBtn onClick={handleSubmit}>{
                            isSpinning ? <Spinner animation="border" />: <IconArrowUp size={"26px"}></IconArrowUp>
                        }</SendBtn>
                    </Control>
                </Container>
            </View>
        </>
    )
}

const Message = ({ message }) => {
    return (
        <div style={{ display: 'flex', margin: '12px 0 28px 0' }}>
            <Avatar color={message.color}>
                {message.icon}
            </Avatar>
            <div>
                <AvatarHeader>{message.user}</AvatarHeader>
                <MessageBody>{message.text}</MessageBody>
            </div>
        </div>)
}

function makeid(length) {
    let result = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    let counter = 0;
    while (counter < length) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
        counter += 1;
    }
    return result;
}

function getHashCode(string) {
    var hash = 0;
    for (var i = 0; i < string.length; i++) {
        var code = string.charCodeAt(i);
        hash = ((hash << 5) - hash) + code;
        hash = hash & hash; // Convert to 32bit integer
    }
    return Math.abs(hash);
}

export default withRouter(withChatState(Chat));