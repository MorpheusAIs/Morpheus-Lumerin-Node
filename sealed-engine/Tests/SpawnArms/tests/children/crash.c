// Child that crashes before any rendezvous (R3-T1 "died before verification"
// shape, R2-T2 marker oracle): first instruction in main writes the marker
// named by argv[1], then it faults before touching $TMPDIR/rv. An engine that
// refuses it must show SEALED_SPAWN_DIED with the marker present (it ran) —
// never a success.
#include <fcntl.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc > 1 && argv[1][0] != '\0') {
        int fd = open(argv[1], O_CREAT | O_WRONLY, 0644);
        if (fd >= 0) close(fd);
    }
    volatile int *p = (volatile int *)(long)16;
    *p = 1;
    _exit(0);
}
