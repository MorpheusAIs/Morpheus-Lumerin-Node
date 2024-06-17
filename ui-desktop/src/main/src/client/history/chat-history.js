import dbManager from '../database';

const getChatHitory = async (sessionId) => {
    return await dbManager.getDb().collection('chat').findAsync({ sessionId });
}

const saveChatHistory = async ({ sessionId, messages }) => {
    const db = dbManager.getDb();
    const collection = db.collection('chat');

    const items = await getChatHitory(sessionId);

    if (!items.length) {
        await collection.insert({ sessionId, messages });
        return;
    }

    await collection.update({ sessionId }, { messages, sessionId }, { upsert: true });
}

const getTitles = async () => {
    return await dbManager.getDb().collection('chat-title').findAsync({});
}

const saveTitle = async ({ sessionId, title }) => {
    const db = dbManager.getDb();
    const collection = db.collection('chat-title');
    await collection.insert({ _id: sessionId, title: title });
}

export default { getChatHitory, saveChatHistory, getTitles, saveTitle };