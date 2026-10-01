# impl/NOTES.md — sealed_spawn.c implementation notes

Task: `TASK-impl.md` (2026-09-30). Scope: the engine-side spawn-time port and
handler parts of R1, R2 (spawn + rendezvous), R3 (pinned-code checks). The
child's own guard already exists in `impl/src/sealed_guard.c` and is not
duplicated here. Interface: `include/sealed_spawn.h`, unchanged.

Sources used, in order of authority: `context/SPEC-R1-R3.md`,
`context/EVIDENCE-2026-09-30.md` (E1, E2, E3, E4, E5, E11 + E11 addendum),
`context/probes/` (probe programs and their raw results).

## Smoke result

`./run-on-test-mac.sh impl/smoke` (the test Mac, macOS 27.0, 2026-09-30), run
twice, both stable:

- Positive arm: pinned child (`build/childA`, cdhash `734fa6908b59a52786c90836fc7bc112364aaf0a`)
  verifies → `result=OK(0)`, marker written (child ran), `PASS`.
- Negative arm: `build/childB` (different CD hash) spawned with the launch
  requirement naming childA's hash → kernel refuses before the child runs
  (marker absent), module refuses `SEALED_SPAWN_CDHASH(15)`, and the reader
  serviced the killed child's exception message (`exception_seen=1`),
  `PASS`.
- `make test exit: 0`. Module compiles warning-free with
  `clang -std=c17 -Wall -Wextra -Werror`.

Status words recorded for R3-T1's phase recording: suspended phase
`0x22011311`, running phase `0x22011311` (same word at both measured points
for this smoke child; not a measured release word — record-only mode).

## Clause → function/line table

Lines are in `impl/src/sealed_spawn.c`.

| Spec clause | Where |
|---|---|
| R1: spawn-time port, one receive right per child generation, no-senders request | `sealed_spawn_child` L344-353 (`mach_port_allocate`, `mach_port_request_notification`) |
| R1: deallocate our send right as soon as the spawn call returns | L444 (`mach_port_deallocate` after `posix_spawn`) |
| R1: dedicated reader thread, started before the port is set, distinct from the guard thread | L355-364 (`pthread_create` before the attributes/spawn), `spawn_port_reader` L159 |
| R1: handler on a message releases thread/task rights unused; E11 measured rule (reply KERN_SUCCESS; SIGKILL through the lifecycle on any fatal type except EXC_CRASH; never KERN_FAILURE) | `service_exc_msg` L129-148, called from `spawn_port_reader` L169/L180 |
| R1: SIGKILL through the lifecycle state's not-yet-reaped check (pid cannot be reused) | `kill_if_not_reaped` L112-120; `sealed_child_kill` L204 |
| R1: reader exits when the child's guard drops the last send right | `spawn_port_reader` L171-187 (no-senders notification → bounded drain, then exit); see deviation D4 |
| R2: one fresh empty 0700 directory per spawn, exclusive, collision fails | L306-319 (`confstr` `_CS_DARWIN_USER_TEMP_DIR` + `mkdtemp`) |
| R2: private Metal cache dir `$TMPDIR/metal-cache` and rendezvous socket `$TMPDIR/rv` (AF_UNIX, backlog 1) created before spawn | L322-343 (`mkdir`, `socket`/`bind`/`listen`) |
| R2: `POSIX_SPAWN_CLOEXEC_DEFAULT`; child inherits only fd 0-2 on /dev/null | L384-388 flags; L409-418 file actions (`addopen` 0/1/2 `/dev/null`) |
| R2: all signals reset to default, empty mask | L384-388 (`setsigdefault` `sigfillset`, `setsigmask` `sigemptyset`) |
| R2: environment exactly `HOME=<dir>` `TMPDIR=<dir>` | L425-429 (`envp`) |
| R2: argv the sealed form, built only from the spec | L286-291 (`spec->argv` passed through) |
| R2: spawn suspended (`POSIX_SPAWN_START_SUSPENDED`), own process group (`POSIX_SPAWN_SETPGROUP`) | L384-386 |
| R2: working directory is TMPDIR | L416 (`posix_spawn_file_actions_addchdir`) |
| R2: launch requirement, kernel-enforced, naming the pinned CD hash (A-13) | `amfi_launch_constraint_set_spawnattr`; attach failure fatal → `SEALED_SPAWN_REQUIREMENT` (batch-1 fix — it returned `SEALED_SPAWN_ATTR` before; r2-inheritance F3) |
| R2: spawn-time port on the guard mask, no port on the null mask, one `posix_spawnattr_setexceptionports_np` call per mask | L391-398 |
| R2: verify, then list the child's descriptors from outside, then resume | suspended phase L449-467; `fds_ok` L244-267; resume L478-481 |
| R2: tolerate exactly one AF_UNIX descriptor in the pre-resume listing | `fds_ok` L255-263 |
| R2: rendezvous — one connection, first byte within the deadline, then token read before anything is sent | L483-533 (poll-bounded accept and read; deadline = `spec->rendezvous_ms`) |
| R2: token computed when read names the sender of the byte (E3); pid must be the spawned child's | L534-539 (`LOCAL_PEERTOKEN`, `token_pid`) |
| R3: `csops(CS_OPS_CDHASH)` before resume equals the pin | L449-454 |
| R3: `csops(CS_OPS_STATUS)` full word vs the measured word for the phase (0 = record only) | L456-467 suspended, L542-551 running; `sealed_status_ok` L189-196 |
| R3: never reap during the spawn window | `child_dead_unreaped` L101-110 (`waitid` `WNOWAIT|WNOHANG` peek only) |
| R3: identity checks after the rendezvous go through the audit token, never a bare pid | L540-551 (`csops_audittoken`, E4: stale token → `ESRCH`) |
| R3: record the status words per phase | L465 (`status_seen_suspended`), L550 (`status_seen_running`) |
| Result codes, fail closed with the matching code | every failure path returns the mapped `SEALED_SPAWN_*` (batch-1 fixes: posix_spawn failure returned ATTR, now EXEC via the stage map — failure-paths F1; `reaped` now only after a real reap — F2; spawn port released once on every path — F3/F4); `spawn_fail` (kill + reap + directory removed) |

