// Arm r2_t2_never_run: SPEC-R1-R3.md R2-T2 — "A child that fails verification
// never runs: a test-build child whose ... first instruction in main writes a
// marker never writes it when the engine refuses it". The module spawns
// suspended (POSIX_SPAWN_START_SUSPENDED) and verifies before resuming, so a
// child it refuses must be killed without its first main instruction running;
// the marker file is that child's first act in main. Two refusal shapes, each
// asserting the specific SEALED_SPAWN_* code and that the marker was never
// written. Both cases carry requirement bytes naming the spawned image's own
// CD hash (the module fails closed on a missing requirement), so the kernel
// admits the image and the module's pre-resume check is what refuses it.
// - a validly signed child with a different CD hash pinned as the good child's
//   hash -> SEALED_SPAWN_CDHASH ("suspended child's CD hash is not the pin");
// - a get-task-allow child pinned by its own hash -> SEALED_SPAWN_STATUS
//   (CS_GET_TASK_ALLOW is forbidden; R3-T1's gta shape through R2-T2's oracle).
// The module fails closed on a missing launch requirement (REQUIREMENT), so
// each case carries the requirement bytes naming the spawned image's own CD
// hash — the kernel admits the image, and it is the module's pre-resume check
// that must refuse it. Positive control in the same program (spec + task
// rules): the pinned good child is accepted and writes its marker, so a module
// that refuses everything (or accepts everything) fails this arm.
#include <stdio.h>
#include "harness.h"

static char why[256], marker[512];

static void refusal_case(const char *label, const char *child, const char *pin,
                         const char *req, int expect, const char *expect_name) {
    sealed_child_spec_t spec;
    sealed_child_t ch;
    if (arm_build_spec(&spec, child, pin, req, marker, 3000) != 0) {
        snprintf(why, sizeof why, "could not build spec for %s", label);
        arm_fail(why);
    }
    int rc = arm_spawn(&spec, &ch);
    if (rc != expect) {
        snprintf(why, sizeof why, "%s: expected %s, got %s(%d)", label, expect_name,
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (arm_marker_written(marker)) {
        snprintf(why, sizeof why, "%s: refused with %s but the child RAN (marker written)",
                 label, expect_name);
        arm_fail(why);
    }
}

int main(void) {
    arm_begin("r2_t2_never_run");
    const char *ok_child = arm_env("CHILD_OK"), *ok_pin = arm_env("HASH_OK");
    const char *alt = arm_env("CHILD_ALT"), *gta = arm_env("CHILD_GTA");
    const char *gta_pin = arm_env("HASH_GTA");
    const char *req = arm_env("REQ_OK"), *work = arm_env("WORKDIR");
    if (!ok_child || !ok_pin || !alt || !gta || !gta_pin || !work)
        arm_fail("env missing (CHILD_OK/HASH_OK/CHILD_ALT/CHILD_GTA/HASH_GTA/WORKDIR)");

    // Positive control: the pinned good child is accepted and runs to its marker.
    sealed_child_spec_t spec;
    sealed_child_t ch;
    snprintf(marker, sizeof marker, "%s/m.t2.control", work);
    if (arm_build_spec(&spec, ok_child, ok_pin, req, marker, 5000) != 0)
        arm_fail("could not build control spec");
    if (arm_spawn(&spec, &ch) != SEALED_SPAWN_OK || !arm_marker_written(marker))
        arm_fail("control failed: pinned child-ok not accepted, or accepted but never ran");

    // Case 1: different CD hash, pinned as the good child's hash (requirement
    // bytes name child-alt itself) -> the module's suspended-phase CD-hash check.
    snprintf(marker, sizeof marker, "%s/m.t2.wronghash", work);
    refusal_case("wrong CD hash (own-hash requirement bytes)", alt, ok_pin, arm_env("REQ_ALT"),
                 SEALED_SPAWN_CDHASH, "SEALED_SPAWN_CDHASH");

    // Case 2: get-task-allow child pinned by its own hash -> only the module's
    // status-word check can refuse it.
    snprintf(marker, sizeof marker, "%s/m.t2.gta", work);
    refusal_case("get-task-allow child", gta, gta_pin, arm_env("REQ_GTA"),
                 SEALED_SPAWN_STATUS, "SEALED_SPAWN_STATUS");

    arm_pass("wrong-hash and get-task-allow children refused (CDHASH/STATUS) before their first main instruction ran; pinned child-ok control accepted");
    return 0;
}
