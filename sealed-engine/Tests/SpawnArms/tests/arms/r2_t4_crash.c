// Arm r2_t4_crash: SPEC-R1-R3.md R2-T4/R3-T1 "died before verification"
// shape — a child that crashes before any rendezvous. The child's first main
// instruction writes the marker (it ran), then it faults (EXC_BAD_ACCESS, in
// the guard mask). The contract's spawn-time port (sealed_child_t
// exception_seen: "the spawn-time port received a message") must receive the
// exception — this is R1's spawn-time port covering the pre-guard window,
// observable through this module's header — and the module must refuse with
// SEALED_SPAWN_DIED ("the child died or raised an exception before
// verification"), never succeed. Positive control in the same program: the
// pinned good child is accepted through the same spawn path, so a module that
// refuses everything (or accepts a crashing child) fails the arm.
#include <stdio.h>
#include "harness.h"

static char why[256];

int main(void) {
    arm_begin("r2_t4_crash");
    const char *ok_child = arm_env("CHILD_OK"), *ok_pin = arm_env("HASH_OK");
    const char *crash = arm_env("CHILD_CRASH"), *crash_pin = arm_env("HASH_CRASH");
    const char *req = arm_env("REQ_OK"), *work = arm_env("WORKDIR");
    if (!ok_child || !ok_pin || !crash || !crash_pin || !work)
        arm_fail("env missing (CHILD_OK/HASH_OK/CHILD_CRASH/HASH_CRASH/WORKDIR)");

    // Positive control: pinned child-ok accepted through the same path.
    sealed_child_spec_t spec;
    sealed_child_t ch;
    char marker[512];
    snprintf(marker, sizeof marker, "%s/m.t4c.control", work);
    if (arm_build_spec(&spec, ok_child, ok_pin, req, marker, 5000) != 0)
        arm_fail("could not build control spec");
    if (arm_spawn(&spec, &ch) != SEALED_SPAWN_OK || !arm_marker_written(marker))
        arm_fail("control failed: pinned child-ok not accepted");

    // Subject: the crashing child, pinned by its own hash (requirement bytes
    // name child-crash itself).
    snprintf(marker, sizeof marker, "%s/m.t4r.crash", work);
    if (arm_build_spec(&spec, crash, crash_pin, arm_env("REQ_CRASH"), marker, 3000) != 0)
        arm_fail("could not build crash spec");
    int rc = arm_spawn(&spec, &ch);
    if (rc != SEALED_SPAWN_DIED) {
        snprintf(why, sizeof why, "crashing child: expected SEALED_SPAWN_DIED, got %s(%d)",
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (!arm_marker_written(marker))
        arm_fail("crashing child never ran (marker absent) — the fault arm is vacuous");
    if (ch.exception_seen == 0)
        arm_fail("spawn-time exception port did not receive the child's crash (exception_seen=0)");
    arm_pass("pre-rendezvous crash -> DIED, spawn-time exception port received the message (exception_seen=1), child had run; pinned child-ok control accepted");
    return 0;
}
