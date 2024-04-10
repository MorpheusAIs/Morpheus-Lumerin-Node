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
    bytes32 modelAgentId;
    uint256 budget;
    uint256 price;
    string closeoutReceipt;
    uint256 closeoutType;
    uint256 openedAt;
    uint256 closedAt;
  }

  struct OnHold {
    uint256 amount;
    uint256 releaseAt;
  }

  // state
  // Number of seconds to delay the stake return when a user closes out a session using a user signed receipt.
  // not clear who is going to trigger this call
  int256 public stakeDelay;

  // dependencies
  Marketplace public marketplace;
  ERC20 public token;

  // storage option 2
  Session[] public sessions;
  mapping(bytes32 => uint256) public map; // sessionId => index
  mapping(address => uint256) public todaysSpend; // user address => spend today
  mapping(address => uint256) public userStake; // user address => balance
  mapping(address => OnHold[]) public providerOnHold; // user address => amount on hold - provider funds that put on hold 

  // events
  event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);
  event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);
  event Staked(address indexed userAddress, uint256 amount);
  event Unstaked(address indexed userAddress, uint256 amount);
  event ProviderClaimed(address indexed providerAddress, uint256 amount);

  // errors
  error NotUserOrProvider();
  error NotUser();
  error NotSenderOrOwner();
  error NotEnoughBalance();
  error NotEnoughComputeBalanceOrStake();

  function initialize(address _token) public initializer {
    __Ownable_init();
    token = ERC20(_token);
    stakeDelay = 0;
  }

  function openSession(bytes32 bidId, uint256 budget) public returns (bytes32 sessionId){
    if (budget > getSpendBalance(_msgSender()) + userStake[_msgSender()]) {
      revert NotEnoughComputeBalanceOrStake();
    }

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

    todaysSpend[_msgSender()] += budget;
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
    uint256 virtualMOR = getSpendBalance(session.user);
    if (cost > virtualMOR) {
      uint256 fromStake = cost - virtualMOR;
      userStake[session.user] -= fromStake;
      todaysSpend[session.user] += virtualMOR;
    } else {
      todaysSpend[session.user] -= session.budget - cost;
    }

    // TODO: decide whether put on hold or transfer to provider
    // put on hold
    providerOnHold[session.provider].push(OnHold({
      amount: cost,
      releaseAt: block.timestamp + uint256(stakeDelay)
    }));
    // or immediately transfer to provider
    // transferVirtualMOR(session.provider, cost);
  }

  function stake(address addr, uint256 amount) public senderOrOwner(addr){
    userStake[addr] += amount;
    token.transferFrom(addr, address(this), amount);
  }

  function unstake(address addr, uint256 amount, address sendToAddr) public senderOrOwner(addr){
    if (userStake[addr] < amount) {
      revert NotEnoughBalance();
    }
    userStake[addr] -= amount;
    token.transfer(sendToAddr, amount);
  }

  // funds related functions

  // Returns claimable balance by provider address.
  function getProviderClaimBalance(address providerAddr) public view returns (uint256) {
    uint256 balance = 0;
    // the only loop that is not avoidable
    for (uint i = 0; i < providerOnHold[providerAddr].length; i++) {
      if (providerOnHold[providerAddr][i].releaseAt < block.timestamp) {
        balance += providerOnHold[providerAddr][i].amount;
      }
    }
    return balance;
  }

  // transfers provider claimable balance to provider address.
  // set amount to 0 to claim all balance.
  function claimProviderBalanceV2(uint256 amount, address to) public {
    uint256 balance = 0;
    address sender = _msgSender();
    // the only loop that is not avoidable
    uint i = 0;

    OnHold[] storage onHoldEntries = providerOnHold[sender];
    while (i < onHoldEntries.length) {
      if (onHoldEntries[i].releaseAt < block.timestamp) {
        balance += onHoldEntries[i].amount;

        if (balance >= amount) {
          uint256 delta = balance - amount;
          onHoldEntries[i].amount = delta;
          token.transfer(to, amount);
          return;
        } 

        onHoldEntries[i] = onHoldEntries[onHoldEntries.length-1];
        onHoldEntries.pop();
      } else {
        i++;
      }
    }
    if (amount == 0) {
      amount = balance;
      token.transfer(to, amount);
      return;
    }

    revert NotEnoughBalance();
  }

  // aka stipend
  function getSpendBalance(address userAddress) public view returns (uint256) {
    return getTodaysBudget() * userStake[userAddress] / getMintedMOR() - todaysSpend[userAddress];
  }

  function getTodaysBudget() public view returns (uint256) {
    // 1% of Compute Balance
    return getComputeBalance() / 100;
  }

  function getComputeBalance() public view returns (uint256) {
    // call layer 1 contract to get daily compute balance
    return 0;
  }

  function transferVirtualMOR(address to, uint256 amount) public {
    address computeBalanceAddr = address(0); // the account that holds daily balance
    token.transferFrom(computeBalanceAddr, to, amount);
  }

  function getMintedMOR() public view returns (uint256) {
    // call layer 1 contract to get minted MOR
    return token.totalSupply();
  }

  function deleteHistory(bytes32 sessionId) public {
    Session storage session = sessions[map[sessionId]];
    _senderOrOwner(session.user);
    session.user = address(0);
  }

  function setStakeDelay(int256 delay) public onlyOwner {
    stakeDelay = delay;
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