# TEE Attestation Architecture — Verifiable Provider Compute

**Status:** v1.0 — Full automated loop: build → sign → deploy → verify attestation  
**Last updated:** 2026-03-11

---

## 1. Goal

A consumer node should be able to **cryptographically verify**, before sending a prompt, that the far-side provider is:

1. Running on genuine TEE hardware (Intel TDX or AMD SEV-SNP)
2. Running an unmodified, known-good version of the proxy-router
3. Configured such that PII/chat logging is disabled and cannot be re-enabled at runtime
4. Not subject to MITM between the consumer and the TEE enclave

---

## 2. Scope — What We're Doing Now (Phase 1)

**In scope (CI/CD supply-chain hardening — our work):**
- Cosign keyless image signing (both standard and TEE images)
- Image digest capture and export
- SBOM generation and attachment (syft)
- Signed TEE attestation manifest (Option 5B — stored in GHCR as OCI artifact)
- Support both Intel TDX and AMD SEV-SNP from day one

**In scope (proxy-router code — developer work, later):**
- `IsTEEModel()` helper for on-chain tag detection
- Consumer-side attestation verification before session creation (fetches quote from SecretVM host endpoint at `:29343`)
- Consumer UI: TEE badge and verification status

**Out of scope for Phase 1:**
- On-chain oracle / DAO governance for measurements
- Model backend (LLM) attestation (accepted gap — later phases co-locate proxy + LLM in same TEE)
- Rating system integration
- CVE scanning gate (later)

### Design Decisions (from review)

| # | Question | Decision |
|---|---|---|
| 1 | Oracle governance | **Automated by CI/CD** — cosign keyless signing via GitHub Actions OIDC. No multi-sig/DAO for now. |
| 2 | SCRT Labs coupling | **Yes, couple** — use their quote-parse API and `reproduce-mr` tool. They control the VM layer; we need their artifacts. See §3. |
| 3 | AMD SEV support | **Both platforms from day one** if feasible; if `reproduce-mr` only handles TDX, compute what we can and extend for SEV in a fast follow. |
| 4 | Model backend trust | **Gap accepted for Phase 1.** Later phases co-locate proxy-router + LLM on the same TEE VM (CPU for proxy, GPU for inference), making RTMR3 cover both. |
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

### 7.3 `IsTEEModel()` helper and model tag convention

**File:** `proxy-router/internal/blockchainapi/model_tags.go`

**What:** Add a helper function to detect TEE models from their on-chain tags. The `ModelRegistry` contract already has `string[] tags` — providers registering TEE models add `"tee"` as a tag.

**Implementation:**
```go
func IsTEEModel(tags []string) bool {
    for _, raw := range tags {
        if strings.ToLower(raw) == "tee" {
            return true
        }
    }
    return false
}
```

**Convention:** The canonical tag string is `"tee"` (lowercase). Document this for providers in the model registration guide.

**Effort:** S

### 7.4 Consumer-side attestation verification in session flow

**File:** `proxy-router/internal/blockchainapi/service.go` (`OpenSessionByModelId`, `tryOpenSession`)

**What:** Before creating a session with a TEE-tagged model, verify the provider's image and hardware attestation. This is the core of the trust loop.

**Flow (inserted into the existing bid-ranking loop in `tryOpenSession`):**
```
if IsTEEModel(model.Tags):
    1. cosign.VerifyImageAttestations() → fetch signed manifest from GHCR
       - Verify signature chain (OIDC issuer = GitHub Actions)
       - Extract predicate: compose_sha256, baked_env, measurements
    2. Confirm baked_env.PROXY_STORE_CHAT_CONTEXT == "false"
    3. Confirm baked_env.ENVIRONMENT == "production"
    4. GET provider attestation endpoint (https://{vm-domain}:29343/cpu.html)
       → get raw TDX quote hex
    5. Extract RTMR3 from TDX quote at byte offset 520 (48 bytes)
       (same extraction logic used in CI/CD verification step)
    6. Compare RTMR3 against expected value from signed manifest
    7. Optionally: verify reportdata contains TLS certificate fingerprint (anti-MITM)
    8. If all pass → proceed with InitiateSession
    9. If any fail → log reason, skip this provider, try next bid
   10. If all providers fail → hard error to user: "No verified TEE providers available"
```

**Note:** The attestation endpoint is provided by the SecretVM host (port 29343), not the proxy-router container. The consumer needs to derive the VM domain from the provider's public URL to reach it.

**Dependencies:**
- `github.com/sigstore/cosign/v2/pkg/cosign` Go library for programmatic verification
- Need to confirm compatibility with Go 1.23.x and impact on binary size (`FROM scratch` image)
- Alternative: shell out to `cosign` binary (simpler but requires adding cosign to the image)

**Effort:** L

### 7.5 Consumer UI: TEE badge and verification status

**File:** `ui-desktop/src/renderer/src/components/` (model list, session views)

**What:** Visual indicators for TEE models in the desktop UI.

**Elements:**
- TEE badge/icon on models tagged `"tee"` in the model browser
- Verification status indicator during session creation ("Verifying TEE attestation...")
- Success state: "Verified — running in TEE enclave"
- Failure state: "Attestation verification failed — provider rejected"

