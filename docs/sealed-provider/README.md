# RFC: one-click sealed (TEE-style) providing from the desktop app on a Mac

Status: draft design proposal with a partial implementation.

**This is not secure yet. Do not use it to make a no-peek claim to consumers.** Nothing in this folder is a finished product. The engine that exists today lacks several of the protections the design requires (see [STATUS-AND-KNOWN-ISSUES.md](STATUS-AND-KNOWN-ISSUES.md)), the code signer is currently the operator, and the independent review of the spawn module found five defects that are still open.

## Problem

A provider on the Morpheus network can read every prompt and answer that passes through the model server it runs. Consumers who care about privacy have no way to check otherwise. The upstream project's own TEE work targets servers (Secret VM on AMD SEV). A Mac running Apple silicon has no equivalent hardware enclave for GPU inference, so a Mac provider cannot offer that guarantee today.

## What the user experiences

The target flow: a provider opens the desktop app, goes to the P-Node wizard, and clicks once. The wizard registers the provider, brings up a user-space tunnel to a relay (no administrator prompt), downloads the model, starts the sealed engine, lists the model on chain with a sealed tag, and posts a bid. The provider is the app wallet, directly. The design goal is that no step needs a terminal or manual key or hash handling; where a step still does, that is a defect against this proposal.

What has been shown: on 2026-09-29 a sealed answer from a small model appeared end to end in the app through this path, using the first-generation engine, not the design in this RFC. That engine does not have the exception-port guard, so its no-peek property does not hold against the provider's own root user.

## How it works

The design replaces "trust the provider" with "trust a specific, hash-pinned program that the provider cannot read into". The paragraphs below summarize it; [ARCHITECTURE.md](ARCHITECTURE.md) has the components and [THREAT-MODEL.md](THREAT-MODEL.md) has the limits.

The engine is a signed macOS app that holds a per-boot App Attest key. Apple's attestation names the engine's code-directory hash, so a consumer can check which program answered. Apple refused to attest on every Mac booted below Full Security that was tried (4 of 4, measured on the first-generation engine).

The engine runs with the hardened runtime and no `get-task-allow`, so another process, root included, cannot obtain its task port. An exception-port guard added first thing in `main` makes any fault end the process with exit 113 and no crash report, so a crash cannot leak memory to a launcher or to a crash-reporting file.

The model server is not inside the engine. Splash (Qwen3.8-27B) and mlx-serve (a Qwen-family MLX model) run as child processes that ship inside the engine bundle. The engine spawns each child suspended, with a kernel-enforced launch requirement that names the child's code-directory hash, and checks the child's hash and code-signing status word before resuming it.

Each child connects back to a private rendezvous socket. The engine accepts the connection only after the audit token of the process that sent the first byte matches the child it spawned. A same-user attacker who squats on or replaces that socket gets a refusal, not a channel.

The child applies a sandbox to itself that denies network and file writes, reads only a manifest of weight files whose SHA-256 values were checked, and persists nothing derived from prompts. Splash needs no change for this; mlx-serve needs a 53-line patch that accepts requests as passed file descriptors instead of listening on a port.

The consumer's node opens a session by sending a challenge. The engine returns an App Attest assertion that binds the challenge to a fresh X25519 key, to the backend (child) identity, and to the weights hash. Frames after that are ChaCha20-Poly1305 under keys derived from that exchange. The provider's node and the relay forward ciphertext.

The weights hash must equal the hash in the model's on-chain listing. For models the app release names as verified, it must also equal a hash compiled into the app.

Remaining trust is stated, not hidden: Apple's platform, the pin publisher until reproducible builds exist (R11), and the memory safety of the pinned children, because one child serves every session of its listing.

Mac-side measurements so far show no speed cost: in the spike, sealed-style runs matched the release binaries within run-to-run noise (table in [STATUS-AND-KNOWN-ISSUES.md](STATUS-AND-KNOWN-ISSUES.md)).

## Status

| Item | Built | Measured | Not built |
|---|---|---|---|
| App wizard, direct-provider mode, user-space tunnel to a relay | yes (first-generation) | one end-to-end sealed answer, 2026-09-29 | v2 catalog entries (Splash, mlx-serve) |
| Engine v1: App Attest, envelope, in-memory weights hash, in-process MLX | yes | answers end to end | R1 guard is not linked into this build |
| R1 exception-port guard | C source exists | prototype behaviour measured (exit 113, no report) | wiring into engine and children |
| R2/R3 spawn module and 19-arm suite | yes | 17 PASS, 2 SKIP, 0 FAIL on the test Mac | five BROKEN review findings listed, fixes not yet re-reviewed |
| Splash and mlx-serve running sealed | spike only | no loss of speed vs release | engine-side channel protocol, tokenizer in the engine |
| Child-side R4-R9 (Metal integrity, sandbox, manifest, config, persistence audit, restart) | no | no | all of it |
| Static mlx-serve build | analysis and patch drafted, never built | no | the build |
| R11 reproducible build, second-party reproduction | build script exists | no second party has reproduced | independent reproduction |
| Same-uid adversary arms | yes (several) | pass as an ordinary user | root run pending |

## Files

- [ARCHITECTURE.md](ARCHITECTURE.md): components, data flow, requirements R1-R11 in one line each.
- [THREAT-MODEL.md](THREAT-MODEL.md): adversary, what is promised, stated residuals, assumptions A-1 to A-21 with status.
- [STATUS-AND-KNOWN-ISSUES.md](STATUS-AND-KNOWN-ISSUES.md): what is built, what review found, measured numbers, what is missing.
- [CREDITS.md](CREDITS.md): the upstream work this composes.
