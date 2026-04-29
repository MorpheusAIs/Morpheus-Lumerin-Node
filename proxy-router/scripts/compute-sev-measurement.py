#!/usr/bin/env python3
"""
Compute expected AMD SEV-SNP launch-digest `measurement` values for a SecretVM
deployment, one per vCPU template (small/medium/large/2xlarge/4xlarge).

This is the byte-for-byte Python port of
``proxy-router/internal/attestation/sev_gctx.go::CalcSevMeasurement`` (PR #718).
The Go function is the runtime source-of-truth used by the consumer's Phase 1
verifier and the P-Node BackendVerifier; this Python tool exists for the
CI/CD pipeline (so it stays in lockstep with the existing
``compute-rtmr3.py`` pattern for Intel TDX) and so providers/auditors can
recompute the measurements independently.

Mathematical model (per AMD SNP spec Section 8.17.2 Table 67):

    PAGE_INFO    = ld_current(48) || contents(48) || u16_LE(0x70)
                   || page_type(1) || imi/perms/reserved(5) || gpa_LE(8)
    ld_next      = SHA-384(PAGE_INFO)

The launch digest accumulates over the OVMF firmware sections, the kernel
hashes table, then one VMSA page per vCPU. Section type semantics:

    1  SNP_SEC_MEM       page_type 0x03 (zero), content = 48 zero bytes
    2  SNP_SECRETS       page_type 0x05,        content = 48 zero bytes
    3  CPUID             page_type 0x06,        content = 48 zero bytes
    4  SVSM_CAA          page_type 0x03 (zero), content = 48 zero bytes
    16 SNP_KERNEL_HASHES page_type 0x01 (normal), content = SHA-384(page)

VMSA pages use page_type 0x02 at fixed GPA 0xFFFFFFFFF000 and content =
SHA-384(vmsa_page). The BSP runs RIP = 0xfffffff0; APs use the registry's
``sev_es_reset_eip`` value.

Cmdline included in the kernel-hashes table:

    "console=ttyS0 loglevel=7[ <cmdline_extra>] " \\
    "docker_compose_hash=<compose-sha256> rootfs_hash=<rootfs-sha256>"

The ``loglevel=7`` token is the Linux *kernel* log level controlled by SecretVM
(NOT the proxy-router application log level — those are baked into the TEE
image via Dockerfile.tee and recorded in the attestation manifest's
``baked_env`` block).

The canonical ``--artifacts-ver`` value comes from
``.github/tee/secretvm.env`` (variable ``SECRETVM_RELEASE``); CI passes it in
automatically. Run ``--help`` for the full flag set.

Usage (single template — ``$RELEASE`` resolved by the caller, e.g. via
``source .github/tee/secretvm.env`` or ``--artifacts-ver "$(grep ^SECRETVM_RELEASE
.github/tee/secretvm.env | cut -d= -f2)"``):

    python3 compute-sev-measurement.py \\
        --registry sev.json \\
        --vm-type prod \\
        --artifacts-ver "$RELEASE" \\
        --compose docker-compose.tee.yml \\
        --template small

Usage (all 5 templates as JSON, used by the CI pipeline):

    python3 compute-sev-measurement.py \\
        --registry sev.json \\
        --vm-type prod \\
        --artifacts-ver "$RELEASE" \\
        --compose docker-compose.tee.yml \\
        --all-templates --json
"""
from __future__ import annotations

import argparse
import hashlib
import json
import sys
from pathlib import Path
from typing import Any


LD_SIZE = 48                          # SHA-384 digest size
VMSA_GPA = 0xFFFFFFFFF000             # fixed VMSA load address
VCPU_SIG_EPYC = 0x00800F12            # CPUID signature stored in RDX
GUEST_FEATURES = 0x1                  # SEV_FEATURES bitmap
BSP_EIP = 0xFFFFFFF0                  # BSP reset vector

