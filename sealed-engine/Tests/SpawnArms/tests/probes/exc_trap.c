// Probe child: explicit hardware trap (__builtin_trap -> EXC_BREAKPOINT leg
// for FIXES-2 4(b)/8). No marker, no rendezvous; the exc_probe receiver owns
// the outcome. Never run on the authoring machine.
#include <unistd.h>

int main(void) {
    __builtin_trap();
    _exit(0);
}
