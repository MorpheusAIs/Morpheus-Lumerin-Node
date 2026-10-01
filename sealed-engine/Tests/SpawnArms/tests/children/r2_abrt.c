// Child that raises SIGABRT after its marker (EXC_SOFTWARE probe leg for
// r1-handler F4: exceptions other than EXC_CRASH reaching the spawn-time port
// before verification). Expected module outcome when refused: DIED, reaped,
// no crash report — the handler SIGKILLs non-CRASH fatal types before
// replying, so the signal never delivers.
#include <fcntl.h>
#include <stdlib.h>
#include <signal.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc > 1 && argv[1][0] != '\0') {
        int fd = open(argv[1], O_CREAT | O_WRONLY, 0644);
        if (fd >= 0) close(fd);
    }
    raise(SIGABRT);
    _exit(0);   // not reached when SIGABRT is at default disposition
}
