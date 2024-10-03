// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IModelStorage {
    struct Model {
        bytes32 ipfsCID; // https://docs.ipfs.tech/concepts/content-addressing/#what-is-a-cid
        uint256 fee;
        uint256 stake;
        address owner;
        string name; // limit name length
        string[] tags; // TODO: limit tags amount
        uint128 createdAt;
        bool isDeleted;
    }

    function getModel(bytes32 modelId) external view returns (Model memory);

    function models(uint256 index) external view returns (bytes32);

    function modelMinimumStake() external view returns (uint256);
}
