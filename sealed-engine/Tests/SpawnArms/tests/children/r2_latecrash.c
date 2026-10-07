// Child that crashes late (FIXES-3 item 1): writes its marker, sleeps about
// 3 s, then faults — without ever connecting to $TMPDIR/rv. Pairs with arm
// r1_t5_late_crash (rendezvous deadline 8000 ms): the fixed module must
// return DIED within about 1 s of the crash; the mutant that orders the
// no-senders registration before the send-right insert has its reader drained
// and exited long before the crash, so the crash is unanswered there and the
// call runs to the deadline (RENDEZVOUS).
#include <fcntl.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc > 1 && argv[1][0] != '\0') {
        int fd = open(argv[1], O_CREAT | O_WRONLY, 0644);
        if (fd >= 0) close(fd);
    }
    sleep(3);
    volatile int *p = (volatile int *)(long)16;
    *p = 1;
    _exit(0);
}
