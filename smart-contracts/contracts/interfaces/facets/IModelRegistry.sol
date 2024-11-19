// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IModelStorage} from "../storage/IModelStorage.sol";

interface IModelRegistry is IModelStorage {
    event ModelRegisteredUpdated(address indexed owner, bytes32 indexed modelId);
    event ModelDeregistered(address indexed owner, bytes32 indexed modelId);
    event ModelMinimumStakeUpdated(uint256 modelMinimumStake);
    error ModelStakeTooLow(uint256 amount, uint256 minAmount);
    error ModelHasAlreadyDeregistered();
    error ModelNotFound();
    error ModelHasActiveBids();

    /**
     * The function to initialize the facet.
     */
    function __ModelRegistry_init() external;

    /**
     * The function to set the minimal stake for models.
     * @param modelMinimumStake_ Amount of tokens
     */
    function modelSetMinStake(uint256 modelMinimumStake_) external;

    /**
     * The function to register the model.
     * @param modelOwner_ The model owner address.
     * @param modelId_ The model ID.
     * @param ipfsCID_ The model IPFS CID.
     * @param fee_ The model fee.
     * @param amount_ The model stake amount.
     * @param name_ The model name.
     * @param tags_ The model tags.
     */
    function modelRegister(
        address modelOwner_,
        bytes32 modelId_,
        bytes32 ipfsCID_,
        uint256 fee_,
        uint256 amount_,
        string calldata name_,
        string[] calldata tags_
    ) external;

    /**
     * The function to deregister the model.
     * @param modelId_ The model ID.
     */
    function modelDeregister(bytes32 modelId_) external;

    /**
     * Form model ID for the user models.
     * @param account_ The address.
     * @param baseModelId_ The base model ID.
     */
    function getModelId(address account_, bytes32 baseModelId_) external pure returns (bytes32);
}
