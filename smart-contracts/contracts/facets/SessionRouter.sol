// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { MessageHashUtils } from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "../libraries/KeySet.sol";
import { ModelRegistry } from "./ModelRegistry.sol";
import { ProviderRegistry } from './ProviderRegistry.sol';
import { AppStorage, Session, Bid, OnHold } from "../AppStorage.sol";
import { LibOwner } from '../libraries/LibOwner.sol';

contract SessionRouter {
  using KeySet for KeySet.Set;
  AppStorage internal s;

  // constants
  uint32 constant DAY = 24*60*60; // 1 day
  uint32 constant MIN_SESSION_DURATION = 5*60; // 5 minutes

  // events
  event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);
  event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);
  event Staked(address indexed userAddress, uint256 amount);
  event Unstaked(address indexed userAddress, uint256 amount);
  event ProviderClaimed(address indexed providerAddress, uint256 amount);

  // errors
  error NotUserOrProvider();
  error NotUser();

  error NotEnoughWithdrawableBalance();
  error NotEnoughStipend();
  error NotEnoughStake();
  error NotEnoughBalance();

  error InvalidSignature();
  error SessionTooShort();
  error SessionNotFound();

  error BidNotFound();
  error BidTaken();

  //===========================
  //         SESSION
  //===========================

  function getSession(bytes32 sessionId) public view returns (Session memory) {
    return s.sessions[s.sessionMap[sessionId]];
  }

  function openSession(bytes32 bidId, uint256 _stake) public returns (bytes32 sessionId){
    address sender = msg.sender;

    Bid memory bid = s.bidMap[bidId];
    if (bid.deletedAt != 0 || bid.createdAt == 0) {
      revert BidNotFound();
    }

    if (s.bidSessionMap[bidId] != 0){
      // TODO: some bids might be already taken by other sessions
      // but the session list in marketplace is ignorant of this fact.
      // Marketplace and SessionRouter contracts should be merged together 
      // to avoid this issue and update indexes by avoiding costly intercontract calls
      revert BidTaken();
    }

    uint256 duration = balanceOfSessionStipend(_stake) / bid.pricePerSecond;
    if (duration < MIN_SESSION_DURATION) {
      revert SessionTooShort();
    }

    sessionId = keccak256(abi.encodePacked(sender, bid.provider, _stake, block.number));
    s.sessions.push(Session({
      id: sessionId,
      user: sender,
      provider: bid.provider,
      modelAgentId: bid.modelAgentId,
      bidID: bidId,
      stake: _stake,
      pricePerSecond: bid.pricePerSecond,
      closeoutReceipt: "",
      closeoutType: 0,
      openedAt: block.timestamp,
      closedAt: 0
    }));

    uint256 sessionIndex = s.sessions.length - 1;
    s.sessionMap[sessionId] = sessionIndex;
    s.bidSessionMap[bidId] = sessionIndex; // marks bid as "taken" by this session

    emit SessionOpened(sender, sessionId, bid.provider);

    s.token.transferFrom(sender, address(this), _stake); // errors with Insufficient Allowance if not approved

    return sessionId;
  }

  // returns expected session duration in seconds
  // should be called daily 00:00:00 UTC
  // returns 24 hours if session should not be closed today
  function getExpectedDuration(uint256 _stake, uint256 pricePerSecond) public view returns (uint256) {
    uint256 stipend = balanceOfSessionStipend(_stake);
    if (stipend > pricePerSecond) {
      return DAY;
    }
    return stipend / pricePerSecond * 60 * 60;
  }

  function closeSession(bytes32 sessionId, bytes memory receiptEncoded, bytes memory signature) public {
    Session storage session = s.sessions[s.sessionMap[sessionId]];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    if (session.user != msg.sender && session.provider != msg.sender) {
      revert NotUserOrProvider();
    }

    s.bidSessionMap[session.bidID] = 0;  // marks bid as available
    session.closeoutReceipt = receiptEncoded;
    session.closedAt = block.timestamp;

    uint256 durationSeconds = session.closedAt - session.openedAt;
    uint256 cost = durationSeconds * session.pricePerSecond;

    // TODO: partially return stake according to the usage
    // and put rest on hold for 24 hours

    if (isValidReceipt(session.provider, receiptEncoded, signature)){
      s.token.transfer(session.provider, cost);
    } else {
      session.closeoutType = 1;
      s.providerOnHold[session.provider].push(OnHold({
        amount: cost,
        releaseAt: block.timestamp + DAY
      }));
    }
  }
  // funds related functions

  function getProviderBalance(address providerAddr) public view returns (uint256 total, uint256 hold) {
    OnHold[] memory onHold = s.providerOnHold[providerAddr];
    for (uint i = 0; i < onHold.length; i++) {
      total += onHold[i].amount;
      if (block.timestamp < onHold[i].releaseAt) {
        hold+=onHold[i].amount;
      }
    }
    return (total, hold);
  }

  // transfers provider claimable balance to provider address.
  // set amount to 0 to claim all balance.
  function claimProviderBalance(uint256 amountToWithdraw, address to) public {
    uint256 balance = 0;
    address sender = msg.sender;
    
    OnHold[] storage onHoldEntries = s.providerOnHold[sender];
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

        onHoldEntries[i] = onHoldEntries[onHoldEntries.length-1];
        onHoldEntries.pop();
      } else {
        i++;
      }
    }

    revert NotEnoughBalance();
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

  function isValidReceipt(address signer, bytes memory receipt, bytes memory signature) public pure returns (bool) {
    if (signature.length == 0){
      return false;
    }
    bytes32 receiptHash = MessageHashUtils.toEthSignedMessageHash(keccak256(receipt));
    return ECDSA.recover(receiptHash, signature) == signer;
  }

  //===========================
  //         STAKING
  //===========================

  function withdrawableStakeBalance(address userAddress) public view returns (uint256) {
    //TODO: return user stake (on hold and withdrawable)
    return 0;
  }

  // return virtual MOR balance of user based on their stake
  // DEPRECATED
  // function balanceOfDailyStipend(address userAddress) public view returns (uint256) {
  //   return getTodaysBudget() * userStake[userAddress] / token.totalSupply() - getTodaysSpend(userAddress);
  // }

  function balanceOfSessionStipend(uint256 sessionStake) public view returns (uint256) {
    return getTodaysBudget() * sessionStake / s.token.totalSupply();
  }


  function getTodaysSpend(address userAddress) public view returns (uint256) {
    // OnHold memory spend = todaysSpend[userAddress];
    // if (block.timestamp > spend.releaseAt) {
    //   return 0;
    // }
    // return spend.amount;
    //
    // TODO: implement global counter of how much was spent today
    return 0; 
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

  //===========================
  //        BIDS
  //===========================

  

  //===========================
  //     ACCESS CONTROL
  //===========================
  
  
}