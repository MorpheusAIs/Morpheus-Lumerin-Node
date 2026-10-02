// Arm r2_t1_control_plain: the negative control for R2-T1's inheritance arm.
// Spec sentence (SPEC-R1-R3.md R2-T1): "Control: a plain posix_spawn without
// the attributes shows each of them inherited." The same reporter child
// (tests/children/r2_reporter.c) is spawned by the arm with plain posix_spawn
// — no attributes, full inherited environment, two extra open descriptors —
// against a listener the arm itself places at <tmpdir>/rv. PASS iff the report
// shows the inherited state: the planted PARENTVAR in the environment, at
// least one descriptor beyond fd 0-2, and the working directory inherited (not
// TMPDIR). This proves the self-report can detect the inheritance that
// r2_t1_inheritance forbids; without it, that arm could pass vacuously against
// a module that merely plain-spawns. This arm pins nothing about the module;
// it pins the detection instrument.
#include <fcntl.h>
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/un.h>
#include <sys/wait.h>
#include <unistd.h>
#include "harness.h"

static char why[256];
static char rep[16384];
extern char **environ;

static int read_report(int fd) {
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

static const char *line(const char *pfx) {
    size_t L = strlen(pfx);
    for (const char *p = rep; (p = strstr(p, pfx)) != NULL; p += L)
        if (p == rep || p[-1] == '\n') return p + L;
    return NULL;
}

int main(void) {
    arm_begin("r2_t1_control_plain");
    const char *rc_child = arm_env("CHILD_R2_REPORTER"), *work = arm_env("WORKDIR");
    if (!rc_child || !work)
        arm_fail("env missing (CHILD_R2_REPORTER/WORKDIR)");
    char pdir[600], prv[640], pmk[600], p0[600];
    snprintf(pdir, sizeof pdir, "%s/plain", work);
    snprintf(prv, sizeof prv, "%s/rv", pdir);
    snprintf(pmk, sizeof pmk, "%s/m.t1.plain", work);
    snprintf(p0, sizeof p0, "%s", rc_child);
    mkdir(pdir, 0700); unlink(prv);
    int ls = socket(AF_UNIX, SOCK_STREAM, 0);
    struct sockaddr_un a;
    memset(&a, 0, sizeof a); a.sun_family = AF_UNIX;
    strncpy(a.sun_path, prv, sizeof a.sun_path - 1);
    if (bind(ls, (struct sockaddr *)&a, sizeof a) || listen(ls, 8))
        arm_fail("control listener setup failed");
    int x1 = open("/dev/null", O_RDWR), x2 = open("/dev/null", O_RDWR); // inherited
    static char hbuf[600], tbuf[600], pbuf[64], *envp[512];
    int ne = 0;
    snprintf(hbuf, sizeof hbuf, "HOME=%s", pdir);
    snprintf(tbuf, sizeof tbuf, "TMPDIR=%s", pdir);
    strcpy(pbuf, "PARENTVAR=leaked-by-control");
    envp[ne++] = hbuf; envp[ne++] = tbuf; envp[ne++] = pbuf;
    for (char **e = environ; *e && ne < 510; e++)
        if (strncmp(*e, "HOME=", 5) && strncmp(*e, "TMPDIR=", 7)) envp[ne++] = *e;
    envp[ne] = NULL;
    char *pargv[3] = {p0, pmk, NULL};
    pid_t pid;
    if (posix_spawn(&pid, rc_child, NULL, NULL, pargv, envp))
        arm_fail("plain posix_spawn failed");
    int cs = accept(ls, NULL, NULL);
    if (cs < 0) arm_fail("child never connected to our listener");
    char sb; read(cs, &sb, 1); // the one fixed byte
    if (read_report(cs)) arm_fail("no report from the plain posix_spawn control");
    close(cs); close(ls); close(x1); close(x2);
    waitpid(pid, NULL, 0);
    int nfds_gt2 = 0;
    for (const char *p = rep; (p = strstr(p, "FD ")) != NULL; p += 3) {
        if (p != rep && p[-1] != '\n') continue;
        int n; if (sscanf(p, "FD %d", &n) == 1 && n > 2) nfds_gt2++;
    }
    const char *pcwd = line("CWD ");
    if (!line("ENV PARENTVAR=") || nfds_gt2 == 0 || (pcwd && !strcmp(pcwd, pdir))) {
        snprintf(why, sizeof why,
                 "control weak: plain posix_spawn showed no inherited env/fds/cwd "
                 "(PARENTVAR=%s, fds>2=%d, cwd-inherited=%d)",
                 line("ENV PARENTVAR=") ? "seen" : "absent", nfds_gt2,
                 pcwd && !strcmp(pcwd, pdir));
        arm_fail(why);
    }
    arm_pass("plain posix_spawn (no attributes) shows the inherited env/fds/cwd the R2-T1 report can catch");
    return 0;
}
