// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "./KeySet.sol";
import { ModelRegistry } from "./ModelRegistry.sol";
import { AgentRegistry } from "./AgentRegistry.sol";

contract Marketplace is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Bid {
    address provider;
    bytes32 modelAgentId;
    uint256 amount;
    uint256 nonce;
    uint256 createdAt;
    uint256 deletedAt;
  }

  // state
  ERC20 public token;
  uint256 public modelBidFee;
  uint256 public agentBidFee;

  // dependencies
  ModelRegistry public modelRegistry;
  AgentRegistry public agentRegistry;

  // storage
  
  // option 1
  // can't get all bids for a provider
  // mapping(address => mapping(bytes32 => Bid[])) public bids; // provider => modelAgentId => bid[], where index is nonce
  
  // option 2
  // not keeping data for old bids
  // mapping(bytes32 => Bid) public bids2; // bidId = keccak256(provider, modelAgentId) => bid
  // mapping(address => KeySet.Set) providerBids; // provider => bidIds
  // mapping(bytes32 => KeySet.Set) modelAgentBids; // modelAgentId => bidIds

  // option 3
  // complex, keeps all bids
  // old bids are accessible only by bidId = keccak256(provider, modelAgentId, nonce)
  mapping(bytes32 => Bid) public map; // bidId = keccak256(provider, modelAgentId, nonce) => bid
  KeySet.Set set; // keeps all active bidIds
  mapping(address => KeySet.Set) providerBids; // provider => bidIds, only stores active bids
  mapping(bytes32 => KeySet.Set) modelAgentBids; // modelAgentId => bidIds, only stores active bids
  mapping(bytes32 => uint256) providerModelAgentNonce; // keccak256(provider, modelAgentId) => last nonce

  // errors
  error ModelOrAgentNotFound();
  error NoActiveBid();
  error NotSenderOrOwner();

  function initialize(
      address _token, 
      address modelRegistryAddr, 
      address agentRegistryAddr
    ) public initializer {

    token = ERC20(_token);
    modelRegistry = ModelRegistry(modelRegistryAddr);
    agentRegistry = AgentRegistry(agentRegistryAddr);
    __Ownable_init();
  }
  
  function getAll() public view returns (Bid[] memory) {
    Bid[] memory _bids = new Bid[](set.count());
    for (uint i = 0; i < set.count(); i++) {
      _bids[i] = map[set.keyAtIndex(i)];
    }
    return _bids;
  }

  function getByIndex(uint index) public view returns (Bid memory bid) {
    return map[set.keyAtIndex(index)];
  }

  function getBidsByProvider(address provider) public view returns (Bid[] memory) {
    KeySet.Set storage providerBidsSet = providerBids[provider];
    Bid[] memory _bids = new Bid[](providerBidsSet.count());
    for (uint i = 0; i < providerBidsSet.count(); i++) {
      _bids[i] = map[providerBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  function getBidsByModelAgent(bytes32 modelAgentId) public view returns (Bid[] memory) {
    KeySet.Set storage modelAgentBidsSet = modelAgentBids[modelAgentId];
    Bid[] memory _bids = new Bid[](modelAgentBidsSet.count());
    for (uint i = 0; i < modelAgentBidsSet.count(); i++) {
      _bids[i] = map[modelAgentBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  function getBidById(bytes32 bidId) public view returns (Bid memory) {
    return map[bidId];
  }

  function postModelBid(address provider, bytes32 modelId, uint256 amount) public senderOrOwner(provider){
    if (!modelRegistry.exists(modelId)){
      revert ModelOrAgentNotFound();
    }

    postModelAgentBid(provider, modelId, amount);
  }

  function postAgentBid(address provider, bytes32 agentId, uint256 amount) public senderOrOwner(provider){
    if (!agentRegistry.exists(agentId)){
      revert ModelOrAgentNotFound();
    }

    postModelAgentBid(provider, agentId, amount);
  }

  function postModelAgentBid(address provider, bytes32 modelAgentId, uint256 amount) internal {
    // TEST IT if it increments nonce correctly
    uint256 nonce = providerModelAgentNonce[keccak256(abi.encodePacked(provider, modelAgentId))]++;
    bytes32 bidId = keccak256(abi.encodePacked(provider, modelAgentId, nonce));

    map[bidId] = Bid({
      provider: provider,
      modelAgentId: modelAgentId,
      amount: amount,
      nonce: nonce,
      createdAt: block.timestamp,
      deletedAt: 0
    });

    set.insert(bidId);
    providerBids[provider].insert(bidId);
    modelAgentBids[modelAgentId].insert(bidId);

    token.transferFrom(_msgSender(), address(this), modelBidFee);
  }

  function deleteModelAgentBid(bytes32 bidId) public {
    Bid storage bid = map[bidId];
    if (bid.createdAt == 0 || bid.deletedAt != 0) {
      revert NoActiveBid();
    }

    _senderOrOwner(bid.provider);

    bid.deletedAt = block.timestamp;
    set.remove(bidId);
    providerBids[bid.provider].remove(bidId);
    modelAgentBids[bid.modelAgentId].remove(bidId);
  }

  function setModelBidFee(uint256 fee) public onlyOwner {
    modelBidFee = fee;
  }

  function setAgentBidFee(uint256 fee) public onlyOwner {
    agentBidFee = fee;
  }

  function withdraw(address addr) public onlyOwner {
    token.transfer(addr, token.balanceOf(address(this)));
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
}