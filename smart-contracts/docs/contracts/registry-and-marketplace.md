# Providers, models and bids

Facets: `ProviderRegistry`, `ModelRegistry`, `Marketplace`. Storage: `ProviderStorage`,
`ModelStorage`, `BidStorage`, `MarketplaceStorage`. All amounts are in MOR (18 decimals);
the token address is `BidStorage.token`, set once by `__Marketplace_init`.

Every write below that names an owner (`provider_`, `modelOwner_`) may be called by that
address or by its delegate with the matching right (see [delegation.md](delegation.md)).
Tokens are always pulled from, and returned to, the owner itself, never the delegate.

## Providers

```solidity
providerRegister(address provider_, uint256 amount_, string endpoint_)
providerDeregister(address provider_)
providerSetMinStake(uint256)                       // owner
```

`Provider` record: `endpoint`, `stake`, `createdAt`, `limitPeriodEnd`, `limitPeriodEarned`,
`isDeleted`.

- **Register** pulls `amount_` from the provider and adds it to `stake`. The resulting stake
  must be at least `providerMinimumStake`. Registering again updates the endpoint and adds
  stake (upsert). A deregistered provider becomes active again.
- The first registration starts the **reward limiter** period:
  `limitPeriodEnd = createdAt + 365 days`. A provider can earn at most `stake` MOR per period
  (see [provider-payouts.md](provider-payouts.md#reward-limiter)).
- **Deregister** requires no active bids. It returns the stake, except the amount the
  provider already earned in the current limiter period: `stake - limitPeriodEarned` while
  the period is running, the full stake after it ends. This stops a provider from earning,
  withdrawing the stake and re-staking from another address to reset the limit. The kept
  part stays in `stake`; to withdraw it, register again (topping up to the minimum stake if
  needed) and deregister after the period ends.
- Debts recorded for a provider stay payable after deregistration.

Views: `getProvider`, `getProviderMinimumStake`, `getActiveProviders(offset, limit)`,
`getIsProviderActive`.

## Models

```solidity
modelRegister(address modelOwner_, bytes32 baseModelId_, bytes32 ipfsCID_, uint256 fee_,
              uint256 amount_, string name_, string[] tags_)
modelDeregister(bytes32 modelId_)
modelSetMinStake(uint256)                          // owner
getModelId(address account_, bytes32 baseModelId_) // keccak256(abi.encodePacked(account, baseModelId))
```

- A model's ID is per owner: two owners registering the same `baseModelId` get different
  models. Registering the same `(owner, baseModelId)` again updates metadata and adds stake.
- The resulting stake must be at least `modelMinimumStake`.
- **Deregister** requires no active bids for the model and returns the whole stake to the
  owner.
- `fee` is stored but not used by any settlement code.

Views: `getModel`, `getModelIds`, `getActiveModelIds`, `getModelMinimumStake`,
`getIsModelActive`.

## Bids

```solidity
postModelBid(address provider_, bytes32 modelId_, uint256 pricePerSecond_) returns (bytes32 bidId)
deleteModelBid(bytes32 bidId_)
setMarketplaceBidFee(uint256)                                   // owner
setMinMaxBidPricePerSecond(uint256 min_, uint256 max_)          // owner; min > 0, min <= max
withdrawFee(address recipient_, uint256 amount_)                // owner
```

- A bid is a provider's price per second for one model. Provider and model must be active,
  and the price must be within `[bidMinPricePerSecond, bidMaxPricePerSecond]`.
- Posting charges `bidFee` (pulled from the provider into `feeBalance`).
- Bid ID: `keccak256(abi.encodePacked(provider, modelId, nonce))`, where `nonce` counts bids
  for that provider/model pair. Posting a new bid deletes the previous active one for the
  same pair, so a pair has at most one active bid.
- Deleting sets `deletedAt`. A deleted bid cannot back a new session; sessions already open
  on it keep settling at its stored price.
- `withdrawFee` pays up to `feeBalance` (the amount is clamped) to `recipient_`.

Views: `getBid`, `isBidActive`, `getProviderBids`, `getProviderActiveBids`,
`getModelBids`, `getModelActiveBids`, `getBidId`, `getProviderModelId`, `getToken`,
`getBidFee`, `getFeeBalance`, `getMinMaxBidPricePerSecond`.

## Events

| Event | When |
|---|---|
| `ProviderRegistered(provider)` | Register or re-register |
| `ProviderDeregistered(provider)` | Deregister |
| `ProviderMinimumStakeUpdated(value)` | Owner setter |
| `ModelRegisteredUpdated(owner, modelId)` | Register or update |
| `ModelDeregistered(owner, modelId)` | Deregister |
| `ModelMinimumStakeUpdated(value)` | Owner setter |
| `MarketplaceBidPosted(provider, modelId, nonce)` | New bid |
| `MarketplaceBidDeleted(provider, modelId, nonce)` | Explicit delete or replacement by a new bid |
| `MaretplaceFeeUpdated(bidFee)` | Owner setter (the name's typo is part of the ABI) |
| `MarketplaceBidMinMaxPriceUpdated(min, max)` | Owner setter |
