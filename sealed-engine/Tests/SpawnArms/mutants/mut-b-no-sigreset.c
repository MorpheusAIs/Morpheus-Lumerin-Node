// R2 spawn and rendezvous, R3 pinned-code checks, and the engine's spawn-time
// exception-port handler (R1). See include/sealed_spawn.h.
// Sources: context/SPEC-R1-R3.md (R1 spawn-time port, R2, R3, A-13, A-14),
// context/EVIDENCE-2026-09-30.md (E1 A-13 SPI, E2 exception ports, E3
// LOCAL_PEERTOKEN, E4 csops_audittoken, E5 fd listing, E11 early-death
// handler rule). libSystem only; the private declarations are the ones the
// probes already use.
#include <sealed_guard.h>
#include <sealed_spawn.h>

#include <bsm/libbsm.h>
#include <errno.h>
#include <fcntl.h>
#include <libproc.h>
#include <mach/mach.h>
#include <poll.h>
#include <pthread.h>
#include <stdatomic.h>
#include <signal.h>
#include <spawn.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/proc_info.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/syslimits.h>
#include <sys/types.h>
#include <sys/un.h>
#include <sys/wait.h>
#include <time.h>
#include <unistd.h>

_Static_assert((SEALED_GUARD_MASK & SEALED_NULL_MASK) == 0, "guard and null masks overlap");

// Private entry points the SDK does not declare. Their signatures are the
// ones the probes confirmed: amfi_launch_constraint_set_spawnattr by the
// pos/neg matrix (E1, probes/a13spi.c), csops and csops_audittoken by
// probes/peertoken.c and results-round3 (E4). pid_resume is refused with
// EPERM for an unprivileged caller (E1); SIGCONT is the measured resume.
extern int amfi_launch_constraint_set_spawnattr(posix_spawnattr_t *, const void *, size_t);
extern int csops(pid_t, unsigned int, void *, size_t);
extern int csops_audittoken(pid_t, unsigned int, void *, size_t, audit_token_t *);
extern int pid_resume(int);

#define CS_OPS_STATUS 0
#define CS_OPS_CDHASH 5

// kern/cs_blobs.h, decoded from the measured status words: a refused child
// reads 0x23000201 where a live suspended child reads 0x22000201 (probes/
// results-round3-2026-09-30.txt). The header omits the name, so it is defined
// here; a KILLED word must never pass, even in record-only mode (measured
// word 0 then still refuses a kernel-killed child).
#define SEALED_CS_KILLED 0x01000000u

#define SPAWN_BEHAVIOR ((exception_behavior_t)(EXCEPTION_DEFAULT | MACH_EXCEPTION_CODES))

// The spawn-time exception message for EXCEPTION_DEFAULT | MACH_EXCEPTION_CODES
// (shape as received by probes/handler.c, S1 arms: id 2405, size 84, no thread
// state, thread and task rights carried as port descriptors).
typedef struct {
    mach_msg_header_t h;
    mach_msg_body_t body;
    mach_msg_port_descriptor_t thread;
    mach_msg_port_descriptor_t task;
    NDR_record_t ndr;
    exception_type_t exception;
    mach_msg_type_number_t code_count;
    int64_t code[2];
} exc_msg_t;

typedef struct {
    mach_msg_header_t h;
    NDR_record_t ndr;
    kern_return_t ret;
} exc_reply_msg_t;

// How long the lifecycle waits for the spawn-port reader to answer a pending
// exception message before falling back to destroying the port, and how long
// the reap loop may run after that fallback. Every stage is bounded so no
// refusal path can block the engine.
#define KILL_GRACE_NS 2000000000ull
#define KILL_REAP_NS 10000000000ull
#define READER_DRAIN_NS 2000000000ull
#define READER_DRAIN_ROUND_MS 200
#define READER_JOIN_NS 2000000000ull

static uint64_t now_ns(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ull + (uint64_t)ts.tv_nsec;
}

// audit_token_t.val[5] is the pid (bsm/libbsm.h); read as a field so the
// module needs no libbsm link.
static pid_t token_pid(const audit_token_t *t) { return (pid_t)t->val[5]; }

// Peeks whether the child is dead without reaping it: the engine does not
// wait on the child during the spawn window, so the pid it verifies is the
// process it spawned (R3). WNOWAIT leaves the child waitable for the later
// reap.
static int child_dead_unreaped(pid_t pid) {
    siginfo_t si;
    memset(&si, 0, sizeof si);
    si.si_pid = 0;
    if (waitid(P_PID, pid, &si, WEXITED | WNOWAIT | WNOHANG) != 0) return 0;
    return si.si_pid == (pid_t)pid;
}

