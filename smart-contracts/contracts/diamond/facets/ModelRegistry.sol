// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

import {BidStorage} from "../storages/BidStorage.sol";
import {ModelStorage} from "../storages/ModelStorage.sol";

import {IModelRegistry} from "../../interfaces/facets/IModelRegistry.sol";

contract ModelRegistry is IModelRegistry, OwnableDiamondStorage, ModelStorage, BidStorage {
    using SafeERC20 for IERC20;

    function __ModelRegistry_init() external initializer(MODEL_STORAGE_SLOT) {}

    function setModelMinimumStake(uint256 modelMinimumStake_) external onlyOwner {
        _setModelMinimumStake(modelMinimumStake_);
        emit ModelMinimumStakeSet(modelMinimumStake_);
    }

    /// @notice Registers or updates existing model
    function modelRegister(
        // TODO: it is not secure (frontrunning) to take the modelId as key
        bytes32 modelId_,
        bytes32 ipfsCID_,
        uint256 fee_,
        uint256 addStake_,
        address owner_,
        string calldata name_,
        string[] calldata tags_
    ) external {
        if (!_isOwnerOrModelOwner(owner_)) {
            // TODO: such that we cannon create a model with the owner as another address
            // Do we need this check?
            revert NotOwnerOrModelOwner();
        }

        Model memory model_ = models(modelId_);
        // TODO: there is no way to decrease the stake
        uint256 newStake_ = model_.stake + addStake_;
        if (newStake_ < modelMinimumStake()) {
            revert StakeTooLow();
        }

        if (addStake_ > 0) {
            getToken().safeTransferFrom(_msgSender(), address(this), addStake_);
        }

        uint128 createdAt_ = model_.createdAt;
        if (createdAt_ == 0) {
            // model never existed
            addModel(modelId_);
            setModelActive(modelId_, true);
            createdAt_ = uint128(block.timestamp);
        } else {
            if (!_isOwnerOrModelOwner(model_.owner)) {
                revert NotOwnerOrModelOwner();
            }
            if (model_.isDeleted) {
                setModelActive(modelId_, true);
            }
        }

        setModel(modelId_, Model(ipfsCID_, fee_, newStake_, owner_, name_, tags_, createdAt_, false));

        emit ModelRegisteredUpdated(owner_, modelId_);
    }

    function modelDeregister(bytes32 modelId_) external {
        Model storage model = models(modelId_);

        if (!isModelExists(modelId_)) {
            revert ModelNotFound();
        }
        if (!_isOwnerOrModelOwner(model.owner)) {
            revert NotOwnerOrModelOwner();
        }
        if (!isModelActiveBidsEmpty(modelId_)) {
            revert ModelHasActiveBids();
        }

        uint256 stake_ = model.stake;

        model.stake = 0;
        model.isDeleted = true;

        setModelActive(modelId_, false);

        getToken().safeTransfer(model.owner, stake_);

        emit ModelDeregistered(model.owner, modelId_);
    }

    function isModelExists(bytes32 modelId_) public view returns (bool) {
        return models(modelId_).createdAt != 0;
    }

    function _isOwnerOrModelOwner(address modelOwner_) internal view returns (bool) {
        return _msgSender() == owner() || _msgSender() == modelOwner_;
    }
}
