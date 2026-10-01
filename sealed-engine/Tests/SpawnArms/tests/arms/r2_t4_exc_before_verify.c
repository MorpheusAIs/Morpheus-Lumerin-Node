// Arm r2_t4_exc_before_verify: FIXES-2 items 4(b) and 8 — a child that raises
// an exception before the rendezvous. Two legs: r2_abrt (raise(SIGABRT)) and
// r2_trap (__builtin_trap). Each leg: own-pin + own-hash requirement bytes
// (reqgen; fallback SKIP with the true reason), spawn through the module, and
// assert SEALED_SPAWN_DIED promptly (ms reported), the child reaped,
// exception_seen set, the marker written (it ran), and NO crash report for
// that child name in either DiagnosticReports folder (matching files listed
// before and after each leg, new entries must not appear). Positive control
// in the same program: the pinned child-ok is accepted and runs to its
// marker. Which Mach exception type each child actually delivers is measured
// by tests/probes/exc_probe.c (probes slot; recorded in tests/NOTES.md).
#include <dirent.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include "harness.h"

static char why[1024];

static int hex_of(const char *child, char out[64]) {
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

// DiagnosticReports entries, both folders, appended across scans; a report is
// "new" if it is not in the first `before` entries.
static char reps[2048][192];
static int nreps;

static void scan_reports(void) {
    const char *home = getenv("HOME");
    char p[512];
    const char *dirs[2];
    snprintf(p, sizeof p, "%s/Library/Logs/DiagnosticReports", home ? home : "/var/root");
    dirs[0] = p;
    dirs[1] = "/Library/Logs/DiagnosticReports";
    for (int d = 0; d < 2; d++) {
        DIR *dp = opendir(dirs[d]);
        if (dp == NULL) continue;
        struct dirent *e;
        while ((e = readdir(dp)) != NULL) {
            if (e->d_name[0] == '.') continue;
            if (nreps < 2048) snprintf(reps[nreps++], 192, "%s", e->d_name);
        }
        closedir(dp);
    }
}

static int in_before(const char *nm, int before) {
    for (int i = 0; i < before; i++)
        if (strcmp(reps[i], nm) == 0) return 1;
    return 0;
}

static const char *new_report(int before) {
    for (int i = before; i < nreps; i++) {
        if (in_before(reps[i], before)) continue;
        if (strstr(reps[i], "r2_abrt") != NULL || strstr(reps[i], "r2_trap") != NULL ||
            strstr(reps[i], "child") != NULL)
            return reps[i];
    }
    return NULL;
}

// One leg: 0 = measured and green (pass text in out), 1 = leg skipped, and
// every violation calls arm_fail directly.
static int leg(const char *child, const char *label, const char *work, char out[512]) {
    char pin[64];
    if (hex_of(child, pin) != 0) {
        snprintf(out, 512, "%s: could not get the child's CD hash via codesign", label);
        return 1;
    }
    char reqp[600];
    snprintf(reqp, sizeof reqp, "%s/req-exc-%s.bin", work, label);
    if (gen_req(work, child, pin, reqp) != 0) {
        snprintf(out, 512, "%s: reqgen unusable (no own-hash requirement bytes)", label);
        return 1;
    }
    char mk[512];
    snprintf(mk, sizeof mk, "%s/m.exc-%s", work, label);
    unlink(mk);
    scan_reports();
    int before = nreps;

    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, child, pin, reqp, mk, 5000) != 0)
        arm_fail("leg spec could not be built");
    double t0 = now_ms();
    int rc = arm_spawn(&spec, &st);
    double ms = now_ms() - t0;
    if (rc == SEALED_SPAWN_REQUIREMENT)
        arm_fail("module refused the spec with REQUIREMENT(13) although reqgen bytes exist");
    if (rc != SEALED_SPAWN_DIED) {
        snprintf(why, sizeof why, "%s leg: expected DIED(%d) for an exception before "
                 "the rendezvous, got %s(%d)", label, SEALED_SPAWN_DIED, arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (st.reaped != 1) arm_fail("leg: the child was not reaped");
    if (st.exception_seen != 1) arm_fail("leg: exception_seen not set (message never serviced)");
    if (!arm_marker_written(mk)) arm_fail("leg: the child never ran (marker missing) — nothing was tested");
    if (ms > 15000.0) {
        snprintf(why, sizeof why, "%s leg: returned after %.0f ms (hang?)", label, ms);
        arm_fail(why);
    }
    sleep(2);   // crash-report writes can lag the death
    scan_reports();
    const char *nr = new_report(before);
    if (nr != NULL) {
        snprintf(why, sizeof why, "%s leg: a crash report appeared for the child: %s", label, nr);
        arm_fail(why);
    }
    snprintf(out, 512, "%s leg: DIED in %.0f ms, reaped, exception_seen=1, ran, no crash report",
             label, ms);
    return 0;
}

int main(void) {
    arm_begin("r2_t4_exc_before_verify");
    const char *ok = arm_env("CHILD_OK");
    const char *req = arm_env("REQ_OK");
    const char *work = arm_env("WORKDIR");
    const char *abrt = arm_env("CHILD_R2_ABRT");
    const char *trap = arm_env("CHILD_R2_TRAP");
    if (ok == NULL || req == NULL || work == NULL)
        arm_fail("env missing (CHILD_OK/REQ_OK/WORKDIR)");

    // Positive control: shipping shape, pinned child accepted.
    char ctrl[512];
    snprintf(ctrl, sizeof ctrl, "%s/m.ctrl-exc", work);
    char pin_ok[64];
    if (hex_of(ok, pin_ok) != 0)
        arm_fail("could not get child-ok's CD hash (codesign failed)");
    unlink(ctrl);
    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, ok, pin_ok, req, ctrl, 5000) != 0)
        arm_fail("control spec could not be built (bad pin hex or req bytes)");
    if (arm_spawn(&spec, &st) != SEALED_SPAWN_OK || !arm_marker_written(ctrl))
        arm_fail("control failed: pinned child-ok not accepted/never ran");
    sealed_child_kill(&st);

    scan_reports();
    int before = nreps;

    char a[512] = "", b[512] = "";
    int a_skip = 1, b_skip = 1;
    if (abrt != NULL) a_skip = leg(abrt, "abrt", work, a);
    if (trap != NULL) b_skip = leg(trap, "trap", work, b);
    if (a_skip && b_skip) {
        snprintf(why, sizeof why, "both legs unrunnable: %s | %s", a[0] ? a : "CHILD_R2_ABRT missing",
                 b[0] ? b : "CHILD_R2_TRAP missing");
        arm_skip(why);
    }
    if (a_skip) arm_fail("abrt leg unrunnable and the trap leg cannot carry this arm alone");
    if (b_skip) arm_fail("trap leg unrunnable and the abrt leg cannot carry this arm alone");

    const char *nr = new_report(before);
    if (nr != NULL) {
        snprintf(why, sizeof why, "a crash report appeared for a control child: %s", nr);
        arm_fail(why);
    }
    snprintf(why, sizeof why, "exceptions before the rendezvous (F4): %s; %s; control accepted",
             a, b);
    arm_pass(why);
    return 0;
}
