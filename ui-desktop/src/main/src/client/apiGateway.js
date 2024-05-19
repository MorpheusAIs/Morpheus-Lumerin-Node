import config from '../../config'

const getAllModels = async () => {
    try {
        const path = `${config.chain.localProxyRouterUrl}/blockchain/models`
        const response = await fetch(path);
        const data = await response.json();
        return data.models;
    }
    catch (e) {
        console.log("Error", e)
        return [];
    }
}

const getBalances = async () => {
    try {
        const path = `${config.chain.localProxyRouterUrl}/blockchain/balance`
        const response = await fetch(path);
        const data = await response.json();
        return data;
    }
    catch (e) {
        console.log("Error", e)
        return [];
    }
}


export {
    getAllModels,
    getBalances
}