// Arm r3_t1_get_task_allow: SPEC-R1-R3.md R3-T1, "verification rejects a
// get-task-allow child". The module requires requirement bytes, so the arm
// attaches bytes naming the spawned child's OWN CD hash (the "launch
// requirement off" configuration expressed through the module API: the kernel
// admits child-gta — ad-hoc, hardened runtime, com.apple.security.get-task-allow
// entitlement — because the requirement matches its hash). Its pin is its own
// CD hash, so the engine's hash check passes and the only failing check is the
// status word's forbidden-flag half of R3: CS_GET_TASK_ALLOW must be excluded.
// Expected refusal: SEALED_SPAWN_STATUS ("status word: required/forbidden
// flags"), before resume, child never runs. (The shipped engine would refuse
// this build at the hash check first; this arm isolates the status check by
// pinning the gta build's own hash.) Requirement bytes are generated at run
// time by reqgen (a13.swift dumpfile route, exits before launching anything).
// Positive control in the same program: the pinned child with the same spec
// shape is accepted (OK) and runs to its marker. SKIP if the module refuses a
// requirement-less spec AND no reqgen bytes could be generated.
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

static int gen_req(const char *work, const char *child, const char *hash, const char *out) {
    char cmd[1600];
    snprintf(cmd, sizeof cmd, "'%s/reqgen' '%s' '%s' dummy '%s' >/dev/null 2>&1",
             work, child, hash, out);
    if (system(cmd) != 0) return 1;
    return access(out, R_OK) != 0;
}

int main(void) {
    arm_begin("r3_t1_get_task_allow");
    const char *ok = arm_env("CHILD_OK");
    const char *gta = arm_env("CHILD_GTA");
    const char *req_ok = arm_env("REQ_OK");
    const char *work = arm_env("WORKDIR");
    if (ok == NULL || gta == NULL || work == NULL)
        arm_fail("R3-T1: env missing (CHILD_OK/CHILD_GTA/WORKDIR)");
    char pin_ok[64], pin_gta[64];
    if (hex_of(ok, arm_env("HASH_OK"), pin_ok) != 0)
        arm_fail("R3-T1: could not get child-ok's CD hash (env HASH_OK and codesign both failed)");
    if (hex_of(gta, arm_env("HASH_GTA"), pin_gta) != 0)
        arm_fail("R3-T1: could not get child-gta's CD hash (env HASH_GTA and codesign both failed)");

    char ctrl[512], ref[512], req_gta[600];
    snprintf(ctrl, sizeof ctrl, "%s/m.ctrl-gta", work);
    snprintf(ref, sizeof ref, "%s/m.ref-gta", work);
    snprintf(req_gta, sizeof req_gta, "%s/req-gta.bin", work);
    const char *ref_req = req_gta;
    if (req_ok != NULL && gen_req(work, gta, pin_gta, req_gta) != 0)
        ref_req = NULL;   // reqgen unusable: fall back to a requirement-less spec

    // Positive control: shipping shape (requirement naming the child's own
    // hash), pinned child accepted.
    unlink(ctrl);
    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, ok, pin_ok, req_ok, ctrl, 5000) != 0)
        arm_fail("R3-T1: control spec could not be built (bad pin hex or req bytes)");
    int rc = arm_spawn(&spec, &st);
    if (rc == SEALED_SPAWN_REQUIREMENT)
        arm_skip("module refuses a spec it cannot attach a launch requirement to; "
                 "this arm needs requirement bytes naming the spawned child's own hash");
    if (rc != SEALED_SPAWN_OK) {
        snprintf(why, sizeof why, "R3-T1: pinned control child refused: result=%s(%d)",
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (!arm_marker_written(ctrl)) {
        arm_fail("R3-T1: control accepted (OK) but the child never wrote its marker");
    }
    sealed_child_kill(&st);

    // get-task-allow child, pinned by its own hash and admitted by the kernel:
    // only the status word's forbidden-flag check refuses it.
    unlink(ref);
    if (arm_build_spec(&spec, gta, pin_gta, ref_req, ref, 5000) != 0)
        arm_fail("R3-T1: refusal spec could not be built");
    rc = arm_spawn(&spec, &st);
    if (rc == SEALED_SPAWN_REQUIREMENT)
        arm_skip("module refuses a spec with no launch requirement and no reqgen "
                 "bytes could be generated; this arm needs the kernel out of the way");
    if (rc != SEALED_SPAWN_STATUS) {
        snprintf(why, sizeof why, "R3-T1: expected STATUS(%d) for the forbidden "
                 "get-task-allow flag, got %s(%d)",
                 SEALED_SPAWN_STATUS, arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (arm_marker_written(ref)) {
        arm_fail("R3-T1: refused but the child ran (marker written)");
    }
    if (st.reaped != 1) {
        arm_fail("R3-T1: refused but the child was not reaped");
    }
    arm_pass("R3-T1: get-task-allow child refused with STATUS before it ran; "
             "pinned control accepted");
    return 0;
}
