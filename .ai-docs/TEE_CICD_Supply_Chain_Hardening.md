# CI/CD Supply-Chain Hardening for Morpheus Docker Images

**Date:** 2026-03-10  
**Branch:** `cicd/tee-supply-chain`  
**PR target:** `dev`  
**First successful run:** [#22920492249](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22920492249)

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
    "note": "RTMR/SEV values will be populated when reproduce-mr is integrated"
  }
}
```

This manifest is signed with cosign and attached to the image using `cosign attest`. The signature uses the same keyless OIDC flow as the image signature, so verification requires no keys — just trust in the GitHub Actions OIDC issuer and Sigstore's certificate transparency.

**What the manifest tells you:**

- **Image provenance**: Which commit, branch, workflow run, and timestamp produced this image. You can trace back to the exact source code.
- **Image chain**: The TEE image's digest AND the base image's digest. You can verify both independently.
- **Config integrity**: SHA256 hashes of `docker-compose.tee.yml` and `Dockerfile.tee`. If either file was modified between the build and a deployment, the hashes won't match.
- **Baked configuration**: The exact environment variables frozen into the image. A verifier can confirm that `PROXY_STORE_CHAT_CONTEXT=false` (no chat logging) and `ENVIRONMENT=production` are immutable — not overridable at runtime.
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

## What This Enables Next

This CI/CD hardening is the **foundation layer** for the full TEE attestation loop. Here's how each artifact feeds into the bigger picture:

```
┌─────────────────────────────────────────────────────────────┐
│                    CI/CD Pipeline (done)                     │
│                                                             │
│  Source Code ──► Build ──► Sign ──► Publish to GHCR         │
│                    │         │         │                     │
│                    │     cosign sig    ├── image + digest    │
│                    │     (Sigstore)    ├── SBOM              │
│                    │                   └── attestation       │
│                    │                       manifest          │
└────────────────────┼────────────────────────┼───────────────┘
                     │                        │
                     ▼                        ▼
┌─────────────────────────┐    ┌──────────────────────────────┐
│  TEE VM Deployment      │    │  Consumer Verification       │
│  (SecretVM / TDX / SEV) │    │  (proxy-router code, later)  │
│                         │    │                              │
│  Image + compose are    │    │  1. Check provider version   │
│  measured into RTMR3    │    │  2. Fetch attestation from   │
│  at boot time by the    │    │     GHCR (cosign verify)     │
│  hardware TEE           │    │  3. Compare compose hash     │
│                         │    │  4. Query TEE hardware quote │
│  compose_sha256 from    │    │  5. Match RTMR3 measurement  │
│  the manifest lets us   │    │  6. If all pass → session    │
│  predict what RTMR3     │    │     If any fail → reject     │
│  should be              │    │                              │
└─────────────────────────┘    └──────────────────────────────┘
```

**Specifically:**

1. **Image signing** → Consumers can verify a provider is running an official image, not a modified fork
2. **Digest pinning** → The attestation manifest references immutable digests, not mutable tags — so a tag-swap attack is detectable
3. **Compose hash** → When combined with SCRT Labs' `reproduce-mr` tool, the compose hash allows us to precompute the expected RTMR3 measurement for the TEE VM, which is the software-layer check in the hardware attestation
4. **Baked ENV record** → A verifier can confirm that chat logging is disabled and the correct chain contracts are configured — without trusting the provider to self-report
5. **SBOM** → Enables vulnerability scanning and dependency auditing of the exact binary running inside the TEE

---

## Files Changed

| File | Change |
|---|---|
| `.github/workflows/build.yml` | Added cosign signing, digest capture, SBOM, and attestation manifest to both GHCR build jobs |
| `proxy-router/Dockerfile.tee` | Unchanged (created in prior PR) — bakes immutable ENV config into the TEE image |
| `proxy-router/docker-compose.tee.yml` | Unchanged (created in prior PR) — canonical compose for TEE deployment with 5 runtime secrets |

---

## Next Steps

| Step | Description | Status |
|---|---|---|
| **RTMR computation** | Integrate SCRT Labs' `reproduce-mr` into CI/CD to compute expected Intel TDX RTMR3 values and populate the attestation manifest `measurements` field | Next |
| **AMD SEV measurement** | Integrate `sev-snp-measure` for AMD platform measurement computation | Next |
| **Healthcheck extension** | Add TEE metadata (platform, image digest, attestation URL) to the proxy-router `/healthcheck` endpoint | Proxy-router code change |
| **Attestation proxy endpoint** | New `/attestation/quote` endpoint on the proxy-router that relays the hardware attestation quote from the TEE platform | Proxy-router code change |
| **Consumer verification** | Implement the full verification loop in the consumer node's session creation flow | Proxy-router code change |
