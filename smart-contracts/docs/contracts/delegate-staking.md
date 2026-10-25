# Delegate staking pools

Facets: `DelegateStaking` (pool sessions) and `DelegateStakingFunding` (grants, funding,
withdrawals, maintenance, parameters, views). Shared logic: `DelegateStakingPool`,
`SessionSettlement`. Storage: `DelegateStakingStorage` (slot
`diamond.standard.delegate.staking.storage`).

A pool lets a consumer node's **hot wallet** open stake sessions with MOR that belongs to
**cold wallets** (funders). The hot wallet never holds the MOR, and nothing in a pool can be
paid to anyone except the funder it belongs to. Funder-facing guarantees and risks are in
[delegate-staking-funder-economics.md](../delegate-staking-funder-economics.md).

## Model

There is one pool per hot wallet (`pools[hot]`). A pool's money is fungible and is tracked
in four aggregates:

| Aggregate | Meaning |
|---|---|
| `free` | Idle MOR, available for withdrawals and new draws |
| `inSession` | MOR currently staked in open pool sessions |
| `locked` | MOR from closed sessions waiting for the next UTC day (day-lock) |
| `pendingTotal` | Queued withdrawals not paid yet |

Each funder has a `Grant` in the pool:

| Field | Meaning |
|---|---|
| `cumulativeFundingCap` | Lifetime cap on what the funder may put in |
| `lifetimeFunded` | Total funded so far (never decreases) |
| `currentPrincipal` | The funder's current share of the pool |
| `exposure` | Part of the principal drawn into open sessions |
| `lockedExposure` | Funder's share of the unreleased day lock of closed sessions it financed; counts against its capacity until the next UTC midnight |
| `pending` | Queued withdrawal amount |
| `expiry`, `revoked` | Grant liveness: live = not revoked and `expiry > now` |
| `approved` | Hot wallet's approval of this funder |

Invariants (checked by the randomized tests after every step):

```
Σ currentPrincipal + pendingTotal == free + inSession + locked
Σ exposure                        == inSession
Σ lockedExposure                  <= locked     (equal until the bucket matures)
lifetimeFunded − Σ paid withdrawals == currentPrincipal + pending      (per funder)
unreleased lock buckets           <= 1
```

## Consent: approval and grant

Both sides must agree before money can be drawn:

| Step | Call | Caller |
|---|---|---|
| Hot approves the funder | `approvePoolFunder(hot, funder)` | Hot or its delegate with `delegation.rules.delegate.staking` |
| Funder sets a grant | `grantStakingAllowance(hot, cap, expiry)` | The funder itself (`msg.sender`) |
| Funder adds MOR | `fundStakingAllowance(hot, amount)` | The funder itself |

- `grantStakingAllowance` creates or updates the grant. `cap` must be at least
  `lifetimeFunded`, `expiry` must be in the future; calling it again clears `revoked`.
- `revokeStakingAllowance(hot)` (funder) and `revokePoolFunderApproval(hot, funder)` (hot
  side) stop further draws and top-ups at once. Neither blocks the funder's withdrawal.
- The hot wallet may fund its own pool (**self-escrow**): it needs its own grant; approval is
  implicit. Self-escrow is drawn last and the hot wallet can withdraw only its own principal.

## Funding

`fundStakingAllowance(hot, amount)` requires: `amount > 0`, approved funder, live grant,
`lifetimeFunded + amount <= cap`, and `currentPrincipal + amount >= minPrincipal`. It pulls
MOR from the funder, adds it to `currentPrincipal` and `free`, records the funder in
`allFunders` on first funding, admits it to the active list (not for self-escrow), and
serves queued withdrawals (newly free MOR goes to the queue first).

## Active funder list (draw order)

Funders with an approved, live grant and non-zero principal sit in a FIFO linked list
(self-escrow is not in the list).

- New entries go to the tail. Updating a live grant keeps the place.
- A funder whose principal drops to 0, or whose grant or approval is revoked, is unlinked at
  once. An expired grant stays linked until a draw or a purge inspects it, so the list can
  temporarily hold inactive entries (`getPoolFunders` shows the raw `inActiveList` flag).
  A funder that was unlinked, or that renews an expired grant, re-enters at the tail.
- Size limit: `maxActiveFunders` (default 64, hard limit 128). When the list is full,
  admission first inspects up to 128 entries from the head and unlinks the inactive ones;
  if the list is still full it reverts with `DelegateStakingActiveFundersLimitExceeded`.
- Lowering `maxActiveFunders` below the current size only blocks new admissions.
- `purgeInactiveFunders(hot, maxEntries ≤ 128)` (anyone) removes inactive entries starting
  from a stored cursor.

## Pool sessions

### Opening

```solidity
openSessionFromPool(address hot_, uint256 amount_, bytes approval_, bytes signature_) returns (bytes32)
```

