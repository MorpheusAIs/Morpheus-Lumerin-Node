const { proxyRouterUrl, agentPassword, agentPerms, agentUsername } = require("./config");

const requestAgentUser = (username, password, perms, allowances) => {
  return fetch(`${proxyRouterUrl}/auth/users/request`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password, perms, allowances }),
  });
};

const MOR_TOKEN = "0x34a285a1b1c166420df5b6630132542923b5b27e";

(async () => {
    const allowances = {
        [`${MOR_TOKEN}`]: 10 * 10 ** 18,
    };
    
    const response = await requestAgentUser(agentUsername, agentPassword, agentPerms, allowances);
    const data = await response.json();
    console.log(data);
})();