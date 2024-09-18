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

    function bids(bytes32 bidId) external view returns (Bid memory) {
        return _getBidStorage().bids[bidId];
    }

    function providerActiveBids(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory) {
        return _getBidStorage().providerActiveBids[provider_].part(offset_, limit_);
    }

    function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) public view returns (bytes32[] memory) {
        return _getBidStorage().modelActiveBids[modelId_].part(offset_, limit_);
    }

    function providerBids(address provider_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory) {
        return _getBidStorage().providerBids[provider_].part(offset_, limit_);
    }

    function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory) {
        return _getBidStorage().modelBids[modelId_].part(offset_, limit_);
    }

    function getToken() internal view returns (IERC20) {
        return _getBidStorage().token;
    }

    function getBid(bytes32 bidId) internal view returns (Bid storage) {
        return _getBidStorage().bids[bidId];
    }

    function addProviderActiveBids(address provider, bytes32 bidId) internal {
        _getBidStorage().providerActiveBids[provider].add(bidId);
    }

    function addModelActiveBids(bytes32 modelId, bytes32 bidId) internal {
        _getBidStorage().modelActiveBids[modelId].add(bidId);
    }

    function removeProviderActiveBids(address provider, bytes32 bidId) internal {
        _getBidStorage().providerActiveBids[provider].remove(bidId);
    }

    function getModelActiveBids(bytes32 modelId) internal view returns (EnumerableSet.Bytes32Set storage) {
        return _getBidStorage().modelActiveBids[modelId];
    }

    function removeModelActiveBids(bytes32 modelId, bytes32 bidId) internal {
        _getBidStorage().modelActiveBids[modelId].remove(bidId);
    }

    function isModelActiveBidsEmpty(bytes32 modelId) internal view returns (bool) {
        return _getBidStorage().modelActiveBids[modelId].length() == 0;
    }

    function isProviderActiveBidsEmpty(address provider) internal view returns (bool) {
        return _getBidStorage().providerActiveBids[provider].length() == 0;
    }

    function addProviderBid(address provider, bytes32 bidId) internal {
        _getBidStorage().providerBids[provider].push(bidId);
    }

    function addModelBid(bytes32 modelId, bytes32 bidId) internal {
        _getBidStorage().modelBids[modelId].push(bidId);
    }

    function addBid(bytes32 bidId, Bid memory bid) internal {
        _getBidStorage().bids[bidId] = bid;
    }

    function _incrementBidNonce(bytes32 providerModelId) internal returns (uint256) {
        return _getBidStorage().providerModelNonce[providerModelId]++;
    }

    function _getBidStorage() internal pure returns (BDStorage storage _ds) {
        bytes32 slot_ = BID_STORAGE_SLOT;

        assembly {
            _ds.slot := slot_
        }
    }
}
