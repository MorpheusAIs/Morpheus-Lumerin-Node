// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {Paginator} from "@solarity/solidity-lib/libs/arrays/Paginator.sol";

import {ISessionStorage} from "../../interfaces/storage/ISessionStorage.sol";

contract SessionStorage is ISessionStorage {
    using Paginator for *;
    using EnumerableSet for EnumerableSet.Bytes32Set;

    struct SessionsStorage {
        // Account which stores the MOR tokens with infinite allowance for this contract
        address fundingAccount;
        // Distribution pools configuration that mirrors L1 contract
        Pool[] pools;
        // Total amount of MOR claimed by providers
        uint256 providersTotalClaimed;
        // Used to generate unique session ID
        uint256 sessionNonce;
        mapping(bytes32 sessionId => Session) sessions;
        // Max ession duration
        uint128 maxSessionDuration;
        // Session registry for providers, users and models
        mapping(address user => EnumerableSet.Bytes32Set) userSessions;
        mapping(address provider => EnumerableSet.Bytes32Set) providerSessions;
        mapping(bytes32 modelId => EnumerableSet.Bytes32Set) modelSessions;
        mapping(address user => OnHold[]) userStakesOnHold;
        mapping(bytes providerApproval => bool) isProviderApprovalUsed;
    }

    bytes32 public constant SESSIONS_STORAGE_SLOT = keccak256("diamond.standard.sessions.storage");
    uint32 public constant MIN_SESSION_DURATION = 5 minutes;
    uint32 public constant SIGNATURE_TTL = 10 minutes;
    uint256 public constant COMPUTE_POOL_INDEX = 3;

    /** PUBLIC, GETTERS */
    function getSession(bytes32 sessionId_) external view returns (Session memory) {
        return _getSessionsStorage().sessions[sessionId_];
    }

    function getUserSessions(address user_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory) {
        return _getSessionsStorage().userSessions[user_].part(offset_, limit_);
    }

    function getProviderSessions(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory) {
        return _getSessionsStorage().providerSessions[provider_].part(offset_, limit_);
    }

    function getModelSessions(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory) {
        return _getSessionsStorage().modelSessions[modelId_].part(offset_, limit_);
    }

    function getPools() external view returns (Pool[] memory) {
        return _getSessionsStorage().pools;
    }

    function getPool(uint256 index_) external view returns (Pool memory) {
        return _getSessionsStorage().pools[index_];
    }

    function getFundingAccount() external view returns (address) {
        return _getSessionsStorage().fundingAccount;
    }

    function getTotalSessions(address provider_) public view returns (uint256) {
        return _getSessionsStorage().providerSessions[provider_].length();
    }

    function getProvidersTotalClaimed() external view returns (uint256) {
        return _getSessionsStorage().providersTotalClaimed;
    }

    function getIsProviderApprovalUsed(bytes memory approval_) external view returns (bool) {
        return _getSessionsStorage().isProviderApprovalUsed[approval_];
    }

    function getMaxSessionDuration() external view returns (uint128) {
        return _getSessionsStorage().maxSessionDuration;
    }

    /** INTERNAL */
    function _getSessionsStorage() internal pure returns (SessionsStorage storage ds) {
        bytes32 slot_ = SESSIONS_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }
}
