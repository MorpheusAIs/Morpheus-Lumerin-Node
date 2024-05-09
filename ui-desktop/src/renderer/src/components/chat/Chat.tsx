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

import 'react-modern-drawer/dist/index.css'
import './Chat.css'
import { ChatHistory } from './ChatHistory';

const lorem = "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."
const colors = [
    '#1899cb', '#da4d76', '#d66b38', '#d39d00', '#b46fc4', '#269c68', '#86858a'
];

const getColor = (name) => {
    return colors[getHashCode(name) % colors.length]
}

const Chat = (props) => {
    const chatBlockRef = useRef<null | HTMLDivElement>(null);

    const [value, setValue] = useState("");

    const [history, setHistory] = useState(["What is Lorem Ipsum?"]);
    const [isSpinning, setIsSpinning] = useState(false);

    const user = props.chat || 'Llama GPT';

    const [messages, setMessages] = useState<any>([]);

    const [isOpen, setIsOpen] = useState(false);
    const toggleDrawer = () => {
        setIsOpen((prevState) => !prevState)
    }

    const scrollToBottom = () => {
        chatBlockRef.current?.scrollIntoView({ behavior: "smooth" })
    }

    const call = async (message) => {

        const response = await fetch(`http://localhost:11434/v1/chat/completions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                        "model": "llama2:latest",
                        "stream": true,
                        "messages": [
                            {
                                "role": "user",
                                "content": message
                            }
                        ]
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

        let answer = ""
        if (response.body != null) {
            const reader = response.body.getReader()

            let memoState = [...messages, { id: "some", user: 'Me', text: value, role: "user", icon: "M", color: colors[getHashCode("M") % colors.length] }];
            
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
    
                    const text = (message?.text || '') + part.choices[0].delta.content;
    
                    const result = [...otherMessages, { id: part.id, user: 'GPT', role: "assistant", text: text, icon: "GPT", color: getColor("GPT") }];
                    memoState = result;
                    setMessages(result);
                })
            }
        }

        try {
            const response = await fetch("http://localhost:11434/v1/chat/completions", {
                method: 'POST',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    model: "llama2:latest",
                    messages: [
                    {
                        role: "user",
                        content: message
                    }]
                })
            });
            const data = await response.json();
            setMessages([...messages, { user: 'GPT', role: "assistant", text: data.choices, icon: "GPT", color: getColor("GPT") }]);
        }
        catch (e) {
            setMessages([...messages, { user: 'GPT', role: "assistant", text: "Ooops, cannot answer", icon: "GPT", color: getColor("GPT") }]);
        }
        finally {
            setIsSpinning(false);
        }
    }

    const handleSubmit = () => {
        if(!value) {
            return;
        }
        
        setIsSpinning(true);
        setMessages([...messages, { id: "some", user: 'Me', text: value, role: "user", icon: "M", color: colors[getHashCode("M") % colors.length] }]);
        scrollToBottom();
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
                <ChatHistory history={history} />
            </Drawer>
            <View>
                <ContainerTitle style={{ padding: '0 2.4rem' }}>
                    <TitleRow>
                        <Title>Chat</Title>
                    </TitleRow>
                </ContainerTitle>
                <ChatTitleContainer>
                    <ChatAvatar>
                        <Avatar style={{ color: 'white' }} color={getColor("GPT")}>
                            GPT
                        </Avatar>
                        <div style={{ marginLeft: '10px' }}>Llama GPT</div>
                    </ChatAvatar>
                    <div>Provider: 0x123...234</div>
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
                        <SendBtn disabled={isSpinning} onClick={handleSubmit}><IconArrowUp size={"26px"}></IconArrowUp></SendBtn>
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

export default Chat;