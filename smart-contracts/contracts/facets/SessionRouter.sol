// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { MessageHashUtils } from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "../libraries/KeySet.sol";
import { ModelRegistry } from "./ModelRegistry.sol";
import { ProviderRegistry } from "./ProviderRegistry.sol";
import { AppStorage, Session, Bid, OnHold, Pool } from "../AppStorage.sol";
import { LibOwner } from "../libraries/LibOwner.sol";
import { LinearDistributionIntervalDecrease } from "../libraries/LinearDistributionIntervalDecrease.sol";

contract SessionRouter {
  using KeySet for KeySet.Set;
  AppStorage internal s;

  // constants
  uint32 public constant DAY = 24 * 60 * 60; // 1 day
  uint32 public constant MIN_SESSION_DURATION = 5 * 60; // 5 minutes

  // events
  event SessionOpened(
    address indexed userAddress,
    bytes32 indexed sessionId,
    address indexed providerId
  );
  event SessionClosed(
    address indexed userAddress,
    bytes32 indexed sessionId,
    address indexed providerId
  );
  event Staked(address indexed userAddress, uint256 amount);
  event Unstaked(address indexed userAddress, uint256 amount);
  event ProviderClaimed(address indexed providerAddress, uint256 amount);

  // errors
  error NotUserOrProvider();
  error NotEnoughWithdrawableBalance();   // means that there is not enough funds at all or some funds are still locked

  error SessionTooShort();
  error SessionNotFound();
  error SessionAlreadyClosed();

  error BidNotFound();
  error BidTaken();

  //===========================
  //         SESSION
  //===========================

  function getSession(bytes32 sessionId) public view returns (Session memory) {
    return s.sessions[s.sessionMap[sessionId]];
  }

  function openSession(
    bytes32 bidId,
    uint256 _stake
  ) public returns (bytes32 sessionId) {
    address sender = msg.sender;

    Bid memory bid = s.bidMap[bidId];
    if (bid.deletedAt != 0 || bid.createdAt == 0) {
      revert BidNotFound();
    }

    if (s.bidSessionMap[bidId] != 0) {
      revert BidTaken();
    }

    uint256 duration = stakeToStipend(_stake, uint128(block.timestamp)) / bid.pricePerSecond;
    if (duration < MIN_SESSION_DURATION) {
      revert SessionTooShort();
    }

    sessionId = keccak256(
      abi.encodePacked(sender, bid.provider, _stake, block.number)
    );
    s.sessions.push(
      Session({
        id: sessionId,
        user: sender,
        provider: bid.provider,
        modelAgentId: bid.modelAgentId,
        bidID: bidId,
        stake: _stake,
        pricePerSecond: bid.pricePerSecond,
        closeoutReceipt: "",
        closeoutType: 0,
        providerWithdrawnAmount: 0,
        openedAt: block.timestamp,
        closedAt: 0
      })
    );

    uint256 sessionIndex = s.sessions.length - 1;
    s.sessionMap[sessionId] = sessionIndex;
    s.bidSessionMap[bidId] = sessionIndex; // marks bid as "taken" by this session

    emit SessionOpened(sender, sessionId, bid.provider);
    s.token.transferFrom(sender, address(this), _stake); // errors with Insufficient Allowance if not approved

    return sessionId;
  }

  // returns expected session duration in seconds
  // should be called daily 00:00:00 UTC or at the beginning of the session
  // returns type(uint256).max if session will not close till the end of the day
  function getSessionEndTime(bytes32 sessionId) public view returns (uint256) {
    Session memory session = s.sessions[s.sessionMap[sessionId]];
    if (session.closedAt != 0) {
      revert SessionAlreadyClosed();
    }

    uint256 stipend = stakeToStipend(session.stake, uint128(block.timestamp));
    uint256 durationSeconds = stipend / session.pricePerSecond;

    uint256 secondsFromStartOfDay = block.timestamp % DAY;
    uint256 startOfToday = block.timestamp - secondsFromStartOfDay;
    uint256 secondsLeftToday = DAY - secondsFromStartOfDay;
    uint256 startOfTheTomorrow = startOfToday + DAY;

    // case 1 
    // started today and will end today 
    if (session.openedAt > startOfToday && session.openedAt + durationSeconds < startOfTheTomorrow) {
      return session.openedAt + durationSeconds;
    }

    // case 2 
    // started today and will end tomorrow (at midnight new stipend issued)
    if (session.openedAt > startOfToday && session.openedAt + durationSeconds >= startOfTheTomorrow) {
      uint256 tomorrowStipend = stakeToStipend(session.stake, uint128(block.timestamp + secondsLeftToday + 1));
      uint256 tomorrowDurationSeconds = tomorrowStipend / session.pricePerSecond;
      return session.openedAt + durationSeconds + tomorrowDurationSeconds;
    }

    // case 3
    // started any time and won't end today
    if (durationSeconds >= DAY) {
      return type(uint256).max;
    }

    // case 4
    // started any time and will end today
    if (durationSeconds < DAY) {
      return startOfToday + durationSeconds;
    }

    return stipend / session.pricePerSecond;
  }

  function closeSession(
    bytes32 sessionId,
    bytes memory receiptEncoded,
    bytes memory signature
  ) public {
    Session storage session = s.sessions[s.sessionMap[sessionId]];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    if (session.user != msg.sender && session.provider != msg.sender) {
      revert NotUserOrProvider();
    }
    if (session.closedAt != 0) {
      revert SessionAlreadyClosed();
    }

    s.bidSessionMap[session.bidID] = 0; // marks bid as available
    session.closeoutReceipt = receiptEncoded;
    session.closedAt = block.timestamp;

    uint256 startOfToday = startOfTheDay(block.timestamp);
    uint256 todaysSessionDurationSeconds = block.timestamp - maxUint256(session.openedAt, startOfToday);
    uint256 todaySessionCost = todaysSessionDurationSeconds * session.pricePerSecond;

    // calculate provider withdraw
    uint256 providerWithdraw;
    if (isValidReceipt(session.provider, receiptEncoded, signature)) {
      // withdraw all remaining provider funds
      uint256 totalSessionDuration = session.closedAt - session.openedAt;
      uint256 totalCost = totalSessionDuration * session.pricePerSecond;
      providerWithdraw = totalCost - session.providerWithdrawnAmount;
    } else {
      // withdraw all funds except for today's session cost
      session.closeoutType = 1;
      uint256 durationTillToday = startOfToday - minUint256(session.openedAt, startOfToday);
      uint256 costTillToday = durationTillToday * session.pricePerSecond;
      providerWithdraw = costTillToday - session.providerWithdrawnAmount;
    }
    session.providerWithdrawnAmount+=providerWithdraw;

    // calculate user withdraw
    uint256 userStakeToLock = stipendToStake(todaySessionCost, uint128(block.timestamp));
    s.userOnHold[session.user].push(OnHold({
      amount: session.stake - userStakeToLock,
      releaseAt: block.timestamp + DAY
    }));
    uint256 userWithdraw = session.stake - userStakeToLock;

    emit SessionClosed(session.user, sessionId, session.provider);
    
    // withdraw provider and user  funds
    s.token.transferFrom(s.fundingAccount, session.provider, providerWithdraw);
    s.token.transfer(session.user, userWithdraw);
  }

  // funds related functions

  // returns total claimanble balance for the provider for particular session
  function getProviderClaimableBalance(bytes32 sessionId) public view returns (uint256) {
    Session memory session = s.sessions[s.sessionMap[sessionId]];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    
    return _getProviderClaimableBalance(session);
  }

  function claimProviderBalance(bytes32 sessionId, uint256 amountToWithdraw, address to) public {
    Session storage session = s.sessions[s.sessionMap[sessionId]];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    LibOwner._senderOrOwner(session.provider);

    uint256 withdrawableAmount = _getProviderClaimableBalance(session);

    if (amountToWithdraw > withdrawableAmount) {
      revert NotEnoughWithdrawableBalance();
    }

    session.providerWithdrawnAmount += amountToWithdraw;
    s.token.transferFrom(s.fundingAccount, to, amountToWithdraw);
    return;
  }

  function _getProviderClaimableBalance(Session memory session) internal view returns (uint256){
     if (session.closedAt == 0) {
      // session is still open
      // provider can claim all funds except for today's session cost
      uint256 startOfToday = startOfTheDay(block.timestamp);
      uint256 durationTillToday = startOfToday - minUint256(session.openedAt, startOfToday);
      uint256 costTillToday = durationTillToday * session.pricePerSecond;
      uint256 withdrawableAmount = costTillToday - session.providerWithdrawnAmount;
      return withdrawableAmount;
    } else if (session.closedAt > startOfTheDay(block.timestamp)){
      // likely not to be the case when session was closed today
      // cause provider already got funds
      uint256 durationTillToday = startOfTheDay(block.timestamp) - 
        minUint256(session.openedAt, startOfTheDay(block.timestamp));
      uint256 costTillToday = durationTillToday * session.pricePerSecond;
      uint256 withdrawableAmount = costTillToday - session.providerWithdrawnAmount;
      return withdrawableAmount;
    } else {
      // session was closed yesterday or earlier
      // provider can claim all funds
      uint256 totalSessionDuration = session.closedAt - session.openedAt;
      uint256 totalCost = totalSessionDuration * session.pricePerSecond;
      uint256 withdrawableAmount = totalCost - session.providerWithdrawnAmount;
      return withdrawableAmount;
    }
  }

  function deleteHistory(bytes32 sessionId) public {
    Session storage session = s.sessions[s.sessionMap[sessionId]];
    LibOwner._senderOrOwner(session.user);
    session.user = address(0);
  }

  function setStakeDelay(int256 delay) public {
    LibOwner._onlyOwner();
    s.stakeDelay = delay;
  }

  function isValidReceipt(
    address signer,
    bytes memory receipt,
    bytes memory signature
  ) public pure returns (bool) {
    if (signature.length == 0) {
      return false;
    }
    bytes32 receiptHash = MessageHashUtils.toEthSignedMessageHash(
      keccak256(receipt)
    );
    return ECDSA.recover(receiptHash, signature) == signer;
  }

  //===========================
  //         STAKING
  //===========================

  // returns amount of user stake withdrawable and on hold
  function withdrawableUserStake(
    address userAddr
  ) public view returns (uint256 avail, uint256 hold) {
    OnHold[] memory onHold = s.userOnHold[userAddr];
    for (uint i = 0; i < onHold.length; i++) {
      uint256 amount = onHold[i].amount;
      if (block.timestamp < onHold[i].releaseAt) {
        hold += amount;
      } else {
        avail += amount;
      }
    }
    return (avail, hold);
  }

  function withdrawUserStake(uint256 amountToWithdraw, address to) public {
    uint256 balance = 0;
    address sender = msg.sender;

    // withdraw all available funds if amountToWithdraw is 0
    if (amountToWithdraw == 0) {
      amountToWithdraw = type(uint256).max;
    }

    OnHold[] storage onHoldEntries = s.userOnHold[sender];
    uint i = 0;

    // the only loop that is not avoidable     
    while (i < onHoldEntries.length) {
      if (block.timestamp > onHoldEntries[i].releaseAt) {
        balance += onHoldEntries[i].amount;

        if (balance >= amountToWithdraw) {
          uint256 delta = balance - amountToWithdraw;
          onHoldEntries[i].amount = delta;
          s.token.transfer(to, amountToWithdraw);
          return;
        }

        // removes entry from array
        onHoldEntries[i] = onHoldEntries[onHoldEntries.length - 1];
        onHoldEntries.pop();
      } else {
        i++;
      }
    }

    if (amountToWithdraw == type(uint256).max) {
      s.token.transfer(to, balance);
      return;
    }

    revert NotEnoughWithdrawableBalance();
  }

  // returns stipend of user based on their stake
  function stakeToStipend(
    uint256 sessionStake, uint128 timestamp
  ) internal view returns (uint256) {
    return (getTodaysBudget(timestamp) * sessionStake) / s.token.totalSupply();
  }

  // returns stake of user based on their stipend
  function stipendToStake(
    uint256 stipend, uint128 timestamp
  ) internal view returns (uint256) {
    return (stipend * s.token.totalSupply()) / getTodaysBudget(timestamp);
  }

  function getTodaysBudget(uint128 timestamp) public view returns (uint256) {
    return getComputeBalance(timestamp) / 100; // 1% of Compute Balance
  }

  function getComputeBalance(uint128 timestamp) public view returns (uint256) {
    // TODO: cache today's budget and compute balance
    return LinearDistributionIntervalDecrease.getPeriodReward(
        s.pool.initialReward,     
        s.pool.rewardDecrease,
        s.pool.payoutStart,
        s.pool.decreaseInterval,
        s.pool.payoutStart,       // should that be payoutStart or 1707350400 ephochSeconds (Feb 8 2024 00:00:00) 
        uint128(timestamp)
      );
  }

  // parameters should be the same as in Ethereum L1 Distribution contract
  // at address 0x47176B2Af9885dC6C4575d4eFd63895f7Aaa4790
  // call 'Distribution.pools(3)' where '3' is a poolId
  function setPoolConfig(Pool calldata pool) public {
    LibOwner._onlyOwner();
    s.pool = pool;
  }

  function startOfTheDay(uint256 timestamp) public pure returns (uint256) {
    return timestamp - (timestamp % DAY);
  }

  function minUint256(uint256 a, uint256 b) internal pure returns (uint256) {
    return a < b ? a : b;
  }

  function maxUint256(uint256 a, uint256 b) internal pure returns (uint256) {
    return a > b ? a : b;
  }
}
