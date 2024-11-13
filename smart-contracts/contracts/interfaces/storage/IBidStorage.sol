// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IBidStorage {
    /**
     * The structure that stores the bid data.
     * @param provider The provider addres.
     * @param modelId The model ID.
     * @param pricePerSecond The price per second.
     * @param nonce The bid creates with this nounce (related to provider nounce).
     * @param createdAt The timestamp when the bid is created.
     * @param deletedAt The timestamp when the bid is deleted.
     */
    struct Bid {
        address provider;
        bytes32 modelId;
        uint256 pricePerSecond;
        uint256 nonce;
        uint128 createdAt;
        uint128 deletedAt;
    }

    /**
     * The function returns the bid structure.
     * @param bidId_ Bid ID.
     */
    function getBid(bytes32 bidId_) external view returns (Bid memory);

    /**
     * The function returns active provider bids.
     * @param provider_ Provider address.
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getProviderActiveBids(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory, uint256);

    /**
     * The function returns active model bids.
     * @param modelId_ Model ID.
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getModelActiveBids(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory, uint256);

    /**
     * The function returns provider bids.
     * @param provider_ Provider address.
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getProviderBids(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory, uint256);

    /**
     * The function returns model bids.
     * @param modelId_ Model ID.
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getModelBids(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory, uint256);

    /**
     * The function returns stake token (MOR).
     */
    function getToken() external view returns (address);

    /**
     * The function returns bid status, active or not.
     * @param bidId_ Bid ID.
     */
    function isBidActive(bytes32 bidId_) external view returns (bool);
}
