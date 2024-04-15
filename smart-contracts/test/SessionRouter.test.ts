// describe.skip("Session actions", function () {
//   it("Should create session", async function () {
//     const { providerRegistry, provider, buyer, tokenMOR, expected, publicClient, buyerBalance } =
//       await loadFixture(deploySingleProvider);

//     const approveHash = await tokenMOR.write.approve([providerRegistry.address, buyerBalance], {
//       account: buyer.account,
//     });
//     await publicClient.waitForTransactionReceipt({ hash: approveHash });

//     const { request, result: sessionId } = await providerRegistry.simulate.startSession(
//       [provider.account.address],
//       { account: buyer.account.address }
//     );
//     const startSessionHash = await providerRegistry.write.startSession(
//       [provider.account.address],
//       request
//     );
//     const startSessionTx = await publicClient.waitForTransactionReceipt({
//       hash: startSessionHash,
//     });

//     const { timestamp } = await publicClient.getBlock({
//       blockNumber: startSessionTx.blockNumber,
//     });

//     // check balances
//     expect(await tokenMOR.read.balanceOf([buyer.account.address])).eq(0n);
//     expect(await tokenMOR.read.balanceOf([providerRegistry.address])).eq(buyerBalance);

//     // check session
//     expect(sessionId).eq(0n);

//     const expDuration = (buyerBalance / expected.price) * 60n;
//     const session = await providerRegistry.read.sessions([sessionId]);
//     console.log("session", session);
//     console.log("exp duration", expDuration);

//     expect(session).to.be.deep.eq([
//       sessionId,
//       getAddress(buyer.account.address),
//       getAddress(provider.account.address),
//       timestamp,
//       timestamp + (buyerBalance / expected.price) * 60n,
//     ]);
//   });

//   it("should close session manually and refund", async function () {
//     const { providerRegistry, provider, buyer, tokenMOR, expected, publicClient, buyerBalance } =
//       await loadFixture(deploySingleProvider);

//     const approveHash = await tokenMOR.write.approve([providerRegistry.address, buyerBalance], {
//       account: buyer.account,
//     });
//     await publicClient.waitForTransactionReceipt({ hash: approveHash });

//     const { request, result: sessionId } = await providerRegistry.simulate.startSession(
//       [provider.account.address],
//       { account: buyer.account.address }
//     );
//     const startSessionHash = await providerRegistry.write.startSession(
//       [provider.account.address],
//       request
//     );
//     const startSessionTx = await publicClient.waitForTransactionReceipt({
//       hash: startSessionHash,
//     });

//     const { timestamp } = await publicClient.getBlock({
//       blockNumber: startSessionTx.blockNumber,
//     });
//     const progress = 1 / 3;
//     const denominator = 10 ** 3;
//     const expectedSpentBalance =
//       (BigInt(Math.round(progress * denominator)) * buyerBalance) / BigInt(denominator);
//     const expectedSpentTimeSeconds = (60n * expectedSpentBalance) / expected.price;
//     const expectedCloseTimestamp = timestamp + expectedSpentTimeSeconds;

//     console.log("exp spent time", expectedSpentTimeSeconds);
//     console.log("exp close", expectedCloseTimestamp);
//     console.log("total balance", buyerBalance);
//     console.log("exp spent balance", expectedSpentBalance);
//     console.log("old buyer balance", await tokenMOR.read.balanceOf([buyer.account.address]));
//     console.log("old contract balance", await tokenMOR.read.balanceOf([providerRegistry.address]));

//     await time.increaseTo(expectedCloseTimestamp);

//     const closeSessionHash = await providerRegistry.write.closeSession([sessionId], {
//       account: buyer.account,
//     });
//     await publicClient.waitForTransactionReceipt({ hash: closeSessionHash });
//     const refundValue = buyerBalance - expectedSpentBalance;

//     const newBuyerBalance = await tokenMOR.read.balanceOf([buyer.account.address]);
//     const newRouterBalance = await tokenMOR.read.balanceOf([providerRegistry.address]);

//     expectAlmostEqual(refundValue, newBuyerBalance, 0.01);
//     expectAlmostEqual(refundValue, newRouterBalance, 0.01);
//   });
// });
