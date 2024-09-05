//     // pool 0 - capital tranche
//     s.pools.push(
//       Pool({
//         payoutStart: 1707393600,
//         decreaseInterval: 86400,
//         initialReward: 3456000000000000000000,
//         rewardDecrease: 592558728240000000
//       })
//     );

import { ISessionStorage } from "../../generated-types/ethers/contracts/interfaces/facets/ISessionRouter";

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

export function getDefaultPools(): ISessionStorage.PoolStruct[] {
  return [
    {
      payoutStart: 1707393600n,
      decreaseInterval: 86400n,
      initialReward: 3456000000000000000000n,
      rewardDecrease: 592558728240000000n,
    },
    {
      payoutStart: 1707393600n,
      decreaseInterval: 86400n,
      initialReward: 3456000000000000000000n,
      rewardDecrease: 592558728240000000n,
    },
    {
      payoutStart: 1707393600n,
      decreaseInterval: 86400n,
      initialReward: 3456000000000000000000n,
      rewardDecrease: 592558728240000000n,
    },
    {
      payoutStart: 1707393600n,
      decreaseInterval: 86400n,
      initialReward: 3456000000000000000000n,
      rewardDecrease: 592558728240000000n,
    },
    {
      payoutStart: 1707393600n,
      decreaseInterval: 86400n,
      initialReward: 576000000000000000000n,
      rewardDecrease: 98759788040000000n,
    },
  ];
}
