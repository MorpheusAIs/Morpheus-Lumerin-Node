// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Create2} from "@openzeppelin/contracts/utils/Create2.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/security/PausableUpgradeable.sol";
import {UpgradeableBeacon} from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import {BeaconProxy} from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";

import {IProvidersDelegate} from "../interfaces/delegate/IProvidersDelegate.sol";
import {IDelegateFactory} from "../interfaces/delegate/IDelegateFactory.sol";
import {IOwnable} from "../interfaces/utils/IOwnable.sol";

contract DelegateFactory is IDelegateFactory, OwnableUpgradeable, PausableUpgradeable, UUPSUpgradeable {
    address public lumerinDiamond;
    address public beacon;

    mapping(address => address[]) public proxies;
    uint128 public minDeregistrationTimeout;

    constructor() {
        _disableInitializers();
    }

    function DelegateFactory_init(
        address lumerinDiamond_,
        address implementation_,
        uint128 minDeregistrationTimeout_
        ) external initializer {
        __Pausable_init();
        __Ownable_init();
        __UUPSUpgradeable_init();

        setMinDeregistrationTimeout(minDeregistrationTimeout_);
        lumerinDiamond = lumerinDiamond_;

        beacon = address(new UpgradeableBeacon(implementation_));
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

      function setMinDeregistrationTimeout(uint128 minDeregistrationTimeout_) public onlyOwner {
        minDeregistrationTimeout = minDeregistrationTimeout_;

        emit MinDeregistrationTimeoutUpdated(minDeregistrationTimeout_);
    }

    function deployProxy(
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_,
        uint128 deregistrationOpenAt
    ) external whenNotPaused returns (address) {
        if (deregistrationOpenAt <= block.timestamp + minDeregistrationTimeout) {
            revert InvalidDeregistrationOpenAt(deregistrationOpenAt, uint128(block.timestamp + minDeregistrationTimeout));
        }

        bytes32 salt_ = _calculatePoolSalt(_msgSender());
        address proxy_ = address(new BeaconProxy{salt: salt_}(beacon, bytes("")));

        proxies[_msgSender()].push(proxy_);

        IProvidersDelegate(proxy_).ProvidersDelegate_init(
            lumerinDiamond,
            feeTreasury_,
            fee_,
            name_,
            endpoint_,
            deregistrationOpenAt
        );
        IOwnable(proxy_).transferOwnership(_msgSender());

        emit ProxyDeployed(proxy_);

        return proxy_;
    }

    function predictProxyAddress(address deployer_) external view returns (address) {
        bytes32 salt_ = _calculatePoolSalt(deployer_);

        bytes32 bytecodeHash_ = keccak256(
            abi.encodePacked(type(BeaconProxy).creationCode, abi.encode(address(beacon), bytes("")))
        );

        return Create2.computeAddress(salt_, bytecodeHash_);
    }

    function updateImplementation(address newImplementation_) external onlyOwner {
        UpgradeableBeacon(beacon).upgradeTo(newImplementation_);
    }

    function version() external pure returns (uint256) {
        return 1;
    }

    function _calculatePoolSalt(address sender_) internal view returns (bytes32) {
        return keccak256(abi.encodePacked(sender_, proxies[sender_].length));
    }

    function _authorizeUpgrade(address) internal view override onlyOwner {}
}
