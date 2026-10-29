# Reference

## Facet routing

The diamond routes each selector of these interfaces (including inherited storage
interfaces) to one facet:

| Facet | Interfaces | Library link |
|---|---|---|
| `ProviderRegistry` | `IProviderRegistry` (+ `IProviderStorage`) | — |
| `ModelRegistry` | `IModelRegistry` (+ `IModelStorage`) | — |
| `Marketplace` | `IMarketplace` (+ `IMarketplaceStorage`), `IBidStorage` | — |
| `SessionRouter` | `ISessionRouter` (+ `ISessionStorage`), `IStatsStorage` | `LinearDistributionIntervalDecrease` |
| `Delegation` | `IDelegation` (+ `IDelegationStorage`) | — |
| `ProviderPayouts` | `IProviderPayouts` | `LinearDistributionIntervalDecrease` |
| `DelegateStaking` | `IDelegateStaking` | `LinearDistributionIntervalDecrease` |
| `DelegateStakingFunding` | `IDelegateStakingFunding` | — |

The upgrade cut onto a diamond running the earlier SessionRouter:

1. **Replace** all `ISessionRouter` and `IStatsStorage` selectors → new `SessionRouter`.
2. **Replace** the selectors that moved → `ProviderPayouts`: `getSessionId`,
   `getSessionEnd`, `stakeToStipend`, `stipendToStake`, `getUserStakesOnHold`,
   `getTodaysBudget`, `getComputeBalance`, `totalMORSupply`, `startOfTheDay`.
3. **Add** the remaining `IProviderPayouts` selectors → `ProviderPayouts`.
4. **Add** `IDelegateStaking` → `DelegateStaking`, `IDelegateStakingFunding` →
   `DelegateStakingFunding`.
5. Init call: `ProviderPayouts.__ProviderPayouts_init()` (sets the
   `lifetimeClaimed` start time).

The operator procedure is in [stage-1-upgrade-runbook.md](../stage-1-upgrade-runbook.md).
`test/helpers/deployers/diamond/facets/diamond-upgrade.ts` builds exactly this cut; the
upgrade tests check the final routing and facet sizes.

| Facet | Bytecode size (bytes, limit 24,576) |
|---|---|
| `DelegateStaking` | 21,817 |
| `SessionRouter` | 19,403 |
| `DelegateStakingFunding` | 17,175 |
| `ProviderPayouts` | 11,557 |

## Storage slots

| Slot string | Struct | Contents |
|---|---|---|
| `diamond.standard.providers.storage` | `PovidersStorage` | Providers, minimum stake, active set |
| `diamond.standard.models.storage` | `ModelsStorage` | Models, minimum stake, ID sets |
| `diamond.standard.bids.storage` | `BidsStorage` | MOR address, bids, bid sets, nonces |
| `diamond.standard.market.storage` | `MarketStorage` | Fee balance, bid fee, price range |
| `diamond.standard.sessions.storage` | `SessionsStorage` | Funding account, pools, sessions, nonce, on-hold, used approvals, `providersTotalClaimed` |
| `diamond.standard.stats.storage` | `STTSStorage` | Provider/model and model stats |
| `diamond.standard.delegation.storage` | `DLGTNStorage` | Registry address |
| `diamond.standard.provider.payout.storage` | `ProviderPayoutsStorage` | Payout targets, debt, `pendingProviderClaims`, `lifetimeClaimed` |
| `diamond.standard.delegate.staking.storage` | `DelegateStakingData` | Pools, pool sessions, parameters |

Slots are `keccak256(slot string)`. Owner and initializer flags live in Solarity's diamond
storage. Existing structs must only be extended at the end, never reordered.

## Access control

