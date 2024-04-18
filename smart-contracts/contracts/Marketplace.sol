// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "./KeySet.sol";
import { ModelRegistry } from "./ModelRegistry.sol";
import { AgentRegistry } from "./AgentRegistry.sol";
import { ProviderRegistry } from './ProviderRegistry.sol';

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
  ProviderRegistry public providerRegistry;

  // storage
  mapping(bytes32 => Bid) public map; // bidId = keccak256(provider, modelAgentId, nonce) => bid
  mapping(bytes32 => uint256) public providerModelAgentNonce; // keccak256(provider, modelAgentId) => last nonce

  // indexes - active bids
  KeySet.Set activeBids; // all active bidIds
  mapping(address => KeySet.Set) providerActiveBids; // provider => active bidIds
  mapping(bytes32 => KeySet.Set) modelAgentActiveBids; // modelAgentId => active bidIds
  
  // indexes - all bids
  mapping(bytes32 => bytes32[]) modelAgentBids; // keccak256(provider, modelAgentId) => all bidIds
  mapping(address => bytes32[]) providerBids; // provider => all bidIds
  
  // events
  event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event FeeUpdated(uint256 modelFee, uint256 agentFee);

  // errors
  error ModelOrAgentNotFound();
  error ActiveBidNotFound();
  error NotSenderOrOwner();
  error NotEnoughWithdrawableBalance();

  function initialize(
      address _token, 
      address modelRegistryAddr,
      address providerRegistryAddr
    ) public initializer {
    __Ownable_init();
    token = ERC20(_token);
    modelRegistry = ModelRegistry(modelRegistryAddr);
    providerRegistry = ProviderRegistry(providerRegistryAddr);
  }

  // returns active bids by provider
  function getActiveBidsByProvider(address provider) public view returns (Bid[] memory) {
    KeySet.Set storage providerBidsSet = providerActiveBids[provider];
    Bid[] memory _bids = new Bid[](providerBidsSet.count());
    for (uint i = 0; i < providerBidsSet.count(); i++) {
      _bids[i] = map[providerBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  // returns active bids by model agent
  function getActiveBidsByModelAgent(bytes32 modelAgentId) public view returns (Bid[] memory) {
    KeySet.Set storage modelAgentBidsSet = modelAgentActiveBids[modelAgentId];
    Bid[] memory _bids = new Bid[](modelAgentBidsSet.count());
    for (uint i = 0; i < modelAgentBidsSet.count(); i++) {
      _bids[i] = map[modelAgentBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  // returns all bids by provider sorted from newest to oldest
  function getBidsByProvider(address provider, uint256 offset, uint8 limit) public view returns (Bid[] memory) {
    uint256 length = providerBids[provider].length;
    if (length < offset){
      return new Bid[](0);
    }
    uint8 size = offset + limit > length ? uint8(length-offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = providerBids[provider][index];
      _bids[i]=map[id];
    }
    return _bids;
  }

  // returns all bids by provider sorted from newest to oldest
  function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) public view returns (Bid[] memory) {
    uint256 length = modelAgentBids[modelAgentId].length;
    if (length < offset){
      return new Bid[](0);
    }
    uint8 size = offset + limit > length ? uint8(length-offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = modelAgentBids[modelAgentId][index];
      _bids[i]=map[id];
    }
    return _bids;
  }

  function postModelBid(address providerAddr, bytes32 modelId, uint256 amount) public  senderOrOwner(providerAddr) returns (bytes32 bidId){
    if (!providerRegistry.exists(providerAddr)){
      revert ModelOrAgentNotFound();
    }
    if (!modelRegistry.exists(modelId)){
      revert ModelOrAgentNotFound();
    }

    return postModelAgentBid(providerAddr, modelId, amount);
  }

  function postModelAgentBid(address provider, bytes32 modelAgentId, uint256 amount) internal returns (bytes32 bidId){
    // remove old bid

    // TEST IT if it increments nonce correctly
    uint256 nonce = providerModelAgentNonce[keccak256(abi.encodePacked(provider, modelAgentId))]++;
    if (nonce > 0) {
      deleteModelAgentBid(keccak256(abi.encodePacked(provider, modelAgentId, nonce-1)));
    }
    
    bidId = keccak256(abi.encodePacked(provider, modelAgentId, nonce));

    map[bidId] = Bid({
      provider: provider,
      modelAgentId: modelAgentId,
      amount: amount,
      nonce: nonce,
      createdAt: block.timestamp,
      deletedAt: 0
    });

    // active indexes
    activeBids.insert(bidId);
    providerActiveBids[provider].insert(bidId);
    modelAgentActiveBids[modelAgentId].insert(bidId);
    
    // all indexes
    providerBids[provider].push(bidId);
    modelAgentBids[modelAgentId].push(bidId);

    emit BidPosted(provider, modelAgentId, nonce);
    
    token.transferFrom(_msgSender(), address(this), modelBidFee);
    return bidId;
  }

  function deleteModelAgentBid(bytes32 bidId) public {
    Bid storage bid = map[bidId];
    if (bid.createdAt == 0 || bid.deletedAt != 0) {
      revert ActiveBidNotFound();
    }

    _senderOrOwner(bid.provider);

    bid.deletedAt = block.timestamp;
    // indexes update
    activeBids.remove(bidId);
    providerActiveBids[bid.provider].remove(bidId);
    modelAgentActiveBids[bid.modelAgentId].remove(bidId);

    emit BidDeleted(bid.provider, bid.modelAgentId, bid.nonce);
  }

  function setBidFee(uint256 modelFee, uint256 agentFee) public onlyOwner {
    modelBidFee = modelFee;
    agentBidFee = agentFee;
    emit FeeUpdated(modelFee, agentFee);
  }

  function withdraw(address addr, uint256 amount) public onlyOwner {
    if (amount == 0) {
      amount = token.balanceOf(address(this));
    }
    // emits ERC-20 transfer event
    // errors with unsufficient balance if amount too big
    token.transfer(addr, amount);
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
}