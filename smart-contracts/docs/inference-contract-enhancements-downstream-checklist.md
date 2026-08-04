# Inference Contract Enhancements — Downstream Work Checklist

**Depends on:** [`inference-contract-enhancements-rfp.md`](inference-contract-enhancements-rfp.md) (diamond work)  
**Companion:** [`inference-contract-enhancements-executive-summary.md`](inference-contract-enhancements-executive-summary.md)  
**Status:** Planning checklist (post-RFP / post-cut)  
**Last updated:** 2026-08-04  
**Audience:** Proxy-router, APIGW, docs, ops, and product owners

---

## Purpose

After the Inference Diamond RFP lands on-chain, this is the **scope of work outside the contracts**: node software, APIs, jobs, auxiliary sites, and nodedocs. Use it as a working checklist — not a second contract RFP.

**Rule of thumb:** the diamond is source of truth (RFP F6). Downstream only wraps, instruments, and explains those surfaces.

---

## 0. Cutover prerequisites (all teams)

- [ ] New/updated ABIs published from the release commit; Go bindings regenerated under `proxy-router/internal/repositories/contracts/bindings/`
- [ ] Diamond addresses / facet loupe verified on Sepolia then mainnet (runbooks in `smart-contracts/docs/`)
- [ ] Compatibility matrix: which proxy-router versions talk to which diamond cut (document in release notes)
- [ ] Feature flags / env defaults chosen for consumer vs provider builds (see §1)
- [ ] Staging (Sepolia / DEV APIGW) green before mainnet client rollouts

---

## 1. Proxy-router — shared (consumer + provider)

Regenerate and wire contract access first; then role-specific work below.

### 1.1 Bindings, registries, config

| Done | Work | RFP |
|------|------|-----|
| [ ] | Regenerate `SessionRouter`, `Marketplace`, `ProviderRegistry`, new `DelegateStaking` (name TBD), `QueryFacet` bindings | all |
| [ ] | Registry wrappers for new views/txs (`quoteSession`, `getClaimable`, `claimAvailable`, pool ops, `updateBidPrice`, `providerUpdateEndpoint`, `getFundingHealth`, fee summary, payout target) | §3.1–§3.5 |
| [ ] | Env / config docs: `proxy-router.all.env`, `.env.example`, `docs/reference/env-proxy-router.mdx` | — |
| [ ] | Version gate / healthcheck fields: diamond facet set or “claimable / pool” summaries (no secrets) | P4, P6, P7 |

### 1.2 Swagger / OpenAPI

| Done | Work | RFP |
|------|------|-----|
| [ ] | Update `proxy-router/docs/swagger.yaml` (+ generated `swagger.json` / `docs.go`) | all HTTP |
| [ ] | Document Basic Auth on every new blockchain route | — |
| [ ] | Mark behavioral change for `directPayment` / session open duration (in-place compatible) | P2 |
| [ ] | Retire “no HTTP route for `withdrawUserStakes`” language everywhere it appears | P4 |

### 1.3 Background jobs / cron-like loops

| Done | Work | Notes | RFP |
|------|------|-------|-----|
| [ ] | **Claim auto-loop** (`STAKE_AUTO_CLAIM` + interval): on startup + periodic `getClaimable` → `claimAvailable` / withdraw for the node wallet | Consumer-critical; useful on providers with `providerOwed` | P4 |
| [ ] | Failures must not block startup; log + retry | F2 | P4 |
| [ ] | Optional: permissionless `settleExpiredSession` sweeper for pools this node funded (or ops bot) | Dead-node backstop | P8 |
| [ ] | Revisit any existing Infra `cast send withdrawUserStakes` crons → prefer node auto-claim or retire | Ops debt | P4 |

---

## 2. Proxy-router — consumer path

| Done | Work | RFP |
|------|------|-----|
| [ ] | **Quotes before open:** call `quoteSession` / `quoteSessionByModel`; surface duration/`endsAt` in API/CLI responses | P1 |
| [ ] | **Session open:** pass `directPayment` correctly; stop assuming stipend math for direct-pay; validate amount vs desired seconds using quote | P2 |
| [ ] | **Session close:** handle “close succeeded, provider unpaid” (no longer a hard revert); surface session state | P3 |
| [ ] | **Claimable API:** `GET` claimable (locked / available / providerOwed); `POST` claim/withdraw | P4 / #827 |
| [ ] | Expose claimable on `/healthcheck` and/or `/blockchain/balance` | P4 |
| [ ] | **Delegated pool mode:** `STAKING_FUND_SOURCE=wallet\|pool\|auto`; `openSessionFromPool` when configured | P7 |
| [ ] | Pool observability: `GET /blockchain/pool` (available, free/locked, pending withdraws, funder counts) | P7 |
| [ ] | Cold-funder UX is off-node (wallet/`cast`/Safe); document three-tx approve/grant/fund + exit paths | P7 |
| [ ] | Rating / session selection: still off-chain; ensure it uses on-chain quotes + prices (not a wrong amount÷price for stake-pool) | P1 |
| [ ] | CLI parity for quote, open (both modes), claimable, pool status | P1–P4, P7 |

