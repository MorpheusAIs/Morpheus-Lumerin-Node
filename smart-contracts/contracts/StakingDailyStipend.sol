// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "./KeySet.sol";
import { Marketplace } from "./Marketplace.sol";
import "hardhat/console.sol";

contract StakingDailyStipend is OwnableUpgradeable {
  struct OnHold {
    uint256 amount;
    uint256 releaseAt; // in epoch seconds TODO: consider using hours to reduce storage cost
  }

  // dependencies
  IERC20 public token;
  address public sessionRouter;
  address public tokenAccount; // account which stores the MOR tokens with infinite allowance for this contract

  // storage
  mapping(address => uint256) public userStake; // user address => stake balance
  mapping(address => OnHold) public todaysSpend; // user address => spend today

  // constants
  uint32 constant DAY = 24*60*60; // 1 day

  // events
  event Staked(address indexed userAddress, uint256 amount);
  event Unstaked(address indexed userAddress, uint256 amount);

  // errors
  error NotSenderOrOwner();
  error NotSessionRouterOrOwner();
  error NotEnoughStake();
  error NotEnoughDailyStipend();

  function initialize(address _token, address _tokenAccount) public initializer {
    __Ownable_init();
    token = IERC20(_token);
    tokenAccount = _tokenAccount;
  }

  function stake(address addr, uint256 amount) public senderOrOwner(addr){
    userStake[addr] += amount;
    token.transferFrom(addr, address(this), amount);
  }

  function unstake(address addr, uint256 amount, address sendToAddr) public senderOrOwner(addr){
    if (amount > withdrawableStakeBalance(addr)) {
      revert NotEnoughStake();
    }

    userStake[addr] -= amount;
    token.transfer(sendToAddr, amount);
  }

  function withdrawableStakeBalance(address userAddress) public view returns (uint256) {
    return userStake[userAddress] - getStakeOnHold(userAddress);
  }
  
  // return virtual MOR balance of user based on their stake
  function balanceOfDailyStipend(address userAddress) public view returns (uint256) {
    return getTodaysBudget() * userStake[userAddress] / token.totalSupply() - getTodaysSpend(userAddress);
  }

  function transferDailyStipend(address from, address to, uint256 amount) public /*onlyOwnerOrSessionRouter*/{
    if (amount > balanceOfDailyStipend(from)) {
      revert NotEnoughDailyStipend();
    }
    todaysSpend[from] = OnHold({
      amount: amount,
      releaseAt: (block.timestamp / DAY + 1) * DAY
    });
    token.transferFrom(address(tokenAccount), to, amount);
  }

  function returnStipend(address to, uint256 amount) public {
    token.transferFrom(_msgSender(), address(tokenAccount), amount);
    uint256 oldSpend = getTodaysSpend(to);
    todaysSpend[to] = OnHold({
      amount: oldSpend > amount ? oldSpend - amount : 0,
      releaseAt: (block.timestamp / DAY + 1) * DAY
    });
  }

  function getTodaysSpend(address userAddress) public view returns (uint256) {
    OnHold memory spend = todaysSpend[userAddress];
    if (block.timestamp > spend.releaseAt) {
      return 0;
    }
    return spend.amount;
  }

  function getStakeOnHold(address userAddress) public view returns (uint256) {
    return getTodaysSpend(userAddress) * token.totalSupply() / getTodaysBudget();
  }

  function getTodaysBudget() public view returns (uint256) {
    // 1% of Compute Balance
    return getComputeBalance() / 100;
  }

  function getComputeBalance() public view returns (uint256) {
    // TODO: or call layer 1 contract to get daily compute balance contract
    //
    // arguments for getPeriodReward call
    // address public constant distributionContractAddr = address(0x0);
    // uint32 public constant distributionRewardStartTime = 1707350400; // ephochSeconds Feb 8 2024 00:00:00
    // uint8 public constant distributionPoolId = 3;
    //
    // return Distribution(distributionContractAddr)
    //   .getPeriodReward(distributionPoolId, distributionRewardStartTime, block.timestamp)
    // return token.allowance(address(token), address(this));
    return 10 * 10**18; // 10 tokens
  }

  modifier onlyOwnerOrSessionRouter() {
    if (_msgSender() != sessionRouter && _msgSender() != owner()) {
      revert NotSessionRouterOrOwner();
    }
    _;
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