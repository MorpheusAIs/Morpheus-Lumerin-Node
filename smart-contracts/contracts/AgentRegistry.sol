// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {KeySet} from "./KeySet.sol";
import "hardhat/console.sol";

contract ModelRegistry is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Agent {
    uint256 fee;
    uint256 stake;
    uint256 timestamp;
    address owner;
    bytes32 uuid;
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
  mapping(bytes32 => Agent) public map;

  function initialize(address _token) public initializer {
    token = ERC20(_token);
    __Ownable_init();
  }

  function getIds() public view returns (bytes32[] memory){
    return set.keys();
  }

  function getAll() public view returns (Agent[] memory){
    Agent[] memory _agents = new Agent[](set.count());
    for (uint i = 0; i < set.count(); i++) {
      _agents[i] = map[set.keyAtIndex(i)];
    }
    return _agents;
  }

  // registers new or updates existing
  function register(uint256 fee, uint256 timestamp, address owner, bytes32 uuid, string memory name, string[] memory tags) public senderOrOwner(owner){
    uint256 amount = token.allowance(owner, address(this));
    uint256 stake = map[uuid].stake;
    if (amount + stake < minStake) {
      revert StakeTooLow();
    }
    if (stake == 0) {
      set.insert(uuid);
    } 
    map[uuid] = Agent({
      fee: fee,
      stake: amount + stake,
      timestamp: timestamp,
      owner: owner,
      uuid: uuid,
      name: name,
      tags: tags
    });
  }

  // avoid loop this by using pointer pattern
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

  // function getModelsByOwner
  // use modelByOwner mapping instead

  // function getModelsByOwner(address addr) public view returns (Model[] memory){
  // Model[] memory _models = new Model[](modelIds.length);
  // for (uint i = 0; i < modelIds.length; i++) {
  //   if (models[modelIds[i]].owner == addr) {
  //     _models[i] = models[modelIds[i]];
  //   }
  // }
  // return _models;
  // }

  // function getModelTypes -- to be implemented when types are defined
  // function getModelsByType

  // function getProvider(address)
  // just use provider mapping contract.provider[address]

  // function getStakeReq()
  // just use minStake variable

  // function updateProvider()
  // use registerProvider instead

  // function owner()
  // inherited from OwnableUpgradeable

  // function transferOwnership()
  // inherited from OwnableUpgradeable
}