// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import {Paginator} from "@solarity/solidity-lib/libs/arrays/Paginator.sol";

import {IBidStorage} from "../../interfaces/storage/IBidStorage.sol";

contract BidStorage is IBidStorage {
    using Paginator for *;
    using EnumerableSet for EnumerableSet.Bytes32Set;

    struct BDStorage {
        IERC20 token; // MOR token
        mapping(bytes32 bidId => Bid) bids; // bidId = keccak256(provider, modelId, nonce)
        mapping(bytes32 modelId => bytes32[]) modelBids; // keccak256(provider, modelId) => all bidIds
        mapping(bytes32 modelId => EnumerableSet.Bytes32Set) modelActiveBids; // modelId => active bidIds
        mapping(address provider => bytes32[]) providerBids; // provider => all bidIds
        mapping(address provider => EnumerableSet.Bytes32Set) providerActiveBids; // provider => active bidIds
        mapping(bytes32 providerModelId => uint256) providerModelNonce; // keccak256(provider, modelId) => last nonce
    }

    bytes32 public constant BID_STORAGE_SLOT = keccak256("diamond.standard.bid.storage");

    /** PUBLIC, GETTERS */
    function getBid(bytes32 bidId_) external view returns (Bid memory) {
        return _getBidStorage().bids[bidId_];
    }

    function getProviderActiveBids(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory) {
        return _getBidStorage().providerActiveBids[provider_].part(offset_, limit_);
    }

    function getModelActiveBids(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) public view returns (bytes32[] memory) {
        return _getBidStorage().modelActiveBids[modelId_].part(offset_, limit_);
    }

    function getProviderBids(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory) {
        return _getBidStorage().providerBids[provider_].part(offset_, limit_);
    }

    function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory) {
        return _getBidStorage().modelBids[modelId_].part(offset_, limit_);
    }

    function getToken() public view returns (IERC20) {
        return _getBidStorage().token;
    }

    function isBidActive(bytes32 bidId_) public view returns (bool) {
        Bid storage bid = _getBidStorage().bids[bidId_];

        return bid.createdAt != 0 && bid.deletedAt == 0;
    }

    /** INTERNAL, GETTERS */
    function bids(bytes32 bidId_) internal view returns (Bid storage) {
        return _getBidStorage().bids[bidId_];
    }

    function isModelActiveBidsEmpty(bytes32 modelId) internal view returns (bool) {
        return _getBidStorage().modelActiveBids[modelId].length() == 0;
    }

    function isProviderActiveBidsEmpty(address provider) internal view returns (bool) {
        return _getBidStorage().providerActiveBids[provider].length() == 0;
    }

    /** INTERNAL, SETTERS */
    function addProviderActiveBids(address provider_, bytes32 bidId_) internal {
        _getBidStorage().providerActiveBids[provider_].add(bidId_);
    }

    function removeProviderActiveBids(address provider_, bytes32 bidId_) internal {
        _getBidStorage().providerActiveBids[provider_].remove(bidId_);
    }

    function addModelActiveBids(bytes32 modelId_, bytes32 bidId_) internal {
        _getBidStorage().modelActiveBids[modelId_].add(bidId_);
    }

    function removeModelActiveBids(bytes32 modelId_, bytes32 bidId_) internal {
        _getBidStorage().modelActiveBids[modelId_].remove(bidId_);
    }

    function addProviderBid(address provider_, bytes32 bidId_) internal {
        _getBidStorage().providerBids[provider_].push(bidId_);
    }

    function addModelBid(bytes32 modelId_, bytes32 bidId_) internal {
        _getBidStorage().modelBids[modelId_].push(bidId_);
    }

    function setToken(IERC20 token_) internal {
        _getBidStorage().token = token_;
    }

    function incrementBidNonce(bytes32 providerModelId_) internal returns (uint256) {
        return _getBidStorage().providerModelNonce[providerModelId_]++;
    }

    /** PRIVATE */
    function _getBidStorage() private pure returns (BDStorage storage ds) {
        bytes32 slot_ = BID_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }
}
