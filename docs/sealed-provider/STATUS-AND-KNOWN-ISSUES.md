# Status and known issues

Dates are 2026-09-29 to 2026-10-01. "The test Mac" is an Apple silicon Mac running macOS 27.0 (build 26A428) with SIP enabled. The spike's speed measurements ran on a different Mac.

Nothing here is a security claim. It is a list of what exists, what was measured, and what is wrong.

## Built

- **First-generation engine.** A signed app with a per-boot App Attest key (code-directory-hash opt-in), a per-session X25519 exchange bound to the consumer's challenge, ChaCha20-Poly1305 frames, and an in-process MLX model host. It reads the model directory once into memory from an allowlist of file names, computes the weights hash over exactly those buffers, and builds the model and tokenizer from the same buffers, so the operator cannot swap files between hashing and loading. It serves loopback HTTP for the node. It has no child processes.
- **R1 exception-port guard.** A C source file implements the guard (mask `0x247e` on the fatal types, `0x1b80` null mask on the rest, a reader thread started before the port is set, `_exit(113)` as its only action), plus a mask read-back gate and a per-thread check. Prototype behaviour was measured: a guarded process exits 113 and leaves no crash report on a main-thread fault, a worker-thread fault and a code-signature kill, where unguarded controls wrote reports. Signals sent from outside (`SEGV`, `ABRT`) still end a guarded process with no report. **The guard is not part of the built engine app**: the engine project file lists only the Swift engine sources, so the shipped first-generation engine does not have it. Until it is wired in, that engine's no-peek property does not hold against the provider's own root user.
- **R2/R3 spawn module and suite.** About 780 lines of C implement the engine-side spawn: suspended spawn, spawn-time exception port with a reader thread, a launch requirement naming the child's code-directory hash, rendezvous with audit-token verification, status-word and hash checks, and a kill-and-reap routine. The arm suite has 19 arms. On the test Mac it was run twice on 2026-10-01 (a staging copy and the committed copy): **17 PASS, 2 SKIP, 0 FAIL**. Both seeded mutants (one that notifies before the port is inserted, one that skips signal reset) fail their intended arm.
- **Skipped arms.** `r2_t3_config_frame` is skipped because config-frame refusals are the child's duty and the module has no way to inject or observe a config frame through its header. `r3_t2_replaced_helper` is skipped because the module refuses a spawn with no launch-requirement bytes, and the arm needs an engine build with the requirement off.
- **Reproducible build script.** It pins the Xcode and Metal toolchain builds, package versions and an MLX patch, builds at fixed paths, and has a signing mode and a no-signature verify mode so a second party can compare. No second party has reproduced a build.

## Measured: spike speed (Splash and mlx-serve sealed-style vs release)

Settings: 256 output tokens, temperature 0, same prompt, medians, arms of one group run back to back. Absolute Splash speed varied between 53 and 75 tok/s across the session with identical binaries, following the Mac's other GPU load. Compare within a row group only.

| Engine | Arm | Decode tok/s | Warm time to first token |
|---|---|---|---|
| Splash 27B, engine-level harness | release engine, socketpair | 52.9 | 0.18 s |
| | source build, socketpair + sandbox (no network, no writes) | 52.8 | 0.19 s |
| mlx-serve Flash-Next, HTTP | release 26.9.2 | 84.1 | 0.07-0.20 s |
| | source HEAD | 87.1 | 0.12-0.19 s |
| | source HEAD, sealed flags + sandbox (loopback only) | 84.5 | 0.13-0.21 s |
| mlx-serve, sealed fd-passing mode | source HEAD + patch, sandbox with no network at all | 104.9 | 0.13-0.15 s |

The three pre-set falsifiers (needs network or disk writes; needs a reachable port; materially slower) were not falsified. mlx-serve in sealed mode needs `--prefix-cache-disk 0 --log-file off`, because the default log holds the first 80 characters of every prompt and the disk cache stores prompt tokens.

## Independent review of the spawn module

Eight reviews from five model families were merged into a ledger of **53 findings**. Counts by severity: **5 BROKEN, 22 GAP, 24 NIT, 2 ungraded** (reviewer gradings split). By status: 6 measured true, 2 measured false, 45 open. Several severities for the earlier entries are derived from the entry text rather than a recorded field. A set of measured facts overrides any reviewer claim that contradicts them.

The five BROKEN findings:

- **L04** `mach_port_deallocate` never destroys the spawn port's receive right. Measured (below).
- **L16** The same defect on every failure exit between the send-right insert and the post-spawn deallocate. Measured.
- **L01** The reap branch overwrites a true "reaped" flag with false when `waitpid` returns ECHILD. Open.
- **L02** `sealed_child_kill` is not idempotent: a second call on the same child stalls about 12 s. Open; measured as the stall in the numbers below.
- **L03** The window between reaping and setting the flag is outside the lock, so a late kill can signal a freed pid. Open, argued from code, not measured.

Measured facts that override reviewer claims:

