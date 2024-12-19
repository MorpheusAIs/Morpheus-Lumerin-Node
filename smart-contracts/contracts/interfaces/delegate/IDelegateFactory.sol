// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IDelegateFactory {
    error InvalidDeregistrationOpenAt(uint128 value, uint128 minimal);

    /**
     * The event that is emitted when the proxy deployed.
     * @param proxyAddress The pool's id.
     */
    event ProxyDeployed(address indexed proxyAddress);

    /**
     * The event that is emitted when the `minDeregistrationTimeout` changed.
     * @param minDeregistrationTimeout_ The pool's id.
     */
    event MinDeregistrationTimeoutUpdated(uint128 minDeregistrationTimeout_);

    /**
     * The function to initialize the contract.
     * @param lumerinDiamond_ The Lumerin protocol address.
     * @param implementation_ The implementation address.
     * @param minDeregistrationTimeout_ The minimal timestamp before deregistration will start
     */
    function DelegateFactory_init(address lumerinDiamond_, address implementation_, uint128 minDeregistrationTimeout_) external;

    /**
     * Triggers stopped state.
     */
    function pause() external;

    /**
     * Returns to normal state.
     */
    function unpause() external;

    /**
     * The function to deploy the new proxy contract.
     * @param feeTreasury_ The subnet fee treasury.
     * @param fee_ The fee percent where 100% = 10^25.
     * @param name_ The Subnet name.
     * @param endpoint_ The subnet endpoint.
     * @param deregistrationOpensAt_ Provider deregistration will be available after this timestamp.
     * @return Deployed proxy address
     */
    function deployProxy(
        address feeTreasury_,
        uint256 fee_,
        string memory name_,
        string memory endpoint_,
        uint128 deregistrationOpensAt_
    ) external returns (address);

    /**
     * The function to predict new proxy address.
     * @param deployer_ The deployer address.
     */
    function predictProxyAddress(address deployer_) external view returns (address);

    /**
     * The function to upgrade the implementation.
     * @param newImplementation_ The new implementation address.
     */
    function updateImplementation(address newImplementation_) external;

    /**
     * @return The contract version.
     */
    function version() external pure returns (uint256);
}
