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

const parse = (decodedChunk) => {
    const lines = decodedChunk.split('\n');
    const trimmedData = lines.map(line => line.replace(/^data: /, "").trim());
    const filteredData = trimmedData.filter(line => !["", "[DONE]"].includes(line));
    const parsedData = filteredData.map(line => JSON.parse(line));

    return parsedData;
}

let abort = false;

const Chat = (props) => {
    const chatBlockRef = useRef<null | HTMLDivElement>(null);

    const [value, setValue] = useState("");

    const [sessions, setSessions] = useState<any>();

    const [isSpinning, setIsSpinning] = useState(false);
    const [meta, setMeta] = useState({ budget: 0, supply: 0 });

    const [activeSession, setActiveSession] = useState<any>(undefined);

    const [chainData, setChainData] = useState<any>(null);

    const [openSessionModal, setOpenSessionModal] = useState(false);
    const [openChangeModal, setOpenChangeModal] = useState(false);

    const [selectedBid, setSelectedBid] = useState<any>(null);

    const modelName = selectedBid?.Model?.Name || "Model";

    const isLocal = selectedBid?.Provider == 'Local';
    const providerAddress = isLocal ? "(local)" : selectedBid?.Provider ? abbreviateAddress(selectedBid?.Provider, 4) : null;

    useEffect(() => {
        props.getMetaInfo().then((meta) => {
            setMeta(meta);
        });
        props.getModelsData().then((chainData) => {
            setChainData(chainData);
            const defaultSelectedBid = (chainData.models
                .find((x: any) => x.bids.find(b => b.Provider == 'Local')) as any).bids.find(b => b.Provider == 'Local');
            setSelectedBid(defaultSelectedBid);
        });
        refreshSessions();
    }, [])

    const [messages, setMessages] = useState<any>([]);

    const [isOpen, setIsOpen] = useState(false);
    const toggleDrawer = () => {
        setIsOpen((prevState) => !prevState)
    }

    const scrollToBottom = () => {
        chatBlockRef.current?.scroll({ top: chatBlockRef.current.scrollHeight, behavior: 'smooth' })
    }

    const onOpenSession = ({ stake }) => {
        console.log("open-session", stake);

        props.onOpenSession({ stake, selectedBid }).then((res) => {
            if (!res) {
                return;
            }
            setActiveSession(res);
            refreshSessions();
        })
    }

    const refreshSessions = () => {
        return props.getSessionsByUser(props.address).then(setSessions);
    }

    const closeSession = (sessionId: string) => {
        props.closeSession(sessionId).then(refreshSessions);
    }

    const call = async (message) => {
        const chatHistory = messages.map(m => ({ role: m.role, content: m.text }))

        const headers = {
            "Accept": "application/json"
        };
        if (!isLocal) {
            headers["session_id"] = activeSession.sessionId;
        }
        
        const response = await fetch(`${props.config.chain.localProxyRouterUrl}/v1/chat/completions`, {
            method: 'POST',
            headers,
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
        }).catch((e) => {
            console.log("Failed to send request", e)
            return null;
        });

        if (!response) {
            return;
        }

        if (!response.ok) {
            console.log("Failed", await response.json())
        }

        const textDecoder = new TextDecoder();

        if (response.body != null) {
            const reader = response.body.getReader()

            let memoState = [...messages, { id: "some", user: 'Me', text: value, role: "user", icon: "M", color: "#20dc8e" }];

            while (true) {
                if (abort) {
                    await reader.cancel();
                    abort = false;
                }
                const { value, done } = await reader.read();
                if (done) {
                    setIsSpinning(false);
                    break;
                }
                const decodedString = textDecoder.decode(value, { stream: true });
                const parts = parse(decodedString);
                parts.forEach(part => {
                    if (!part?.id) {
                        return;
                    }
                    const message = memoState.find(m => m.id == part.id);
                    const otherMessages = memoState.filter(m => m.id != part.id);
                    const text = `${message?.text || ''}${part?.choices[0]?.delta?.content || ''}`;
                    const result = [...otherMessages, { id: part.id, user: modelName, role: "assistant", text: text, icon: modelName.toUpperCase()[0], color: getColor(modelName.toUpperCase()[0]) }];
                    memoState = result;
                    setMessages(result);
                    scrollToBottom();
                })
            }

        }
    }

    const handleSubmit = () => {
        if (isSpinning) {
            abort = true;
            setIsSpinning(false);
            return;
        }

        if (!value) {
            return;
        }

        setIsSpinning(true);
        setMessages([...messages, { id: "some", user: 'Me', text: value, role: "user", icon: "M", color: "#20dc8e" }]);
        call(value).then(() => {
            setIsSpinning(false);
        });
        setValue("");
        scrollToBottom();
    }

    const price = selectedBid?.PricePerSecond ? selectedBid?.PricePerSecond / (10 ** 18) : 0;

    return (
        <>
            <Drawer
                open={isOpen}
                onClose={toggleDrawer}
                direction='right'
                className='history-drawer'
            >
                <ChatHistory
                    sessions={sessions}
                    onCloseSession={closeSession} />
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
                                                    <span>{price} MOR/sec</span>
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
                                : (!isLocal && !activeSession && <div className='session-container' style={{ width: '400px' }}>
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
                            disabled={!activeSession && !isLocal}
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
                        <SendBtn disabled={!activeSession && !isLocal} onClick={handleSubmit}>{
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
                    setOpenSessionModal(false)
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
                    setMessages([]);
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
