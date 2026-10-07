// Child that traps after its marker (EXC_BREAKPOINT leg for r1-handler F4:
// an explicit hardware-trap exception, not a plain bad access). Expected
// module outcome when refused: DIED, reaped, no crash report, no hang — the
// handler replies KERN_SUCCESS without re-running the fault, per the
// measured rule in E11.
#include <fcntl.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc > 1 && argv[1][0] != '\0') {
        int fd = open(argv[1], O_CREAT | O_WRONLY, 0644);
        if (fd >= 0) close(fd);
    }
    __builtin_trap();
    _exit(0);   // not reached
}
