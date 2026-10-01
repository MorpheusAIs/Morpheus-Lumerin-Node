# Session and audio stream hardening (change notes)

## Session front door (MORRPC)

- `session.request` requires the supplied key to derive to the supplied wallet
  address (`PubKeyBytesToAddr` must equal `user`).
- The stored `user:<addr>` record is a cache, not a trust anchor. Every handler
  that authenticates a session owner (`session.prompt`, `session.report`,
  `session.prompt.stream.*`) re-derives the stored key and requires it to match
  the session's on-chain user before verifying the signature
  (`sessionUserPubKey`). A stale or mismatched record cannot authenticate.
- `AddUser` allows overwrite. Ownership is proven by the caller before the
  write, so a later write for the same address can only come from the wallet
  owner; this also lets an owner replace a stale record on the next
  `session.request` without operator intervention.
- Session expiry uses server time (`sessionExpiredByServerTime`); the client
  timestamp is not used for expiry.
- A session is only usable on the provider it was opened with
  (`session.ProviderAddr()` must equal this node's wallet).

## Audio stream sandbox (MORRPC)

- The client `streamid` is accepted for wire compatibility (existing consumers
  reuse it on `chunk`/`end`) but is validated (`^[A-Za-z0-9_-]{1,64}$`) and
  used only as an in-memory key scoped to the owning session
  (`sessionID:streamID`). It never influences a filesystem path.
- Temp files are created with `os.CreateTemp` under a fixed directory
  (`morpheus-audio-streams`) and verified with `PathUnderDir`.
- Bounds per stream: `totalchunks` in `1..4096`, `filesize` in `1..256 MiB`,
  decoded chunk `<= 8 MiB`, cumulative bytes `<= filesize`, at most 4 concurrent
  streams per session. Chunks must arrive in order.
- Remove/process only operate on managed paths
  (`SafeRemoveManagedAudioPath` / `IsManagedAudioPath`); a `FilePath` on
  `session.prompt` that is not a managed path is refused.

## Consumer side

- `sendStreamStart` continues with the stream ID the provider acknowledges.
  Providers echo the client ID today, so this is a no-op; it keeps consumers
  compatible if a provider ever issues its own.

## HTTP (`:8082`, operator-authenticated)

- Audio upload `createTempFile` uses `os.CreateTemp`; only an allowlisted
  audio extension is preserved from the client filename so the engine can
  infer the format.
- Chat file storage builds paths from sanitized ids and rejects path elements.
- `POST /ipfs/add` is unchanged: it is operator-authenticated and reads a file
  the operator chose on their own machine (MorpheusUI model pinning).

## Compatibility

All MORRPC changes are wire-compatible with unpatched consumers. Providers can
upgrade first; consumers can upgrade at their own pace.

## Ops

- Run the process as a non-root user.
- Consumers should not expose inbound MORRPC `:3333` publicly.

## Tests

- `internal/lib`: `PathUnderDir`, `PubKeyBytesToAddr`
- `internal/storages`: `AddUser` overwrite semantics
- `internal/proxyapi`: stream id validation, session-scoped lookup, bounds,
  ordered append, managed-path remove, stored-key derivation check, server-time
  expiry helper
- `internal/chatstorage`: id sanitization
