// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IModelStorage } from "../../interfaces/storage/IModelStorage.sol";

contract ModelStorage is IModelStorage {
  struct MDLStorage {
    mapping(bytes32 => Model) modelMap; // modelId => Model
    // mapping(address => bytes32[]) public modelsByOwner; // owner to modelIds
    bytes32[] models; // all model ids
    mapping(bytes32 => bool) activeModels; // active model ids
    uint256 modelMinimumStake;
  }

  bytes32 public constant MODEL_STORAGE_SLOT = keccak256("diamond.standard.model.storage");

  function getModel(bytes32 modelId) external view returns (Model memory) {
    return _getModelStorage().modelMap[modelId];
  }

  function models(uint256 index) external view returns (bytes32) {
    return _getModelStorage().models[index];
  }

  function modelMinimumStake() public view returns (uint256) {
    return _getModelStorage().modelMinimumStake;
  }

  function setModelActive(bytes32 modelId, bool isActive) internal {
    _getModelStorage().activeModels[modelId] = isActive;
  }

  function addModel(bytes32 modelId) internal {
    _getModelStorage().models.push(modelId);
  }

  function setModel(bytes32 modelId, Model memory model) internal {
    _getModelStorage().modelMap[modelId] = model;
  }

  function setModelMinimumStake(uint256 _modelMinimumStake) internal {
    _getModelStorage().modelMinimumStake = _modelMinimumStake;
  }

  function modelMap(bytes32 id) internal view returns (Model storage) {
    return _getModelStorage().modelMap[id];
  }

  function isModelActive(bytes32 modelId) internal view returns (bool) {
    return _getModelStorage().activeModels[modelId];
  }

  function _getModelStorage() internal pure returns (MDLStorage storage _ds) {
    bytes32 slot_ = MODEL_STORAGE_SLOT;

    assembly {
      _ds.slot := slot_
    }
  }
}
