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

    function __Marketplace_init(address token_) external initializer(MARKETPLACE_STORAGE_SLOT) {
        setToken(IERC20(token_));
    }

    function setMarketplaceBidFee(uint256 bidFee_) external onlyOwner {
        setBidFee(bidFee_);

        emit MaretplaceFeeUpdated(bidFee_);
    }

    function postModelBid(bytes32 modelId_, uint256 pricePerSecond_) external returns (bytes32 bidId) {
        address provider_ = _msgSender();

        if (!getIsProviderActive(provider_)) {
            revert MarketplaceProviderNotFound();
        }
        if (!getIsModelActive(modelId_)) {
            revert MarketplaceModelNotFound();
        }

        uint256 fee_ = getBidFee();
        getToken().safeTransferFrom(_msgSender(), address(this), fee_);

        setFeeBalance(getFeeBalance() + fee_);

        bytes32 providerModelId_ = getProviderModelId(provider_, modelId_);
        uint256 providerModelNonce_ = incrementBidNonce(providerModelId_);
        bytes32 bidId_ = getBidId(provider_, modelId_, providerModelNonce_);

        if (providerModelNonce_ != 0) {
            bytes32 oldBidId_ = getBidId(provider_, modelId_, providerModelNonce_ - 1);
            if (isBidActive(oldBidId_)) {
                _deleteBid(oldBidId_);
            }
        }

        Bid storage bid = bids(bidId_);
        bid.provider = provider_;
        bid.modelId = modelId_;
        bid.pricePerSecond = pricePerSecond_;
        bid.nonce = providerModelNonce_;
        bid.createdAt = uint128(block.timestamp);

        addProviderBid(provider_, bidId_);
        addModelBid(modelId_, bidId_);

        addProviderActiveBids(provider_, bidId_);
        addModelActiveBids(modelId_, bidId_);

        emit MarketplaceBidPosted(provider_, modelId_, providerModelNonce_);

        return bidId_;
    }

    function deleteModelBid(bytes32 bidId_) external {
        _onlyAccount(bids(bidId_).provider);

        if (!isBidActive(bidId_)) {
            revert MarketplaceActiveBidNotFound();
        }

        _deleteBid(bidId_);
    }

    function withdraw(address recipient_, uint256 amount_) external onlyOwner {
        uint256 feeBalance_ = getFeeBalance();
        amount_ = amount_ > feeBalance_ ? feeBalance_ : amount_;

        setFeeBalance(getFeeBalance() - amount_);

        getToken().safeTransfer(recipient_, amount_);
    }

    function _deleteBid(bytes32 bidId_) private {
        Bid storage bid = bids(bidId_);

        bid.deletedAt = uint128(block.timestamp);

        removeProviderActiveBids(bid.provider, bidId_);
        removeModelActiveBids(bid.modelId, bidId_);

        emit MarketplaceBidDeleted(bid.provider, bid.modelId, bid.nonce);
    }

    function getBidId(address provider_, bytes32 modelId_, uint256 nonce_) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(provider_, modelId_, nonce_));
    }

    function getProviderModelId(address provider_, bytes32 modelId_) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(provider_, modelId_));
    }
}