// The lifecycle kill (R1): SIGKILL goes to the pid only while the state says
// the child has not been reaped, so the pid cannot have been reused. Used by
// the spawn-port reader and by every kill path in this module.
static void kill_if_not_reaped(sealed_child_t *child) {
    if (child == NULL || child->lock == NULL) return;
    pthread_mutex_t *mu = child->lock;
    pthread_mutex_lock(mu);
    if (!child->reaped && child->pid > 0) kill(child->pid, SIGKILL);
    pthread_mutex_unlock(mu);
}

// One exception message serviced per the measured E11 rule (probes/
// results-handler-policy S1/S2 arms): release the thread and task rights the
// message carries without using them; on EXC_CRASH reply KERN_SUCCESS; on any
// other fatal-type exception SIGKILL the child through the lifecycle, then
// reply KERN_SUCCESS. The reply is required: holding the reply right held the
// child unreaped (E11 S1/S2), and replying KERN_FAILURE passes the event on
// to crash reporting. Replying KERN_SUCCESS without the kill retries the
// fault, so the kill is required. It never reads, copies or logs anything
// from the message beyond the exception type that selects the branch.
static void service_exc_msg(sealed_child_t *child, const exc_msg_t *req) {
    (void)mach_port_deallocate(mach_task_self(), req->thread.name);
    (void)mach_port_deallocate(mach_task_self(), req->task.name);
    if (child->lock != NULL) {
        pthread_mutex_lock(child->lock);
        child->exception_seen = 1;
        pthread_mutex_unlock(child->lock);
    } else {
        child->exception_seen = 1;
    }
    if (req->exception != EXC_CRASH) kill_if_not_reaped(child);
    exc_reply_msg_t rep;
    memset(&rep, 0, sizeof rep);
    rep.h.msgh_bits = MACH_MSGH_BITS(MACH_MSG_TYPE_MOVE_SEND_ONCE, 0);
    rep.h.msgh_size = sizeof rep;
    rep.h.msgh_remote_port = req->h.msgh_remote_port;
    rep.h.msgh_id = req->h.msgh_id + 100;
    rep.ndr = NDR_record;
    rep.ret = KERN_SUCCESS;
    kern_return_t sr = mach_msg(&rep.h, MACH_SEND_MSG, sizeof rep, 0, MACH_PORT_NULL,
                                MACH_MSG_TIMEOUT_NONE, MACH_PORT_NULL);
    // r1-handler F7: a failed reply send leaves the reply's send-once right
    // unconsumed — destroy it or the right leaks per exception message.
    if (sr != KERN_SUCCESS) mach_msg_destroy(&rep.h);
}

// The spawn-port reader: one thread per child generation, started before the
// port is set, distinct from the engine's guard thread (R1). It must be
// running before anything waits to reap: a child killed at spawn while its
// exception message is unanswered is held in exit (measured E1, and the
// 2026-09-30 smoke run), and only the reader's reply or the port's
// destruction releases it. A no-senders notification (simple message) means
// the last send right dropped — the child's guard replaced the port, or the
// child was torn down. The reader does not exit on it: the refused child's
// exception message can be delivered after the notification (teardown drops
// the exception-port right before the message lands), so after a notification
// the reader drains for a bounded time, answering any exception message that
// follows, and only then exits. Otherwise it parks until the lifecycle
// destroys the port.
// The reader's context: the port name travels by value (failure-paths F4 —
// the reader never dereferences the child struct for the port, and the
// lifecycle stops and drains the thread with a bound before it returns, so
// freeing the child struct at generation end is safe).
typedef struct reader_ctx {
    sealed_child_t *child;
    mach_port_t port;
    pthread_t thr;
    atomic_int stop;
    atomic_int done;
} reader_ctx_t;

static pthread_mutex_t readers_mu = PTHREAD_MUTEX_INITIALIZER;
typedef struct reader_node {
    reader_ctx_t *ctx;
    struct reader_node *next;
} reader_node_t;
static reader_node_t *readers;

