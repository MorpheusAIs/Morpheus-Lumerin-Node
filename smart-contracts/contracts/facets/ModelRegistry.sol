// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { AppStorage, Model, ModelStats } from "../AppStorage.sol";
import { KeySet } from "../libraries/KeySet.sol";
import { LibOwner } from "../libraries/LibOwner.sol";

contract ModelRegistry {
  using KeySet for KeySet.Set;

  AppStorage internal s;

  event ModelRegisteredUpdated(address indexed owner, bytes32 indexed modelId);
  event ModelDeregistered(address indexed owner, bytes32 indexed modelId);
  event ModelMinStakeUpdated(uint256 newStake);

  error ModelNotFound();
  error StakeTooLow();

  /// @notice Returns model struct by id
  function modelMap(bytes32 id) external view returns (Model memory) {
    return s.modelMap[id];
  }

  /// @notice Returns model id by index
  function models(uint256 index) external view returns (bytes32) {
    return s.models[index];
  }

  /// @notice Returns active (undeleted) model IDs
  function modelGetIds() external view returns (bytes32[] memory) {
    return s.activeModels.keys();
  }

  /// @notice Returns count of active models
  function modelGetCount() external view returns (uint count) {
    return s.activeModels.count();
  }

  /// @notice Returns all models
  /// @return ids    array of model ids
  /// @return models array of model structs
  function modelGetAll() external view returns (bytes32[] memory, Model[] memory) {
    uint256 len = s.activeModels.count();
    Model[] memory _models = new Model[](len);
    bytes32[] memory ids = new bytes32[](len);
    for (uint i = 0; i < len; i++) {
      bytes32 id = s.activeModels.keyAtIndex(i);
      ids[i] = id;
      _models[i] = s.modelMap[id];
    }
    return (ids, _models);
  }

  /// @notice Returns active model struct by index
  function modelGetByIndex(uint index) external view returns (bytes32 modelId, Model memory model) {
    modelId = s.activeModels.keyAtIndex(index);
    return (modelId, s.modelMap[modelId]);
  }

  /// @notice Checks if model exists
  function modelExists(bytes32 id) external view returns (bool) {
    return s.activeModels.exists(id);
  }

  /// @notice Registers or updates existing model
  function modelRegister(
    bytes32 modelId,
    bytes32 ipfsCID,
    uint256 fee,
    uint256 addStake,
    address owner,
    string memory name,
    string[] memory tags
  ) external {
    LibOwner._senderOrOwner(owner);
    Model memory model = s.modelMap[modelId];
    uint256 newStake = model.stake + addStake;
    if (newStake < s.modelMinStake) {
      revert StakeTooLow();
    }
    if (model.stake == 0) {
      s.activeModels.insert(modelId);
      s.models.push(modelId);
    } else {
      LibOwner._senderOrOwner(s.modelMap[modelId].owner);
    }

    s.modelMap[modelId] = Model({
      fee: fee,
      stake: newStake,
      createdAt: uint128(block.timestamp),
      ipfsCID: ipfsCID,
      owner: owner,
      name: name,
      tags: tags,
      isDeleted: false
    });

    emit ModelRegisteredUpdated(owner, modelId);
    s.token.transferFrom(msg.sender, address(this), addStake); // reverts with ERC20InsufficientAllowance()
  }

  /// @notice Deregisters a model
  function modelDeregister(bytes32 id) external {
    Model storage model = s.modelMap[id];
    LibOwner._senderOrOwner(model.owner);

    s.activeModels.remove(id); // reverts with KeyNotFound()
    model.isDeleted = true;
    uint256 stake = model.stake;

    emit ModelDeregistered(model.owner, id);
    s.token.transfer(model.owner, stake);
  }

  /// @notice Sets the minimum stake required for a model
  function modelSetMinStake(uint256 _minStake) external {
    LibOwner._onlyOwner();
    s.modelMinStake = _minStake;
    emit ModelMinStakeUpdated(s.modelMinStake);
  }

  /// @notice Returns the minimum stake required for a model
  function modelMinStake() external view returns (uint256) {
    return s.modelMinStake;
  }

  function modelStats(bytes32 id) external view returns (ModelStats memory) {
    return s.modelStats[id];
  }

  function modelResetStats(bytes32 id) external {
    LibOwner._onlyOwner();
    delete s.modelStats[id];
  }

  // TODO: implement these functions
  // function getModelsByOwner(address addr) external view returns (Model[] memory){
  //   Model[] memory _models = new Model[](modelIds.length);
  //   for (uint i = 0; i < modelIds.length; i++) {
  //     if (models[modelIds[i]].owner == addr) {
  //       _models[i] = models[modelIds[i]];
  //     }
  //   }
  //   return _models;
  // }

  // function getModelTypes -- to be implemented when types are defined
  // function getModelsByType
}
