// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";

import {PRECISION} from "@solarity/solidity-lib/utils/Globals.sol";

import {IProvidersDelegate} from "../interfaces/delegate/IProvidersDelegate.sol";
import {IBidStorage} from "../interfaces/storage/IBidStorage.sol";
import {ISessionRouter} from "../interfaces/facets/ISessionRouter.sol";
import {IProviderRegistry} from "../interfaces/facets/IProviderRegistry.sol";
import {IMarketplace} from "../interfaces/facets/IMarketplace.sol";

contract ProvidersDelegate is IProvidersDelegate, OwnableUpgradeable {
    using SafeERC20 for IERC20;
    using Math for uint256;

    // The contract deps
    address public lumerinDiamond;
    address public token;

    // The owner fee
    address public feeTreasury;
    uint256 public fee;

    // The contract metadata
    string public name;
    string public endpoint;

    // The main calculation storage
    uint256 public totalStaked;
    uint256 public totalRate;
    uint256 public lastContractBalance;

    // The Staker data
    bool public isStakeClosed;
    mapping(address => Staker) public stakers;

    // Deregistration limits
    uint128 public deregistrationOpensAt;

    constructor() {
        _disableInitializers();
    }

    function ProvidersDelegate_init(
        address lumerinDiamond_,
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_,
        uint128 deregistrationOpensAt_
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
        deregistrationOpensAt = deregistrationOpensAt_;

        IERC20(token).approve(lumerinDiamond_, type(uint256).max);
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
        _stake(_msgSender(), amount_);
    }

    function _stake(address staker_, uint256 amount_) private {
        if (isStakeClosed && !isDeregisterAvailable()) {
            revert StakeClosed();
        }
        if (amount_ == 0) {
            revert InsufficientAmount();
        }

        Staker storage staker = stakers[staker_];

        (uint256 currentRate_, uint256 contractBalance_) = getCurrentRate();
        uint256 pendingRewards_ = _getCurrentStakerRewards(currentRate_, staker);

        IERC20(token).safeTransferFrom(staker_, address(this), amount_);

        totalRate = currentRate_;
        totalStaked += amount_;

        lastContractBalance = contractBalance_;

        staker.rate = currentRate_;
        staker.staked += amount_;
        staker.pendingRewards = pendingRewards_;

        IProviderRegistry(lumerinDiamond).providerRegister(address(this), amount_, endpoint);

        emit Staked(staker_, staker.staked, totalStaked, staker.rate);
    } 

    function restake(address staker_, uint256 amount_) external {
        if (_msgSender() != staker_ && _msgSender() != owner()) {
            revert RestakeInvalidCaller(_msgSender(), staker_);
        }
        if (_msgSender() == owner() && stakers[staker_].isRestakeDisabled) {
            revert RestakeDisabled(staker_);
        }

        amount_ = claim(staker_, amount_);
        _stake(staker_, amount_);
    }

    function claim(address staker_, uint256 amount_) public returns (uint256) {
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
        if (feeAmount_ != 0) {
            IERC20(token).safeTransfer(feeTreasury, feeAmount_);

            amount_ -= feeAmount_;

            emit FeeClaimed(feeTreasury, feeAmount_);
        }

        IERC20(token).safeTransfer(staker_, amount_);

        emit Claimed(staker_, staker.claimed, staker.rate);

        return amount_;
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
        if (!isDeregisterAvailable()) {
            _checkOwner();
        }

        _deleteModelBids(bidIds_);
        IProviderRegistry(lumerinDiamond).providerDeregister(address(this));

        fee = 0;
    }

    function postModelBid(bytes32 modelId_, uint256 pricePerSecond_) external onlyOwner returns (bytes32) {
        if (isDeregisterAvailable()) {
            revert BidCannotBeCreatedDuringThisPeriod();
        }

        return IMarketplace(lumerinDiamond).postModelBid(address(this), modelId_, pricePerSecond_);
    }

    function deleteModelBids(bytes32[] calldata bidIds_) external {
        if (!isDeregisterAvailable()) {
            _checkOwner();
        }

        _deleteModelBids(bidIds_);
    }

    function claimForProvider(bytes32 sessionId_) external {
        ISessionRouter(lumerinDiamond).claimForProvider(sessionId_);
    }

    function isDeregisterAvailable() public view returns (bool) {
        return block.timestamp >= deregistrationOpensAt;
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
