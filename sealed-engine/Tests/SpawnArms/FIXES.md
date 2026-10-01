# Fix batch 1 for impl/src/sealed_spawn.c (verified by the orchestrator against the code)

Each item: fix it in impl/src/sealed_spawn.c, and make sure an arm in tests/arms/ fails on the old code and passes on the fixed code (add or extend an arm where none covers it). Reviews with full detail: review/failure-paths.md, review/r2-inheritance.md, review/r2r3-identity.md.

1. posix_spawn failure returns SEALED_SPAWN_ATTR, never SEALED_SPAWN_EXEC (L431-439: pid stays -1). Map the failing stage to its code. (failure-paths F1)
2. Launch-requirement attach failure returns SEALED_SPAWN_ATTR, not SEALED_SPAWN_REQUIREMENT (L404-409, L439). (r2-inheritance F3)
3. sealed_child_kill sets reaped=1 even when the bounded waitpid loop timed out (L224-235). Set it only after a real reap; an unreapable child stays kill-guarded, as the comment above the function promises. (failure-paths F2)
4. The reader thread holds a pointer to the child struct and rereads child->spawn_port each receive (L159-186) while sealed_child_kill writes it unlocked and the engine frees the struct: use-after-free and cross-generation port-name reuse. Pass the port by value, guard the field, and make sealed_child_kill stop and join (bounded) the reader before returning. (failure-paths F4)
5. The drain receive passes a timeout without MACH_RCV_TIMEOUT (L177-178), so it blocks forever; and READER_DRAIN_ROUND_NS (200000000) is passed where mach_msg takes milliseconds. Fix both.
6. The spawn-time port is never released on normal teardown or on failures before the pid exists (L210 early return; L218-223 only on the unreapable path). Release it on every path, once. (failure-paths F3, r2-inheritance F4)
7. Signal reset is inert: flags lack POSIX_SPAWN_SETSIGDEF and POSIX_SPAWN_SETSIGMASK (L382-388; sys/spawn.h:47-48). (r2-inheritance F1)
8. fds_ok's 64-entry PROC_PIDLISTFDS buffer truncates silently (L245-248); refuse or grow when full. (three reviews)
9. The fresh directory is used as /var/... while the child's working directory reads /private/var/...; R2 wants canonical paths: realpath the directory once after mkdtemp and use that everywhere. Also check snprintf truncation of child->dir and that the rv path fits sun_path (refuse with SEALED_SPAWN_DIR). (r2r3-identity D1, r2-inheritance further-1)
10. tests/arms/r1_t5_crash_before_rv.c fails: a child that crashes before connecting holds the module for the whole deadline and returns RENDEZVOUS instead of DIED promptly. Likely cause: items 4-5 plus the no-senders notification registered before any sender exists (failure-paths F5: if it fires immediately the reader drains and never answers). Find the actual cause with a measurement, record it in impl/NOTES.md, fix it, and make the arm pass.
11. First-byte loop: treat POLLHUP/POLLERR as end of connection and re-check child death each round (r2r3-identity D2).

Also: tests/Makefile's `test` target exists twice (lines ~34-47 are a stale partial copy); remove the stale copy, and make any output line starting with FAIL give a nonzero exit. Prove it: temporarily seed a FAIL and show make test exits nonzero, then remove the seed.

Do not change include/. Record each fix and its arm in impl/NOTES.md (update the deviations and the clause table; correct the claims the reviews showed were wrong).
