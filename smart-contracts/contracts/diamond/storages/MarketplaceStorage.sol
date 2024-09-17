// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IMarketplaceStorage} from "../../interfaces/storage/IMarketplaceStorage.sol";

contract MarketplaceStorage is IMarketplaceStorage {
    struct MPStorage {
        uint256 feeBalance; // total fees balance of the contract
        uint256 bidFee;
    }

    bytes32 public constant MARKETPLACE_STORAGE_SLOT = keccak256("diamond.standard.marketplace.storage");

    function getBidFee() public view returns (uint256) {
        return _getMarketplaceStorage().bidFee;
    }

    function getFeeBalance() internal view returns (uint256) {
        return _getMarketplaceStorage().feeBalance;
    }

    function increaseFeeBalance(uint256 amount) internal {
        _getMarketplaceStorage().feeBalance += amount;
    }

    function decreaseFeeBalance(uint256 amount) internal {
        _getMarketplaceStorage().feeBalance -= amount;
    }

    function _getMarketplaceStorage() internal pure returns (MPStorage storage _ds) {
        bytes32 slot_ = MARKETPLACE_STORAGE_SLOT;

        assembly {
            _ds.slot := slot_
        }
    }
}
