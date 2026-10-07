// Arm r1_t5_late_crash: FIXES-3 item 1 — a child that crashes about 3 s after
// starting, before ever connecting (pinned by its own CD hash with
// own-hash requirement bytes), with the arm's rendezvous deadline at 8000 ms.
// Assert SEALED_SPAWN_DIED within about 1 s of the crash (ms from spawn
// reported; the crash lands ~3 s in, so anything past ~4.5 s is a missed
// answer), the child reaped, exception_seen set, marker written (it ran).
// Positive control in the same program: the pinned child-ok, shipping shape,
// accepted. Against the fixed module this arm PASSes (the live spawn-port
// reader services the late crash); against mutants/mut-a-notify-before-insert
// (notification registered before the send-right insert — the measured
// immediately-delivered order) it FAILs: that reader has already drained and
// exited by the time the crash arrives, the crash is unanswered, and the call
// waits out the deadline.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include "harness.h"

static char why[512];

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

static double now_ms(void) {
    struct timespec t;
    if (clock_gettime(CLOCK_MONOTONIC, &t) != 0) return 0.0;
    return t.tv_sec * 1000.0 + t.tv_nsec / 1e6;
}

static int gen_req(const char *work, const char *child, const char *hash, const char *out) {
    char cmd[1600];
    snprintf(cmd, sizeof cmd, "'%s/reqgen' '%s' '%s' dummy '%s' >/dev/null 2>&1",
             work, child, hash, out);
    if (system(cmd) != 0) return 1;
    return access(out, R_OK) != 0;
}

int main(void) {
    arm_begin("r1_t5_late_crash");
    const char *ok = arm_env("CHILD_OK");
    const char *late = arm_env("CHILD_R2_LATECRASH");
    const char *req = arm_env("REQ_OK");
    const char *work = arm_env("WORKDIR");
    if (ok == NULL || late == NULL || req == NULL || work == NULL)
        arm_fail("env missing (CHILD_OK/CHILD_R2_LATECRASH/REQ_OK/WORKDIR)");
    char pin_ok[64], pin_late[64];
    if (hex_of(ok, arm_env("HASH_OK"), pin_ok) != 0)
        arm_fail("could not get child-ok's CD hash (env HASH_OK and codesign both failed)");
    if (hex_of(late, NULL, pin_late) != 0)
        arm_fail("could not get the late-crash child's CD hash via codesign");

    char ctrl[512], ref[512], req_late[600];
    snprintf(ctrl, sizeof ctrl, "%s/m.ctrl-late", work);
    snprintf(ref, sizeof ref, "%s/m.ref-late", work);
    snprintf(req_late, sizeof req_late, "%s/req-late.bin", work);
    const char *ref_req = req_late;
    if (gen_req(work, late, pin_late, req_late) != 0)
        ref_req = NULL;   // reqgen unusable: fall back to a requirement-less spec

    // Positive control: shipping shape, pinned child accepted.
    unlink(ctrl);
    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, ok, pin_ok, req, ctrl, 5000) != 0)
        arm_fail("control spec could not be built (bad pin hex or req bytes)");
    int rc = arm_spawn(&spec, &st);
    if (rc != SEALED_SPAWN_OK) {
        snprintf(why, sizeof why, "pinned control child refused: result=%s(%d)",
                 arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (!arm_marker_written(ctrl))
        arm_fail("control accepted (OK) but the child never wrote its marker");
    sealed_child_kill(&st);

    // Late-crash child: runs, marker, ~3 s silent (no connection), faults.
    unlink(ref);
    if (arm_build_spec(&spec, late, pin_late, ref_req, ref, 8000) != 0)
        arm_fail("refusal spec could not be built");
    double t0 = now_ms();
    rc = arm_spawn(&spec, &st);
    double ms = now_ms() - t0;
    if (rc == SEALED_SPAWN_REQUIREMENT)
        arm_skip("module refuses a spec with no launch requirement and no reqgen "
                 "bytes could be generated for the late-crash child");
    if (rc != SEALED_SPAWN_DIED) {
        snprintf(why, sizeof why, "expected DIED(%d) for a child that crashed ~3 s in, "
                 "got %s(%d) after %.0f ms (deadline 8000)",
                 SEALED_SPAWN_DIED, arm_code_name(rc), rc, ms);
        arm_fail(why);
    }
    if (ms > 4500.0) {
        snprintf(why, sizeof why, "DIED but only after %.0f ms — more than ~1 s past the "
                 "crash (~3 s in): the exception was not answered on arrival", ms);
        arm_fail(why);
    }
    if (st.reaped != 1) arm_fail("refused but the child was not reaped");
    if (st.exception_seen != 1) arm_fail("exception_seen not set (crash message never serviced)");
    if (!arm_marker_written(ref))
        arm_fail("the late-crash child never ran (marker missing) — nothing was tested");
    snprintf(why, sizeof why, "crash ~3 s in, no connection ever: DIED within the crash's "
             "~1 s window (%.0f ms from spawn, deadline 8000), reaped, exception_seen=1; "
             "control accepted", ms);
    arm_pass(why);
    return 0;
}
