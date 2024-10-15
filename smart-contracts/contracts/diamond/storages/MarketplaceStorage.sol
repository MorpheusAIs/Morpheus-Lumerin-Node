// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IMarketplaceStorage} from "../../interfaces/storage/IMarketplaceStorage.sol";

contract MarketplaceStorage is IMarketplaceStorage {
    struct MPStorage {
        uint256 feeBalance; // Total fees balance of the contract
        uint256 bidFee;
    }

    bytes32 public constant MARKETPLACE_STORAGE_SLOT = keccak256("diamond.standard.marketplace.storage");

    /** PUBLIC, GETTERS */
    function getBidFee() public view returns (uint256) {
        return _getMarketplaceStorage().bidFee;
    }

    function getFeeBalance() public view returns (uint256) {
        return _getMarketplaceStorage().feeBalance;
    }

    /** INTERNAL, SETTERS */
    function setBidFee(uint256 bidFee_) internal {
        _getMarketplaceStorage().bidFee = bidFee_;
    }

    function setFeeBalance(uint256 feeBalance_) internal {
        _getMarketplaceStorage().feeBalance = feeBalance_;
    }

    /** PRIVATE */
    function _getMarketplaceStorage() private pure returns (MPStorage storage ds) {
        bytes32 slot_ = MARKETPLACE_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }
}
