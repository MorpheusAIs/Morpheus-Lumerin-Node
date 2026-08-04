# SessionRouter stipend day-lock — diamond upgrade runbook

Two-step, owner-signed upgrade for the #830 day-lock fix. **Same ceremony on BASE
Sepolia and BASE mainnet:** any EOA deploys facet bytecode; the diamond owner signs
exactly one `diamondCut`. No proxy-router / node changes.

| | BASE Sepolia | BASE mainnet |
|---|---|---|
| Diamond | `0x6e4d0B775E3C3b02683A6F277Ac80240C4aFF930` | `0x6aBE1d282f72B474E54527D93b979A4f64d3030a` |
| Owner | EOA `0x19ec1E4b714990620edf41fE28e9a1552953a7F4` | Gnosis Safe **5-of-9** `0x1FE04BC15Cf2c5A2d41a0b3a96725596676eBa1E` |
| Who signs the cut | Key-holder directly | Safe proposal → 5 sigs → execute |

Re-read `owner()` on-chain before starting (the calldata script prints it).

## What the cut does

Atomic `diamondCut`:

1. **Remove** every selector on the live SessionRouter facet (fresh loupe read).
2. **Add** all `ISessionRouter` selectors → new SessionRouter facet.
3. **Add** all `IStatsStorage` selectors → new SessionRouter facet.

ABI unchanged. Behavior change: session close day-locks the used stipend even on
late close (anchors to `min(closedAt, endsAt)`).

**Forward-only.** Do not re-add a pre-fix SessionRouter facet. Hotfix = new facet
via the same two-step ceremony.

## Step 0 — preflight (deployer)

From `smart-contracts/` on the release commit:

```bash
npm ci && npx hardhat compile
npx hardhat test test/diamond/facets/SessionRouter.test.ts
```

`.env`: `PRIVATE_KEY` (throwaway EOA), `ALCHEMY_KEY` (or public RPC), `BASESCAN_API_KEY`.
Fund the deployer with gas ETH.

## Step 1 — deploy facet (permissionless)

```bash
# Sepolia
npx hardhat migrate --network base_sepolia --only 5 --verify

# Mainnet (same steps)
npx hardhat migrate --network base --only 5 --verify
```

Records library + new SessionRouter addresses. **No diamond state changes yet.**

## Step 2 — generate owner calldata (read-only)

```bash
# Sepolia (DIAMOND defaults to config_base_sepolia.json)
NEW_SESSION_ROUTER_FACET=0x... \
  npx hardhat run scripts/session-day-lock-cut-calldata.ts --network base_sepolia

# Mainnet — DIAMOND required (config-parser is Sepolia-default)
NEW_SESSION_ROUTER_FACET=0x... \
  DIAMOND=0x6aBE1d282f72B474E54527D93b979A4f64d3030a \
  npx hardhat run scripts/session-day-lock-cut-calldata.ts --network base
```

Prints owner, selector lists, and the single `to` / `value: 0` / `data` hex.

## Step 3 — review (before signing)

- [ ] New facet verified on Basescan; source matches the release commit.
- [ ] Remove target equals live SessionRouter (`facetAddress(closeSession)`).
- [ ] No unexpected selector drops; brand-new list is empty (ABI unchanged).
- [ ] Simulate the cut (Tenderly / Safe UI).
- [ ] `to` = diamond, `value` = 0.

## Step 4 — owner signs

- **Sepolia:** EOA sends the printed tx to the diamond.
- **Mainnet:** Safe → Transaction Builder → paste `data` (or `diamondCut`) → 5-of-9 → execute.

## Step 5 — verify

- Loupe: old SessionRouter gone; new facet serves `ISessionRouter` + `IStatsStorage`.
- Smoke: open a short funded session, late-close same day → `getUserStakesOnHold` shows a lock until next 00:00 UTC; unused stake returned immediately.
