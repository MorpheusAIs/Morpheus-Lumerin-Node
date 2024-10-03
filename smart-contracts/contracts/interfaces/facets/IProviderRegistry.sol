// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IProviderStorage} from "../storage/IProviderStorage.sol";

interface IProviderRegistry is IProviderStorage {
    event ProviderRegisteredUpdated(address indexed provider);
    event ProviderDeregistered(address indexed provider);
    event ProviderMinStakeUpdated(uint256 newStake);
    event ProviderWithdrawnStake(address indexed provider, uint256 amount);
    error StakeTooLow();
    error ErrProviderNotDeleted();
    error ErrNoStake();
    error ErrNoWithdrawableStake();
    error ProviderHasActiveBids();
    error NotOwnerOrProvider();
    error ProviderNotFound();

    function __ProviderRegistry_init() external;

    function providerSetMinStake(uint256 providerMinimumStake_) external;

    function providerRegister(address providerAddress_, uint256 amount_, string memory endpoint_) external;

    function providerDeregister(address provider_) external;

    function providerWithdrawStake(address provider_) external;
}
