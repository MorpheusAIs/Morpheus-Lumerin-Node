// Arm r3_t2_replaced_helper: SPEC-R1-R3.md R3-T2 — "On a test-build engine
// with the launch requirement off (so the engine's own check is what refuses),
// between requests, the helper on disk is replaced and the child restarted:
// the next request is refused and the session ends. Controls: an unmodified
// restart continues; the replacement is a validly signed different build, so
// only the engine's check can refuse it."
// This module refuses a spec with no requirement bytes (SEALED_SPAWN_REQUIREMENT,
// measured on the test Mac), so the launch-requirement-off test-build engine R3-T2
// is named by is not expressible through include/sealed_spawn.h: attaching the
// shipping requirement bytes (naming the pin) makes the kernel refuse the
// replaced helper first (E1), and the pass would say nothing about the engine.
// The arm runs its positive control (an unmodified restart: the pinned child
// accepted, OK + marker) and then SKIPs the R3-T2-specific part with this
// reason. The closest expressible evidence that the engine's own check
// refuses a different-hash child when the kernel admits it is
// r3_t1_different_hash.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include "harness.h"

static char why[256];

static int hex_of(const char *child, const char *envh, char out[64]) {
    if (envh != NULL && strlen(envh) == 2 * SEALED_CDHASH_LEN) {
        snprintf(out, 64, "%s", envh);
        return 0;
    }
    char cmd[1200];
    snprintf(cmd, sizeof cmd, "codesign -dvvv '%s' 2>&1", child);
    FILE *f = popen(cmd, "r");
    if (f == NULL) return 1;
    char line[512], hex[64];
    int ok = 1;
    while (fgets(line, sizeof line, f) != NULL) {
        char *p = strstr(line, "CDHash=");
        if (p != NULL) {
            p += 7;
            char *e = strchr(p, '\n');
            if (e != NULL) *e = '\0';
            if (strlen(p) == 2 * SEALED_CDHASH_LEN) {
                snprintf(hex, sizeof hex, "%s", p);
                ok = 0;
            }
        }
    }
    pclose(f);
    if (ok) return 1;
    snprintf(out, 64, "%s", hex);
    return 0;
}

int main(void) {
    arm_begin("r3_t2_replaced_helper");
    const char *ok = arm_env("CHILD_OK");
    const char *req_ok = arm_env("REQ_OK");
    const char *work = arm_env("WORKDIR");
    if (ok == NULL || req_ok == NULL || work == NULL)
        arm_fail("R3-T2: env missing (CHILD_OK/REQ_OK/WORKDIR)");
    char pin[64];
    if (hex_of(ok, arm_env("HASH_OK"), pin) != 0)
        arm_fail("R3-T2: could not get child-ok's CD hash (env HASH_OK and codesign both failed)");

    // Control (R3-T2's "an unmodified restart continues"): the pinned child,
    // shipping shape, accepted and running to its marker.
    char ctrl[512];
    snprintf(ctrl, sizeof ctrl, "%s/m.ctrl-r32", work);
    unlink(ctrl);
    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, ok, pin, req_ok, ctrl, 5000) != 0)
        arm_fail("R3-T2: control spec could not be built (bad pin hex or req bytes)");
    int rc = arm_spawn(&spec, &st);
    if (rc == SEALED_SPAWN_REQUIREMENT)
        arm_skip("module refuses a spec it cannot attach a launch requirement to");
    if (rc != SEALED_SPAWN_OK) {
        snprintf(why, sizeof why, "R3-T2: unmodified-restart control refused: result=%s(%d)",
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (!arm_marker_written(ctrl)) {
        arm_fail("R3-T2: control accepted (OK) but the child never wrote its marker");
    }
    sealed_child_kill(&st);

    arm_skip("R3-T2 needs a test-build engine with the launch requirement off so the "
             "engine's own check refuses the replaced helper; the module refuses a "
             "spec with no requirement bytes (REQUIREMENT) and with the shipping "
             "requirement bytes the kernel refuses the replacement first (E1), so "
             "only the engine's check cannot be isolated through "
             "include/sealed_spawn.h — see r3_t1_different_hash for the closest "
             "expressible evidence");
    return 0;
}
