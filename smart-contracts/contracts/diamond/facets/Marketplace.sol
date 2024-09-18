// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

import {BidStorage} from "../storages/BidStorage.sol";
import {ModelStorage} from "../storages/ModelStorage.sol";
import {ProviderStorage} from "../storages/ProviderStorage.sol";
import {MarketplaceStorage} from "../storages/MarketplaceStorage.sol";

import {IMarketplace} from "../../interfaces/facets/IMarketplace.sol";

contract Marketplace is
    IMarketplace,
    OwnableDiamondStorage,
    MarketplaceStorage,
    ProviderStorage,
    ModelStorage,
    BidStorage
{
    using SafeERC20 for IERC20;

    function __Marketplace_init(
        address token_
    ) external initializer(MARKETPLACE_STORAGE_SLOT) initializer(BID_STORAGE_SLOT) {
        _getBidStorage().token = IERC20(token_);
    }

    /// @notice sets a bid fee
    function setBidFee(uint256 bidFee_) external onlyOwner {
        _getMarketplaceStorage().bidFee = bidFee_;
        emit FeeUpdated(bidFee_);
    }

    /// @notice posts a new bid for a model
    function postModelBid(
        address provider_,
        bytes32 modelId_,
        uint256 pricePerSecond_
    ) external returns (bytes32 bidId) {
        if (!_ownerOrProvider(provider_)) {
            revert NotOwnerOrProvider();
        }
        if (!isProviderActive(provider_)) {
            revert ProviderNotFound();
        }
        if (!isModelActive(modelId_)) {
            revert ModelNotFound();
        }

        return _postModelBid(provider_, modelId_, pricePerSecond_);
    }

    /// @notice deletes a bid
    function deleteModelBid(bytes32 bidId_) external {
        if (!_isBidActive(bidId_)) {
            revert ActiveBidNotFound();
        }
        if (!_ownerOrProvider(getBid(bidId_).provider)) {
            revert NotOwnerOrProvider();
        }

        _deleteBid(bidId_);
    }

    /// @notice withdraws the fee balance
    function withdraw(address recipient_, uint256 amount_) external onlyOwner {
        if (amount_ > getFeeBalance()) {
            revert NotEnoughBalance();
        }

        decreaseFeeBalance(amount_);
        getToken().safeTransfer(recipient_, amount_);
    }

    function _incrementBidNonce(address provider_, bytes32 modelId_) private returns (uint256) {
        return _incrementBidNonce(getProviderModelId(provider_, modelId_));
    }

    function _postModelBid(address provider_, bytes32 modelId_, uint256 pricePerSecond_) private returns (bytes32) {
        uint256 fee = getBidFee();
        getToken().safeTransferFrom(_msgSender(), address(this), fee);
        increaseFeeBalance(fee);

        // TEST IT if it increments nonce correctly
        uint256 nonce_ = _incrementBidNonce(provider_, modelId_);
        if (nonce_ != 0) {
            bytes32 oldBidId_ = getBidId(provider_, modelId_, nonce_ - 1);
            if (_isBidActive(oldBidId_)) {
                _deleteBid(oldBidId_);
            }
        }

        bytes32 bidId = getBidId(provider_, modelId_, nonce_);

        addBid(bidId, Bid(provider_, modelId_, pricePerSecond_, nonce_, uint128(block.timestamp), 0));

        addProviderActiveBids(provider_, bidId);
        addModelActiveBids(modelId_, bidId);

        addProviderBid(provider_, bidId);
        addModelBid(modelId_, bidId);

        emit BidPosted(provider_, modelId_, nonce_);

        return bidId;
    }

    /// @dev passing bidId and bid storage to avoid double storage access
    function _deleteBid(bytes32 bidId_) private {
        Bid storage bid = getBid(bidId_);
        bid.deletedAt = uint128(block.timestamp);

        removeProviderActiveBids(bid.provider, bidId_);
        removeModelActiveBids(bid.modelId, bidId_);

        emit BidDeleted(bid.provider, bid.modelId, bid.nonce);
    }

    function getBidId(address provider_, bytes32 modelId_, uint256 nonce_) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(provider_, modelId_, nonce_));
    }

    function getProviderModelId(address provider_, bytes32 modelId_) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(provider_, modelId_));
    }

    function _ownerOrProvider(address provider_) private view returns (bool) {
        return _msgSender() == owner() || _msgSender() == provider_;
    }

    function _isBidActive(bytes32 bidId_) private view returns (bool) {
        Bid memory bid_ = getBid(bidId_);

        return bid_.createdAt != 0 && bid_.deletedAt == 0;
    }
}
