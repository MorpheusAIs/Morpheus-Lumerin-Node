// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IProvidersDelegator {
    event NameUpdated(string name);
    event EndpointUpdated(string endpoint);
    event FeeUpdated(uint256 fee, address feeTreasury);
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

    struct Staker {
        uint256 staked;
        uint256 rate;
        uint256 pendingRewards;
        bool isRestakeDisabled;
    }

    function ProvidersDelegator_init(
        address lumerinDiamond_,
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_
    ) external;

    function setName(string memory name_) external;

    function setEndpoint(string memory endpoint_) external;

    function setFee(address feeTreasury_, uint256 fee_) external;

    function setIsStakeClosed(bool isStakeClosed_) external;

    function setIsRestakeDisabled(bool isRestakeDisabled_) external;

    function stake(uint256 amount_) external;

    function restake(address staker_, uint256 amount_) external;

    function claim(address staker_, uint256 amount_) external;

    function getCurrentRate() external view returns (uint256, uint256);

    function getCurrentStakerRewards(address staker_) external view returns (uint256);

    function providerDeregister() external;

    function postModelBid(
        bytes32 modelId_,
        uint256 pricePerSecond_
    ) external returns (bytes32);

    function deleteModelBid(bytes32 bidId_) external;

    function version() external pure returns (uint256);
}
