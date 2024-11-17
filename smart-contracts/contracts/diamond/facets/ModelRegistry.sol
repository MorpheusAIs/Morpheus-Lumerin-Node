// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

import {BidStorage} from "../storages/BidStorage.sol";
import {ModelStorage} from "../storages/ModelStorage.sol";
import {DelegationStorage} from "../storages/DelegationStorage.sol";

import {IModelRegistry} from "../../interfaces/facets/IModelRegistry.sol";

contract ModelRegistry is IModelRegistry, OwnableDiamondStorage, ModelStorage, BidStorage, DelegationStorage {
    using EnumerableSet for EnumerableSet.Bytes32Set;
    using SafeERC20 for IERC20;

    function __ModelRegistry_init() external initializer(MODELS_STORAGE_SLOT) {}

    function modelSetMinStake(uint256 modelMinimumStake_) external onlyOwner {
        ModelsStorage storage modelsStorage = _getModelsStorage();
        modelsStorage.modelMinimumStake = modelMinimumStake_;

        emit ModelMinimumStakeUpdated(modelMinimumStake_);
    }

    function modelRegister(
        address modelOwner_,
        bytes32 baseModelId_,
        bytes32 ipfsCID_,
        uint256 fee_,
        uint256 amount_,
        string calldata name_,
        string[] memory tags_
    ) external {
        _validateDelegatee(_msgSender(), modelOwner_, DELEGATION_RULES_MODEL);

        ModelsStorage storage modelsStorage = _getModelsStorage();

        bytes32 modelId_ = getModelId(modelOwner_, baseModelId_);
        Model storage model = modelsStorage.models[modelId_];

        uint256 newStake_ = model.stake + amount_;
        uint256 minStake_ = modelsStorage.modelMinimumStake;
        if (newStake_ < minStake_) {
            revert ModelStakeTooLow(newStake_, minStake_);
        }

        if (amount_ > 0) {
            BidsStorage storage bidsStorage = _getBidsStorage();
            IERC20(bidsStorage.token).safeTransferFrom(modelOwner_, address(this), amount_);
        }

        if (model.createdAt == 0) {
            modelsStorage.modelIds.add(modelId_);

            model.createdAt = uint128(block.timestamp);
            model.owner = modelOwner_;
        }

        model.stake = newStake_;
        model.ipfsCID = ipfsCID_;
        model.fee = fee_; // TODO: validate fee and get usage places
        model.name = name_;
        model.tags = tags_;
        model.isDeleted = false;

        modelsStorage.activeModels.add(modelId_);

        emit ModelRegisteredUpdated(modelOwner_, modelId_);
    }

    function modelDeregister(bytes32 modelId_) external {
        ModelsStorage storage modelsStorage = _getModelsStorage();
        Model storage model = modelsStorage.models[modelId_];

        _validateDelegatee(_msgSender(), model.owner, DELEGATION_RULES_MODEL);        

        if (!_isModelActiveBidsEmpty(modelId_)) {
            revert ModelHasActiveBids();
        }
        if (model.isDeleted) {
            revert ModelHasAlreadyDeregistered();
        }

        uint256 withdrawAmount_ = model.stake;

        model.stake = 0;
        model.isDeleted = true;

        modelsStorage.activeModels.remove(modelId_);

        BidsStorage storage bidsStorage = _getBidsStorage();
        IERC20(bidsStorage.token).safeTransfer(model.owner, withdrawAmount_);

        emit ModelDeregistered(model.owner, modelId_);
    }

    function getModelId(address account_, bytes32 baseModelId_) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(account_, baseModelId_));
    }
}
