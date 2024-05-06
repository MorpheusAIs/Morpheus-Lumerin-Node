//SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { LibDiamond } from "../diamond/libraries/LibDiamond.sol";

library LibOwner {
  error NotSenderOrOwner();

  modifier onlyOwner() {
    LibDiamond.enforceIsContractOwner();
    _;
  }

  modifier senderOrOwner(address addr) {
    _senderOrOwner(addr);
    _;
  }

  function _senderOrOwner(address resourceOwner) internal view {
    address owner = LibDiamond.diamondStorage().contractOwner;
    if (msg.sender != resourceOwner && msg.sender != owner) {
      revert NotSenderOrOwner();
    }
  }

  function _onlyOwner() internal view {
    LibDiamond.enforceIsContractOwner();
  }
}
