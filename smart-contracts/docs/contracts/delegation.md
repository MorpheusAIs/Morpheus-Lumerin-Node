# Delegation

Facet: `Delegation`. Storage: `DelegationStorage` (slot
`diamond.standard.delegation.storage`). The diamond keeps no delegation data of its own; it
asks an external [delegate.xyz v2](https://docs.delegate.xyz) registry, set by the owner
with `setRegistry` (BASE: `0x00000000000000447e69651d841bD8D104Bed493`).

## How a check works

```solidity
_validateDelegatee(caller, principal, right)
// passes if caller == principal
// or registry.checkDelegateForContract(caller, principal, address(diamond), right)
```

The registry returns true for:

- a delegation with exactly that `right`, either for all contracts (`ALL`) or for the
  diamond (`CONTRACT`);
- a **full** delegation (empty rights) of type `ALL` or `CONTRACT` for the diamond.

Delegation never changes where money goes: tokens are pulled from and paid to the
principal (or, for provider rewards, the provider's payout target).

## Rights

| Right | `bytes32` | Lets a delegate of… |
|---|---|---|
| `DELEGATION_RULES_PROVIDER` | `keccak256("delegation.rules.provider")` | …a provider: `providerRegister`, `providerDeregister`, and sign approvals and receipts for it |
| `DELEGATION_RULES_MODEL` | `keccak256("delegation.rules.model")` | …a model owner: `modelRegister`, `modelDeregister` |
| `DELEGATION_RULES_MARKETPLACE` | `keccak256("delegation.rules.marketplace")` | …a provider: `postModelBid`, `deleteModelBid` |
| `DELEGATION_RULES_SESSION` | `keccak256("delegation.rules.session")` | …a user or hot wallet: `openSession`, `closeSession`, `withdrawUserStakes`, `openSessionFromPool`, `closePoolSession` |
| Delegate staking | `keccak256("delegation.rules.delegate.staking")` | …a hot wallet: `approvePoolFunder`, `revokePoolFunderApproval`; …a funder: `requestStakingWithdrawal`, `cancelStakingWithdrawal` |
| Payout | `keccak256("delegation.rules.payout")` | …a provider: `setPayoutTarget` (explicit right only, see below) |

The first four are exposed as constants on the `Delegation` facet. The delegate staking
right is returned by `getDelegateStakingDelegationRight()`, the payout right by
`DELEGATION_RULES_PAYOUT()`.

## Exceptions

- **Payout right.** `setPayoutTarget` does not use `checkDelegateForContract`. It reads the
  provider's outgoing delegations and accepts only an entry with exactly the payout right,
  of type `ALL` or `CONTRACT` for the diamond. A full delegation does not grant it, so
  handing a node operator full control does not let it redirect rewards.
- **Funder-only calls.** `grantStakingAllowance`, `revokeStakingAllowance` and
  `fundStakingAllowance` act for `msg.sender` and cannot be delegated.
- **Permissionless calls** need no right: `claimForProvider`, `settleExpiredSession`,
  `servePoolWithdrawals`, `releasePoolLocks`, `purgeInactiveFunders`.

## Full delegations of a hot wallet

A full delegation from a hot wallet lets the delegate open and close pool sessions and
approve or revoke funders. It cannot move other funders' MOR to the hot wallet or to
itself: every pool payment goes to the funder that owns it. The only pool money that can
reach the hot wallet is its own self-escrow principal, through the normal withdrawal path.
A full delegation from a funder lets the delegate request or cancel that funder's
withdrawals, which still pay the funder.

## Signer delegation for providers

Approvals and receipts are accepted from the provider, a delegate with
`DELEGATION_RULES_PROVIDER` (or a full delegation), or the `owner()` of a contract provider
(static call capped at 100,000 gas). This is how `ProvidersDelegate` proxies sign: their
owner signs.
