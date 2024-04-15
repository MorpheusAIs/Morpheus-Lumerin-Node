// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
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
    uint256 releaseAt; // in epoch seconds TODO: consider using hours to reduce storage cost
  }

  // state
  // Number of seconds to delay the stake return when a user closes out a session using a user signed receipt.
  // not clear who is going to trigger this call
  int256 public stakeDelay;

  // dependencies
  Marketplace public marketplace;
  IERC20 public token;

  // arguments for getPeriodReward call
  // address public constant distributionContractAddr = address(0x0);
  // uint32 public constant distributionRewardStartTime = 1707350400; // ephochSeconds Feb 8 2024 00:00:00
  // uint8 public constant distributionPoolId = 3;

  // storage
  Session[] public sessions;
  mapping(bytes32 => uint256) public map; // sessionId => index
  mapping(address => OnHold) public todaysSpend; // user address => spend today
  
  mapping(address => uint256) public userStake; // user address => balance
  mapping(address => OnHold) public userStakeOnHold; // user address => amount on hold - user funds that put on hold
  // mapping(address => OnHold[]) public providerOnHold; // user address => amount on hold - provider funds that put on hold 

  // constants
  uint32 constant DAY = 24*60*60; // 1 day

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
    token = IERC20(_token);
    stakeDelay = 0;
  }

  function openSession(bytes32 bidId, uint256 budget) public returns (bytes32 sessionId){
    address sender = _msgSender();
    uint256 virtualMORBalance = getUserVirtualMORBalance(sender);
    if (budget > virtualMORBalance + userStake[sender]) {
      revert NotEnoughComputeBalanceOrStake();
    }
    uint256 vMORToBeUsed = min(budget, virtualMORBalance);
    setTodaysSpend(sender, getTodaysSpend(sender) + vMORToBeUsed);
    setStakeOnHold(sender, getStakeOnHold(sender) + getStakeCorrespondingToVirtualMOR(vMORToBeUsed));

    (address provider, bytes32 modelAgentId, uint256 amount, , ,)  = marketplace.map(bidId);

    sessionId = keccak256(abi.encodePacked(sender, provider, budget, block.number));
    sessions.push(Session({
      id: sessionId,
      user: sender,
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

    return sessionId;
  }

  function closeSession(bytes32 sessionId, string memory receipt, string memory signature) public {
    Session storage session = sessions[map[sessionId]];
    if (session.user != _msgSender() && session.provider != _msgSender()) {
      revert NotUserOrProvider();
    }

    session.closeoutReceipt = receipt;
    session.closedAt = block.timestamp;

    uint256 durationSeconds = session.closedAt - session.openedAt;
    uint256 cost = durationSeconds * session.price;

    uint256 virtualMOR = getUserVirtualMORBalance(session.user);
    if (cost > virtualMOR) {
      uint256 spentStake = cost - virtualMOR;
      userStake[session.user] -= spentStake;
    } else {
      setTodaysSpend(session.user, getTodaysSpend(session.user) - session.budget + cost);
    }

    // TODO: decide whether put on hold or transfer to provider
    // wait for dispute design to be finalized
    //
    // put on hold
    //
    // providerOnHold[session.provider].push(OnHold({
    //   amount: cost,
    //   releaseAt: block.timestamp + uint256(stakeDelay)
    // }));
    //
    // or immediately transfer to provider
    //
    transferVirtualMOR(session.provider, cost);
  }

  function stake(address addr, uint256 amount) public senderOrOwner(addr){
    userStake[addr] += amount;
    token.transferFrom(addr, address(this), amount);
  }

  function unstake(address addr, uint256 amount, address sendToAddr) public senderOrOwner(addr){
    OnHold memory stakeOnHold = userStakeOnHold[addr];
    if (stakeOnHold.releaseAt < block.timestamp) {
      stakeOnHold.amount = 0;
    }
    uint256 stakeAvailable = userStake[addr] - stakeOnHold.amount;
    if (amount > stakeAvailable) {
      revert NotEnoughBalance();
    }

    userStake[addr] -= amount;
    token.transfer(sendToAddr, amount);
  }

  // funds related functions

  // Returns claimable balance by provider address.
  function getProviderClaimBalance(address providerAddr) public view returns (uint256) {
    // uint256 balance = 0;
    // the only loop that is not avoidable
    // for (uint i = 0; i < providerOnHold[providerAddr].length; i++) {
    //   if (providerOnHold[providerAddr][i].releaseAt < block.timestamp) {
    //     balance += providerOnHold[providerAddr][i].amount;
    //   }
    // }
    return 0;
  }

  // transfers provider claimable balance to provider address.
  // set amount to 0 to claim all balance.
  function claimProviderBalance(uint256 amount, address to) public {
    // uint256 balance = 0;
    // address sender = _msgSender();
    // // the only loop that is not avoidable
    // uint i = 0;

    // OnHold[] storage onHoldEntries = providerOnHold[sender];
    // while (i < onHoldEntries.length) {
    //   if (onHoldEntries[i].releaseAt < block.timestamp) {
    //     balance += onHoldEntries[i].amount;

    //     if (balance >= amount) {
    //       uint256 delta = balance - amount;
    //       onHoldEntries[i].amount = delta;
    //       token.transfer(to, amount);
    //       return;
    //     } 

    //     onHoldEntries[i] = onHoldEntries[onHoldEntries.length-1];
    //     onHoldEntries.pop();
    //   } else {
    //     i++;
    //   }
    // }
    // if (amount == 0) {
    //   amount = balance;
    //   token.transfer(to, amount);
    //   return;
    // }

    revert NotEnoughBalance();
  }

  function getTodaysSpend(address userAddress) public view returns (uint256) {
    OnHold memory spend = todaysSpend[userAddress];
    if (block.timestamp > spend.releaseAt) {
      return 0;
    }
    return spend.amount;
  }

  function setTodaysSpend(address userAddress, uint256 amount) public {
    todaysSpend[userAddress] = OnHold({
      amount: amount,
      releaseAt: (block.timestamp / DAY + 1) * DAY
    });
  }

  // Returns stake on hold by user address for the current day.
  function getStakeOnHold(address userAddress) public view returns (uint256) {
    OnHold memory onHold = userStakeOnHold[userAddress];
    if (block.timestamp > onHold.releaseAt) {
      return 0;
    }
    return onHold.amount;
  }

  // Puts user stake on hold for a day
  function setStakeOnHold(address userAddress, uint256 amount) public {
    userStakeOnHold[userAddress] = OnHold({
      amount: amount,
      releaseAt: (block.timestamp / DAY + 1) * DAY
    });
  }

  // return virtual MOR balance of user based on their stake
  function getUserVirtualMORBalance(address userAddress) public view returns (uint256) {
    return getTodaysBudget() * userStake[userAddress] / getMintedMOR() - getTodaysSpend(userAddress);
  }

  // returns stake to put on hold to get 1 virtual MOR
  function getStakeCorrespondingToVirtualMOR(uint256 virtualMOR) public view returns (uint256) {
    return virtualMOR * getMintedMOR() / getTodaysBudget();
  }

  function getTodaysBudget() public view returns (uint256) {
    // 1% of Compute Balance
    return getComputeBalance() / 100;
  }

  function getComputeBalance() public view returns (uint256) {
    // TODO: or call layer 1 contract to get daily compute balance contract somehow
    // return Distribution(distributionContractAddr)
    //   .getPeriodReward(distributionPoolId, distributionRewardStartTime, block.timestamp)
    return token.allowance(address(token), address(this));
  }

  function transferVirtualMOR(address to, uint256 amount) public {
    address computeBalanceAddr = address(0); // the account that holds daily balance
    token.transferFrom(computeBalanceAddr, to, amount);
  }

  function getMintedMOR() public view returns (uint256) {
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

  function min(uint256 a, uint256 b) internal pure returns (uint256) {
    return a < b ? a : b;
  }

  function max(uint256 a, uint256 b) internal pure returns (uint256) {
    return a > b ? a : b;
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