# Maps SecretVM template name -> vCPU count. Matches sev_gctx.go::VcpuMap.
VCPU_MAP: dict[str, int] = {
    "small":   1,
    "medium":  2,
    "large":   4,
    "2xlarge": 8,
    "4xlarge": 16,
}

SEV_HASH_TABLE_HEADER_GUID = "9438d606-4f22-4cc9-b479-a793d411fd21"
SEV_KERNEL_ENTRY_GUID      = "4de79437-abd2-427f-b835-d5b172d2045b"
SEV_INITRD_ENTRY_GUID      = "44baf731-3a2f-4bd7-9af1-41e29169781d"
SEV_CMDLINE_ENTRY_GUID     = "97d02dd8-bd20-4c94-aa78-e7714d36ab2a"


# ---------------------------------------------------------------------------
# GCTX page-update primitive (mirrors sev_gctx.go::gctxUpdate)
# ---------------------------------------------------------------------------

def _sha384(data: bytes) -> bytes:
    return hashlib.sha384(data).digest()


def _gctx_update(ld: bytes, page_type: int, gpa: int, contents: bytes) -> bytes:
    """One GCTX page-update: SHA-384 over a 0x70-byte PAGE_INFO struct."""
    if len(ld) != LD_SIZE:
        raise ValueError(f"ld must be {LD_SIZE} bytes, got {len(ld)}")
    if len(contents) != LD_SIZE:
        raise ValueError(f"contents must be {LD_SIZE} bytes, got {len(contents)}")

    buf = bytearray(0x70)
    buf[0:48] = ld
    buf[48:96] = contents
    buf[96:98] = (0x70).to_bytes(2, "little")  # page_info_len
    buf[98] = page_type & 0xFF
    # buf[99..103] = 0   (imi flag, vmpl perms, reserved) — already zero
    buf[104:112] = gpa.to_bytes(8, "little")
    return _sha384(bytes(buf))


def _gctx_update_normal_pages(ld: bytes, start_gpa: int, data: bytes) -> bytes:
    """page_type 0x01, content = SHA-384(page) — for SNP_KERNEL_HASHES."""
    for offset in range(0, len(data), 4096):
        page = data[offset:offset + 4096]
        ld = _gctx_update(ld, 0x01, start_gpa + offset, _sha384(page))
    return ld


def _gctx_update_zero_pages(ld: bytes, gpa: int, size: int) -> bytes:
    """page_type 0x03, content = 48 zero bytes — for SNP_SEC_MEM, SVSM_CAA."""
    zeros = b"\x00" * LD_SIZE
    for offset in range(0, size, 4096):
        ld = _gctx_update(ld, 0x03, gpa + offset, zeros)
    return ld


def _gctx_update_secrets_page(ld: bytes, gpa: int) -> bytes:
    """page_type 0x05, content = 48 zero bytes — for SNP_SECRETS."""
    return _gctx_update(ld, 0x05, gpa, b"\x00" * LD_SIZE)


def _gctx_update_cpuid_page(ld: bytes, gpa: int) -> bytes:
    """page_type 0x06, content = 48 zero bytes — for CPUID."""
    return _gctx_update(ld, 0x06, gpa, b"\x00" * LD_SIZE)


def _gctx_update_vmsa_page(ld: bytes, vmsa: bytes) -> bytes:
    """page_type 0x02 at fixed VMSA_GPA, content = SHA-384(vmsa_page)."""
    return _gctx_update(ld, 0x02, VMSA_GPA, _sha384(vmsa))


# ---------------------------------------------------------------------------
# Hashes-table page (mirrors sev_gctx.go::buildHashesPage)
# ---------------------------------------------------------------------------

def _uuid_to_le(guid: str) -> bytes:
    """RFC 4122 mixed-endian: groups 1-3 reverse, groups 4-5 stay BE."""
    raw = bytes.fromhex(guid.replace("-", ""))
    le = bytearray(raw)
    le[0], le[1], le[2], le[3] = raw[3], raw[2], raw[1], raw[0]
    le[4], le[5] = raw[5], raw[4]
    le[6], le[7] = raw[7], raw[6]
    return bytes(le)


