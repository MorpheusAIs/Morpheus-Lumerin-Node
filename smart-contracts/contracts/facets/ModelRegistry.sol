// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { AppStorage, Model } from "../AppStorage.sol";
import { KeySet } from "../libraries/KeySet.sol";
import { LibOwner } from "../libraries/LibOwner.sol";

contract ModelRegistry {
  using KeySet for KeySet.Set;

  AppStorage internal s;

  error ModelNotFound();
  error StakeTooLow();

  event ModelRegisteredUpdated(address indexed owner, bytes32 indexed modelId);
  event ModelDeregistered(address indexed owner, bytes32 indexed modelId);
  event ModelMinStakeUpdated(uint256 newStake);

  function modelMap(
    bytes32 id
  )
    public
    view
    returns (bytes32, uint256, uint256, address, string memory, uint128, bool)
  {
    Model memory model = s.modelMap[id];
    return (
      model.ipfsCID,
      model.fee,
      model.stake,
      model.owner,
      model.name,
      model.timestamp,
      model.isDeleted
    );
  }

  function models(uint256 index) public view returns (bytes32) {
    return s.models[index];
  }

  function modelGetIds() public view returns (bytes32[] memory) {
    return s.activeModels.keys();
  }

  function modelGetCount() public view returns (uint count) {
    return s.activeModels.count();
  }

  function modelGetAll() public view returns (Model[] memory) {
    Model[] memory _models = new Model[](s.activeModels.count());
    for (uint i = 0; i < s.activeModels.count(); i++) {
      _models[i] = s.modelMap[s.activeModels.keyAtIndex(i)];
    }
    return _models;
  }

  function modelGetByIndex(
    uint index
  ) public view returns (bytes32 modelId, Model memory model) {
    modelId = s.activeModels.keyAtIndex(index);
    return (modelId, s.modelMap[modelId]);
  }

  function modelExists(bytes32 id) public view returns (bool) {
    return s.activeModels.exists(id);
  }

  // registers new model or updates existing
  function modelRegister(
    bytes32 modelId,
    bytes32 ipfsCID,
    uint256 fee,
    uint256 addStake,
    address owner,
    string memory name,
    string[] memory tags
  ) public {
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
      timestamp: uint128(block.timestamp),
      ipfsCID: ipfsCID,
      owner: owner,
      name: name,
      tags: tags,
      isDeleted: false
    });

    emit ModelRegisteredUpdated(owner, modelId);
    s.token.transferFrom(msg.sender, address(this), addStake); // reverts with ERC20InsufficientAllowance
  }

  function modelDeregister(bytes32 id) public {
    Model storage model = s.modelMap[id];
    LibOwner._senderOrOwner(model.owner);

    s.activeModels.remove(id);
    model.isDeleted = true;
    uint256 stake = model.stake;

    emit ModelDeregistered(model.owner, id);
    s.token.transfer(model.owner, stake);
  }

  function modelSetMinStake(uint256 _minStake) public {
    LibOwner._onlyOwner();
    s.modelMinStake = _minStake;
    emit ModelMinStakeUpdated(s.modelMinStake);
  }

  function modelMinStake() public view returns (uint256) {
    return s.modelMinStake;
  }

  // TODO: implement these functions
  // function getModelsByOwner(address addr) public view returns (Model[] memory){
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
