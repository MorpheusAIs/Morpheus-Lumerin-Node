// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { MessageHashUtils } from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import { KeySet, Uint256Set } from "../libraries/KeySet.sol";
import { AppStorage, Session, Bid, OnHold, Pool } from "../AppStorage.sol";
import { LibOwner } from "../libraries/LibOwner.sol";
import { LinearDistributionIntervalDecrease } from "../libraries/LinearDistributionIntervalDecrease.sol";

contract SessionRouter {
  using KeySet for KeySet.Set;
  using Uint256Set for Uint256Set.Set;

  AppStorage internal s;

  // constants
  uint32 public constant MIN_SESSION_DURATION = 5 minutes;
  uint32 public constant MAX_SESSION_DURATION = 7 days;
  uint32 public constant SIGNATURE_TTL = 10 minutes;

  // events
  event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);
  event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);

  // errors
  error NotUserOrProvider();
  error NotEnoughWithdrawableBalance(); // means that there is not enough funds at all or some funds are still locked
  error ProviderSignatureMismatch();
  error SignatureExpired();
  error DuplicateApproval();

  error SessionTooShort();
  error SessionNotFound();
  error SessionAlreadyClosed();

  error BidNotFound();
  error CannotDecodeAbi();

  //===========================
  //         SESSION
  //===========================

  /// @notice returns session by sessionId
  function getSession(bytes32 sessionId) public view returns (Session memory) {
    return s.sessions[s.sessionMap[sessionId]];
  }

  function getActiveSessionsByUser(address user) public view returns (Session[] memory) {
    Uint256Set.Set storage userSessions = s.userActiveSessions[user];
    uint256 size = userSessions.count();
    Session[] memory sessions = new Session[](size);
    for (uint i = 0; i < size; i++) {
      sessions[i] = s.sessions[userSessions.keyAtIndex(i)];
    }
    return sessions;
  }

  function getActiveSessionsByProvider(address provider) public view returns (Session[] memory) {
    Uint256Set.Set storage providerSessions = s.providerActiveSessions[provider];
    uint256 size = providerSessions.count();
    Session[] memory sessions = new Session[](size);
    for (uint i = 0; i < size; i++) {
      sessions[i] = s.sessions[providerSessions.keyAtIndex(i)];
    }
    return sessions;
  }

  function getSessionsByProvider(address provider, uint256 offset, uint8 limit) public view returns (Session[] memory) {
    return paginate(s.providerSessions[provider], offset, limit);
  }

  function getSessionsByUser(address user, uint256 offset, uint8 limit) public view returns (Session[] memory) {
    return paginate(s.userSessions[user], offset, limit);
  }

  function getSessionsByModel(bytes32 modelId, uint256 offset, uint8 limit) public view returns (Session[] memory) {
    return paginate(s.modelSessions[modelId], offset, limit);
  }

  function paginate(uint256[] memory indexes, uint256 offset, uint8 limit) private view returns (Session[] memory) {
    uint256 length = indexes.length;
    if (length < offset) {
      return (new Session[](0));
    }
    uint8 size = offset + limit > length ? uint8(length - offset) : limit;
    Session[] memory sessions = new Session[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      sessions[i] = s.sessions[indexes[index]];
    }
    return sessions;
  }

  function activeSessionsCount() public view returns (uint256) {
    return s.activeSessionsCount;
  }

  function sessionsCount() public view returns (uint256) {
    return s.sessions.length;
  }

  function openSession(
    uint256 _stake,
    bytes memory providerApproval,
    bytes memory signature
  ) public returns (bytes32 sessionId) {
    address sender = msg.sender;

    // reverts without specific error if cannot decode abi
    (bytes32 bidId, uint128 timestampMs) = abi.decode(providerApproval, (bytes32, uint128));

    if (timestampMs / 1000 < block.timestamp - SIGNATURE_TTL) {
      revert SignatureExpired();
    }

    Bid memory bid = s.bidMap[bidId];
    if (bid.deletedAt != 0 || bid.createdAt == 0) {
      revert BidNotFound();
    }

    if (!isValidReceipt(bid.provider, providerApproval, signature)) {
      revert ProviderSignatureMismatch();
    }

    if (s.approvalMap[providerApproval]) {
      revert DuplicateApproval();
    }
    s.approvalMap[providerApproval] = true;

    uint256 startOfToday = startOfTheDay(block.timestamp);
    uint256 duration = stakeToStipend(_stake, startOfToday) / bid.pricePerSecond;

    if (duration < MIN_SESSION_DURATION) {
      revert SessionTooShort();
    }

    sessionId = keccak256(abi.encodePacked(sender, bid.provider, _stake, block.number));
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
        openedAt: uint128(block.timestamp),
        endsAt: whenSessionEnds(_stake, bid.pricePerSecond, block.timestamp),
        closedAt: 0
      })
    );

    uint256 sessionIndex = s.sessions.length - 1;
    s.sessionMap[sessionId] = sessionIndex;
    s.userSessions[sender].push(sessionIndex);
    s.providerSessions[bid.provider].push(sessionIndex);
    s.modelSessions[bid.modelAgentId].push(sessionIndex);

    s.userActiveSessions[sender].insert(sessionIndex);
    s.providerActiveSessions[bid.provider].insert(sessionIndex);
    s.activeSessionsCount++;

    emit SessionOpened(sender, sessionId, bid.provider);
    s.token.transferFrom(sender, address(this), _stake); // errors with Insufficient Allowance if not approved

    return sessionId;
  }

  function closeSession(bytes memory receiptEncoded, bytes memory signature) public {
    // reverts without specific error if cannot decode abi
    (bytes32 sessionId, uint128 timestampMs, ) = abi.decode(receiptEncoded, (bytes32, uint128, uint32));
    if (timestampMs / 1000 < block.timestamp - SIGNATURE_TTL) {
      revert SignatureExpired();
    }

    uint256 sessionIndex = s.sessionMap[sessionId];
    Session storage session = s.sessions[sessionIndex];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    if (session.user != msg.sender && session.provider != msg.sender) {
      revert NotUserOrProvider();
    }
    if (session.closedAt != 0) {
      revert SessionAlreadyClosed();
    }

    // update indexes
    s.userActiveSessions[session.user].remove(sessionIndex);
    s.providerActiveSessions[session.provider].remove(sessionIndex);
    s.activeSessionsCount--;

    // update session record
    session.closeoutReceipt = receiptEncoded;
    session.closedAt = uint128(block.timestamp);

    // calculate provider withdraw
    uint256 providerWithdraw;
    bool isClosingLate = startOfTheDay(block.timestamp) > startOfTheDay(session.endsAt);
    bool noDispute = isValidReceipt(session.provider, receiptEncoded, signature);

    if (noDispute || isClosingLate) {
      // session was closed without dispute or next day after it expected to end
      uint256 duration = minUint256(block.timestamp, session.endsAt) - session.openedAt;
      uint256 cost = duration * session.pricePerSecond;
      providerWithdraw = cost - session.providerWithdrawnAmount;
    } else {
      // session was closed on the same day or earlier with dispute
      // withdraw all funds except for today's session cost
      uint256 durationTillToday = startOfTheDay(block.timestamp) -
        minUint256(session.openedAt, startOfTheDay(block.timestamp));
      uint256 costTillToday = durationTillToday * session.pricePerSecond;
      providerWithdraw = costTillToday - session.providerWithdrawnAmount;
    }

    if (!noDispute) {
      session.closeoutType = 1;
    }

    session.providerWithdrawnAmount += providerWithdraw;

    // calculate user withdraw
    uint256 userStakeToLock = 0;
    if (!isClosingLate) {
      // session was closed on the same day
      // lock today's stake
      uint256 todaysDuration = minUint256(session.endsAt, block.timestamp) -
        maxUint256(startOfTheDay(block.timestamp), session.openedAt);
      uint256 todaysCost = todaysDuration * session.pricePerSecond;
      userStakeToLock = stipendToStake(todaysCost, startOfTheDay(block.timestamp));
      s.userOnHold[session.user].push(
        OnHold({ amount: userStakeToLock, releaseAt: uint128(block.timestamp + 1 days) })
      );
    }

    uint256 userWithdraw = session.stake - userStakeToLock;

    emit SessionClosed(session.user, sessionId, session.provider);

    // withdraw provider and user funds
    s.token.transferFrom(s.fundingAccount, session.provider, providerWithdraw);
    s.token.transfer(session.user, userWithdraw);
  }

  // funds related functions

  /// @notice returns total claimanble balance for the provider for particular session
  function getProviderClaimableBalance(bytes32 sessionId) public view returns (uint256) {
    Session memory session = s.sessions[s.sessionMap[sessionId]];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    return _getProviderClaimableBalance(session);
  }

  /// @notice allows provider to claim their funds
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
    s.totalClaimed += amountToWithdraw;
    s.token.transferFrom(s.fundingAccount, to, amountToWithdraw);
    return;
  }

  function _getProviderClaimableBalance(Session memory session) internal view returns (uint256) {
    // if session was closed with no dispute - provider already got all funds
    //
    // if session was closed with dispute   -
    // if session was ended but not closed  -
    // if session was not ended             - provider can claim all funds except for today's session cost

    uint256 claimIntervalEnd = minUint256(startOfTheDay(block.timestamp), session.endsAt);
    uint256 claimableDuration = maxUint256(claimIntervalEnd, session.openedAt) - session.openedAt;
    uint256 totalCost = claimableDuration * session.pricePerSecond;
    uint256 withdrawableAmount = totalCost - session.providerWithdrawnAmount;

    return withdrawableAmount;
  }

  /// @notice deletes session from the history
  function deleteHistory(bytes32 sessionId) public {
    Session storage session = s.sessions[s.sessionMap[sessionId]];
    LibOwner._senderOrOwner(session.user);
    session.user = address(0);
  }

  /// @notice checks if receipt is valid
  function isValidReceipt(address signer, bytes memory receipt, bytes memory signature) public pure returns (bool) {
    if (signature.length == 0) {
      return false;
    }
    bytes32 receiptHash = MessageHashUtils.toEthSignedMessageHash(keccak256(receipt));
    return ECDSA.recover(receiptHash, signature) == signer;
  }

  /// @notice returns amount of withdrawable user stake and one on hold
  function withdrawableUserStake(address userAddr) public view returns (uint256 avail, uint256 hold) {
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

  /// @notice withdraws user stake
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

  /// @notice returns stipend of user based on their stake
  function stakeToStipend(uint256 sessionStake, uint256 timestamp) public view returns (uint256) {
    uint256 time = startOfTheDay(timestamp);
    uint256 totalMOR = totalMORSupply(time);
    uint todaysBudget = getTodaysBudget(time);
    uint stake = sessionStake / (totalMOR / todaysBudget);
    return stake;
  }

  /// @notice returns stake of user based on their stipend
  function stipendToStake(uint256 stipend, uint256 timestamp) public view returns (uint256) {
    return stipend * (totalMORSupply(timestamp) / getTodaysBudget(timestamp));
  }

  /// @dev make it pure
  function whenSessionEnds(
    uint256 sessionStake,
    uint256 pricePerSecond,
    uint256 openedAt
  ) private view returns (uint256) {
    uint256 lastDay = whenStipendLessThanDailyPrice(sessionStake, pricePerSecond);
    if (lastDay == 0) {
      lastDay = openedAt;
    }

    uint256 endTime = lastDay + stakeToStipend(sessionStake, lastDay) / pricePerSecond;

    // if session ends after today then count the next day stipend
    if (startOfTheDay(endTime) > startOfTheDay(lastDay)) {
      uint256 nextDayDuration = stakeToStipend(sessionStake, lastDay + 1 days) / pricePerSecond;
      endTime = startOfTheDay(endTime) + nextDayDuration;
    }

    return minUint256(endTime, openedAt + MAX_SESSION_DURATION);
  }

  /// @notice returns the time when stipend will be less than daily price
  function whenStipendLessThanDailyPrice(uint256 sessionStake, uint256 pricePerSecond) public view returns (uint256) {
    uint256 pricePerDay = pricePerSecond * 1 days;
    uint256 minComputeBalance = (pricePerDay * 100 * totalMORSupply(block.timestamp)) / sessionStake;
    return whenComputeBalanceIsLessThan(minComputeBalance);
  }

  /// @notice returns today's budget in MOR
  function getTodaysBudget(uint256 timestamp) public view returns (uint256) {
    return getComputeBalance(timestamp) / 100; // 1% of Compute Balance
  }

  /// @notice returns today's compute balance in MOR
  function getComputeBalance(uint256 timestamp) public view returns (uint256) {
    Pool memory pool = s.pools[3];
    uint256 periodReward = LinearDistributionIntervalDecrease.getPeriodReward(
      pool.initialReward,
      pool.rewardDecrease,
      pool.payoutStart,
      pool.decreaseInterval,
      pool.payoutStart,
      uint128(startOfTheDay(timestamp))
    );

    return periodReward - s.totalClaimed;
  }

  // returns total amount of MOR tokens that were distributed across all pools
  function totalMORSupply(uint256 timestamp) public view returns (uint256) {
    uint256 totalSupply = 0;
    for (uint i = 0; i < s.pools.length; i++) {
      Pool memory pool = s.pools[i];
      uint256 sup = LinearDistributionIntervalDecrease.getPeriodReward(
        pool.initialReward,
        pool.rewardDecrease,
        pool.payoutStart,
        pool.decreaseInterval,
        pool.payoutStart,
        uint128(startOfTheDay(timestamp))
      );
      totalSupply += sup;
    }
    return totalSupply;
  }

  /// @notice returns the time when compute balance will be less than targetReward
  /// @dev returns 0 if targetReward is greater than initial reward
  function whenComputeBalanceIsLessThan(uint256 targetReward) public view returns (uint256) {
    Pool memory pool = s.pools[3];
    if (targetReward >= pool.initialReward) {
      return 0;
    }
    return ((pool.initialReward - targetReward) / pool.rewardDecrease) * pool.decreaseInterval + pool.payoutStart;
  }

  /// @notice sets distibution pool configuration
  /// @dev parameters should be the same as in Ethereum L1 Distribution contract
  /// @dev at address 0x47176B2Af9885dC6C4575d4eFd63895f7Aaa4790
  /// @dev call 'Distribution.pools(3)' where '3' is a poolId
  function setPoolConfig(uint256 index, Pool calldata pool) public {
    LibOwner._onlyOwner();
    s.pools[index] = pool;
  }

  function startOfTheDay(uint256 timestamp) public pure returns (uint256) {
    return timestamp - (timestamp % 1 days);
  }

  function minUint256(uint256 a, uint256 b) internal pure returns (uint256) {
    return a < b ? a : b;
  }

  function maxUint256(uint256 a, uint256 b) internal pure returns (uint256) {
    return a > b ? a : b;
  }
}
