// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IMarketplaceStorage} from "../storage/IMarketplaceStorage.sol";

interface IMarketplace is IMarketplaceStorage {
    event MaretplaceFeeUpdated(uint256 bidFee);
    event MarketplaceBidPosted(address indexed provider, bytes32 indexed modelId, uint256 nonce);
    event MarketplaceBidDeleted(address indexed provider, bytes32 indexed modelId, uint256 nonce);
    event MarketplaceBidMinMaxPriceUpdated(uint256 bidMinPricePerSecond, uint256 bidMaxPricePerSecond);

    error MarketplaceProviderNotFound();
    error MarketplaceModelNotFound();
    error MarketplaceActiveBidNotFound();
    error MarketplaceBidMinPricePerSecondIsZero();
    error MarketplaceBidMinPricePerSecondIsInvalid();
    error MarketplaceBidPricePerSecondInvalid();

    /**
     * The function to initialize the facet.
     * @param token_ Stake token (MOR)
     * @param bidMinPricePerSecond_ Min price per second for bid
     * @param bidMaxPricePerSecond_ Max price per second for bid
     */
    function __Marketplace_init(address token_, uint256 bidMinPricePerSecond_, uint256 bidMaxPricePerSecond_) external;

    /**
     * The function to set the bidFee.
     * @param bidFee_ Amount of tokens
     */
    function setMarketplaceBidFee(uint256 bidFee_) external;

    /**
     * The function to set the min and max price per second for bid.
     * @param bidMinPricePerSecond_ Min price per second for bid
     * @param bidMaxPricePerSecond_ Max price per second for bid
     */
    function setMinMaxBidPricePerSecond(
        uint256 bidMinPricePerSecond_,
        uint256 bidMaxPricePerSecond_
    ) external;

    /**
     * The function to create the bid.
     * @param modelId_ The mode ID
     * @param pricePerSecond_ The price per second
     */
    function postModelBid(bytes32 modelId_, uint256 pricePerSecond_) external returns (bytes32);

    /**
     * The function to delete the bid.
     * @param bidId_ The bid ID
     */
    function deleteModelBid(bytes32 bidId_) external;

    /**
     * The function to withdraw the stake amount.
     * @param recipient_ The recipient address.
     * @param amount_ The amount.
     */
    function withdraw(address recipient_, uint256 amount_) external;

    /**
     * The function to get bid ID.
     * @param provider_ The provider address.
     * @param modelId_  The model ID.
     * @param nonce_ The nonce.
     */
    function getBidId(address provider_, bytes32 modelId_, uint256 nonce_) external view returns (bytes32);

    /**
     * The function to returns provider model ID
     * @param provider_ The provider address.
     * @param modelId_  The model ID.
     */
    function getProviderModelId(address provider_, bytes32 modelId_) external view returns (bytes32);
}