---

## 3. Proxy-router — provider path

| Done | Work | RFP |
|------|------|-----|
| [ ] | **Bid update:** `PATCH` (or equivalent) → `updateBidPrice`; show `bidUpdateFee` | P10 |
| [ ] | **Endpoint update:** route → `providerUpdateEndpoint` (stop teaching `providerRegister(..., 0)`) | P11 |
| [ ] | **Payout target:** set/get `payoutTarget`; ensure claims (incl. deferred `providerOwed`) land there | P9 |
| [ ] | **Earnings / owed:** wrap `getProviderEarningsStatus` (incl. `providerOwed`) | P6 |
| [ ] | **Claim path:** keep/extend `providerClaim` / claimable balance; align with unified `getClaimable` | P4, P9 |
| [ ] | **Fee awareness:** display net-of-fee expectations; read `getProtocolFeeSummary` if useful for ops | P5 |
| [ ] | Healthcheck: optional funding-health is **consumer/treasury** concern; providers need **owed + payout target** visibility | P6, P9 |
| [ ] | MyProvider / GUI contracts: same bid update + endpoint update + payout target flows | P9–P11 |

---

## 4. API Gateway (`api.mor.org` / Morpheus-Infra `03-morpheus_api`)

Hosted OpenAI-compatible gateway runs a **consumer** proxy-router (C-Node). It does not replace the diamond; it must adopt consumer behaviors at scale.

| Done | Work | RFP |
|------|------|-----|
| [ ] | Session open/close paths honor direct-pay vs stake-pool duration semantics | P2 |
| [ ] | Prefer on-chain quote (or cached quote) when sizing stake / session duration | P1 |
| [ ] | **Delegated staking pool** for production float: cold Safe(s) → grant/fund → C-Node `STAKING_FUND_SOURCE=pool\|auto` | P7 |
| [ ] | Secrets / terragrunt: pool mode env, auto-claim env, any new cookie/wallet assumptions | P4, P7 |
| [ ] | Auto-claim enabled on C-Node so day-locked / releasable MOR does not accumulate unnoticed | P4 |
| [ ] | Close-path resilience: funding shortfall must not brick user sessions (diamond CLOSE-R1); monitor `providerOwed` / funding health for **treasury** wallet | P3, P6 |
| [ ] | Optional ops: `settleExpiredSession` coverage if C-Node dies mid-flight | P8 |
| [ ] | Idle-session early-close policy re-validated against day-lock + claim auto-loop (no silent stranded float) | P4, P7 |
| [ ] | apidocs.mor.org: session economics / stake / refund language updated if user-visible | P1, P2, P4 |
| [ ] | Insights / reporting: new events (`UserStakeReleased`, provider debt, fee charged, pool draws) — lake + dashboards if needed | P4, P5 |

---

## 5. Auxiliary sites

### 5.1 [active.mor.org](https://active.mor.org)

Live marketplace JSON/UI (models, bids, status).

| Done | Work | RFP |
|------|------|-----|
| [ ] | Confirm bid list still correct after `updateBidPrice` (no delete+repost); refresh cadence OK | P10 |
| [ ] | If enriched QueryFacet helps, optionally reduce multicalls in the active-models builder | P12 |
| [ ] | No requirement to show pool/claimable state here (node-local) — unless product wants a global “funding health” badge later | P6 |

### 5.2 [tech.mor.org](https://tech.mor.org) (calc / session explainers)

| Done | Work | RFP |
|------|------|-----|
| [ ] | Session calculator: **two modes** — stake-pool (`stakeToStipend`) vs direct-pay (`amount / pricePerSecond`) | P1, P2 |
| [ ] | Update day-lock / unused vs used stipend copy to match diamond + claim UX | P4 |
| [ ] | Capital / runway tools: optional `getFundingHealth` inputs if exposed publicly or via internal API | P6 |
| [ ] | `CALC_DATA_SOURCES.md` / `AI_ASSISTANT_GUIDE.md` aligned | P1, P2 |

### 5.3 Other surfaces

