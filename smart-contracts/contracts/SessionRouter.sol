// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "./KeySet.sol";
import { Marketplace } from "./Marketplace.sol";
import { StakingDailyStipend } from './StakingDailyStipend.sol';
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
    bytes closeoutReceipt;
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
  StakingDailyStipend public stakingDailyStipend;
  IERC20 public token;

  // arguments for getPeriodReward call
  // address public constant distributionContractAddr = address(0x0);
  // uint32 public constant distributionRewardStartTime = 1707350400; // ephochSeconds Feb 8 2024 00:00:00
  // uint8 public constant distributionPoolId = 3;

  // storage
  Session[] public sessions;
  mapping(bytes32 => uint256) public map; // sessionId => index

  mapping(address => OnHold[]) public providerOnHold; // user address => balance
  // mapping(address => OnHold) public todaysSpend; // user address => spend today
  
  // mapping(address => uint256) public userStake; // user address => balance
  // mapping(address => OnHold) public userStakeOnHold; // user address => amount on hold - user funds that put on hold
  // mapping(address => OnHold[]) public providerOnHold; // user address => amount on hold - provider funds that put on hold 

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
  error NotSenderOrOwner();
  error NotEnoughBalance();
  error NotEnoughStipend();
  error BidNotFound();
  error InvalidSignature();
  error SessionTooShort();
  error SessionNotFound();

  function initialize(address _token, address _stakingDailyStipend, address _marketplace) public initializer {
    __Ownable_init();
    token = IERC20(_token);
    stakingDailyStipend = StakingDailyStipend(_stakingDailyStipend);
    marketplace = Marketplace(_marketplace);
    stakeDelay = 0;
    sessions.push(Session({
      id: bytes32(0),
      user: address(0),
      provider: address(0),
      modelAgentId: bytes32(0),
      budget: 0,
      price: 0,
      closeoutReceipt: "",
      closeoutType: 0,
      openedAt: 0,
      closedAt: 0
    }));
  }

  function openSession(bytes32 bidId, uint256 budget) public returns (bytes32 sessionId){
    address sender = _msgSender();
    uint256 stipend = stakingDailyStipend.balanceOfDailyStipend(sender);
    if (budget > stipend) {
      revert NotEnoughStipend();
    }

    (address provider, bytes32 modelAgentId, uint256 amount, , uint256 createdAt, uint256 deletedAt)  = marketplace.map(bidId);
    if (deletedAt != 0 || createdAt == 0) {
      revert BidNotFound();
    }

    uint256 duration = budget / amount;
    if (duration < MIN_SESSION_DURATION) {
      revert SessionTooShort();
    }

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

    emit SessionOpened(sender, sessionId, provider);

    stakingDailyStipend.transferDailyStipend(sender, address(this), budget);
    return sessionId;
  }


  function getEthSignedMessageHash(bytes memory _encodedMessage) public pure returns (bytes32) {
    return keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", keccak256(_encodedMessage)));
  }

  function closeSession(bytes32 sessionId, bytes memory receiptEncoded, bytes memory signature) public {
    Session storage session = sessions[map[sessionId]];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    if (session.user != _msgSender() && session.provider != _msgSender()) {
      revert NotUserOrProvider();
    }

    session.closeoutReceipt = receiptEncoded;
    session.closedAt = block.timestamp;

    bool isDispute = true;

    if (signature.length != 0){
      bytes32 receiptHash = getEthSignedMessageHash(receiptEncoded);
      address signer = recoverSigner(receiptHash, signature);
      
      if (signer == session.provider) {
        isDispute = false;
      }
    }

    uint256 durationSeconds = session.closedAt - session.openedAt;
    uint256 cost = durationSeconds * session.price;

    if (cost < session.budget) {
      uint256 refund = session.budget - cost;
      token.approve(address(stakingDailyStipend), refund);
      stakingDailyStipend.returnStipend(session.user, refund);
    }

    if (isDispute){
      session.closeoutType = 1;
      providerOnHold[session.provider].push(OnHold({
        amount: session.budget,
        releaseAt: block.timestamp + DAY
      }));
    } else {
      token.transfer(session.provider, cost);
    }
  }


  // funds related functions

  // Returns claimable balance by provider address.
  function getProviderClaimBalance(address providerAddr) public view returns (uint256) {
   
  }

  function getProviderBalance(address providerAddr) public view returns (uint256 total, uint256 hold) {
    OnHold[] memory onHold = providerOnHold[providerAddr];
    for (uint i = 0; i < onHold.length; i++) {
      hold += onHold[i].amount;
      if (onHold[i].releaseAt < block.timestamp) {
        total+=onHold[i].amount;
      }
    }
    return (total, hold);
  }

  // transfers provider claimable balance to provider address.
  // set amount to 0 to claim all balance.
  function claimProviderBalance(uint256 amount, address to) public {
    uint256 balance = 0;
    address sender = _msgSender();
    // the only loop that is not avoidable
    uint i = 0;

    OnHold[] storage onHoldEntries = providerOnHold[sender];
    while (i < onHoldEntries.length) {
      if (onHoldEntries[i].releaseAt < block.timestamp) {
        balance += onHoldEntries[i].amount;

        if (balance >= amount) {
          uint256 delta = balance - amount;
          onHoldEntries[i].amount = delta;
          token.transfer(to, amount);
          return;
        } 

        onHoldEntries[i] = onHoldEntries[onHoldEntries.length-1];
        onHoldEntries.pop();
      } else {
        i++;
      }
    }
    if (amount == 0) {
      amount = balance;
      token.transfer(to, amount);
      return;
    }

    revert NotEnoughBalance();
  }

  function deleteHistory(bytes32 sessionId) public {
    Session storage session = sessions[map[sessionId]];
    _senderOrOwner(session.user);
    session.user = address(0);
  }

  function setStakeDelay(int256 delay) public onlyOwner {
    stakeDelay = delay;
  }

  function recoverSigner(bytes32 message, bytes memory sig) public pure returns (address) {
    (bytes32 r, bytes32 s, uint8 v) = splitSignature(sig);
    return ecrecover(message, v, r, s);
  }

  function splitSignature(bytes memory sig) public pure returns (bytes32 r, bytes32 s, uint8 v) {
    if (sig.length != 65) {
      revert InvalidSignature();
    }

    assembly {
        /*
        First 32 bytes stores the length of the signature

        add(sig, 32) = pointer of sig + 32
        effectively, skips first 32 bytes of signature

        mload(p) loads next 32 bytes starting at the memory address p into memory
        */

        // first 32 bytes, after the length prefix
        r := mload(add(sig, 32))
        // second 32 bytes
        s := mload(add(sig, 64))
        // final byte (first byte of the next 32 bytes)
        v := byte(0, mload(add(sig, 96)))
    }

    // implicitly return (r, s, v)
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