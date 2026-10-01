// R1 exception-port guard (DESIGN-sealed-backends v2.26, R1).
//
// One object file, shared by the engine and both children, byte-identical in
// release and test builds (R1 build gate). The guard's only action on an
// exception message is _exit(113): a process that faults dies before any
// crash report can snapshot it, and no exception message reaches a port that
// root or the engine could use to read the process.
#ifndef SEALED_GUARD_H
#define SEALED_GUARD_H

#include <mach/mach.h>

#ifdef __cplusplus
extern "C" {
#endif

// Fatal types the guard catches: BAD_ACCESS, BAD_INSTRUCTION, ARITHMETIC,
// EMULATION, SOFTWARE, BREAKPOINT, CRASH, CORPSE_NOTIFY.
#define SEALED_GUARD_MASK 0x247eu
// Every other valid type: SYSCALL, MACH_SYSCALL, RPC_ALERT, RESOURCE, GUARD.
// Set to no port, so nothing inherited survives and a non-fatal notice does
// not end a healthy process (operator, 2026-09-30).
#define SEALED_NULL_MASK 0x1b80u
#define SEALED_VALID_MASK 0x3ffeu

// One old task-level port the guard's swaps returned, combined by the port
// name it had in this task (send rights to one port share one name, so a port
// returned partly by each swap is one entry).
typedef struct {
    mach_port_name_t name;
    exception_mask_t mask;
    exception_behavior_t behavior;
    thread_state_flavor_t flavor;
} sealed_old_port_t;

#define SEALED_MAX_OLD_PORTS 16

// Installs the guard: receive port, reader thread, task ports for the guard
// mask, no port for the null mask, every thread's ports cleared for every
// valid type, then the fatal-signal handlers. Writes the old ports it
// replaced (combined by name) to `old` and their count to `*old_count`, and
// deallocates the send rights they named. Returns 0, or a nonzero step number
// for the call that failed; the caller then ends the process with _exit(113).
int sealed_guard_install(sealed_old_port_t *old, unsigned *old_count);

// Mask gate: every type in the guard mask is the guard's port with the
// guard's behavior; every type in the null mask has no port. 0 if so.
int sealed_guard_readback(void);

// Every thread's exception ports are null for every valid type (threads that
// exit between task_threads and the query are skipped). 0 if so.
int sealed_guard_threads_clear(void);

// 1 if the behavior would deliver a task or thread right or thread state to
// whoever holds the port; 0 only for the identity-protected behaviors.
int sealed_behavior_exposes(exception_behavior_t behavior);

// The engine's rule for the ports launchd handed it (R1): every old port must
// match an entry of `expected` exactly (mask, behavior, flavor) and no old port
// may expose rights, whatever the table says. 0 if so.
int sealed_old_ports_match(const sealed_old_port_t *old, unsigned old_count,
                           const sealed_old_port_t *expected, unsigned expected_count);

#ifdef __cplusplus
}
#endif

#endif