## Deviations from the spec text (with reasons)

1. **The spawn-time handler replies; R1's literal "never replies" is not
   implemented.** R1 says the thread never replies; EVIDENCE E11 (later,
   measured, and R1 itself defers release to R1-T5) measured that a
   never-reply handler holds children unreaped and that the working rule is:
   release rights, reply `KERN_SUCCESS`, with `SIGKILL` first for every
   fatal-type exception except `EXC_CRASH`. Implemented per E11 (S1: refused
   launch reaped 5/5 at ~0.31 s, 0 reports; S2: null-store crash reaped 5/5,
   0 reports). The 2026-09-30 smoke run confirmed the same on the refusal
   path (`exception_seen=1`, no hang after the fix).

2. **`posix_spawn_file_actions_addchdir` instead of `addchdir_np`.** macOS 26
   renamed the `_np` function (SDK deprecation note: "posix_spawn_file_
   actions_addchdir(3) has replaced posix_spawn_file_actions_addchdir_np(3)");
   same semantics. Keeping `_np` would not compile `-Werror`.

3. **`mach_port_deallocate` instead of `mach_port_destroy`** for releasing
   the spawn port's receive right (macOS 12 deprecation). Deallocating the
   last receive right destroys the port with the same release/wake effect.

4. **The reader does not exit immediately on the no-senders notification.**
   On the refusal path the kernel tears down the killed-at-spawn task and the
   no-senders notification can be delivered *before* the refused child's
   exception message lands; a reader that exits on the notification leaves
   the exception unanswered and the child held in exit (observed 2026-09-30:
   child in `UE` while the driver hung in the reap). The reader instead
   drains for a bounded time (2 s) after a notification, answering any
   exception message that follows, then exits. The notification itself is
   still registered as R1 requires.

5. **Non-absolute `spec->path` refused (`SEALED_SPAWN_ATTR`).** The spawn
   chdirs the child into its fresh directory before exec, so a relative
   executable path would silently resolve in the wrong directory (measured:
   first smoke run returned `ENOENT` from `posix_spawn`).

6. **`SEALED_CS_KILLED` (`0x01000000u`) is defined in the module**, not the
   header (the header cannot be changed). A refused child's measured status
   is `0x23000201` where a live suspended child reads `0x22000201`
   (results-round3). The bit is refused in addition to `sealed_status_ok`,
   so a kernel-killed child fails even in record-only mode (`measured == 0`).

7. **The rv listening socket is closed and unlinked at the end of successful
   verification.** In the full engine the listening socket closes only after
   the config frame is sent; this module's contract ends at the verified
   channel, and closing at that point is strictly more conservative (a
   foreign connect after verification can no longer reach the listener).

## Unimplemented here (belongs to other modules)

- R2's post-acknowledgement steps: reading the child's acknowledgement,
  the authoritative post-ack descriptor comparison with socket identity
  (`PROC_PIDFDSOCKETINFO` peer check, R2-T4(i)), R5's probe files, and the
  config frame. The channel protocol owns them; this module only delivers
  `child->channel_fd` and the recorded status words.
- R2's generation selection, manifest reads, and restart bounds; R9's
  timeouts; R1's engine self-check, launch-state check, constructor, and
  hypervisor check; R7's runtime re-check. The child's guard exists
  separately (`sealed_guard.c`).
- R2's contingency variables are not in force for this build: the spawn
  environment is exactly `HOME` and `TMPDIR`. If R4-T4 later shows Metal
  reads `MTL_SHADER_CACHE_SIZE` pre-`main`, the envp construction at
  L425-429 gains that one fixed variable.

## Measured during fix batch 1 (the test Mac, 2026-10-01)

- **FIXES 10** — symptom (old code): `r1_t5_crash_before_rv` and `r2_t4_crash`
  returned `RENDEZVOUS(18)` at the full 5000 ms deadline for a child that
  crashed before touching `$TMPDIR/rv`, with `exception_seen=1` and reaped.
  Measurement: `exception_seen=1` in the *failing* run proves the spawn-port
  reader serviced the crash exception promptly — so the reader's drain and the
  no-senders registration (FIXES's suspected items 4-5/F5) were NOT the
  proximate cause on this OS: the notification registered before any sender
  exists did not preempt the exception message. The actual cause was the
  rendezvous loop: the accept (and byte) loop called poll(2) with the whole
  remaining rendezvous window as one timeout and re-checked
  `child_dead_unreaped` only at the loop top, so a child dying during the
  window was noticed only when that single poll finally timed out, and the
  timeout branch returned RENDEZVOUS with no death check. Fix: 100 ms poll
  slices with a child-death re-check per slice in both loops. Measured on the
  fixed code: `DIED(20)` in 224 ms, reaped, marker written, control accepted.
- **FIXES 8** — first attempt (size-probe `proc_pidinfo(pid,
  PROC_PIDLISTFDS, 0, NULL, 0)`) is not usable on this OS: the probe returned
  ≤ 0 for a live child, refusing every spawn with `FDS(17)` (whole suite red).
  Measured working shape: start at 64 entries and grow (×2) while the buffer
  comes back exactly full, accepting only a short-of-buffer listing; fail
  closed if it never does.
- **FIXES-2 4(a)** — MEASURED (tests/probes/nosend_probe.c, the test Mac,
  2026-10-01): a `MACH_NOTIFY_NO_SENDERS` notification registered on a fresh
  port with zero send rights IS delivered immediately (zero-timeout receive
  returns it, msgh_id 70 = `MACH_NOTIFY_NO_SENDERS`); with the task holding a
  MAKE_SEND send right on the port it is NOT delivered within 200 ms.
  Consequence for the module's old order: the notification was registered
  *before* the MAKE_SEND insert, so every reader got the notification as its
  first message and drained-then-exited at spawn time — a refused child's
  later exception message could only be caught inside that 2 s drain window,
  never by a parked reader. Fix (shipping order): allocate → insert
  MAKE_SEND → register the notification → start the reader; our send right is
  still deallocated as soon as the spawn call returns (R1), leaving the child
  (through the kernel's rights copy) the only sender. This partially answers
  the open question below (whether the notification ever fires): it fires
  immediately when no senders exist, and not while ours is held.
- **FIXES-2 9** — the reader's stop flag in `sealed_child_kill` moved from
  before the SIGKILL to after the grace loop (just before the first
  `reader_wait`; the pre-pid and post-reap paths also stop the reader now,
  since they no longer inherit the early stop). Effect: the reader stays live
  through the SIGKILL and the `KILL_GRACE_NS` window, so an exception message
  arriving in either is answered on the measured answer-first path; the cost
  on the plain (no-exception) kill path is one bounded reader round.
  Kill-path time on a live held child (the `r1_t5_crash_before_rv` control,
  new `MEASURED kill` line): **30 ms before → 232 ms after** (before: stop
  already set, reader torn down immediately; after: the reader answers
  through the window and exits within one 200 ms round). Crash-arm spawn
  times unchanged (DIED in 218 ms before → 225 ms after). Full suite
  (`SLOT=fix ./run-on-test-mac.sh tests`) and smoke
  (`SLOT=fix ./run-on-test-mac.sh impl/smoke`) both `make test exit: 0` on
  the test Mac 2026-10-01.
- **FIXES-2 7** — MEASURED (r2_t4_bad_requirement arm, per-variant codes on
  the result line; the test Mac 2026-10-01): the launch-requirement SPI accepts
  every malformed shape up to the shipping length — 1-byte truncation
  `STATUS(16)`, 4-byte truncation `STATUS(16)`, 82 bytes of 0xFF `STATUS(16)`,
  82 zeroed bytes `STATUS(16)` — and the kernel then refuses the exec'd image
  (the module's status-word check sees the exec-refusal word), child never
  ran (marker absent, asserted per variant). Only the oversize shape (1 MiB
  of 0xFF) is rejected by the SPI itself: `REQUIREMENT(13)` with no child
  created (`reaped` legitimately 0; the arm no longer claims "reaped" on the
  SPI-reject path). So a byte-level SPI rejection is reachable only via
  oversize bytes; every accepted-but-malformed shape is caught by the kernel
  stage, and the arm now FAILs on any variant that ends in
  `SEALED_SPAWN_OK` or a written marker (seal failure).

## Fix batch 1 record (change + arm evidence, runs on the test Mac 2026-10-01)

| FIXES | Change in impl/src/sealed_spawn.c | Arm evidence (before → after) |
|---|---|---|
| 1 | posix_spawn failure maps to `SEALED_SPAWN_EXEC` (`failed_code` stage at the spawn call site, mapped in the shared failure exit) | pending: r2_t4_exec_fail arm (absolute nonexistent path; before ATTR(12) → after EXEC(14)) |
| 2 | requirement-attach failure maps to `SEALED_SPAWN_REQUIREMENT` (`failed_code` at the SPI call) | pending: garbage-requirement-bytes arm (before ATTR(12) → after REQUIREMENT(13)) |
| 3 | `reaped = 1` only after the `waitpid` loop actually reaped; an unreapable child stays kill-guarded | invariant (reaped ⇒ child gone) asserted in the pending conn-die arm; the wedge branch itself is not arm-triggerable unprivileged (needs an unanswered exception message) |
| 4 | reader context carries the port by value; `spawn_port` guarded by the mutex; static registry; `sealed_child_kill` stops the reader (`stop`, checked each 200 ms round) and drains it with a bound before returning | pending: exercised by every arm once the churn arm lands; before: UAF/cross-generation port-name race |
| 5 | drain receive gets `MACH_RCV_TIMEOUT` with a millisecond timeout (`READER_DRAIN_ROUND_MS` = 200); timed-out rounds continue to the drain deadline | covered by the crash arms after the fix (reader services the message; `exception_seen=1`) |
| 6 | `release_spawn_port` on every path exactly once, including failures before the pid exists (pre-pid branch of `sealed_child_kill`) | pending: port-leak arm deferred to the churn arm; verified by the full suite not leaking (all arms reaped) |
| 7 | `POSIX_SPAWN_SETSIGDEF | POSIX_SPAWN_SETSIGMASK` added to the spawn flags | pending: r2_t1_sig_reset arm (arm ignores SIGSEGV + blocks SIGUSR1 before spawn; before: child reports SIG 11 IGN / nonzero mask → FAIL, after: DFL / mask 0 → PASS) |
| 8 | `fds_ok` grows the listing while the buffer comes back exactly full; fails closed; size-probe unusable (measured above) | SKIP-stated arm planned: no unprivileged way to open extra fds in a suspended child pre-exec (needs the A-14 test hook) |
| 9 | `realpath` the fresh directory once (canonical everywhere: dir, env, rv socket); refuse on `child->dir` snprintf truncation and on an rv path ≥ `sun_path` (104) | live discriminator: `r2_t1_inheritance` failed on old code (cwd=/private/var/... vs HOME/TMPDIR=/var/...); passes once that arm's own `line()` harness bug is fixed (pending arm-step fix) |
| 10 | measured cause + fix — see the two entries above (rendezvous loops: 100 ms poll slices + per-slice death re-check) | `r1_t5_crash_before_rv` and `r2_t4_crash`: FAIL (RENDEZVOUS, exception_seen=1, full 5000 ms) → PASS (DIED(20), 224/298 ms, reaped) |
| 11 | byte loop: per-round child-death re-check; `POLLHUP`/`POLLERR` end the connection (DIED if the child is dead, else RENDEZVOUS); EOF read → death check | pending: r2_t4_conn_die arm (child connects, closes, exits; before RENDEZVOUS → after DIED) |
| Makefile | stale duplicate `test:` target already absent (verified: one `test:` in tests/Makefile); FAIL-line gate added (any output line starting with FAIL forces nonzero exit) | proven: seeded `zzz_seed_fail` arm printing FAIL with exit 0 → `make test exit: 2`; seed removed |

## Unmeasured / open (for R1-T5 and the design team)

- The no-senders notification registered on the spawn port with the port
  itself as the notify port is a design choice; whether the notification is
  ever delivered on macOS 27 (the kernel's own notification send-once could
  count as a sender) is unmeasured. It is not load-bearing for correctness:
  the reader also exits when the lifecycle destroys the port.
- Message shapes other than the measured ones (S1: EXC_CRASH id 2405 size 84;
  S2: EXC_BAD_ACCESS) have not been exercised through this module; the fuzz
  arm of R1-T5 must cover them.
- Concurrent children sharing nothing but the module (each with its own port,
  reader, and lifecycle state) are untested here; R1-T5's "two children fault
  together" arm covers that.
- `mach_port_deallocate` on the receive right while the reader is blocked in
  `mach_msg` — observed to wake the reader (reader exits, receive returns
  non-success) in the smoke runs, but a dedicated probe has not confirmed the
  kernel's behavior for a port destroyed mid-exception-delivery; the observed
  2026-09-30 hang (child held in `UE` while the driver was alive) suggests a
  port destroyed mid-delivery can wedge the hold, which is why the lifecycle
  now answers-first and destroys only as a fallback.
- If a child is still unreapable after the destroy fallback, `sealed_child_kill`
  leaves it unreaped (still kill-guarded) rather than blocking the caller
  forever; the engine's supervision must observe such a child. Unmeasured on
  this OS (all smoke paths reaped).

## Interface notes

- `sealed_child_t.lock` is a heap `pthread_mutex_t` allocated by
  `sealed_spawn_child`; the engine frees it (and the struct) at generation
  end. The mutex is initialized there; `sealed_child_kill` tolerates `lock`
  being NULL.
- The smoke driver bounds each arm with its own 45 s alarm (`ARM TIMEOUT`
  + exit 124), well under run-on-test-mac.sh's 600 s cap, so a hang is a named
  failure.
- The 82-byte requirement bytes are the measured `NSTask.launchRequirementData`
  encoding (E1); the smoke generates them with `probes/a13.swift` via
  `swiftc` on the test Mac. Building a `LaunchCodeRequirement` has no C API
  (no C header ships); the engine build must keep that step.
