# Morpheus Inference Contracts

Technical documentation for the contracts in `smart-contracts/contracts/`: the inference
marketplace and the RFP v2 Stage 1 additions (delegate staking pools, provider debt,
provider payout target, rating events).

## Documents

| Document | Contents |
|---|---|
| [registry-and-marketplace.md](registry-and-marketplace.md) | Providers, models, bids, fees, stake withdrawal |
| [sessions.md](sessions.md) | Regular sessions: approvals, receipts, stipend math, close, day-lock, claims, stats |
| [provider-payouts.md](provider-payouts.md) | Reward accrual, reward limiter, provider debt and its priority, payout target, emission views |
| [delegate-staking.md](delegate-staking.md) | Cold/hot staking pools: grants, funding, draws, locks, withdrawal queue, expired-session settlement |
| [delegation.md](delegation.md) | Delegation rights and which function checks which right |
| [providers-delegate.md](providers-delegate.md) | `ProvidersDelegate` staking proxy and `DelegateFactory` |
| [reference.md](reference.md) | Facet routing, access control matrix, parameters, events, errors, storage slots |

Related documents in `smart-contracts/docs/`:

- [client-impact.md](../client-impact.md): what Stage 1 changes for proxy-router, wallets and indexers.
- [delegate-staking-funder-economics.md](../delegate-staking-funder-economics.md): what a pool funder can rely on.
- [privacy-assessment.md](../privacy-assessment.md): on-chain linkability of the new features.
- [session-day-lock-upgrade-runbook.md](../session-day-lock-upgrade-runbook.md): the previous SessionRouter cut ceremony.
- Product docs (user-facing behaviour): `docs/` at the repository root.

## What the system does

A consumer pays for AI inference by **staking** MOR, not by spending it. The amount staked
buys a session length (the *stipend*). When the session ends, the consumer gets the stake
back; the provider is paid from the protocol **funding account**, which holds the compute
share of the MOR emission. A consumer may instead choose **direct pay**: the provider is
paid out of the session stake and only the rest returns.

Actors:

| Actor | Role |
|---|---|
| Provider | Registers with a stake, posts bids (price per second) for models, signs session approvals and close receipts, receives rewards |
| Model owner | Registers a model with a stake |
| User (consumer) | Opens sessions with a provider approval, closes them, withdraws day-locked stake |
| Hot wallet | Stage 1: the consumer node wallet that opens pool sessions without holding MOR |
| Funder (cold wallet) | Stage 1: puts MOR into a hot wallet's pool; only the funder can get it back |
| Payout target | Stage 1: the address that receives a provider's rewards (default: the provider) |
| Keeper | Anyone: the permissionless calls (claims, settlement, queue service, lock release, purge) |
| Owner | Diamond owner (Gnosis Safe on BASE mainnet): facet cuts and parameters |
| Funding account | External account that holds compute MOR and approves the diamond to pull it |

## Architecture

