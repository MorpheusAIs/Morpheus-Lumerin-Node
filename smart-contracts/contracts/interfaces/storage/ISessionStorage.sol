// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface ISessionStorage {
    /**
     * The structure that stores the bid data.
     * @param user The user address. User opens the session
     * @param bidId The bid ID.
     * @param stake The stake amount.
     * @param closeoutReceipt The receipt from provider when session closed.
     * @param closeoutType The closeout type.
     * @param providerWithdrawnAmount Provider withdrawn amount fot this session.
     * @param openedAt The timestamp when the session is opened. Setted on creation.
     * @param endsAt The timestamp when the session is ends. Setted on creation.
     * @param closedAt The timestamp when the session is closed. Setted on close.
     * @param isActive The session status.
     * @param isDirectPaymentFromUser If active, user pay for provider from his stake.
     */
    struct Session {
        address user;
        bytes32 bidId;
        uint256 stake;
        bytes closeoutReceipt;
        uint256 closeoutType;
        uint256 providerWithdrawnAmount;
        uint128 openedAt;
        uint128 endsAt;
        uint128 closedAt;
        bool isActive;
        bool isDirectPaymentFromUser;
    }

    /**
     * The structure that stores information about locked user funds.
     * @param amount The locked amount.
     * @param releaseAt The timestampl when funds will be available.
     */
    struct OnHold {
        uint256 amount;
        // In epoch seconds. TODO: consider using hours to reduce storage cost
        uint128 releaseAt;
    }

    /**
     * The structure that stores the Pool data. Should be the same with 0x47176B2Af9885dC6C4575d4eFd63895f7Aaa4790 on the Eth mainnet
     */
    struct Pool {
        uint256 initialReward;
        uint256 rewardDecrease;
        uint128 payoutStart;
        uint128 decreaseInterval;
    }

    /**
     * The function returns the session structure.
     * @param sessionId_ Session ID
     */
    function getSession(bytes32 sessionId_) external view returns (Session memory);

    /**
     * The function returns the user session IDs.
     * @param user_ The user address
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getUserSessions(address user_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory);

    /**
     * The function returns the provider session IDs.
     * @param provider_ The provider address
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getProviderSessions(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory);

    /**
     * The function returns the model session IDs.
     * @param modelId_ The model ID
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getModelSessions(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory);

    /**
     * The function returns the pools info.
     */
    function getPools() external view returns (Pool[] memory);

    /**
     * The function returns the pools info.
     * @param index_ Pool index
     */
    function getPool(uint256 index_) external view returns (Pool memory);

    /**
     * The function returns the funcding (treasury) address for providers payments.
     */
    function getFundingAccount() external view returns (address);

    /**
     * The function returns total amount of sessions for the provider.
     * @param provider_ Provider address
     */
    function getTotalSessions(address provider_) external view returns (uint256);

    /**
     * The function returns total amount of claimed token by providers.
     */
    function getProvidersTotalClaimed() external view returns (uint256);

    /**
     * Check the approval for usage.
     * @param approval_ Approval from provider
     */
    function getIsProviderApprovalUsed(bytes memory approval_) external view returns (bool);

    /**
     * The function returns max session duration.
     */
    function getMaxSessionDuration() external view returns (uint128);
}
