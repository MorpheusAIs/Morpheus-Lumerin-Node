// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { MessageHashUtils } from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { ModelRegistry } from "./ModelRegistry.sol";
import { ProviderRegistry } from './ProviderRegistry.sol';
import { AppStorage, Bid } from "../AppStorage.sol";
import { KeySet } from "../libraries/KeySet.sol";
import { LibOwner } from '../libraries/LibOwner.sol';

contract Marketplace {
  using KeySet for KeySet.Set;

  AppStorage internal s;

  event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event FeeUpdated(uint256 bidFee);

  error ProviderNotFound();
  error ModelOrAgentNotFound();
  error ActiveBidNotFound();
  error BidNotFound();
  error BidTaken();
  error NotEnoughBalance();

  function bidMap(bytes32 bidId) public view returns (Bid memory) {
    return s.bidMap[bidId];
  }

  function getActiveBidsByProvider(address provider) public view returns (Bid[] memory) {
    KeySet.Set storage providerBidsSet = s.providerActiveBids[provider];
    Bid[] memory _bids = new Bid[](providerBidsSet.count());
    for (uint i = 0; i < providerBidsSet.count(); i++) {
      _bids[i] = s.bidMap[providerBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  // returns active bids by model agent
  function getActiveBidsByModelAgent(bytes32 modelAgentId) public view returns (Bid[] memory) {
    KeySet.Set storage modelAgentBidsSet = s.modelAgentActiveBids[modelAgentId];
    Bid[] memory _bids = new Bid[](modelAgentBidsSet.count());
    for (uint i = 0; i < modelAgentBidsSet.count(); i++) {
      _bids[i] = s.bidMap[modelAgentBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  // returns all bids by provider sorted from newest to oldest
  function getBidsByProvider(address provider, uint256 offset, uint8 limit) public view returns (Bid[] memory) {
    uint256 length = s.providerBids[provider].length;
    if (length < offset){
      return new Bid[](0);
    }
    uint8 size = offset + limit > length ? uint8(length-offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = s.providerBids[provider][index];
      _bids[i]=s.bidMap[id];
    }
    return _bids;
  }

  // returns all bids by provider sorted from newest to oldest
  function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) public view returns (Bid[] memory) {
    uint256 length = s.modelAgentBids[modelAgentId].length;
    if (length < offset){
      return new Bid[](0);
    }
    uint8 size = offset + limit > length ? uint8(length-offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = s.modelAgentBids[modelAgentId][index];
      _bids[i]=s.bidMap[id];
    }
    return _bids;
  }

  function postModelBid(address providerAddr, bytes32 modelId, uint256 pricePerSecond) public returns (bytes32 bidId){
    LibOwner._senderOrOwner(providerAddr);
    if (s.providerMap[providerAddr].isDeleted){
      revert ProviderNotFound();
    }
    if (s.modelMap[modelId].isDeleted){
      revert ModelOrAgentNotFound();
    }

    return postModelAgentBid(providerAddr, modelId, pricePerSecond);
  }

  function postModelAgentBid(address provider, bytes32 modelAgentId, uint256 pricePerSecond) internal returns (bytes32 bidId){
    // remove old bid

    // TEST IT if it increments nonce correctly
    uint256 nonce = s.providerModelAgentNonce[keccak256(abi.encodePacked(provider, modelAgentId))]++;
    if (nonce > 0) {
      deleteModelAgentBid(keccak256(abi.encodePacked(provider, modelAgentId, nonce-1)));
    }

    bidId = keccak256(abi.encodePacked(provider, modelAgentId, nonce));

    s.bidMap[bidId] = Bid({
      provider: provider,
      modelAgentId: modelAgentId,
      pricePerSecond: pricePerSecond,
      nonce: nonce,
      createdAt: block.timestamp,
      deletedAt: 0
    });

    // active indexes
    s.activeBids.insert(bidId);
    s.providerActiveBids[provider].insert(bidId);
    s.modelAgentActiveBids[modelAgentId].insert(bidId);
    
    // all indexes
    s.providerBids[provider].push(bidId);
    s.modelAgentBids[modelAgentId].push(bidId);

    emit BidPosted(provider, modelAgentId, nonce);
    
    s.token.transferFrom(msg.sender, address(this), s.bidFee);
    s.feeBalance += s.bidFee;

    return bidId;
  }

  function deleteModelAgentBid(bytes32 bidId) public {
    Bid storage bid = s.bidMap[bidId];
    if (bid.createdAt == 0 || bid.deletedAt != 0) {
      revert ActiveBidNotFound();
    }

    LibOwner._senderOrOwner(bid.provider);

    bid.deletedAt = block.timestamp;
    // indexes update
    s.activeBids.remove(bidId);
    s.providerActiveBids[bid.provider].remove(bidId);
    s.modelAgentActiveBids[bid.modelAgentId].remove(bidId);

    emit BidDeleted(bid.provider, bid.modelAgentId, bid.nonce);
  }

  function setBidFee(uint256 _bidFee) public {
    LibOwner._onlyOwner();
    s.bidFee = _bidFee;
    emit FeeUpdated(_bidFee);
  }

  function bidFee() public view returns (uint256) {
    return s.bidFee;
  }

  function withdraw(address addr, uint256 amount) public {
    LibOwner._onlyOwner();
    if (amount > s.feeBalance) {
      revert NotEnoughBalance();
    }
    // emits ERC-20 transfer event
    s.feeBalance -= amount;
    s.token.transfer(addr, amount);
  }
}