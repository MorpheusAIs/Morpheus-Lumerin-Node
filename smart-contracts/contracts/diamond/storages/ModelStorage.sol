// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Paginator} from "@solarity/solidity-lib/libs/arrays/Paginator.sol";

import {IModelStorage} from "../../interfaces/storage/IModelStorage.sol";

contract ModelStorage is IModelStorage {
    using Paginator for *;

    struct MDLStorage {
        uint256 modelMinimumStake;
        bytes32[] modelIds;
        mapping(bytes32 modelId => Model) models;
    }

    bytes32 public constant MODEL_STORAGE_SLOT = keccak256("diamond.standard.model.storage");

    /** PUBLIC, GETTERS */
    function getModel(bytes32 modelId_) external view returns (Model memory) {
        return _getModelStorage().models[modelId_];
    }

    function getModelIds(uint256 offset_, uint256 limit_) external view returns (bytes32[] memory) {
        return _getModelStorage().modelIds.part(offset_, limit_);
    }

    function getModelMinimumStake() public view returns (uint256) {
        return _getModelStorage().modelMinimumStake;
    }

    function getIsModelActive(bytes32 modelId_) public view returns (bool) {
        return !_getModelStorage().models[modelId_].isDeleted;
    }

    /** INTERNAL, GETTERS */
    function models(bytes32 modelId_) internal view returns (Model storage) {
        return _getModelStorage().models[modelId_];
    }

    /** INTERNAL, SETTERS */
    function addModelId(bytes32 modelId_) internal {
        _getModelStorage().modelIds.push(modelId_);
    }

    function setModelMinimumStake(uint256 modelMinimumStake_) internal {
        _getModelStorage().modelMinimumStake = modelMinimumStake_;
    }

    /** PRIVATE */
    function _getModelStorage() private pure returns (MDLStorage storage ds) {
        bytes32 slot_ = MODEL_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }
}
