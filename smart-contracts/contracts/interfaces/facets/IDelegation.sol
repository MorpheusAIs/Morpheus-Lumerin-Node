// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IDelegationStorage} from "../storage/IDelegationStorage.sol";

interface IDelegation is IDelegationStorage {
    event DelegationRegistryUpdated(address registry);

    /**
     * The function to initialize the facet.
     * @param registry_ https://docs.delegate.xyz/technical-documentation/delegate-registry/contract-addresses
     */
    function __Delegation_init(address registry_) external;

    /**
     * The function to set the registry.
     * @param registry_ https://docs.delegate.xyz/technical-documentation/delegate-registry/contract-addresses
     */
    function setRegistry(address registry_) external;
}
