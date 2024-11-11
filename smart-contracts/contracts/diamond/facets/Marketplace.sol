// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
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
    using EnumerableSet for EnumerableSet.Bytes32Set;

    function __Marketplace_init(
        address token_,
        uint256 bidMinPricePerSecond_,
        uint256 bidMaxPricePerSecond_
    ) external initializer(BIDS_STORAGE_SLOT) {
        BidsStorage storage bidsStorage = _getBidsStorage();
        bidsStorage.token = token_;

        setMinMaxBidPricePerSecond(bidMinPricePerSecond_, bidMaxPricePerSecond_);
    }

    function setMarketplaceBidFee(uint256 bidFee_) external onlyOwner {
        MarketStorage storage marketStorage = _getMarketStorage();
        marketStorage.bidFee = bidFee_;

        emit MaretplaceFeeUpdated(bidFee_);
    }

    function setMinMaxBidPricePerSecond(uint256 bidMinPricePerSecond_, uint256 bidMaxPricePerSecond_) public onlyOwner {
        if (bidMinPricePerSecond_ == 0) {
            revert MarketplaceBidMinPricePerSecondIsZero();
        }

        if (bidMinPricePerSecond_ > bidMaxPricePerSecond_) {
            revert MarketplaceBidMinPricePerSecondIsInvalid();
        }

        MarketStorage storage marketStorage = _getMarketStorage();
        marketStorage.bidMinPricePerSecond = bidMinPricePerSecond_;
        marketStorage.bidMaxPricePerSecond = bidMaxPricePerSecond_;

        emit MarketplaceBidMinMaxPriceUpdated(bidMinPricePerSecond_, bidMaxPricePerSecond_);
    }

    function postModelBid(bytes32 modelId_, uint256 pricePerSecond_) external returns (bytes32 bidId) {
        address provider_ = _msgSender();

        if (!getIsProviderActive(provider_)) {
            revert MarketplaceProviderNotFound();
        }
        if (!getIsModelActive(modelId_)) {
            revert MarketplaceModelNotFound();
        }

        BidsStorage storage bidsStorage = _getBidsStorage();
        MarketStorage storage marketStorage = _getMarketStorage();

        if (
            pricePerSecond_ < marketStorage.bidMinPricePerSecond || pricePerSecond_ > marketStorage.bidMaxPricePerSecond
        ) {
            revert MarketplaceBidPricePerSecondInvalid();
        }

        IERC20(bidsStorage.token).safeTransferFrom(provider_, address(this), marketStorage.bidFee);
        marketStorage.feeBalance += marketStorage.bidFee;

        bytes32 providerModelId_ = getProviderModelId(provider_, modelId_);
        uint256 providerModelNonce_ = bidsStorage.providerModelNonce[providerModelId_]++;
        bytes32 bidId_ = getBidId(provider_, modelId_, providerModelNonce_);

        if (providerModelNonce_ != 0) {
            bytes32 oldBidId_ = getBidId(provider_, modelId_, providerModelNonce_ - 1);
            if (isBidActive(oldBidId_)) {
                _deleteBid(oldBidId_);
            }
        }

        Bid storage bid = bidsStorage.bids[bidId_];
        bid.provider = provider_;
        bid.modelId = modelId_;
        bid.pricePerSecond = pricePerSecond_;
        bid.nonce = providerModelNonce_;
        bid.createdAt = uint128(block.timestamp);

        bidsStorage.providerBids[provider_].add(bidId_);
        bidsStorage.providerActiveBids[provider_].add(bidId_);
        bidsStorage.modelBids[modelId_].add(bidId_);
        bidsStorage.modelActiveBids[modelId_].add(bidId_);

        emit MarketplaceBidPosted(provider_, modelId_, providerModelNonce_);

        return bidId_;
    }

    function deleteModelBid(bytes32 bidId_) external {
        BidsStorage storage bidsStorage = _getBidsStorage();
        _onlyAccount(bidsStorage.bids[bidId_].provider);

        if (!isBidActive(bidId_)) {
            revert MarketplaceActiveBidNotFound();
        }

        _deleteBid(bidId_);
    }

    function withdrawFee(address recipient_, uint256 amount_) external onlyOwner {
        BidsStorage storage bidsStorage = _getBidsStorage();
        MarketStorage storage marketStorage = _getMarketStorage();

        amount_ = amount_ > marketStorage.feeBalance ? marketStorage.feeBalance : amount_;
        if (amount_ == 0) {
            revert MarketplaceFeeAmountIsZero();
        }

        marketStorage.feeBalance -= amount_;

        IERC20(bidsStorage.token).safeTransfer(recipient_, amount_);
    }

    function _deleteBid(bytes32 bidId_) private {
        BidsStorage storage bidsStorage = _getBidsStorage();
        Bid storage bid = bidsStorage.bids[bidId_];

        bid.deletedAt = uint128(block.timestamp);

        bidsStorage.providerActiveBids[bid.provider].remove(bidId_);
        bidsStorage.modelActiveBids[bid.modelId].remove(bidId_);

        emit MarketplaceBidDeleted(bid.provider, bid.modelId, bid.nonce);
    }

    function getBidId(address provider_, bytes32 modelId_, uint256 nonce_) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(provider_, modelId_, nonce_));
    }

    function getProviderModelId(address provider_, bytes32 modelId_) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(provider_, modelId_));
    }
}
