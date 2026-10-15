// R1 exception-port guard. See include/sealed_guard.h and the spec's R1.
// Built as one object without LTO (Makefile / project settings), so the
// machine code that fault tests exercise is the code that ships.
#include "sealed_guard.h"

#include <pthread.h>
#include <signal.h>
#include <string.h>
#include <unistd.h>

_Static_assert((SEALED_GUARD_MASK & SEALED_NULL_MASK) == 0, "guard and null masks overlap");
_Static_assert((SEALED_GUARD_MASK | SEALED_NULL_MASK) ==
                   (EXC_MASK_ALL | EXC_MASK_CRASH | EXC_MASK_CORPSE_NOTIFY),
               "guard and null masks must cover exactly the SDK's valid types");
_Static_assert(SEALED_VALID_MASK == (EXC_MASK_ALL | EXC_MASK_CRASH | EXC_MASK_CORPSE_NOTIFY),
               "valid mask drifted from the SDK");

#define GUARD_BEHAVIOR ((exception_behavior_t)(EXCEPTION_DEFAULT | MACH_EXCEPTION_CODES))

static mach_port_t guard_port = MACH_PORT_NULL;

// The reader's only action is _exit(113). It never replies, never reads the
// message's contents and never touches the rights it carries: the process is
// ending, and the kernel reclaims everything with it.
static void *guard_reader(void *unused) {
    (void)unused;
    struct {
        mach_msg_header_t header;
        char body[1024];
    } msg;
    for (;;) {
        kern_return_t kr = mach_msg(&msg.header, MACH_RCV_MSG | MACH_RCV_LARGE, 0, sizeof msg,
                                    guard_port, MACH_MSG_TIMEOUT_NONE, MACH_PORT_NULL);
        if (kr == MACH_RCV_INTERRUPTED) continue;
        _exit(113);
    }
    return NULL;
}

static void fatal_signal(int sig) {
    (void)sig;
    _exit(113);
}

int sealed_behavior_exposes(exception_behavior_t behavior) {
    exception_behavior_t base = behavior & ~MACH_EXCEPTION_MASK;
    return !(base == EXCEPTION_IDENTITY_PROTECTED || base == EXCEPTION_STATE_IDENTITY_PROTECTED);
}

// Adds the ports one swap returned to `old`, combining entries by port name.
// Returns 0, or 1 if `old` is full.
static int record_old(sealed_old_port_t *old, unsigned *count, const exception_mask_t *masks,
                      mach_msg_type_number_t n, const mach_port_t *ports,
                      const exception_behavior_t *behaviors, const thread_state_flavor_t *flavors) {
    for (mach_msg_type_number_t i = 0; i < n; i++) {
        if (!MACH_PORT_VALID(ports[i])) continue;
        unsigned j = 0;
        for (; j < *count; j++) {
            if (old[j].name == ports[i] && old[j].behavior == behaviors[i] && old[j].flavor == flavors[i]) {
                old[j].mask |= masks[i];
                break;
            }
        }
        if (j == *count) {
            if (*count >= SEALED_MAX_OLD_PORTS) return 1;
            old[j] = (sealed_old_port_t){ports[i], masks[i], behaviors[i], flavors[i]};
            (*count)++;
        }
    }
    return 0;
}

static void release_rights(mach_msg_type_number_t n, const mach_port_t *ports) {
    for (mach_msg_type_number_t i = 0; i < n; i++)
        if (MACH_PORT_VALID(ports[i])) mach_port_deallocate(mach_task_self(), ports[i]);
}

static int clear_thread_ports(void) {
    thread_act_array_t threads;
    mach_msg_type_number_t n = 0;
    if (task_threads(mach_task_self(), &threads, &n) != KERN_SUCCESS) return 1;
    int failed = 0;
    for (mach_msg_type_number_t i = 0; i < n; i++) {
        kern_return_t kr = thread_set_exception_ports(threads[i], SEALED_VALID_MASK, MACH_PORT_NULL,
                                                      EXCEPTION_DEFAULT, THREAD_STATE_NONE);
        // A thread that exited after task_threads is skipped, not a failure.
        if (kr != KERN_SUCCESS && kr != KERN_TERMINATED && kr != MACH_SEND_INVALID_DEST) failed = 1;
        mach_port_deallocate(mach_task_self(), threads[i]);
    }
    vm_deallocate(mach_task_self(), (vm_address_t)threads, n * sizeof(thread_act_t));
    return failed;
}

