// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { ISessionStorage } from "../../interfaces/storage/ISessionStorage.sol";

contract SessionStorage is ISessionStorage {
  struct SNStorage {
    // all sessions
    Session[] sessions;
    mapping(bytes32 => uint256) sessionMap; // sessionId => session index
    mapping(address => uint256[]) userSessions; // user address => all session indexes
    mapping(address => uint256[]) providerSessions; // provider address => all session indexes
    mapping(bytes32 => uint256[]) modelSessions; // modelId => all session indexes
    uint256 sessionNonce; // used to generate unique session id
    // active sessions
    mapping(address => mapping(uint256 => bool)) userActiveSessions; // user address => active session indexes
    mapping(address => mapping(uint256 => bool)) providerActiveSessions; // provider address => active session indexes
    mapping(address => OnHold[]) userOnHold; // user address => balance
    mapping(bytes => bool) approvalMap; // provider approval => true if approval was already used
    uint64 activeSessionsCount;
    uint256 totalClaimed; // total amount of MOR claimed by providers
    // other
    address fundingAccount; // account which stores the MOR tokens with infinite allowance for this contract
    Pool[] pools; // distribution pools configuration that mirrors L1 contract
  }

  bytes32 public constant SESSION_STORAGE_SLOT = keccak256("diamond.standard.session.storage");

  function sessionMap(bytes32 sessionId) external view returns (uint256) {
    return _getSessionStorage().sessionMap[sessionId];
  }

  function getPools() internal view returns (Pool[] storage) {
    return _getSessionStorage().pools;
  }

  function getPool(uint256 poolIndex) internal view returns (Pool storage) {
    return _getSessionStorage().pools[poolIndex];
  }

  function getFundingAccount() internal view returns (address) {
    return _getSessionStorage().fundingAccount;
  }

  function addSession(Session memory session) internal {
    _getSessionStorage().sessions.push(session);
  }

  function setSession(bytes32 sessionId, Session memory session) internal {
    _getSessionStorage().sessions[getSessionIndex(sessionId)] = session;
  }

  function setActiveUserSession(address user, uint256 sessionIndex, bool active) internal {
    _getSessionStorage().userActiveSessions[user][sessionIndex] = active;
  }

  function setActiveProviderSession(address provider, uint256 sessionIndex, bool active) internal {
    _getSessionStorage().providerActiveSessions[provider][sessionIndex] = active;
  }

  function incrementActiveSessionsCount() internal {
    _getSessionStorage().activeSessionsCount++;
  }

  function decrementActiveSessionsCount() internal {
    _getSessionStorage().activeSessionsCount--;
  }

  function setSession(bytes32 sessionId, uint256 sessionIndex) internal {
    _getSessionStorage().sessionMap[sessionId] = sessionIndex;
  }

  function addUserSession(address user, uint256 sessionIndex) internal {
    _getSessionStorage().userSessions[user].push(sessionIndex);
  }

  function addProviderSession(address provider, uint256 sessionIndex) internal {
    _getSessionStorage().providerSessions[provider].push(sessionIndex);
  }

  function totalSessions(address providerAddr) internal view returns (uint256) {
    return _getSessionStorage().providerSessions[providerAddr].length;
  }

  function addModelSession(bytes32 modelAgentId, uint256 sessionIndex) internal {
    _getSessionStorage().modelSessions[modelAgentId].push(sessionIndex);
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

  function getActiveSessionsCount() internal view returns (uint256) {
    return _getSessionStorage().activeSessionsCount;
  }

  function getSession(bytes32 sessionId) internal view returns (Session storage) {
    return _getSessionStorage().sessions[getSessionIndex(sessionId)];
  }

  function getSession(uint256 sessionIndex) internal view returns (Session storage) {
    return _getSessionStorage().sessions[sessionIndex];
  }

  function getSessionIndex(bytes32 sessionId) internal view returns (uint256) {
    return _getSessionStorage().sessionMap[sessionId];
  }

  function incrementSessionNonce() internal returns (uint256) {
    return _getSessionStorage().sessionNonce++;
  }

  function getNextSessionIndex() internal view returns (uint256) {
    return _getSessionStorage().sessions.length - 1;
  }

  function isApproved(bytes memory approval) internal view returns (bool) {
    return _getSessionStorage().approvalMap[approval];
  }

  function approve(bytes memory approval) internal {
    _getSessionStorage().approvalMap[approval] = true;
  }

  function _getSessionStorage() internal pure returns (SNStorage storage _ds) {
    bytes32 slot_ = SESSION_STORAGE_SLOT;

    assembly {
      _ds.slot := slot_
    }
  }
}