- Caller: the hot wallet or its delegate with `delegation.rules.session`.
- The provider approval must name `hot_` as the user (`SessionApprovedForAnotherUser`).
  Signature, TTL, single use, bid and duration rules are those of a regular session
  ([sessions.md](sessions.md)); pool sessions are always stake mode (no direct pay).
- Draw:
  1. Release matured lock buckets.
  2. Liquidity: `amount <= free − pendingTotal`. Queued withdrawals come first; otherwise
     `DelegateStakingInsufficientLiquidity`.
  3. Walk the active list from the head (at most 128 nodes), unlinking inactive entries,
     and take `min(max(0, currentPrincipal − exposure − lockedExposure), remaining)` from each
     funder (after a withdrawal `exposure + lockedExposure` may exceed the principal).
  4. Take the rest from self-escrow.
  5. If something is still missing: `DelegateStakingInsufficientCapacity`.
- Each draw is stored with the session (`getPoolSession`) and emitted as
  `DelegateStakingPoolDrawn(sessionId, hot, funder, amount)`. A session has at most 129
  draws (128 funders + self-escrow).
- `free −= amount`, `inSession += amount`; queued withdrawals are served afterwards.

### Closing

```solidity
closePoolSession(bytes receipt_, bytes signature_)
```

- Caller: the hot wallet or its session delegate. Regular sessions are rejected here, and
  `SessionRouter.closeSession` rejects pool sessions.
- The user settlement is the regular one: the part of the stake used on the last day is
  locked until the next UTC midnight, the rest is free at once. In pool terms:
  `inSession −= stake`, `free += stake − lock`, a lock bucket gets `lock`.
- Every draw is returned to its funder's exposure (`DelegateStakingPoolReturned`). Principal
  does not change: sessions never spend funder money.
- The lock is split across the draws pro rata to the drawn amounts (the shares add up to
  the lock exactly) and kept as each funder's `lockedExposure` until the lock date. Until
  then that part of the principal cannot be drawn again, so a new session cannot be backed
  by idle MOR of funders who left the pool.
- Provider reward, dispute hold, stats and `SessionRating` are exactly as for a regular
  close; the reward is paid from the funding account or becomes debt
  ([provider-payouts.md](provider-payouts.md)).

### Settling an expired session (dead hot wallet)

```solidity
settleExpiredSession(bytes32 sessionId_)   // anyone
```

- Available once `block.timestamp >= endsAt + settlementGrace` for an open pool session.
- Returns the stake to the pool like a close (with `sessionEnd = endsAt`), marks the session
  closed and `settled`, and accrues the provider reward as **debt** (no transfer), up to
  what the reward limiter allows; a capped remainder stays claimable later. It does not
  update stats and emits no `SessionRating`.
- Events: `SessionClosed(hot, sessionId, provider)`,
  `DelegateStakingPoolSessionSettled(sessionId, hot, provider, caller)`.
- The provider collects later with `claimForProvider`.

## Day locks

The locked part of a closed session goes into a bucket keyed by its release time (the next
UTC midnight after the session end). Buckets are released lazily: every draw, close,
settlement, funding and withdrawal request first releases matured buckets (up to
`maxAutoReleaseDays`), and anyone can call `releasePoolLocks(hot, maxBuckets ≤ 32)`.

Because every lock is dated the next UTC day and every operation that creates a lock first
releases matured ones, a pool never has more than one unreleased bucket. The bucket queue
and its limits are a safety margin.

## Withdrawals

```solidity
requestStakingWithdrawal(address hot_, address funder_, uint256 amount_)
cancelStakingWithdrawal(address hot_, address funder_)
servePoolWithdrawals(address hot_, uint256 maxEntries_)   // anyone, ≤ 32
```

- Request and cancel: the funder or its delegate with `delegation.rules.delegate.staking`
  (a full delegation passes too, see [delegation.md](delegation.md)). Money always goes to
  the funder.
- No live grant or approval is needed: a revoked or expired funder can always leave.
- `amount <= currentPrincipal`, and what remains must be 0 or at least `minPrincipal`
  (a full exit is always allowed).
- Principal decreases at once. If the queue is empty, `min(amount, free)` is paid
  immediately; the rest is queued. If the queue is not empty, everything is queued behind
  it. One queue entry per funder; a repeated request adds to it and moves the whole entry,
  including its earlier part, to the queue tail, so later money waits behind earlier
  requests of other funders. To keep its place, a funder waits until the entry is paid
  before requesting more.
- The queue is served in FIFO order from `free`, paying the head partially when needed. It
  is served automatically (up to `maxAutoService` entries) after funding, draws, closes,
  settlements, lock releases and requests, and by `servePoolWithdrawals`.
- **Last out waits.** Liquidity is shared: if others withdrew first, a funder may wait for
  sessions to close and locks to release even if "its" money is not in a session. Other
  funders are never locked pro rata.
