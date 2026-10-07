// Arm r2_t4_bad_requirement: FIXES-2 items 1 and 7 — requirement byte-strings
// that must not end in an accepted, running child. On the unfixed code a
// SPI-side rejection happened before posix_spawn_file_actions_init, so the
// module destroyed an uninitialized file-actions object (UB) and then
// reported ATTR(12); it must return SEALED_SPAWN_REQUIREMENT without crashing.
// The shipping encoding is 82 bytes, so truncated/garbage shapes exercise both
// the SPI-reject path (REQUIREMENT(13), no child exists) and the
// SPI-accepts-then-kernel-refuses path (kernel-stage codes, child never ran).
// Per FIXES-2 7 every variant asserts rc != SEALED_SPAWN_OK and an absent
// marker, reports its code on the result line, and never claims "reaped" on
// the SPI-reject path (no child exists there). Positive control in the same
// program: the pinned child-ok with the shipping requirement bytes is
// accepted (OK) and runs to its marker.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include "harness.h"

static char why[1024];

int main(void) {
    arm_begin("r2_t4_bad_requirement");
    const char *ok = arm_env("CHILD_OK");
    const char *req = arm_env("REQ_OK");
    const char *work = arm_env("WORKDIR");
    if (ok == NULL || req == NULL || work == NULL)
        arm_fail("env missing (CHILD_OK/REQ_OK/WORKDIR)");
    const char *pin = arm_env("HASH_OK");
    if (pin == NULL || strlen(pin) != 2 * SEALED_CDHASH_LEN)
        arm_fail("env HASH_OK missing");

    // Positive control: shipping shape, pinned child accepted.
    char ctrl[512];
    snprintf(ctrl, sizeof ctrl, "%s/m.ctrl-badreq", work);
    unlink(ctrl);
    sealed_child_spec_t spec;
    sealed_child_t st;
    if (arm_build_spec(&spec, ok, pin, req, ctrl, 5000) != 0)
        arm_fail("control spec could not be built");
    if (arm_spawn(&spec, &st) != SEALED_SPAWN_OK || !arm_marker_written(ctrl))
        arm_fail("control failed: pinned child-ok not accepted/never ran");
    sealed_child_kill(&st);

    // Subject: requirement byte-strings that must not end in an accepted,
    // running child (FIXES-2 7). For EVERY variant: rc != SEALED_SPAWN_OK and
    // the child never ran (marker absent) — a malformed requirement that the
    // SPI accepts and the kernel then runs is a seal failure and fails the
    // arm. Every variant's code is reported on the result line.
    static uint8_t ship[4096];
    static uint8_t over[1 << 20];
    size_t ship_len = 0;
    if (arm_load_req(req, ship, sizeof ship, &ship_len) != 0 || ship_len < 8)
        arm_fail("could not load the shipping requirement bytes");
    memset(over, 0xFF, sizeof over);
    static uint8_t v82_ff[82], v82_zero[82];
    memset(v82_ff, 0xFF, sizeof v82_ff);
    memset(v82_zero, 0, sizeof v82_zero);

    struct { const uint8_t *bytes; size_t len; const char *name; } variants[] = {
        {ship, 1, "1-byte truncation"},
        {ship, 4, "4-byte truncation"},
        {v82_ff, sizeof v82_ff, "82 bytes of 0xFF"},
        {v82_zero, sizeof v82_zero, "82 zeroed bytes"},
        {over, sizeof over, "1 MiB of 0xFF (oversize)"},
    };

    char m[512];
    snprintf(m, sizeof m, "%s/m.badreq", work);
    char codes[1024];
    codes[0] = '\0';
    int spi_rejected = 0, accepted_any = 0, ran_any = 0;
    for (size_t v = 0; v < sizeof variants / sizeof variants[0]; v++) {
        memset(&spec, 0, sizeof spec);
        if (arm_hex_pin(pin, spec.cdhash) != 0) arm_fail("bad pin hex");
        spec.path = ok;
        spec.requirement = variants[v].bytes;
        spec.requirement_len = variants[v].len;
        spec.rendezvous_ms = 5000;
        // argv is unused on the refusal paths here (the module returns before
        // the rendezvous); pass a marker path anyway for the harness print.
        static const char *cargv[2];
        cargv[0] = ok;
        cargv[1] = m;
        spec.argv = cargv;
        unlink(m);
        int rc = arm_spawn(&spec, &st);
        char one[96];
        snprintf(one, sizeof one, "%s: %s(%d); ", variants[v].name, arm_code_name(rc), rc);
        strlcat(codes, one, sizeof codes);
        if (rc == SEALED_SPAWN_OK) accepted_any = 1;
        if (rc == SEALED_SPAWN_REQUIREMENT) spi_rejected = 1;
        if (arm_marker_written(m)) ran_any = 1;
        sealed_child_kill(&st);
    }
    if (accepted_any || ran_any) {
        snprintf(why, sizeof why, "SEAL FAILURE: a malformed requirement ended with an "
                 "accepted/running child (accepted=%d ran=%d); %s",
                 accepted_any, ran_any, codes);
        arm_fail(why);
    }
    snprintf(why, sizeof why, "every malformed requirement shape was refused before the "
             "child ran (%s); codes: %s; control accepted",
             spi_rejected ? "REQUIREMENT(13) = the SPI rejected the bytes itself"
                          : "the SPI accepted the bytes; the kernel/verification stage refused",
             codes);
    arm_pass(why);
    return 0;
}
