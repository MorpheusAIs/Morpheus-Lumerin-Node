// // SPDX-License-Identifier: MIT
// pragma solidity ^0.8.24;

// //******************************************************************************\
// //* Author: Nick Mudge <nick@perfectabstractions.com> (https://twitter.com/mudgen)
// //* EIP-2535 Diamonds: https://eips.ethereum.org/EIPS/eip-2535
// //*
// //* Implementation of a diamond.
// /******************************************************************************/

// import { LibDiamond } from "./libraries/LibDiamond.sol";
// import { IDiamondLoupe } from "./interfaces/IDiamondLoupe.sol";
// import { IDiamondCut } from "./interfaces/IDiamondCut.sol";
// import { IERC173 } from "./interfaces/IERC173.sol";
// import { IERC165 } from "./interfaces/IERC165.sol";
// import { LibAppStorage, AppStorage, Session, Pool } from "./AppStorage.sol";
// import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

// // Diamond contract resources
// // https://github.com/mudgen/awesome-diamonds/blob/main/README.md

// // It is expected that this contract is customized if you want to deploy your diamond
// // with data from a deployment script. Use the init function to initialize state variables
// // of your diamond. Add parameters to the init funciton if you need to.

// // Adding parameters to the `init` or other functions you add here can make a single deployed
// // DiamondInit contract reusable accross upgrades, and can be used for multiple diamonds.

// contract DiamondInit {
//   // You can add parameters to this function in order to pass in
//   // data to set your own state variables
//   function init(address _token, address _fundingAccount) external {
//     // adding ERC165 data
//     LibDiamond.DiamondStorage storage ds = LibDiamond.diamondStorage();
//     ds.supportedInterfaces[type(IERC165).interfaceId] = true;
//     ds.supportedInterfaces[type(IDiamondCut).interfaceId] = true;
//     ds.supportedInterfaces[type(IDiamondLoupe).interfaceId] = true;
//     ds.supportedInterfaces[type(IERC173).interfaceId] = true;

//     // add your own state variables
//     AppStorage storage s = LibAppStorage.appStorage();
//     s.token = IERC20(_token);

//     s.fundingAccount = _fundingAccount;

//     // we need to add a dummy session to avoid index 0
//     s.sessions.push(
//       Session({
//         id: bytes32(0),
//         user: address(0),
//         provider: address(0),
//         modelAgentId: bytes32(0),
//         bidID: bytes32(0),
//         stake: 0,
//         pricePerSecond: 0,
//         closeoutReceipt: "",
//         closeoutType: 0,
//         providerWithdrawnAmount: 0,
//         openedAt: 0,
//         endsAt: 0,
//         closedAt: 0
//       })
//     );

//     // default values for pool
//     // watch also (SessionRouter.sol).setPoolConfig

//     // pool 0 - capital tranche
//     s.pools.push(
//       Pool({
//         payoutStart: 1707393600,
//         decreaseInterval: 86400,
//         initialReward: 3456000000000000000000,
//         rewardDecrease: 592558728240000000
//       })
//     );

//     // pool 1 - code tranche
//     s.pools.push(
//       Pool({
//         payoutStart: 1707393600,
//         decreaseInterval: 86400,
//         initialReward: 3456000000000000000000,
//         rewardDecrease: 592558728240000000
//       })
//     );

//     // pool 2 - community tranche
//     s.pools.push(
//       Pool({
//         payoutStart: 1707393600,
//         decreaseInterval: 86400,
//         initialReward: 3456000000000000000000,
//         rewardDecrease: 592558728240000000
//       })
//     );

//     // pool 3 - compute tranche
//     s.pools.push(
//       Pool({
//         payoutStart: 1707393600,
//         decreaseInterval: 86400,
//         initialReward: 3456000000000000000000,
//         rewardDecrease: 592558728240000000
//       })
//     );

//     // pool 4 - protection tranche
//     s.pools.push(
//       Pool({
//         payoutStart: 1707393600,
//         decreaseInterval: 86400,
//         initialReward: 576000000000000000000,
//         rewardDecrease: 98759788040000000
//       })
//     );

//     // EIP-2535 specifies that the `diamondCut` function takes two optional
//     // arguments: address _init and bytes calldata _calldata
//     // These arguments are used to execute an arbitrary function using delegatecall
//     // in order to set state variables in the diamond during deployment or an upgrade
//     // More info here: https://eips.ethereum.org/EIPS/eip-2535#diamond-interface
//   }
// }
