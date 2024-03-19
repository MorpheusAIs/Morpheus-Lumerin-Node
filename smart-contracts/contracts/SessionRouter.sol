// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "hardhat/console.sol";

contract SessionRouter {
    //
    // structs
    //
    struct Provider {
        address addr;
        string name;
        string url;
        uint256 pricePerMinute;
    }

    struct Session {
        uint id;
        address buyer;
        address provider;
        uint startTime;
        uint endTime;
    }

    //
    // state
    //
    address payable public owner;
    ERC20 public tokenMOR;
    uint public constant minute = 60; // minute in seconds

    //
    // mappings
    //
    mapping(address => Provider) public providers;
    address[] public providerList;

    mapping(uint => Session) public sessions;
    uint[] public sessionList;

    //
    // events
    //
    event SessionStarted(address indexed user, uint indexed sessionId);
    event SessionClosedManually(address indexed user, uint indexed sessionId);
    event ProviderAdded(address indexed provider);
    // event ProviderRemoved(address indexed provider);

    constructor(address _morpheusTokenAddr) payable {
        owner = payable(msg.sender);
        tokenMOR = ERC20(_morpheusTokenAddr);
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "only owner can call this function");
        _;
    }

    function addProvider(address _addr, string memory _name, string memory _url, uint _pricePerMinute) public onlyOwner {
        require(providers[_addr].addr == address(0), "provider already exists");
        providers[_addr] = Provider(_addr, _name, _url, _pricePerMinute);
        providerList.push(_addr);
        emit ProviderAdded(_addr);
    }

    function startSession(address _provider) public payable returns (uint) {
        uint256 actualAllowance = tokenMOR.allowance(msg.sender, address(this));
        console.log("msg sender", msg.sender);
        console.log("allowance", actualAllowance);
        bool tokensTransfered = tokenMOR.transferFrom(
            msg.sender,
            address(this),
            actualAllowance
        );
        require(tokensTransfered, "morheus transfer failed");
        
        Provider memory provider = providers[_provider];
        require(provider.addr != address(0), "provider not found");

        uint durationMinutes = actualAllowance / provider.pricePerMinute;
        uint endTime = block.timestamp + durationMinutes * 60;

        uint sessionId = sessionList.length;
        sessions[sessionId] = Session(sessionId, msg.sender, _provider, block.timestamp, endTime);
        sessionList.push(sessionId);

        emit SessionStarted(msg.sender, sessionId);
        return sessionId;
    }

    function closeSession(uint _sessionId) public {
        Session storage session = sessions[_sessionId];
        require(session.buyer == msg.sender, "only user can close session");
        require(session.endTime > block.timestamp, "session already closed or expired");
        uint durationSeconds = block.timestamp - session.startTime;
        
        Provider memory provider = providers[session.provider];
        require(provider.addr != address(0), "provider not found");

        uint refund = durationSeconds * provider.pricePerMinute / minute;
        tokenMOR.transfer(msg.sender, refund);

        session.endTime = block.timestamp;
        emit SessionClosedManually(msg.sender, _sessionId);
    }
}
