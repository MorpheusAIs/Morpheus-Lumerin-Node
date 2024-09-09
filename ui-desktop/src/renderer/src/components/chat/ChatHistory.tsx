import { IconPencil, IconTrash, IconSearch } from "@tabler/icons-react";
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
    sessions: any[];
    models: any[];
    sessionTitles: { sessionId: string, title: string, createdAt: any }[]
}

export const ChatHistory = (props: ChatHistoryProps) => {
    const sessions = props.sessions;
    const [search, setSearch] = useState<string | undefined>();

    const wrapDelete = (e, id) => {
        e.stopPropagation();
        props.deleteHistory(id);
    }

    useEffect(() => {
        setSearch("");
    }, [props.open])

    const renderTitlesGroup = (items) => {
        return items.map(t => {
            return (<components.HistoryEntryTitle onClick={() => props.onSelectSession(t)}>
                <span>{t?.title || ""}</span>
                <div>
                    <IconPencil style={{ marginRight: '1rem' }} size={18}></IconPencil>
                    <IconTrash size={18} onClick={(e) => wrapDelete(e, t.sessionId)}></IconTrash>
                </div>
            </components.HistoryEntryTitle>)
        })
    }

    const getGroupHistory = (items) => {
        const getPreviousDate = (shift) => {
            const d = new Date();
            d.setDate(d.getDate() - shift);
            return d;
        }
        const source = (items || []).filter(i => i.createdAt).sort((a, b) => b.createdAt - a.createdAt);
        console.log("🚀 ~ getGroupHistory ~ source:", source)
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