// Removes and returns the reader context registered for `child` (NULL if none).
static reader_ctx_t *reader_take(sealed_child_t *child) {
    pthread_mutex_lock(&readers_mu);
    reader_node_t **pp = &readers;
    while (*pp != NULL && (*pp)->ctx->child != child) pp = &(*pp)->next;
    reader_node_t *n = *pp;
    if (n == NULL) {
        pthread_mutex_unlock(&readers_mu);
        return NULL;
    }
    *pp = n->next;
    pthread_mutex_unlock(&readers_mu);
    reader_ctx_t *ctx = n->ctx;
    free(n);
    return ctx;
}

// Bounded wait for the reader to confirm exit: on confirmation the context is
// freed and *ctxp cleared; if the bound expires the detached thread is still
// finishing, so the context is left allocated (freeing it would hand the
// thread a dangling pointer).
static void reader_wait(reader_ctx_t **ctxp) {
    reader_ctx_t *ctx = *ctxp;
    if (ctx == NULL) return;
    uint64_t until = now_ns() + READER_JOIN_NS;
    while (atomic_load(&ctx->done) == 0 && now_ns() < until) usleep(10000);
    if (atomic_load(&ctx->done) != 0) {
        free(ctx);
        *ctxp = NULL;
    }
}

static void *spawn_port_reader(void *arg) {
    reader_ctx_t *ctx = arg;
    sealed_child_t *child = ctx->child;
    union {
        exc_msg_t exc;
        uint8_t raw[4096];
    } msg;
    for (;;) {
        if (atomic_load(&ctx->stop) != 0) break;
        memset(&msg, 0, sizeof msg);
        // Bounded rounds so the reader observes `stop` while parked; a queued
        // exception message is serviced in the round it arrives in.
        kern_return_t kr = mach_msg(&msg.exc.h, MACH_RCV_MSG | MACH_RCV_LARGE | MACH_RCV_TIMEOUT,
                                    0, sizeof msg, ctx->port, READER_DRAIN_ROUND_MS,
                                    MACH_PORT_NULL);
        if (kr == MACH_RCV_TIMED_OUT) continue;
        if (kr != KERN_SUCCESS) break;
        if (msg.exc.h.msgh_bits & MACH_MSGH_BITS_COMPLEX) {
            service_exc_msg(child, &msg.exc);
            continue;
        }
        uint64_t until = now_ns() + READER_DRAIN_NS;
        while (atomic_load(&ctx->stop) == 0 && now_ns() < until) {
            memset(&msg, 0, sizeof msg);
            kr = mach_msg(&msg.exc.h, MACH_RCV_MSG | MACH_RCV_LARGE | MACH_RCV_TIMEOUT,
                          0, sizeof msg, ctx->port, READER_DRAIN_ROUND_MS, MACH_PORT_NULL);
            if (kr == MACH_RCV_TIMED_OUT) continue;
            if (kr != KERN_SUCCESS) break;
            if (msg.exc.h.msgh_bits & MACH_MSGH_BITS_COMPLEX) {
                service_exc_msg(child, &msg.exc);
                until = now_ns() + READER_DRAIN_NS;
            }
        }
        break;
    }
    atomic_store(&ctx->done, 1);
    return NULL;
}

int sealed_status_ok(uint32_t status, uint32_t measured) {
    if ((status & SEALED_CS_REQUIRED) != SEALED_CS_REQUIRED) return 1;
    if ((status & SEALED_CS_FORBIDDEN) != 0) return 2;
    if (measured != 0 && status != measured) return 3;
    return 0;
}

// Takes the spawn port's name out of the guarded field; the caller releases
// the rights exactly once.
static mach_port_t take_spawn_port(sealed_child_t *child) {
    if (child->lock == NULL) return MACH_PORT_NULL;
    pthread_mutex_t *mu = child->lock;
    pthread_mutex_lock(mu);
    mach_port_t p = child->spawn_port;
    child->spawn_port = MACH_PORT_NULL;
    pthread_mutex_unlock(mu);
    return p;
}

// Releases the spawn-time port on every teardown path, once (failure-paths F3,
// r2-inheritance F4): one deallocate releases a reference to each right the
// task holds for the name — the receive right, plus a MAKE_SEND right still
// held on pre-spawn failure paths — and wakes a parked reader.
static void release_spawn_port(sealed_child_t *child) {
    mach_port_t p = take_spawn_port(child);
    if (MACH_PORT_VALID(p)) (void)mach_port_deallocate(mach_task_self(), p);
}

