// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {KeySet} from "./KeySet.sol";
import "hardhat/console.sol";

contract ModelRegistry is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Model {
    uint256 fee;
    uint256 stake;
    uint256 timestamp;
    bytes32 ipfsCID;    // https://docs.ipfs.tech/concepts/content-addressing/#what-is-a-cid
    bytes32 uuid;
    address owner;
    string name;        // limit name length
    string[] tags;      // TODO: limit tags amount
  }

  error StakeTooLow();
  error NotSenderOrOwner();
  error ModelNotFound();

  event ModelRegisteredUpdated(address indexed id);
  event ModelDeregistered(address indexed id);

  // state
  uint256 public minStake;
  ERC20 public token;

  // storage
  KeySet.Set set;
  mapping(bytes32 => Model) public map;
  // mapping(address => bytes32[]) public modelsByOwner; // owner to uuids

  function initialize(address _token) public initializer {
    token = ERC20(_token);
    __Ownable_init();
  }

  function getModelIds() public view returns (bytes32[] memory){
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

  // registers new model or updates existing
  function register(uint256 fee, uint256 timestamp, bytes32 ipfsCID, bytes32 uuid, address owner, string memory name, string[] memory tags) public senderOrOwner(owner){
    uint256 amount = token.allowance(owner, address(this));
    uint256 stake = map[uuid].stake;
    if (amount + stake < minStake) {
      revert StakeTooLow();
    }
    if (stake == 0) {
      set.insert(uuid);
    } 
    map[uuid] = Model({
      fee: fee,
      stake: amount + stake,
      timestamp: timestamp,
      ipfsCID: ipfsCID,
      uuid: uuid,
      owner: owner,
      name: name,
      tags: tags
    });
  }

  function deregister(bytes32 id) public {
    _senderOrOwner(map[id].owner);
    set.remove(id);
    uint256 stake = map[id].stake;
    delete map[id];
    token.transfer(map[id].owner, stake);
  }

  function setMinStake(uint256 _minStake) public onlyOwner {
    minStake = _minStake;
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