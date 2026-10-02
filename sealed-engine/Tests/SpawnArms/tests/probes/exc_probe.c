// FIXES-2 4(b)/8 probe: which Mach exception type does each raising child
// deliver to a task-level spawn-time port, and can EXC_SOFTWARE or
// EXC_CORPSE_NOTIFY be provoked unprivileged? For each child: set the port on
// the CRASH|SOFTWARE|CORPSE mask via posix_spawnattr_setexceptionports_np
// (the module's SPI/behavior/flavor), spawn suspended, resume, receive every
// message until a 3 s window closes, print msgh_id + exception type + codes,
// reply KERN_SUCCESS (the module's measured rule), then reap.
#include <mach/mach.h>
#include <signal.h>
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>

extern int posix_spawnattr_setexceptionports_np(posix_spawnattr_t *,
    exception_mask_t, mach_port_t, exception_behavior_t, thread_state_flavor_t);

#define SPAWN_BEHAVIOR ((exception_behavior_t)(EXCEPTION_DEFAULT | MACH_EXCEPTION_CODES))

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

static const char *exc_name(int e) {
    switch (e) {
    case EXC_BAD_ACCESS: return "EXC_BAD_ACCESS";
    case EXC_BAD_INSTRUCTION: return "EXC_BAD_INSTRUCTION";
    case EXC_ARITHMETIC: return "EXC_ARITHMETIC";
    case EXC_SOFTWARE: return "EXC_SOFTWARE";
    case EXC_BREAKPOINT: return "EXC_BREAKPOINT";
    case EXC_CRASH: return "EXC_CRASH";
    case EXC_CORPSE_NOTIFY: return "EXC_CORPSE_NOTIFY";
    default: return "?";
    }
}

static void one_leg(const char *child) {
    mach_port_t port = MACH_PORT_NULL;
    if (mach_port_allocate(mach_task_self(), MACH_PORT_RIGHT_RECEIVE, &port) != KERN_SUCCESS) {
        printf("MEASURED exc_probe %s: port allocate failed\n", child);
        return;
    }
    if (mach_port_insert_right(mach_task_self(), port, port,
                               MACH_MSG_TYPE_MAKE_SEND) != KERN_SUCCESS) {
        printf("MEASURED exc_probe %s: insert_right failed\n", child);
        return;
    }
    posix_spawnattr_t a;
    posix_spawnattr_init(&a);
    posix_spawnattr_setflags(&a, POSIX_SPAWN_START_SUSPENDED);
    exception_mask_t mask = EXC_MASK_CRASH | EXC_MASK_SOFTWARE | EXC_MASK_CORPSE_NOTIFY;
    int spi = posix_spawnattr_setexceptionports_np(&a, mask, port, SPAWN_BEHAVIOR,
                                                   THREAD_STATE_NONE);
    printf("MEASURED exc_probe %s: spawnattr exception ports (CRASH|SOFTWARE|CORPSE) "
           "-> %s (%d)\n", child, spi == 0 ? "accepted" : "REFUSED", spi);
    if (spi != 0) {
        int spi2 = posix_spawnattr_setexceptionports_np(&a, EXC_MASK_CRASH | EXC_MASK_SOFTWARE,
                                                        port, SPAWN_BEHAVIOR, THREAD_STATE_NONE);
        printf("MEASURED exc_probe %s: retry without the CORPSE mask -> %s (%d)\n",
               child, spi2 == 0 ? "accepted" : "REFUSED", spi2);
        if (spi2 != 0) {
            posix_spawnattr_destroy(&a);
            (void)mach_port_deallocate(mach_task_self(), port);
            return;
        }
    }
    extern char **environ;
    char *argv[] = { (char *)child, NULL };
    pid_t pid = -1;
    if (posix_spawn(&pid, child, NULL, &a, argv, environ) != 0 || pid <= 0) {
        printf("MEASURED exc_probe %s: posix_spawn failed (errno context: rc=%d)\n",
               child, (int)pid);
        posix_spawnattr_destroy(&a);
        (void)mach_port_deallocate(mach_task_self(), port);
        return;
    }
    posix_spawnattr_destroy(&a);
    (void)mach_port_deallocate(mach_task_self(), port);   // our send right only
    (void)kill(pid, SIGCONT);

    union { exc_msg_t exc; uint8_t raw[4096]; } msg;
    int msgs = 0;
    for (;;) {
        memset(&msg, 0, sizeof msg);
        kern_return_t kr = mach_msg(&msg.exc.h, MACH_RCV_MSG, 0, sizeof msg.raw, port,
                                    3000, MACH_PORT_NULL);
        if (kr != KERN_SUCCESS) {
            printf("MEASURED exc_probe %s: receive window closed after %d message(s) "
                   "(%s)\n", child, msgs, mach_error_string(kr));
            break;
        }
        printf("MEASURED exc_probe %s: message id=%d size=%u exception=%d (%s) "
               "code0=0x%llx code1=0x%llx\n", child, msg.exc.h.msgh_id,
               msg.exc.h.msgh_size, (int)msg.exc.exception,
               exc_name((int)msg.exc.exception),
               (unsigned long long)msg.exc.code[0], (unsigned long long)msg.exc.code[1]);
        (void)mach_port_deallocate(mach_task_self(), msg.exc.thread.name);
        (void)mach_port_deallocate(mach_task_self(), msg.exc.task.name);
        exc_reply_msg_t rep;
        memset(&rep, 0, sizeof rep);
        rep.h.msgh_bits = MACH_MSGH_BITS(MACH_MSG_TYPE_MOVE_SEND_ONCE, 0);
        rep.h.msgh_size = sizeof rep;
        rep.h.msgh_remote_port = msg.exc.h.msgh_remote_port;
        rep.h.msgh_id = msg.exc.h.msgh_id + 100;
        rep.ndr = NDR_record;
        rep.ret = KERN_SUCCESS;
        (void)mach_msg(&rep.h, MACH_SEND_MSG, sizeof rep, 0, MACH_PORT_NULL,
                       MACH_MSG_TIMEOUT_NONE, MACH_PORT_NULL);
        if (++msgs >= 3) { printf("MEASURED exc_probe %s: message cap\n", child); break; }
    }
    int st = 0;
    if (waitpid(pid, &st, 0) == pid)
        printf("MEASURED exc_probe %s: child reaped status=%#x\n", child, st);
    (void)mach_port_deallocate(mach_task_self(), port);
}

int main(void) {
    one_leg("./exc_abrt");
    one_leg("./exc_trap");
    printf("MEASURED exc_probe: done\n");
    return 0;
}
