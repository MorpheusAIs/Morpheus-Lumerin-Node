// Harness helpers for the gating-test arms (TASK-tests.md part 2).
// An arm builds a sealed_child_spec_t for a named child and pin, runs
// sealed_spawn_child, checks the marker file the child writes to an absolute
// path, and reports PASS / FAIL / SKIP with a reason on one line.
#ifndef TEST_HARNESS_H
#define TEST_HARNESS_H

#include <stddef.h>
#include <stdint.h>

#include "sealed_spawn.h"

// 40 hex chars -> 20 CD-hash bytes. 0 on success.
int arm_hex_pin(const char *hex, uint8_t out[SEALED_CDHASH_LEN]);

// Reads a launch-requirement bytes file (a13.swift dumpfile route) into
// buf/len. path may be NULL or "" for no requirement (len = 0). 0 on success.
int arm_load_req(const char *path, uint8_t *buf, size_t max, size_t *len);

// Fills spec: executable `child`, pinned CD hash `pin_hex` (40 hex chars),
// requirement bytes from `req_file` (NULL/"" = none), marker path passed to
// the child as argv[1], rendezvous deadline `rendezvous_ms` (0 -> 5000).
// Storage for the requirement bytes and argv is static in the harness, so one
// spec per arm at a time. 0 on success.
int arm_build_spec(sealed_child_spec_t *spec, const char *child,
                   const char *pin_hex, const char *req_file,
                   const char *marker, unsigned rendezvous_ms);

// Symbolic name of a sealed_spawn result code ("OK", "CDHASH", ...).
const char *arm_code_name(int rc);

// Runs sealed_spawn_child and prints one diagnostic line
// "arm: result=NAME(rc) marker=N suspended_status=0x.. running_status=0x..".
int arm_spawn(const sealed_child_spec_t *spec, sealed_child_t *child);

// 1 iff the marker file exists (the child ran to its first instruction).
int arm_marker_written(const char *marker);

// Call once at the top of main with the arm's name; it prefixes the
// PASS / FAIL / SKIP line ("PASS <name>: <reason>").
void arm_begin(const char *name);

// Report the arm result and exit (0 on PASS/SKIP, 1 on FAIL), exactly one
// line: "PASS <name>: <reason>" / "FAIL <name>: <reason>" /
// "SKIP (stated) <name>: <reason>".
void arm_pass(const char *reason);
void arm_fail(const char *reason);
void arm_skip(const char *reason);

// getenv convenience for the env the Makefile exports (CHILD_OK, HASH_OK, ...).
const char *arm_env(const char *name);

#endif
