// Implementation of tests/lib/harness.h: see the header for the contracts.
#include "harness.h"

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static char arm_name[64];

void arm_begin(const char *name) {
    snprintf(arm_name, sizeof arm_name, "%s", name ? name : "?");
}

static void arm_print(const char *tag, const char *reason) {
    if (arm_name[0] == '\0') snprintf(arm_name, sizeof arm_name, "?");
    printf("%s %s: %s\n", tag, arm_name, reason);
    fflush(stdout);
}

void arm_pass(const char *reason) { arm_print("PASS", reason); exit(0); }
void arm_fail(const char *reason) { arm_print("FAIL", reason); exit(1); }
void arm_skip(const char *reason) { arm_print("SKIP (stated)", reason); exit(0); }

const char *arm_env(const char *name) { return getenv(name); }

int arm_hex_pin(const char *hex, uint8_t out[SEALED_CDHASH_LEN]) {
    if (hex == NULL || strlen(hex) != 2 * SEALED_CDHASH_LEN) return 1;
    for (int i = 0; i < SEALED_CDHASH_LEN; i++) {
        unsigned b = 0;
        for (int j = 0; j < 2; j++) {
            char c = hex[2 * i + j];
            unsigned v;
            if (c >= '0' && c <= '9') v = (unsigned)(c - '0');
            else if (c >= 'a' && c <= 'f') v = (unsigned)(c - 'a' + 10);
            else if (c >= 'A' && c <= 'F') v = (unsigned)(c - 'A' + 10);
            else return 1;
            b = b * 16 + v;
        }
        out[i] = (uint8_t)b;
    }
    return 0;
}

int arm_load_req(const char *path, uint8_t *buf, size_t max, size_t *len) {
    *len = 0;
    if (path == NULL || path[0] == '\0') return 0;
    FILE *f = fopen(path, "rb");
    if (f == NULL) return 1;
    *len = fread(buf, 1, max, f);
    fclose(f);
    return *len == 0;
}

int arm_build_spec(sealed_child_spec_t *spec, const char *child,
                   const char *pin_hex, const char *req_file,
                   const char *marker, unsigned rendezvous_ms) {
    static uint8_t req[65536];
    static const char *cargv[3];
    size_t req_len = 0;
    memset(spec, 0, sizeof *spec);
    if (child == NULL || marker == NULL) return 1;
    if (arm_hex_pin(pin_hex, spec->cdhash) != 0) return 1;
    if (arm_load_req(req_file, req, sizeof req, &req_len) != 0) return 1;
    cargv[0] = child;
    cargv[1] = marker;
    cargv[2] = NULL;
    spec->path = child;
    spec->argv = cargv;
    spec->requirement = req_len ? req : NULL;
    spec->requirement_len = req_len;
    spec->rendezvous_ms = rendezvous_ms ? rendezvous_ms : 5000;
    return 0;
}

const char *arm_code_name(int rc) {
    switch (rc) {
    case SEALED_SPAWN_OK: return "OK";
    case SEALED_SPAWN_DIR: return "DIR";
    case SEALED_SPAWN_PORT: return "PORT";
    case SEALED_SPAWN_ATTR: return "ATTR";
    case SEALED_SPAWN_REQUIREMENT: return "REQUIREMENT";
    case SEALED_SPAWN_EXEC: return "EXEC";
    case SEALED_SPAWN_CDHASH: return "CDHASH";
    case SEALED_SPAWN_STATUS: return "STATUS";
    case SEALED_SPAWN_FDS: return "FDS";
    case SEALED_SPAWN_RENDEZVOUS: return "RENDEZVOUS";
    case SEALED_SPAWN_TOKEN: return "TOKEN";
    case SEALED_SPAWN_DIED: return "DIED";
    default: return "?";
    }
}

int arm_spawn(const sealed_child_spec_t *spec, sealed_child_t *child) {
    memset(child, 0, sizeof *child);
    child->channel_fd = -1;
    int rc = sealed_spawn_child(spec, child);
    printf("arm: child=%s result=%s(%d) marker=%d suspended_status=0x%x "
           "running_status=0x%x exception_seen=%d\n",
           spec->path, arm_code_name(rc), rc, arm_marker_written(spec->argv[1]),
           child->status_seen_suspended, child->status_seen_running,
           child->exception_seen);
    fflush(stdout);
    return rc;
}

int arm_marker_written(const char *marker) {
    if (marker == NULL) return 0;
    return access(marker, F_OK) == 0;
}
