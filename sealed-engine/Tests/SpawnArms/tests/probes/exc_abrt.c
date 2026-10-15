// Probe child: raises SIGABRT as its first act (EXC_SOFTWARE/EXC_CRASH leg
// for FIXES-2 4(b)/8). No marker, no rendezvous; the exc_probe receiver owns
// the outcome. Never run on the authoring machine.
#include <signal.h>
#include <unistd.h>

int main(void) {
    raise(SIGABRT);
    _exit(0);
}
