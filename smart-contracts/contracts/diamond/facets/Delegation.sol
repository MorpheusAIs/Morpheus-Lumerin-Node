// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

import {DelegationStorage} from "../storages/DelegationStorage.sol";

import {IDelegation} from "../../interfaces/facets/IDelegation.sol";

contract Delegation is IDelegation, OwnableDiamondStorage, DelegationStorage {
    function __Delegation_init(address registry_) external initializer(DELEGATION_STORAGE_SLOT) {
        setRegistry(registry_);
    }

    function setRegistry(address registry_) public onlyOwner {
        DLGTNStorage storage delegationStorage = _getDelegationStorage();
        delegationStorage.registry = registry_;

        emit DelegationRegistryUpdated(registry_);
    }
}
