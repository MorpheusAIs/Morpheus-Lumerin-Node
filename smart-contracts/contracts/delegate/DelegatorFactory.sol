// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Create2} from "@openzeppelin/contracts/utils/Create2.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/security/PausableUpgradeable.sol";
import {UpgradeableBeacon} from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import {BeaconProxy} from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";

import {IProvidersDelegator} from "../interfaces/delegate/IProvidersDelegator.sol";
import {IOwnable} from "../interfaces/utils/IOwnable.sol";

contract DelegatorFactory is OwnableUpgradeable, PausableUpgradeable, UUPSUpgradeable {
    address public lumerinDiamond;
    address public beacon;
    mapping(address => address[]) public proxies;

    event ProxyDeployed(address indexed proxyAddress);
    event ImplementationUpdated(address indexed newImplementation);

    constructor() {
        _disableInitializers();
    }

    function DelegatorFactory_init(address _lumerinDiamond, address _implementation) external initializer {
        __Pausable_init();
        __Ownable_init();
        __UUPSUpgradeable_init();

        lumerinDiamond = _lumerinDiamond;

        beacon = address(new UpgradeableBeacon(_implementation));
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    function deployProxy(
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_
    ) external whenNotPaused returns (address) {
        bytes32 salt_ = _calculatePoolSalt(_msgSender());
        address proxy_ = address(new BeaconProxy{salt: salt_}(beacon, bytes("")));

        proxies[_msgSender()].push(address(proxy_));

        IProvidersDelegator(proxy_).ProvidersDelegator_init(lumerinDiamond, feeTreasury_, fee_, name_, endpoint_);
        IOwnable(proxy_).transferOwnership(_msgSender());

        emit ProxyDeployed(address(proxy_));

        return address(proxy_);
    }

    function predictProxyAddress(address _deployer) external view returns (address) {
        bytes32 salt_ = _calculatePoolSalt(_deployer);

        bytes32 bytecodeHash_ = keccak256(
            abi.encodePacked(type(BeaconProxy).creationCode, abi.encode(address(beacon), bytes("")))
        );

        return Create2.computeAddress(salt_, bytecodeHash_);
    }

    function updateImplementation(address _newImplementation) external onlyOwner {
        UpgradeableBeacon(beacon).upgradeTo(_newImplementation);

        emit ImplementationUpdated(_newImplementation);
    }

    function version() external pure returns (uint256) {
        return 1;
    }

    function _calculatePoolSalt(address sender_) internal view returns (bytes32) {
        return keccak256(abi.encodePacked(sender_, proxies[sender_].length));
    }

    function _authorizeUpgrade(address) internal view override onlyOwner {}
}
