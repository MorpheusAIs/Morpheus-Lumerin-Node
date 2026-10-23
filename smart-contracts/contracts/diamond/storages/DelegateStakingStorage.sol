// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IDelegateStakingStorage} from "../../interfaces/storage/IDelegateStakingStorage.sol";

contract DelegateStakingStorage is IDelegateStakingStorage {
    struct Grant {
        uint256 cumulativeFundingCap;
        uint256 lifetimeFunded;
        uint128 currentPrincipal;
        uint128 pending;
        uint128 exposure;
        uint64 expiry;
        bool revoked;
        bool approved;
        bool known;
    }

    struct LinkedNode {
        address previous;
        address next;
        bool linked;
    }

    struct PoolSessionDrawData {
        address funder;
        uint96 amount;
    }

    struct PoolData {
        uint128 free;
        uint128 inSession;
        uint128 locked;
        uint128 pendingTotal;
        mapping(address funder => Grant) grants;
        address[] allFunders;
        address activeHead;
        uint8 activeCount;
        address activeTail;
        address purgeCursor;
        mapping(address funder => LinkedNode) activeNodes;
        address withdrawalHead;
        address withdrawalTail;
        uint256 withdrawalCount;
        mapping(address funder => LinkedNode) withdrawalNodes;
        uint256[] lockReleaseDates;
        uint256 lockHead;
        mapping(uint256 releaseAt => uint256 amount) lockAmounts;
    }

    struct PoolSessionData {
        address hot;
        bool exists;
        bool closed;
        bool settled;
        PoolSessionDrawData[] draws;
    }

    struct DelegateStakingData {
        mapping(address hot => PoolData) pools;
        mapping(bytes32 sessionId => PoolSessionData) poolSessions;
        uint8 maxActiveFunders;
        uint96 minPrincipal;
        uint32 settlementGrace;
        uint8 maxAutoService;
        uint8 maxAutoReleaseDays;
    }

    bytes32 internal constant DELEGATE_STAKING_STORAGE_SLOT = keccak256("diamond.standard.delegate.staking.storage");
    bytes32 internal constant DELEGATION_RULES_DELEGATE_STAKING = keccak256("delegation.rules.delegate.staking");

    uint256 internal constant MAX_ACTIVE_FUNDERS_LIMIT = 128;
    uint256 internal constant MAX_POOL_SESSION_DRAWS = MAX_ACTIVE_FUNDERS_LIMIT + 1;
    uint256 internal constant MAX_AUTO_SERVICE_LIMIT = 32;
    uint256 internal constant MAX_AUTO_RELEASE_DAYS_LIMIT = 32;
    uint256 internal constant MAX_SETTLEMENT_GRACE_LIMIT = 7 days;
    uint256 internal constant MAX_MIN_PRINCIPAL_LIMIT = 1_000_000_000 ether;

    uint256 internal constant DEFAULT_MAX_ACTIVE_FUNDERS = 64;
    uint256 internal constant DEFAULT_MIN_PRINCIPAL = 1 ether;
    uint256 internal constant DEFAULT_SETTLEMENT_GRACE = 1 days;
    uint256 internal constant DEFAULT_MAX_AUTO_SERVICE = 5;
    uint256 internal constant DEFAULT_MAX_AUTO_RELEASE_DAYS = 8;

    function _getDelegateStakingStorage() internal pure returns (DelegateStakingData storage ds) {
        bytes32 slot_ = DELEGATE_STAKING_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }

    function _isPoolSession(bytes32 sessionId_) internal view returns (bool) {
        return _getDelegateStakingStorage().poolSessions[sessionId_].exists;
    }

    function _maxActiveFunders() internal view returns (uint256) {
        uint256 value_ = _getDelegateStakingStorage().maxActiveFunders;
        return value_ == 0 ? DEFAULT_MAX_ACTIVE_FUNDERS : value_;
    }

    function _minPrincipal() internal view returns (uint256) {
        uint256 value_ = _getDelegateStakingStorage().minPrincipal;
        return value_ == 0 ? DEFAULT_MIN_PRINCIPAL : value_;
    }

    function _settlementGrace() internal view returns (uint256) {
        uint256 value_ = _getDelegateStakingStorage().settlementGrace;
        return value_ == 0 ? DEFAULT_SETTLEMENT_GRACE : value_;
    }

    function _maxAutoService() internal view returns (uint256) {
        uint256 value_ = _getDelegateStakingStorage().maxAutoService;
        return value_ == 0 ? DEFAULT_MAX_AUTO_SERVICE : value_;
    }

    function _maxAutoReleaseDays() internal view returns (uint256) {
        uint256 value_ = _getDelegateStakingStorage().maxAutoReleaseDays;
        return value_ == 0 ? DEFAULT_MAX_AUTO_RELEASE_DAYS : value_;
    }
}
