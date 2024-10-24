// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IModelStorage {
    /**
     * The structure that stores the model data.
     * @param ipfsCID https://docs.ipfs.tech/concepts/content-addressing/#what-is-a-cid. Up to the model maintainer to keep up to date.
     * @param fee The model fee. Readonly for now.
     * @param stake The stake amount.
     * @param owner The owner.
     * @param name The name. Readonly for now.
     * @param tags The tags. Readonly for now.
     * @param createdAt The timestamp when the model is created.
     * @param isDeleted The model status.
     */
    struct Model {
        bytes32 ipfsCID;
        uint256 fee; // The fee is a royalty placeholder that isn't currently used
        uint256 stake;
        address owner;
        string name; // TODO: Limit name length. Up to the model maintainer to keep up to date
        string[] tags; // TODO: Limit tags amount. Up to the model maintainer to keep up to date
        uint128 createdAt;
        bool isDeleted;
    }

    /**
     * The function returns the model structure.
     * @param modelId_ Model ID.
     */
    function getModel(bytes32 modelId_) external view returns (Model memory);

    /**
     * The function returns the model IDs.
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getModelIds(uint256 offset_, uint256 limit_) external view returns (bytes32[] memory);

    /**
     * The function returns the model minimal stake.
     */
    function getModelMinimumStake() external view returns (uint256);

    /**
     * The function returns active model IDs.
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getActiveModelIds(uint256 offset_, uint256 limit_) external view returns (bytes32[] memory);

    /**
     * The function returns the model status, active or not.
     * @param modelId_ Model ID.
     */
    function getIsModelActive(bytes32 modelId_) external view returns (bool);
}
