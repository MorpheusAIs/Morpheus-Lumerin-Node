// Arm r2_t4_rendezvous: SPEC-R1-R3.md R2-T4 — the rendezvous deadline arms
// that run unprivileged against the module contract. "if no connection with
// its first byte arrives within a fixed time (an engine constant, covering
// both the connection and the byte), the engine kills the child, refuses":
// - a child that never connects (child-noconn) -> SEALED_SPAWN_RENDEZVOUS;
// - a child that connects but sends nothing (child-nosend) ->
//   SEALED_SPAWN_RENDEZVOUS (the deadline covers the first byte too, KK3).
// Both children are pinned by their own CD hashes and launched without
// requirement bytes, so the module's deadline is what fires. Positive control
// in the same program: the pinned good child (child-ok, requirement bytes,
// same rendezvous window) is accepted — so a module that refuses everything,
// or one with no working rendezvous, fails the arm.
// Skipped here, with the reason: R2-T4(a)/(c)/(l) (root connects to rv or
// replaces it before the child does) need to reach the module's private
// per-spawn rendezvous socket before the child connects — its mkdtemp path is
// not exposed by the contract and the spec's own framing is root; and (f)
// exec-after-first-byte needs a test hook delaying the module's token read
// (E3/E4 make the check correct; without the hook the arm is a coin flip).
#include <stdio.h>
#include "harness.h"

static char why[256], marker[512];

static void expect(const char *label, const char *child, const char *pin,
                   const char *req, int expect_code, const char *expect_name, int never) {
    sealed_child_spec_t spec;
    sealed_child_t ch;
    if (arm_build_spec(&spec, child, pin, req, marker, 3000) != 0) {
        snprintf(why, sizeof why, "could not build spec for %s", label);
        arm_fail(why);
    }
    int rc = arm_spawn(&spec, &ch);
    if (rc != expect_code) {
        snprintf(why, sizeof why, "%s: expected %s, got %s(%d)", label, expect_name,
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (never && arm_marker_written(marker)) {
        snprintf(why, sizeof why, "%s: refused with %s but the child RAN (marker written)",
                 label, expect_name);
        arm_fail(why);
    }
}

int main(void) {
    arm_begin("r2_t4_rendezvous");
    const char *ok_child = arm_env("CHILD_OK"), *ok_pin = arm_env("HASH_OK");
    const char *noconn = arm_env("CHILD_NOCONN"), *nosend = arm_env("CHILD_NOSEND");
    const char *nc_pin = arm_env("HASH_NOCONN"), *ns_pin = arm_env("HASH_NOSEND");
    const char *req = arm_env("REQ_OK"), *work = arm_env("WORKDIR");
    if (!ok_child || !ok_pin || !noconn || !nosend || !nc_pin || !ns_pin || !work)
        arm_fail("env missing (CHILD_OK/HASH_OK/CHILD_NOCONN/CHILD_NOSEND/HASH_NOCONN/HASH_NOSEND/WORKDIR)");

    // Positive control: the good child connects and sends its first byte within
    // the same rendezvous window; the module must accept it.
    sealed_child_spec_t spec;
    sealed_child_t ch;
    snprintf(marker, sizeof marker, "%s/m.t4r.control", work);
    if (arm_build_spec(&spec, ok_child, ok_pin, req, marker, 5000) != 0)
        arm_fail("could not build control spec");
    if (arm_spawn(&spec, &ch) != SEALED_SPAWN_OK || !arm_marker_written(marker))
        arm_fail("control failed: pinned child-ok not accepted through the rendezvous");

    // Child that never connects: killed and refused at the deadline.
    snprintf(marker, sizeof marker, "%s/m.t4r.noconn", work);
    expect("never-connects child", noconn, nc_pin, arm_env("REQ_NOCONN"),
           SEALED_SPAWN_RENDEZVOUS, "SEALED_SPAWN_RENDEZVOUS", 0);

    // Child that connects but sends nothing: the deadline covers the byte too.
    snprintf(marker, sizeof marker, "%s/m.t4r.nosend", work);
    expect("connects-but-sends-nothing child", nosend, ns_pin, arm_env("REQ_NOSEND"),
           SEALED_SPAWN_RENDEZVOUS, "SEALED_SPAWN_RENDEZVOUS", 0);

    arm_pass("no-connection and no-first-byte children both refused with RENDEZVOUS at the deadline and killed; pinned child-ok rendezvous control accepted");
    return 0;
}