| Caller | Functions |
|---|---|
| Owner | `diamondCut`, `transferOwnership`, `providerSetMinStake`, `modelSetMinStake`, `setMarketplaceBidFee`, `setMinMaxBidPricePerSecond`, `withdrawFee`, `setPoolConfig`, `setMaxSessionDuration`, `setRegistry`, `setMaxActiveFunders`, `setMinPrincipal`, `setSettlementGrace`, `setMaxAutoService`, `setMaxAutoReleaseDays` |
| Principal or delegate (right in [delegation.md](delegation.md)) | `providerRegister`, `providerDeregister`, `modelRegister`, `modelDeregister`, `postModelBid`, `deleteModelBid`, `openSession`, `closeSession`, `withdrawUserStakes`, `openSessionFromPool`, `closePoolSession`, `approvePoolFunder`, `revokePoolFunderApproval`, `requestStakingWithdrawal`, `cancelStakingWithdrawal` |
| Provider or explicit payout delegate | `setPayoutTarget` |
| Funder itself only | `grantStakingAllowance`, `revokeStakingAllowance`, `fundStakingAllowance` |
| Anyone | `claimForProvider`, `settleExpiredSession`, `servePoolWithdrawals`, `releasePoolLocks`, `purgeInactiveFunders`, all views |
| Once, at deploy or cut | `__LumerinDiamond_init`, `__ProviderRegistry_init`, `__ModelRegistry_init`, `__Marketplace_init`, `__SessionRouter_init`, `__Delegation_init`, `__ProviderPayouts_init` |

All initialisers are one-shot (a flag per storage slot). `__Marketplace_init`,
`__SessionRouter_init` and `__Delegation_init` call `onlyOwner` setters, so only the owner
can run them. `__LumerinDiamond_init` (it makes the caller the owner),
`__ProviderRegistry_init`, `__ModelRegistry_init` and `__ProviderPayouts_init` have no owner
check: run each right away. The upgrade cut calls `__ProviderPayouts_init` inside the cut
transaction; the deploy script calls the others immediately after deployment.

## Constants and parameters

