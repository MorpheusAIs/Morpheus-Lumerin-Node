// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { MessageHashUtils } from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { KeySet } from "./KeySet.sol";
import { ModelRegistry } from "./ModelRegistry.sol";
import { ProviderRegistry } from './ProviderRegistry.sol';
import "hardhat/console.sol";

contract SessionRouter is OwnableUpgradeable {
  using KeySet for KeySet.Set;

  struct Session {
    bytes32 id;
    address user;
    address provider;
    bytes32 modelAgentId;
    bytes32 bidID;
    uint256 budget;
    uint256 price;
    bytes closeoutReceipt;
    uint256 closeoutType;
    uint256 openedAt;
    uint256 closedAt;
  }

  struct Bid {
    address provider;
    bytes32 modelAgentId;
    uint256 amount;
    uint256 nonce;
    uint256 createdAt;
    uint256 deletedAt;
  }

  struct OnHold {
    uint256 amount;
    uint256 releaseAt; // in epoch seconds TODO: consider using hours to reduce storage cost
  }

  // state
  // Number of seconds to delay the stake return when a user closes out a session using a user signed receipt.
  // not clear who is going to trigger this call
  int256 public stakeDelay;
  uint256 public modelBidFee;
  uint256 public agentBidFee;
  uint256 public feeBalance;

  // dependencies
  address public tokenAccount; // account which stores the MOR tokens with infinite allowance for this contract
  IERC20 public token;
  ModelRegistry public modelRegistry;
  ProviderRegistry public providerRegistry;

  // arguments for getPeriodReward call
  // address public constant distributionContractAddr = address(0x0);
  // uint32 public constant distributionRewardStartTime = 1707350400; // ephochSeconds Feb 8 2024 00:00:00
  // uint8 public constant distributionPoolId = 3;

  // storage - sessions
  Session[] public sessions;
  mapping(bytes32 => uint256) public sessionMap; // sessionId => session index
  mapping(bytes32 => uint256) public bidSessionMap; // bidId => session index
  
  // storage - provider stake
  mapping(address => OnHold[]) public providerOnHold; // provider address => balance

  // storage - users stake
  mapping(address => uint256) public userStake; // user address => stake balance
  mapping(address => OnHold) public todaysSpend; // user address => spend today

  // storage - bids
  mapping(bytes32 => Bid) public bidMap; // bidId = keccak256(provider, modelAgentId, nonce) => bid
  mapping(bytes32 => uint256) public providerModelAgentNonce; // keccak256(provider, modelAgentId) => last nonce

  // storage - active bids indexes
  KeySet.Set activeBids; // all active bidIds
  mapping(address => KeySet.Set) providerActiveBids; // provider => active bidIds
  mapping(bytes32 => KeySet.Set) modelAgentActiveBids; // modelAgentId => active bidIds

  // storage - all bids indexes
  mapping(bytes32 => bytes32[]) modelAgentBids; // keccak256(provider, modelAgentId) => all bidIds
  mapping(address => bytes32[]) providerBids; // provider => all bidIds

  // constants
  uint32 constant DAY = 24*60*60; // 1 day
  uint32 constant MIN_SESSION_DURATION = 5*60; // 5 minutes

  // events
  event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);
  event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId);
  event Staked(address indexed userAddress, uint256 amount);
  event Unstaked(address indexed userAddress, uint256 amount);
  event ProviderClaimed(address indexed providerAddress, uint256 amount);
  event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
  event FeeUpdated(uint256 modelFee, uint256 agentFee);

  // errors
  error NotUserOrProvider();
  error NotUser();
  error NotSenderOrOwner();

  error NotEnoughBalance();
  error NotEnoughWithdrawableBalance();
  error NotEnoughStipend();
  error NotEnoughStake();
  
  error BidNotFound();
  error BidTaken();

  error InvalidSignature();
  error SessionTooShort();
  error SessionNotFound();

  error ModelOrAgentNotFound();
  error ActiveBidNotFound();

  function initialize(
    address _token, 
    address _tokenAccount,
    address modelRegistryAddr,
    address providerRegistryAddr
  ) public initializer {
    __Ownable_init();

    token = IERC20(_token);
    modelRegistry = ModelRegistry(modelRegistryAddr);
    providerRegistry = ProviderRegistry(providerRegistryAddr);

    stakeDelay = 0;
    tokenAccount = _tokenAccount;
    sessions.push(Session({
      id: bytes32(0),
      user: address(0),
      provider: address(0),
      modelAgentId: bytes32(0),
      bidID: bytes32(0),
      budget: 0,
      price: 0,
      closeoutReceipt: "",
      closeoutType: 0,
      openedAt: 0,
      closedAt: 0
    }));
  }

  //===========================
  //         SESSION
  //===========================

  function getSession(bytes32 sessionId) public view returns (Session memory) {
    return sessions[sessionMap[sessionId]];
  }

  function openSession(bytes32 bidId, uint256 budget) public returns (bytes32 sessionId){
    address sender = _msgSender();
    uint256 stipend = balanceOfDailyStipend(sender);
    if (budget > stipend) {
      revert NotEnoughStipend();
    }

    Bid memory bid = bidMap[bidId];
    if (bid.deletedAt != 0 || bid.createdAt == 0) {
      revert BidNotFound();
    }

    if (bidSessionMap[bidId] != 0){
      // TODO: some bids might be already taken by other sessions
      // but the session list in marketplace is ignorant of this fact.
      // Marketplace and SessionRouter contracts should be merged together 
      // to avoid this issue and update indexes by avoiding costly intercontract calls
      revert BidTaken();
    }

    uint256 duration = budget / bid.amount;
    if (duration < MIN_SESSION_DURATION) {
      revert SessionTooShort();
    }

    sessionId = keccak256(abi.encodePacked(sender, bid.provider, budget, block.number));
    sessions.push(Session({
      id: sessionId,
      user: sender,
      provider: bid.provider,
      modelAgentId: bid.modelAgentId,
      bidID: bidId,
      budget: budget,
      price: bid.amount,
      closeoutReceipt: "",
      closeoutType: 0,
      openedAt: block.timestamp,
      closedAt: 0
    }));

    uint256 sessionIndex = sessions.length - 1;
    sessionMap[sessionId] = sessionIndex;
    bidSessionMap[bidId] = sessionIndex; // marks bid as "taken" by this session

    emit SessionOpened(sender, sessionId, bid.provider);

    transferDailyStipend(sender, address(this), budget);
    return sessionId;
  }

  function closeSession(bytes32 sessionId, bytes memory receiptEncoded, bytes memory signature) public {
    Session storage session = sessions[sessionMap[sessionId]];
    if (session.openedAt == 0) {
      revert SessionNotFound();
    }
    if (session.user != _msgSender() && session.provider != _msgSender()) {
      revert NotUserOrProvider();
    }

    bidSessionMap[session.bidID] = 0;  // marks bid as available
    session.closeoutReceipt = receiptEncoded;
    session.closedAt = block.timestamp;

    uint256 durationSeconds = session.closedAt - session.openedAt;
    uint256 cost = durationSeconds * session.price;

    if (cost < session.budget) {
      uint256 refund = session.budget - cost;
      returnStipend(session.user, refund);
    } 

    if (isValidReceipt(session.provider, receiptEncoded, signature)){
      token.transfer(session.provider, cost);
    } else {
      session.closeoutType = 1;
      providerOnHold[session.provider].push(OnHold({
        amount: cost,
        releaseAt: block.timestamp + DAY
      }));
    }
  }
  // funds related functions

  function getProviderBalance(address providerAddr) public view returns (uint256 total, uint256 hold) {
    OnHold[] memory onHold = providerOnHold[providerAddr];
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
    address sender = _msgSender();
    
    OnHold[] storage onHoldEntries = providerOnHold[sender];
    uint i = 0;
    // the only loop that is not avoidable
    while (i < onHoldEntries.length) {
      if (block.timestamp > onHoldEntries[i].releaseAt) {
        balance += onHoldEntries[i].amount;


        if (balance >= amountToWithdraw) {
          uint256 delta = balance - amountToWithdraw;
          onHoldEntries[i].amount = delta;
          token.transfer(to, amountToWithdraw);
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
    Session storage session = sessions[sessionMap[sessionId]];
    _senderOrOwner(session.user);
    session.user = address(0);
  }

  function setStakeDelay(int256 delay) public onlyOwner {
    stakeDelay = delay;
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

  function transferDailyStipend(address from, address to, uint256 amount) internal /*onlyOwnerOrSessionRouter*/{
    if (amount > balanceOfDailyStipend(from)) {
      revert NotEnoughStipend();
    }
    todaysSpend[from] = OnHold({
      amount: getTodaysSpend(from) + amount,
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

  //===========================
  //        BIDS
  //===========================

  function getActiveBidsByProvider(address provider) public view returns (Bid[] memory) {
    KeySet.Set storage providerBidsSet = providerActiveBids[provider];
    Bid[] memory _bids = new Bid[](providerBidsSet.count());
    for (uint i = 0; i < providerBidsSet.count(); i++) {
      _bids[i] = bidMap[providerBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  // returns active bids by model agent
  function getActiveBidsByModelAgent(bytes32 modelAgentId) public view returns (Bid[] memory) {
    KeySet.Set storage modelAgentBidsSet = modelAgentActiveBids[modelAgentId];
    Bid[] memory _bids = new Bid[](modelAgentBidsSet.count());
    for (uint i = 0; i < modelAgentBidsSet.count(); i++) {
      _bids[i] = bidMap[modelAgentBidsSet.keyAtIndex(i)];
    }
    return _bids;
  }

  // returns all bids by provider sorted from newest to oldest
  function getBidsByProvider(address provider, uint256 offset, uint8 limit) public view returns (Bid[] memory) {
    uint256 length = providerBids[provider].length;
    if (length < offset){
      return new Bid[](0);
    }
    uint8 size = offset + limit > length ? uint8(length-offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = providerBids[provider][index];
      _bids[i]=bidMap[id];
    }
    return _bids;
  }

  // returns all bids by provider sorted from newest to oldest
  function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) public view returns (Bid[] memory) {
    uint256 length = modelAgentBids[modelAgentId].length;
    if (length < offset){
      return new Bid[](0);
    }
    uint8 size = offset + limit > length ? uint8(length-offset) : limit;
    Bid[] memory _bids = new Bid[](size);
    for (uint i = 0; i < size; i++) {
      uint256 index = length - offset - i - 1;
      bytes32 id = modelAgentBids[modelAgentId][index];
      _bids[i]=bidMap[id];
    }
    return _bids;
  }

  function postModelBid(address providerAddr, bytes32 modelId, uint256 amount) public  senderOrOwner(providerAddr) returns (bytes32 bidId){
    if (!providerRegistry.exists(providerAddr)){
      revert ModelOrAgentNotFound();
    }
    if (!modelRegistry.exists(modelId)){
      revert ModelOrAgentNotFound();
    }

    return postModelAgentBid(providerAddr, modelId, amount);
  }

  function postModelAgentBid(address provider, bytes32 modelAgentId, uint256 amount) internal returns (bytes32 bidId){
    // remove old bid

    // TEST IT if it increments nonce correctly
    uint256 nonce = providerModelAgentNonce[keccak256(abi.encodePacked(provider, modelAgentId))]++;
    if (nonce > 0) {
      deleteModelAgentBid(keccak256(abi.encodePacked(provider, modelAgentId, nonce-1)));
    }

    bidId = keccak256(abi.encodePacked(provider, modelAgentId, nonce));

    bidMap[bidId] = Bid({
      provider: provider,
      modelAgentId: modelAgentId,
      amount: amount,
      nonce: nonce,
      createdAt: block.timestamp,
      deletedAt: 0
    });

    // active indexes
    activeBids.insert(bidId);
    providerActiveBids[provider].insert(bidId);
    modelAgentActiveBids[modelAgentId].insert(bidId);
    
    // all indexes
    providerBids[provider].push(bidId);
    modelAgentBids[modelAgentId].push(bidId);

    emit BidPosted(provider, modelAgentId, nonce);
    
    token.transferFrom(_msgSender(), address(this), modelBidFee);
    feeBalance += modelBidFee;

    return bidId;
  }

  function deleteModelAgentBid(bytes32 bidId) public {
    Bid storage bid = bidMap[bidId];
    if (bid.createdAt == 0 || bid.deletedAt != 0) {
      revert ActiveBidNotFound();
    }

    _senderOrOwner(bid.provider);

    bid.deletedAt = block.timestamp;
    // indexes update
    activeBids.remove(bidId);
    providerActiveBids[bid.provider].remove(bidId);
    modelAgentActiveBids[bid.modelAgentId].remove(bidId);

    emit BidDeleted(bid.provider, bid.modelAgentId, bid.nonce);
  }

  function setBidFee(uint256 modelFee, uint256 agentFee) public onlyOwner {
    modelBidFee = modelFee;
    agentBidFee = agentFee;
    emit FeeUpdated(modelFee, agentFee);
  }

  function withdraw(address addr, uint256 amount) public onlyOwner {
    if (amount > feeBalance) {
      revert NotEnoughBalance();
    }
    // emits ERC-20 transfer event
    feeBalance -= amount;
    token.transfer(addr, amount);
  }

  //===========================
  //     ACCESS CONTROL
  //===========================
  
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