- Cancel returns the whole pending amount to principal (subject to `minPrincipal` while the
  funder is approved and live). If the funder is not the hot wallet, is approved and live, and
  is not linked, it is admitted at the tail of the active list. If that list is full (for
  example after a full withdrawal freed the slot for another funder, or after the owner
  lowered `maxActiveFunders`), the cancel reverts with `DelegateStakingActiveFundersLimitExceeded`
  and the withdrawal stays queued and is paid as usual.

## Permissionless maintenance

| Call | Effect | Bound |
|---|---|---|
| `settleExpiredSession(sessionId)` | Close an expired pool session, provider reward → debt | One session, ≤ 129 draws |
| `servePoolWithdrawals(hot, n)` | Pay queued withdrawals | n ≤ 32 |
| `releasePoolLocks(hot, n)` | Release matured buckets, then serve the queue | n ≤ 32 |
| `purgeInactiveFunders(hot, n)` | Unlink inactive funders | n ≤ 128 |
| `claimForProvider(sessionId)` | Accrue session reward and pay provider debt | One session |

The caller never receives funds. With the hot wallet, the provider and the owner all
unavailable, funders still recover their capital through these calls.

## Parameters

Owner setters in `DelegateStakingFunding`. Zero restores the default; values above the hard
limit revert with `DelegateStakingLimitExceeded`.

| Parameter | Setter | Default | Hard limit |
|---|---|---|---|
| Active funder list size | `setMaxActiveFunders` | 64 | 128 |
| Minimum principal | `setMinPrincipal` | 1 MOR | 1,000,000,000 MOR |
| Grace before settlement | `setSettlementGrace` | 1 day | 7 days |
| Withdrawals served automatically | `setMaxAutoService` | 5 | 32 |
| Lock buckets released automatically | `setMaxAutoReleaseDays` | 8 | 32 |

`getDelegateStakingParameters()` returns the effective values, `getDelegateStakingLimits()`
the hard limits.

## Views

| View | Returns |
|---|---|
| `getStakingAllowance(hot, funder)` | Grant fields plus `active` (live) and `approved` |
| `isPoolFunderApproved(hot, funder)` | Approval, including implicit self-escrow |
| `getAvailablePoolStake(hot)` | What the next `openSessionFromPool` can draw: `min(free + matured buckets − pendingTotal, capacity)` |
| `getPoolBalances(hot)` | `free`, `inSession`, `locked`, `pendingTotal` |
| `getPoolFunders(hot, offset, limit ≤ 128)` | Page of every funder that ever funded, with grant and `inActiveList` |
| `getStakingWithdrawal(hot, funder)` | Pending amount |
| `getPoolWithdrawalQueue(hot, cursor, limit ≤ 128)` | Queue page in service order, next cursor, total |
| `getPoolLockBuckets(hot, offset, limit ≤ 32)` | Unreleased buckets and their count |
| `getPoolSession(sessionId)` | Hot wallet, draws (while open), `exists`, `closed`, `settled` |
| `getDelegateStakingDelegationRight()` | `keccak256("delegation.rules.delegate.staking")` |

## Gas at the limits

Measured in `test/diamond/ProtocolProperties.test.ts`; every case stays below the 10M test
budget.

| Case | Gas |
|---|---|
| Open with 129 draws | 5.24M |
| Close that session | 1.79M |
| Settle that session | 1.63M |
| Admit a funder after purging 128 expired grants | 1.77M |
| Serve 32 queued withdrawals | 1.27M |

## Events

| Event | When |
|---|---|
| `DelegateStakingPoolFunderApproved(hot, funder)` / `…ApprovalRevoked` | Hot-side approval changes |
| `DelegateStakingAllowanceGranted(hot, funder, cap, expiry)` / `…Revoked(hot, funder)` | Grant changes |
| `DelegateStakingAllowanceFunded(hot, funder, amount)` | Funding |
| `DelegateStakingPoolDrawn(sessionId, hot, funder, amount)` | Per draw at open |
| `DelegateStakingPoolReturned(sessionId, hot, funder, amount)` | Per draw at close or settlement |
| `DelegateStakingPoolLockCreated(hot, releaseAt, amount)` / `…LockReleased` | Day-lock buckets |
| `DelegateStakingWithdrawalRequested(hot, funder, amount, paid, pending)` | Request (pending = funder's total pending) |
| `DelegateStakingWithdrawalPaid(hot, funder, amount)` | Every payment, immediate or from the queue |
| `DelegateStakingWithdrawalCancelled(hot, funder, amount)` | Cancel |
| `DelegateStakingPoolSessionSettled(sessionId, hot, provider, caller)` | Expired-session settlement |
| `DelegateStaking…Updated(value)` | Parameter setters |

Pool sessions also emit the regular `SessionOpened`, `SessionClosed`, `SessionRating` and
`ProviderRewardAccrued`; `ProviderDebtPaid` comes from a later `claimForProvider`.
