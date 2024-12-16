// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";

import {PRECISION} from "@solarity/solidity-lib/utils/Globals.sol";

import {IProvidersDelegator} from "../interfaces/delegate/IProvidersDelegator.sol";
import {IBidStorage} from "../interfaces/storage/IBidStorage.sol";
import {IProviderRegistry} from "../interfaces/facets/IProviderRegistry.sol";
import {IMarketplace} from "../interfaces/facets/IMarketplace.sol";

contract ProvidersDelegator is IProvidersDelegator, OwnableUpgradeable {
    using SafeERC20 for IERC20;
    using Math for uint256;

    // Deps
    address public lumerinDiamond;
    address public token;

    // Fee
    address public feeTreasury;
    uint256 public fee;

    // Metadata
    string public name;
    string public endpoint;

    // Main calculation storage
    uint256 public totalStaked;
    uint256 public totalRate;
    uint256 public lastContractBalance;
    bool public isStakeClosed;
    mapping(address => Staker) public stakers;

    // Deregistration limits
    uint128 public deregistrationOpenAt;
    uint128 public deregistrationTimeout;
    uint128 public deregistrationNonFeeOpened;
    uint128 public deregistrationNonFeePeriod;

    constructor() {
        _disableInitializers();
    }

    function ProvidersDelegator_init(
        address lumerinDiamond_,
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_,
        uint128 deregistrationTimeout_,
        uint128 deregistrationNonFeePeriod_
    ) external initializer {
        __Ownable_init();

        lumerinDiamond = lumerinDiamond_;
        token = IBidStorage(lumerinDiamond).getToken();

        setName(name_);
        setEndpoint(endpoint_);
        setFeeTreasury(feeTreasury_);

        if (fee_ > PRECISION) {
            revert InvalidFee(fee_, PRECISION);
        }
        fee = fee_;

        IERC20(token).approve(lumerinDiamond_, type(uint256).max);

        deregistrationTimeout = deregistrationTimeout_;
        deregistrationOpenAt = uint128(block.timestamp) + deregistrationTimeout_;
        deregistrationNonFeePeriod = deregistrationNonFeePeriod_;
    }

    function setName(string memory name_) public onlyOwner {
        if (bytes(name_).length == 0) {
            revert InvalidNameLength();
        }

        name = name_;

        emit NameUpdated(name_);
    }

    function setEndpoint(string memory endpoint_) public onlyOwner {
        if (bytes(endpoint_).length == 0) {
            revert InvalidEndpointLength();
        }

        endpoint = endpoint_;

        emit EndpointUpdated(endpoint_);
    }

    function setFeeTreasury(address feeTreasury_) public onlyOwner {
        if (feeTreasury_ == address(0)) {
            revert InvalidFeeTreasuryAddress();
        }

        feeTreasury = feeTreasury_;

        emit FeeTreasuryUpdated(feeTreasury_);
    }

    function setIsStakeClosed(bool isStakeClosed_) public onlyOwner {
        isStakeClosed = isStakeClosed_;

        emit IsStakeClosedUpdated(isStakeClosed_);
    }

    function setIsRestakeDisabled(bool isRestakeDisabled_) external {
        stakers[_msgSender()].isRestakeDisabled = isRestakeDisabled_;

        emit IsRestakeDisabledUpdated(_msgSender(), isRestakeDisabled_);
    }

    function stake(uint256 amount_) external {
        if (isStakeClosed) {
            revert StakeClosed();
        }

        if (amount_ == 0) {
            revert InsufficientAmount();
        }

        address user_ = _msgSender();
        Staker storage staker = stakers[user_];

        (uint256 currentRate_, uint256 contractBalance_) = getCurrentRate();
        uint256 pendingRewards_ = _getCurrentStakerRewards(currentRate_, staker);

        IERC20(token).safeTransferFrom(user_, address(this), amount_);

        totalRate = currentRate_;
        totalStaked += amount_;

        lastContractBalance = contractBalance_;

        staker.rate = currentRate_;
        staker.staked += amount_;
        staker.pendingRewards = pendingRewards_;

        IProviderRegistry(lumerinDiamond).providerRegister(address(this), amount_, endpoint);

        emit Staked(user_, staker.staked, staker.pendingRewards, staker.rate);
    }

    function restake(address staker_, uint256 amount_) external {
        if (_msgSender() != staker_ && _msgSender() != owner()) {
            revert RestakeInvalidCaller(_msgSender(), staker_);
        }

        Staker storage staker = stakers[staker_];
        if (staker.isRestakeDisabled) {
            revert RestakeDisabled(staker_);
        }

        (uint256 currentRate_, uint256 contractBalance_) = getCurrentRate();
        uint256 pendingRewards_ = _getCurrentStakerRewards(currentRate_, staker);

        amount_ = amount_.min(contractBalance_).min(pendingRewards_);
        if (amount_ == 0) {
            revert InsufficientAmount();
        }

        totalRate = currentRate_;
        totalStaked += amount_;

        lastContractBalance = contractBalance_ - amount_;

        staker.rate = currentRate_;
        staker.staked += amount_;
        staker.claimed += amount_;
        staker.pendingRewards = pendingRewards_ - amount_;

        IProviderRegistry(lumerinDiamond).providerRegister(address(this), amount_, endpoint);

        emit Restaked(staker_, staker.staked, staker.pendingRewards, staker.rate);
    }

    function claim(address staker_, uint256 amount_) external {
        Staker storage staker = stakers[staker_];

        (uint256 currentRate_, uint256 contractBalance_) = getCurrentRate();
        uint256 pendingRewards_ = _getCurrentStakerRewards(currentRate_, staker);

        amount_ = amount_.min(contractBalance_).min(pendingRewards_);
        if (amount_ == 0) {
            revert ClaimAmountIsZero();
        }

        totalRate = currentRate_;

        lastContractBalance = contractBalance_ - amount_;

        staker.rate = currentRate_;
        staker.pendingRewards = pendingRewards_ - amount_;
        staker.claimed += amount_;

        uint256 feeAmount_ = (amount_ * fee) / PRECISION;
        if (feeAmount_ != 0 && block.timestamp > deregistrationNonFeeOpened + deregistrationNonFeePeriod) {
            IERC20(token).safeTransfer(feeTreasury, feeAmount_);

            amount_ -= feeAmount_;

            emit FeeClaimed(feeTreasury, feeAmount_);
        }

        IERC20(token).safeTransfer(staker_, amount_);

        emit Claimed(staker_, staker.staked, staker.pendingRewards, staker.rate);
    }

    function getCurrentRate() public view returns (uint256, uint256) {
        uint256 contractBalance_ = IERC20(token).balanceOf(address(this));

        if (totalStaked == 0) {
            return (totalRate, contractBalance_);
        }

        uint256 reward_ = contractBalance_ - lastContractBalance;
        uint256 rate_ = totalRate + (reward_ * PRECISION) / totalStaked;

        return (rate_, contractBalance_);
    }

    function getCurrentStakerRewards(address staker_) public view returns (uint256) {
        (uint256 currentRate_, ) = getCurrentRate();

        return _getCurrentStakerRewards(currentRate_, stakers[staker_]);
    }

    function providerDeregister(bytes32[] calldata bidIds_) external {
        if (block.timestamp < deregistrationOpenAt) {
            _checkOwner();
        } else {
            deregistrationOpenAt = uint128(block.timestamp) + deregistrationTimeout;
        }

        _deleteModelBids(bidIds_);
        IProviderRegistry(lumerinDiamond).providerDeregister(address(this));

        deregistrationNonFeeOpened = uint128(block.timestamp);
    }

    function postModelBid(bytes32 modelId_, uint256 pricePerSecond_) external onlyOwner returns (bytes32) {
        return IMarketplace(lumerinDiamond).postModelBid(address(this), modelId_, pricePerSecond_);
    }

    function deleteModelBids(bytes32[] calldata bidIds_) external {
        if (block.timestamp < deregistrationOpenAt) {
            _checkOwner();
        }

        _deleteModelBids(bidIds_);
    }

    function _deleteModelBids(bytes32[] calldata bidIds_) private {
        address lumerinDiamond_ = lumerinDiamond;

        for (uint256 i = 0; i < bidIds_.length; i++) {
            IMarketplace(lumerinDiamond_).deleteModelBid(bidIds_[i]);
        }
    }

    function _getCurrentStakerRewards(uint256 delegatorRate_, Staker memory staker_) private pure returns (uint256) {
        uint256 newRewards_ = ((delegatorRate_ - staker_.rate) * staker_.staked) / PRECISION;

        return staker_.pendingRewards + newRewards_;
    }

    function version() external pure returns (uint256) {
        return 1;
    }
}
