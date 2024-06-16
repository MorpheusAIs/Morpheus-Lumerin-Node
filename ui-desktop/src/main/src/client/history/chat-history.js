import dbManager from '../database';

const getChatHitory = async (sessionId) => {
    const collection = await dbManager.getDb().collection('chat').findAsync({ _id: sessionId });
    return collection;
}

const saveChatHistory = async ({ sessionId, messages }) => {
    const db = dbManager.getDb();
    const collection = db.collection('chat');
    await collection.insert(
        {
            _id: sessionId,
            messages: messages,
        });
}

export default { getChatHitory, saveChatHistory };