// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "../libraries/KeySet.sol";

contract AgentRegistry is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Agent {
    bytes32 agentId;
    uint256 fee;
    uint256 stake;
    uint256 timestamp;
    address owner;
    string name; // limit name length
    string[] tags; // TODO: limit tags amount
  }

  error StakeTooLow();
  error NotSenderOrOwner();
  error ModelNotFound();

  event RegisteredUpdated(address indexed owner, bytes32 indexed agentId);
  event Deregistered(address indexed owner, bytes32 indexed agentId);
  event MinStakeUpdated(uint256 newStake);

  // state
  uint256 public minStake;
  ERC20 public token;

  // model storage
  KeySet.Set private set;
  mapping(bytes32 => Agent) public map;

  function initialize(address _token) public initializer {
    token = ERC20(_token);
    __Ownable_init();
  }

  function getIds() public view returns (bytes32[] memory) {
    return set.keys();
  }

  function getAll() public view returns (Agent[] memory) {
    Agent[] memory _agents = new Agent[](set.count());
    for (uint i = 0; i < set.count(); i++) {
      _agents[i] = map[set.keyAtIndex(i)];
    }
    return _agents;
  }

  function exists(bytes32 id) public view returns (bool) {
    return set.exists(id);
  }

  // registers new or updates existing
  function register(
    uint256 addStake,
    uint256 fee,
    address owner,
    bytes32 agentId,
    string memory name,
    string[] memory tags
  ) public senderOrOwner(owner) {
    uint256 stake = map[agentId].stake;
    uint256 newStake = stake + addStake;
    if (newStake < minStake) {
      revert StakeTooLow();
    }

    if (stake == 0) {
      set.insert(agentId);
    } else {
      _senderOrOwner(map[agentId].owner);
    }

    map[agentId] = Agent({
      fee: fee,
      stake: newStake,
      timestamp: block.timestamp,
      owner: owner,
      agentId: agentId,
      name: name,
      tags: tags
    });

    emit RegisteredUpdated(owner, agentId);
    token.transferFrom(_msgSender(), address(this), addStake); // reverts with ERC20InsufficientAllowance
  }

  // avoid loop this by using pointer pattern
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

  function _senderOrOwner(address resourceOwner) internal view {
    if (_msgSender() != resourceOwner && _msgSender() != owner()) {
      revert NotSenderOrOwner();
    }
  }

  // TODO: implement the following functions using mapping
  // function getAgentsByOwner(address addr) public view returns (Model[] memory){
  // Model[] memory _models = new Model[](modelIds.length);
  // for (uint i = 0; i < modelIds.length; i++) {
  //   if (models[modelIds[i]].owner == addr) {
  //     _models[i] = models[modelIds[i]];
  //   }
  // }
  // return _models;
  // }
}
