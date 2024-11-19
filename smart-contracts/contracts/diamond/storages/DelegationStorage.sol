// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {Paginator} from "@solarity/solidity-lib/libs/arrays/Paginator.sol";

import {IDelegationStorage} from "../../interfaces/storage/IDelegationStorage.sol";
import {IDelegateRegistry} from "../../interfaces/deps/IDelegateRegistry.sol";

contract DelegationStorage is IDelegationStorage {
    struct DLGTNStorage {
        address registry;
    }

    bytes32 public constant DELEGATION_STORAGE_SLOT = keccak256("diamond.standard.delegation.storage");
    bytes32 public constant DELEGATION_RULES_PROVIDER = keccak256("delegation.rules.provider");
    bytes32 public constant DELEGATION_RULES_MODEL = keccak256("delegation.rules.model");
    bytes32 public constant DELEGATION_RULES_MARKETPLACE = keccak256("delegation.rules.marketplace");
    bytes32 public constant DELEGATION_RULES_SESSION = keccak256("delegation.rules.session");

    /** PUBLIC, GETTERS */
    function getRegistry() external view returns (address) {
        return _getDelegationStorage().registry;
    }

    function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) public view returns (bool) {
        DLGTNStorage storage delegationStorage = _getDelegationStorage();

        return
            IDelegateRegistry(delegationStorage.registry).checkDelegateForContract(
                delegatee_,
                delegator_,
                address(this),
                rights_
            );
    }

    /** INTERNAL */
    function _validateDelegatee(address delegatee_, address delegator_, bytes32 rights_) internal view {
        if (delegatee_ != delegator_ && !isRightsDelegated(delegatee_, delegator_, rights_)) {
            revert InsufficientRightsForOperation(delegator_, delegatee_);
        }
    }

    function _getDelegationStorage() internal pure returns (DLGTNStorage storage ds) {
        bytes32 slot_ = DELEGATION_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }
}
