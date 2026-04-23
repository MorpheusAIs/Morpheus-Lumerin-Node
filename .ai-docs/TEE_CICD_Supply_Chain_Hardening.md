# CI/CD Supply-Chain Hardening for Morpheus Docker Images

**Last updated:** 2026-04-22
**First successful run (Phase 1a — signing):** [#22920492249](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22920492249)
**First end-to-end run (Phase 1b — deploy + verify):** [#22969993910](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22969993910)

> **v7.0.0 release status.** The CI/CD hardening described here is the foundation that every downstream trust check depends on. Both **Phase 1c** (consumer-side proxy-router verification of the P-Node) and **Phase 2** (P-Node verifies its own backend LLM) have shipped on top of it — see [`TEE_Attestation_Architecture.md`](TEE_Attestation_Architecture.md) §7.4 and §7.7 for the code-level flow. The CI/CD pipeline itself remains unchanged from Phase 1b in this release; v7 is the *consumer* of the artifacts this pipeline produces.

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

## What This Enables — The Full Loop (as of v7.0.0)

This CI/CD hardening is the **foundation layer** for the full TEE attestation loop. As of v7.0.0, the loop is complete end-to-end — both consumer-side Phase 1 and P-Node-side Phase 2 are shipped:

```
┌──────────────────────────────────────────────────────────────────────┐
│                    CI/CD Pipeline (Phase 1a + 1b — DONE)             │
│                                                                      │
│  Source Code ──► Build ──► Sign ──► Compute RTMR3 ──► Publish GHCR   │
│                    │         │           │               │            │
│                    │     cosign sig    RTMR3 in        ├── image     │
│                    │     (Sigstore)    manifest         ├── SBOM     │
│                    │                                    └── manifest │
│                    ▼                                                 │
│                  Deploy to SecretVM ──► Verify live RTMR3 matches    │
│                  (secretvm-cli)         (polls attestation quote)    │
└──────────────────────────────────────────────────────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────┐
│     Phase 1c — Consumer verifies P-Node (DONE in v6.0.0)             │
│                                                                      │
│  C-Node (v6.0.0+) session open + every prompt:                       │
│    1. IsTeeModel(on-chain tags) == true                              │
│    2. cosign fetch signed manifest from GHCR                         │
│    3. GET provider :29343/cpu → raw TDX quote                        │
│    4. POST to TEE_PORTAL_URL → quote is genuine                      │
│    5. Compare RTMR3 against manifest golden value                    │
│    6. reportData[0:32] == SHA-256(peer TLS cert) → anti-MITM         │
│    7. Cache snapshot; fast-verify on every prompt                    │
│  (attestation/verifier.go; PR #686, #689, #699)                      │
└──────────────────────────────────────────────────────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────┐
│     Phase 2  — P-Node verifies its Backend LLM (DONE in v7.0.0)      │
│                                                                      │
│  P-Node (-tee image, v7.0.0+) startup + every prompt:                │
│    1. GET backend :29343/cpu → backend TDX quote (portal-verified)   │
│    2. TLS pinning via reportData[0:32]                               │
│    3. Artifact-registry lookup for MRTD + RTMR0-2                    │
│    4. Replay RTMR3 from backend docker-compose.yaml (SHA-384 chain)  │
│    5. GET backend :29343/gpu → CPU-GPU binding via reportData[32:64] │
│    6. NVIDIA NRAS v4 attestation of GPU evidence                     │
│    7. Fast-verify on every prompt; PinnedHTTPClient for inference    │
│    8. State exposed at GET /v1/models/attestation                    │
│  (attestation/backend_verifier.go, workload_verifier.go,             │
│   nras_verifier.go, artifacts_registry.go; PR #699, #700, #708-#709) │
└──────────────────────────────────────────────────────────────────────┘
```

**Why v6+ consumers are forward-compatible with v7+ providers:** Phase 2 runs **entirely inside the P-Node** — the consumer never talks to the backend LLM and never sees the backend's attestation quote. The consumer trusts Phase 2 transitively because it has already attested (via Phase 1) that the P-Node is running the exact `-tee` binary that enforces Phase 2. No client-side upgrade is required to get Phase 2 guarantees.

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

### Completed (Phase 1a + 1b — CI/CD)

| Step | Description | Status |
|---|---|---|
| **Cosign signing + SBOM** | Keyless signing, digest capture, SPDX SBOM for TEE image | **Done** (v6.0.0) |
| **TEE attestation manifest** | Signed JSON with digests, hashes, baked env, build provenance | **Done** (v6.0.0) |
| **RTMR3 computation** | Computed in CI/CD from deployed compose + SecretVM rootfs; embedded in signed manifest | **Done** (v6.0.0) |
| **Auto-deploy to SecretVM** | `Deploy-SecretVM-Test` job deploys digest-pinned compose to test VM via `secretvm-cli` | **Done** (v6.0.0) |
| **Post-deploy verification** | Polls live VM attestation, extracts RTMR3 from raw TDX quote, compares against CI-computed value | **Done** (v6.0.0) |
| **ECS deploy timing hardening** | Retry + stabilization-timeout improvements so post-deploy healthchecks don't race ECS | **Done** (PR #694/#695, #701) |

### Completed (Phase 1c — Consumer verifies P-Node, v6.0.0 → v6.2.x)

| Step | Description | Status |
|---|---|---|
| **`IsTeeModel()` helper** | Detect `"tee"` tag on blockchain-registered models; drives both hops of the trust chain | **Done** — PR #708, #709 (consolidated as sole TEE switch) |
| **Consumer-side verification** | Fetch attestation from `:29343`, verify quote via SecretAI portal, compare RTMR3 against signed manifest, pin TLS cert — all before opening session | **Done** (`attestation/verifier.go`) |
| **Per-prompt fast-verify** | Re-fetch quote, compare hash + TLS fingerprint on every forwarded prompt | **Done** — PR #686, #689 |
| **Consumer UI TEE badge** | Visual indicator for TEE-verified models + session status | **Done** |

### Completed (Phase 2a — P-Node verifies its Backend LLM, v7.0.0)

| Step | Description | Status |
|---|---|---|
| **`BackendVerifier.AttestBackend`** | Startup full attestation: portal-verified CPU quote, TLS binding, workload RTMR3 replay, CPU-GPU nonce binding, NRAS | **Done** — PR #699 |
| **`FastVerifyBackend`** | Per-prompt hot-path re-check with hash + TLS fingerprint; no TTL | **Done** — PR #699 |
| **`ArtifactRegistry`** | Auto-refreshed SecretVM TDX artifact CSV for MRTD + RTMR0-2 lookup | **Done** — PR #699 |
| **`NrasVerifier`** | NVIDIA NRAS v4 API integration for GPU attestation | **Done** — PR #699 |
| **`PinnedHTTPClient`** | Onward inference rejects any TLS cert whose SHA-256 differs from attested fingerprint | **Done** — PR #699 |
| **`GET /v1/models/attestation`** | Per-model attestation state endpoint for monitoring and forensics | **Done** — PR #699 |
| **New env vars** | `TEE_PORTAL_URL`, `TEE_IMAGE_REPO`, `ARTIFACT_REGISTRY_URL`, `ARTIFACT_REGISTRY_REFRESH_INTERVAL` | **Done** — PR #699 |

### Remaining (Lower Priority / Future)

| Area | Step | Status |
|---|---|---|
| CI/CD | Full RTMR0-2 *recomputation* in CI (today we verify RTMR0-2 by artifact-registry lookup, which is sufficient) | TODO — blocked on ACPI templates |
| CI/CD | AMD SEV-SNP measurement via `sev-snp-measure` | TODO — TDX-only today |
| CI/CD | CVE scanning (Trivy/Grype) — advisory then gating | TODO |
| Proxy-router | Verifiable per-message signing using SecretVM TEE-bound key | Deferred to Phase 2b |
| Proxy-router | Local in-process quote verification (remove `quote-parse` dependency on SCRT Labs) | Deferred to Phase 2b |
| Proxy-router | Co-located proxy-router + LLM in a single TDX VM (collapses both hops into one RTMR3) | Deferred to Phase 2b |
| Proxy-router | NRAS alternatives for non-NVIDIA GPU vendors | Deferred to Phase 2b |
