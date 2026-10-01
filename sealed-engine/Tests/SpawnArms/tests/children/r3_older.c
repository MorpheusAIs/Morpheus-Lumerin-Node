// An "older child build" for R3-T1 (tests/children/r3_older.c): a different
// build of the well-behaved child — same behavior as children/good.c (marker
// argv[1], $TMPDIR/rv rendezvous, one fixed byte, stays connected), separate
// translation unit, so its CD hash differs from every other child's. Built and
// ad-hoc signed by the Makefile like the other children.
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>

static const char older_build_tag[] = "r3-older-build";

int main(int argc, char **argv) {
    (void)older_build_tag;
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
    sleep(120);   // stay connected; a verified arm kills us via the harness
    close(s);
    return 0;
}
