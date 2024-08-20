import { useEffect, useRef, useState } from 'react'
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
    SendBtn,
    LoadingCover,
    ImageContainer,
    SubPriceLabel
} from './Chat.styles';
import { BtnAccent } from '../dashboard/BalanceBlock.styles';
import { withRouter } from 'react-router-dom';
import withChatState from '../../store/hocs/withChatState';
import { abbreviateAddress } from '../../utils'

import 'react-modern-drawer/dist/index.css'
import './Chat.css'
import { ChatHistory } from './ChatHistory';
import Spinner from 'react-bootstrap/Spinner';
import { formatSmallNumber } from './utils';
import ModelSelectionModal from './modals/ModelSelectionModal';
import { parseDataChunk, makeId, getColor, isClosed } from './utils';
import { Cooldown } from './Cooldown';
import ImageViewer from "react-simple-image-viewer";

let abort = false;
let cancelScroll = false;
const userMessage = { user: 'Me', role: "user", icon: "M", color: "#20dc8e" };

const Chat = (props) => {
    const chatBlockRef = useRef<null | HTMLDivElement>(null);

    const [value, setValue] = useState("");
    const [isLoading, setIsLoading] = useState(true);
    const [messages, setMessages] = useState<any>([]);
    const [isOpen, setIsOpen] = useState(false);
    const [sessions, setSessions] = useState<any>();

    const [isSpinning, setIsSpinning] = useState(false);
    const [meta, setMeta] = useState({ budget: 0, supply: 0 });

    const [imagePreview, setImagePreview] = useState<string>();
    const [activeSession, setActiveSession] = useState<any>(undefined);

    const [chainData, setChainData] = useState<any>(null);
    const [sessionTitles, setSessionTitles] = useState<{ sessionId: string, title: string }[]>([]);

    const [openChangeModal, setOpenChangeModal] = useState(false);
    const [isReadonly, setIsReadonly] = useState(false);

    const [selectedBid, setSelectedBid] = useState<any>(null);
    const [selectedModel, setSelectedModel] = useState<any>(undefined);
    const [requiredStake, setRequiredStake] = useState<{ min: Number, max: number }>({ min: 0, max: 0 })
    const [balances, setBalances] = useState<{ eth: Number, mor: number }>({ eth: 0, mor: 0 });

    const modelName = selectedModel?.Name || "Model";
    const isLocal = selectedModel?.useLocal;

    const providerAddress = isLocal ? "(local)" : selectedBid?.Provider ? abbreviateAddress(selectedBid?.Provider, 5) : null;
    const isDisabled = (!activeSession && !isLocal) || isReadonly;
    const isEnoughFunds = Number(balances.mor) > Number(requiredStake.min);

    useEffect(() => {
        (async () => {
            const [meta, chainData, titles, userBalances] = await Promise.all([
                props.getMetaInfo(),
                props.getModelsData(),
                props.client.getTitles(),
                props.getBalances()]);

            setBalances(userBalances)

            setSessionTitles(titles.map(t => ({ sessionId: t._id, title: t.title })));

            const sessions = await props.getSessionsByUser(props.address);
            const openSessions = sessions.filter(s => !isClosed(s));

            if (openSessions.length) {
                const latestSession = openSessions[0];
                const latestSessionModel = (chainData.models.find((m: any) => m.Id == latestSession.ModelAgentId));
                if (latestSessionModel) {
                    setSelectedModel(latestSessionModel);
                }

                const openBid = latestSessionModel?.bids?.find(b => b.Id == latestSession.BidID);
                if (openBid) {
                    setSelectedBid(openBid);
                }
                await onSetActiveSession({ sessionId: latestSession.Id, endDate: latestSession.EndsAt })
            }
            else {
                const localModel = (chainData?.models?.find((m: any) => m.hasLocal));
                if (localModel) {
                    setSelectedModel({ ...localModel, useLocal: true });
                }
            }

            setMeta(meta);
            setChainData(chainData)
            setSessions(sessions);
        })().then(() => {
            setIsLoading(false);
        })
    }, [])

    const toggleDrawer = () => {
        setIsOpen((prevState) => !prevState)
    }

    const selectLocalModel = () => {
        const localModel = (chainData?.models?.find((m: any) => m.hasLocal));
        if (localModel) {
            setSelectedModel({ ...localModel, useLocal: true });
        }
    }

    const scrollToBottom = () => {
        if (!cancelScroll) {
            chatBlockRef.current?.scroll({ top: chatBlockRef.current.scrollHeight, behavior: 'smooth' })
        }
    }

    const calculateAcceptableDuration = (pricePerSecond: number, balance: number, stakingInfo) => {
        const delta = 60; // 1 minute

        if (balance > requiredStake.max) {
            return 24 * 60 * 60; // 1 day in seconds
        }

        const targetDuration = Math.round((balance * Number(stakingInfo.budget)) / (Number(stakingInfo.supply) * pricePerSecond))

        if (targetDuration - delta < 5 * 60) {
            return 5 * 60;
        }

        return (targetDuration - (targetDuration % 60)) - delta;
    }

    const onOpenSession = async () => {
        setIsLoading(true);

        const prices = selectedModel.bids.map(x => x.PricePerSecond);
        const maxPrice = Math.max(prices);
        const duration = calculateAcceptableDuration(maxPrice, Number(balances.mor), meta);

        console.log("open-session", duration);

        try {
            const openedSession = await props.onOpenSession({ modelId: selectedModel.Id, duration });
            if (!openedSession) {
                return;
            }
            setActiveSession({ sessionId: openedSession });
            var allSessions = await refreshSessions();
            var targetSessionData = allSessions.find(x => x.Id == openedSession);
            var targetModel = chainData.models.find(x => x.Id == targetSessionData.ModelAgentId)
            var targetBid = targetModel.bids.find(x => x.Id == targetSessionData.BidID);
            setSelectedBid(targetBid);
        }
        finally {
            setIsLoading(false);
        }
    }

    const onSetActiveSession = async (session) => {
        setActiveSession(session);
        if (session) {
            try {
                const history = await props.client.getChatHistory(session.sessionId);
                setMessages(history.length ? (history[0].messages || []) : []);
            }
            catch (e) {
                props.toasts.toast('error', 'Failed to load chat history');
            }
        }
        scrollToBottom(); 
    }

    const refreshSessions = async () => {
        const sessions = await props.getSessionsByUser(props.address);
        setSessions(sessions);
        return sessions;
    }

    const closeSession = async (sessionId: string) => {
        await props.closeSession(sessionId);
        await refreshSessions();

        if (activeSession.sessionId == sessionId) {
            selectLocalModel();
            setMessages([]);
        }
    }

    const selectSession = async (sessionId: string) => {
        const findBid = (id) => {
            return (chainData.models
                .find((x: any) => x.bids.find(b => b.Id == id)) as any)
                .bids.find(b => b.Id == id);
        }

        console.log("select-session", sessionId)
        toggleDrawer();

        const openSessions = sessions.filter(s => !isClosed(s));
        const openSession = openSessions.find(s => s.Id == sessionId);
        if (!openSession) {
            setIsReadonly(true)

            const closedSession = sessions.find(s => s.Id == sessionId);
            if (closedSession) {
                await onSetActiveSession({ sessionId: closedSession.Id })
                const selectedBid = findBid(closedSession.BidID);
                setSelectedBid(selectedBid);
                const selectedModel = chainData.models.find((m: any) => m.Id == closedSession.ModelAgentId);
                setSelectedModel(selectedModel);
            }
            return;
        }
        else {
            setIsReadonly(false)
            await onSetActiveSession({ sessionId: openSession.Id, endDate: openSession.EndsAt })
            const selectedBid = findBid(openSession.BidID);
            setSelectedBid(selectedBid);
            const selectedModel = chainData.models.find((m: any) => m.Id == openSession.ModelAgentId);
            setSelectedModel(selectedModel);
        }
        setTimeout(scrollToBottom, 400);
    }

    const registerScrollEvent = (register) => {
        cancelScroll = false;
        const handler = (event: any) => {
            const isUp = event.wheelDelta ? event.wheelDelta > 0 : event.deltaY < 0;
            if (isUp) {
                cancelScroll = true;
            }
            else {
                if (!chatBlockRef?.current || !cancelScroll) {
                    return;
                }
                // Return scrolling if scrolled to div end 
                if ((chatBlockRef.current.offsetHeight + chatBlockRef.current.scrollTop) >= chatBlockRef.current.scrollHeight) {
                    cancelScroll = false;
                }
            }
        };

        if (register) {
            chatBlockRef?.current?.addEventListener('wheel', handler);
        }
        else {
            chatBlockRef?.current?.removeEventListener('wheel', handler);
        }
    }

    const call = async (message) => {
        scrollToBottom();
        const chatHistory = messages.map(m => ({ role: m.role, content: m.text, isImageContent: m.isImageContent }))

        let memoState = [...messages, { id: makeId(6), text: value, ...userMessage }];
        setMessages(memoState);

        const headers = {
            "Accept": "application/json"
        };
        if (isLocal) {
            headers["model_id"] = selectedModel.Id;
        } else {
            headers["session_id"] = activeSession.sessionId;
        }

        const hasImageHistory = chatHistory.some(x => x.isImageContent);
        const incommingMessage = { role: "user", content: message };
        const payload = {
            stream: true,
            messages: hasImageHistory ? [incommingMessage] : [...chatHistory, incommingMessage]
        };

        // If image take only last message
        const response = await fetch(`${props.config.chain.localProxyRouterUrl}/v1/chat/completions`, {
            method: 'POST',
            headers,
            body: JSON.stringify(payload)
        }).catch((e) => {
            console.log("Failed to send request", e)
            return null;
        });

        if (!response) {
            return;
        }

        if (!response.ok) {
            console.log("Failed", await response.json())
            props.toasts.toast('error', 'Failed to send prompt');
            return;
        }

        const textDecoder = new TextDecoder();

        if (!response.body) {
            console.error("Body is missed");
            return;
        }

        const reader = response.body.getReader()
        registerScrollEvent(true);

        const iconProps = { icon: modelName.toUpperCase()[0], color: getColor(modelName.toUpperCase()[0]) };
        try {
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
                const parts = parseDataChunk(decodedString);
                parts.forEach(part => {
                    if (part.error) {
                        console.warn(part.error);
                        return;
                    }
                    const imageContent = part.imageUrl;

                    if (!part?.id && !imageContent) {
                        return;
                    }

                    let result: any[] = [];
                    const message = memoState.find(m => m.id == part.id);
                    const otherMessages = memoState.filter(m => m.id != part.id);
                    if (imageContent) {
                        result = [...otherMessages, { id: part.job, user: modelName, role: "assistant", text: imageContent, isImageContent: true, ...iconProps }];
                    }
                    else {
                        const text = `${message?.text || ''}${part?.choices[0]?.delta?.content || ''}`.replace("<|im_start|>", "").replace("<|im_end|>", "");
                        result = [...otherMessages, { id: part.id, user: modelName, role: "assistant", text: text, ...iconProps }];
                    }
                    memoState = result;
                    setMessages(result);
                    scrollToBottom();
                })
            }
        }
        catch (e) {
            console.error(e);
        }

        console.log("Flush to storage");
        await props.client.saveChatHistory({ sessionId: activeSession.sessionId, messages: memoState });
        console.log("Stored succesfully");

        registerScrollEvent(false);
        return memoState;
    }

    const handleSubmit = () => {
        if (!value) {
            return;
        }

        if (abort) {
            abort = false;
        }

        if (isSpinning) {
            abort = true;
            setIsSpinning(false);
            return;
        }

        if (!isLocal && messages.length === 0) {
            const title = { sessionId: activeSession.sessionId, title: value };
            props.client.saveTitle(title).then(() => {
                setSessionTitles([...sessionTitles, title]);
            }).catch(console.error);
        }

        setIsSpinning(true);
        call(value).finally(() => setIsSpinning(false));
        setValue("");
    }


    const calculateStake = (pricePerSecond, durationInMin) => {
        const totalCost = pricePerSecond * durationInMin * 60;
        const stake = totalCost * Number(meta.supply) / Number(meta.budget);
        return stake;
    }

    const onBidSelect = ({ modelId, isLocal }) => {
        // TODO: Add support for custom Bid.
        setMessages([]);
        setActiveSession(undefined);
        setSelectedBid(undefined);
        setIsReadonly(false);
        abort = true;

        if (isLocal) {
            const localModel = (chainData?.models?.find((m: { Id: string }) => m.Id == modelId));
            if (localModel) {
                setSelectedModel({ ...localModel, useLocal: true });
            } else {
                props.toasts.toast('error', 'Failed to select local model');
            }
            return;
        }

        const selectedModel = chainData.models.find((m: any) => m.Id == modelId);
        setSelectedModel(selectedModel);

        const openSessions = sessions.filter(s => !isClosed(s));
        const openModelSession = openSessions.find(s => s.ModelAgentId == modelId);

        if (openModelSession) {
            const selectedBid = selectedModel.bids.find(b => b.Id == openModelSession.BidID);
            if (selectedBid) {
                setSelectedBid(selectedBid);
            }
            onSetActiveSession({ sessionId: openModelSession.Id })
            return;
        }

        const prices = selectedModel.bids.map(x => x.PricePerSecond);
        const maxPrice = Math.max(prices);

        setRequiredStake({ min: calculateStake(maxPrice, 5), max: calculateStake(maxPrice, 24 * 60) })
    }

    return (
        <>
            {
                isLoading &&
                <LoadingCover>
                    <Spinner style={{ width: '5rem', height: '5rem' }} animation="border" variant="success" />
                </LoadingCover>
            }
            <Drawer
                open={isOpen}
                onClose={toggleDrawer}
                direction='right'
                className='history-drawer'
            >
                <ChatHistory
                    sessionTitles={sessionTitles}
                    sessions={sessions}
                    models={chainData?.models || []}
                    onSelectSession={selectSession}
                    refreshSessions={refreshSessions}
                    onCloseSession={closeSession} />
            </Drawer>
            <View>
                <ContainerTitle>
                    <TitleRow>
                        <Title>Chat</Title>
                        <div className='d-flex' style={{ alignItems: 'center' }}>
                            <div className='d-flex model-selector'>
                                <div className='model-selector__info'>
                                    <h3>{modelName}</h3>
                                    {
                                        isLocal ?
                                            (
                                                <>
                                                    <span>(local)</span>
                                                </>
                                            )
                                            : ( 
                                                <>
                                                    <SubPriceLabel>{selectedBid ? formatSmallNumber(selectedBid?.PricePerSecond / (10 ** 18)) : 0} MOR/s</SubPriceLabel>
                                                </>
                                            )
                                    }
                                </div>
                                {

                                    !isLocal && activeSession?.endDate && (
                                        <div className='model-selector__icons'>
                                            <Cooldown endDate={activeSession?.endDate} />
                                        </div>
                                    )
                                }

                            </div>
                            <BtnAccent className='change-modal' onClick={() => setOpenChangeModal(true)}>Change Model</BtnAccent>
                        </div>
                    </TitleRow>
                </ContainerTitle>
                <ChatTitleContainer>
                    <ChatAvatar>
                        <Avatar style={{ color: 'white' }} color={getColor(modelName[0])}>
                            {modelName[0]}
                        </Avatar>
                        <div style={{ marginLeft: '10px' }}>{modelName}</div>
                    </ChatAvatar>
                    { 
                        (selectedBid || isLocal) && <div>
                            <span style={{ color: 'white' }}>Provider:</span> {isLocal ? "(local)" : providerAddress}
                        </div>
                    }
                    <div>
                        <div onClick={toggleDrawer}>
                            <IconHistory size={"2.4rem"}></IconHistory>
                        </div>
                    </div>
                </ChatTitleContainer>

                {imagePreview && (
                    <ImageViewer
                    src={[imagePreview]}
                    onClose={() => setImagePreview("")}
                    disableScroll={false}
                    backgroundStyle={{
                        backgroundColor: "rgba(0,0,0,0.9)",
                        zIndex: 1000
                    }}
                    closeOnClickOutside={true}
                    />
                )}

                <Container>
                    <ChatBlock ref={chatBlockRef} className={!messages?.length ? 'createSessionMode' : null}>
                        {
                            messages?.length ? messages.map(x => (
                                <Message key={makeId(6)} message={x} onOpenImage={setImagePreview}></Message>
                            ))
                                : (!isLocal && !activeSession &&
                                    <div className='session-container' style={{ width: '400px' }}>
                                        {
                                            isEnoughFunds ?
                                                <>
                                                    <div className='session-title'>Staked MOR funds will be reserved to start session</div>
                                                    <div className='session-title'>Session may last from 5 mins to 24 hours depending on staked funds (min: {(Number(requiredStake.min) / 10 ** 18).toFixed(2)}, max: {(Number(requiredStake.max) / 10 ** 18).toFixed(2)} MOR)</div>
                                                </> :
                                                <div className='session-title'>To start session required balance should be at least {(Number(requiredStake.min) / 10 ** 18).toFixed(2)} MOR</div>
                                        }
                                        <div>
                                            <BtnAccent
                                                data-modal="receive"
                                                data-testid="receive-btn"
                                                styles={{ marginLeft: '0' }}
                                                block={requiredStake.min}
                                                onClick={onOpenSession}
                                                disabled={!isEnoughFunds}
                                            >
                                                Start
                                            </BtnAccent></div>
                                    </div>)
                        }
                    </ChatBlock>
                    <Control>
                        <CustomTextArrea
                            disabled={isDisabled}
                            onKeyPress={(e) => {
                                if (e.key === 'Enter') {
                                    e.preventDefault();
                                    handleSubmit();
                                }
                            }}
                            value={value}
                            onChange={ev => setValue(ev.target.value)}
                            placeholder={isReadonly ? "Session is closed. Chat in ReadOnly Mode" : "Ask me anything..."}
                            minRows={1}
                            maxRows={6} />
                        <SendBtn disabled={isDisabled} onClick={handleSubmit}>{
                            isSpinning ? <Spinner animation="border" /> : <IconArrowUp size={"26px"}></IconArrowUp>
                        }</SendBtn>
                    </Control>
                </Container>
            </View>
            <ModelSelectionModal
                models={(chainData as any)?.models}
                isActive={openChangeModal}
                onChangeModel={(eventData) => {
                    onBidSelect(eventData);
                }}
                handleClose={() => setOpenChangeModal(false)} />
        </>
    )
}

const Message = ({ message, onOpenImage }) => {
    return (
        <div style={{ display: 'flex', margin: '12px 0 28px 0' }}>
            <Avatar color={message.color}>
                {message.icon}
            </Avatar>
            <div>
                <AvatarHeader>{message.user}</AvatarHeader>
                {
                    message.isImageContent
                        ? (<MessageBody>{<ImageContainer src={message.text} onClick={() => onOpenImage(message.text)} />}</MessageBody>)
                        : (<MessageBody>{message.text}</MessageBody>)
                }
            </div>
        </div>)
}

export default withRouter(withChatState(Chat));
