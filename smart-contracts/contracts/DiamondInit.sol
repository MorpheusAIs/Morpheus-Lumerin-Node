// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

//******************************************************************************\
//* Author: Nick Mudge <nick@perfectabstractions.com> (https://twitter.com/mudgen)
//* EIP-2535 Diamonds: https://eips.ethereum.org/EIPS/eip-2535
//*
//* Implementation of a diamond.
/******************************************************************************/

import { LibDiamond } from "./diamond/libraries/LibDiamond.sol";
import { IDiamondLoupe } from "./diamond/interfaces/IDiamondLoupe.sol";
import { IDiamondCut } from "./diamond/interfaces/IDiamondCut.sol";
import { IERC173 } from "./diamond/interfaces/IERC173.sol";
import { IERC165 } from "./diamond/interfaces/IERC165.sol";
import { LibAppStorage, AppStorage, Session } from "./AppStorage.sol";
import { IERC20 } from '@openzeppelin/contracts/token/ERC20/IERC20.sol';

// It is expected that this contract is customized if you want to deploy your diamond
// with data from a deployment script. Use the init function to initialize state variables
// of your diamond. Add parameters to the init funciton if you need to.

// Adding parameters to the `init` or other functions you add here can make a single deployed
// DiamondInit contract reusable accross upgrades, and can be used for multiple diamonds.

contract DiamondInit {    

    // You can add parameters to this function in order to pass in 
    // data to set your own state variables
    function init(address _token, address _tokenAccount) external {
        // adding ERC165 data
        LibDiamond.DiamondStorage storage ds = LibDiamond.diamondStorage();
        ds.supportedInterfaces[type(IERC165).interfaceId] = true;
        ds.supportedInterfaces[type(IDiamondCut).interfaceId] = true;
        ds.supportedInterfaces[type(IDiamondLoupe).interfaceId] = true;
        ds.supportedInterfaces[type(IERC173).interfaceId] = true;

        // add your own state variables 
        AppStorage storage s = LibAppStorage.appStorage();
        s.token = IERC20(_token);

        s.stakeDelay = 0;
        s.tokenAccount = _tokenAccount;
        s.sessions.push(Session({
            id: bytes32(0),
            user: address(0),
            provider: address(0),
            modelAgentId: bytes32(0),
            bidID: bytes32(0),
            stake: 0,
            pricePerSecond: 0,
            closeoutReceipt: "",
            closeoutType: 0,
            openedAt: 0,
            closedAt: 0
        }));
        // EIP-2535 specifies that the `diamondCut` function takes two optional 
        // arguments: address _init and bytes calldata _calldata
        // These arguments are used to execute an arbitrary function using delegatecall
        // in order to set state variables in the diamond during deployment or an upgrade
        // More info here: https://eips.ethereum.org/EIPS/eip-2535#diamond-interface 
    }
}