// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {DiamondOwnableStorage} from "@solarity/solidity-lib/diamond/access/ownable/DiamondOwnableStorage.sol";

import {Context} from "@openzeppelin/contracts/utils/Context.sol";

abstract contract OwnableDiamondStorage is DiamondOwnableStorage, Context {
    function _onlyOwner() internal view virtual override {
        require(owner() == _msgSender(), "OwnableDiamondStorage: not an owner");
    }
}
