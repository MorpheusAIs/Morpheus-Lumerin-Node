// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {Paginator} from "@solarity/solidity-lib/libs/arrays/Paginator.sol";

import {SessionStorage} from "../storages/SessionStorage.sol";
import {BidStorage} from "../storages/BidStorage.sol";
import {ModelStorage} from "../storages/ModelStorage.sol";
import {ProviderStorage} from "../storages/ProviderStorage.sol";

import {IQueryFacet} from "../../interfaces/facets/IQueryFacet.sol";

/**
 * @title QueryFacet
 * @notice Consolidated read-only aggregation facet for developer and client SDK efficiency (RFP H5).
 */
contract QueryFacet is
    IQueryFacet,
    SessionStorage,
    BidStorage,
    ModelStorage,
    ProviderStorage
{
    using Paginator for *;
    using EnumerableSet for EnumerableSet.Bytes32Set;

    /**
     * @notice Consolidates session, bid, model, and provider data into a single RPC query.
     */
    function getSessionView(bytes32 sessionId_) external view returns (SessionView memory view_) {
        Session memory session = _getSessionsStorage().sessions[sessionId_];
        Bid memory bid = _getBidsStorage().bids[session.bidId];
        Model memory model = _getModelsStorage().models[bid.modelId];
        Provider memory provider = _getProvidersStorage().providers[bid.provider];

        return SessionView({
            session: session,
            bid: bid,
            model: model,
            provider: provider
        });
    }

    /**
     * @notice Consolidates bid and provider data into a single RPC query.
     */
    function getBidView(bytes32 bidId_) external view returns (ModelBidView memory view_) {
        Bid memory bid = _getBidsStorage().bids[bidId_];
        Provider memory provider = _getProvidersStorage().providers[bid.provider];

        return ModelBidView({
            bid: bid,
            provider: provider
        });
    }

    /**
     * @notice Returns paginated active bids for a model enriched with provider metadata.
     */
    function getModelBidsEnriched(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (ModelBidView[] memory views_, uint256 total_) {
        EnumerableSet.Bytes32Set storage modelActiveBids = _getBidsStorage().modelActiveBids[modelId_];
        total_ = modelActiveBids.length();
        bytes32[] memory bidIds = modelActiveBids.part(offset_, limit_);

        views_ = new ModelBidView[](bidIds.length);
        for (uint256 i = 0; i < bidIds.length; i++) {
            Bid memory bid = _getBidsStorage().bids[bidIds[i]];
            Provider memory provider = _getProvidersStorage().providers[bid.provider];
            views_[i] = ModelBidView({
                bid: bid,
                provider: provider
            });
        }
    }
}