The protocol is one [EIP-2535](https://eips.ethereum.org/EIPS/eip-2535) diamond,
`LumerinDiamond` (Solarity `OwnableDiamond`). All facets run in the diamond's context and
share its MOR balance. Each storage area is a struct at a fixed `keccak256` slot
("namespaced storage"), so facets can be replaced without moving data.

```
                         LumerinDiamond (owner, diamondCut, loupe)
   ┌───────────────┬──────────────┬─────────────┬───────────────┬──────────────────┐
ProviderRegistry ModelRegistry  Marketplace   SessionRouter   Delegation
                                              (regular        (registry address)
                                               sessions)
   Stage 1:  ProviderPayouts      DelegateStaking        DelegateStakingFunding
             (payout target,      (pool sessions:        (grants, funding,
              debt views,          open, close,           withdrawals, locks,
              moved views)         settle expired)        parameters, views)

   Shared code (abstract, compiled into each facet that inherits it):
     SessionSettlement   — approval/receipt checks, stipend math, user settlement,
                           reward accrual, debt payment, stats
     DelegateStakingPool — pool lists, draws, returns, locks, withdrawal service
```

Stage 1 split the logic into several facets because of the 24,576-byte contract size limit
(EIP-170). `SessionRouter`, `DelegateStaking` and `ProviderPayouts` all inherit
`SessionSettlement`, so the regular and pool paths run the same settlement code.

Outside the diamond:

| Contract | Purpose |
|---|---|
| `DelegateFactory` | UUPS-upgradeable factory that deploys `ProvidersDelegate` beacon proxies |
| `ProvidersDelegate` | A provider owned by many stakers: pools their MOR as provider stake and shares rewards |
| Delegate registry | External [delegate.xyz v2](https://docs.delegate.xyz) registry used for all delegation checks |
| MOR | The token: stake, fees, rewards. Assumed to be a standard ERC-20 (no fees, hooks, rebasing or blacklist) |

`contracts/tokens/LumerinToken.sol` is a legacy 8-decimal token and is not used by the
diamond. `contracts/mock/` contains test-only contracts (MOR mock, registry copy,
`SessionRouterLegacy` for upgrade tests, `ProviderOwnerMock`).

## Where MOR sits

All MOR held by the diamond is in one balance. Accounting separates it into:

| Bucket | Funded by | Leaves to |
|---|---|---|
| Provider stakes | `providerRegister` | Provider on `providerDeregister` (limited by the current period's earnings) |
| Model stakes | `modelRegister` | Model owner on `modelDeregister` |
| Bid fees (`feeBalance`) | `postModelBid` | Owner's recipient via `withdrawFee` |
| Regular session stakes | `openSession` | User at close; provider (direct pay); day-locked part via `withdrawUserStakes` |
| User on-hold entries | Day-locked part of a regular close | User via `withdrawUserStakes` after release |
| Pool balances (`free`, `inSession`, `locked`, `pendingTotal`) | `fundStakingAllowance` | Only the funder, via the withdrawal path |

Provider rewards for non-direct sessions never pass through the diamond's balance: they are
pulled with `transferFrom` from the funding account straight to the payout target. Unpaid
rewards are recorded as debt and paid the same way later.

## Time conventions

- Day boundaries are UTC midnights (`timestamp - timestamp % 1 days`).
- Signatures (approvals and receipts) are valid for `SIGNATURE_TTL = 10 minutes`.
- Minimum session length is `MIN_SESSION_DURATION = 5 minutes`.
- Provider reward limiter period: 365 days.

## Build and test

From `smart-contracts/`:

```bash
npm ci
npm run compile          # hardhat compile, solc 0.8.24, optimizer 200 runs, evmVersion paris
npm test                 # generate-types + full hardhat test suite
npx hardhat test test/diamond/facets/DelegateStaking.test.ts   # one file
npm run test:gas         # gas report
npm run coverage
npm run bindings-go      # Go bindings for proxy-router
```

Test layout: `test/diamond/facets/*.test.ts` (one file per facet),
`test/diamond/DiamondUpgrade.test.ts` (cut from the legacy router, storage and behaviour at
the boundary, facet sizes), `test/diamond/ProtocolProperties.test.ts` (randomized invariant
runs, gas at caps, bus-factor drill), `test/delegate/` (ProvidersDelegate, factory).

## Deployments

| Network | Diamond | Owner |
|---|---|---|
| BASE mainnet | `0x6aBE1d282f72B474E54527D93b979A4f64d3030a` | Gnosis Safe 5-of-9 `0x1FE04BC15Cf2c5A2d41a0b3a96725596676eBa1E` |
| BASE Sepolia | `0xA328196f2438DADA5ab729E39388D86896c27c85` | EOA `0x94C58f682A9A518EfF83c7B53F3738294bB0F357` |

Network configs (MOR, funding account, registry, pools, minimum stakes) are in
`deploy/data/config_*.json`. The Stage 1 facets are live only after the Stage 1
`diamondCut` is executed on a network; the cut replaces the SessionRouter selectors, moves
the views listed in [reference.md](reference.md#facet-routing) to `ProviderPayouts`, adds
`DelegateStaking` and `DelegateStakingFunding`, and calls `__ProviderPayouts_init`.

Upgrades are forward-only: a fix ships as a new facet. Never cut back a pre-Stage-1
SessionRouter, or remove `DelegateStaking`, while pool state exists.

## Known limitations

- `Model.fee` is stored but not used anywhere.
- Regular `openSession` does not check the user named in the provider approval (legacy
  behaviour); the pool path does.
- A regular session has no permissionless settlement: if the user never closes it, the
  provider can still be paid with `claimForProvider`, but the stake stays until the user
  (or its session delegate) closes. Pool sessions have `settleExpiredSession`.
- `userStakesOnHold` grows by one entry per day-locked regular close; `withdrawUserStakes`
  takes an iteration limit.
- The running-average code in `LibSD` has open TODOs about overflow; it only feeds the
  stats views.
- `setPayoutTarget` rejects the diamond address but not other addresses that cannot spend
  tokens (for example the MOR contract); a wrong target is the provider's own risk.
- `setRegistry` does not validate the registry address. A wrong value makes every delegated
  call revert, and it also breaks the signature check for any approval or receipt not signed
  by the provider address itself (contract providers such as `ProvidersDelegate`, delegated
  signers, disputed receipts): opening and closing those sessions reverts even when the user
  calls directly. Only direct calls whose messages are signed by the provider EOA keep
  working.
