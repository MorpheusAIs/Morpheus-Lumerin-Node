#!/usr/bin/env python3
"""
Compute expected RTMR3 (Intel TDX) for a SecretVM deployment.

RTMR3 = replayRTMR(sha256(docker-compose), sha256(rootfs) [, sha256(docker-files)])

Algorithm matches scrtlabs/reproduce-mr  (internal/mr.go  lines 642-657).
Only requires the compose file and rootfs digest — no firmware/kernel/templates needed.

Usage:
    python3 compute-rtmr3.py <docker-compose.yaml> <rootfs.iso> [docker-files] [--json]
    python3 compute-rtmr3.py <docker-compose.yaml> --rootfs-sha256=<hex> [--json]

The --rootfs-sha256 form is for CI when the SecretVM rootfs ISO is not downloadable
(scrtlabs/secret-vm-build GitHub releases were removed) but the SHA-256 is still
pinned in .github/tee/secretvm.env and published in secretvm-verify's tdx.csv
(npm/PyPI data/).

Output (stdout):
    96-char lowercase hex RTMR3 value  (SHA-384, 48 bytes)
"""
import hashlib
import json
import sys


def sha256_file(path: str) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        while True:
            chunk = f.read(1 << 16)
            if not chunk:
                break
            h.update(chunk)
    return h.hexdigest()


def replay_rtmr(entries: list[str]) -> str:
    """Replay the RTMR extension chain.

    Each entry is a hex-encoded SHA-256 hash (64 hex chars = 32 bytes).
    The register starts at 48 zero bytes.  For each entry:
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


def main() -> None:
    json_output = "--json" in sys.argv
    rootfs_sha_arg = None
    for a in sys.argv[1:]:
        if a.startswith("--rootfs-sha256="):
            rootfs_sha_arg = a.split("=", 1)[1].strip().lower().removeprefix("0x")
        elif a == "--rootfs-sha256":
            print("ERROR: use --rootfs-sha256=<hex>", file=sys.stderr)
            sys.exit(2)

    positional = [
        a for a in sys.argv[1:]
        if not a.startswith("--")
    ]

    if not positional:
        print(
            f"Usage: {sys.argv[0]} <docker-compose.yaml> <rootfs.iso> [docker-files] [--json]\n"
            f"       {sys.argv[0]} <docker-compose.yaml> --rootfs-sha256=<hex> [--json]",
            file=sys.stderr,
        )
        sys.exit(1)

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
            print(
                f"Usage: {sys.argv[0]} <docker-compose.yaml> <rootfs.iso> [docker-files] [--json]\n"
                f"       {sys.argv[0]} <docker-compose.yaml> --rootfs-sha256=<hex> [--json]",
                file=sys.stderr,
            )
            sys.exit(1)
        rootfs_hash = sha256_file(positional[1])

    entries = [compose_hash, rootfs_hash]

    docker_files_hash = None
    # docker-files only applies when hashing a real ISO (positional[2])
    if rootfs_sha_arg is None and len(positional) > 2:
        docker_files_hash = sha256_file(positional[2])
        entries.append(docker_files_hash)

    rtmr3 = replay_rtmr(entries)

    if json_output:
        result = {
            "rtmr3": rtmr3,
            "compose_sha256": compose_hash,
            "rootfs_sha256": rootfs_hash,
        }
        if docker_files_hash:
            result["docker_files_sha256"] = docker_files_hash
        json.dump(result, sys.stdout, indent=2)
        print()
    else:
        print(rtmr3)


if __name__ == "__main__":
    main()