def _sev_hash_table_entry(guid: str, hash_bytes: bytes) -> bytes:
    """50-byte entry: GUID(16 LE) + length(u16 LE = 50) + hash(32)."""
    if len(hash_bytes) != 32:
        raise ValueError(f"hash must be 32 bytes, got {len(hash_bytes)}")
    entry = bytearray(50)
    entry[0:16] = _uuid_to_le(guid)
    entry[16:18] = (50).to_bytes(2, "little")
    entry[18:50] = hash_bytes
    return bytes(entry)


def _build_hashes_page(
    kernel_hash_hex: str,
    initrd_hash_hex: str,
    cmdline: str,
    offset_in_page: int,
) -> bytes:
    """4096-byte page containing the QEMU SEV kernel-hashes table at ``offset_in_page``."""
    kernel_hash = bytes.fromhex(kernel_hash_hex)
    if initrd_hash_hex:
        initrd_hash = bytes.fromhex(initrd_hash_hex)
    else:
        initrd_hash = hashlib.sha256(b"").digest()

    # cmdline is null-terminated before hashing; empty cmdline hashes a single \0
    cmdline_bytes = (cmdline.encode("utf-8") + b"\x00") if cmdline else b"\x00"
    cmdline_hash = hashlib.sha256(cmdline_bytes).digest()

    # SevHashTable layout: GUID(16) + length(u16) + cmdline(50) + initrd(50) + kernel(50) = 168
    ht = bytearray(168)
    ht[0:16] = _uuid_to_le(SEV_HASH_TABLE_HEADER_GUID)
    ht[16:18] = (168).to_bytes(2, "little")
    ht[18:68]  = _sev_hash_table_entry(SEV_CMDLINE_ENTRY_GUID, cmdline_hash)
    ht[68:118] = _sev_hash_table_entry(SEV_INITRD_ENTRY_GUID,  initrd_hash)
    ht[118:168] = _sev_hash_table_entry(SEV_KERNEL_ENTRY_GUID, kernel_hash)

    # Pad to 16-byte alignment (176 bytes), then place inside a 4096-byte page.
    padded = bytearray(176)
    padded[0:168] = ht

    page = bytearray(4096)
    page[offset_in_page:offset_in_page + len(padded)] = padded
    return bytes(page)


# ---------------------------------------------------------------------------
# VMSA page (mirrors sev_gctx.go::buildVmsaPage)
# ---------------------------------------------------------------------------

def _vmcb_seg(page: bytearray, off: int, sel: int, attr: int, lim: int, base: int) -> None:
    page[off:off + 2]      = (sel & 0xFFFF).to_bytes(2, "little")
    page[off + 2:off + 4]  = (attr & 0xFFFF).to_bytes(2, "little")
    page[off + 4:off + 8]  = (lim & 0xFFFFFFFF).to_bytes(4, "little")
    page[off + 8:off + 16] = (base & 0xFFFFFFFFFFFFFFFF).to_bytes(8, "little")


