import { IconPencil, IconTrash, IconSearch, IconCheck, IconX } from "@tabler/icons-react";
import { abbreviateAddress } from '../../utils';
import { isClosed } from './utils';
import Tab from 'react-bootstrap/Tab';
import Tabs from 'react-bootstrap/Tabs';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Badge from 'react-bootstrap/Badge';
import { SearchContainer } from '../contracts/modals/CreateContractModal.styles';
import * as components from './ChatHistory.styles';
import { useEffect, useState } from "react";

interface ChatHistoryProps {
    open: boolean,
    onCloseSession: (string) => void;
    onSelectSession: (session) => void;
    refreshSessions: () => void;
    deleteHistory: (string) => void
    onChangeTitle: (data: { id, title }) => Promise<void>;
    sessions: any[];
    models: any[];
    activeChat: any,
    sessionTitles: { sessionId: string, title: string, createdAt: any }[]
}

const HisotryEntry = ({ entry, deleteHistory, onSelectSession, isActive, onChangeTitle }) => {
    const [isEdit, setIsEdit] = useState<boolean>(false);
    const [title, setTitle] = useState<string>(entry?.title || "");

    const wrapDelete = (e, id) => {
        e.stopPropagation();
        deleteHistory(id);
    }

    const changetTitle = (e, id) => {
        setIsEdit(!isEdit);
        e.stopPropagation();
        onChangeTitle({ id, title });
    }

    return (
        <components.HistoryEntryTitle data-active={isActive ? true : undefined} onClick={() => onSelectSession(entry)}>
            {
                !isEdit ?
                    (

                        <>
                            <span>{title}</span>
                            {
                                isActive && (
                                    <components.IconsContainer>
                                        <IconPencil onClick={(e) => {
                                            e.stopPropagation();
                                            setIsEdit(!isEdit);
                                        }} style={{ marginRight: '1.5rem' }} size={22} />
                                        <IconTrash size={22} onClick={(e) => wrapDelete(e, entry.sessionId)} />
                                    </components.IconsContainer>
                                )
                            }

                        </>
                    ) :
                    (
                        <components.ChangeTitleContainer>
                            <InputGroup onClick={(e) => e.stopPropagation()}>
                                <Form.Control
                                    type="text"
                                    value={title}
                                    onChange={(e) => setTitle(e.target.value)}
                                />
                            </InputGroup>
                            <components.IconsContainer>
                                <IconCheck onClick={(e) => {
                                    changetTitle(e, entry.sessionId)
                                }} style={{ margin: '0 1.5rem' }} size={22} />
                                <IconX size={22} onClick={(e) => wrapDelete(e, entry.sessionId)} />
                            </components.IconsContainer>
                        </components.ChangeTitleContainer>
                    )
            }
        </components.HistoryEntryTitle>)
}

export const ChatHistory = (props: ChatHistoryProps) => {
    const sessions = props.sessions;
    const [search, setSearch] = useState<string | undefined>();

    useEffect(() => {
        setSearch("");
    }, [props.open])

    const renderTitlesGroup = (items) => {
        return items.map(t => {
            return (<HisotryEntry
                onChangeTitle={props.onChangeTitle}
                isActive={props.activeChat?.id == t.sessionId}
                key={t.sessionId}
                entry={t}
                deleteHistory={props.deleteHistory}
                onSelectSession={props.onSelectSession} />)
        })
    }

    const getGroupHistory = (items) => {
        const getPreviousDate = (shift) => {
            const d = new Date();
            d.setDate(d.getDate() - shift);
            return d;
        }
        const source = (items || []).filter(i => i.createdAt).sort((a, b) => b.createdAt - a.createdAt);
        const result: any = {
            today: [],
            last7: [],
            last30: [],
            older: [],
        };
        const yesterday = getPreviousDate(1);
        const last7 = getPreviousDate(7);
        const last30 = getPreviousDate(30);
        for (const item of source) {
            if (item.createdAt > yesterday) {
                result.today.push(item);
            }
            else if (item.createdAt > last7) {
                result.last7.push(item);
            }
            else if (item.createdAt > last30) {
                result.last30.push(item);
            }
            else {
                result.older.push(item);
            }
        }
        return result;
    }

    const filterdModels = props.sessionTitles && search ? props.sessionTitles.filter(m => m.title.includes(search)) : (props.sessionTitles || []);
    const groupedItems = getGroupHistory(filterdModels);

    return (
        <components.Container>
            <div className='history-scroll-block'>

                <Tabs
                    defaultActiveKey="history"
                    id="uncontrolled-tab-example"
                    className="mb-3"
                >
                    <Tab eventKey="history" title="History">
                        <SearchContainer>
                            <InputGroup style={{ marginBottom: '15px' }}>
                                <InputGroup.Text>
                                    <IconSearch />
                                </InputGroup.Text>
                                <Form.Control
                                    type="text"
                                    placeholder="Search..."
                                    value={search}
                                    onChange={(e) => setSearch(e.target.value)}
                                />
                            </InputGroup>
                        </SearchContainer>
                        <div className='history-block'>
                            {
                                groupedItems.today.length ? (
                                    <>
                                        <span>Today</span>
                                        {renderTitlesGroup(groupedItems.today)}
                                    </>
                                ) : null
                            }
                            {
                                groupedItems.last7.length ? (
                                    <>
                                        <span>Previous 7 Days</span>
                                        {renderTitlesGroup(groupedItems.last7)}
                                    </>
                                ) : null
                            }
                            {
                                groupedItems.last30.length ? (
                                    <>
                                        <span>Previous 30 Days</span>
                                        {renderTitlesGroup(groupedItems.last30)}
                                    </>
                                ) : null
                            }
                            {
                                groupedItems.older.length ? (
                                    <>
                                        <span>Older</span>
                                        {renderTitlesGroup(groupedItems.older)}
                                    </>
                                ) : null
                            }
                        </div>
                    </Tab>
                    <Tab eventKey="sessions" title="Sessions">
                        {
                            sessions?.length ? (
                                sessions.map(a => {
                                    const model = props.models.find(x => x.Id == a.ModelAgentId);

                                    return (
                                        <components.HistoryEntryContainer key={a.Id}>
                                            <div>
                                                {!isClosed(a) ?
                                                    <components.FlexSpaceBetween>
                                                        <Badge bg="success">Active</Badge>
                                                        <components.CloseBtn onClick={() => props.onCloseSession(a.Id)}>Close</components.CloseBtn>
                                                    </components.FlexSpaceBetween> : null}
                                            </div>
                                            <components.HistoryItem>
                                                <components.ModelName data-rh={abbreviateAddress(a.Id, 3)} data-rh-negative>{model?.Name}</components.ModelName>
                                                <components.Duration>{((a.EndsAt - a.OpenedAt) / 60).toFixed(0)} min</components.Duration>
                                            </components.HistoryItem>
                                        </components.HistoryEntryContainer>
                                    )
                                }
                                ))
                                : <div>You have not any sessions</div>
                        }
                    </Tab>
                </Tabs>
            </div>
        </components.Container>
    )
}