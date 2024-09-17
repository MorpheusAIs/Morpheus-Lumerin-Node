// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IModelStorage} from "../storage/IModelStorage.sol";

interface IModelRegistry is IModelStorage {
    event ModelRegisteredUpdated(address indexed owner, bytes32 indexed modelId);
    event ModelDeregistered(address indexed owner, bytes32 indexed modelId);
    event ModelMinStakeUpdated(uint256 newStake);

    error ModelNotFound();
    error StakeTooLow();
    error ModelHasActiveBids();
    error NotOwnerOrModelOwner();

    function __ModelRegistry_init() external;

    function modelSetMinStake(uint256 modelMinimumStake_) external;

    function modelRegister(
        bytes32 modelId_,
        bytes32 ipfsCID_,
        uint256 fee_,
        uint256 stake_,
        address owner_,
        string memory name_,
        string[] memory tags_
    ) external;

    function modelDeregister(bytes32 modelId_) external;
}