- **Receive-right leak.** On a receive-only name `mach_port_deallocate` returns 17 (`KERN_INVALID_RIGHT`) and the receive right remains. On receive plus send, one or two deallocates remove the send right only. `mach_port_mod_refs(RECEIVE, -1)` destroys it. 100 iterations of the module's allocate, insert, notify, deallocate sequence leak 100 port names; with a real destroy, 0. The module calls deallocate on the spawn port at four places (lines 285, 513, 523, 643). Line 643 on the success path must keep its deallocate so the no-senders notification fires when the serving child dies; the destroy belongs in the release routine and the early failure exits.
- **Double kill.** The first `sealed_child_kill` takes about 220 ms and reports reaped. A second call on the same child takes about 12,000 ms and leaves reaped at 0 on a gone pid. With `SIGCHLD` set to ignore, the kill takes about 12,100 ms and leaves reaped at 0. A copy of the module that treats ECHILD as reaped makes all three fast and correct. Per-phase timing reproduced the second-call figure (12,041 ms).
- **Dead peer token.** After a child writes its first byte and exits, reading the local peer token fails (EINVAL while exiting, ENOTCONN as a zombie and after reap). No token is produced for a dead peer, so the "OK for a dead child" claim is false; a final death re-check remains cheap defence in depth.
- **Status words.** No arm sets the suspended or running status word (record-only, measured 0, in every run). The word-compare path is therefore untested end to end (finding L49).
- **Stale note.** An old failing line for `r1_t5_crash_before_rv` in the suite notes is stale; that arm passes.
- **Refuted claim.** A reviewer claim that a spawn-file-action symbol used by the module does not exist is refuted by the build.

The remaining findings are mostly GAP and NIT items about bounds, error handling in cleanup, reader-thread lifetime, unvalidated rendezvous byte value and token length, and missing test coverage for kill, reap, port accounting and rendezvous arms. A fix batch is grouped by area in the ledger. The ledger lists fixes as candidates and its status column marks 45 findings open; no fixed version has been re-reviewed.

## Child code-signing status words

Measured on the test Mac by spawning signed copies suspended and reading the status word before resume. Both real children, Splash and mlx-serve, signed ad-hoc with the hardened runtime, read **0x22011311**, identical to the test child.

| Signing form | Word | Against R3's masks |
|---|---|---|
| ad-hoc as linked (linker signature) | 0x22020201 | fails required mask (no hardened-runtime bits) |
| ad-hoc, hardened runtime | 0x22011311 | passes |
| hardened runtime + `get-task-allow` | 0x22011315 | fails forbidden mask |
| `-o runtime,library` | 0x22013301 | fails required mask: sets `CS_REQUIRE_LV` (0x2000) instead of `CS_FORCED_LV` (0x10) |

Two bits that the spec's named masks do not list, `CS_NO_UNTRUSTED_HELPERS` (0x02000000) and `CS_SIGNED` (0x20000000), appear in every word and are part of the full-word pin. These rows are from scratch copies. mlx-serve is still dynamically linked against non-system libraries in that form. The final pin must be measured from the release-signed, statically linked children, and a different macOS build needs its own rows.

## Same-user adversary arms (root run pending)

All of these ran as an ordinary user, so the attacker had the engine's own uid. The same binaries split privileges when run as root; that run has not been done. Each arm has a positive control in the same run and, where recorded, a mutant that fails it.

| Arm | Result |
|---|---|
| Attacker connects first and sends a wrong byte | refused with TOKEN code in 234 ms; child reaped; directory removed. Mutant without the token check: accepted in 17 ms |
| Attacker connects first and stays silent | refused with RENDEZVOUS code in about 5.2 s (deadline 5 s); no hang. Mutant with a stretched deadline: harness timeout at 60 s |
| Attacker replaces the rendezvous socket with its own listener | refused with RENDEZVOUS in about 5.2 s; the attacker learns only the child's fixed first byte (0x53) |
| Child connects first, attacker second (two arms; in the first the scheduler ordered the attacker first and the spawn was refused with RENDEZVOUS in 221 ms, so the second arm gates the attacker on the child's connection existing) | child verified in 15 ms with channel peer equal to the child; attacker's connect found the socket already removed |
| Early `SIGCONT` raced into the suspended window, 20 spawns | OK 20 of 20, status word unchanged, all reaped |
| `SIGKILL` while suspended or before connect | module side at most 330 ms per spawn |

With these arms added, the suite read 18 PASS, 2 SKIP in the run that included the signal arm. A timing-instrumented arm that kills a child after OK and then calls kill twice exceeded its 28 s budget (28,735 ms); 12,041 ms of that was the second kill call, which is the double-kill defect above, not an arm defect.

## Not built

- **Channel protocol between engine and children.** The module hands back a verified channel descriptor. What runs on it (Splash's token-ID protocol; mlx-serve's control socket and per-request socketpairs) is not implemented in the engine. Its framing, request IDs and protocol-error handling are specified.
- **Engine front end for Splash.** Tokenizer, chat template with reasoning-effort setting, and detokenization in the engine are specified, not built.
- **Child-side R4 to R9.** Metal integrity, the child's self-applied sandbox, the weights manifest with verified hashes and closed-window loading, argument allowlist and compiled-out logging, the persistence audit, and child restart handling. The spike ran the children under an external sandbox profile, which is deprecated and was for measurement only.
- **Static mlx-serve build.** A read-only analysis lists every non-system library the current build needs (several dynamic libraries from MLX and llama.cpp, plus one image library) and a patch was drafted. No build was run and the patch is not applied. Open risks recorded there: runtime-compiled Metal kernels depend on assumption A-6, and the build identity fingerprint must be re-homed.
- **R10 consumer-side v2 checks.** Backend ID parsing, per-build tables and the verified-model weights pin. The first-generation counterpart exists in the node; the v2 checks are specified only.
- **R11 reproducible builds closed.** A script exists. A second party must rebuild the engine and each child and match the published pin; until then A-10 is trust in the publisher.
- **Per-pair assumption measurements.** A-4, A-5, A-6, A-9, A-15, A-16, A-17 and A-20 are unmeasured (see [THREAT-MODEL.md](THREAT-MODEL.md)). No pair can be labelled tested.
- **Wizard entries for the v2 models**, and the app's own Direct Pay display and secure-badge copy fixes noted during development.

## What this means for use

Do not advertise a no-peek guarantee to consumers from any of this. The first-generation engine lacks the guard; the v2 pieces exist as a spawn module with known defects and no child-side protections; and the code signer is the operator.
