## Environment Variables:

Chat context is two independent flags. **Storing** history does not prepend it onto the next prompt.

### `PROXY_STORE_CHAT_CONTEXT` (default `true`)

Persist chats to local files for the history drawer and `/v1/chats` (see swagger). Frozen `false` in `-tee` images.

- **`true`:** save sessions locally; list / rename / delete via the chats API. Send a stable `chat_id` header if you want turns grouped into one conversation.
- **`false`:** nothing is persisted. Clients must send the full `messages[]` transcript each request.

### `PROXY_FORWARD_CHAT_CONTEXT` (default `false`)

Prepend stored history onto the outgoing prompt. Requires store on and a reused `chat_id`.

- **`false` (default):** the client owns the transcript. Leave this off for OpenAI-compatible clients that already send the full `messages[]` — otherwise context is duplicated.
- **`true`:** opt-in for clients that send only the latest turn (the bundled Desktop UI sets this). The router prepends stored user/assistant turns before the new prompt.

Restart the proxy-router after changing either variable.

### `PROXY_REDACT_SECRETS` (default `hybrid`)

Scrub wallet secrets out of prompts before they are signed and sent to a provider. The prompt still goes through — only the secret is replaced, with `[redacted:private-key]` or `[redacted:seed-phrase]`.

An Ethereum private key is 32 random bytes, and so is a tx hash, a block hash, and a session ID. Nothing about the string distinguishes them, so the mode picks which way to err.

- **`hybrid` (default):** redact bare 64-hex blobs and checksum-valid BIP-39 phrases always; redact `0x`-prefixed 64-hex only when a private-key cue (`private key`, `PRIVATE_KEY=`, `mnemonic`, `keystore`, …) sits within 120 characters. Keys get pasted bare and hashes get pasted with the prefix, so a user asking about a tx hash or session ID still gets an answer.
- **`strict`:** redact every 64-hex blob and every run of 12+ BIP-39 words, checksum or not. Nothing key-shaped escapes, but pasted tx hashes and session IDs are redacted too.
- **`off`:** no scrubbing.

Only prompts bound for a **remote** model are scrubbed — the bundled local model never leaves the machine. Addresses (20 bytes / 40 hex) are never redacted. A redaction logs a `WARN` with the count only, never the matched value.

## CapacityPolicy strategies (models-config.json):

#### `simple`

Assign a slot to each session upon initiation, blocking new sessions when all slots are occupied, regardless of activity level.

- Each new session consumes one slot from the total available slots N (`concurrentSlots`).

- Do not allow new sessions when slots_in_use >= N.

- Slots remain occupied until the user explicitly closes the session or times out.


#### `idle_timeout`

Free up slots occupied by inactive sessions by setting an idle timeout period.

Timeout is 15 minutes.

- If no prompt is received within the idle timeout period, mark the session as idle.

- Release the slot associated with the idle session, making it available for new users.
