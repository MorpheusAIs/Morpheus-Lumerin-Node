// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ISessionStorage} from "../../interfaces/storage/ISessionStorage.sol";

import {Paginator} from "@solarity/solidity-lib/libs/arrays/Paginator.sol";

contract SessionStorage is ISessionStorage {
    using Paginator for *;

    struct SNStorage {
        // all sessions
        uint256 sessionNonce; // used to generate unique session id
        mapping(bytes32 sessionId => Session) sessions;
        mapping(address user => bytes32[]) userSessionIds;
        mapping(address provider => bytes32[]) providerSessionIds;
        mapping(bytes32 modelId => bytes32[]) modelSessionIds;
        // active sessions
        uint256 totalClaimed; // total amount of MOR claimed by providers
        mapping(address user => mapping(bytes32 => bool)) isUserSessionActive; // user address => active session indexes
        mapping(address provider => mapping(bytes32 => bool)) isProviderSessionActive; // provider address => active session indexes
        mapping(address user => OnHold[]) userOnHold; // user address => balance
        mapping(bytes providerApproval => bool) isApprovalUsed;
        // other
        address fundingAccount; // account which stores the MOR tokens with infinite allowance for this contract
        Pool[] pools; // distribution pools configuration that mirrors L1 contract
    }

    bytes32 public constant SESSION_STORAGE_SLOT = keccak256("diamond.standard.session.storage");

    function sessions(bytes32 sessionId) external view returns (Session memory) {
        return _getSessionStorage().sessions[sessionId];
    }

    function getSessionsByUser(address user, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory) {
        return _getSessionStorage().userSessionIds[user].part(offset_, limit_);
    }

    function pools() external view returns (Pool[] memory) {
        return _getSessionStorage().pools;
    }

    function getPools() internal view returns (Pool[] storage) {
        return _getSessionStorage().pools;
    }

    function getPool(uint256 poolIndex) internal view returns (Pool storage) {
        return _getSessionStorage().pools[poolIndex];
    }

    function getFundingAccount() public view returns (address) {
        return _getSessionStorage().fundingAccount;
    }

    function setSession(bytes32 sessionId, Session memory session) internal {
        _getSessionStorage().sessions[sessionId] = session;
    }

    function setUserSessionActive(address user, bytes32 sessionId, bool active) internal {
        _getSessionStorage().isUserSessionActive[user][sessionId] = active;
    }

    function setProviderSessionActive(address provider, bytes32 sessionId, bool active) internal {
        _getSessionStorage().isProviderSessionActive[provider][sessionId] = active;
    }

    function addUserSessionId(address user, bytes32 sessionId) internal {
        _getSessionStorage().userSessionIds[user].push(sessionId);
    }

    function addProviderSessionId(address provider, bytes32 sessionId) internal {
        _getSessionStorage().providerSessionIds[provider].push(sessionId);
    }

    function totalSessions(address providerAddr) internal view returns (uint256) {
        return _getSessionStorage().providerSessionIds[providerAddr].length;
    }

    function addModelSessionId(bytes32 modelId, bytes32 sessionId) internal {
        _getSessionStorage().modelSessionIds[modelId].push(sessionId);
    }

    function addOnHold(address user, OnHold memory onHold) internal {
        _getSessionStorage().userOnHold[user].push(onHold);
    }

    function increaseTotalClaimed(uint256 amount) internal {
        _getSessionStorage().totalClaimed += amount;
    }

    function totalClaimed() internal view returns (uint256) {
        return _getSessionStorage().totalClaimed;
    }

    function getOnHold(address user) internal view returns (OnHold[] storage) {
        return _getSessionStorage().userOnHold[user];
    }

    function _getSession(bytes32 sessionId) internal view returns (Session storage) {
        return _getSessionStorage().sessions[sessionId];
    }

    function incrementSessionNonce() internal returns (uint256) {
        return _getSessionStorage().sessionNonce++;
    }

    function isApproved(bytes memory approval) internal view returns (bool) {
        return _getSessionStorage().isApprovalUsed[approval];
    }

    function approve(bytes memory approval) internal {
        _getSessionStorage().isApprovalUsed[approval] = true;
    }

    function _getSessionStorage() internal pure returns (SNStorage storage _ds) {
        bytes32 slot_ = SESSION_STORAGE_SLOT;

        assembly {
            _ds.slot := slot_
        }
    }
}
