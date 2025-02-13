export const getSessionsByUser = async (url, user, headers) => {
    if(!user || !url) {
      return;
    }

    const getSessions = async (user, offset, limit) => {
      try {
        const path = `${url}/blockchain/sessions/user?user=${user}&offset=${offset}&limit=${limit}&order=desc`;
        const response = await fetch(path, {
          headers,
          method: 'GET',
        });
        const data = await response.json();
        return data.sessions;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    } 
    
    const limit = 50;
    let offset = 0;
    let sessions: any[] = [];
    let all = false;

    while (!all) {
      const sessionsRes = await getSessions(user, offset, limit);
      sessions.push(...sessionsRes);

      if(sessionsRes.length != limit) {
        all = true;
      }
      else {
        offset += limit;
      }
    }

    return sessions;
}

export const getBidsByModelId = async (url, modelId, headers) => {
  if(!modelId || !url) {
    return;
  }

  const getBidsByModels = async (modelId, offset, limit) => {
    try {
      const path = `${url}/blockchain/models/${modelId}/bids?offset=${offset}&limit=${limit}&order=desc`
      const response = await fetch(path, {
        headers,
      });
      const data = await response.json();
      return data.bids;
    }
    catch (e) {
      console.log("Error", e)
      return [];
    }
  }
  
  const limit = 50;
  let offset = 0;
  let bids: any[] = [];
  let all = false;

  while (!all) {
    const bidsRes = await getBidsByModels(modelId, offset, limit);
    bids.push(...bidsRes);

    if(bids.length != limit) {
      all = true;
    }
    else {
      offset += limit;
    }
  }

  return bids;
}

export const getBidInfoById = async (url, id, headers) => {
  try {
    const path = `${url}/blockchain/bids/${id}`
    const response = await fetch(path, {
      headers,
    });
    const data = await response.json();
    return data.bid;
  }
  catch (e) {
    console.log("Error", e)
    return undefined;
  }
}