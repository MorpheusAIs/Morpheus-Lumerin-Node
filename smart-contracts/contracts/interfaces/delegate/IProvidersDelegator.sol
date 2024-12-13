// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IProvidersDelegator {
    event NameUpdated(string name);
    event EndpointUpdated(string endpoint);
    event FeeTreasuryUpdated(address feeTreasury);
    event IsStakeClosedUpdated(bool isStakeClosed);
    event IsRestakeDisabledUpdated(address staker, bool isRestakeDisabled);
    event Staked(address staker, uint256 staked, uint256 pendingRewards, uint256 rate);
    event Restaked(address staker, uint256 staked, uint256 pendingRewards, uint256 rate);
    event Claimed(address staker, uint256 staked, uint256 pendingRewards, uint256 rate);
    event FeeClaimed(address feeTreasury, uint256 feeAmount);

    error InvalidNameLength();
    error InvalidEndpointLength();
    error InvalidFeeTreasuryAddress();
    error InvalidFee(uint256 current, uint256 max);
    error StakeClosed();
    error InsufficientAmount();
    error RestakeDisabled(address staker);
    error RestakeInvalidCaller(address caller, address staker);
    error ClaimAmountIsZero();

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
     * @param deregistrationTimeout_ Provider deregistration will be available after this timeout.
     * @param deregistrationNonFeePeriod_ Period after deregistration when Stakers can claim rewards without fee.
     */
    function ProvidersDelegator_init(
        address lumerinDiamond_,
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_,
        uint128 deregistrationTimeout_,
        uint128 deregistrationNonFeePeriod_
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
    function claim(address staker_, uint256 amount_) external;

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
