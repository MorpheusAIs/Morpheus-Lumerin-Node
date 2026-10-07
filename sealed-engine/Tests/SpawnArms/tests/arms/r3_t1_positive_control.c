// Arm r3_t1_positive_control: the positive control for every R3-T1 refusal
// arm. Spec sentence (SPEC-R1-R3.md, R3-T1): "Controls: the pinned child is
// accepted". The pinned child is child-ok (ad-hoc signed, hardened runtime,
// valid status flags), pinned by its measured CD hash, with the A-13
// launch-requirement bytes naming that hash attached the way run-a13.sh
// produces them. PASS iff sealed_spawn_child returns SEALED_SPAWN_OK and the
// child ran to its marker; any other code or a missing marker is FAIL, so a
// harness that refuses everything cannot pass.
#include <stdio.h>
#include "harness.h"

static char why[256];

int main(void) {
    arm_begin("r3_t1_positive_control");
    const char *child = arm_env("CHILD_OK");
    const char *pin = arm_env("HASH_OK");
    const char *req = arm_env("REQ_OK");
    const char *work = arm_env("WORKDIR");
    if (child == NULL || pin == NULL || work == NULL)
        arm_fail("env missing (CHILD_OK/HASH_OK/REQ_OK/WORKDIR)");
    char marker[512];
    snprintf(marker, sizeof marker, "%s/m.positive", work);
    sealed_child_spec_t spec;
    sealed_child_t child_st;
    if (arm_build_spec(&spec, child, pin, req, marker, 5000) != 0)
        arm_fail("could not build spec (bad pin hex or unreadable req bytes)");
    int rc = arm_spawn(&spec, &child_st);
    if (rc != SEALED_SPAWN_OK) {
        snprintf(why, sizeof why, "pinned child not accepted: result=%s(%d)",
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (!arm_marker_written(marker)) {
        arm_fail("accepted (OK) but the child never wrote its marker");
    }
    arm_pass("pinned child accepted and ran to its marker");
    return 0;
}
