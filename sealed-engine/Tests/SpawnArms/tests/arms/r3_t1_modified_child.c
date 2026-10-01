// Arm r3_t1_modified_child: SPEC-R1-R3.md R3-T1, "verification rejects a
// modified child". Shipping shape: the launch requirement names the pinned
// child-ok CD hash; the helper file on the path has been replaced by a
// different valid build (child-alt). The kernel refuses it at exec (A-13), the
// suspended child never runs (E1: it reads CS_KILLED before resume, marker
// never written), and the engine refuses with SEALED_SPAWN_CDHASH — csops on
// the suspended child returns the executed image's CD hash, which is not the
// pin. Positive control in the same program: the pinned child with the same
// pin and the same requirement bytes is accepted (OK) and runs to its marker,
// so a module that refuses everything fails this arm.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include "harness.h"

static char why[256];

// CD hash hex for a child: env hash if valid, else what codesign reports for
// the file (the same value the Makefile's cdhash macro reads).
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

static void expect_accepted(const char *child, const char *pin, const char *req,
                            const char *marker) {
    unlink(marker);
    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, child, pin, req, marker, 5000) != 0)
        arm_fail("R3-T1: control spec could not be built (bad pin hex or req bytes)");
    int rc = arm_spawn(&spec, &st);
    if (rc != SEALED_SPAWN_OK) {
        snprintf(why, sizeof why, "R3-T1: pinned control child refused: result=%s(%d)",
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (!arm_marker_written(marker)) {
        arm_fail("R3-T1: control accepted (OK) but the child never wrote its marker");
    }
    sealed_child_kill(&st);
}

static void expect_refused(const char *child, const char *pin, const char *req,
                           const char *marker, int want) {
    unlink(marker);
    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, child, pin, req, marker, 5000) != 0)
        arm_fail("R3-T1: refusal spec could not be built");
    int rc = arm_spawn(&spec, &st);
    if (rc != want) {
        snprintf(why, sizeof why, "R3-T1: expected %s(%d), got %s(%d)",
                 arm_code_name(want), want, arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (arm_marker_written(marker)) {
        arm_fail("R3-T1: refused but the child ran (marker written)");
    }
    if (st.reaped != 1) {
        arm_fail("R3-T1: refused but the child was not reaped");
    }
}

int main(void) {
    arm_begin("r3_t1_modified_child");
    const char *ok = arm_env("CHILD_OK");
    const char *alt = arm_env("CHILD_ALT");
    const char *req = arm_env("REQ_OK");
    const char *work = arm_env("WORKDIR");
    if (ok == NULL || alt == NULL || req == NULL || work == NULL)
        arm_fail("R3-T1: env missing (CHILD_OK/CHILD_ALT/REQ_OK/WORKDIR)");
    char pin[64];
    if (hex_of(ok, arm_env("HASH_OK"), pin) != 0)
        arm_fail("R3-T1: could not get child-ok's CD hash (env HASH_OK and codesign both failed)");
    char ctrl[512], ref[512];
    snprintf(ctrl, sizeof ctrl, "%s/m.ctrl-mod", work);
    snprintf(ref, sizeof ref, "%s/m.ref-mod", work);
    expect_accepted(ok, pin, req, ctrl);
    expect_refused(alt, pin, req, ref, SEALED_SPAWN_CDHASH);
    arm_pass("R3-T1: modified child refused with CDHASH before it ran; pinned control accepted");
    return 0;
}
