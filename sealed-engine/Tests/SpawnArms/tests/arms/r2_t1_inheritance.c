// Arm r2_t1_inheritance: SPEC-R1-R3.md R2-T1, the sentences the spawn module's
// contract can carry unprivileged (the module ends at the verified channel):
// - "before the rendezvous, exactly fd 0-2 on /dev/null" — from the child's
//   pre-connect self-report (CLOEXEC_DEFAULT inheritance);
// - "environment exactly HOME and TMPDIR" (R2's contingency names tolerated
//   only with their fixed values; any other name or stray MTL_* fails);
// - "all signals reset to default and an empty mask" (mask + dispositions);
// - working directory is TMPDIR and the spawn is into its own process group
//   (POSIX_SPAWN_SETPGROUP), via getcwd/getpgrp/getppid in the report;
// - the fresh 0700 directory holds only an empty metal-cache dir and rv;
// - RLIMIT_CPU/FSIZE are RLIM_INFINITY (R1's constants the spawn must leave).
// Control in the same program (task rules): the pinned child-ok acceptance, so
// a module that refuses everything fails. The negative control (a plain
// posix_spawn without the attributes must show the inherited env/fds/cwd the
// report can catch) is the separate arm r2_t1_control_plain. The reporter
// child (tests/children/r2_reporter.c) is spawned with requirement bytes naming
// the reporter itself (the module fails closed on a missing requirement); A-13
// is exercised by r3_t1 and
// r2_t4_kernel_refusal. See tests/NOTES.md for the arm row.
#include <fcntl.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include "harness.h"

static char why[256];
static char rep[16384];

static int read_report(int fd) { // until the END line or EOF; 0 = report seen
    size_t n = 0;
    rep[0] = '\0';
    while (n < sizeof rep - 1) {
        ssize_t k = read(fd, rep + n, sizeof rep - 1 - n);
        if (k <= 0) break;
        n += (size_t)k;
        rep[n] = '\0';
        if (strstr(rep, "\nEND\n")) return 0;
    }
    return 1;
}

// line() previously returned the pointer into `rep` without terminating at
// the newline, so every strcmp against it compared the whole rest of the
// report — cwd/mask/etc. could never match. Copy the line's text into a small
// rotating ring of buffers (callers may hold up to three at once).
static char lbufs[8][600];
static int lidx;
static const char *line(const char *pfx) {
    size_t L = strlen(pfx);
    for (const char *p = rep; (p = strstr(p, pfx)) != NULL; p += L)
        if (p == rep || p[-1] == '\n') {
            const char *s = p + L, *e = strchr(s, '\n');
            if (e == NULL) return NULL;
            size_t n = (size_t)(e - s);
            if (n >= sizeof lbufs[0]) n = sizeof lbufs[0] - 1;
            memcpy(lbufs[lidx % 8], s, n);
            lbufs[lidx % 8][n] = '\0';
            return lbufs[lidx++ % 8];
        }
    return NULL;
}

static int fds_minimal(void) { // exactly fd 0-2, all /dev/null
    unsigned seen = 0;
    for (const char *p = rep; (p = strstr(p, "FD ")) != NULL; p += 3) {
        if (p != rep && p[-1] != '\n') continue;
        int n; char kind[16];
        if (sscanf(p, "FD %d %15s", &n, kind) != 2) return 0;
        if (n < 0 || n > 2 || strcmp(kind, "DEVNULL")) return 0;
        seen |= 1u << (n + 1);
    }
    return seen == 0xEu;
}

