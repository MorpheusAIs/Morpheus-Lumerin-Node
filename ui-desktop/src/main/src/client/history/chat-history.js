import dbManager from '../database';

const getChatHitory = async (sessionId) => {
    const collection = await dbManager.getDb().collection('chat').findAsync({ id: sessionId });
    return collection;
}

const saveChatHistory = ({ sessionId, messages }) => {
    const db = getDb();
    const collection = db.collection('chat');
    collection.insert(
        {
            id: sessionId,
            messages: messages,
        });
}

export default { getChatHitory, saveChatHistory };