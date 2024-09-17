// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IBidStorage} from "../storage/IBidStorage.sol";
import {IMarketplaceStorage} from "../storage/IMarketplaceStorage.sol";

interface IMarketplace is IBidStorage, IMarketplaceStorage {
    event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
    event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce);
    event FeeUpdated(uint256 bidFee);

    error ProviderNotFound();
    error ModelOrAgentNotFound();
    error ActiveBidNotFound();
    error BidTaken();
    error NotEnoughBalance();
    error NotOwnerOrProvider();

    function __Marketplace_init(address token_) external;

    function setBidFee(uint256 bidFee_) external;

    function postModelBid(
        address provider_,
        bytes32 modelId_,
        uint256 pricePerSecond_
    ) external returns (bytes32 bidId);

    function deleteModelAgentBid(bytes32 bidId_) external;

    function withdraw(address recipient_, uint256 amount_) external;

    function getBidId(address provider_, bytes32 modelAgentId_, uint256 nonce_) external view returns (bytes32);

    function getProviderModelAgentId(address provider_, bytes32 modelAgentId_) external view returns (bytes32);
}
