// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import { InitializableStorage } from "@solarity/solidity-lib/diamond/utils/InitializableStorage.sol";

import { Context } from "@openzeppelin/contracts/utils/Context.sol";

/**
 * @notice The Diamond standard module
 *
 * This is an Ownable Storage contract with Diamond Standard support
 */
abstract contract DiamondOwnableStorage is InitializableStorage, Context {
  bytes32 public constant DIAMOND_OWNABLE_STORAGE_SLOT = keccak256("diamond.standard.diamond.ownable.storage");

  struct DOStorage {
    address owner;
  }

  modifier onlyOwner() {
    _onlyOwner();
    _;
  }

  function _getDiamondOwnableStorage() internal pure returns (DOStorage storage _dos) {
    bytes32 slot_ = DIAMOND_OWNABLE_STORAGE_SLOT;

    assembly {
      _dos.slot := slot_
    }
  }

  /**
   * @notice The function to get the Diamond owner
   * @return the owner of the Diamond
   */
  function owner() public view virtual returns (address) {
    return _getDiamondOwnableStorage().owner;
  }

  /**
   * @notice The function to check if `_msgSender` is the owner
   */
  function _onlyOwner() internal view virtual {
    require(owner() == _msgSender(), "DiamondOwnable: not an owner");
  }
}
