# Delegate staking pool — funder economics (v1)

Covers RFP §3.4.1 COLDC-R19 / AC-COLDC-10. Applies to the `DelegateStaking` and
`DelegateStakingFunding` facets.

## Who funds a pool

A pool belongs to one hot wallet (the consumer node that opens sessions). Cold wallets
(funders) put MOR into that pool so the hot wallet can stake it in sessions without ever
holding it.

- **Funders are operator-affiliated.** v1 assumes the funders are the node operator's own
  cold wallets or parties compensated off-chain. The contract has no notion of a market
  between unrelated funders and hot wallets.
- **No on-chain yield.** Funding a pool earns nothing on-chain. Provider rewards are paid
  from the protocol funding account, never from the pool, and the pool receives nothing
  back except its own principal.
- **Admission is controlled by the hot wallet.** Only funders approved by the hot wallet
  (or its delegate with the `delegation.rules.delegate.staking` right, including a full
  delegation) can add money. Revoking the approval stops further draws and top-ups; the
  funder can still withdraw.

## What a funder can rely on

- Principal is never sent to the hot wallet or to anyone except the funder itself (or a
  session escrow inside the diamond while a session is open). There is no path from another
  funder's principal to the hot wallet's balance; the hot wallet can only withdraw its own
  self-escrow.
- Everything a session draws returns to the pool in full: the unused part at close, the
  used part after the day-lock (the next UTC day). Funder principal does not change
  because of sessions.
- Withdrawals never need the hot wallet, the provider, or the owner:
  `requestStakingWithdrawal` by the funder (or its delegate), and permissionless
  `servePoolWithdrawals`, `releasePoolLocks` and `settleExpiredSession` for a dead node.
  Someone has to send these transactions; the contract does not act on its own.
- Owner parameters are bounded by hard constants. They cannot prevent a full exit, but
  they can delay it (`settlementGrace`, up to 7 days) or constrain partial withdrawals
  (`minPrincipal`: a partial withdrawal must leave at least the minimum, withdrawing
  everything is always allowed).

## Shared liquidity (disclosure)

Money in a pool is fungible. A funder owns a share (`currentPrincipal`), not specific
tokens.

- If no withdrawal is queued, a withdrawal is paid at once from the pool's free balance and
  any shortfall is queued; if others are already queued, the whole amount joins the queue
  behind them (one queue entry per funder; a repeated request moves the whole entry,
  including the earlier part, to the tail). The queue is paid in order as sessions close
  and day-locks release.
- **Last out waits.** If another funder withdrew first, your withdrawal may wait even if
  "your" part of the pool is not in a session. Other funders are never locked pro-rata.
- Pending withdrawals take priority over new session draws: the hot wallet cannot draw
  money that is needed to pay the queue.
- The hot wallet may fund its own pool (self-escrow). Self-escrow is drawn last, and the
  hot wallet can withdraw only up to its own recorded principal; because liquidity is
  shared, those tokens may physically be the same tokens other funders deposited.

## Limits and risks

- Draw order is FIFO by admission: older funders are drawn first, self-escrow last. A
  funder whose principal drops to zero or whose grant expires loses its place and goes to
  the tail when it is active again.
- The active funder list is capped (`maxActiveFunders`, default 64, hard limit 128).
  Revoked grants and approvals leave the list at once. Expired grants are unlinked by
  draws, by admission when the list is full, and by the permissionless
  `purgeInactiveFunders`.
- With a dead hot wallet, capital becomes permissionlessly recoverable after the session
  end (`endsAt`, at most `maxSessionDuration` after opening) plus `settlementGrace`
  (default 1 day, hard limit 7 days), once someone calls `settleExpiredSession` and the
  funder or a keeper serves the withdrawal. With a grace of one day or more, settlement
  returns the whole stake to the free balance at once; a day-lock applies only if the owner
  sets the grace below one day.
- The contract assumes MOR is a standard ERC-20 (no transfer fees, hooks or blacklists).
- **Delegation.** A cold funder does not need to delegate anything to the hot wallet: it
  only grants, funds and withdraws, all from its own key. Do not give the hot wallet a full
  delegation or `delegation.rules.session`. With either, the hot wallet can open regular
  sessions in the funder's name that pull MOR from the funder's wallet up to its remaining
  allowance to the diamond (and, in direct-pay mode, pay it to a provider it controls); a
  full delegation also lets it cancel the funder's withdrawals. Approve the diamond for
  exactly the amount being funded. To let someone else manage withdrawals, delegate only
  `delegation.rules.delegate.staking`.
