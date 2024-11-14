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
    let sessions: any[] = [];
    let all = false;

    while (!all) {
      console.log("Getting session for user: ", user, offset, limit)
      const ids = await getIds(user, offset, limit);
      const results = await Promise.all(ids.map(id => getSessioInfo(id)));
      sessions.push(...results);

      if(ids.length != limit) {
        all = true;
      }
      else {
        offset++;
      }
    }

    return sessions;
  }