**Effort:** M

### 7.6 Verifiable message signing (Phase 2, lower priority)

**File:** New code in proxy-router

**What:** Use SecretVM's internal signing service (`http://172.17.0.1:49153/sign`) to sign every response with a TEE-bound key. The public key is embedded in the attestation quote's `reportdata`, so consumers can verify that each response genuinely came from inside the TEE.

**Why this matters:** Steps 7.1-7.5 verify attestation at session start. Per-message signing provides continuous proof throughout the session — not just "this provider was in a TEE when the session started" but "this specific response came from the TEE."

**Effort:** L  
**Priority:** Phase 2

---

## 8. Remaining Open Questions

1. **~~`reproduce-mr` artifact availability~~** — **RESOLVED.** Artifacts are publicly downloadable from [secret-vm-build/releases](https://github.com/scrtlabs/secret-vm-build/releases). Rootfs is ~464MB; cached in CI/CD. RTMR3 computation works with rootfs alone. Full RTMR0-2 still needs ACPI templates (not yet in the reproduce-mr repo).

2. **~~SecretVM release pinning~~** — **RESOLVED.** Pinned in `.github/tee/secretvm.env`. Attestation manifest includes `secretvm_release` and `rootfs_sha256`. Providers must deploy on the matching release for RTMR3 to match.

3. **Anti-MITM in practice:** The `reportdata` field contains the TLS certificate fingerprint. When the consumer connects to the provider over HTTPS, does it see the VM's TLS cert (proving it's inside the TEE), or does it see a CDN/load-balancer cert (breaking the chain)? If providers use Cloudflare/nginx in front, the anti-MITM guarantee breaks. Need to understand the typical network topology.

4. **Cosign verification in Go:** The consumer proxy-router is written in Go. Cosign has a Go library (`github.com/sigstore/cosign/v2/pkg/cosign`) for programmatic verification. Need to confirm it's compatible with the proxy-router's Go version (1.22.x) and doesn't bloat the binary excessively (relevant for `FROM scratch` image).

5. **ACPI templates for full RTMR0-2:** The `reproduce-mr` tool requires `template_qemu_cpu{N}.hex` files for ACPI table generation. These aren't published in the reproduce-mr repo. Need to either obtain from SCRT Labs or generate from a QEMU build. Not blocking — RTMR3 is the critical register for our software layer verification.

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

### Phase 1c — Proxy-Router Code (Developer Work)

| # | Task | Effort | Status | Reference |
|---|---|---|---|---|
| ~~1.15~~ | ~~Extend `/healthcheck` with TEE metadata~~ | ~~S~~ | **DROPPED** | Not needed; TEE capability discovered via on-chain tag |
| ~~1.16~~ | ~~Add `/attestation/quote` proxy endpoint~~ | ~~M~~ | **DROPPED** | Not needed; SecretVM host already exposes `:29343/cpu.html` |
| 1.17 | `IsTEEModel()` helper in `model_tags.go` | S | TODO | §7.3 |
| 1.18 | Consumer-side attestation verification in session flow | L | TODO | §7.4 — fetch quote from `:29343`, verify RTMR3 against signed manifest |
| 1.19 | Consumer UI: TEE badge and verification status | M | TODO | §7.5 — frontend work, after 1.18 |

### Phase 2 — Deeper Guarantees (future)

| # | Task | Effort | Status |
|---|---|---|---|
| 2.1 | Co-locate proxy-router + LLM in same TEE VM | L | TODO |
| 2.2 | Verifiable per-message signing (TEE-bound key) | L | TODO |
| 2.3 | On-chain measurement registry (if needed beyond cosign) | L | TODO |
| 2.4 | Local quote verification (remove SCRT Labs API dependency) | L | TODO |

---

## 10. Reference Links

| Resource | URL |
|---|---|
| **Delivered artifacts** | |
| Supply-chain hardening doc (in LMN) | `Morpheus-Lumerin-Node/.ai-docs/TEE_CICD_Supply_Chain_Hardening.md` |
| Provider/consumer TEE guide (in LMN) | `Morpheus-Lumerin-Node/docs/02.3-proxy-router-tee.md` |
| PR #644 — TEE image + BASE migration | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/644 |
| PR #646 — Supply-chain hardening | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/646 |
| PR #648 — Compose canonical format | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/648 |
| PR #650 — User-facing TEE docs | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/650 |
| PR #652 — SecretVM CLI instructions (from SCRT Labs) | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/652 |
| PR #669 — Testnet/mainnet TEE image split | https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/pull/669 |
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
| Model tags detection | `proxy-router/internal/blockchainapi/model_tags.go` |
| Session creation flow | `proxy-router/internal/blockchainapi/service.go` (`OpenSessionByModelId`, `tryOpenSession`) |
| InitiateSession handshake | `proxy-router/internal/proxyapi/proxy_sender.go`, `proxy_receiver.go` |
| ModelRegistry contract | `smart-contracts/contracts/diamond/facets/ModelRegistry.sol` |
| Model struct (with tags) | `smart-contracts/contracts/interfaces/storage/IModelStorage.sol` |
