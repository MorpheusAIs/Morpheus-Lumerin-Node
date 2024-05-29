import React, { createRef, useContext, useEffect, useRef, useState } from 'react'
// import component 👇
import Drawer from 'react-modern-drawer'
import { IconHistory, IconArrowUp, IconServer, IconWorld } from '@tabler/icons-react';
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
import { BtnAccent } from '../dashboard/BalanceBlock.styles';
import { withRouter } from 'react-router-dom';
import withChatState from '../../store/hocs/withChatState';
import { abbreviateAddress } from '../../utils'

import 'react-modern-drawer/dist/index.css'
import './Chat.css'
import { ChatHistory } from './ChatHistory';
import Spinner from 'react-bootstrap/Spinner';
import OpenSessionModal from './modals/OpenSessionModal';
import ModelSelectionModal from './modals/ModelSelectionModal';

const colors = [
    '#1899cb', '#da4d76', '#d66b38', '#d39d00', '#b46fc4', '#269c68', '#86858a'
];

const getColor = (name) => {
    if (!name) {
        return;
    }
    return colors[(getHashCode(name) + 1) % colors.length]
}

const Chat = (props) => {
    const chatBlockRef = useRef<null | HTMLDivElement>(null);

    const [value, setValue] = useState("");
    const [hasSession, setHasSession] = useState(true);

    const [chatHistory, setChatHistory] = useState<{ id: string, title: string }[]>();

    const [isSpinning, setIsSpinning] = useState(false);
    const [meta, setMeta] = useState({ budget: 0, supply: 0 });

    const [sessions, setSessions] = useState([{ id: "1245", title: "What can I do to save animals?" }])
    const [activeSession, setActiveSession] = useState("");

    const [chainData, setChainData] = useState<any>(null);

    const [openSessionModal, setOpenSessionModal] = useState(false);
    const [openChangeModal, setOpenChangeModal] = useState(false);

    const [selectedBid, setSelectedBid] = useState<any>(null);

    const modelName = props?.model?.Name || "Model";

    const isLocal = selectedBid?.Provider == 'Local';
    const providerAddress = isLocal ? "(local)" : selectedBid?.Provider ? abbreviateAddress(selectedBid?.Provider, 4) : null;

    useEffect(() => {
        props.getMetaInfo().then(setMeta);
        props.getModelsData().then((chainData) => {
            setChainData(chainData);
            const defaultSelectedBid = (chainData.models
                .find((x: any) => x.bids.find(b => b.Provider == 'Local')) as any).bids.find(b => b.Provider == 'Local');
            setSelectedBid(defaultSelectedBid);
        });
        // if(!props.activeSession) {
        //     props.history.push("/models");
        //     return;
        // }
    }, [])

    const [messages, setMessages] = useState<any>([]);

    const [isOpen, setIsOpen] = useState(false);
    const toggleDrawer = () => {
        setIsOpen((prevState) => !prevState)
    }

    const scrollToBottom = () => {
        chatBlockRef.current?.scrollIntoView({ behavior: "smooth", block: 'end' })
    }

    const onOpenSession = (stake) => {
        console.log("open-session", stake);
    }

    const closeSession = (sessionId: string) => {

    }

    const call = async (message) => {
        setIsSpinning(true);
        const chatHistory = messages.map(m => ({ role: m.role, content: m.text }))
        let response;

        if(isLocal) {
            response = await fetch(`${props.config.chain.localProxyRouterUrl}/v1/chat/completions`, {
                method: 'POST',
                headers: {
                    "Accept": "application/json"
                },
                body: JSON.stringify({
                    model: "llama2:latest",
                    stream: true,
                    messages: [
                        ...chatHistory,
                        {
                            role: "user",
                            content: message
                        }
                    ]
                })
            });
        }
        else {
            response = await fetch(`${props.config.chain.localProxyRouterUrl}/proxy/sessions/${props.activeSession.sessionId}/prompt`, {
                method: 'POST',
                body: JSON.stringify({
                    prompt: {
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
        }
        

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
        if (!value) {
            return;
        }

        setIsSpinning(true);
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
                <ChatHistory history={sessions} onCloseSession={closeSession} />
            </Drawer>
            <View>
                <ContainerTitle>
                    <TitleRow>
                        <Title>Chat</Title>
                        <div className='d-flex' style={{ alignItems: 'center' }}>
                            <div className='d-flex model-selector'>
                                <div className='model-selector__info'>
                                    <h3>{selectedBid?.Model?.Name}</h3>
                                    {
                                        isLocal ?
                                            (
                                                <>
                                                    <span>(local)</span>
                                                    <span>0 MOR/sec</span>
                                                </>
                                            )
                                            : (
                                                <>
                                                    <span>{providerAddress}</span>
                                                    <span>{selectedBid?.PricePerSecond || 0} MOR/sec</span>
                                                </>
                                            )
                                    }
                                </div>
                                <div className='model-selector__icons'>
                                    <IconServer width={'1.5rem'} color='#20dc8e'></IconServer>
                                    <IconWorld width={'1.5rem'}></IconWorld>
                                </div>
                            </div>
                            <BtnAccent className='change-modal' onClick={() => setOpenChangeModal(true)}>Change Model</BtnAccent>
                        </div>
                    </TitleRow>
                </ContainerTitle>
                <ChatTitleContainer>
                    <ChatAvatar>
                        <Avatar style={{ color: 'white' }} color={getColor(selectedBid?.Model?.Name[0])}>
                            {selectedBid?.Model?.Name[0]}
                        </Avatar>
                        <div style={{ marginLeft: '10px' }}>{selectedBid?.Model?.Name}</div>
                    </ChatAvatar>
                    <div>Provider: {isLocal ? "(local)" : providerAddress}</div>
                    <div>
                        <div onClick={toggleDrawer}>
                            <IconHistory size={"2.4rem"}></IconHistory>
                        </div>
                    </div>
                </ChatTitleContainer>

                <Container>
                    <ChatBlock ref={chatBlockRef} className={!messages?.length ? 'createSessionMode' : null}>
                        {
                            messages?.length ? messages.map(x => (
                                <Message key={makeid(6)} message={x}></Message>
                            ))
                                : (!isLocal && <div className='session-container' style={{ width: '400px' }}>
                                    <div className='session-title'>To perform promt please create session and choose desired session time</div>
                                    <div className='session-title'>Session will be created for selected Model</div>
                                    <div>
                                        <BtnAccent
                                            data-modal="receive"
                                            data-testid="receive-btn"
                                            styles={{ marginLeft: '0' }}
                                            onClick={() => setOpenSessionModal(true)}
                                            block
                                        >
                                            Create Session
                                        </BtnAccent></div>
                                </div>)
                        }
                    </ChatBlock>
                    <Control>
                        <CustomTextArrea
                            disabled={!hasSession}
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
                        <SendBtn disabled={!hasSession} onClick={handleSubmit}>{
                            isSpinning ? <Spinner animation="border" /> : <IconArrowUp size={"26px"}></IconArrowUp>
                        }</SendBtn>
                    </Control>
                </Container>
            </View>
            <OpenSessionModal
                pricePerSecond={selectedBid?.PricePerSecond}
                {...meta}
                isActive={openSessionModal}
                triggerOpen={(data) => {
                    onOpenSession(data);
                }}
                handleClose={() => setOpenSessionModal(false)} />
            <ModelSelectionModal
                models={(chainData as any)?.models}
                isActive={openChangeModal}
                onChangeModel={(id) => {
                    const defaultSelectedBid = (chainData.models
                        .find((x: any) => x.bids.find(b => b.Id == id)) as any)
                        .bids.find(b => b.Id == id);

                    setSelectedBid(defaultSelectedBid);
                }}
                handleClose={() => setOpenChangeModal(false)} />
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