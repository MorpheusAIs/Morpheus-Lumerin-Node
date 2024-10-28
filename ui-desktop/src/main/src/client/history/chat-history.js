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

const saveTitle = async (data) => {
    const db = dbManager.getDb();
    const collection = db.collection('chat-title');
    await collection.insert({ _id: data.sessionId, ...data });
}

const deleteTitle = async (id) => {
    const db = dbManager.getDb();
    await db.collection('chat-title').remove({ _id: id })
    await db.collection('chat').remove({ sessionId: id })
}

const updateChatTitle = async ({ id, title }) => {
    const db = dbManager.getDb();
    const collection = db.collection('chat-title');
    const { _id, ...stored } = await collection.findAsync({ _id:  id })
    const data = { ...stored[0], title };
    await collection.update({ _id: id }, data);
}

export default { getChatHitory, saveChatHistory, getTitles, saveTitle, deleteTitle, updateChatTitle };