// FIXES-2 4(a) measurement probe: is a MACH_NOTIFY_NO_SENDERS notification
// registered on a fresh port with no send rights delivered immediately?
// Variant B: the same registration while the task holds a MAKE_SEND right —
// expected: no delivery. Informational (measurement, not a gate): always
// exits 0; its output is recorded in impl/NOTES.md.
#include <mach/mach.h>
#include <stdio.h>
#include <string.h>

// 1 = message received (id_out set), 0 = timed out, -1 = receive error.
static int try_receive(mach_port_t port, unsigned ms, mach_msg_id_t *id_out) {
    union {
        mach_msg_header_t h;
        uint8_t raw[512];
    } msg;
    memset(&msg, 0, sizeof msg);
    kern_return_t kr = mach_msg(&msg.h, MACH_RCV_MSG | MACH_RCV_LARGE | MACH_RCV_TIMEOUT, 0,
                                sizeof msg, port, ms, MACH_PORT_NULL);
    if (kr == MACH_RCV_TIMED_OUT) return 0;
    if (kr != KERN_SUCCESS) return -1;
    *id_out = msg.h.msgh_id;
    return 1;
}

int main(void) {
    mach_msg_id_t id = -1;
    mach_port_t p = MACH_PORT_NULL;
    mach_port_name_t prev = MACH_PORT_NULL;
    if (mach_port_allocate(mach_task_self(), MACH_PORT_RIGHT_RECEIVE, &p) != KERN_SUCCESS) {
        printf("PROBE nosend: allocate failed\n");
        return 1;
    }
    kern_return_t kr = mach_port_request_notification(mach_task_self(), p,
                                                      MACH_NOTIFY_NO_SENDERS, 0, p,
                                                      MACH_MSG_TYPE_MAKE_SEND_ONCE, &prev);
    if (kr != KERN_SUCCESS) {
        printf("PROBE nosend: request_notification failed kr=%d\n", kr);
        return 1;
    }
    int r = try_receive(p, 0, &id);
    printf("PROBE nosend: senderless port, notification delivered immediately: %s "
           "(rc=%d msgh_id=%d)\n", r == 1 ? "YES" : r == 0 ? "NO" : "ERR", r, id);
    if (r == 0) {
        r = try_receive(p, 200, &id);
        printf("PROBE nosend: delivered within a further 200 ms: %s (rc=%d id=%d)\n",
               r == 1 ? "YES" : r == 0 ? "NO" : "ERR", r, id);
    }

    // Variant B: module shape with a send right held before registration.
    mach_port_t q = MACH_PORT_NULL;
    if (mach_port_allocate(mach_task_self(), MACH_PORT_RIGHT_RECEIVE, &q) != KERN_SUCCESS)
        return 1;
    if (mach_port_insert_right(mach_task_self(), q, q, MACH_MSG_TYPE_MAKE_SEND) != KERN_SUCCESS)
        return 1;
    kr = mach_port_request_notification(mach_task_self(), q, MACH_NOTIFY_NO_SENDERS, 0, q,
                                        MACH_MSG_TYPE_MAKE_SEND_ONCE, &prev);
    if (kr != KERN_SUCCESS) {
        printf("PROBE withsend: request_notification failed kr=%d\n", kr);
        return 1;
    }
    r = try_receive(q, 0, &id);
    if (r == 0) r = try_receive(q, 200, &id);
    printf("PROBE withsend: port holding our MAKE_SEND right, notification "
           "delivered: %s (rc=%d id=%d)\n", r == 1 ? "YES" : r == 0 ? "NO" : "ERR", r, id);
    printf("PROBE done\n");
    return 0;
}
