// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IDelegationStorage {
    error InsufficientRightsForOperation(address delegator, address delegatee);

    /**
     * @return `keccak256("delegation.rules.provider")`
     */
    function DELEGATION_RULES_PROVIDER() external view returns (bytes32);
    
    /**
     * @return `keccak256("delegation.rules.model")`
     */
    function DELEGATION_RULES_MODEL() external view returns (bytes32);
    
    /**
     * @return `keccak256("delegation.rules.marketplace")`
     */
    function DELEGATION_RULES_MARKETPLACE() external view returns (bytes32);
    
    /**
     * @return `keccak256("delegation.rules.session")`
     */
    function DELEGATION_RULES_SESSION() external view returns (bytes32);
    
    /**
     * @return The registry address.
     */
    function getRegistry() external view returns (address);
    
    /**
     * The function to check is `delegator_` add permissions for `delegatee_`.
     * @return `true` when delegated.
     */
    function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) external view returns (bool);
}
