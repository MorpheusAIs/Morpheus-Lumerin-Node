// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

contract Staking is Ownable {
  IERC20 public immutable lmrToken;
  IERC20 public immutable morToken;
  mapping(address => Stake) public stakes;
  address[] public stakers;

  struct Stake {
    uint256 index;
    uint256 amount;
    uint128 lastClaimedDay; // days from the epoch
  }

  event Staked(address indexed user, uint256 amount);
  event Withdrawn(address indexed user, uint256 amount);

  error NoStakeToWithdraw();

  constructor(IERC20 _lmrToken, IERC20 _morToken) Ownable(msg.sender) {
    lmrToken = _lmrToken;
    morToken = _morToken;
    stakers.push(address(0));
  }

  function stake(uint256 _amount) external {
    Stake storage userStake = stakes[msg.sender];

    userStake.amount += _amount;
    userStake.lastClaimedDay = todaysDay();

    if (userStake.amount > 0 && userStake.index == 0) {
      stakers.push(msg.sender);
      userStake.index = stakers.length;
    }
    lmrToken.transferFrom(msg.sender, address(this), _amount);
  }

  function withdraw() external {
    Stake storage userStake = stakes[msg.sender];
    if (userStake.amount == 0) {
      revert NoStakeToWithdraw();
    }
    userStake.amount = 0;

    // update stakers array
    stakers[userStake.index] = stakers[stakers.length];
    stakers.pop();

    lmrToken.transfer(msg.sender, userStake.amount);
  }

  function claimDailyReward(address addr) public {
    Stake storage userStake = stakes[addr];
    uint128 daysSinceLastClaim = todaysDay() - userStake.lastClaimedDay;
    if (daysSinceLastClaim == 0) {
      return;
    }
    uint256 reward = calculateDailyReward(userStake.amount) * daysSinceLastClaim;
    userStake.lastClaimedDay = todaysDay();
    morToken.transfer(addr, reward);
  }

  /// @notice Claim daily rewards for multiple addresses, intended to be invoked by extrernal script
  /// @param addrs Array of addresses to claim daily rewards for
  function batchClaimDailyReward(address[] calldata addrs) external {
    for (uint256 i = 0; i < addrs.length; i++) {
      claimDailyReward(addrs[i]);
    }
  }

  function todaysDay() private view returns (uint128) {
    return uint128(block.timestamp / 1 days);
  }

  function calculateDailyReward(uint256 _stake) private pure returns (uint256) {
    //TODO: implement this function
    return _stake / 100;
  }
}
