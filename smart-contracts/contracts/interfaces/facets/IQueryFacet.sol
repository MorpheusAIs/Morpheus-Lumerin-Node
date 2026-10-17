// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ISessionStorage} from "../storage/ISessionStorage.sol";
import {IBidStorage} from "../storage/IBidStorage.sol";
import {IModelStorage} from "../storage/IModelStorage.sol";
import {IProviderStorage} from "../storage/IProviderStorage.sol";

/**
 * @title IQueryFacet
 * @notice Consolidated read-only view facet joining multi-facet state across the Inference Diamond (RFP H5).
 */
interface IQueryFacet is ISessionStorage, IBidStorage, IModelStorage, IProviderStorage {
    struct SessionView {
        Session session;
        Bid bid;
        Model model;
        Provider provider;
    }

    struct ModelBidView {
        Bid bid;
        Provider provider;
    }

    /**
     * @notice Returns comprehensive consolidated session view across all facets.
     * @param sessionId_ The unique session ID.
     */
    function getSessionView(bytes32 sessionId_) external view returns (SessionView memory view_);

    /**
     * @notice Returns consolidated bid and provider view.
     * @param bidId_ The unique bid ID.
     */
    function getBidView(bytes32 bidId_) external view returns (ModelBidView memory view_);

    /**
     * @notice Returns paginated active bids for a model enriched with provider metadata.
     * @param modelId_ The model ID.
     * @param offset_ Paginator offset.
     * @param limit_ Paginator limit.
     */
    function getModelBidsEnriched(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (ModelBidView[] memory views_, uint256 total_);
}