| Done | Work | Notes |
|------|------|-------|
| [ ] | [myprovider.mor.org](https://myprovider.mor.org) | Bid update, endpoint update, payout target, fee/net display |
| [ ] | [app.mor.org](https://app.mor.org) / MorpheusUI | Quotes, direct-pay vs stake copy, claimable balances, optional pool status |
| [ ] | CLI (`cli/`) | Parity with proxy-router blockchain routes |
| [ ] | Agents / SDKs under `agents/` | Session open amount sizing via quote; claimable if agents hold wallets |

---

## 6. Nodedocs ([nodedocs.mor.org](https://nodedocs.mor.org))

Mintlify source: `/docs`. Update pages + `docs.json` nav; bump `last_verified`.

### 6.1 Cross-cutting concepts

| Done | Page / topic | RFP |
|------|--------------|-----|
| [ ] | [Session states / stake / close / claim](https://nodedocs.mor.org/concepts/sessions-stake-close-recover) | P2–P4, P8 |
| [ ] | [Where is my MOR?](https://nodedocs.mor.org/ai/where-is-my-mor) / [Why locked?](https://nodedocs.mor.org/ai/why-locked-in-contract) | P4 |
| [ ] | [Tokens and fees](https://nodedocs.mor.org/concepts/tokens-and-fees) / [Rewards](https://nodedocs.mor.org/concepts/rewards-and-economics) | P5 |
| [ ] | Glossary: `getClaimable`, pool terms, `payoutTarget`, fee bps, retire “no HTTP route” | P4, P7, P9 |
| [ ] | Myths: direct-pay ≠ stake math; open escrows; claim paths | P2 |

### 6.2 Consumer — operate / update / manage

| Done | Topic | RFP |
|------|-------|-----|
| [ ] | Quickstart: quote before open; choose stake-pool vs direct-pay deliberately | P1, P2 |
| [ ] | How to read claimable (locked vs available) and claim (API + auto-claim env) | P4 |
| [ ] | **Delegated staking guide:** cold grant/fund/revoke; `STAKING_FUND_SOURCE`; pool status; shared-liquidity disclosure; settleExpired backstop | P7, P8 |
| [ ] | Env reference for new consumer vars | P4, P7 |
| [ ] | Troubleshooting: “closed but MOR missing” → claimable / releaseAt; “funding short” no longer blocks consumer close | P3, P4 |

### 6.3 Provider — operate / update / manage

| Done | Topic | RFP |
|------|-------|-----|
| [ ] | Register / bid: prefer `updateBidPrice` over delete+repost; fee table | P10 |
| [ ] | Endpoint changes via `providerUpdateEndpoint` | P11 |
| [ ] | Set `payoutTarget` (cold) for earnings + deferred owed | P9 |
| [ ] | Read earnings, `providerOwed`, claim when funding catches up | P4, P6 |
| [ ] | Fee: provider nets `(1 - feeBps)`; consumer refunds untouched | P5 |
| [ ] | Pricing docs: still use active.mor.org; note in-place bid updates | P10 |
| [ ] | MyProvider walkthrough updates for the above | P9–P11 |

### 6.4 API reference tab

| Done | Work |
|------|------|
| [ ] | Regenerated from swagger after proxy-router routes land |
| [ ] | Examples for quote, claimable, pool, bid PATCH, endpoint update, payout target |

---

## 7. Suggested delivery waves

Order is operational, not contractual.

| Wave | Focus | Unblocks |
|------|-------|----------|
| **A** | Bindings + claimable HTTP + auto-claim + nodedocs “where is my MOR” / retire no-route language | P4, #827 |
| **B** | Quotes + direct-pay behavior in open/close + tech.mor.org calculators + consumer docs | P1, P2 |
| **C** | Provider bid/endpoint/payout-target routes + MyProvider + provider docs | P9–P11 |
| **D** | Fee display + earnings/funding-health views + rewards docs | P5, P6 |
| **E** | Delegated pool (node + APIGW + cold Safe runbook + consumer staking docs) + settle sweeper | P7, P8 |
| **F** | QueryFacet adoption (optional perf), active.mor.org multicall trim, Insights events | P12 |

---

## 8. Traceability (RFP problem → primary downstream)

| RFP | Primary downstream owners |
|-----|---------------------------|
| P1 Quotes | Proxy-router consumer, CLI, UI, tech.mor.org, nodedocs consumers |
| P2 Direct-pay math | Proxy-router open/close, APIGW, tech.mor.org, nodedocs myths/sessions |
| P3 Close vs funding | Proxy-router close handling, APIGW/treasury monitoring, nodedocs sessions |
| P4 Claimable | Proxy-router + auto-claim, Infra cron retirement, nodedocs MOR pages, swagger |
| P5 Provider fee | Proxy-router/MyProvider display, nodedocs rewards/fees, Insights |
| P6 Earnings / funding health | Proxy-router views, treasury ops, optional tech.mor.org |
| P7 Cold/hot pool | Proxy-router pool mode, APIGW Safe funding, nodedocs delegated staking |
| P8 Dead node settle | Optional sweeper job, nodedocs pool ops, F2 drills |
| P9 Payout target | Proxy-router provider + MyProvider + nodedocs provider manage |
| P10 Bid update | Proxy-router + MyProvider + active.mor.org sanity + provider docs |
| P11 Endpoint update | Proxy-router + MyProvider + provider docs |
| P12 Query facet | Proxy-router registries, optional active.mor.org builder |

---

## 9. Explicitly out of this checklist

- Solidity / diamondCut / Hardhat work (the RFP itself)
- ModelRegistry redesign, provider bond redesign, reward-limiter redesign (RFP out of scope)
- Changing day-lock policy
- Hosted Inference API **billing product** redesign (only C-Node economics + docs as above)

---

## 10. Revision

| Date | Change |
|------|--------|
| 2026-08-04 | Initial downstream checklist aligned to trimmed contract RFP |