def _build_vmsa_page(eip: int, vcpu_sig: int, guest_features: int) -> bytes:
    """Mirrors buildVmsaPage(eip, vcpu_sig, guest_features) in sev_gctx.go."""
    page = bytearray(4096)

    cs_base = eip & 0xFFFF0000
    rip     = eip & 0x0000FFFF

    _vmcb_seg(page, 0x000, 0,      0x0093, 0xFFFF, 0)        # es
    _vmcb_seg(page, 0x010, 0xF000, 0x009B, 0xFFFF, cs_base)  # cs
    _vmcb_seg(page, 0x020, 0,      0x0093, 0xFFFF, 0)        # ss
    _vmcb_seg(page, 0x030, 0,      0x0093, 0xFFFF, 0)        # ds
    _vmcb_seg(page, 0x040, 0,      0x0093, 0xFFFF, 0)        # fs
    _vmcb_seg(page, 0x050, 0,      0x0093, 0xFFFF, 0)        # gs
    _vmcb_seg(page, 0x060, 0,      0x0000, 0xFFFF, 0)        # gdtr
    _vmcb_seg(page, 0x070, 0,      0x0082, 0xFFFF, 0)        # ldtr
    _vmcb_seg(page, 0x080, 0,      0x0000, 0xFFFF, 0)        # idtr
    _vmcb_seg(page, 0x090, 0,      0x008B, 0xFFFF, 0)        # tr

    def put_u64(off: int, val: int) -> None:
        page[off:off + 8] = (val & 0xFFFFFFFFFFFFFFFF).to_bytes(8, "little")

    def put_u32(off: int, val: int) -> None:
        page[off:off + 4] = (val & 0xFFFFFFFF).to_bytes(4, "little")

    def put_u16(off: int, val: int) -> None:
        page[off:off + 2] = (val & 0xFFFF).to_bytes(2, "little")

    put_u64(0x0D0, 0x1000)              # efer (SVME)
    put_u64(0x148, 0x40)                # cr4 (MCE)
    put_u64(0x158, 0x10)                # cr0 (PE)
    put_u64(0x160, 0x400)               # dr7
    put_u64(0x168, 0xFFFF0FF0)          # dr6
    put_u64(0x170, 0x2)                 # rflags
    put_u64(0x178, rip)                 # rip
    put_u64(0x268, 0x0007040600070406)  # g_pat
    put_u64(0x310, vcpu_sig)            # rdx (CPUID sig)
    put_u64(0x3B0, guest_features)      # sev_features
    put_u64(0x3E8, 0x1)                 # xcr0
    put_u32(0x408, 0x1F80)              # mxcsr
    put_u16(0x410, 0x037F)              # x87_fcw

    return bytes(page)


# ---------------------------------------------------------------------------
# Top-level measurement (mirrors sev_gctx.go::CalcSevMeasurement)
# ---------------------------------------------------------------------------

def calc_sev_measurement(entry: dict[str, Any], vcpus: int, cmdline: str) -> str:
    """Compute the SEV-SNP launch digest for the given registry entry + vCPU count."""
    ld = bytes.fromhex(entry["ovmf_hash"])
    if len(ld) != LD_SIZE:
        raise ValueError(
            f"ovmf_hash must decode to {LD_SIZE} bytes (96 hex chars), got {len(ld)}"
        )

    sev_hashes_table_gpa = int(entry["sev_hashes_table_gpa"])
    offset_in_page = sev_hashes_table_gpa & 0xFFF
    hashes_page = _build_hashes_page(
        entry["kernel_hash"],
        entry.get("initrd_hash", ""),
        cmdline,
        offset_in_page,
    )

    for sec in entry["ovmf_sections"]:
        gpa = int(sec["gpa"])
        size = int(sec["size"])
        st = int(sec["section_type"])
        if st == 1:        # SNP_SEC_MEM
            ld = _gctx_update_zero_pages(ld, gpa, size)
        elif st == 2:      # SNP_SECRETS
            ld = _gctx_update_secrets_page(ld, gpa)
        elif st == 3:      # CPUID
            ld = _gctx_update_cpuid_page(ld, gpa)
        elif st == 4:      # SVSM_CAA
            ld = _gctx_update_zero_pages(ld, gpa, size)
        elif st == 0x10:   # SNP_KERNEL_HASHES
            ld = _gctx_update_normal_pages(ld, gpa, hashes_page)
        # Unknown section types are silently ignored — Go behaviour.

    ap_eip = int(entry["sev_es_reset_eip"])
    for i in range(vcpus):
        eip = BSP_EIP if i == 0 else ap_eip
        vmsa = _build_vmsa_page(eip, VCPU_SIG_EPYC, GUEST_FEATURES)
        ld = _gctx_update_vmsa_page(ld, vmsa)

    return ld.hex()


