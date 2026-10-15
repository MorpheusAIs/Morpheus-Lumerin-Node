// Child that never connects (R2 rendezvous timeout shape): first instruction
// in main writes the marker named by argv[1], then it sleeps without touching
// $TMPDIR/rv. The engine must kill it and refuse with SEALED_SPAWN_RENDEZVOUS
// (no connection and first byte within the deadline), marker present.
#include <fcntl.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc > 1 && argv[1][0] != '\0') {
        int fd = open(argv[1], O_CREAT | O_WRONLY, 0644);
        if (fd >= 0) close(fd);
    }
    for (;;) pause();
}
