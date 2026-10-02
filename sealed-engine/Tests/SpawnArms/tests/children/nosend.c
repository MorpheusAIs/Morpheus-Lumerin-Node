// Child that connects but sends nothing (R2 "first byte" shape): writes the
// marker named by argv[1], connects to $TMPDIR/rv, then stays silent — the
// engine must refuse with SEALED_SPAWN_RENDEZVOUS, since the deadline covers
// both the connection and the byte (KK3).
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <unistd.h>

int main(int argc, char **argv) {
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
    for (;;) pause();
}
