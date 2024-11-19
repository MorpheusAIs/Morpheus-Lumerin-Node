// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IMarketplaceStorage {
    /**
     * The function returns bid fee on creation.
     */
    function getBidFee() external view returns (uint256);

    /**
     * The function returns fee balance.
     */
    function getFeeBalance() external view returns (uint256);

    /**
     * The function returns min and max price per second for bid.
     * @return Min bid price per second
     * @return Max bid price per second
     */
    function getMinMaxBidPricePerSecond() external view returns (uint256, uint256);
}