static int env_exact(const char *dir, char *off, size_t on) { // HOME+TMPDIR only
    int have_h = 0, have_t = 0, nenv = 0, ok = 1;
    for (const char *p = rep; (p = strstr(p, "ENV ")) != NULL; p += 4) {
        if (p != rep && p[-1] != '\n') continue;
        const char *s = p + 4, *eq = strchr(s, '='), *e = strchr(s, '\n');
        if (!eq || eq > e) { snprintf(off, on, "unparsable ENV line"); return 0; }
        printf("MEASURED env: %.*s\n", (int)(e - s), s);   // full env, arm output (FIXES-2 5)
        size_t nl = (size_t)(eq - s), vl = (size_t)(e - eq - 1);
        char val[600];
        if (vl >= sizeof val) vl = sizeof val - 1;
        memcpy(val, eq + 1, vl); val[vl] = '\0';
        #define IS(x) (nl == strlen(x) && !strncmp(s, x, nl))
        if (IS("HOME")) { have_h = 1; if (strcmp(val, dir)) { ok = 0; snprintf(off, on, "HOME=%s (len %zu, dir len %zu)", val, vl, strlen(dir)); } }
        else if (IS("TMPDIR")) { have_t = 1; if (strcmp(val, dir)) { ok = 0; snprintf(off, on, "TMPDIR=%s (len %zu, dir len %zu)", val, vl, strlen(dir)); } }
        else { snprintf(off, on, "env %.*s=%s", (int)nl, s, val); ok = 0; }
        #undef IS
        nenv++;
    }
    if (!nenv || !have_h || !have_t) { snprintf(off, on, "missing HOME/TMPDIR"); return 0; }
    return ok;
}

