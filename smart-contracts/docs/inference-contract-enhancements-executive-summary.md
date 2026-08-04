# Inference Contract Enhancements: Executive Summary

**Companion to:** [`inference-contract-enhancements-rfp.md`](inference-contract-enhancements-rfp.md)  
**Date:** 2026-08-04  
**Scope:** Inference Diamond contract changes only (proxy-router / crons are follow-on callers)

---

## Guiding principles (RFP §2)

- **F1** Additive upgrades; no Breaking cutovers (in-place ABI-compatible tighten OK).
- **F2** No stranded funds; settlement/claim paths stay permissionless without privileged keys.
- **F3** Minimal governance — owner tunes bounds; everyone can bid/settle/claim.
- **F4** Direct-pay ≠ stake-pool math; emission/fee accounting conserves; fee path never bricks close.
- **F5** Bounded gas; delegated pool MOR never touches the hot wallet.
- **F6** Diamond is source of truth for quotes, debts, and claims.

## Problems → what we need (RFP §1.1)

| # | Problem | What we need |
|---|---------|--------------|
| P1 | Opaque session duration (mode-dependent math) | On-chain quotes with `openSession` parity **per mode** (§3.3.1) |
| P2 | Direct-pay still follows staking/stipend math | Direct-pay = consumer escrow × `pricePerSecond` → seconds; immediate unused return. **Not stake-pool** (§3.3.2) |
| P3 | Close reverts when funding wallet is short | Finish consumer close; record provider debt at accrual (§3.3.3) |
| P4 | Diamond debts hard to see/claim | Unified `getClaimable(addr)` + bounded claim — consumer on-hold **and** provider owed (§3.3.3) |
| P5 | No protocol take on payouts | Configurable provider fee; conserved emission accounting (§3.3.4) |
| P6 | Earnings / funding runway opaque | `getProviderEarningsStatus` (incl. owed) + `getFundingHealth` (§3.3.5) |
| P7 | Hot-wallet treasury exposure | Cold→hot purpose pool; `openSessionFromPool`; funds never to hot (§3.4.1) |
| P8 | Dead node strands capital | `settleExpiredSession`; provider claim decoupled from funding wallet (§3.4.1, §3.3.3) |
| P9 | Provider income on hot EOA | `payoutTarget` for immediate and deferred claims (§3.4.2) |
| P10 | Bid update = delete + repost fee | `updateBidPrice` (§3.1.1) |
| P11 | Endpoint update via re-register | `providerUpdateEndpoint` + bounds (§3.2.1) |
| P12 | Fragmented reads | QueryFacet enriched views (§3.5) |

Full requirements: [`inference-contract-enhancements-rfp.md`](inference-contract-enhancements-rfp.md).  
Post-RFP downstream checklist (proxy-router, APIGW, sites, nodedocs): [`inference-contract-enhancements-downstream-checklist.md`](inference-contract-enhancements-downstream-checklist.md).
