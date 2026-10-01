# Session and audio stream hardening (change notes)

## Session front door (MORRPC)

- `session.request` requires the supplied key to derive to the supplied wallet
  address (`PubKeyBytesToAddr` must equal `user`).
- The stored `user:<addr>` record is a cache, not a trust anchor. Every handler
  that authenticates a session owner (`session.prompt`, `session.report`,
  `session.prompt.stream.*`) re-derives the stored key and requires it to match
  the session's on-chain user before verifying the signature
  (`sessionUserPubKey`). A stale or mismatched record cannot authenticate.
- Records learned from inbound peers are written through `AddUserBound`, which
  itself refuses any key that does not derive to the address, so the
  key<->wallet invariant is enforced by the store and not only by the call
  site. Overwrite is allowed because only the wallet owner can satisfy the
  bound write; this also lets an owner replace a stale record on the next
  `session.request` without operator intervention.
- The unbound `AddUser` remains for consumer-side provider records only: a
  contract provider's record legitimately holds the owner's key under the
  contract address after the caller validated it against the resolved signer.
- Session expiry uses server time (`sessionExpiredByServerTime`); the client
  timestamp is not used for expiry.
- A session is only usable on the provider it was opened with. `isSessionValid`
  (prompt/stream) checks the on-chain `ProviderAddr()`; `session.report` checks
  the cached session's `ProviderAddr`. The check (`servesProvider`) accepts the
  node's own key address directly, and a contract provider whose on-chain
  `owner()` is this node's key via the shared `ProviderAuthResolver` (cached
  per provider). Fails closed on an unknown node key or resolver error.

## Audio stream sandbox (MORRPC)

- The client `streamid` is accepted for wire compatibility (existing consumers
  reuse it on `chunk`/`end`) but is validated (`^[A-Za-z0-9_-]{1,64}$`) and
  used only as an in-memory key scoped to the owning session
  (`sessionID:streamID`). It never influences a filesystem path.
- Temp files are created with `os.CreateTemp` under a fixed directory
  (`morpheus-audio-streams`) and verified with `PathUnderDir`.
- Bounds per stream: `totalchunks` in `1..4096` and `<= filesize`, `filesize`
  in `1..256 MiB`, decoded chunk `<= 8 MiB`, cumulative bytes `<= filesize`,
  and `end` requires exactly `filesize` bytes. At most 4 concurrent streams per
  session and 2 GiB of declared size in flight across all sessions. Chunks must
  arrive in order.
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
  the operator chose on their own machine (MorpheusUI model pinning). Accepted
  residual: if `:8082` admin credentials are compromised the endpoint can read
  any path the process can; the same credentials already expose
  wallet-spending endpoints, so the control for this surface is keeping `:8082`
  off the public network and running non-root, not a path allowlist.

## Known follow-ups (out of scope for this hotfix)

- Consumer-side provider key records (`proxy_sender.go` `InitiateSession`) are
  verified self-consistently against the response signature but not bound to
  the provider address; binding needs the resolved signer for contract
  providers. Not part of the inbound provider surface this hotfix addresses.

## Compatibility

All MORRPC changes are wire-compatible with unpatched consumers. Providers can
upgrade first; consumers can upgrade at their own pace.

## Ops

- Run the process as a non-root user.
- Consumers should not expose inbound MORRPC `:3333` publicly.

## Tests

- `internal/lib`: `PathUnderDir`, `PubKeyBytesToAddr`
- `internal/storages`: `AddUserBound` key<->address enforcement, `AddUser`
  overwrite semantics
- `internal/proxyapi`: stream id validation, session-scoped lookup, per-stream
  and global bounds, ordered append, managed-path remove, stored-key derivation
  check, server-time expiry helper
- `internal/chatstorage`: id sanitization
