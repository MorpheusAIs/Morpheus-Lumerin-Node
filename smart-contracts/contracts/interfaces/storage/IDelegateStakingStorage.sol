// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IDelegateStakingStorage {
    /**
     * The structure that stores the funder grant data.
     * @param cumulativeFundingCap The cap on the total amount the funder may fund over the grant lifetime.
     * @param expiry The timestamp when the grant expires.
     * @param revoked If true, the grant is revoked by the funder.
     * @param lifetimeFunded The total amount funded, never decreases.
     * @param currentPrincipal The funder share of the pool.
     * @param exposure The part of the principal drawn into open sessions.
     * @param pending The queued withdrawal amount.
     * @param active If true, the grant is live (not revoked and not expired).
     * @param approved If true, the funder is approved by the hot wallet (always true for self-escrow).
     */
    struct StakingAllowance {
        uint256 cumulativeFundingCap;
        uint128 expiry;
        bool revoked;
        uint256 lifetimeFunded;
        uint256 currentPrincipal;
        uint256 exposure;
        uint256 pending;
        bool active;
        bool approved;
    }

    /**
     * The structure that stores the pool balances.
     * @param free The idle amount, available for withdrawals and new draws.
     * @param inSession The amount staked in open pool sessions.
     * @param locked The amount day-locked after session close.
     * @param pendingTotal The total queued withdrawal amount.
     */
    struct PoolBalances {
        uint256 free;
        uint256 inSession;
        uint256 locked;
        uint256 pendingTotal;
    }

    /**
     * The structure that stores the pool funder data.
     * @param funder The funder address.
     * @param allowance The funder grant and accounting state.
     * @param inActiveList If true, the funder is linked in the active funder list.
     */
    struct PoolFunder {
        address funder;
        StakingAllowance allowance;
        bool inActiveList;
    }

    /**
     * The structure that stores the queued withdrawal data.
     * @param funder The funder address.
     * @param amount The pending amount.
     */
    struct PoolWithdrawal {
        address funder;
        uint256 amount;
    }

    /**
     * The structure that stores the lock bucket data.
     * @param releaseAt The timestamp when the bucket is released.
     * @param amount The locked amount.
     */
    struct PoolLockBucket {
        uint128 releaseAt;
        uint256 amount;
    }

    /**
     * The structure that stores the session draw data.
     * @param funder The funder address, the hot wallet for self-escrow.
     * @param amount The drawn amount.
     */
    struct PoolSessionDraw {
        address funder;
        uint256 amount;
    }

    /**
     * The structure that stores the pool session data.
     * @param hot The hot wallet whose pool backs the session.
     * @param draws The funding draws, empty after close or settlement.
     * @param exists If true, the session is a pool session.
     * @param closed If true, the stake is returned to the pool.
     * @param settled If true, the session is settled by `settleExpiredSession`.
     */
    struct PoolSession {
        address hot;
        PoolSessionDraw[] draws;
        bool exists;
        bool closed;
        bool settled;
    }

    /**
     * The structure that stores the effective delegate staking parameters.
     * @param maxActiveFunders The active funder list size limit.
     * @param minPrincipal The minimum funder principal.
     * @param settlementGrace The period after the session end before settlement is allowed.
     * @param maxAutoService The number of withdrawals served automatically.
     * @param maxAutoReleaseDays The number of lock buckets released automatically.
     */
    struct DelegateStakingParameters {
        uint256 maxActiveFunders;
        uint256 minPrincipal;
        uint256 settlementGrace;
        uint256 maxAutoService;
        uint256 maxAutoReleaseDays;
    }

    /**
     * The structure that stores the hard limits for delegate staking parameters.
     * @param maxActiveFunders The max active funder list size.
     * @param maxAutoService The max number of withdrawals served in one call.
     * @param maxAutoReleaseDays The max number of lock buckets released in one call.
     * @param maxSettlementGrace The max settlement grace.
     * @param maxMinPrincipal The max minimum principal.
     */
    struct DelegateStakingLimits {
        uint256 maxActiveFunders;
        uint256 maxAutoService;
        uint256 maxAutoReleaseDays;
        uint256 maxSettlementGrace;
        uint256 maxMinPrincipal;
    }

    error DelegateStakingInvalidAddress();
    error DelegateStakingInvalidAmount();
    error DelegateStakingInvalidExpiry();
    error DelegateStakingAllowanceNotActive();
    error DelegateStakingPoolFunderNotApproved();
    error DelegateStakingCapBelowLifetimeFunding();
    error DelegateStakingFundingCapExceeded();
    error DelegateStakingPrincipalTooLow();
    error DelegateStakingActiveFundersLimitExceeded();
    error DelegateStakingInsufficientPrincipal();
    error DelegateStakingInsufficientLiquidity();
    error DelegateStakingInsufficientCapacity();
    error DelegateStakingWithdrawalNotFound();
    error DelegateStakingInvalidWithdrawalCursor();
    error DelegateStakingLimitExceeded();
    error DelegateStakingPoolSessionAlreadyExists();
    error DelegateStakingPoolSessionNotFound();
    error DelegateStakingInvalidPoolSettlement();
    error DelegateStakingSessionSlotNotEmpty();
    error DelegateStakingSettlementNotReady();

    /**
     * The event that is emitted when the funder grant created or updated.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     * @param cumulativeFundingCap The new cumulative funding cap.
     * @param expiry The new expiry timestamp.
     */
    event DelegateStakingAllowanceGranted(
        address indexed hot,
        address indexed funder,
        uint256 cumulativeFundingCap,
        uint128 expiry
    );

    /**
     * The event that is emitted when the funder grant revoked.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     */
    event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder);

    /**
     * The event that is emitted when the funder funded the pool.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     * @param amount The funded amount.
     */
    event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount);

    /**
     * The event that is emitted when the hot wallet approved the funder.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     */
    event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder);

    /**
     * The event that is emitted when the hot wallet revoked the funder approval.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     */
    event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder);

    /**
     * The event that is emitted when the session drew stake from the funder.
     * @param sessionId The session ID.
     * @param hot The hot wallet address.
     * @param funder The funder address, the hot wallet for self-escrow.
     * @param amount The drawn amount.
     */
    event DelegateStakingPoolDrawn(
        bytes32 indexed sessionId,
        address indexed hot,
        address indexed funder,
        uint256 amount
    );

    /**
     * The event that is emitted when the session draw returned to the funder exposure.
     * @param sessionId The session ID.
     * @param hot The hot wallet address.
     * @param funder The funder address, the hot wallet for self-escrow.
     * @param amount The returned amount.
     */
    event DelegateStakingPoolReturned(
        bytes32 indexed sessionId,
        address indexed hot,
        address indexed funder,
        uint256 amount
    );

    /**
     * The event that is emitted when the withdrawal requested.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     * @param amount The requested amount.
     * @param paid The amount paid immediately.
     * @param pending The total pending amount of the funder after the request.
     */
    event DelegateStakingWithdrawalRequested(
        address indexed hot,
        address indexed funder,
        uint256 amount,
        uint256 paid,
        uint256 pending
    );

    /**
     * The event that is emitted when the pending withdrawal cancelled.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     * @param amount The amount returned to the principal.
     */
    event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount);

    /**
     * The event that is emitted when the withdrawal paid, immediately or from the queue.
     * @param hot The hot wallet address.
     * @param funder The funder address.
     * @param amount The paid amount.
     */
    event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount);

    /**
     * The event that is emitted when the session stake day-locked in the pool.
     * @param hot The hot wallet address.
     * @param releaseAt The bucket release timestamp.
     * @param amount The locked amount.
     */
    event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount);

    /**
     * The event that is emitted when the matured lock bucket released.
     * @param hot The hot wallet address.
     * @param releaseAt The bucket release timestamp.
     * @param amount The released amount.
     */
    event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount);

    /**
     * The event that is emitted when the `maxActiveFunders` updated.
     * @param value The new value, zero for the default.
     */
    event DelegateStakingMaxActiveFundersUpdated(uint256 value);

    /**
     * The event that is emitted when the `minPrincipal` updated.
     * @param value The new value, zero for the default.
     */
    event DelegateStakingMinPrincipalUpdated(uint256 value);

    /**
     * The event that is emitted when the `settlementGrace` updated.
     * @param value The new value, zero for the default.
     */
    event DelegateStakingSettlementGraceUpdated(uint256 value);

    /**
     * The event that is emitted when the `maxAutoService` updated.
     * @param value The new value, zero for the default.
     */
    event DelegateStakingMaxAutoServiceUpdated(uint256 value);

    /**
     * The event that is emitted when the `maxAutoReleaseDays` updated.
     * @param value The new value, zero for the default.
     */
    event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value);

    /**
     * The event that is emitted when the expired pool session settled.
     * @param sessionId The session ID.
     * @param hot The hot wallet address.
     * @param provider The provider address.
     * @param caller The caller address.
     */
    event DelegateStakingPoolSessionSettled(
        bytes32 indexed sessionId,
        address indexed hot,
        address indexed provider,
        address caller
    );
}
