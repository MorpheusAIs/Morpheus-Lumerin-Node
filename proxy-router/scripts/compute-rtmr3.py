#!/usr/bin/env python3
"""
Compute expected RTMR3 (Intel TDX) for a SecretVM deployment.

SecretVM measures the workload into RTMR3 with one of two schemes. Which one a
VM uses is decided by its init: images with a dstack/gramine KMS (SecretVM
portal artifacts v0.0.34+) report a non-empty ``dstack_app_id`` on GET /info
and extend RTMR3 with dstack event digests; older images extend the raw
hashes (attest-tool). Mirrors scrtlabs/secretvm-verify 0.13.0 node/src/rtmr.ts
and proxy-router/internal/attestation/rtmr.go (kept in lockstep by
internal/attestation/rtmr3_python_parity_test.go).

attest-tool (no app-id):
    RTMR3 = replay( sha256(docker-compose), sha256(rootfs) [, sha256(docker-files)] )

dstack-util (app-id present):
    RTMR3 = replay( digest("app-id",        app_id_bytes),
                    digest("compose-hash",  sha256(docker-compose)),
                    digest("os-image-hash", sha256(rootfs))
                  [, digest("docker-files-hash", sha256(docker-files))] )
    digest(name, payload) = sha384( LE32(0x08000001) || ":" || name || ":" || payload )

Both: mr starts at 48 zero bytes; for each entry mr = SHA-384(mr || pad48(entry)).

Usage:
    python3 compute-rtmr3.py <docker-compose.yaml> <rootfs.iso> [docker-files] [--json]
    python3 compute-rtmr3.py <docker-compose.yaml> --rootfs-sha256=<hex> [--json]
    python3 compute-rtmr3.py <docker-compose.yaml> --rootfs-sha256=<hex> --dstack-app-id=<hex> [--json]

The --rootfs-sha256 form is for CI when the SecretVM rootfs ISO is not downloadable
(scrtlabs/secret-vm-build GitHub releases were removed) but the SHA-256 is still
pinned in .github/tee/secretvm.env and published in secretvm-verify's tdx.csv.

The --dstack-app-id form reproduces a *specific* dstack VM's RTMR3. The app-id
is per VM (read it from https://<vm>:29343/info), so it is only known after the
VM exists; CI publishes the attest-tool value plus the compose/rootfs digests
and the verifier recomputes the dstack value at attestation time.

Output (stdout):
    96-char lowercase hex RTMR3 value  (SHA-384, 48 bytes)
"""
import hashlib
import json
import struct
import sys

# dstack cc-eventlog event type for RTMR3 app events. Always little-endian on
# the measuring host (x86_64 TDX).
DSTACK_EVENT_TYPE = 0x08000001


def sha256_file(path: str) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        while True:
            chunk = f.read(1 << 16)
            if not chunk:
                break
            h.update(chunk)
    return h.hexdigest()


def normalize_hex(value: str) -> str:
    v = value.strip().lower()
    if v.startswith("sha256:"):
        v = v[len("sha256:"):]
    if v.startswith("0x"):
        v = v[2:]
    return v


def dstack_event_digest(event_name: str, payload: bytes) -> str:
    """sha384(LE32(event_type) || ":" || name || ":" || payload) as hex."""
    h = hashlib.sha384()
    h.update(struct.pack("<I", DSTACK_EVENT_TYPE))
    h.update(b":")
    h.update(event_name.encode("utf-8"))
    h.update(b":")
    h.update(payload)
    return h.hexdigest()


def replay_rtmr(entries: list[str]) -> str:
    """Replay the RTMR extension chain.

    Each entry is a hex-encoded hash. The register starts at 48 zero bytes.
    For each entry:
        content = decode_hex(entry), right-padded to 48 bytes
        mr      = SHA-384(mr || content)
    """
    mr = bytes(48)
    for entry_hex in entries:
        content = bytes.fromhex(entry_hex)
        if len(content) < 48:
            content += bytes(48 - len(content))
        h = hashlib.sha384()
        h.update(mr + content)
        mr = h.digest()
    return mr.hex()


