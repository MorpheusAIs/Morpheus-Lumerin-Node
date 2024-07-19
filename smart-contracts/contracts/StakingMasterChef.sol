import { IERC20 } from "@openzeppelin/contracts/interfaces/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { SafeMath } from "./SafeMath.sol";
import "hardhat/console.sol";

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract StakingMasterChef {
  using SafeMath for uint256;
  using SafeERC20 for IERC20;

  uint256 constant PRECISION = 1e12;

  IERC20 public immutable rewardToken;
  IERC20 public immutable stakingToken;

  uint256 rewardPerSecond; // Reward tokens per second
  uint256 lastRewardTime; // Last time rewards were distributed
  uint256 accRewardPerShareScaled; // Accumulated reward per share, times 1e12
  uint256 totalShares; // Total shares of reward token
  uint256 public startTime;
  address public owner;
  address fundingAccount; // account which stores the reward tokens with allowance for this contract

  // Info of each user.
  struct UserInfo {
    uint256 amount; // How many LP tokens the user has provided.
    uint256 shareAmount; // How many shares the user has received
    uint256 rewardDebt; // Reward debt. See explanation below.
    uint256 lockEndsAt; // When the lock period ends
  }
  mapping(address => UserInfo) public userInfo; // userAddress => UserInfo

  struct LockPeriod {
    uint256 lockPeriodSeconds;
    uint256 multiplierScaled;
  }
  LockPeriod[4] lockPeriod = [
    LockPeriod(7 days, (100 * PRECISION) / 100),
    LockPeriod(30 days, (115 * PRECISION) / 100),
    LockPeriod(180 days, (135 * PRECISION) / 100),
    LockPeriod(365 days, (150 * PRECISION) / 100)
  ];

  event Deposit(address indexed user, uint256 amount);
  event Withdraw(address indexed user, uint256 amount);

  constructor(IERC20 _rewardToken, IERC20 _stakingToken, address _fundingAccount, uint256 _rewardPerSecond) {
    owner = msg.sender;
    rewardToken = _rewardToken;
    stakingToken = _stakingToken;
    fundingAccount = _fundingAccount;
    lastRewardTime = block.timestamp;
    rewardPerSecond = _rewardPerSecond;
  }

  // Update reward variables of the given pool to be up-to-date.
  function updatePool() public {
    // TODO:
    // Place to withdraw MOR from distribution contract
    if (block.timestamp <= lastRewardTime) {
      return;
    }
    // uint256 totalStaked = stakingToken.balanceOf(address(this));
    // if (totalStaked == 0) {
    //   lastRewardTime = block.timestamp;
    //   return;
    // }
    if (totalShares == 0) {
      lastRewardTime = block.timestamp;
      return;
    }
    uint256 rewardScaled = (block.timestamp - lastRewardTime) * rewardPerSecond * PRECISION;
    accRewardPerShareScaled = accRewardPerShareScaled + (rewardScaled / totalShares);
    lastRewardTime = block.timestamp;
  }

  // Deposit staking token
  function deposit(uint256 _amount, uint8 _duration) public {
    UserInfo storage user = userInfo[msg.sender];
    updatePool();

    // if user already staked TODO
    if (user.amount > 0) {
      revert("User already staked");
      // uint256 rewardFromStart = (user.shareAmount * accRewardPerShareScaled) / PRECISION;
      // uint256 pending = rewardFromStart - user.rewardDebt;
      // safeTransfer(msg.sender, pending);
    }

    stakingToken.safeTransferFrom(address(msg.sender), address(this), _amount);
    (uint256 lockDuration, uint256 multiplierScaled) = getLockPeriod(_duration);

    uint256 userShares = (_amount * multiplierScaled) / PRECISION;
    totalShares += userShares;

    user.shareAmount = userShares;
    user.amount += _amount;
    user.rewardDebt = (userShares * accRewardPerShareScaled) / PRECISION;
    user.lockEndsAt = block.timestamp + lockDuration;

    emit Deposit(msg.sender, _amount);
  }

  // Withdraw LP tokens from MasterChef.
  function withdraw(uint256 _amount) public {
    UserInfo storage user = userInfo[msg.sender];
    require(block.timestamp >= user.lockEndsAt, "withdraw: lock period not over");
    require(user.amount >= _amount, "withdraw: insufficient balance");
    updatePool();

    uint256 reward = (user.shareAmount * accRewardPerShareScaled) / PRECISION - user.rewardDebt;
    safeTransfer(msg.sender, reward);

    uint256 withdrawShares = (((_amount * PRECISION) / user.amount) * user.shareAmount) / PRECISION;
    user.amount = user.amount - _amount;
    user.shareAmount = user.shareAmount - withdrawShares;

    totalShares -= withdrawShares;

    user.rewardDebt = (user.shareAmount * accRewardPerShareScaled) / PRECISION;
    stakingToken.safeTransfer(address(msg.sender), _amount);
    emit Withdraw(msg.sender, _amount);
  }

  // Safe reward transfer function, just in case if rounding error causes pool to not have enough reward token.
  function safeTransfer(address _to, uint256 _amount) internal {
    uint256 morRewardBalance = rewardToken.allowance(fundingAccount, address(this));
    rewardToken.transferFrom(fundingAccount, _to, min(morRewardBalance, _amount));
  }

  // function getLockPeriod(uint8 _period) public pure returns (uint256, uint256) {
  //   if (_period == 0) {
  //     return (7 days, (100 * PRECISION) / 100);
  //   } else if (_period == 1) {
  //     return (30 days, (115 * PRECISION) / 100);
  //   } else if (_period == 2) {
  //     return (180 days, (135 * PRECISION) / 100);
  //   } else if (_period == 3) {
  //     return (365 days, (150 * PRECISION) / 100);
  //   } else {
  //     revert("Invalid lock period");
  //   }
  // }

  function getLockPeriod(uint8 _period) public view returns (uint256, uint256) {
    return (lockPeriod[_period].lockPeriodSeconds, lockPeriod[_period].multiplierScaled);
  }

  function max(uint256 a, uint256 b) internal pure returns (uint256) {
    return a > b ? a : b;
  }

  function min(uint256 a, uint256 b) internal pure returns (uint256) {
    return a < b ? a : b;
  }

  modifier onlyOwner() {
    require(msg.sender == owner, "not authorized");
    _;
  }
}