| Name | Value | Where |
|---|---|---|
| `SIGNATURE_TTL` | 10 minutes | Approvals and receipts |
| `MIN_SESSION_DURATION` | 5 minutes | Session length and `setMaxSessionDuration` floor |
| `COMPUTE_POOL_INDEX` | 3 | Compute pool in `pools` |
| `PROVIDER_REWARD_LIMITER_PERIOD` | 365 days | Reward limiter |
| `PROVIDER_OWNER_STATICCALL_GAS` | 100,000 | `owner()` lookup for contract providers |
| `maxSessionDuration` | owner, ≥ 5 minutes (deploy: 7 days) | Session length cap |
| `providerMinimumStake`, `modelMinimumStake`, `bidFee`, bid price range | owner | Registry and marketplace |
| Pool parameters | see [delegate-staking.md](delegate-staking.md#parameters) | Delegate staking |
| `MAX_POOL_SESSION_DRAWS` | 129 | Draws per pool session |

## Events by facet

| Facet | Events |
|---|---|
| `ProviderRegistry` | `ProviderRegistered`, `ProviderDeregistered`, `ProviderMinimumStakeUpdated` |
| `ModelRegistry` | `ModelRegisteredUpdated`, `ModelDeregistered`, `ModelMinimumStakeUpdated` |
| `Marketplace` | `MarketplaceBidPosted`, `MarketplaceBidDeleted`, `MaretplaceFeeUpdated`, `MarketplaceBidMinMaxPriceUpdated` |
| `SessionRouter` | `SessionOpened`, `SessionClosed`, `UserWithdrawn`, `SessionRating`, `ProviderRewardAccrued`, `ProviderDebtPaid` |
| `DelegateStaking` | `SessionOpened`, `SessionClosed`, `SessionRating`, `ProviderRewardAccrued`, `DelegateStakingPoolDrawn`, `DelegateStakingPoolReturned`, `DelegateStakingPoolLockCreated`, `DelegateStakingPoolLockReleased`, `DelegateStakingWithdrawalPaid`, `DelegateStakingPoolSessionSettled` |
| `DelegateStakingFunding` | All other `DelegateStaking*` events (grants, approvals, funding, withdrawals, locks, parameters) |
| `ProviderPayouts` | `ProviderPayoutTargetUpdated` (plus `ProviderRewardAccrued`/`ProviderDebtPaid` declared in its storage interface) |
| `Delegation` | `DelegationRegistryUpdated` |

Field-level descriptions are in the topic documents.

## Errors

### Sessions (`ISessionRouter`)

| Error | Raised when |
|---|---|
| `SessionApprovedForAnotherChainId` / `SessionReceiptForAnotherChainId` | `chainId` in the message is not this chain |
| `SessionApproveExpired` / `SessionReceiptExpired` | Message older than `SIGNATURE_TTL` |
| `SessionProviderSignatureMismatch` | Approval not signed by the provider, its delegate or its contract owner |
| `SessionDuplicateApproval` | Approval already used |
| `SessionApprovedForAnotherUser` | Pool open: approval names another user than the hot wallet |
| `SessionBidNotFound` | Bid missing or deleted |
| `SessionTooShort` | Session would last less than 5 minutes |
| `SessionStakeTooLow` | Direct pay: stake does not cover `duration * price` |
| `SessionAlreadyClosed` | Close or settle of a closed session |
| `SessionNotEndedOrNotExist` | `claimForProvider` on an unknown session |
| `SessionPoolSessionCannotBeClosed` | `closeSession` on a pool session |
| `SessionUserAmountToWithdrawIsZero` | Nothing released to withdraw |
| `SessionPoolIndexOutOfBounds` | `setPoolConfig` with a missing index |
| `SessionMaxDurationTooShort` | `setMaxSessionDuration` below 5 minutes |

`SessionProviderNothingToClaimInThisPeriod` is declared but no longer raised.

### Delegate staking (`IDelegateStakingStorage`)

| Error | Raised when |
|---|---|
| `DelegateStakingInvalidAddress` | Zero funder in `approvePoolFunder` / `revokePoolFunderApproval`, zero hot wallet in `grantStakingAllowance` (other calls do not check) |
| `DelegateStakingInvalidAmount` | Zero amount |
| `DelegateStakingInvalidExpiry` | Grant expiry not in the future |
| `DelegateStakingCapBelowLifetimeFunding` | New cap below what was already funded |
| `DelegateStakingAllowanceNotActive` | Funding or revoking a grant that is not live |
| `DelegateStakingPoolFunderNotApproved` | Funding without the hot wallet's approval |
| `DelegateStakingFundingCapExceeded` | Funding above the cap |
| `DelegateStakingPrincipalTooLow` | Principal would end below `minPrincipal` (and not 0) |
| `DelegateStakingActiveFundersLimitExceeded` | Active list full even after purging |
| `DelegateStakingInsufficientPrincipal` | Withdrawal above principal |
| `DelegateStakingInsufficientLiquidity` | Draw above `free − pendingTotal` |
| `DelegateStakingInsufficientCapacity` | Funders' undrawn principal does not cover the draw |
| `DelegateStakingWithdrawalNotFound` | Cancel without a queued withdrawal |
| `DelegateStakingInvalidWithdrawalCursor` | Queue view cursor not in the queue |
| `DelegateStakingLimitExceeded` | Batch size, page size or parameter above its hard limit |
| `DelegateStakingPoolSessionNotFound` | Pool call on a session that is not a pool session |
| `DelegateStakingPoolSessionAlreadyExists` / `DelegateStakingSessionSlotNotEmpty` | Session ID already used (defensive) |
| `DelegateStakingSettlementNotReady` | Settle before `endsAt + settlementGrace` |
| `DelegateStakingInvalidPoolSettlement` | Internal consistency check failed (should never happen) |

### Other

| Error | Facet | Raised when |
|---|---|---|
| `InsufficientRightsForOperation(delegator, delegatee)` | all | Caller is not the principal and has no right |
| `OwnableUnauthorizedAccount(account)` | all | Owner-only call by another account |
| `ProviderPayoutTargetInvalid` | `ProviderPayouts` | Target is the diamond |
| `ProviderStakeTooLow`, `ProviderNotFound`, `ProviderHasActiveBids`, `ProviderHasAlreadyDeregistered` | `ProviderRegistry` | See [registry-and-marketplace.md](registry-and-marketplace.md) |
| `ModelStakeTooLow`, `ModelHasActiveBids`, `ModelHasAlreadyDeregistered` | `ModelRegistry` | 〃 |
| `MarketplaceProviderNotFound`, `MarketplaceModelNotFound`, `MarketplaceActiveBidNotFound`, `MarketplaceBidPricePerSecondInvalid`, `MarketplaceBidMinPricePerSecondIsZero`, `MarketplaceBidMinPricePerSecondIsInvalid`, `MarketplaceFeeAmountIsZero` | `Marketplace` | 〃 |

Declared but not raised: `ProviderNotDeregistered`, `ProviderNoStake`,
`ProviderNothingToWithdraw`, `ModelNotFound`. The event `ProviderWithdrawn` is declared but
never emitted; deregistration emits only `ProviderDeregistered`.
