# ProvidersDelegate and DelegateFactory

`contracts/delegate/`. These contracts are outside the diamond. A `ProvidersDelegate` is a
provider owned by many stakers: it registers itself in the diamond as a provider, uses the
stakers' MOR as the provider stake, and shares the rewards it receives.

## DelegateFactory

UUPS-upgradeable, pausable, ownable. Holds an `UpgradeableBeacon` for the
`ProvidersDelegate` implementation.

```solidity
deployProxy(address feeTreasury_, uint256 fee_, string name_, string endpoint_,
            uint128 deregistrationOpensAt_) returns (address)          // whenNotPaused
predictProxyAddress(address deployer_) returns (address)               // next proxy of a deployer
```

- Each call deploys a `BeaconProxy` with CREATE2 (salt = `keccak256(deployer, number of
  proxies the deployer already has)`), initialises it and transfers its ownership to the
  caller. `proxies(deployer, i)` lists them.
- `deregistrationOpensAt_` must be later than `now + minDeregistrationTimeout`. The proxy
  initialiser also requires `fee_ <= PRECISION`, a non-empty name and endpoint, and a
  non-zero fee treasury.
- Owner: `pause`, `unpause`, `setMinDeregistrationTimeout`, `updateImplementation`
  (upgrades the beacon, so **all** proxies at once), UUPS upgrade of the factory.

Events: `ProxyDeployed(proxy)`, `MinDeregistrationTimeoutUpdated(value)`.

## ProvidersDelegate

State: diamond and token addresses, `feeTreasury`, `fee` (share of claims, out of
`PRECISION = 10^25`), `name`, `endpoint`, staking totals, `stakers`, `isStakeClosed`,
`deregistrationOpensAt`. The initialiser approves the diamond to pull unlimited MOR from the
proxy.

### Staking

- `stake(amount)` moves MOR from the staker to the proxy and calls
  `providerRegister(proxy, amount, endpoint)`, so the diamond pulls it as provider stake.
- `isStakeClosed` (owner) blocks staking, except when deregistration is open and the
  provider is deregistered with stake left in the diamond.
- There is no unstake. The stake comes back when the provider is deregistered: the diamond
  returns it to the proxy and it is distributed like rewards.

### Rewards

- MOR that arrives at the proxy while `totalStaked > 0` (provider rewards, returned stake)
  is shared pro rata to `staked` through a rate accumulator:
  `rate += (balance − lastContractBalance) * PRECISION / totalStaked`. A balance present
  before the first stake becomes the accounting baseline and is not distributed.
- `claim(staker, amount)` can be called by anyone and pays the staker
  `min(amount, balance, pending rewards)`; `fee` of it goes to `feeTreasury`. It reverts with
  `ClaimAmountIsZero` if that is 0. `stakers(s).claimed` counts the gross amount, before the
  fee.
- `restake(staker, amount)` claims and stakes again. Caller: the staker, or the owner unless
  the staker set `setIsRestakeDisabled(true)`.
- `claimForProvider(sessionId)` forwards to the diamond.
- The proxy has no call to `setPayoutTarget`, so its rewards always arrive at the proxy and
  enter staker accounting.

### Provider operations

| Function | Caller |
|---|---|
| `postModelBid(modelId, pricePerSecond)` | Owner, only before `deregistrationOpensAt`; the owner pays the bid fee |
| `deleteModelBids(bidIds)` | Owner; anyone once deregistration is open |
| `providerDeregister(bidIds)` | Owner; anyone once deregistration is open. Deletes the bids, deregisters, sets `fee = 0` so returned stake is not charged |
| `setName`, `setEndpoint`, `setFeeTreasury`, `setIsStakeClosed` | Owner |

Approvals and receipts for sessions with this provider are signed by the proxy's `owner()`
(see [delegation.md](delegation.md#signer-delegation-for-providers)).

The diamond's limiter applies to the proxy like to any provider: at deregistration, the
amount earned in the current period stays in the diamond until the period ends
([registry-and-marketplace.md](registry-and-marketplace.md#providers)).

Events: `Staked`, `Claimed`, `FeeClaimed`, `NameUpdated`, `EndpointUpdated`,
`FeeTreasuryUpdated`, `IsStakeClosedUpdated`, `IsRestakeDisabledUpdated`.
