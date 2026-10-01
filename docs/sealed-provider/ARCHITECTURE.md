# Architecture

Draft design. Parts marked "built" exist in source; parts marked "specified" do not yet. [STATUS-AND-KNOWN-ISSUES.md](STATUS-AND-KNOWN-ISSUES.md) is the authority on which is which.

## Components and data flow

```
 consumer app / router                         provider Mac
 ------------------------                      ----------------------------------------------------
                                               desktop app (P-Node wizard, app wallet)
 challenge --------\                                    |
                    \   relay (ciphertext only)         v
                     +--------------------------> node (proxy-router, sealed lane)
 sealed frames  <----/                                  |  loopback HTTP: attest, session, probe, infer
                                                        v
                                               sealed engine (attested, hardened runtime)
                                                 |  spawn suspended + launch requirement
                                                 |  rendezvous socket, token check
                                       +---------+----------+
                                       v                    v
                              Splash child            mlx-serve child
                              (Qwen3.8-27B)           (accept-fd mode)

 chain: model listing -> declares the weights hash the session must carry
```

The paragraphs below follow the arrows.

**Desktop app and wizard.** The app's P-Node wizard walks the provider through registration, model selection, price and concurrency, listing and bidding. The provider is the app wallet, directly; the wizard does not ask for a separate provider wallet. A user-space tunnel connects the app to a relay, so no administrator prompt is needed and the provider needs no inbound port. Built in the first generation; the v2 catalog entries are specified.

**Node (proxy-router, sealed lane).** The node is the upstream proxy-router with a sealed lane added. It talks to the engine over loopback HTTP on four endpoints: `/attest`, `/session`, `/probe`, `/infer`. It admits sessions on the basis of an on-chain check and a soft concurrency cap. The node and the relay see ciphertext, request sizes and timing, nothing else.

**Sealed engine.** A signed macOS app with the hardened runtime and without `get-task-allow`. It holds a per-boot App Attest key, created with the code-directory-hash opt-in entitlement so that every attestation and assertion names the engine binary that produced it. It holds the session keys, decrypts and encrypts consumer frames, selects a backend, and manages child lifecycle. For Splash it also tokenizes, applies the chat template and detokenizes, because Splash's native protocol carries token IDs. The built engine today is a first-generation form with an in-process MLX model host; spawning children is specified (the spawn module is built and tested separately).

**Pinned children.** Splash (`serve-native`, token IDs over a socketpair using its SPLH v6 protocol) and mlx-serve (a patch adds `MLX_SERVE_ACCEPT_FD`: no listening socket, one socketpair per request passed with `SCM_RIGHTS` over a private control socket). Children ship in the engine bundle's helper directory, are ad-hoc signed with the hardened runtime, and are statically linked except for system libraries. Their code-directory hashes are compiled into the engine. A child exits when its channel closes.

**Spawn and rendezvous.** The engine spawns a child suspended, with minimal inheritance (descriptors 0-2 on `/dev/null`, signals reset, environment limited to a private home and temporary directory), with its own task exception ports set at spawn, and with a launch requirement naming the child's code-directory hash attached through the kernel's launch-constraint mechanism. Before resuming, it reads the child's code-directory hash and code-signing status word. After resuming, the child connects to a rendezvous socket in the fresh private directory and sends one byte. The engine accepts the channel only after the audit token of the process that sent that byte names the spawned child and passes the same checks.

**App Attest session keys.** For each consumer session the router sends a 32-byte challenge. The engine produces an App Attest assertion that binds that challenge to a fresh ephemeral X25519 public key. Both sides derive two directional keys with HKDF-SHA256 (salt: the challenge; info: a version string and both public keys). Each frame is ChaCha20-Poly1305 with a counter nonce and a direction label as associated data. The key never leaves the engine process.

**Backend identity and weights hash.** The attested session data carries a 32-byte backend ID (kind, the child's 20-byte code-directory hash, zero padding) and the weights hash, which occupies bytes 96 to 127 of that data. The child hashes the bytes it computes with; the engine binds that hash into the attested data. The router refuses a session whose backend ID is not one it pins, or whose weights hash differs from the listing.

**On-chain listing.** Each sealed model is listed in the registry with a tag marking it sealed on Apple platforms and with the weights hash the owner declares. The router reads the listing once per session. For models the app release names as verified, the release also carries a weights hash and the session must match both.

**Relay.** The relay forwards the tunnel's traffic. It sees ciphertext and timing. It cannot produce a valid assertion because it holds no attested key.

## Requirements R1-R11 (summary)

Each line is a one-line summary of a requirement in the design specification, not a claim that it is implemented.

- **R1** Every sealed process (engine and both children) installs an exception-port guard as its first act: fatal exception types go to a port whose only action is `_exit(113)`, all other types to no port, so no fault produces a report or reaches a launcher.
- **R2** Minimal inheritance at spawn and a verified channel: no inherited channel descriptor, fresh private directory, rendezvous socket, and the audit token of the first byte's sender must match the spawned child.
- **R3** Run only pinned child code: the child's code-directory hash must equal a compiled-in constant and its full code-signing status word must equal a measured per-build value before it is resumed.
- **R4** Metal integrity in children: metal libraries come from hashed bytes inside the signed executable, shader cache is switched off, and compile keys are on a fixed list.
- **R5** Each child applies a sandbox to itself with a profile compiled into the child, denying network and file writes, after the guard and rendezvous and before any model file is opened.
- **R6** Weights manifest and integrity: a manifest of every file the child opens, each with size and SHA-256, is verified, and the resulting hash is bound into the attested session.
- **R7** Child configuration: each child accepts only one exact argument form; disk prefix cache, file logging, prompt logging, metrics and download code are compiled out.
- **R8** Persistence audit: a named list of every site that could write prompt-derived bytes is checked by symbol scans and a closed-world trace.
- **R9** Child restart: a child ends through `_exit` when its channel closes, sockets suppress `SIGPIPE`, and the engine closes every descriptor for the old child before any respawn.
- **R10** Consumer (router): the router generates challenges, holds the pins, reads the chain through its own node and verifies every assertion, including the backend ID format and the weights hash.
- **R11** Build and reproducibility: pinned toolchain, no host-dependent code generation, and a second party's rebuild must match the published signed engine and each child pin.
