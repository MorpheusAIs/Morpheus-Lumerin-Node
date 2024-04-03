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
  error AllowanceTooLow();
  error NotSenderOrOwner();
  error ProviderNotFound();

  event RegisteredUpdated(address indexed provider);
  event Deregistered(address indexed provider);
  event MinStakeUpdated(uint256 newStake);

  // state
  uint256 public minStake;
  ERC20 public token;

  // providers storage
  AddressSet.Set set;
  mapping(address => Provider) public map;

  function initialize(address _token) public initializer {
      token = ERC20(_token);
      __Ownable_init();
  }

  function getIds() public view returns (address[] memory) {
      return set.keys();
  }

  function getCount() public view returns (uint count) {
      return set.count();
  }

  function getByIndex(uint index) public view returns (Provider memory provider) {
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
  function register(address addr, uint256 addStake, string memory endpoint) public senderOrOwner(addr) {
    uint256 stake = map[addr].stake;
    uint256 newStake = stake + addStake;
    if (newStake < minStake) {
        revert StakeTooLow();
    }
    if (stake == 0) {
        set.insert(addr);
    } else {
        _senderOrOwner(addr);
    }

    map[addr] = Provider(endpoint, newStake, block.timestamp);
    emit RegisteredUpdated(addr);
    token.transferFrom(_msgSender(), address(this), addStake); // reverts with ERC20InsufficientAllowance
  }

  // avoid loop this by using pointer pattern
  function deregister(address addr) public senderOrOwner(addr) {
    set.remove(addr);
    
    uint256 stake = map[addr].stake;
    delete map[addr];
    
    emit Deregistered(addr);
    token.transfer(addr, stake);
  }

  function setMinStake(uint256 _minStake) public onlyOwner {
    minStake = _minStake;
    emit MinStakeUpdated(_minStake);
  }

  modifier senderOrOwner(address addr) {
    _senderOrOwner(addr);
    _;
  }

  function _senderOrOwner(address addr) internal view {
    if (addr != _msgSender() && addr != owner()) {
        revert NotSenderOrOwner();
    }
  }
}
