#!/usr/bin/env python3
"""
Compute expected RTMR3 (Intel TDX) for a SecretVM deployment.

RTMR3 = replayRTMR(sha256(docker-compose), sha256(rootfs) [, sha256(docker-files)])

Algorithm matches scrtlabs/reproduce-mr  (internal/mr.go  lines 642-657).
Only requires the compose file and rootfs — no firmware/kernel/templates needed.

Usage:
    python3 compute-rtmr3.py <docker-compose.yaml> <rootfs.iso> [docker-files]

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
    if len(sys.argv) < 3:
        print(
            f"Usage: {sys.argv[0]} <docker-compose.yaml> <rootfs.iso> [docker-files]",
            file=sys.stderr,
        )
        sys.exit(1)

    compose_path = sys.argv[1]
    rootfs_path = sys.argv[2]

    compose_hash = sha256_file(compose_path)
    rootfs_hash = sha256_file(rootfs_path)

    entries = [compose_hash, rootfs_hash]

    if len(sys.argv) > 3:
        docker_files_hash = sha256_file(sys.argv[3])
        entries.append(docker_files_hash)

    rtmr3 = replay_rtmr(entries)

    if "--json" in sys.argv:
        result = {
            "rtmr3": rtmr3,
            "compose_sha256": compose_hash,
            "rootfs_sha256": rootfs_hash,
        }
        if len(sys.argv) > 3 and sys.argv[3] != "--json":
            result["docker_files_sha256"] = sha256_file(sys.argv[3])
        json.dump(result, sys.stdout, indent=2)
        print()
    else:
        print(rtmr3)


if __name__ == "__main__":
    main()
