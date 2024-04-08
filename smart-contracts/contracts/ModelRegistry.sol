// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {KeySet} from "./KeySet.sol";
import "hardhat/console.sol";

contract ModelRegistry is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Model {
    bytes32 modelId;
    bytes32 ipfsCID;    // https://docs.ipfs.tech/concepts/content-addressing/#what-is-a-cid
    uint256 fee;
    uint256 stake;
    uint256 timestamp;
    address owner;
    string name;        // limit name length
    string[] tags;      // TODO: limit tags amount
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
  KeySet.Set set;
  mapping(bytes32 => Model) public map;
  // mapping(address => bytes32[]) public modelsByOwner; // owner to modelIds

  function initialize(address _token) public initializer {
    token = ERC20(_token);
    __Ownable_init();
  }

  function getIds() public view returns (bytes32[] memory){
    return set.keys();
  }

  function getCount() public view returns(uint count) {
    return set.count();
  }

  function getAll() public view returns (Model[] memory){
    Model[] memory _models = new Model[](set.count());
    for (uint i = 0; i < set.count(); i++) {
      _models[i] = map[set.keyAtIndex(i)];
    }
    return _models;
  }

  function getByIndex(uint index) public view returns(Model memory model) {
    return map[set.keyAtIndex(index)];
  }

  function exists(bytes32 id) public view returns (bool) {
    return set.exists(id);
  }

  // registers new model or updates existing
  function register(uint256 addStake, uint256 fee, bytes32 ipfsCID, bytes32 modelId, address owner, string memory name, string[] memory tags) public senderOrOwner(owner){
    uint256 stake = map[modelId].stake;
    uint256 newStake = stake + addStake;
    if (newStake < minStake) {
      revert StakeTooLow();
    }

    if (stake == 0) {
      set.insert(modelId);
    } else {
      _senderOrOwner(map[modelId].owner);
    }

    map[modelId] = Model({
      fee: fee,
      stake: newStake,
      timestamp: block.timestamp,
      ipfsCID: ipfsCID,
      modelId: modelId,
      owner: owner,
      name: name,
      tags: tags
    });

    emit RegisteredUpdated(owner, modelId);
    token.transferFrom(_msgSender(), address(this), addStake); // reverts with ERC20InsufficientAllowance
  }

  function deregister(bytes32 id) public {
    address owner = map[id].owner;
    _senderOrOwner(owner);
    
    set.remove(id);
    uint256 stake = map[id].stake;
    delete map[id];

    emit Deregistered(owner, id);
    token.transfer(owner, stake);
  }

  function setMinStake(uint256 _minStake) public onlyOwner {
    minStake = _minStake;
    emit MinStakeUpdated(minStake);
  }

  modifier senderOrOwner(address addr) {
    _senderOrOwner(addr);
    _;
  }

  function _senderOrOwner(address addr) internal view {
    if (addr != _msgSender() && addr != owner()) {
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