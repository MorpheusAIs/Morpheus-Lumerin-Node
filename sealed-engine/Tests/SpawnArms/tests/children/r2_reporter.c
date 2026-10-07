// Reporting child for the R2-T1 inheritance arm (tests/arms/r2_t1_inheritance.c).
// It writes the marker named by argv[1] (its first act in main), then collects
// — strictly before connecting — its descriptor table (with /dev/null
// identification), environment, working directory, process group and parent,
// signal mask and dispositions, rlimits, and the stat and contents of $TMPDIR.
// Then it connects to $TMPDIR/rv, sends the one fixed byte 'S' (the R2
// rendezvous contract), sends the whole pre-connect report over the same
// connection, and waits for EOF so no child is left behind between arms. It
// enforces nothing itself: the arm owns every assertion, so the same binary
// serves the module spawn (must show the minimal inherited state) and the
// plain-posix_spawn control (must show the inherited state it would have
// leaked). Built by the generic r2_* child rule; CHILD_R2_REPORTER and
// HASH_R2_REPORTER are exported by the Makefile's test target.
#include <dirent.h>
#include <fcntl.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/un.h>
#include <unistd.h>

static void report_fds(FILE *out) {
    DIR *d = opendir("/dev/fd");
    if (d == NULL) { fprintf(out, "FDSCAN -1\n"); return; }
    int dfd = dirfd(d);
    struct stat nullst;
    int have_null = stat("/dev/null", &nullst) == 0;
    struct dirent *e;
    while ((e = readdir(d)) != NULL) {
        if (e->d_name[0] < '0' || e->d_name[0] > '9') continue;
        int fd = atoi(e->d_name);
        if (fd == dfd) continue;
        struct stat st;
        if (fstat(fd, &st) != 0) { fprintf(out, "FD %d BAD\n", fd); continue; }
        const char *kind = S_ISCHR(st.st_mode) ? "CHR" : "OTHER";
        if (have_null && S_ISCHR(st.st_mode) && st.st_rdev == nullst.st_rdev)
            kind = "DEVNULL";
        fprintf(out, "FD %d %s\n", fd, kind);
    }
    closedir(d);
}

static void report_env(FILE *out) {
    extern char **environ;
    for (char **e = environ; *e != NULL; e++)
        fprintf(out, "ENV %s\n", *e);
}

static void report_signals(FILE *out) {
    static const int sigs[] = {SIGSEGV, SIGBUS, SIGILL, SIGFPE, SIGABRT, SIGTRAP};
    for (size_t i = 0; i < sizeof sigs / sizeof sigs[0]; i++) {
        struct sigaction sa;
        memset(&sa, 0, sizeof sa);
        if (sigaction(sigs[i], NULL, &sa) != 0) { fprintf(out, "SIG %zu BAD\n", i); continue; }
        const char *k = sa.sa_handler == SIG_DFL ? "DFL" :
                        sa.sa_handler == SIG_IGN ? "IGN" : "OTHER";
        fprintf(out, "SIG %d %s\n", sigs[i], k);
    }
    sigset_t m;
    sigprocmask(SIG_SETMASK, NULL, &m);
    unsigned mask = 0;
    for (int s = 1; s < 32; s++)
        if (sigismember(&m, s) == 1) mask |= 1u << s;
    fprintf(out, "MASK %x\n", mask);
}

static void report_rlimits(FILE *out) {
    static const int which[] = {RLIMIT_CORE, RLIMIT_CPU, RLIMIT_FSIZE,
                                RLIMIT_NOFILE, RLIMIT_STACK, RLIMIT_DATA, RLIMIT_AS};
    for (size_t i = 0; i < sizeof which / sizeof which[0]; i++) {
        struct rlimit rl;
        if (getrlimit(which[i], &rl) != 0) { fprintf(out, "RLIM %zu BAD\n", i); continue; }
        fprintf(out, "RLIM %zu %llu %llu\n", i,
                rl.rlim_cur == RLIM_INFINITY ? 0xffffffffffffffffULL : (unsigned long long)rl.rlim_cur,
                rl.rlim_max == RLIM_INFINITY ? 0xffffffffffffffffULL : (unsigned long long)rl.rlim_max);
    }
}

static void report_tmpdir(FILE *out) {
    const char *tmp = getenv("TMPDIR");
    if (tmp == NULL) { fprintf(out, "TDIR -\n"); return; }
    struct stat st;
    if (stat(tmp, &st) != 0) { fprintf(out, "TDIR -\n"); return; }
    fprintf(out, "TDIR %o %llu\n", st.st_mode & 07777, (unsigned long long)st.st_uid);
    DIR *d = opendir(tmp);
    if (d == NULL) return;
    struct dirent *e;
    while ((e = readdir(d)) != NULL) {
        if (strcmp(e->d_name, ".") == 0 || strcmp(e->d_name, "..") == 0) continue;
        char p[1200];
        snprintf(p, sizeof p, "%s/%s", tmp, e->d_name);
        struct stat s2;
        const char *t = "OTHER";
        if (stat(p, &s2) == 0)
            t = S_ISDIR(s2.st_mode) ? "DIR" : S_ISSOCK(s2.st_mode) ? "SOCK" : "OTHER";
        fprintf(out, "ENTRY %s %s\n", e->d_name, t);
        if (S_ISDIR(s2.st_mode)) {
            DIR *d2 = opendir(p);
            int n = -1;
            if (d2 != NULL) {
                n = 0;
                struct dirent *e2;
                while ((e2 = readdir(d2)) != NULL)
                    if (strcmp(e2->d_name, ".") && strcmp(e2->d_name, "..")) n++;
                closedir(d2);
            }
            fprintf(out, "DIRCOUNT %s %d\n", e->d_name, n);
        }
    }
    closedir(d);
}

int main(int argc, char **argv) {
    if (argc > 1 && argv[1][0] != '\0') {
        int fd = open(argv[1], O_CREAT | O_WRONLY, 0644);
        if (fd >= 0) close(fd);
    }
    char *buf = NULL;
    size_t sz = 0;
    FILE *out = open_memstream(&buf, &sz);
    if (out == NULL) _exit(113);
    char cwd[1024];
    if (getcwd(cwd, sizeof cwd) == NULL) snprintf(cwd, sizeof cwd, "?");
    fprintf(out, "CWD %s\n", cwd);
    fprintf(out, "PGRP %d\n", getpgrp());
    fprintf(out, "PPID %d\n", getppid());
    report_fds(out);
    report_env(out);
    report_signals(out);
    report_rlimits(out);
    report_tmpdir(out);
    fprintf(out, "END\n");
    fclose(out);
    const char *tmp = getenv("TMPDIR");
    char path[1024];
    if (tmp == NULL || snprintf(path, sizeof path, "%s/rv", tmp) >= (int)sizeof path)
        _exit(113);
    int s = socket(AF_UNIX, SOCK_STREAM, 0);
    if (s < 0) _exit(113);
    struct sockaddr_un a;
    memset(&a, 0, sizeof a);
    a.sun_family = AF_UNIX;
    strncpy(a.sun_path, path, sizeof a.sun_path - 1);
    if (connect(s, (struct sockaddr *)&a, sizeof a) != 0) _exit(113);
    char c = 'S';
    if (write(s, &c, 1) != 1) _exit(113);
    size_t off = 0;
    while (buf != NULL && off < sz) {
        ssize_t n = write(s, buf + off, sz - off);
        if (n <= 0) _exit(113);
        off += (size_t)n;
    }
    char sink[64];
    for (;;) {
        ssize_t n = read(s, sink, sizeof sink);
        if (n <= 0) break;
    }
    return 0;
}
