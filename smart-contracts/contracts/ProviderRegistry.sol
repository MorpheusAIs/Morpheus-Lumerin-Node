// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {AddressSet} from "./KeySet.sol";
import "hardhat/console.sol";

contract ProviderRegistry is OwnableUpgradeable {
    using AddressSet for AddressSet.Set;

    struct Provider {
        string endpoint; // example 'domain.com:1234'
        uint256 stake; // stake amount
        uint256 timestamp; // timestamp of the registration
    }

    error StakeTooLow();
    error NotSenderOrOwner();
    error ProviderNotFound();

    event RegisteredUpdated(address indexed provider);
    event Deregistered(address indexed provider);
    
    // state
    uint256 public minStake;
    ERC20 public token;

    // providers storage
    AddressSet.Set set;
    mapping(address => Provider) public map;

    modifier senderOrOwner(address addr) {
        if (addr != _msgSender() && addr != owner()) {
            revert NotSenderOrOwner();
        }
        _;
    }

    function initialize(address _token) public initializer {
        token = ERC20(_token);
        __Ownable_init();
    }

    function getIds() public view returns (address[] memory) {
        return set.keys();
    }

    function getCount() public view returns(uint count) {
        return set.count();
    }

    function getByIndex(uint index) public view returns(Provider memory provider) {
        return map[set.keyAtIndex(index)];
    }

    function getAll() public view returns (Provider[] memory) {
        Provider[] memory _providers = new Provider[](set.count());
        for (uint i = 0; i < set.count(); i++) {
            _providers[i] = map[set.keyAtIndex(i)];
        }
        return _providers;
    }

    // registers new provider or updates existing
    function register(address addr, string memory endpoint) public senderOrOwner(addr) {
        uint256 amount = token.allowance(addr, address(this));
        uint256 stake = map[addr].stake;
        if (amount + stake < minStake) {
            revert StakeTooLow();
        }
        if (stake == 0) {
          set.insert(addr);
        }
        
        map[addr] = Provider(endpoint, minStake, block.timestamp);
        token.transferFrom(addr, address(this), amount);
    }

    // avoid loop this by using pointer pattern
    function deregister(address addr) public senderOrOwner(addr) {
        set.remove(addr);
        uint256 stake = map[addr].stake;
        delete map[addr];
        token.transfer(addr, stake);
    }

    function setMinStake(uint256 _minStake) public onlyOwner {
        minStake = _minStake;
    }

    // function getProvider(address)
    // just use provider mapping contract.provider[address]

    // function getStakeReq()
    // just use minStake variable

    // function updateProvider()
    // use registerProvider instead

    // function owner()
    // inherited from OwnableUpgradeable

    // function transferOwnership()
    // inherited from OwnableUpgradeable
}
