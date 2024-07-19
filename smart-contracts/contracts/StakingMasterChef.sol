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

  IERC20 public immutable morToken;
  IERC20 public immutable lmrToken;

  uint256 rewardPerSecond; // How many allocation points assigned to this pool. SUSHIs to distribute per block.
  uint256 lastRewardTime; // Last time rewards were distributed
  uint256 accRewardPerShareScaled; // Accumulated SUSHIs per share, times 1e12. See below.
  uint256 public startTime;
  address public owner;
  address fundingAccount; // account which stores the MOR tokens with infinite allowance for this contract

  // Info of each user.
  struct UserInfo {
    uint256 amount; // How many LP tokens the user has provided.
    uint256 rewardDebt; // Reward debt. See explanation below.
  }
  mapping(address => UserInfo) public userInfo; // userAddress => UserInfo

  event Deposit(address indexed user, uint256 amount);
  event Withdraw(address indexed user, uint256 amount);

  constructor(IERC20 _morToken, IERC20 _lmrToken, address _fundingAccount, uint256 _rewardPerSecond) {
    owner = msg.sender;
    morToken = _morToken;
    lmrToken = _lmrToken;
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
    uint256 lmrSupply = lmrToken.balanceOf(address(this));
    if (lmrSupply == 0) {
      lastRewardTime = block.timestamp;
      return;
    }
    uint256 morReward = (block.timestamp - lastRewardTime) * rewardPerSecond;
    console.log("morReward: %s", morReward);

    accRewardPerShareScaled = accRewardPerShareScaled + ((morReward * PRECISION) / lmrSupply);
    console.log("accaccRewardPerShareScaled: %s", accRewardPerShareScaled);
    lastRewardTime = block.timestamp;
  }

  // Deposit LP tokens to MasterChef for SUSHI allocation.
  function deposit(uint256 _amount) public {
    UserInfo storage user = userInfo[msg.sender];
    updatePool();
    if (user.amount > 0) {
      uint256 pending = (user.amount * accRewardPerShareScaled) / PRECISION - user.rewardDebt;
      safeTransfer(msg.sender, pending);
    }
    lmrToken.safeTransferFrom(address(msg.sender), address(this), _amount);
    user.amount += _amount;
    user.rewardDebt = (user.amount * accRewardPerShareScaled) / PRECISION;
    emit Deposit(msg.sender, _amount);
  }

  // Withdraw LP tokens from MasterChef.
  function withdraw(uint256 _amount) public {
    UserInfo storage user = userInfo[msg.sender];
    require(user.amount >= _amount, "withdraw: not good");
    updatePool();
    console.log("user.amount: %s", user.amount);
    console.log("accaccRewardPerShareScaled: %s", accRewardPerShareScaled);
    console.log("user.rewardDebt: %s", user.rewardDebt);
    uint256 pending = (user.amount * accRewardPerShareScaled) / PRECISION - user.rewardDebt;
    safeTransfer(msg.sender, pending);
    user.amount = user.amount.sub(_amount);
    user.rewardDebt = (user.amount * accRewardPerShareScaled) / PRECISION;
    lmrToken.safeTransfer(address(msg.sender), _amount);
    emit Withdraw(msg.sender, _amount);
  }

  // Safe sushi transfer function, just in case if rounding error causes pool to not have enough SUSHIs.
  function safeTransfer(address _to, uint256 _amount) internal {
    uint256 morRewardBalance = morToken.allowance(fundingAccount, address(this));
    console.log("transferred %s MOR tokens", min(morRewardBalance, _amount));
    morToken.transferFrom(fundingAccount, _to, min(morRewardBalance, _amount));
  }

  function getLockPeriod(uint8 _period) public pure returns (uint256, uint256) {
    if (_period == 0) {
      return (7 days, (100 * PRECISION) / 100);
    } else if (_period == 1) {
      return (30 days, (115 * PRECISION) / 100);
    } else if (_period == 2) {
      return (180 days, (135 * PRECISION) / 100);
    } else if (_period == 3) {
      return (365 days, (150 * PRECISION) / 100);
    } else {
      revert("Invalid lock period");
    }
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
