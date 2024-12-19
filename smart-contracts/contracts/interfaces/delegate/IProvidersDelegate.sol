// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IProvidersDelegate {
    error InvalidNameLength();
    error InvalidEndpointLength();
    error InvalidFeeTreasuryAddress();
    error InvalidFee(uint256 current, uint256 max);
    error StakeClosed();
    error ProviderDeregistered();
    error BidCannotBeCreatedDuringThisPeriod();
    error InsufficientAmount();
    error RestakeDisabled(address staker);
    error RestakeInvalidCaller(address caller, address staker);
    error ClaimAmountIsZero();

    /**
     * The event that is emitted when the name updated.
     * @param name The new value.
     */
    event NameUpdated(string name);

    /**
     * The event that is emitted when the endpoint updated.
     * @param endpoint The new value.
     */
    event EndpointUpdated(string endpoint);

    /**
     * The event that is emitted when the `feeTreasury` updated.
     * @param feeTreasury The new value.
     */
    event FeeTreasuryUpdated(address feeTreasury);

    /**
     * The event that is emitted when the stake closed or opened for all users.
     * @param isStakeClosed The new value.
     */
    event IsStakeClosedUpdated(bool isStakeClosed);

    /**
     * The event that is emitted when restake was disabled or enabled by user.
     * @param staker The user address.
     * @param isRestakeDisabled The new value.
     */
    event IsRestakeDisabledUpdated(address staker, bool isRestakeDisabled);

    /**
     * The event that is emitted when user staked.
     * @param staker The user address.
     * @param staked The total staked amount for user.
     * @param totalStaked The total staked amount for the contract.
     * @param rate The contract rate.
     */
    event Staked(address staker, uint256 staked, uint256 totalStaked, uint256 rate);

    /**
     * The event that is emitted when user claimed.
     * @param staker The user address.
     * @param claimed The total claimed amount for user.
     * @param rate The contract rate.
     */
    event Claimed(address staker, uint256 claimed, uint256 rate);

    /**
     * The event that is emitted when rewards claimed for owner.
     * @param feeTreasury The fee treasury address.
     * @param feeAmount The fee amount.
     */
    event FeeClaimed(address feeTreasury, uint256 feeAmount);



    /**
     * @param staked Staked amount.
     * @param claimed Claimed amount.
     * @param rate The user internal rate.
     * @param pendingRewards Pending rewards for claim.
     * @param isRestakeDisabled If true, restake isn't available.
     */
    struct Staker {
        uint256 staked;
        uint256 claimed;
        uint256 rate;
        uint256 pendingRewards;
        bool isRestakeDisabled;
    }

    /**
     * The function to initialize the contract.
     * @param lumerinDiamond_ The Lumerin protocol address.
     * @param feeTreasury_ The subnet fee treasury.
     * @param fee_ The fee percent where 100% = 10^25.
     * @param name_ The Subnet name.
     * @param endpoint_ The subnet endpoint.
     * @param deregistrationOpensAt_ Provider deregistration will be available after this time.
     */
    function ProvidersDelegate_init(
        address lumerinDiamond_,
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_,
        uint128 deregistrationOpensAt_
    ) external;

    /**
     * The function to set the Subnet name.
     * @param name_ New name.
     */
    function setName(string memory name_) external;

    /**
     * The function to set the new endpoint.
     * @param endpoint_ New endpoint.
     */
    function setEndpoint(string memory endpoint_) external;

    /**
     * The function to set fee treasury address.
     * @param feeTreasury_ New address
     */
    function setFeeTreasury(address feeTreasury_) external;

    /**
     * The function close or open possibility to stake new tokens.
     * @param isStakeClosed_ True or False.
     */
    function setIsStakeClosed(bool isStakeClosed_) external;

    /**
     * The function to disabled possibility for restake.
     * @param isRestakeDisabled_ True or False.
     */
    function setIsRestakeDisabled(bool isRestakeDisabled_) external;

    /**
     * The function to stake tokens.
     * @param amount_ Amount to stake.
     */
    function stake(uint256 amount_) external;

    /**
     * The function to restake rewards.
     * @param staker_ Staker address.
     * @param amount_ Amount to stake.
     */
    function restake(address staker_, uint256 amount_) external;

    /**
     * The function to claim rewards.
     * @param staker_ Staker address.
     * @param amount_ Amount to stake.
     */
    function claim(address staker_, uint256 amount_) external returns(uint256);

    /**
     * The function to return the current rate.
     */
    function getCurrentRate() external view returns (uint256, uint256);

    /**
     * The function to get amount of the Staker rewards.
     * @param staker_ Staker address.
     */
    function getCurrentStakerRewards(address staker_) external view returns (uint256);

    /**
     * The function to deregister the provider.
     * @param bidIds_ Bid IDs.
     */
    function providerDeregister(bytes32[] calldata bidIds_) external;

    /**
     * The function create the model bid.
     * @param modelId_ Model ID.
     * @param pricePerSecond_ Price per second.
     */
    function postModelBid(bytes32 modelId_, uint256 pricePerSecond_) external returns (bytes32);

    /**
     * The function to delete model bids.
     * @param bidIds_ Bid IDs.
     */
    function deleteModelBids(bytes32[] calldata bidIds_) external;

    /**
     * The function to get contract version.
     */
    function version() external pure returns (uint256);
}
