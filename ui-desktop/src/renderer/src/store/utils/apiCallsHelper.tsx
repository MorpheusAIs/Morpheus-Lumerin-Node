export const getSessionsByUser = async (url, user) => {
    if(!user || !url) {
      return;
    }

    const getIds = async (user, offset, limit) => {
      try {
        const path = `${url}/blockchain/sessions/user/ids?user=${user}&offset=${offset}&limit=${limit}`;
        const response = await fetch(path);
        const data = await response.json();
        return data;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    } 

    const getSessioInfo = async (id) => {
      try {
        const path = `${url}/blockchain/sessions/${id}`;
        const response = await fetch(path);
        const data = await response.json();
        return data.session;
      }
      catch (e) {
        console.log("Error", e)
        return [];
      }
    }
    
    const limit = 50;
    let offset = 0;
    let resultIds: string[] = [];
    let all = false;

    while (!all) {
      const ids = await getIds(user, offset, limit);
      resultIds.push(...ids);
      if(ids.length != limit) {
        all = true;
      }
      else {
        offset++;
      }
    }

    const results = await Promise.all(resultIds.map(id => getSessioInfo(id)));
    return results;
  }