// Kills the child if it has not been reaped (so the pid cannot have been
// reused), then reaps it. Order matters: while the child is held in exit
// behind its exception message it is not reapable, so the lifecycle first
// gives the reader its window to answer (the measured release path), and only
// if the child is still not reapable stops the reader, destroys the port (the
// measured fallback) and reaps with a bounded loop. If the child is still not
// reapable after that, it is left unreaped and kill-guarded rather than
// blocking the caller (recorded in impl/NOTES.md). Every path also releases
// the spawn port's receive right (once) and drains the reader with a bound
// before returning, so the engine may free the child struct afterwards.
void sealed_child_kill(sealed_child_t *child) {
    if (child == NULL || child->lock == NULL) return;
    pthread_mutex_t *mu = child->lock;
    reader_ctx_t *ctx = reader_take(child);
    pthread_mutex_lock(mu);
    if (!child->reaped && child->pid > 0) kill(child->pid, SIGKILL);
    pthread_mutex_unlock(mu);
    if (child->pid <= 0) {
        // Failures before the child exists: no kill, but the port and the
        // reader still need their teardown (r2-inheritance F4).
        if (ctx != NULL) atomic_store(&ctx->stop, 1);
        reader_wait(&ctx);
        release_spawn_port(child);
        return;
    }

    uint64_t until = now_ns() + KILL_GRACE_NS;
    for (;;) {
        if (child_dead_unreaped(child->pid)) break;
        if (now_ns() >= until) break;
        usleep(20000);
    }
    if (!child_dead_unreaped(child->pid)) {
        // FIXES-2 9: the reader stayed live through the SIGKILL and this
        // grace window, so an exception message arriving in either is
        // answered on the measured answer-first path. Only now is it told to
        // stop (one last round answers anything already queued) and the port
        // destroyed, so the reader is never parked on a destroyed (and
        // potentially reused) port name.
        if (ctx != NULL) atomic_store(&ctx->stop, 1);
        reader_wait(&ctx);
        release_spawn_port(child);
        until = now_ns() + KILL_REAP_NS;
        while (!child_dead_unreaped(child->pid) && now_ns() < until) usleep(20000);
    }
    int st;
    int reaped_now = 0;
    for (;;) {
        pid_t w = waitpid(child->pid, &st, WNOHANG);
        if (w == child->pid) {
            reaped_now = 1;
            break;
        }
        if (w < 0 && errno == EINTR) continue;
        if (w < 0) break;
        if (now_ns() >= until) break;
        usleep(20000);
    }
    // failure-paths F2: `reaped` is set only after a real reap — an unreapable
    // child (still held behind its exception message) stays kill-guarded, as
    // the comment above this function promises; the timed-out waitpid loop
    // must not disable the guard.
    pthread_mutex_lock(mu);
    child->reaped = reaped_now;
    pthread_mutex_unlock(mu);
    // The grace loop can also exit via the child-dead branch, which never set
    // the stop flag; stop the reader here so the wait below stays bounded.
    if (ctx != NULL) atomic_store(&ctx->stop, 1);
    reader_wait(&ctx);
    release_spawn_port(child);
}

// Outside listing of the suspended child's descriptors (R2; PROC_PIDLISTFDS
// works same-uid on a hardened child, E5): every descriptor must be
// /dev/null, except at most one AF_UNIX socket (the rendezvous socket, at
// whatever descriptor the child connected on; the backend-specific table
// applies only to the post-acknowledgement comparison, which is not this
// module's check).
static int fds_ok(pid_t pid) {
    // A fixed buffer silently truncates a child with more descriptors than it
    // holds entries (failure-paths F6): a full buffer may be truncated, so
    // grow and re-list until the listing comes back short of the buffer, and
    // fail closed if it never does.
    size_t cap = 64;
    struct proc_fdinfo *fds = NULL;
    int n = 0;
    for (;;) {
        free(fds);
        fds = malloc(cap * sizeof *fds);
        if (fds == NULL) return 0;
        n = (int)proc_pidinfo(pid, PROC_PIDLISTFDS, 0, fds, (int)(cap * sizeof *fds));
        if (n <= 0) {
            free(fds);
            return 0;
        }
        if ((size_t)n < cap * sizeof *fds) break;
        if (cap >= 65536) {
            free(fds);
            return 0;
        }
        cap *= 2;
    }
    size_t count = (size_t)n / sizeof fds[0];
    int ok = 0;
    int unix_sockets = 0;
    for (size_t i = 0; i < count; i++) {
        if (fds[i].proc_fdtype == PROX_FDTYPE_VNODE) {
            struct vnode_fdinfowithpath vi;
            if (proc_pidfdinfo(pid, fds[i].proc_fd, PROC_PIDFDVNODEPATHINFO, &vi, sizeof vi) <= 0)
                goto out;
            if (strcmp(vi.pvip.vip_path, "/dev/null") != 0) goto out;
        } else if (fds[i].proc_fdtype == PROX_FDTYPE_SOCKET) {
            if (++unix_sockets > 1) goto out;
            struct socket_fdinfo sfi;
            if (proc_pidfdinfo(pid, fds[i].proc_fd, PROC_PIDFDSOCKETINFO, &sfi, sizeof sfi) <= 0)
                goto out;
            if (sfi.psi.soi_family != AF_UNIX) goto out;
        } else {
            goto out;
        }
    }
    ok = 1;
out:
    free(fds);
    return ok;
}

