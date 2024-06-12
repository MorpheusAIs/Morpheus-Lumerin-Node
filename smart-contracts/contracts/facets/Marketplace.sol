// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { AppStorage, Bid } from "../AppStorage.sol";
import { KeySet, AddressSet } from "../libraries/KeySet.sol";
import { LibOwner } from "../libraries/LibOwner.sol";

contract Marketplace {
  using KeySet for KeySet.Set;
  using AddressSet for AddressSet.Set;

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

  function bidMap(bytes32 bidId) external view returns (Bid memory) {
    return s.bidMap[bidId];
  }

  function getActiveBidsByProvider(address provider) external view returns (Bid[] memory) {
    KeySet.Set storage providerBidsSet = s.providerActiveBids[provider];
    uint256 length = providerBidsSet.count();

    Bid[] memory _bids = new Bid[](length);
    bytes32[] memory bidIds = new bytes32[](length);
    for (uint i = 0; i < providerBidsSet.count(); i++) {
      bytes32 id = providerBidsSet.keyAtIndex(i);
      bidIds[i] = id;
      _bids[i] = s.bidMap[id];
    }
    return _bids;
  }

  /// @notice returns active bids by model or agent id
  function getActiveBidsByModelAgent(bytes32 modelAgentId) external view returns (Bid[] memory) {
    KeySet.Set storage modelAgentBidsSet = s.modelAgentActiveBids[modelAgentId];
    uint256 length = modelAgentBidsSet.count();

    Bid[] memory _bids = new Bid[](length);
    bytes32[] memory bidIds = new bytes32[](length);
    for (uint i = 0; i < length; i++) {
      bytes32 id = modelAgentBidsSet.keyAtIndex(i);
      bidIds[i] = id;
      _bids[i] = s.bidMap[id];
    }
    return _bids;
  }

  /// @notice returns all bids by provider sorted from newest to oldest
  function getBidsByProvider(
    address provider,
    uint256 offset,
    uint8 limit
  ) external view returns (bytes32[] memory, Bid[] memory) {
    uint256 length = s.providerBids[provider].length;
    if (length < offset) {
      return (new bytes32[](0), new Bid[](0));
    }
    uint8 size = offset + limit > length ? uint8(length - offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    bytes32[] memory bidIds = new bytes32[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = s.providerBids[provider][index];
      bidIds[i] = id;
      _bids[i] = s.bidMap[id];
    }
    return (bidIds, _bids);
  }

  /// @notice returns all bids by model or agent Id sorted from newest to oldest
  function getBidsByModelAgent(
    bytes32 modelAgentId,
    uint256 offset,
    uint8 limit
  ) external view returns (bytes32[] memory, Bid[] memory) {
    uint256 length = s.modelAgentBids[modelAgentId].length;
    if (length < offset) {
      return (new bytes32[](0), new Bid[](0));
    }
    uint8 size = offset + limit > length ? uint8(length - offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    bytes32[] memory bidIds = new bytes32[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = s.modelAgentBids[modelAgentId][index];
      bidIds[i] = id;
      _bids[i] = s.bidMap[id];
    }
    return (bidIds, _bids);
  }

  /// @notice posts a new bid for a model
  function postModelBid(
    address providerAddr,
    bytes32 modelId,
    uint256 pricePerSecond
  ) external returns (bytes32 bidId) {
    LibOwner._senderOrOwner(providerAddr);
    if (!s.activeProviders.exists(providerAddr)) {
      revert ProviderNotFound();
    }
    if (!s.activeModels.exists(modelId)) {
      revert ModelOrAgentNotFound();
    }

    return postModelAgentBid(providerAddr, modelId, pricePerSecond);
  }

  function postModelAgentBid(
    address provider,
    bytes32 modelAgentId,
    uint256 pricePerSecond
  ) internal returns (bytes32 bidId) {
    // remove old bid

    // TEST IT if it increments nonce correctly
    uint256 nonce = s.providerModelAgentNonce[keccak256(abi.encodePacked(provider, modelAgentId))]++;
    if (nonce > 0) {
      deleteModelAgentBid(keccak256(abi.encodePacked(provider, modelAgentId, nonce - 1)));
    }

    bidId = keccak256(abi.encodePacked(provider, modelAgentId, nonce));

    s.bidMap[bidId] = Bid({
      provider: provider,
      modelAgentId: modelAgentId,
      pricePerSecond: pricePerSecond,
      nonce: nonce,
      createdAt: uint128(block.timestamp),
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

  /// @notice deletes a bid
  function deleteModelAgentBid(bytes32 bidId) public {
    Bid storage bid = s.bidMap[bidId];
    if (bid.createdAt == 0 || bid.deletedAt != 0) {
      revert ActiveBidNotFound();
    }

    LibOwner._senderOrOwner(bid.provider);

    bid.deletedAt = uint128(block.timestamp);
    // indexes update
    s.activeBids.remove(bidId);
    s.providerActiveBids[bid.provider].remove(bidId);
    s.modelAgentActiveBids[bid.modelAgentId].remove(bidId);

    emit BidDeleted(bid.provider, bid.modelAgentId, bid.nonce);
  }

  /// @notice sets a bid fee
  function setBidFee(uint256 _bidFee) external {
    LibOwner._onlyOwner();
    s.bidFee = _bidFee;
    emit FeeUpdated(_bidFee);
  }

  /// @notice returns the bid fee
  function bidFee() external view returns (uint256) {
    return s.bidFee;
  }

  /// @notice withdraws the fee balance (OWNER ONLY)
  function withdraw(address addr, uint256 amount) external {
    LibOwner._onlyOwner();
    if (amount > s.feeBalance) {
      revert NotEnoughBalance();
    }
    // emits ERC-20 transfer event
    s.feeBalance -= amount;
    s.token.transfer(addr, amount);
  }
}
