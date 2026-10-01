// Arm r2_t4_kernel_refusal: SPEC-R1-R3.md R2-T4(b) and its control (h) —
// "root replaces the helper file before the spawn: the kernel refuses to exec
// it under the launch requirement and the spawn fails (A-13)"; control (h):
// "with the requirement naming the pinned hash, a validly signed helper with a
// different CD hash is refused by the kernel; an arm with the requirement
// absent records that the rogue runs". Expressed unprivileged against the
// module contract: the "replaced helper" is a different validly signed image
// (child-alt) whose CD hash differs from the requirement's named hash —
// exactly the post-swap state the kernel refuses (evidence E1: the suspended
// child reads CS_KILLED before resume and never runs). Pre-resume, the module
// can only see a suspended process whose CD hash is not the pin, so the
// specific expected code is SEALED_SPAWN_CDHASH; the refusal arm also asserts
// the child never ran (its first main instruction never executes). Positive
// controls in the same program: child-ok accepted under the same requirement
// bytes (so a module that refuses everything fails), and — the (h) negative
// half — the same wrong image with requirement bytes naming child-alt itself,
// which the kernel would run; the module must still refuse it
// (SEALED_SPAWN_CDHASH), recording that the rogue image is what the kernel
// would have run. (The module fails closed on a missing requirement
// (REQUIREMENT), so "requirement absent" is not expressible through this
// contract; the own-hash requirement is the expressible (h) half.)
#include <stdio.h>
#include "harness.h"

static char why[256], marker[512];

static void expect(const char *label, const char *child, const char *pin,
                   const char *req, int expect_code, const char *expect_name,
                   int never_runs) {
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
    if (never_runs && arm_marker_written(marker)) {
        snprintf(why, sizeof why, "%s: refused with %s but the child RAN (marker written)",
                 label, expect_name);
        arm_fail(why);
    }
}

int main(void) {
    arm_begin("r2_t4_kernel_refusal");
    const char *ok_child = arm_env("CHILD_OK"), *ok_pin = arm_env("HASH_OK");
    const char *alt = arm_env("CHILD_ALT");
    const char *req = arm_env("REQ_OK"), *work = arm_env("WORKDIR");
    if (!ok_child || !ok_pin || !alt || !req || !work)
        arm_fail("env missing (CHILD_OK/HASH_OK/CHILD_ALT/REQ_OK/WORKDIR)");

    // Positive control: pinned child-ok under the same requirement bytes.
    sealed_child_spec_t spec;
    sealed_child_t ch;
    snprintf(marker, sizeof marker, "%s/m.t4k.control", work);
    if (arm_build_spec(&spec, ok_child, ok_pin, req, marker, 5000) != 0)
        arm_fail("could not build control spec");
    if (arm_spawn(&spec, &ch) != SEALED_SPAWN_OK || !arm_marker_written(marker))
        arm_fail("control failed: pinned child-ok not accepted under the requirement bytes");

    // (b): requirement names child-ok's hash, the spawned image is child-alt —
    // the kernel refuses to exec it (A-13); the module sees a suspended process
    // whose CD hash is not the pin and must refuse before resume, never runs.
    snprintf(marker, sizeof marker, "%s/m.t4k.wrongimage", work);
    expect("wrong image under requirement (b)", alt, ok_pin, req,
           SEALED_SPAWN_CDHASH, "SEALED_SPAWN_CDHASH", 1);

    // (h) negative half: the same wrong image, requirement bytes naming
    // child-alt itself — the kernel would run it; only the module's own check
    // can refuse it.
    snprintf(marker, sizeof marker, "%s/m.t4k.noreq", work);
    expect("wrong image with own-hash requirement (h)", alt, ok_pin, arm_env("REQ_ALT"),
           SEALED_SPAWN_CDHASH, "SEALED_SPAWN_CDHASH", 1);

    arm_pass("requirement-naming wrong image refused pre-resume (CDHASH, child never ran), both with the swapped-hash requirement and with the image's own-hash requirement; pinned child-ok control accepted");
    return 0;
}
