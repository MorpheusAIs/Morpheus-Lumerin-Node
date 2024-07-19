import { IERC20 } from "@openzeppelin/contracts/interfaces/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

// Refer to https://www.rareskills.io/post/staking-algorithm
contract StakingMasterChef {
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

  error AlreadyStaked();
  error LockPeriodNotOver();
  error InsufficientBalance();
  error Unauthorized();

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
    if (block.timestamp <= lastRewardTime) {
      return;
    }

    if (totalShares == 0) {
      lastRewardTime = block.timestamp;
      return;
    }
    accRewardPerShareScaled = getRewardPerShareScaled();
    lastRewardTime = block.timestamp;
  }

  function getRewardPerShareScaled() private view returns (uint256) {
    uint256 rewardScaled = (block.timestamp - lastRewardTime) * rewardPerSecond * PRECISION;
    return accRewardPerShareScaled + (rewardScaled / totalShares);
  }

  // Deposit staking token
  function deposit(uint256 _amount, uint8 _duration) external {
    UserInfo storage user = userInfo[msg.sender];
    updatePool();

    if (user.amount > 0) {
      revert AlreadyStaked();
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

  /// @notice Withdraw staking token
  function withdraw(uint256 _amount) external {
    UserInfo storage user = userInfo[msg.sender];
    if (block.timestamp < user.lockEndsAt) {
      revert LockPeriodNotOver();
    }
    if (user.amount < _amount) {
      revert InsufficientBalance();
    }
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

  /// @notice View function to see up-to-date reward on frontend.
  function getReward(address _user) external view returns (uint256) {
    if (block.timestamp <= lastRewardTime) {
      return 0;
    }

    if (totalShares == 0) {
      return 0;
    }
    UserInfo memory user = userInfo[_user];
    return (user.shareAmount * getRewardPerShareScaled()) / PRECISION - user.rewardDebt;
  }

  function withdrawReward() external {
    updatePool();

    UserInfo storage user = userInfo[msg.sender];
    uint256 rewardFromStart = (user.shareAmount * accRewardPerShareScaled) / PRECISION;
    uint256 reward = rewardFromStart - user.rewardDebt;
    user.rewardDebt = rewardFromStart;
    safeTransfer(msg.sender, reward);
  }

  // Safe reward transfer function, just in case if rounding error causes pool to not have enough reward token.
  function safeTransfer(address _to, uint256 _amount) internal {
    uint256 morRewardBalance = rewardToken.allowance(fundingAccount, address(this));
    rewardToken.safeTransferFrom(fundingAccount, _to, min(morRewardBalance, _amount));
  }

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
    if (msg.sender != owner) {
      revert Unauthorized();
    }
    _;
  }
}
