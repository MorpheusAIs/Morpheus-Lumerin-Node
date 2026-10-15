// Well-behaved pinned child (R2 rendezvous shape, per SPEC-R1-R3.md and
// probes/marker.c + impl/smoke/child.c): its first act in main is to write the
// marker file named by argv[1] (the R2-T2 "the child ran" oracle), then it
// derives $TMPDIR/rv, connects, sends one fixed byte and stays connected so
// the engine can read the peer audit token after that byte and finish the
// rendezvous checks. Any failure -> _exit(113). Compiled three times with
// -DVARIANT=1/2/3: each build has its own CD hash; VARIANT 3 is the
// get-task-allow child (signed with the get-task-allow entitlement).
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>

#ifndef VARIANT
#define VARIANT 1
#endif

int main(int argc, char **argv) {
    static volatile int variant = VARIANT;
    if (argc > 1 && argv[1][0] != '\0') {
        int fd = open(argv[1], O_CREAT | O_WRONLY, 0644);
        if (fd >= 0) close(fd);
    }
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
    sleep(120);   // stay connected; the verified arm kills us via the harness
    close(s);
    return variant == VARIANT ? 0 : 1;
}
