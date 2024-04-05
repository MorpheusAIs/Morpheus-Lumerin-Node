// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "./KeySet.sol";
import { Marketplace } from "./Marketplace.sol";
import "hardhat/console.sol";

contract SessionRouter is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Session {
    bytes32 id;
    address user;
    address provider;
    bytes32 modelAgentId; // not clear
    uint256 budget;
    uint256 price;
    string closeoutReceipt;
    uint256 closeoutType;
    uint256 openedAt;
    uint256 closedAt;
  }

  error NotUserOrProvider();
  error NotUser();
  error NotSenderOrOwner();
  error NotEnoughBalance();

  // state
  // Number of seconds to delay the stake return when a user closes out a session using a user signed receipt.
  // not clear who is going to trigger this call
  int256 public stakeDelay;

  // dependencies
  Marketplace public marketplace;
  ERC20 public token;

  // storage option 1
  // KeySet.Set set;
  // mapping(bytes32 => Session) public map; // sessionId => Session

  // storage option 2
  Session[] public sessions;
  mapping(bytes32 => uint256) public map; // sessionId => index

  function initialize(address _token) public initializer {
    token = ERC20(_token);
    stakeDelay = 0;
    __Ownable_init();
  }

  function openSession(bytes32 bidId, uint256 budget) public returns (bytes32 sessionId){
    (address provider, bytes32 modelAgentId, uint256 amount, , ,)  = marketplace.map(bidId);

    sessionId = keccak256(abi.encodePacked(_msgSender(), provider, budget, block.number));
    sessions.push(Session({
      id: sessionId,
      user: _msgSender(),
      provider: provider,
      modelAgentId: modelAgentId,
      budget: budget,
      price: amount,
      closeoutReceipt: "",
      closeoutType: 0,
      openedAt: block.timestamp,
      closedAt: 0
    }));
    map[sessionId] = sessions.length - 1;

    // TODO: check allowed amount of virtual MOR and stake
    token.transferFrom(msg.sender, address(this), budget);

    return sessionId;
  }

  function closeSession(bytes32 sessionId, string memory receipt, string memory signature) public {
    Session storage session = sessions[map[sessionId]];
    if (session.user != _msgSender() && session.provider != _msgSender()) {
      revert NotUserOrProvider();
    }
    session.closeoutReceipt = receipt;
    session.closedAt = block.timestamp;

    // TODO: calculate it considering virtual MOR credits
    uint256 durationSeconds = session.closedAt - session.openedAt;
    uint256 cost = durationSeconds * session.price;
    uint256 refund = session.budget - cost;
    token.transfer(session.user, refund);
  }

  function deleteHistory(bytes32 sessionId) public {
    Session storage session = sessions[map[sessionId]];
    _senderOrOwner(session.user);
    session.user = address(0);
  }

  function setStakeDelay(int256 delay) public onlyOwner {
    stakeDelay = delay;
  }

  // funds related functions

  // Returns claimable balance by provider address.
  function getProviderClaimBalance(address providerAddr) public view returns (uint256) {
    return 0;
  }

  // transfers provider claimable balance to provider address.
  // set amount to 0 to claim all balance.
  function claimProviderBalance(uint256 amount, address to) public {
    if (amount > getProviderClaimBalance(_msgSender())) {
      revert NotEnoughBalance();
    }
    if (amount == 0) {
      amount = getProviderClaimBalance(_msgSender());
    }
    token.transfer(to, amount);
  }

  function getSpendBalance(address userAddress) public view returns (uint256) {
    uint256 userBalance = token.balanceOf(userAddress);
    return getTodaysBudget() * userBalance / getMintedMOR() - getTodaysSpend();
  }

  function getTodaysSpend() public view returns (uint256) {
    // todays closed sessions spend + today opened sessions budget
    return 0;
  }

  function getTodaysBudget() public view returns (uint256) {
    // 1% of Compute Balance
    return getComputeBalance() / 100;
  }

  function getComputeBalance() public view returns (uint256) {
    // call layer 1 contract to get daily compute balance
    return 0;
  }

  function getMintedMOR() public view returns (uint256) {
    // call layer 1 contract to get minted MOR
    return token.totalSupply();
  }

  function claimAmount(uint256 amount) public {
    // TODO: check if provider registered
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