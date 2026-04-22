# TEE Attestation Architecture — Verifiable Provider Compute

**Status:** v2.0 — Full two-hop trust chain shipped. Phase 1 (consumer → P-Node) in v6.0.0+; Phase 2 (P-Node → backend LLM) in v7.0.0+.
**Last updated:** 2026-04-22

> **Shipping summary (as of v7.0.0):**
> - **Phase 1a (CI/CD supply-chain hardening)** — DONE (v6.0.0)
> - **Phase 1b (RTMR3 computation + automated deployment + post-deploy attestation)** — DONE (v6.0.0)
> - **Phase 1c (proxy-router code: consumer verifies P-Node)** — DONE (v6.0.0, refined through v6.2.x)
> - **Phase 2a (P-Node verifies backend LLM: CPU quote, TLS pinning, RTMR3 replay, CPU-GPU binding, NRAS)** — DONE (v7.0.0, PR #699 and follow-ups #700, #703/#704, #705, #708/#709)
> - **Phase 2b (verifiable per-message signing, local quote verification, co-located CPU+GPU TDX VM)** — still future work
>
> The `tee` on-chain model tag is the single switch that turns on **both** hops: consumer-side P-Node verification (Phase 1) and P-Node-side backend verification (Phase 2). A v6.0.0+ consumer paired with a v7.0.0+ provider gets the full chain transparently with no client-side upgrade.

---

## 1. Goal

A consumer node should be able to **cryptographically verify**, before sending a prompt, that the far-side provider is:

1. Running on genuine TEE hardware (Intel TDX or AMD SEV-SNP)
2. Running an unmodified, known-good version of the proxy-router
3. Configured such that PII/chat logging is disabled and cannot be re-enabled at runtime
4. Not subject to MITM between the consumer and the TEE enclave

---

## 2. Scope — What Shipped

### Phase 1 — DONE (v6.0.0)

**CI/CD supply-chain hardening:**
- Cosign keyless image signing (both standard and TEE images) — **DONE**
- Image digest capture and export — **DONE**
- SBOM generation and attachment (syft) — **DONE**
- Signed TEE attestation manifest (Option 5B — stored in GHCR as OCI artifact) — **DONE**
- Intel TDX RTMR3 computed in CI and published in the signed manifest — **DONE**
- Auto-deploy to SecretVM test instance + post-deploy RTMR3 verification gate — **DONE**
- AMD SEV-SNP measurement path — still TODO (blocked on upstream tooling)

**Proxy-router code (consumer verifies P-Node):**
- `IsTeeModel()` helper for on-chain tag detection — **DONE** (`blockchainapi/model_tags.go`)
- Consumer-side attestation verification before session creation (fetches quote from SecretVM host endpoint at `:29343`, verifies via SecretAI portal, compares to manifest) — **DONE** (`attestation/verifier.go`, called from `blockchainapi/service.go` and `proxyapi/proxy_sender.go`)
- Per-prompt `VerifyProviderQuick` fast-path re-check — **DONE** (hash + TLS fingerprint compare)
- Consumer UI: TEE badge and verification status — **DONE**

### Phase 2 — DONE (v7.0.0, PR #699 + #700)

**P-Node verifies its own backend LLM** (the gap previously "accepted" for Phase 1 is now closed without co-locating):

- `BackendVerifier.AttestBackend` at startup — **DONE**
  - Fetch backend CPU quote from `:29343/cpu`; verify via SecretAI portal
  - TLS binding: compare TLS cert SHA-256 with CPU quote `reportData[0:32]`
  - Workload verification: parse TDX quote, look up MRTD + RTMR0-2 in the published SecretVM TDX artifact registry CSV, replay RTMR3 using SHA-384 extend chain over `SHA-256(docker-compose)` + rootfs data
  - GPU attestation: fetch from `:29343/gpu`, verify CPU-GPU binding via `reportData[32:64] == GPU nonce`
  - NVIDIA NRAS v4 API: submit GPU evidence for independent hardware validation (non-fatal if unreachable)
- `BackendVerifier.FastVerifyBackend` on every prompt — **DONE**
  - Always re-fetches CPU quote (~50 ms), compares hash + TLS fingerprint against cached snapshot
  - Mismatch on hash → trigger full re-attestation
  - Mismatch on TLS fingerprint → immediate hard fail (MITM signal)
- Pinned-cert HTTP client for onward inference traffic (`PinnedHTTPClient`) — **DONE**
  - Custom `VerifyPeerCertificate` refuses any TLS cert whose SHA-256 doesn't match the attested fingerprint
- Artifact registry auto-refresh (`ArtifactRegistry`) — **DONE** (configurable via `ARTIFACT_REGISTRY_URL`, `ARTIFACT_REGISTRY_REFRESH_INTERVAL`)
- Health endpoint `GET /v1/models/attestation` — **DONE**

### Still out of scope (future)

- On-chain oracle / DAO governance for golden-measurement updates (cosign keyless via GHA OIDC remains the signer)
- Verifiable per-message signing using a TEE-bound key (§7.6) — slipped to Phase 2b
- Local quote verification in Go (today we still call SCRT Labs `quote-parse`)
- Co-located proxy-router + LLM in a single TDX VM (would collapse both hops into one RTMR3)
- Rating system integration
- CVE scanning gate

### Design Decisions (from review)

| # | Question | Decision |
|---|---|---|
| 1 | Oracle governance | **Automated by CI/CD** — cosign keyless signing via GitHub Actions OIDC. No multi-sig/DAO for now. |
| 2 | SCRT Labs coupling | **Yes, couple** — use their quote-parse API and `reproduce-mr` tool. They control the VM layer; we need their artifacts. See §3. |
| 3 | AMD SEV support | **Both platforms from day one** if feasible; if `reproduce-mr` only handles TDX, compute what we can and extend for SEV in a fast follow. |
| 4 | Model backend trust | **Closed in Phase 2 (v7.0.0).** Instead of co-locating, the P-Node actively verifies the backend LLM over the network: CPU quote + TLS pinning + RTMR3 replay of the backend's `docker-compose.yaml` + CPU-GPU nonce binding + NVIDIA NRAS. See §7.7. Co-location (single TDX VM for proxy-router + LLM) remains a future simplification that would collapse both hops into one measurement chain. |
| 5 | RTMR computation in CI/CD | **Attempt to integrate** `reproduce-mr` using SCRT Labs release artifacts from GitHub. If not available in CI, publish all inputs so it can be run independently. |
| 6 | Attestation freshness | **Version-based, not clock-based.** Support current version and N-2 prior versions. When a new version publishes, the oldest attestation manifest becomes stale. This is a release-cadence policy. |
| 7 | Consumer opt-out | **No opt-out.** If a consumer selects a TEE-tagged model, verification is mandatory. Failure = hard error, no prompt sent. |
| 8 | Rating system | **Ignore for now.** |

---

## 3. SCRT Labs Infrastructure Coupling

We intentionally couple to SCRT Labs' infrastructure because they control the TEE VM platform layer (firmware, kernel, initramfs, rootfs) that produces RTMR0-2. Our software only controls RTMR3.

### What we use from SCRT Labs

| Asset | Where | Purpose |
|---|---|---|
| **Quote parse API** | `POST https://pccs.scrtlabs.com/dcap-tools/quote-parse` | Parse Intel TDX / AMD SEV quotes into human-readable fields. Used by consumers for verification. |
| **Attestation portal** | `https://secretai.scrtlabs.com/attestation` | Web-based quick verification (paste compose + quote). Useful for manual spot-checks. |
| **`reproduce-mr` tool** | [github.com/scrtlabs/reproduce-mr](https://github.com/scrtlabs/reproduce-mr) | Compute expected RTMR values (Intel TDX) from VM artifacts + docker-compose. Run in CI/CD. |
| **`sev-snp-measure` tool** | [github.com/virtee/sev-snp-measure](https://github.com/virtee/sev-snp-measure) | Compute expected AMD SEV measurement. Run in CI/CD. |
| **SecretVM build artifacts** | [github.com/scrtlabs/secret-vm-build/releases](https://github.com/scrtlabs/secret-vm-build/releases) | Firmware (ovmf.fd), kernel (bzImage), initramfs, rootfs — needed by `reproduce-mr`. |
| **Host attestation service** | `https://<vm-domain>:29343/cpu.html` | Returns the live attestation quote from a running SecretVM instance. Queried directly by the consumer — served by the TEE host, outside the application container. |
| **Verifiable signing service** | `http://172.17.0.1:49153/sign` (internal only) | Signs messages with a TEE-bound key whose public key is embedded in attestation `reportdata`. Phase 2 feature. |
| **SecretVM CLI** | `npm install --global secretvm-cli` / [CLI docs](https://docs.scrt.network/secret-network-documentation/secretvm-confidential-virtual-machines/secretvm-cli) | CLI tool for provisioning, managing, and monitoring SecretVM instances. Alternative to the web portal. |
| **SecretAI dev portal** | `https://secretai.scrtlabs.com/secret-vms/create` | Web UI for creating and managing SecretVM instances. |

### SecretVM deployment methods

Two ways to deploy a TEE VM on SecretVM (contributed by SCRT Labs via [PR #652](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/652)):

**Web portal:** Navigate to https://secretai.scrtlabs.com/secret-vms/create, paste compose + secrets, select hardware (TDX/SEV), deploy.

**CLI (`secretvm-cli`):**
```bash
# Install
sudo npm install --global secretvm-cli

# Authenticate
secretvm-cli auth login

# Create VM with compose + secrets
secretvm-cli -k <api_key> vm create \
  --name <vm_name> \
  --type small \
  --persistence \
  --upgradeability \
  --docker-compose proxy-router/docker-compose.tee.yml \
  --env <path_to_env_file_with_5_secrets> \
  --platform tdx   # or sev

# Check status
secretvm-cli -k <api_key> vm list
```

The CLI is useful for automation and scripting — could be integrated into deployment pipelines or used for programmatic VM lifecycle management.

### Why this coupling is acceptable

- SCRT Labs owns the VM layer — there is no way around needing their artifacts for RTMR0-2 computation
- Their quote-parse API is a convenience, not a trust dependency — the cryptographic verification (Intel/AMD signature chain) is independently verifiable
- If SCRT Labs' API goes down, consumers can still verify using Intel's DCAP libraries directly
- The `reproduce-mr` tool is open-source and can be forked/self-hosted

### RTMR Measurement Layers (Intel TDX)

| Register | What It Measures | Who Controls It |
|---|---|---|
| **MRTD** | Firmware hash | SCRT Labs (build-time, hardware root of trust) |
| **RTMR0** | Firmware configuration (CFV, TDHOB, ACPI) | SCRT Labs |
| **RTMR1** | Linux kernel (bzImage) | SCRT Labs |
| **RTMR2** | Kernel cmdline + initramfs | SCRT Labs |
| **RTMR3** | **Root filesystem + docker-compose.yaml** | **Us (image + compose)** |

### AMD SEV-SNP Difference

AMD SEV-SNP produces a **single cumulative `measurement`** hash over the entire initial guest state. Both platforms are x86_64 — the same Docker image works on both, but the measurement format and computation tool differ.

---

## 4. Attestation Loop (Full Flow)

```
Consumer                          GHCR (signed manifest)            Provider (TEE VM)
   │                                  │                                  │
   │  1. Browse models (tag: "tee")   │                                  │
   │──────────────────────────────────>│ (blockchain)                     │
   │                                  │                                  │
   │  2. Fetch signed attestation     │                                  │
   │     manifest from GHCR (NOT      │                                  │
   │     from the provider)           │                                  │
   │──────────────────────────────────>│ (cosign verify-attestation)     │
   │  { tee_image: ...@sha256:DIGEST, │                                  │
   │    compose_sha256, rtmr3,        │                                  │
   │    baked_env, rootfs_sha256 }    │                                  │
   │<──────────────────────────────────│                                  │
   │                                  │                                  │
   │  3. GET :29343/cpu.html          │                                  │
   │─────────────────────────────────────────────────────────────────────>│
   │  (raw TDX quote from TEE host)   │  (served by VM host, NOT the    │
   │<─────────────────────────────────────── proxy-router container)     │
   │                                  │                                  │
   │  4. Verify (consumer-side):      │                                  │
   │     a) Cosign sig valid (OIDC)   │                                  │
   │     b) Quote sig valid (HW)      │                                  │
   │     c) RTMR3 == manifest.rtmr3   │                                  │
   │     d) baked_env checks          │                                  │
   │                                  │                                  │
   │  5. If OK → InitiateSession      │                                  │
   │─────────────────────────────────────────────────────────────────────>│
   │     If FAIL → hard error         │  (port 3333: inference traffic)  │
```

### Why the RTMR3 is NOT self-attested

The provider never reports its own expected RTMR3. The trust chain is:

1. **Expected RTMR3** comes from the **GHCR attestation manifest** — signed by CI/CD's OIDC key (Sigstore), cryptographically verifiable, P-Node cannot modify it
2. **Actual RTMR3** comes from **Intel/AMD hardware attestation** — signed by the CPU's attestation key, independently verifiable against Intel/AMD root certificates
3. **Consumer** compares (1) vs (2). The P-Node is merely a conduit for the hardware quote; it cannot fake the CPU's cryptographic signature

### Digest-pinned compose eliminates mutability

The compose file used for RTMR3 computation references the image by **immutable digest** (`image: ...@sha256:DIGEST`), not by mutable tag. This means:

- A tag can be overwritten to point to a different image; a digest cannot
- The compose content is deterministic once the image is built
- RTMR3 = f(compose_bytes, rootfs_bytes) — both are immutable for a given version
- Any change to the image produces a different digest → different compose → different RTMR3
- **Critical: exact byte content.** RTMR3 is computed from the raw bytes of the compose file. The compose must end with a single newline (`\n`) after `proxy_data: null` — standard POSIX line ending, **no trailing blank line**. An earlier assumption that SecretVM appends a trailing newline was incorrect; live testing confirmed that the file on the VM's disk matches what the user pastes (single `\n`). The `docker-compose.tee.yml` template is 19 lines ending with `null\n`. Verified via the VM's `/docker-compose` endpoint (2026-03-11).
- **Portal normalization.** The SecretVM portal strips trailing blank lines — even if you paste extra blank lines at the end, the portal normalizes to a single trailing `\n`. This is safe; our template already matches this behavior.
- **Advanced settings that affect RTMR3.** Only two portal settings change RTMR3: (1) **Additional Files** — uploading a `.tar` archive adds a third `sha256(dockerFiles)` entry to the RTMR3 chain; must be left empty for our use case. (2) **Platform** — Intel TDX vs AMD SEV uses different rootfs ISOs. All other advanced settings (persistence, upgrades, hide runtime info, zkVerify, app cert) do not affect RTMR3.

### Chicken-and-egg: resolved

The concern was: "the compose needs the image digest, but the image needs RTMR3 baked in." The resolution: **RTMR3 is NOT baked into the image — it goes into the signed attestation manifest in GHCR.** The flow is strictly sequential:

```
Build image → capture digest → generate compose with digest → compute RTMR3 → sign into manifest
```

No circular dependency.

---

## 5. Phase 1a — CI/CD Supply-Chain Hardening — DONE

**Completed 2026-03-10.** All changes merged to `dev` and verified on the `test` branch build.

### What was delivered

| Artifact | Description | Applied To |
|---|---|---|
| **Cosign keyless signature** | OIDC-based signing via GitHub Actions + Sigstore Fulcio/Rekor | Both standard and TEE images |
| **Image digest** | Immutable `sha256:` manifest digest captured via `--metadata-file` and exported as job output | Both images |
| **SBOM** | SPDX JSON generated by syft, attached via `cosign attach sbom` | TEE image |
| **TEE attestation manifest** | Signed JSON predicate with image digests, compose/Dockerfile hashes, baked ENV, build provenance — attached via `cosign attest` | TEE image |

### Pipeline changes

- Added `permissions: { contents: read, packages: write, id-token: write }` to both GHCR jobs
- Added `sigstore/cosign-installer@v4` step to both GHCR jobs
- Added `anchore/sbom-action/download-syft@v0` step to TEE job
- Added `--metadata-file` to `docker buildx build` for reliable digest extraction
- New steps: `Capture image digest`, `Sign image with cosign (keyless)`, `Generate and attach SBOM`, `Generate TEE attestation manifest`, `Sign and attach TEE attestation manifest`

### Verified output

First successful run: [#22920492249](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22920492249)

```bash
# Signature verification — confirmed
cosign verify \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'MorpheusAIs/Morpheus-Lumerin-Node' \
  ghcr.io/morpheusais/morpheus-lumerin-node-tee:v5.14.7-tee-supply-chain

# Attestation manifest — confirmed, shows all baked ENV, digests, hashes
cosign verify-attestation \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'MorpheusAIs/Morpheus-Lumerin-Node' \
  --type https://morpheusais.github.io/tee-attestation/v1 \
  ghcr.io/morpheusais/morpheus-lumerin-node-tee:v5.14.7-tee-supply-chain

# Artifact tree — confirmed: sig + att + sbom all attached
cosign tree ghcr.io/morpheusais/morpheus-lumerin-node-tee:v5.14.7-tee-supply-chain
```

### PRs and artifacts

| Item | Link |
|---|---|
| PR #644 — TEE image + BASE migration | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/644 |
| PR #646 — Supply-chain hardening (cosign, SBOM, attestation) | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/646 |
| PR #648 — Compose canonical format fix | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/648 |
| PR #650 — User-facing TEE setup & verification docs | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/650 |
| Supply-chain hardening deep-dive doc | `.ai-docs/TEE_CICD_Supply_Chain_Hardening.md` in Morpheus-Lumerin-Node |
| Provider setup & consumer verification guide | `docs/02.3-proxy-router-tee.md` in Morpheus-Lumerin-Node |

### Attestation manifest example (real output from verified run)

```json
{
  "tee_image": "ghcr.io/morpheusais/morpheus-lumerin-node-tee:v5.14.7-tee-supply-chain",
  "tee_image_digest": "sha256:3bc2f2f90308b8fd4bd0d9c03648962b5a40e856abc65b0fa3f689960d2c5899",
  "base_image": "ghcr.io/morpheusais/morpheus-lumerin-node:v5.14.7-tee-supply-chain",
  "base_image_digest": "sha256:67dbc8595e0b5acff3ace058e4b18acebaedbd0d65eb06e8c3f62129ed6e47da",
  "compose_sha256": "sha256:9b4b4fce6ef862999f10410ce9be9b289c28c1eb39d0cdbddd268285f5781987",
  "dockerfile_tee_sha256": "sha256:30094e96fff43dcf335dc6bcadc4265327230196a60c6120495fba1e17acbb6d",
  "build": {
    "commit": "369e9027dc048b52003ca8abd4fbeb278196cba4",
    "ref": "refs/heads/cicd/tee-supply-chain",
    "run_url": "https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22920492249",
    "timestamp": "2026-03-10T19:46:38Z"
  },
  "baked_env": {
    "network": "mainnet",
    "DIAMOND_CONTRACT_ADDRESS": "0x6aBE1d282f72B474E54527D93b979A4f64d3030a",
    "MOR_TOKEN_ADDRESS": "0x7431ada8a591c955a994a21710752ef9b882b8e3",
    "BLOCKSCOUT_API_URL": "https://base.blockscout.com/api/v2",
    "ETH_NODE_CHAIN_ID": "8453",
    "PROXY_STORE_CHAT_CONTEXT": "false",
    "ENVIRONMENT": "production",
    "LOG_JSON": "true",
    "LOG_IS_PROD": "true"
  },
  "runtime_secrets_only": [
    "WALLET_PRIVATE_KEY", "ETH_NODE_ADDRESS", "MODELS_CONFIG_CONTENT",
    "WEB_PUBLIC_URL", "COOKIE_CONTENT"
  ],
  "measurements": {
    "note": "RTMR/SEV values will be populated when reproduce-mr is integrated (Phase 1b)"
  }
}
```

### Version freshness policy (N-2)

Attestation manifests are published per version. The policy is:
- **Current version** + **2 prior versions** are considered valid
- When a new version is published, the oldest of the 3 becomes stale
- This is enforced at the consumer level (proxy-router code, later) — the CI/CD just publishes

This is a **release cadence** concern, not a clock-based TTL. A version doesn't expire on a timer — it expires when 3 newer versions have been published.

---

## 6. Phase 1b — RTMR3 Computation & Automated Deployment — DONE

### 6.1 RTMR3 computation for Intel TDX — DONE

**Goal:** Compute expected RTMR3 in CI/CD using the deployed compose (with immutable digest) and SCRT Labs rootfs. Publish the value in the signed attestation manifest.

**Implementation (branch `cicd/scrt-generate-rtmr3`):**

1. **Standalone computation script:** `proxy-router/scripts/compute-rtmr3.py`
   - Matches the `replayRTMR` algorithm from [scrtlabs/reproduce-mr](https://github.com/scrtlabs/reproduce-mr) (internal/mr.go lines 642-657)
   - RTMR3 = replayRTMR(sha256(compose), sha256(rootfs))
   - Each extension step: content padded to 48 bytes, mr = SHA-384(mr || padded_content)
   - Can be run locally by providers/consumers for independent verification

2. **SecretVM artifact config:** `.github/tee/secretvm.env`
   - Pins SecretVM release version (`v0.0.25`) and rootfs variant (`rootfs-prod-tdx`)
   - Pins rootfs URL and expected SHA-256 hash
   - All pipeline references to rootfs filenames are derived from `SECRETVM_ROOTFS_VARIANT` — no hardcoded filenames in `build.yml`
   - Separate entries for TDX / SEV / GPU variants (SEV/GPU commented out for now)
   - **Important:** Always use the `prod` rootfs variant — SecretVM runs "environment prod" even for developer portal deployments

3. **Compose generation with immutable digest reference:**
   - The repo's `docker-compose.tee.yml` is a **template** (tag-based, for reference)
   - CI/CD generates the **deployed** version with `image: ...@sha256:DIGEST` (immutable)
   - `sed` replaces the image line; all other content is preserved byte-for-byte
   - The deployed compose is uploaded as a build artifact for providers to download

4. **Pipeline flow (in `GHCR-Build-and-Push-TEE` job):**
   ```
   Load network config (main.env or test.env based on branch)
     → Build TEE image with network-specific build args → capture digest → sign image
     → load secretvm.env → download rootfs (cached)
     → generate compose with @sha256:DIGEST
     → compute RTMR3 from compose + rootfs
     → generate attestation manifest (with RTMR3 + network-specific baked_env)
     → sign and attach manifest → upload deployed compose
   ```

5. **Enhanced attestation manifest structure:**
   ```json
   {
     "tee_image": "ghcr.io/.../morpheus-lumerin-node-tee@sha256:DIGEST",
     "tee_image_tag": "ghcr.io/.../morpheus-lumerin-node-tee:vX.Y.Z-branch",
     "compose_sha256": "sha256:...",
     "compose_image_reference": "ghcr.io/.../morpheus-lumerin-node-tee@sha256:DIGEST",
     "baked_env": {
       "network": "mainnet | testnet",
       "DIAMOND_CONTRACT_ADDRESS": "...",
       "MOR_TOKEN_ADDRESS": "...",
       "BLOCKSCOUT_API_URL": "...",
       "ETH_NODE_CHAIN_ID": "8453 | 84532",
       "PROXY_STORE_CHAT_CONTEXT": "false",
       "ENVIRONMENT": "production"
     },
     "measurements": {
       "intel_tdx": {
         "rtmr3": "96-char-hex-value",
         "secretvm_release": "v0.0.25",
         "rootfs_variant": "rootfs-prod-tdx",
         "rootfs_sha256": "<sha256-of-rootfs-prod-v0.0.25-tdx.iso>"
       }
     }
   }
   ```

**Why RTMR3 only (not RTMR0-3):**
- RTMR3 measures rootfs + docker-compose.yaml — the only registers **our software** controls
- RTMR0-2 measure firmware, kernel, initramfs, ACPI tables — all controlled by SCRT Labs
- Computing RTMR0-2 requires `reproduce-mr` with ACPI templates (not yet available) and exact VM config (memory, CPU, cmdline)
- RTMR3 can be computed with standard tools (SHA-256 + SHA-384) — no external dependencies

### 6.2 Integrate full `reproduce-mr` for RTMR0-2 (follow-up)

**Goal:** Compute all RTMR values (0-3) for complete verification.

**What's needed:**
- ACPI table templates (`template_qemu_cpu{N}.hex`) — not currently available in the `reproduce-mr` repo; need to obtain from SCRT Labs or generate from QEMU build
- Exact kernel cmdline used by SecretVM (from docs: `console=ttyS0 loglevel=7 clearcpuid=mtrr,rtmr ro initrd=initrd`)
- VM configuration: memory size, CPU count (maps to SecretVM "small"/"medium"/"large" types)
- TCB version parameter (6 or 7, determines MRTD computation variant)

**Approach:** `go install github.com/scrtlabs/reproduce-mr@latest` in CI/CD, download full artifact set (firmware, kernel, initrd, rootfs, templates), run with correct parameters.

**Effort:** M  
**Blocked by:** Obtaining ACPI templates from SCRT Labs

### 6.3 Integrate `sev-snp-measure` for AMD SEV

**Goal:** Compute expected AMD SEV-SNP `measurement` hash.

**Approach:** Same rootfs, different computation tool: [virtee/sev-snp-measure](https://github.com/virtee/sev-snp-measure) (Python).

**Effort:** M  
**Blocked by:** Same template/config dependency as 6.2

### 6.4 SecretVM release version tracking — DONE

**Implementation:** `.github/tee/secretvm.env` pins the release version and rootfs variant. The pipeline derives all rootfs filenames, cache keys, and download paths from `SECRETVM_RELEASE` and `SECRETVM_ROOTFS_VARIANT` — no hardcoded filenames in `build.yml`. The attestation manifest includes `measurements.intel_tdx.secretvm_release`, `rootfs_variant`, and `rootfs_sha256`.

**Upgrade procedure:** When SCRT Labs publishes a new release, edit `secretvm.env` (version, URL, clear SHA256), push, let CI capture the new rootfs hash from the step summary, then pin it back. Full instructions in `docs/02.3-proxy-router-tee.md` § "Upgrading SecretVM Artifacts".

**Version detection:** There is no SCRT Labs push notification or dedicated API for new releases. The GitHub Releases API (`https://api.github.com/repos/scrtlabs/secret-vm-build/releases`) can be polled, but newer versions are often marked as **pre-release** on GitHub while already deployed to the SecretVM portal (e.g., `v0.0.25` is a pre-release on GitHub but portal runs "artifacts v0.0.25, environment prod"). The `/releases/latest` endpoint only returns non-prerelease versions and should NOT be relied on. Future consideration: a scheduled GitHub Action that polls for new releases and opens an issue or PR.

### 6.5 Auto-deploy to SecretVM test instance — DONE

**Goal:** Close the automation loop: push code → build TEE image → compute RTMR3 → sign → auto-update running test VM on SecretVM.

**Implementation:** A dedicated `Deploy-SecretVM-Test` job runs after `GHCR-Build-and-Push-TEE`, scoped to the `test` branch and bound to the GitHub Actions **`test` environment**. This ensures the three required secrets are environment-scoped (not repo-wide), allowing separate `production` secrets later for `main` branch deployments.

**Network-aware builds:** The TEE image built for the `test` branch uses **testnet** values (Base Sepolia — chain ID 84532, testnet contracts) sourced from `.github/workflows/test.env`. The `main` branch uses **mainnet** values (Base Mainnet — chain ID 8453) from `.github/workflows/main.env`. The `Dockerfile.tee` blockchain values are parameterized via `ARG` with mainnet defaults, overridden via `--build-arg` at build time. All other hardened settings (logging, timeouts, proxy config) are identical. The attestation manifest `baked_env` section reflects the actual network used, including a `network` field ("mainnet" or "testnet").

The job downloads the deployed compose artifact from the build job, then calls `secretvm-cli vm edit` to update the running test VM.

**GitHub Environment Secrets required** (set in Settings → Environments → `test`):

| Secret | Value |
|---|---|
| `SECRETVM_API_KEY` | SCRT Labs API key (from portal account settings) |
| `SECRETVM_TEST_VM_UUID` | UUID of the test VM to update |
| `SECRETVM_TEST_ENV` | The 5 runtime secrets as a `.env` file, base64-encoded |

**Encoding the env file:**
```bash
# Create env file with the 5 runtime secrets
cat > secrets.env <<EOF
WALLET_PRIVATE_KEY=0x...
ETH_NODE_ADDRESS=wss://base-mainnet.g.alchemy.com/v2/YOUR_KEY
MODELS_CONFIG_CONTENT={"models":[...]}
WEB_PUBLIC_URL=https://your-domain:8082
COOKIE_CONTENT=admin:yourpassword
EOF

# Base64-encode and copy to clipboard
base64 -i secrets.env | pbcopy
# Paste into GitHub Secret SECRETVM_TEST_ENV
rm secrets.env
```

**Pipeline flow:**
```
GHCR-Build-and-Push-TEE (builds, signs, uploads compose artifact, exports rtmr3)
  ↓
Deploy-SecretVM-Test (separate job, environment: test, test + cicd/* branches)
  → download compose artifact
  → install secretvm-cli
  → decode SECRETVM_TEST_ENV from base64 to /tmp/secretvm-test.env
  → secretvm-cli vm edit (compose + env)
  → cleanup env file
  → VERIFY: poll vm attestation → extract RTMR3 from raw TDX quote → compare against CI-computed RTMR3
  → log match/mismatch to step summary
```

**Attestation verification detail:**
The `vm attestation` CLI command returns the raw TDX Quote v4 as a hex blob (not parsed JSON). RTMR3 is extracted from the known offset in the TDX quote structure: 48-byte header + 472 bytes into the TD Report Body = byte offset 520, 48 bytes long (96 hex chars). The step polls up to 12 times at 30-second intervals (6 minutes) to allow the VM to reboot with the new image. If RTMR3 doesn't match after all attempts, the job fails.

**Environment branch policy:**
The `test` environment uses custom deployment branch policies. The allowed branches must include both `test` and `cicd/*` patterns (configured in Settings → Environments → test → Deployment branches).

**First successful end-to-end run:** https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22969993910

**Future: production deployment**
Add a `Deploy-SecretVM-Prod` job bound to `environment: production`, triggered only on `main` branch, with its own set of secrets (`SECRETVM_API_KEY`, `SECRETVM_PROD_VM_UUID`, `SECRETVM_PROD_ENV`). Environment protection rules (reviewers, wait timer) can gate production deploys.

**Security notes:**
- The env file is decoded to `/tmp` and deleted immediately after `vm edit`
- The `vm edit` command re-sends all env vars (old ones are not preserved)
- The API key and secrets are never printed to logs (GitHub masks `${{ secrets.* }}`)
- The raw TDX quote is parsed in-pipeline but never stored; only the 96-char RTMR3 hex is logged

### 6.6 CVE scanning gate (optional, lower priority)

**Goal:** Scan the TEE image for known vulnerabilities before publishing.

**Approach:** Add [Trivy](https://github.com/aquasecurity/trivy) or [Grype](https://github.com/anchore/grype) to the TEE job as a non-blocking advisory step initially.

**Effort:** S

---

## 7. Next Steps — Proxy-Router Code (Developer Work)

These are changes to the proxy-router Go code and UI. They consume the CI/CD artifacts to complete the attestation loop.

### ~~7.1 Extend `/healthcheck` with TEE metadata~~ — DROPPED

**Reason:** Not required. The consumer discovers TEE capability from the on-chain `"tee"` model tag (see §7.3). The image digest is available in the cosign-signed attestation manifest on GHCR, and the platform can be inferred from the attestation quote. Adding TEE fields to `/healthcheck` would be nice-to-have but isn't on the critical path.

### ~~7.2 Add `/attestation/quote` proxy endpoint~~ — DROPPED

**Reason:** Not required. The SecretVM host already exposes the attestation quote at `https://{vm-domain}:29343/cpu.html` (raw TDX quote hex). The consumer can fetch it directly from the provider's VM domain — no need for the proxy-router to re-serve it. The VM domain is discoverable from the provider's public URL.

Example: `https://<your-vm-name>.vm.scrtlabs.com:29343/cpu.html`

The `secretvm-cli vm attestation <uuid>` command also returns the same data (used in our CI/CD verification step).

**Port security note:** This design has a beneficial side-effect for providers. The management API on port 8082 (which includes `/healthcheck`, wallet config, session management) can be firewalled off — it's an admin surface that providers may not want publicly exposed. The consumer verification flow only requires:
- **Port 3333** — MOR proxy protocol (inference traffic)
- **Port 29343** — SecretVM attestation endpoint (platform-provided, outside the container)
- **GHCR** — cosign-signed attestation manifest (public, no auth)

The attestation endpoint at `:29343` is served by the TEE host, not the proxy-router container. Even a compromised application cannot fake or suppress it — strong separation of concerns.

### 7.3 `IsTeeModel()` helper and model tag convention — DONE

**File:** `proxy-router/internal/blockchainapi/model_tags.go`

**What shipped:** Helper function that detects the `tee` tag on an on-chain model's `tags` array. The `ModelRegistry` contract already has `string[] tags` — providers registering TEE models add `"tee"` as a tag. This single tag drives **both** hops of the trust chain: the consumer uses it to decide whether to verify the P-Node (Phase 1), and the P-Node uses it to decide whether to verify its own backend on every prompt (Phase 2).

**Implementation (current):**
```go
func IsTeeModel(tags []string) bool {
    for _, raw := range tags {
        if strings.ToLower(raw) == "tee" {
            return true
        }
    }
    return false
}
```

**Convention:** The canonical tag string is `"tee"` (matched case-insensitively). The v7 "tee tag for everything" refactor (PR #708, #709) removed the local `isTee` field from `models-config.json` entirely — the on-chain tag is now the sole source of truth.

**Effort:** S — **DONE**

### 7.4 Consumer-side attestation verification in session flow — DONE (v6.0.0, refined v6.2.x)

**Files:**
- `proxy-router/internal/attestation/verifier.go` — core `Verifier`, `VerifyProvider` (full), `VerifyProviderQuick` (fast path)
- `proxy-router/internal/blockchainapi/service.go` — calls `VerifyProvider` from `tryOpenSession` when `IsTeeSession` is true
- `proxy-router/internal/proxyapi/proxy_sender.go` — calls `verifyTEEAttestation` (→ `VerifyProviderQuick`) before every `SendPrompt`, `SendTranscription`, `SendSpeech`, and session-init handshake

**What shipped:** When the consumer's C-Node picks a bid whose model carries the `tee` tag, it verifies the provider's P-Node attestation before opening the session and re-verifies on every prompt. Failure means fail-over to the next bid at session open, or a hard prompt-level error once a session is live.

**Actual flow in `tryOpenSession` (v6+):**
```
if IsTeeSession && attestationVerifier != nil:
    1. Derive attestation VM domain from provider's Url
    2. Verifier.VerifyProvider(vmDomain, providerAddr):
       a. Fetch cosign-signed attestation manifest from GHCR for the `-tee` image
          (cosign Go library, GitHub-Actions OIDC identity pattern,
           type https://morpheusais.github.io/tee-attestation/v1)
       b. Extract expected RTMR3 + baked_env from manifest predicate
       c. Confirm baked_env.PROXY_STORE_CHAT_CONTEXT == "false",
          ENVIRONMENT == "production", correct ETH_NODE_CHAIN_ID for network
       d. GET https://{vm-domain}:29343/cpu → raw TDX quote hex
       e. POST quote to TEE_PORTAL_URL (SecretAI quote-parse) — confirms genuine TDX
       f. Extract RTMR3 from quote at byte offset 520 (48 bytes), compare to manifest
       g. Extract reportData[0:32], compare to SHA-256 of the connection's
          peer TLS certificate — anti-MITM
       h. Cache snapshot (quote hash, TLS fingerprint, expiry-free)
    3. If pass → InitiateSession / continue
       If fail → skip this provider, try next bid; all fail → hard error
```

**Per-prompt fast path (`VerifyProviderQuick`, called from `proxy_sender.go` before every forwarded prompt):**
```
1. Look up cached snapshot for providerAddr; no cache → re-run full VerifyProvider
2. Re-fetch :29343/cpu, compute SHA-256 of the quote
3. If quote hash == cached hash AND TLS fingerprint == cached fingerprint → pass
4. If quote hash changed → trigger full re-attestation
5. If TLS fingerprint changed → hard fail (MITM signal), refuse to forward
```

**Dependencies (resolved):**
- `github.com/sigstore/cosign/v2/pkg/cosign` — in-process Go verification, no shell-out, no binary added to image. Binary size impact acceptable for the `-tee` image (consumer proxy-router uses the same library).
- Go 1.25.x (per CI/CD)

**Effort:** L — **DONE**

### 7.5 Consumer UI: TEE badge and verification status — DONE

**File:** `ui-desktop/src/renderer/src/components/` (model list, session views)

**What shipped:** Visual indicators for TEE models in the desktop UI.

**Elements:**
- TEE badge/icon on models tagged `"tee"` in the model browser
- Verification status indicator during session creation
- Success and failure states

**Effort:** M — **DONE**

### 7.6 Verifiable per-message signing — DEFERRED to Phase 2b

**File:** New code in proxy-router (not yet added)

**What:** Use SecretVM's internal signing service (`http://172.17.0.1:49153/sign`) to sign every response with a TEE-bound key. The public key is embedded in the attestation quote's `reportdata`, so consumers can verify that each response genuinely came from inside the TEE.

**Why this matters:** Per-prompt fast-verify (§7.4) already re-checks the provider's CPU quote and TLS fingerprint on every request, so a compromise between session-open and the first prompt is detected before any inference ships. Per-message signing would tighten that further — prove that the response payload itself was produced inside the enclave, not just that the connection terminated there.

**Status:** Not shipped in v7.0.0. Deferred as "Phase 2b, lower priority" because the combination of TLS pinning (§7.7) + per-prompt fast-verify makes the practical attack window too narrow to justify the added per-response overhead at release time.

**Effort:** L
**Priority:** Phase 2b (post-v7)

### 7.7 Phase 2 — P-Node verifies its own backend LLM — DONE (v7.0.0, PR #699)

This is the Phase 2 capability that makes v7 the "full TEE" release. It closes the model-backend gap previously accepted in Phase 1 (Design Decision §2.4) without requiring CPU+GPU co-location in a single TDX VM.

**Files (all under `proxy-router/internal/attestation/`):**
- `backend_verifier.go` — `BackendVerifier.AttestBackend` (full) and `FastVerifyBackend` (per-prompt fast path)
- `workload_verifier.go` — registry lookup + RTMR3 recalculation for the backend
- `artifacts_registry.go` — downloads and caches the SecretVM TDX artifact registry CSV on a configurable interval (`ARTIFACT_REGISTRY_URL`, `ARTIFACT_REGISTRY_REFRESH_INTERVAL`)
- `nras_verifier.go` — NVIDIA NRAS v4 API integration; validates the returned JWT Entity Attestation Token
- `tdx_quote.go` — parses raw TDX quotes to extract MRTD + RTMR0-3 + reportData
- `rtmr.go` — SHA-384 extend chain implementation used to replay expected RTMR3

**Integration points:**
- `proxy-router/cmd/main.go` — wires `BackendVerifier` into startup; calls `AttestBackend` once per `tee`-tagged model
- `proxy-router/internal/proxyapi/proxy_receiver.go` — calls `FastVerifyBackend` in `SessionPrompt` before forwarding any inference for `tee`-tagged model sessions (hot path)
- `proxy-router/internal/aiengine/ai_engine.go` — returns a `PinnedHTTPClient` for TEE models; the client's `VerifyPeerCertificate` callback refuses any onward TLS cert whose SHA-256 doesn't match the attested fingerprint
- `proxy-router/internal/proxyapi/controller_http.go` — `GET /v1/models/attestation` returns per-model attestation state (verified / pending / failed + workload-match / last-success timestamp / error detail)

**Backend attestation endpoints (derived from each TEE model's `apiUrl` host + standard port 29343):**
- `:29343/cpu` — raw TDX CPU quote hex
- `:29343/gpu` — JSON containing `nonce`, `arch`, `evidence_list`
- `:29343/docker-compose` — exact `docker-compose.yaml` loaded into the backend VM (used for RTMR3 replay)

**Full attestation sequence (`AttestBackend`):**
```
1. GET :29343/cpu            → raw TDX quote
2. POST quote to TEE_PORTAL_URL (quote-parse) → portal verification
3. Extract reportData[0:32]   → compare with SHA-256 of live TLS cert
                                   → TLS binding proven
4. if ArtifactRegistry loaded:
     GET :29343/docker-compose
     Parse TDX quote: MRTD + RTMR0-3 + reportData
     Lookup (MRTD, RTMR0, RTMR1, RTMR2) in registry CSV → rootfs_data + secretvm_release
     Replay RTMR3 = SHA-384-extend-chain( SHA-256(docker-compose) + rootfs_data )
     if replayed RTMR3 != quote RTMR3 → fail (workload mismatch)
5. GET :29343/gpu             → {nonce, arch, evidence_list}
6. Extract reportData[32:64]  → compare with GPU nonce
                                   → CPU-GPU binding proven
7. if NRAS configured:
     POST evidence to NVIDIA NRAS /v2/attest/gpu
     Validate returned JWT EAT signature + claims
     (non-fatal on network failure — NRAS outage does not block inference,
      but CPU-GPU binding is still enforced)
8. Cache snapshot: {quote_hash, tls_fingerprint, workload_status,
                    last_verified_ts, compose_sha256}
```

**Per-prompt fast verify (`FastVerifyBackend`):**
- No TTL. Runs unconditionally on every inference prompt for a `tee`-tagged model (inside `proxy_receiver.SessionPrompt`, on the hot path).
- Always re-fetches `:29343/cpu` (~50 ms TLS handshake).
- `SHA-256(new_quote) == cached_quote_hash` AND `live_tls_fp == cached_tls_fp` → pass.
- Quote-hash change → run full `AttestBackend` again (backend restart / redeploy).
- TLS-fingerprint change → immediate hard fail, prompt refused (MITM signal).
- No cache → reject ("model not attested").
- Cached status `failed` → reject ("attestation failed").

**Measurement-register semantics (recap):**

| Register | Content | How verified |
|---|---|---|
| MRTD | VM firmware measurement (set at launch, immutable) | Artifact registry lookup |
| RTMR0 | VM configuration | Artifact registry lookup |
| RTMR1 | Kernel | Artifact registry lookup |
| RTMR2 | Initramfs | Artifact registry lookup |
| RTMR3 | Rootfs + docker-compose.yaml | Client-side replay and compared to quote |

**reportData layout (recap):**

| Bytes | Content | Purpose |
|---|---|---|
| 0–31 | SHA-256(TLS cert) | Anti-MITM: pins the attested TEE to a specific TLS identity |
| 32–63 | GPU nonce | CPU-GPU binding: GPU evidence can't be replayed from another box |

**What Phase 2 proves end-to-end:**
- The backend is genuine Intel TDX hardware running a known SecretVM firmware/kernel/initramfs build.
- The exact set of loaded models (the `MODELS=...` line in `docker-compose.yaml`) is what the operator declared — swapping any model, port, or env var changes RTMR3 and fails verification.
- The TLS endpoint serving inference terminates inside the attested enclave (no CDN/reverse-proxy MITM can sit between the P-Node and the backend).
- The GPU evidence is genuine NVIDIA hardware (per NRAS), and cryptographically bound to the same CPU quote (per `reportData[32:64]`).
- The backend identity hasn't changed since initial attestation, on a per-prompt basis.

**Effort:** L — **DONE**
**Priority:** Phase 2a — **shipped in v7.0.0**
**PRs:** #699 (Phase 2 main), #700 (Phase 2 merge to test), #704 (error wrapping), #705 (request_id in logs), #708 / #709 (consolidate `tee` tag as sole TEE switch)

See also: [`proxy-router/docs/tee-backend-verification.md`](../proxy-router/docs/tee-backend-verification.md) (developer reference with mermaid sequence + trust-chain diagrams).

---

## 8. Remaining Open Questions

1. **~~`reproduce-mr` artifact availability~~** — **RESOLVED.** Artifacts are publicly downloadable from [secret-vm-build/releases](https://github.com/scrtlabs/secret-vm-build/releases). Rootfs is ~464MB; cached in CI/CD. RTMR3 computation works with rootfs alone. Full RTMR0-2 still needs ACPI templates (not yet in the reproduce-mr repo).

2. **~~SecretVM release pinning~~** — **RESOLVED.** Pinned in `.github/tee/secretvm.env`. Attestation manifest includes `secretvm_release` and `rootfs_sha256`. Providers must deploy on the matching release for RTMR3 to match.

3. **~~Anti-MITM in practice~~** — **RESOLVED in Phase 2.** The approach is now enforced on both hops:
   - *Consumer → P-Node (Phase 1):* The consumer extracts the peer TLS cert during the `:29343/cpu` fetch and compares its SHA-256 to `reportData[0:32]` of the quote. Mismatch = session refused.
   - *P-Node → Backend (Phase 2):* Same check at `AttestBackend`, **plus** the `PinnedHTTPClient` used for all onward inference refuses any TLS certificate whose SHA-256 doesn't match the attested fingerprint (in `VerifyPeerCertificate`). This means a TLS-terminating CDN/reverse-proxy in front of the backend *cannot* be used for `tee`-tagged models — the inference connection will refuse to establish. Operators are documented to expose SecretVM's port 443 (Traefik sidecar with SecretVM certs) directly for `tee`-tagged deployments.

4. **~~Cosign verification in Go~~** — **RESOLVED.** `github.com/sigstore/cosign/v2/pkg/cosign` is compiled into the proxy-router binary (Go 1.25.x). Binary-size impact accepted. Used both for consumer-side Phase 1 attestation manifest verification and for any in-process cosign verification needs.

5. **ACPI templates for full RTMR0-2:** The `reproduce-mr` tool requires `template_qemu_cpu{N}.hex` files for ACPI table generation — not needed for Phase 1 CI/CD (we only verify RTMR3 against the published manifest) *nor* for Phase 2 (we verify MRTD + RTMR0-2 by lookup in the published SecretVM TDX artifact registry CSV, not by recomputation). Remaining only if we ever want to recompute RTMR0-2 ourselves.

---

## 9. Phased Implementation Plan

### Phase 1a — CI/CD Supply-Chain Hardening — DONE

| # | Task | Effort | Status |
|---|---|---|---|
| 1.1 | Add `id-token: write` permission to GHCR jobs | S | **DONE** — PR #646 |
| 1.2 | Add cosign keyless signing to `GHCR-Build-and-Push` | S | **DONE** — PR #646 |
| 1.3 | Add cosign keyless signing to `GHCR-Build-and-Push-TEE` | S | **DONE** — PR #646 |
| 1.4 | Capture image digest in both GHCR jobs, export as output | S | **DONE** — PR #646 |
| 1.5 | Add SBOM generation (syft) for TEE image | S | **DONE** — PR #646 |
| 1.6 | Generate TEE attestation manifest JSON | M | **DONE** — PR #646 |
| 1.7 | Sign and attach attestation manifest via `cosign attest` | S | **DONE** — PR #646 |
| 1.8 | Canonical compose format for SecretVM | S | **DONE** — PR #648 |
| 1.9 | User-facing TEE setup and verification docs | M | **DONE** — PR #650 |

### Phase 1b — RTMR3 Computation & Automated Deployment — DONE

| # | Task | Effort | Status | Notes |
|---|---|---|---|---|
| 1.10 | Standalone RTMR3 computation script (`compute-rtmr3.py`) | S | **DONE** | Matches reproduce-mr algorithm |
| 1.11 | SecretVM artifact config (`.github/tee/secretvm.env`) | S | **DONE** | Pins v0.0.25 prod rootfs; variant + version fully variablized |
| 1.12 | Generate deployed compose with `@sha256:DIGEST` reference | S | **DONE** | Template + sed, uploaded as artifact |
| 1.13 | Compute RTMR3 in CI/CD and populate attestation manifest | M | **DONE** | `measurements.intel_tdx.rtmr3` |
| 1.13a | Fix compose byte content for RTMR3 parity with SecretVM | S | **DONE** | Single `\n` ending (no blank line); verified via VM `/docker-compose` endpoint |
| 1.13b | Add compose SHA-256 + RTMR3 debug output to CI/CD step summary | S | **DONE** | Aids future troubleshooting of measurement mismatches |
| 1.13c | Variablize rootfs config — no hardcoded filenames in pipeline | S | **DONE** | `SECRETVM_ROOTFS_VARIANT` flows to cache key, download, RTMR3 compute, attestation manifest |
| 1.13d | Switch rootfs from `dev` to `prod` variant (v0.0.25) | S | **DONE** | SecretVM runs "environment prod"; `dev` rootfs produces wrong RTMR3 |
| 1.13e | Document upgrade procedure for SecretVM releases | S | **DONE** | `docs/02.3-proxy-router-tee.md` § "Upgrading SecretVM Artifacts" |
| 1.14 | Full `reproduce-mr` for RTMR0-2 (Intel TDX) | M | TODO | Blocked: ACPI templates from SCRT Labs |
| 1.15 | `sev-snp-measure` for AMD SEV measurement | M | TODO | Blocked: same as 1.14 |
| 1.16 | CVE scanning (Trivy/Grype) — advisory, then gate | S | TODO | — |
| 1.17 | Automated SecretVM release monitoring (scheduled GHA) | S | TODO | GitHub API pre-release caveat; see §6.4 |
| 1.18 | Auto-deploy to SecretVM test instance via CLI | S | **DONE** | Separate `Deploy-SecretVM-Test` job; `environment: test`; `test` + `cicd/*` branches |
| 1.19 | Post-deploy attestation verification | S | **DONE** | Extracts RTMR3 from raw TDX quote at byte offset 520; polls 12×30s; fails job on mismatch |
| 1.20 | Temporarily disable service/UI builds on cicd/* branches | S | **DONE** | `Build-Service-Executables` set to `if: false`; cascades to all UI jobs. Re-enable before merge to main |
| 1.21 | Testnet/mainnet TEE image split | S | **DONE** | `Dockerfile.tee` parameterized via ARG; pipeline sources `test.env` or `main.env` per branch — PR #669 |

### Phase 1c — Proxy-Router Code (Consumer verifies P-Node) — DONE (v6.0.0)

| # | Task | Effort | Status | Reference |
|---|---|---|---|---|
| ~~1.15~~ | ~~Extend `/healthcheck` with TEE metadata~~ | ~~S~~ | **DROPPED** | Not needed; TEE capability discovered via on-chain tag |
| ~~1.16~~ | ~~Add `/attestation/quote` proxy endpoint~~ | ~~M~~ | **DROPPED** | Not needed; SecretVM host already exposes `:29343/cpu` |
| 1.17 | `IsTeeModel()` helper in `model_tags.go` | S | **DONE** | §7.3 — consolidated in PR #708/#709 as sole TEE switch |
| 1.18 | Consumer-side attestation verification in session flow | L | **DONE** | §7.4 — `attestation/verifier.go`, called from `blockchainapi/service.go` + `proxyapi/proxy_sender.go` |
| 1.18a | Per-prompt `VerifyProviderQuick` fast path | M | **DONE** | PR #686 + #689 (quote-mismatch re-verify instead of hard fail) |
| 1.18b | Storage-layer activity tracking for TEE sessions | S | **DONE** | PR #692/#693 (per-entry Badger keys, improved GC) |
| 1.18c | Error-wrapping + request_id propagation for TEE failures | S | **DONE** | PR #704, #705 |
| 1.19 | Consumer UI: TEE badge and verification status | M | **DONE** | §7.5 — renderer components + session-view states |

### Phase 2a — P-Node verifies Backend LLM — DONE (v7.0.0)

| # | Task | Effort | Status | Reference |
|---|---|---|---|---|
| 2a.1 | `BackendVerifier` — full `AttestBackend` at startup | L | **DONE** | `attestation/backend_verifier.go`; PR #699 |
| 2a.2 | `FastVerifyBackend` per-prompt hot-path check | M | **DONE** | Hooked from `proxy_receiver.SessionPrompt`; PR #699 |
| 2a.3 | `WorkloadVerifier` — MRTD + RTMR0-2 registry lookup, RTMR3 replay | M | **DONE** | `attestation/workload_verifier.go`, `rtmr.go`, `tdx_quote.go` |
| 2a.4 | `ArtifactRegistry` — download + cache SecretVM TDX artifact CSV | S | **DONE** | `attestation/artifacts_registry.go`; configurable interval |
| 2a.5 | `NrasVerifier` — NVIDIA NRAS v4 evidence submission + JWT EAT validation | M | **DONE** | `attestation/nras_verifier.go` |
| 2a.6 | `PinnedHTTPClient` — refuse onward TLS certs whose fingerprint doesn't match attested value | S | **DONE** | `aiengine/` — custom `VerifyPeerCertificate` |
| 2a.7 | Health endpoint `GET /v1/models/attestation` | S | **DONE** | `proxyapi/controller_http.go` |
| 2a.8 | New env config: `TEE_PORTAL_URL`, `TEE_IMAGE_REPO`, `ARTIFACT_REGISTRY_URL`, `ARTIFACT_REGISTRY_REFRESH_INTERVAL` | S | **DONE** | `config/config.go` (§7.7) |
| 2a.9 | Consolidate `tee` on-chain tag as sole TEE switch (remove `isTee` field from models config schema) | S | **DONE** | PR #708 / #709 |
| 2a.10 | Test coverage: `backend_verifier_test.go`, `golden_test.go`, `workload_verifier_test.go`, `workload_rytn_test.go`, `nras_verifier_test.go` | M | **DONE** | PR #699 |

### Phase 2b — Deeper Guarantees (future, post-v7)

| # | Task | Effort | Status | Notes |
|---|---|---|---|---|
| 2b.1 | Co-locate proxy-router + LLM in same TEE VM (single RTMR3 covers both hops) | L | TODO | Collapses Phase 1 + Phase 2 into one measurement chain |
| 2b.2 | Verifiable per-message signing (TEE-bound key via SecretVM internal signer) | L | TODO | §7.6; deferred because Phase 2a fast-verify narrows the attack window significantly |
| 2b.3 | AMD SEV-SNP measurement path in CI/CD | M | TODO | Blocked on upstream tooling; TDX-only today |
| 2b.4 | Local quote verification in-process (remove SCRT Labs `quote-parse` dependency) | L | TODO | Requires PCK cert chain handling + Intel quote-verification library in Go |
| 2b.5 | On-chain measurement registry (if ever needed beyond cosign signatures) | L | TODO | Cosign keyless via GHA OIDC is sufficient today |
| 2b.6 | NRAS alternatives for non-NVIDIA GPU vendors | M | TODO | NVIDIA-only today |
| 2b.7 | CVE scanning gate in CI/CD (Trivy/Grype) | S | TODO | §6.6 |

---

## 10. Reference Links

| Resource | URL |
|---|---|
| **Delivered artifacts** | |
| Supply-chain hardening doc (in LMN) | `Morpheus-Lumerin-Node/.ai-docs/TEE_CICD_Supply_Chain_Hardening.md` |
| Provider/consumer TEE guide (in LMN) | `Morpheus-Lumerin-Node/docs/02.3-proxy-router-tee.md` |
| SecretVM provider quick-start (in LMN) | `Morpheus-Lumerin-Node/docs/02.4-proxy-router-secretvm-quickstart.md` |
| **Phase 2 developer reference (in LMN)** | `Morpheus-Lumerin-Node/proxy-router/docs/tee-backend-verification.md` |
| PR #644 — TEE image + BASE migration | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/644 |
| PR #646 — Supply-chain hardening | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/646 |
| PR #648 — Compose canonical format | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/648 |
| PR #650 — User-facing TEE docs | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/650 |
| PR #652 — SecretVM CLI instructions (from SCRT Labs) | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/652 |
| PR #669 — Testnet/mainnet TEE image split | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/669 |
| PR #686 — per-prompt TEE attestation with TLS binding and quote caching | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/686 |
| PR #689 — re-verify provider on quote hash mismatch instead of failing | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/689 |
| PR #692 — per-entry Badger activity tracking, improved GC reclaim | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/692 |
| **PR #699 — Phase 2 TEE backend verification (main)** | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/699 |
| PR #700 — Phase 2 merge to test | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/700 |
| PR #703 / #704 — correct error wrapping on P-Node TEE attestation fail | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/704 |
| PR #705 — request_id propagation in TEE logs | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/705 |
| PR #708 / #709 — `tee` on-chain tag as sole TEE switch (removed `isTee` from models-config schema) | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/709 |
| First verified pipeline run (build + sign) | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22920492249 |
| First end-to-end run (build → sign → deploy → verify attestation) | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions/runs/22969993910 |
| **SCRT Labs / TEE platform** | |
| SecretVM Attestation Docs | https://docs.scrt.network/secret-network-documentation/secretvm-confidential-virtual-machines/attestation |
| SecretVM CLI Docs | https://docs.scrt.network/secret-network-documentation/secretvm-confidential-virtual-machines/secretvm-cli |
| SCRT Labs reproduce-mr | https://github.com/scrtlabs/reproduce-mr |
| SCRT Labs secret-vm-build | https://github.com/scrtlabs/secret-vm-build |
| SCRT Labs quote parser API | `POST https://pccs.scrtlabs.com/dcap-tools/quote-parse` |
| AMD sev-snp-measure | https://github.com/virtee/sev-snp-measure |
| **Sigstore / supply-chain tools** | |
| Sigstore cosign | https://github.com/sigstore/cosign |
| Cosign keyless signing docs | https://docs.sigstore.dev/cosign/signing/signing_with_containers/ |
| Cosign Go library | https://pkg.go.dev/github.com/sigstore/cosign/v2/pkg/cosign |
| Syft SBOM generator | https://github.com/anchore/syft |
| SLSA framework | https://slsa.dev |
| **Proxy-router code references** | |
| Healthcheck handler | `proxy-router/internal/system/controller.go:72-101` |
| Healthcheck response struct | `proxy-router/internal/system/structs.go:19-24` |
| Model tags detection (`IsTeeModel`) | `proxy-router/internal/blockchainapi/model_tags.go` |
| Session creation flow + TEE branch | `proxy-router/internal/blockchainapi/service.go` (`OpenSessionByModelId`, `tryOpenSession`, sets `IsTee` on session) |
| InitiateSession handshake + `VerifyProviderQuick` callsites | `proxy-router/internal/proxyapi/proxy_sender.go`, `proxy_receiver.go` |
| **Phase 1 consumer verifier** | `proxy-router/internal/attestation/verifier.go` (`Verifier`, `VerifyProvider`, `VerifyProviderQuick`) |
| **Phase 2 backend verifier** | `proxy-router/internal/attestation/backend_verifier.go` (`BackendVerifier.AttestBackend`, `FastVerifyBackend`) |
| **Phase 2 workload verifier + RTMR3 replay** | `proxy-router/internal/attestation/workload_verifier.go`, `rtmr.go`, `tdx_quote.go` |
| **Phase 2 artifact registry** | `proxy-router/internal/attestation/artifacts_registry.go` |
| **Phase 2 NVIDIA NRAS client** | `proxy-router/internal/attestation/nras_verifier.go` |
| **Phase 2 pinned TLS client** | `proxy-router/internal/aiengine/ai_engine.go` (returns `PinnedHTTPClient` for TEE models) |
| **Phase 2 health endpoint** | `proxy-router/internal/proxyapi/controller_http.go` (`GET /v1/models/attestation`) |
| TEE config struct | `proxy-router/internal/config/config.go` — `TEE` section (lines ~87-92) |
| Session repository (tracks `IsTee`) | `proxy-router/internal/repositories/session/session_model.go`, `session_repo.go` |
| ModelRegistry contract | `smart-contracts/contracts/diamond/facets/ModelRegistry.sol` |
| Model struct (with tags) | `smart-contracts/contracts/interfaces/storage/IModelStorage.sol` |
