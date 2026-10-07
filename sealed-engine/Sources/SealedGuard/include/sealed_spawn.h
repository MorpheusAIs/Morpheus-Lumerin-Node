// R2 spawn and rendezvous, R3 pinned-code checks (DESIGN-sealed-backends v2.26).
#ifndef SEALED_SPAWN_H
#define SEALED_SPAWN_H

#include <mach/mach.h>
#include <stdint.h>
#include <sys/types.h>

#ifdef __cplusplus
extern "C" {
#endif

#define SEALED_CDHASH_LEN 20

// Code-signing status flags (kern/cs_blobs.h; values checked against the SDK).
#define SEALED_CS_VALID 0x00000001u
#define SEALED_CS_GET_TASK_ALLOW 0x00000004u
#define SEALED_CS_INSTALLER 0x00000008u
#define SEALED_CS_FORCED_LV 0x00000010u
#define SEALED_CS_INVALID_ALLOWED 0x00000020u
#define SEALED_CS_HARD 0x00000100u
#define SEALED_CS_KILL 0x00000200u
#define SEALED_CS_ENFORCEMENT 0x00001000u
#define SEALED_CS_RUNTIME 0x00010000u
#define SEALED_CS_PLATFORM_BINARY 0x04000000u
#define SEALED_CS_DEBUGGED 0x10000000u

#define SEALED_CS_REQUIRED \
    (SEALED_CS_VALID | SEALED_CS_HARD | SEALED_CS_KILL | SEALED_CS_RUNTIME | SEALED_CS_ENFORCEMENT | SEALED_CS_FORCED_LV)
#define SEALED_CS_FORBIDDEN                                                                          \
    (SEALED_CS_GET_TASK_ALLOW | SEALED_CS_DEBUGGED | SEALED_CS_INVALID_ALLOWED | SEALED_CS_PLATFORM_BINARY | \
     SEALED_CS_INSTALLER)

// What the engine knows about one pinned child before spawning it.
typedef struct {
    const char *path;                       // the child executable
    const char *const *argv;                // the sealed argv form (argv[0] first), NULL-terminated
    uint8_t cdhash[SEALED_CDHASH_LEN];      // pinned CD hash
    const uint8_t *requirement;             // LaunchCodeRequirement bytes naming `cdhash` (A-13)
    size_t requirement_len;
    uint32_t status_suspended;              // measured full status word, suspended phase (0 = record only)
    uint32_t status_running;                // measured full status word after the rendezvous (0 = record only)
    unsigned rendezvous_ms;                 // deadline for the connection and its first byte (KK3)
} sealed_child_spec_t;

// Lifecycle state shared by the spawn path, the spawn-time exception handler
// and every kill the engine addresses to the child (R1, R3): SIGKILL goes to
// the pid only while `reaped` is 0, so the pid cannot have been reused.
typedef struct sealed_child {
    pid_t pid;
    int reaped;
    int exception_seen;                     // the spawn-time port received a message
    mach_port_t spawn_port;                 // receive right for this generation's spawn-time port
    int channel_fd;                         // verified connection (the channel), -1 until verified
    char dir[256];                          // the child's fresh HOME/TMPDIR
    uint32_t status_seen_suspended;         // recorded for R3-T1's phase recording
    uint32_t status_seen_running;
    void *lock;                             // pthread mutex (opaque here)
} sealed_child_t;

// Result codes: 0 = verified and serving; anything else = refused, child killed.
enum {
    SEALED_SPAWN_OK = 0,
    SEALED_SPAWN_DIR = 10,          // fresh directory, metal-cache or rv socket
    SEALED_SPAWN_PORT = 11,         // spawn-time port or its handler thread
    SEALED_SPAWN_ATTR = 12,         // spawn attributes, exception ports or file actions
    SEALED_SPAWN_REQUIREMENT = 13,  // launch requirement could not be attached (A-13)
    SEALED_SPAWN_EXEC = 14,         // posix_spawn failed
    SEALED_SPAWN_CDHASH = 15,       // suspended child's CD hash is not the pin
    SEALED_SPAWN_STATUS = 16,       // status word: required/forbidden flags or measured word
    SEALED_SPAWN_FDS = 17,          // descriptor table before resume
    SEALED_SPAWN_RENDEZVOUS = 18,   // no connection and first byte within the deadline
    SEALED_SPAWN_TOKEN = 19,        // audit token: wrong pid, stale, or wrong code identity
    SEALED_SPAWN_DIED = 20,         // the child died or raised an exception before verification
};

// Spawns and verifies one child generation. On success `child` holds the
// verified channel; on failure the child has been killed through its
// lifecycle state and reaped, and the directory removed.
int sealed_spawn_child(const sealed_child_spec_t *spec, sealed_child_t *child);

// Kills the child if (and only if) it has not been reaped, then reaps it.
void sealed_child_kill(sealed_child_t *child);

// Checks a status word against the required and forbidden flags and, when
// `measured` is nonzero, against the measured word for the phase. 0 if it passes.
int sealed_status_ok(uint32_t status, uint32_t measured);

#ifdef __cplusplus
}
#endif

#endif
