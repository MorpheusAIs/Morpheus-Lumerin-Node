## Full Flow Example for an Agent User Using the Proxy-Router API

### How to run

1. **Run the proxy-router.**
2. **Update `config.js`** with the desired values:
   - **`proxyRouterUrl`** – The HTTP proxy-router URL.
   - **`modelId`** – The local model to use.
   - **`agentUsername`, `agentPassword`, `agentPerms`** – The agent user data to be created.
3. **Run `node ./agent-user-request.js`.** An agent user request will be sent to the proxy-router.
4. **Approve the agent user creation** using an admin user in the proxy-router
   (`http://localhost:8082/swagger/index.html#/auth/post_auth_users_confirm`).
5. **Run `node agent-run.js`.**