int sealed_guard_install(sealed_old_port_t *old, unsigned *old_count) {
    *old_count = 0;
    task_t self = mach_task_self();

    // 1. The port and the thread that waits on it exist before the port is
    //    set, so a fault right after the swap cannot queue a message nobody reads.
    if (mach_port_allocate(self, MACH_PORT_RIGHT_RECEIVE, &guard_port) != KERN_SUCCESS) return 1;
    if (mach_port_insert_right(self, guard_port, guard_port, MACH_MSG_TYPE_MAKE_SEND) != KERN_SUCCESS) return 2;
    pthread_t reader;
    if (pthread_create(&reader, NULL, guard_reader, NULL) != 0) return 3;
    pthread_detach(reader);

    // 2. Guard mask -> the guard's port.
    exception_mask_t masks[EXC_TYPES_COUNT];
    mach_port_t ports[EXC_TYPES_COUNT];
    exception_behavior_t behaviors[EXC_TYPES_COUNT];
    thread_state_flavor_t flavors[EXC_TYPES_COUNT];
    mach_msg_type_number_t n = EXC_TYPES_COUNT;
    if (task_swap_exception_ports(self, SEALED_GUARD_MASK, guard_port, GUARD_BEHAVIOR, THREAD_STATE_NONE,
                                  masks, &n, ports, behaviors, flavors) != KERN_SUCCESS) return 4;
    int full = record_old(old, old_count, masks, n, ports, behaviors, flavors);

    // 3. Null mask -> no port.
    exception_mask_t masks2[EXC_TYPES_COUNT];
    mach_port_t ports2[EXC_TYPES_COUNT];
    exception_behavior_t behaviors2[EXC_TYPES_COUNT];
    thread_state_flavor_t flavors2[EXC_TYPES_COUNT];
    mach_msg_type_number_t n2 = EXC_TYPES_COUNT;
    if (task_swap_exception_ports(self, SEALED_NULL_MASK, MACH_PORT_NULL, EXCEPTION_DEFAULT, THREAD_STATE_NONE,
                                  masks2, &n2, ports2, behaviors2, flavors2) != KERN_SUCCESS) return 5;
    full |= record_old(old, old_count, masks2, n2, ports2, behaviors2, flavors2);

    // The swaps returned send rights; record first (names combine), then release.
    release_rights(n, ports);
    release_rights(n2, ports2);
    if (full) return 6;

    // 4. A thread-level port takes precedence over the task's: clear every
    //    existing thread for every valid type.
    if (clear_thread_ports() != 0) return 7;

    // 5. Fatal signals end the process at once (R1, R7's fixed expectation).
    static const int fatal[] = {SIGSEGV, SIGBUS, SIGILL, SIGFPE, SIGABRT, SIGTRAP,
                                SIGSYS,  SIGEMT, SIGQUIT, SIGXCPU, SIGXFSZ};
    for (size_t i = 0; i < sizeof fatal / sizeof fatal[0]; i++) {
        struct sigaction sa;
        memset(&sa, 0, sizeof sa);
        sa.sa_handler = fatal_signal;
        sigemptyset(&sa.sa_mask);
        if (sigaction(fatal[i], &sa, NULL) != 0) return 8;
    }

    // 6. Mask gate.
    if (sealed_guard_readback() != 0) return 9;
    return 0;
}

int sealed_guard_readback(void) {
    if (!MACH_PORT_VALID(guard_port)) return 1;
    exception_mask_t masks[EXC_TYPES_COUNT];
    mach_port_t ports[EXC_TYPES_COUNT];
    exception_behavior_t behaviors[EXC_TYPES_COUNT];
    thread_state_flavor_t flavors[EXC_TYPES_COUNT];
    mach_msg_type_number_t n = EXC_TYPES_COUNT;
    if (task_get_exception_ports(mach_task_self(), SEALED_VALID_MASK, masks, &n, ports, behaviors,
                                 flavors) != KERN_SUCCESS)
        return 2;
    exception_mask_t guarded = 0;
    int bad = 0;
    for (mach_msg_type_number_t i = 0; i < n; i++) {
        if (MACH_PORT_VALID(ports[i])) {
            if (ports[i] != guard_port || behaviors[i] != GUARD_BEHAVIOR ||
                flavors[i] != THREAD_STATE_NONE || (masks[i] & SEALED_NULL_MASK))
                bad = 1;
            else
                guarded |= masks[i];
        }
    }
    release_rights(n, ports);
    if (bad) return 3;
    return guarded == SEALED_GUARD_MASK ? 0 : 4;
}

int sealed_guard_threads_clear(void) {
    thread_act_array_t threads;
    mach_msg_type_number_t count = 0;
    if (task_threads(mach_task_self(), &threads, &count) != KERN_SUCCESS) return 1;
    int bad = 0;
    for (mach_msg_type_number_t t = 0; t < count; t++) {
        exception_mask_t masks[EXC_TYPES_COUNT];
        mach_port_t ports[EXC_TYPES_COUNT];
        exception_behavior_t behaviors[EXC_TYPES_COUNT];
        thread_state_flavor_t flavors[EXC_TYPES_COUNT];
        mach_msg_type_number_t n = EXC_TYPES_COUNT;
        kern_return_t kr = thread_get_exception_ports(threads[t], SEALED_VALID_MASK, masks, &n, ports,
                                                      behaviors, flavors);
        if (kr == KERN_SUCCESS) {
            for (mach_msg_type_number_t i = 0; i < n; i++)
                if (MACH_PORT_VALID(ports[i])) bad = 1;
            release_rights(n, ports);
        } else if (kr != KERN_TERMINATED && kr != MACH_SEND_INVALID_DEST) {
            bad = 1;
        }
        mach_port_deallocate(mach_task_self(), threads[t]);
    }
    vm_deallocate(mach_task_self(), (vm_address_t)threads, count * sizeof(thread_act_t));
    return bad;
}

int sealed_old_ports_match(const sealed_old_port_t *old, unsigned old_count,
                           const sealed_old_port_t *expected, unsigned expected_count) {
    if (old_count != expected_count) return 1;
    for (unsigned i = 0; i < old_count; i++) {
        if (sealed_behavior_exposes(old[i].behavior)) return 2;
        int found = 0;
        for (unsigned j = 0; j < expected_count; j++)
            if (old[i].mask == expected[j].mask && old[i].behavior == expected[j].behavior &&
                old[i].flavor == expected[j].flavor)
                found = 1;
        if (!found) return 3;
    }
    return 0;
}