static void cleanup_dir(const char *dir) {
    char path[512];
    if (snprintf(path, sizeof path, "%s/rv", dir) < (int)sizeof path) unlink(path);
    if (snprintf(path, sizeof path, "%s/metal-cache", dir) < (int)sizeof path) rmdir(path);
    rmdir(dir);
}

// Shared failure exit: close what is open, kill + reap the child through the
// lifecycle, remove the per-spawn directory.
static int spawn_fail(sealed_child_t *child, int listen_fd, int conn_fd, int code) {
    if (listen_fd >= 0) close(listen_fd);
    if (conn_fd >= 0) close(conn_fd);
    sealed_child_kill(child);
    if (child->dir[0]) cleanup_dir(child->dir);
    return code;
}

int sealed_spawn_child(const sealed_child_spec_t *spec, sealed_child_t *child) {
    memset(child, 0, sizeof *child);
    child->channel_fd = -1;
    child->spawn_port = MACH_PORT_NULL;
    if (spec == NULL || spec->path == NULL || spec->argv == NULL || spec->argv[0] == NULL)
        return SEALED_SPAWN_ATTR;
    // The spawn chdirs the child into its fresh directory before exec, so a
    // relative executable path would resolve in the wrong directory (and R2
    // requires absolute, canonical paths): refuse anything not absolute.
    if (spec->path[0] != '/') return SEALED_SPAWN_ATTR;
    if (spec->requirement == NULL || spec->requirement_len == 0) return SEALED_SPAWN_REQUIREMENT;

    pthread_mutex_t *mu = malloc(sizeof *mu);
    if (mu == NULL || pthread_mutex_init(mu, NULL) != 0) {
        free(mu);
        return SEALED_SPAWN_PORT;
    }
    child->lock = mu;

    // R2: one fresh empty 0700 directory per spawn, created exclusively
    // (mkdtemp); a collision fails the spawn.
    char base[PATH_MAX];
    size_t blen = confstr(_CS_DARWIN_USER_TEMP_DIR, base, sizeof base);
    if (blen == 0 || blen > sizeof base) return SEALED_SPAWN_DIR;
    if (blen > 1 && base[blen - 2] != '/') {
        if (blen + 1 > sizeof base) return SEALED_SPAWN_DIR;
        base[blen - 1] = '/';
        base[blen] = '\0';
    }
    char dirtpl[256];
    if (snprintf(dirtpl, sizeof dirtpl, "%ssealed-child-XXXXXXXX", base) >= (int)sizeof dirtpl)
        return SEALED_SPAWN_DIR;
    if (mkdtemp(dirtpl) == NULL) return SEALED_SPAWN_DIR;
    // R2 wants canonical paths (r2r3-identity D1): the kernel resolves the
    // chdir target to the /private/var/... form, so realpath the directory
    // once and use that everywhere (dir, HOME/TMPDIR env, rv socket).
    char canon[PATH_MAX];
    if (realpath(dirtpl, canon) == NULL) {
        rmdir(dirtpl);
        return SEALED_SPAWN_DIR;
    }
    if (snprintf(child->dir, sizeof child->dir, "%s", canon) >= (int)sizeof child->dir) {
        rmdir(canon);
        return SEALED_SPAWN_DIR;
    }
    const char *dir = child->dir;

    char path[512];
    if (snprintf(path, sizeof path, "%s/metal-cache", dir) >= (int)sizeof path ||
        mkdir(path, 0700) != 0)
        return spawn_fail(child, -1, -1, SEALED_SPAWN_DIR);
    if (snprintf(path, sizeof path, "%s/rv", dir) >= (int)sizeof path ||
        strlen(path) >= sizeof ((struct sockaddr_un *)0)->sun_path)   // sun_path is 104 bytes on macOS
        return spawn_fail(child, -1, -1, SEALED_SPAWN_DIR);
    int listen_fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (listen_fd < 0) return spawn_fail(child, -1, -1, SEALED_SPAWN_DIR);
    struct sockaddr_un sa;
    memset(&sa, 0, sizeof sa);
    sa.sun_family = AF_UNIX;
    strncpy(sa.sun_path, path, sizeof sa.sun_path - 1);
    unlink(path);
    if (bind(listen_fd, (struct sockaddr *)&sa, sizeof sa) != 0 || listen(listen_fd, 1) != 0)
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_DIR);
    const char *rvpath = strdup(path);
    if (rvpath == NULL) return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_DIR);

    // R1 spawn-time port: one receive right per child generation, with a
    // no-senders request so the reader learns when the last send right drops.
    mach_port_t port;
    if (mach_port_allocate(mach_task_self(), MACH_PORT_RIGHT_RECEIVE, &port) != KERN_SUCCESS) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_PORT);
    }
    // FIXES-2 4a, measured (impl/NOTES.md): a no-senders notification
    // registered on a senderless port is delivered immediately (msgh_id 70),
    // which would end the reader before the child even spawns; a held MAKE_SEND
    // right suppresses it. Insert the send right first, register after. The
    // send right is deallocated as soon as the spawn call returns (R1),
    // leaving the child (through the kernel) the only sender.
    if (mach_port_insert_right(mach_task_self(), port, port, MACH_MSG_TYPE_MAKE_SEND) !=
        KERN_SUCCESS) {
        (void)mach_port_deallocate(mach_task_self(), port);
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_PORT);
    }
    mach_port_name_t previous = MACH_PORT_NULL;
    kern_return_t kr = mach_port_request_notification(mach_task_self(), port,
                                                      MACH_NOTIFY_NO_SENDERS, 0, port,
                                                      MACH_MSG_TYPE_MAKE_SEND_ONCE,
                                                      &previous);
    if (kr != KERN_SUCCESS) {
        (void)mach_port_deallocate(mach_task_self(), port);
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_PORT);
    }
    if (MACH_PORT_VALID(previous)) mach_port_deallocate(mach_task_self(), previous);
    pthread_mutex_lock(mu);
    child->spawn_port = port;
    pthread_mutex_unlock(mu);
    reader_ctx_t *ctx = malloc(sizeof *ctx);
    if (ctx == NULL) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_PORT);
    }
    ctx->child = child;
    ctx->port = port;
    atomic_store(&ctx->stop, 0);
    atomic_store(&ctx->done, 0);
    reader_node_t *node = malloc(sizeof *node);
    if (node == NULL) {
        free(ctx);
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_PORT);
    }
    node->ctx = ctx;
    pthread_mutex_lock(&readers_mu);
    node->next = readers;
    readers = node;
    pthread_mutex_unlock(&readers_mu);
    if (pthread_create(&ctx->thr, NULL, spawn_port_reader, ctx) != 0) {
        reader_take(child);
        free(ctx);
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_PORT);
    }
    pthread_detach(ctx->thr);

    // R2 spawn attributes: cloexec default, start suspended, own process
    // group, all signals default, empty mask.
    posix_spawnattr_t attr;
    posix_spawnattr_t *ap = &attr;
    if (posix_spawnattr_init(ap) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_ATTR);
    }
    int failed = 0;
    // Failing stage -> result code (FIXES 1-2): only the requirement attach and
    // the spawn call itself have their own codes; everything else is ATTR.
    int failed_code = 0;
    if (!failed) {
        short flags = (short)(POSIX_SPAWN_CLOEXEC_DEFAULT | POSIX_SPAWN_START_SUSPENDED |
                              POSIX_SPAWN_SETPGROUP |
                              0);
        sigset_t all, empty;
        failed = posix_spawnattr_setflags(ap, flags) != 0 ||
                 posix_spawnattr_setpgroup(ap, 0) != 0 || sigfillset(&all) != 0 ||
                 sigemptyset(&empty) != 0 || posix_spawnattr_setsigdefault(ap, &all) != 0 ||
                 posix_spawnattr_setsigmask(ap, &empty) != 0;
    }
    // R1: the child's exception ports are set explicitly for the same two
    // masks the guard sets — the spawn-time port on the guard mask, no port
    // on the null mask — so nothing inherited reaches the child (E2: a
    // spawned task inherits its parent's task ports for every unset type).
    // One call per mask.
    if (!failed &&
        posix_spawnattr_setexceptionports_np(ap, SEALED_GUARD_MASK, port, SPAWN_BEHAVIOR,
                                             THREAD_STATE_NONE) != 0)
        failed = 1;
    if (!failed &&
        posix_spawnattr_setexceptionports_np(ap, SEALED_NULL_MASK, MACH_PORT_NULL,
                                             EXCEPTION_DEFAULT, THREAD_STATE_NONE) != 0)
        failed = 1;
    // A-13: the kernel-enforced launch code requirement naming the pinned CD
    // hash, attached with the undeclared libxpc SPI (E1; 82-byte
    // NSTask.launchRequirementData encoding). Failing to attach is fatal
    // (A-14): without it there is no v2 sealed serving.
    if (!failed && amfi_launch_constraint_set_spawnattr(ap, spec->requirement,
                                                         spec->requirement_len) != 0)
        failed = 1, failed_code = SEALED_SPAWN_REQUIREMENT;

    // R2: fd 0-2 on /dev/null, working directory TMPDIR; with
    // POSIX_SPAWN_CLOEXEC_DEFAULT everything else is closed.
    posix_spawn_file_actions_t fa;
    // r1-handler F: `fa` is only initialized here when nothing failed before;
    // destroying an uninitialized (stack-garbage) file-actions object is UB.
    int fa_init = 0;
    if (!failed) {
        if (posix_spawn_file_actions_init(&fa) != 0) failed = 1;
        else fa_init = 1;
    }
    if (!failed) {
        failed = posix_spawn_file_actions_addchdir(&fa, dir) != 0 ||
                 posix_spawn_file_actions_addopen(&fa, 0, "/dev/null", O_RDONLY, 0) != 0 ||
                 posix_spawn_file_actions_addopen(&fa, 1, "/dev/null", O_WRONLY, 0) != 0 ||
                 posix_spawn_file_actions_addopen(&fa, 2, "/dev/null", O_WRONLY, 0) != 0;
    }

    // R2: environment exactly HOME and TMPDIR (no contingency variable is in
    // force on this build); argv the sealed form.
    char envbuf[2][300];
    char *envp[3] = {envbuf[0], envbuf[1], NULL};
    if (!failed &&
        (snprintf(envbuf[0], sizeof envbuf[0], "HOME=%s", dir) >= (int)sizeof envbuf[0] ||
         snprintf(envbuf[1], sizeof envbuf[1], "TMPDIR=%s", dir) >= (int)sizeof envbuf[1]))
        failed = 1;

    pid_t pid = -1;
    if (!failed && posix_spawn(&pid, spec->path, &fa, ap, (char *const *)spec->argv, envp) != 0)
        failed = 1, failed_code = SEALED_SPAWN_EXEC;

    if (fa_init) posix_spawn_file_actions_destroy(&fa);
    posix_spawnattr_destroy(ap);
    if (failed) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1,
                          pid >= 0 || failed_code == SEALED_SPAWN_EXEC ? SEALED_SPAWN_EXEC
                          : failed_code != 0 ? failed_code : SEALED_SPAWN_ATTR);
    }

    // R1: deallocate our send right as soon as the spawn call returns; the
    // child (through the kernel's exception-port copy) holds the only senders.
    (void)mach_port_deallocate(mach_task_self(), port);
    pthread_mutex_lock(mu);
    child->pid = pid;
    pthread_mutex_unlock(mu);

    // R3, suspended phase: the kernel's record of the executed image, before
    // resume. Never reap during the spawn window.
    uint8_t hash[SEALED_CDHASH_LEN];
    if (csops(pid, CS_OPS_CDHASH, hash, sizeof hash) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_DIED);
    }
    if (memcmp(hash, spec->cdhash, SEALED_CDHASH_LEN) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_CDHASH);
    }
    uint32_t st = 0;
    if (csops(pid, CS_OPS_STATUS, &st, sizeof st) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_DIED);
    }
    child->status_seen_suspended = st;
    if ((st & SEALED_CS_KILLED) || sealed_status_ok(st, spec->status_suspended) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_STATUS);
    }

    // R2: list the child's descriptors from outside and refuse on anything
    // but /dev/null plus at most the one rendezvous socket.
    if (!fds_ok(pid)) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, -1, SEALED_SPAWN_FDS);
    }

    // Resume. pid_resume is refused with EPERM for an unprivileged engine
    // (measured, E1); SIGCONT resumes a START_SUSPENDED child (measured).
    (void)pid_resume(pid);
    kill(pid, SIGCONT);

    // R2 rendezvous: one connection with its first byte within the deadline.
    uint64_t deadline = now_ns() + (uint64_t)spec->rendezvous_ms * 1000000ull;
    int conn = -1;
    for (;;) {
        if (child_dead_unreaped(pid)) {
            free((void *)rvpath);
            return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_DIED);
        }
        uint64_t left = deadline > now_ns() ? deadline - now_ns() : 0;
        if (left == 0) {
            free((void *)rvpath);
            return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_RENDEZVOUS);
        }
        // Cap each poll slice so child death is re-checked promptly (FIXES 10:
        // a crash before the rendezvous otherwise rides the whole deadline out
        // in one poll and surfaces as RENDEZVOUS instead of DIED).
        uint64_t slice_ms = left / 1000000ull + 1;
        if (slice_ms > 100) slice_ms = 100;
        struct pollfd pf = {.fd = listen_fd, .events = POLLIN, .revents = 0};
        int pr = poll(&pf, 1, (int)slice_ms);
        if (pr > 0 && (pf.revents & POLLIN)) {
            conn = accept(listen_fd, NULL, NULL);
            if (conn >= 0) break;
        } else if (pr < 0 && errno == EINTR) {
            continue;
        }
        // pr == 0 or unexpected revents: the next round re-checks death and
        // the deadline.
    }
    char byte = 0;
    for (;;) {
        // r2r3-identity D2: re-check child death each round, so a child that
        // dies after connecting is refused DIED, not RENDEZVOUS at the deadline.
        if (child_dead_unreaped(pid)) {
            free((void *)rvpath);
            return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_DIED);
        }
        uint64_t left = deadline > now_ns() ? deadline - now_ns() : 0;
        if (left == 0) {
            free((void *)rvpath);
            return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_RENDEZVOUS);
        }
        struct pollfd pf = {.fd = conn, .events = POLLIN, .revents = 0};
        uint64_t slice_ms = left / 1000000ull + 1;
        if (slice_ms > 100) slice_ms = 100;   // re-check child death each slice
        int pr = poll(&pf, 1, (int)slice_ms);
        if (pr < 0 && errno == EINTR) continue;
        if (pr > 0 && (pf.revents & (POLLHUP | POLLERR)) && !(pf.revents & POLLIN)) {
            // Peer closed without sending: the connection is over either way.
            free((void *)rvpath);
            return spawn_fail(child, listen_fd, conn,
                              child_dead_unreaped(pid) ? SEALED_SPAWN_DIED : SEALED_SPAWN_RENDEZVOUS);
        }
        if (pr <= 0) continue;
        ssize_t n = read(conn, &byte, 1);
        if (n == 1) break;
        if (n < 0 && errno == EINTR) continue;
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, conn,
                          child_dead_unreaped(pid) ? SEALED_SPAWN_DIED : SEALED_SPAWN_RENDEZVOUS);
    }

    // R2: the token is computed when read, from the process that sent the
    // byte (E3); csops_audittoken refuses it once that process has exec'd or
    // its pid was reused (E4). On failure the engine never sends anything on
    // the connection.
    audit_token_t token;
    socklen_t tlen = sizeof token;
    if (getsockopt(conn, SOL_LOCAL, LOCAL_PEERTOKEN, &token, &tlen) != 0 ||
        token_pid(&token) != pid) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_TOKEN);
    }
    if (csops_audittoken(pid, CS_OPS_CDHASH, hash, sizeof hash, &token) != 0 ||
        memcmp(hash, spec->cdhash, SEALED_CDHASH_LEN) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_TOKEN);
    }
    uint32_t rst = 0;
    if (csops_audittoken(pid, CS_OPS_STATUS, &rst, sizeof rst, &token) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_TOKEN);
    }
    child->status_seen_running = rst;
    if ((rst & SEALED_CS_KILLED) || sealed_status_ok(rst, spec->status_running) != 0) {
        free((void *)rvpath);
        return spawn_fail(child, listen_fd, conn, SEALED_SPAWN_STATUS);
    }

    child->channel_fd = conn;
    close(listen_fd);
    unlink(rvpath);
    free((void *)rvpath);
    return SEALED_SPAWN_OK;
}