# ---------------------------------------------------------------------------
# Helpers + CLI
# ---------------------------------------------------------------------------

def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        while True:
            chunk = f.read(1 << 16)
            if not chunk:
                break
            h.update(chunk)
    return h.hexdigest()


def build_cmdline(compose_sha256: str, rootfs_hash: str, cmdline_extra: str = "") -> str:
    prefix = "console=ttyS0 loglevel=7"
    if cmdline_extra:
        prefix += " " + cmdline_extra
    return f"{prefix} docker_compose_hash={compose_sha256} rootfs_hash={rootfs_hash}"


def find_registry_entry(
    registry: list[dict[str, Any]], vm_type: str, artifacts_ver: str
) -> dict[str, Any]:
    for e in registry:
        if e.get("vm_type") == vm_type and e.get("artifacts_ver") == artifacts_ver:
            return e
    raise SystemExit(
        f"no SEV registry entry found for vm_type={vm_type} artifacts_ver={artifacts_ver}"
    )


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--registry", required=True, type=Path, help="path to SEV artifact registry JSON")
    p.add_argument("--vm-type", required=True, help="registry vm_type (e.g. prod, dev, gpu_prod)")
    p.add_argument("--artifacts-ver", required=True,
                   help="registry artifacts_ver — must match SECRETVM_RELEASE from "
                        ".github/tee/secretvm.env (the SCRT Labs portal-mandated release)")
    p.add_argument("--compose", required=True, type=Path, help="path to docker-compose.tee.yml (deployed/digest-pinned)")
    p.add_argument(
        "--template",
        choices=sorted(VCPU_MAP.keys()),
        help="single template to compute (default: all 5 if --all-templates is set)",
    )
    p.add_argument("--all-templates", action="store_true", help="compute measurement for every template in VCPU_MAP")
    p.add_argument("--json", action="store_true", help="emit a structured JSON object")
    args = p.parse_args()

    if not args.template and not args.all_templates:
        p.error("specify either --template or --all-templates")

    registry = json.loads(args.registry.read_text())
    if not isinstance(registry, list):
        raise SystemExit("registry must be a JSON array")

    entry = find_registry_entry(registry, args.vm_type, args.artifacts_ver)
    compose_sha = sha256_file(args.compose)
    cmdline = build_cmdline(compose_sha, entry["rootfs_hash"], entry.get("cmdline_extra", ""))

    if args.template:
        templates = [args.template]
    else:
        templates = list(VCPU_MAP.keys())

    measurements: dict[str, str] = {}
    for tmpl in templates:
        measurements[tmpl] = calc_sev_measurement(entry, VCPU_MAP[tmpl], cmdline)

    if args.json:
        out = {
            "vm_type":               args.vm_type,
            "artifacts_ver":         args.artifacts_ver,
            "compose_sha256":        compose_sha,
            "kernel_hash":           entry["kernel_hash"],
            "initrd_hash":           entry["initrd_hash"],
            "ovmf_hash":             entry["ovmf_hash"],
            "rootfs_hash":           entry["rootfs_hash"],
            "sev_hashes_table_gpa":  entry["sev_hashes_table_gpa"],
            "sev_es_reset_eip":      entry["sev_es_reset_eip"],
            "kernel_cmdline":        cmdline,
            "vcpu_type":             entry.get("vcpu_type", "EPYC"),
            "per_template":          measurements,
        }
        if entry.get("cmdline_extra"):
            out["cmdline_extra"] = entry["cmdline_extra"]
        json.dump(out, sys.stdout, indent=2)
        print()
    else:
        for tmpl, m in measurements.items():
            print(f"{tmpl}\t{m}")


if __name__ == "__main__":
    main()
