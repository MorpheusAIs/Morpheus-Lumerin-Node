import { IERC20 } from "@openzeppelin/contracts/interfaces/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { SafeMath } from "./SafeMath.sol";
import "hardhat/console.sol";

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract StakingMasterChef {
  using SafeMath for uint256;
  using SafeERC20 for IERC20;

  IERC20 public immutable morToken;
  uint256 constant PRECISION = 1e12;
  // IERC20 public immutable rewardsToken;

  // Info of each user.
  struct UserInfo {
    uint256 amount; // How many LP tokens the user has provided.
    uint256 rewardDebt; // Reward debt. See explanation below.
    //
    // We do some fancy math here. Basically, any point in time, the amount of SUSHIs
    // entitled to a user but is pending to be distributed is:
    //
    //   pending reward = (user.amount * pool.accSushiPerShare) - user.rewardDebt
    //
    // Whenever a user deposits or withdraws LP tokens to a pool. Here's what happens:
    //   1. The pool's `accSushiPerShare` (and `lastRewardBlock`) gets updated.
    //   2. User receives the pending reward sent to his/her address.
    //   3. User's `amount` gets updated.
    //   4. User's `rewardDebt` gets updated.
  }

  // Info of each pool.
  struct PoolInfo {
    IERC20 lmrToken; // Address of LP token contract.
    uint256 rewardPerSecond; // How many allocation points assigned to this pool. SUSHIs to distribute per block.
    uint256 lastRewardTime; // Last time rewards were distributed
    uint256 accRewardPerShareScaled; // Accumulated SUSHIs per share, times 1e12. See below.
  }

  struct LockPeriodBonus {
    uint256 lockPeriodSeconds;
    uint256 multiplierScaled; // scaled by PRECISION
  }

  // Info of each pool.
  PoolInfo[] public poolInfo;

  // Info of each user that stakes LP tokens.
  mapping(uint256 => mapping(address => UserInfo)) public userInfo;

  // The block number when SUSHI mining starts.
  uint256 public startTime;
  address public owner;
  address fundingAccount; // account which stores the MOR tokens with infinite allowance for this contract

  event Deposit(address indexed user, uint256 indexed pid, uint256 amount);
  event Withdraw(address indexed user, uint256 indexed pid, uint256 amount);

  constructor(IERC20 _morToken, address _fundingAccount) {
    owner = msg.sender;
    morToken = _morToken;
    fundingAccount = _fundingAccount;
  }

  // Add a new lp to the pool. Can only be called by the owner.
  // XXX DO NOT add the same LP token more than once. Rewards will be messed up if you do.
  function add(uint256 rewardPerSecond, IERC20 _lpToken, bool _withUpdate) public onlyOwner {
    if (_withUpdate) {
      massUpdatePools();
    }
    // uint256 lastRewardTime = block.timestamp > startTime ? block.timestamp : startTime;
    uint256 lastRewardTime = block.timestamp;
    // totalAllocPoint = totalAllocPoint.add(rewardPerSecond);
    poolInfo.push(
      PoolInfo({
        lmrToken: _lpToken,
        rewardPerSecond: rewardPerSecond,
        lastRewardTime: lastRewardTime,
        accRewardPerShareScaled: 0
      })
    );
  }

  // Update reward variables of the given pool to be up-to-date.
  function updatePool(uint256 _pid) public {
    // TODO:
    // Place to withdraw MOR from distribution contract
    PoolInfo storage pool = poolInfo[_pid];
    if (block.timestamp <= pool.lastRewardTime) {
      return;
    }
    uint256 lmrSupply = pool.lmrToken.balanceOf(address(this));
    if (lmrSupply == 0) {
      pool.lastRewardTime = block.timestamp;
      return;
    }
    uint256 morReward = (block.timestamp - pool.lastRewardTime) * pool.rewardPerSecond;
    console.log("morReward: %s", morReward);

    pool.accRewardPerShareScaled = pool.accRewardPerShareScaled + ((morReward * PRECISION) / lmrSupply);
    console.log("pool.accRewardPerShareScaled: %s", pool.accRewardPerShareScaled);
    pool.lastRewardTime = block.timestamp;
  }

  // Update reward vairables for all pools. Be careful of gas spending!
  function massUpdatePools() public {
    uint256 length = poolInfo.length;
    for (uint256 pid = 0; pid < length; ++pid) {
      updatePool(pid);
    }
  }

  // Deposit LP tokens to MasterChef for SUSHI allocation.
  function deposit(uint256 _pid, uint256 _amount) public {
    PoolInfo storage pool = poolInfo[_pid];
    UserInfo storage user = userInfo[_pid][msg.sender];
    updatePool(_pid);
    if (user.amount > 0) {
      uint256 pending = (user.amount * pool.accRewardPerShareScaled) / PRECISION - user.rewardDebt;
      safeTransfer(msg.sender, pending);
    }
    pool.lmrToken.safeTransferFrom(address(msg.sender), address(this), _amount);
    user.amount += _amount;
    user.rewardDebt = (user.amount * pool.accRewardPerShareScaled) / PRECISION;
    emit Deposit(msg.sender, _pid, _amount);
  }

  // Withdraw LP tokens from MasterChef.
  function withdraw(uint256 _pid, uint256 _amount) public {
    PoolInfo storage pool = poolInfo[_pid];
    UserInfo storage user = userInfo[_pid][msg.sender];
    require(user.amount >= _amount, "withdraw: not good");
    updatePool(_pid);
    console.log("user.amount: %s", user.amount);
    console.log("pool.accRewardPerShareScaled: %s", pool.accRewardPerShareScaled);
    console.log("user.rewardDebt: %s", user.rewardDebt);
    uint256 pending = (user.amount * pool.accRewardPerShareScaled) / PRECISION - user.rewardDebt;
    safeTransfer(msg.sender, pending);
    user.amount = user.amount.sub(_amount);
    user.rewardDebt = (user.amount * pool.accRewardPerShareScaled) / PRECISION;
    pool.lmrToken.safeTransfer(address(msg.sender), _amount);
    emit Withdraw(msg.sender, _pid, _amount);
  }

  // Safe sushi transfer function, just in case if rounding error causes pool to not have enough SUSHIs.
  function safeTransfer(address _to, uint256 _amount) internal {
    uint256 morRewardBalance = morToken.allowance(fundingAccount, address(this));
    console.log("transferred %s MOR tokens", min(morRewardBalance, _amount));
    morToken.transferFrom(fundingAccount, _to, min(morRewardBalance, _amount));
  }

  // Return reward multiplier over the given _from to _to block.
  // function getMultiplier(uint256 _from, uint256 _to) public view returns (uint256) {
  //   return 1;
  // if (_to <= bonusEndBlock) {
  //   return _to.sub(_from).mul(BONUS_MULTIPLIER);
  // } else if (_from >= bonusEndBlock) {
  //   return _to.sub(_from);
  // } else {
  //   return bonusEndBlock.sub(_from).mul(BONUS_MULTIPLIER).add(_to.sub(bonusEndBlock));
  // }
  // }

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
