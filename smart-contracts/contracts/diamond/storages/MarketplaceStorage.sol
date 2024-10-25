// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IMarketplaceStorage} from "../../interfaces/storage/IMarketplaceStorage.sol";

contract MarketplaceStorage is IMarketplaceStorage {
    struct MarketStorage {
        uint256 feeBalance; // Total fees balance of the contract
        uint256 bidFee;
        uint256 bidMinPricePerSecond;
        uint256 bidMaxPricePerSecond;
    }

    bytes32 public constant MARKET_STORAGE_SLOT = keccak256("diamond.standard.market.storage");

    /** PUBLIC, GETTERS */
    function getBidFee() external view returns (uint256) {
        return getMarketStorage().bidFee;
    }

    function getFeeBalance() external view returns (uint256) {
        return getMarketStorage().feeBalance;
    }

    function getMinMaxBidPricePerSecond() external view returns (uint256, uint256) {
        return (getMarketStorage().bidMinPricePerSecond, getMarketStorage().bidMaxPricePerSecond);
    }

    /** INTERNAL */
    function getMarketStorage() internal pure returns (MarketStorage storage ds) {
        bytes32 slot_ = MARKET_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }
}
