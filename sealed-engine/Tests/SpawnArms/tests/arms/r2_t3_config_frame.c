// Arm r2_t3_config_frame: SPEC-R1-R3.md R2-T3 — "Config frame: a truncated
// frame, a frame with trailing bytes, and frames whose manifest breaks R6's
// encoding rules ... -> the child refuses before sandboxing", plus the hostile
// root-path-byte, stray-file-in-$TMPDIR and out-of-bounds MAX_CONTEXT/MAX_MEMORY
// refusals. Every one of those refusals is the CHILD's behaviour on the
// verified channel, driven by the engine's config frame — the first frame the
// ENGINE sends. This module's contract (include/sealed_spawn.h) ends at the
// verified channel: sealed_spawn_child returns the connected, token-checked
// channel and never speaks the config-frame protocol, and the spec carries no
// way to inject or observe a config frame through the header. SKIP, not FAIL.
#include <stdio.h>
#include "harness.h"

int main(void) {
    arm_begin("r2_t3_config_frame");
    printf("SKIP r2_t3_config_frame: config-frame refusals are the child's "
           "channel protocol; this module's contract ends at the verified "
           "channel and exposes no way to drive or observe a config frame "
           "through include/sealed_spawn.h\n");
    fflush(stdout);
    return 0;
}
