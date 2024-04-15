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
  mapping(bytes32 => Bid) public map; // bidId = keccak256(provider, modelAgentId, nonce) => bid
  mapping(bytes32 => uint256) public providerModelAgentNonce; // keccak256(provider, modelAgentId) => last nonce
  mapping(address => uint256) public userStake; // user stakes
  uint256 public totalUserStake; // total user stakes that should be returned

  // indexes
  // active bids
  KeySet.Set activeBids; // all active bidIds
  mapping(address => KeySet.Set) providerActiveBids; // provider => active bidIds
  mapping(bytes32 => KeySet.Set) modelAgentActiveBids; // modelAgentId => active bidIds
  // all bids
  mapping(bytes32 => bytes32[]) modelAgentBids; // keccak256(provider, modelAgentId) => all bidIds
  mapping(address => bytes32[]) providerBids; // provider => all bidIds
  
  // events
  event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event Staked(address indexed user, uint256 amount);
  event Unstaked(address indexed user, uint256 amount);
  event Withdrawn(uint256 amount);
  event FeeUpdated(uint256 modelFee, uint256 agentFee);

  // errors
  error ModelOrAgentNotFound();
  error ActiveBidNotFound();
  error NotSenderOrOwner();
  error NotEnoughWithdrawableBalance();

  function initialize(
      address _token, 
      address modelRegistryAddr, 
      address agentRegistryAddr
    ) public initializer {
    __Ownable_init();
    token = ERC20(_token);
    modelRegistry = ModelRegistry(modelRegistryAddr);
    agentRegistry = AgentRegistry(agentRegistryAddr);
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

    // active indexes
    activeBids.insert(bidId);
    providerActiveBids[provider].insert(bidId);
    modelAgentActiveBids[modelAgentId].insert(bidId);
    
    // all indexes
    providerBids[provider].push(bidId);
    modelAgentBids[modelAgentId].push(bidId);

    emit BidPosted(provider, modelAgentId, nonce);
    
    token.transferFrom(_msgSender(), address(this), modelBidFee);
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

  function stake(uint256 amount) public {
    totalUserStake += amount;
    userStake[_msgSender()] += amount;

    emit Staked(_msgSender(), amount);
    token.transferFrom(_msgSender(), address(this), amount);
  }

  function unstake(uint256 amount) public {
    if (userStake[_msgSender()] < amount) {
      revert NotEnoughWithdrawableBalance();
    }
    totalUserStake -= amount;
    userStake[_msgSender()] -= amount;

    emit Unstaked(_msgSender(), amount);
    token.transfer(_msgSender(), amount);
  }

  function setBidFee(uint256 modelFee, uint256 agentFee) public onlyOwner {
    modelBidFee = modelFee;
    agentBidFee = agentFee;
    emit FeeUpdated(modelFee, agentFee);
  }

  function withdraw(address addr, uint256 amount) public onlyOwner {
    uint256 withdrawableBalance = token.balanceOf(address(this)) - totalUserStake;
    if (amount > withdrawableBalance) {
      revert NotEnoughWithdrawableBalance();
    }
    
    emit Withdrawn(amount);
    token.transfer(addr, withdrawableBalance);
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