## Environment Variables: 

`PROXY_STORE_CHAT_CONTEXT`

- **Type:** Boolean
- **Purpose:** Controls whether the proxy-router stores chat contexts locally.

#### When `PROXY_STORE_CHAT_CONTEXT` is set to `TRUE`

- **Local Storage of Chats:**
  - The proxy saves chat sessions to local files.
  - Chat histories are maintained, allowing for persistent conversations.
  
- **API Operations:** (check swagger - `/v1/chats`)
  - **Retrieve Chats:** Access stored chats via API endpoints.
  - **Update Titles:** Modify chat titles using API calls.
  - **Delete Chats:** Remove chat sessions through the API.

- **Using `chat_id`:**
  - Include a `chat_id` in the request header.
  - The proxy-router automatically injects the corresponding chat context.
  - **Request Simplification:** Only the latest message needs to be sent in the request body.

#### When `PROXY_STORE_CHAT_CONTEXT` is set to `FALSE`

- **No Chat Storage:**
  - The proxy-router does not save any chat sessions locally.
  
- **Client-Side Context Management:**
  - Clients must include the entire conversation history in each request.
  - The proxy-router forwards the request to the AI model without adding any context.

---

Ensure the proxy-router is restarted after changing the environment variable to apply the new configuration.