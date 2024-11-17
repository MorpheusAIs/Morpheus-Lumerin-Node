export const getSessionsByUser = async (url, user) => {
    if(!user || !url) {
      return;
    }

    const getSessions = async (user, offset, limit) => {
      try {
        const path = `${url}/blockchain/sessions/user?user=${user}&offset=${offset}&limit=${limit}`;
        const response = await fetch(path);
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
      console.log("Getting session for user: ", user, offset, limit)
      const sessionsRes = await getSessions(user, offset, limit);
      sessions.push(...sessionsRes);

      if(sessionsRes.length != limit) {
        all = true;
      }
      else {
        offset++;
      }
    }

    return sessions;
  }