def rtmr3_entries(compose_hash: str, rootfs_hash: str,
                  docker_files_hash: str | None, dstack_app_id: str) -> list[str]:
    """Build the RTMR3 event list for the scheme selected by dstack_app_id."""
    if not dstack_app_id:
        entries = [compose_hash, rootfs_hash]
        if docker_files_hash:
            entries.append(docker_files_hash)
        return entries

    entries = [
        dstack_event_digest("app-id", bytes.fromhex(dstack_app_id)),
        dstack_event_digest("compose-hash", bytes.fromhex(compose_hash)),
        dstack_event_digest("os-image-hash", bytes.fromhex(rootfs_hash)),
    ]
    if docker_files_hash:
        entries.append(dstack_event_digest("docker-files-hash", bytes.fromhex(docker_files_hash)))
    return entries


def usage() -> None:
    print(
        f"Usage: {sys.argv[0]} <docker-compose.yaml> <rootfs.iso> [docker-files] [--json]\n"
        f"       {sys.argv[0]} <docker-compose.yaml> --rootfs-sha256=<hex> [--dstack-app-id=<hex>] [--json]",
        file=sys.stderr,
    )
    sys.exit(1)


def main() -> None:
    json_output = "--json" in sys.argv
    rootfs_sha_arg = None
    dstack_app_id = ""
    for a in sys.argv[1:]:
        if a.startswith("--rootfs-sha256="):
            rootfs_sha_arg = normalize_hex(a.split("=", 1)[1])
        elif a == "--rootfs-sha256":
            print("ERROR: use --rootfs-sha256=<hex>", file=sys.stderr)
            sys.exit(2)
        elif a.startswith("--dstack-app-id="):
            dstack_app_id = normalize_hex(a.split("=", 1)[1])
        elif a == "--dstack-app-id":
            print("ERROR: use --dstack-app-id=<hex>", file=sys.stderr)
            sys.exit(2)

    if dstack_app_id:
        try:
            bytes.fromhex(dstack_app_id)
        except ValueError:
            print("ERROR: --dstack-app-id must be hex", file=sys.stderr)
            sys.exit(2)

    positional = [a for a in sys.argv[1:] if not a.startswith("--")]
    if not positional:
        usage()

    compose_path = positional[0]
    compose_hash = sha256_file(compose_path)

    if rootfs_sha_arg is not None:
        if len(rootfs_sha_arg) != 64 or any(c not in "0123456789abcdef" for c in rootfs_sha_arg):
            print(
                f"ERROR: --rootfs-sha256 must be 64 hex chars, got {len(rootfs_sha_arg)} chars",
                file=sys.stderr,
            )
            sys.exit(2)
        rootfs_hash = rootfs_sha_arg
        if len(positional) > 1:
            print(
                "ERROR: do not pass a rootfs.iso path when using --rootfs-sha256",
                file=sys.stderr,
            )
            sys.exit(2)
    else:
        if len(positional) < 2:
            usage()
        rootfs_hash = sha256_file(positional[1])

    docker_files_hash = None
    # docker-files only applies when hashing a real ISO (positional[2])
    if rootfs_sha_arg is None and len(positional) > 2:
        docker_files_hash = sha256_file(positional[2])

    scheme = "dstack" if dstack_app_id else "attest-tool"
    rtmr3 = replay_rtmr(rtmr3_entries(compose_hash, rootfs_hash, docker_files_hash, dstack_app_id))

    if json_output:
        result = {
            "rtmr3": rtmr3,
            "rtmr3_scheme": scheme,
            "compose_sha256": compose_hash,
            "rootfs_sha256": rootfs_hash,
        }
        if dstack_app_id:
            result["dstack_app_id"] = dstack_app_id
            # The attest-tool value for the same image, for manifests that
            # publish both.
            result["rtmr3_attest_tool"] = replay_rtmr(
                rtmr3_entries(compose_hash, rootfs_hash, docker_files_hash, "")
            )
        if docker_files_hash:
            result["docker_files_sha256"] = docker_files_hash
        json.dump(result, sys.stdout, indent=2)
        print()
    else:
        print(rtmr3)


if __name__ == "__main__":
    main()
