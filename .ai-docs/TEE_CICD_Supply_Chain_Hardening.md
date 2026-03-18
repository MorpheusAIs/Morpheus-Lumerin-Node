# CI/CD Supply-Chain Hardening for Morpheus Docker Images

**Last updated:** 2026-03-11  
**First successful run (Phase 1a — signing):** [#22920492249](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22920492249)  
**First end-to-end run (Phase 1b — deploy + verify):** [#22969993910](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22969993910)

---

## Why This Matters

Morpheus is building toward **verifiable, trustless compute** — where a consumer can cryptographically confirm that a provider is running genuine, untampered software inside a Trusted Execution Environment (TEE) before sending a single prompt.

That trust chain starts at the CI/CD pipeline. If we can't prove that a Docker image was built by our official workflow from a known commit, then nothing downstream — not the TEE hardware attestation, not the RTMR measurements, not the secure enclave — can be meaningfully trusted.

This document describes the supply-chain hardening we've added to the Morpheus-Lumerin-Node build pipeline to close that gap.

---

## What Changed

### Before

The pipeline built Docker images and pushed them to GitHub Container Registry (GHCR). That was it. There was:

- No cryptographic proof of who built the image
- No immutable identifier — only mutable tags like `:v5.14.0`
- No inventory of what's inside the image
- No machine-readable record of what configuration the TEE image was built with

Anyone with GHCR write access could have silently replaced an image. A consumer had no way to distinguish a legitimate image from a compromised one.

### After

Every image build now produces four verifiable artifacts, all attached directly to the image in GHCR:

| Artifact | What It Proves | Attached To |
|---|---|---|
| **Cosign signature** | This image was built by the official GitHub Actions workflow from the MorpheusAIs org | Both standard and TEE images |
| **Image digest** | Immutable `sha256:...` content address — tags can be moved, digests cannot | Both images (exported as job output) |
| **SBOM** | Complete inventory of every binary and Go dependency inside the image (SPDX JSON) | TEE image |
| **TEE attestation manifest** | Signed record of exact image digests, config file hashes, baked environment variables, and build provenance | TEE image |

---

## How It Works

### 1. Cosign Keyless Signing

We use [Sigstore cosign](https://github.com/sigstore/cosign) in **keyless mode**. There is no signing key to manage, rotate, or protect. Instead:

1. GitHub Actions mints an OIDC token during the workflow run
2. Cosign exchanges it for a short-lived certificate from [Fulcio](https://docs.sigstore.dev/fulcio/overview/) (Sigstore's certificate authority)
3. The certificate's identity is bound to the workflow: `https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/.github/workflows/build.yml@refs/heads/{branch}`
4. The signature is recorded in [Rekor](https://docs.sigstore.dev/rekor/overview/) (Sigstore's public transparency log) — immutable and publicly auditable

This means the signature proves three things simultaneously:
- **Who** built it: the MorpheusAIs GitHub organization
- **What** built it: the `build.yml` workflow
- **When** it was built: timestamp from the transparency log

### 2. Image Digest Capture

After each `docker buildx build --push`, we extract the manifest digest from BuildKit's metadata file:

```bash
DIGEST=$(jq -r '.["containerimage.digest"]' /tmp/build-metadata.json)
```

This `sha256:...` digest is:
- Exported as a GitHub Actions job output (available to downstream jobs)
- Used as the target for cosign signing (signatures bind to digests, not tags)
- Included in the TEE attestation manifest

Tags like `:v5.14.0` are human-friendly aliases that can be moved. The digest is the image's true identity.

### 3. SBOM (Software Bill of Materials)

We generate an SBOM for the TEE image using [syft](https://github.com/anchore/syft) in SPDX JSON format:

```bash
syft "$TEE_IMAGE@$DIGEST" -o spdx-json=sbom-tee.spdx.json
cosign attach sbom --sbom sbom-tee.spdx.json "$TEE_IMAGE@$DIGEST"
```

Even though the TEE image is built `FROM scratch` (containing only a single Go binary), syft extracts the Go module build info embedded in the binary. This produces a full dependency inventory — every Go package, every version — that can be audited for known vulnerabilities.

The SBOM is attached to the image in GHCR as an OCI artifact and travels with the image wherever it goes.

### 4. TEE Attestation Manifest

This is the most important new artifact. It's a signed JSON document that records everything needed to verify a TEE deployment:

```json
{
  "tee_image": "ghcr.io/morpheusais/morpheus-lumerin-node-tee:v5.14.7-tee-supply-chain",
  "tee_image_digest": "sha256:3bc2f2f9...",
  "base_image": "ghcr.io/morpheusais/morpheus-lumerin-node:v5.14.7-tee-supply-chain",
  "base_image_digest": "sha256:67dbc859...",
  "compose_file": "proxy-router/docker-compose.tee.yml",
  "compose_sha256": "sha256:9b4b4fce...",
  "dockerfile_tee_sha256": "sha256:30094e96...",
  "build": {
    "commit": "369e9027dc048b52003ca8abd4fbeb278196cba4",
    "ref": "refs/heads/cicd/tee-supply-chain",
    "workflow": "build.yml",
    "run_id": "22920492249",
    "run_url": "https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22920492249",
    "timestamp": "2026-03-10T19:46:38Z"
  },
  "tee_platforms": ["intel-tdx", "amd-sev-snp"],
  "runtime_secrets_only": [
    "WALLET_PRIVATE_KEY",
    "ETH_NODE_ADDRESS",
    "MODELS_CONFIG_CONTENT",
    "WEB_PUBLIC_URL",
    "COOKIE_CONTENT"
  ],
  "baked_env": {
    "network": "mainnet",
    "DIAMOND_CONTRACT_ADDRESS": "0x6aBE1d282f72B474E54527D93b979A4f64d3030a",
    "MOR_TOKEN_ADDRESS": "0x7431ada8a591c955a994a21710752ef9b882b8e3",
    "BLOCKSCOUT_API_URL": "https://base.blockscout.com/api/v2",
    "ETH_NODE_CHAIN_ID": "8453",
    "PROXY_STORE_CHAT_CONTEXT": "false",
    "LOG_COLOR": "false",
    "LOG_JSON": "true",
    "LOG_IS_PROD": "true",
    "ENVIRONMENT": "production"
  },
  "measurements": {
    "intel_tdx": {
      "rtmr3": "<96-char-hex — computed from sha256(compose) + sha256(rootfs)>",
      "secretvm_release": "v0.0.25",
      "rootfs_variant": "rootfs-prod-tdx",
      "rootfs_sha256": "<sha256 of rootfs-prod-v0.0.25-tdx.iso>"
    }
  }
}
```

This manifest is signed with cosign and attached to the image using `cosign attest`. The signature uses the same keyless OIDC flow as the image signature, so verification requires no keys — just trust in the GitHub Actions OIDC issuer and Sigstore's certificate transparency.

**What the manifest tells you:**

- **Image provenance**: Which commit, branch, workflow run, and timestamp produced this image. You can trace back to the exact source code.
- **Image chain**: The TEE image's digest AND the base image's digest. You can verify both independently.
- **Config integrity**: SHA256 hashes of `docker-compose.tee.yml` and `Dockerfile.tee`. If either file was modified between the build and a deployment, the hashes won't match.
- **Baked configuration**: The exact environment variables frozen into the image. A verifier can confirm that `PROXY_STORE_CHAT_CONTEXT=false` (no chat logging) and `ENVIRONMENT=production` are immutable — not overridable at runtime. The `network` field ("mainnet" or "testnet") and corresponding blockchain values (contract addresses, chain ID, blockscout URL) are set at build time based on the branch: `main` → mainnet (Base Mainnet 8453), `test` → testnet (Base Sepolia 84532). All other hardened settings are identical across networks.
- **Runtime secret boundary**: Explicitly lists the 5 variables that ARE injectable at runtime. Everything else is sealed.
- **Platform targets**: Confirms the image is built for both Intel TDX and AMD SEV-SNP TEE platforms.

---

## How to Verify

Anyone with `cosign` installed can verify the entire supply chain. No keys, no accounts, no special access needed.

### Verify the image signature

```bash
cosign verify \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'MorpheusAIs/Morpheus-Lumerin-Node' \
  ghcr.io/morpheusais/morpheus-lumerin-node-tee:<tag>
```

This confirms the image was built by the official MorpheusAIs GitHub Actions workflow. The output includes the exact commit SHA, branch, and workflow that produced it.

### Verify and extract the TEE attestation manifest

```bash
cosign verify-attestation \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'MorpheusAIs/Morpheus-Lumerin-Node' \
  --type https://morpheusais.github.io/tee-attestation/v1 \
  ghcr.io/morpheusais/morpheus-lumerin-node-tee:<tag>
```

This both verifies the signature AND returns the manifest JSON. To extract just the predicate:

```bash
cosign verify-attestation \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'MorpheusAIs/Morpheus-Lumerin-Node' \
  --type https://morpheusais.github.io/tee-attestation/v1 \
  ghcr.io/morpheusais/morpheus-lumerin-node-tee:<tag> \
  2>/dev/null | jq -r '.payload' | base64 -d | jq '.predicate'
```

### View the supply-chain artifact tree

```bash
cosign tree ghcr.io/morpheusais/morpheus-lumerin-node-tee:<tag>
```

This shows all attached artifacts — signature, attestation, and SBOM — in a tree view.

---

## What This Enables — The Full Loop

This CI/CD hardening is the **foundation layer** for the full TEE attestation loop. As of Phase 1b, the pipeline is fully automated end-to-end:

```
┌──────────────────────────────────────────────────────────────────────┐
│                         CI/CD Pipeline (done)                         │
│                                                                      │
│  Source Code ──► Build ──► Sign ──► Compute RTMR3 ──► Publish GHCR   │
│                    │         │           │               │            │
│                    │     cosign sig    RTMR3 in       ├── image      │
│                    │     (Sigstore)    manifest        ├── SBOM      │
│                    │                                   └── manifest  │
│                    ▼                                                  │
│                  Deploy to SecretVM ──► Verify live RTMR3 matches     │
│                  (secretvm-cli)         (polls attestation quote)     │
└──────────────────────────────────────────────────────────────────────┘
                                              │
                                              ▼
                              ┌──────────────────────────────┐
                              │  Consumer Verification       │
                              │  (proxy-router code, next)   │
                              │                              │
                              │  1. Detect "tee" model tag   │
                              │  2. Fetch manifest from GHCR │
                              │  3. Fetch quote from :29343  │
                              │  4. Compare RTMR3            │
                              │  5. If match → session       │
                              │     If fail → reject         │
                              └──────────────────────────────┘
```

**How each artifact feeds the trust chain:**

1. **Image signing** → Consumers can verify a provider is running an official image, not a modified fork
2. **Digest pinning** → The attestation manifest references immutable digests, not mutable tags — so a tag-swap attack is detectable
3. **RTMR3 computation** → The compose hash + rootfs hash produce a predictable RTMR3 that can be compared against live hardware attestation
4. **Auto-deploy + verify** → Every CI/CD push automatically deploys to a test VM and verifies the live RTMR3 matches — catching measurement mismatches before they reach providers
5. **Baked ENV record** → A verifier can confirm that chat logging is disabled and the correct chain contracts are configured — without trusting the provider to self-report
6. **SBOM** → Enables vulnerability scanning and dependency auditing of the exact binary running inside the TEE

---

## Files Changed

| File | Change |
|---|---|
| `.github/workflows/build.yml` | Cosign signing, digest capture, SBOM, attestation manifest, RTMR3 computation, auto-deploy, and post-deploy verification. Also: GitHub Actions upgraded to Node 24-compatible versions, Go version updated to 1.23.x. |
| `.github/tee/secretvm.env` | Pins SecretVM release version, rootfs variant, URL, and SHA-256. All pipeline rootfs references derive from this file. |
| `proxy-router/scripts/compute-rtmr3.py` | Standalone RTMR3 computation script matching the `reproduce-mr` algorithm. Can be run locally for independent verification. |
| `proxy-router/Dockerfile.tee` | Bakes immutable ENV config into the TEE image. Blockchain values (diamond, token, blockscout, chain ID) are parameterized via `ARG` with mainnet defaults; overridden via `--build-arg` for testnet builds. |
| `proxy-router/docker-compose.tee.yml` | Canonical compose template for TEE deployment with 5 runtime secrets |
| `docs/02.3-proxy-router-tee.md` | Provider setup and consumer verification guide |

---

## Current Status and Next Steps

### Completed (Phase 1a + 1b)

| Step | Description | Status |
|---|---|---|
| **Cosign signing + SBOM** | Keyless signing, digest capture, SPDX SBOM for TEE image | **Done** |
| **TEE attestation manifest** | Signed JSON with digests, hashes, baked env, build provenance | **Done** |
| **RTMR3 computation** | Computed in CI/CD from deployed compose + SecretVM rootfs; embedded in signed manifest | **Done** |
| **Auto-deploy to SecretVM** | `Deploy-SecretVM-Test` job deploys digest-pinned compose to test VM via `secretvm-cli` | **Done** |
| **Post-deploy verification** | Polls live VM attestation, extracts RTMR3 from raw TDX quote, compares against CI-computed value | **Done** |

### Remaining (Developer Work — Proxy-Router Code)

| Step | Description | Status |
|---|---|---|
| **`IsTEEModel()` helper** | Detect `"tee"` tag on blockchain-registered models | TODO |
| **Consumer-side verification** | Fetch attestation from `:29343`, verify RTMR3 against signed manifest before opening session | TODO |
| **Consumer UI TEE badge** | Visual indicator for TEE-verified models | TODO |

### Lower Priority (CI/CD)

| Step | Description | Status |
|---|---|---|
| **Full RTMR0-2** | Integrate `reproduce-mr` for firmware/kernel layers (blocked on ACPI templates) | TODO |
| **AMD SEV measurement** | Integrate `sev-snp-measure` for AMD platform | TODO |
| **CVE scanning** | Trivy/Grype scan as advisory step, then gating | TODO |
