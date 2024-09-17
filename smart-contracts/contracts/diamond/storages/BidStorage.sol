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
        mapping(bytes32 => Bid) bidMap; // bidId = keccak256(provider, modelAgentId, nonce) => bid
        mapping(bytes32 => uint256) providerModelAgentNonce; // keccak256(provider, modelAgentId) => last nonce
        mapping(bytes32 => bool) activeBids; // all active bidIds
        mapping(address => EnumerableSet.Bytes32Set) providerActiveBids; // provider => active bidIds
        mapping(bytes32 => EnumerableSet.Bytes32Set) modelAgentActiveBids; // modelAgentId => active bidIds
        mapping(bytes32 => bytes32[]) modelAgentBids; // keccak256(provider, modelAgentId) => all bidIds
        mapping(address => bytes32[]) providerBids; // provider => all bidIds
    }

    bytes32 public constant BID_STORAGE_SLOT = keccak256("diamond.standard.bid.storage");

    function bidMap(bytes32 bidId) external view returns (Bid memory) {
        return _getBidStorage().bidMap[bidId];
    }

    function providerActiveBids(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory) {
        return _getBidStorage().providerActiveBids[provider_].part(offset_, limit_);
    }

    function modelAgentActiveBids(
        bytes32 modelAgentId_,
        uint256 offset_,
        uint256 limit_
    ) public view returns (bytes32[] memory) {
        return _getBidStorage().modelAgentActiveBids[modelAgentId_].part(offset_, limit_);
    }

    function providerBids(address provider_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory) {
        return _getBidStorage().providerBids[provider_].part(offset_, limit_);
    }

    function modelAgentBids(
        bytes32 modelAgentId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory) {
        return _getBidStorage().modelAgentBids[modelAgentId_].part(offset_, limit_);
    }

    function getToken() internal view returns (IERC20) {
        return _getBidStorage().token;
    }

    function getBid(bytes32 bidId) internal view returns (Bid storage) {
        return _getBidStorage().bidMap[bidId];
    }

    function setBidActive(bytes32 bidId, bool active) internal {
        _getBidStorage().activeBids[bidId] = active;
    }

    function addProviderActiveBids(address provider, bytes32 bidId) internal {
        _getBidStorage().providerActiveBids[provider].add(bidId);
    }

    function addModelAgentActiveBids(bytes32 modelAgentId, bytes32 bidId) internal {
        _getBidStorage().modelAgentActiveBids[modelAgentId].add(bidId);
    }

    function removeProviderActiveBids(address provider, bytes32 bidId) internal {
        _getBidStorage().providerActiveBids[provider].remove(bidId);
    }

    function getModelAgentActiveBids(bytes32 modelAgentId) internal view returns (EnumerableSet.Bytes32Set storage) {
        return _getBidStorage().modelAgentActiveBids[modelAgentId];
    }

    function removeModelAgentActiveBids(bytes32 modelAgentId, bytes32 bidId) internal {
        _getBidStorage().modelAgentActiveBids[modelAgentId].remove(bidId);
    }

    function isModelAgentActiveBidsEmpty(bytes32 modelAgentId) internal view returns (bool) {
        return _getBidStorage().modelAgentActiveBids[modelAgentId].length() == 0;
    }

    function isProviderActiveBidsEmpty(address provider) internal view returns (bool) {
        return _getBidStorage().providerActiveBids[provider].length() == 0;
    }

    function addProviderBid(address provider, bytes32 bidId) internal {
        _getBidStorage().providerBids[provider].push(bidId);
    }

    function addModelAgentBid(bytes32 modelAgentId, bytes32 bidId) internal {
        _getBidStorage().modelAgentBids[modelAgentId].push(bidId);
    }

    function addBid(bytes32 bidId, Bid memory bid) internal {
        _getBidStorage().bidMap[bidId] = bid;
    }

    function incrementNonce(bytes32 providerModelAgentId) internal returns (uint256) {
        return _getBidStorage().providerModelAgentNonce[providerModelAgentId]++;
    }

    function _getBidStorage() internal pure returns (BDStorage storage _ds) {
        bytes32 slot_ = BID_STORAGE_SLOT;

        assembly {
            _ds.slot := slot_
        }
    }
}