int main(void) {
    arm_begin("r2_t1_inheritance");
    const char *ok_child = arm_env("CHILD_OK"), *ok_pin = arm_env("HASH_OK");
    const char *req = arm_env("REQ_OK"), *work = arm_env("WORKDIR");
    const char *rc_child = arm_env("CHILD_R2_REPORTER"), *rc_pin = arm_env("HASH_R2_REPORTER");
    const char *rc_req = arm_env("REQ_R2_REPORTER");
    if (!ok_child || !ok_pin || !work || !rc_child || !rc_pin || !rc_req)
        arm_fail("env missing (CHILD_OK/HASH_OK/REQ_OK/WORKDIR/CHILD_R2_REPORTER/HASH_R2_REPORTER/REQ_R2_REPORTER)");
    sealed_child_spec_t spec; sealed_child_t ch; char marker[512], mdir[512];

    // FIXES-3 2: make the parent's own signal state non-default BEFORE any
    // spawn (SIGPIPE and SIGUSR2 ignored, SIGUSR1 blocked), so a module that
    // drops POSIX_SPAWN_SETSIGDEF/SETSIGMASK (mutants/mut-b-no-sigreset.c)
    // leaks the inherited state into the child and the report's SIG/MASK
    // lines catch it. Restored before the arm exits.
    struct sigaction ign, save_pipe, save_usr2;
    memset(&ign, 0, sizeof ign);
    ign.sa_handler = SIG_IGN;
    sigaction(SIGPIPE, &ign, &save_pipe);
    sigaction(SIGUSR2, &ign, &save_usr2);
    sigset_t blk, oldmask;
    sigemptyset(&blk);
    sigaddset(&blk, SIGUSR1);
    sigprocmask(SIG_BLOCK, &blk, &oldmask);

    // Control: the pinned good child is accepted (refuses-everything guard).
    snprintf(marker, sizeof marker, "%s/m.t1.control", work);
    if (arm_build_spec(&spec, ok_child, ok_pin, req, marker, 5000) != 0)
        arm_fail("could not build control spec");
    if (arm_spawn(&spec, &ch) != SEALED_SPAWN_OK || !arm_marker_written(marker))
        arm_fail("control failed: pinned child-ok not accepted/never ran");

    // Subject: the module spawns the reporting child (requirement bytes name
    // the reporter itself, so the kernel admits it and the module's checks run).
    snprintf(mdir, sizeof mdir, "%s/m.t1.subject", work);
    if (arm_build_spec(&spec, rc_child, rc_pin, rc_req, mdir, 5000) != 0)
        arm_fail("could not build reporter spec");
    int rc = arm_spawn(&spec, &ch);
    if (rc != SEALED_SPAWN_OK) {
        snprintf(why, sizeof why, "reporter not accepted: result=%s(%d)", arm_code_name(rc), rc);
        arm_fail(why);
    }
    if (!arm_marker_written(mdir)) arm_fail("reporter accepted but never wrote its marker");
    if (read_report(ch.channel_fd) != 0) arm_fail("no self-report over the verified channel");

    if (!fds_minimal()) arm_fail("pre-connect descriptors are not exactly fd 0-2 on /dev/null");
    char off[192] = "";
    const char *dir = ch.dir;
    if (!env_exact(dir, off, sizeof off)) {
        snprintf(why, sizeof why, "environment not exactly HOME/TMPDIR(=%s): %s", dir, off);
        arm_fail(why);
    }
    const char *cwd = line("CWD "), *pg = line("PGRP "), *pp = line("PPID ");
    if (!cwd || strcmp(cwd, dir)) {
        snprintf(why, sizeof why, "working directory %s is not TMPDIR %s", cwd ? cwd : "(none)", dir);
        arm_fail(why);
    }
    if (!pg || atoi(pg) != ch.pid) arm_fail("child is not in its own process group");
    if (!pp || atoi(pp) != (int)getpid()) arm_fail("child's parent is not the spawner");
    if (!line("SIG ") || !line("MASK ")) arm_fail("no signal report");
    for (const char *p = rep; (p = strstr(p, "SIG ")) != NULL; p += 4) {
        if (p != rep && p[-1] != '\n') continue;
        int sn; char k[8];
        if (sscanf(p, "SIG %d %7s", &sn, k) == 2 && strcmp(k, "DFL")) {
            snprintf(why, sizeof why, "signal %d is not at default (=%s)", sn, k);
            arm_fail(why);
        }
    }
    const char *mk = line("MASK ");
    if (!mk || strcmp(mk, "0")) arm_fail("child's signal mask is not empty");
    for (const char *p = rep; (p = strstr(p, "RLIM ")) != NULL; p += 5) {
        if (p != rep && p[-1] != '\n') continue;
        int i; char cur[32], mx[32];
        if (sscanf(p, "RLIM %d %31s %31s", &i, cur, mx) == 3 && (i == 1 || i == 2) &&
            (strcmp(cur, "18446744073709551615") || strcmp(mx, "18446744073709551615"))) {
            snprintf(why, sizeof why, "RLIMIT %s is not RLIM_INFINITY (cur=%s max=%s)",
                     i == 1 ? "CPU" : "FSIZE", cur, mx);
            arm_fail(why);
        }
    }
    unsigned tmode = 0; unsigned long long tuid = 0;
    const char *td = line("TDIR ");
    if (!td || sscanf(td, "%o %llu", &tmode, &tuid) != 2 || tmode != 0700 || tuid != getuid()) {
        snprintf(why, sizeof why, "TMPDIR not a fresh 0700 dir of our uid (mode=%o uid=%llu)", tmode, tuid);
        arm_fail(why);
    }
    int nent = 0, e_ok = 1;
    for (const char *p = rep; (p = strstr(p, "ENTRY ")) != NULL; p += 6) {
        if (p != rep && p[-1] != '\n') continue;
        char nm[128], ty[8]; nent++;
        if (sscanf(p, "ENTRY %127s %7s", nm, ty) != 2) e_ok = 0;
        else if (!strcmp(nm, "metal-cache") && !strcmp(ty, "DIR")) {
            const char *dc = line("DIRCOUNT metal-cache ");
            if (!dc || atoi(dc)) e_ok = 0;
        } else if (strcmp(nm, "rv") || strcmp(ty, "SOCK")) e_ok = 0;
    }
    if (nent != 2 || !e_ok) {
        snprintf(why, sizeof why, "TMPDIR contents not exactly empty metal-cache + rv (%d entries)", nent);
        arm_fail(why);
    }
    close(ch.channel_fd);
    sigaction(SIGPIPE, &save_pipe, NULL);
    sigaction(SIGUSR2, &save_usr2, NULL);
    sigprocmask(SIG_SETMASK, &oldmask, NULL);
    arm_pass("module spawn: fd 0-2 /dev/null only, env exactly HOME/TMPDIR, 0700 fresh dir with empty metal-cache + rv, cwd/pgroup/parent/signals/rlimits as spec'd; control: pinned child-ok accepted");
    return 0;
}
