// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {KeySet} from "./KeySet.sol";
import "hardhat/console.sol";

contract ModelRegistry is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Model {
    bytes32 ipfsCID;    // https://docs.ipfs.tech/concepts/content-addressing/#what-is-a-cid
    uint256 fee;
    uint256 stake;
    address owner;
    string name;        // limit name length
    string[] tags;      // TODO: limit tags amount
    uint128 timestamp;
    bool isDeleted;
  }

  error StakeTooLow();
  error NotSenderOrOwner();
  error ModelNotFound();

  event RegisteredUpdated(address indexed owner, bytes32 indexed modelId);
  event Deregistered(address indexed owner, bytes32 indexed modelId);
  event MinStakeUpdated(uint256 newStake);

  // state
  uint256 public minStake;
  ERC20 public token;

  // storage
  mapping(bytes32 => Model) public map; // modelId => Model
  bytes32[] public models; // all model ids
  KeySet.Set activeModels; // active model ids
  // mapping(address => bytes32[]) public modelsByOwner; // owner to modelIds

  function initialize(address _token) public initializer {
    __Ownable_init();
    token = ERC20(_token);
  }

  function getIds() public view returns (bytes32[] memory){
    return activeModels.keys();
  }

  function getCount() public view returns(uint count) {
    return activeModels.count();
  }

  function getAll() public view returns (Model[] memory){
    Model[] memory _models = new Model[](activeModels.count());
    for (uint i = 0; i < activeModels.count(); i++) {
      _models[i] = map[activeModels.keyAtIndex(i)];
    }
    return _models;
  }

  function getByIndex(uint index) public view returns(bytes32 modelId, Model memory model) {
      modelId = activeModels.keyAtIndex(index);
      return (modelId, map[modelId]);
  }

  function exists(bytes32 id) public view returns (bool) {
    return activeModels.exists(id);
  }

  // registers new model or updates existing
  function register(bytes32 modelId, bytes32 ipfsCID,  uint256 fee, uint256 addStake, address owner, string memory name, string[] memory tags) public senderOrOwner(owner){
    Model memory model = map[modelId];
    uint256 newStake = model.stake + addStake;
    if (newStake < minStake) {
      revert StakeTooLow();
    }
    if (model.stake == 0) {
      activeModels.insert(modelId);
      models.push(modelId);
    } else {
      _senderOrOwner(map[modelId].owner);
    }

    map[modelId] = Model({
      fee: fee,
      stake: newStake,
      timestamp: uint128(block.timestamp),
      ipfsCID: ipfsCID,
      owner: owner,
      name: name,
      tags: tags,
      isDeleted: false
    });

    emit RegisteredUpdated(owner, modelId);
    token.transferFrom(_msgSender(), address(this), addStake); // reverts with ERC20InsufficientAllowance
  }

  function deregister(bytes32 id) public {
    Model storage model = map[id];
    _senderOrOwner(model.owner);

    activeModels.remove(id);
    model.isDeleted = true;
    uint256 stake = model.stake;

    emit Deregistered(model.owner, id);
    token.transfer(model.owner, stake);
  }

  function setMinStake(uint256 _minStake) public onlyOwner {
    minStake = _minStake;
    emit MinStakeUpdated(minStake);
  }

  modifier senderOrOwner(address addr) {
    _senderOrOwner(addr);
    _;
  }

  function _senderOrOwner(address resourceOwner) internal view {
    if (_msgSender() != resourceOwner && _msgSender() != owner()) {
      revert NotSenderOrOwner();
    }
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