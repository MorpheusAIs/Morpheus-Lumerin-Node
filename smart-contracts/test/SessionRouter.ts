import {
  time,
  loadFixture,
  takeSnapshot,
  SnapshotRestorer,
} from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { getAddress, parseGwei, parseUnits } from "viem";

describe("Session router", function () {
  let singleProviderSnap: SnapshotRestorer;
  // We define a fixture to reuse the same setup in every test.
  // We use loadFixture to run this setup once, snapshot that state,
  // and reset Hardhat Network to that snapshot in every test.
  async function deployMorpheusTokenAndSessionRouter() {
    // Contracts are deployed using the first signer/account by default
    const [owner, provider, buyer] = await hre.viem.getWalletClients();

    const tokenMOR = await hre.viem.deployContract("MorpheusToken", []);
    const sessionRouter = await hre.viem.deployContract("SessionRouter", [tokenMOR.address]);
    const publicClient = await hre.viem.getPublicClient();
    const decimalsMOR = await tokenMOR.read.decimals();

    return {
      tokenMOR,
      decimalsMOR,
      sessionRouter,
      owner,
      provider,
      buyer,
      publicClient,
    };
  }

  async function deployWithSingleProvider(){
    const {sessionRouter, owner, provider, publicClient, decimalsMOR, tokenMOR, buyer} = await deployMorpheusTokenAndSessionRouter();
    const expected = {
      address: provider.account.address,
      name: "Best provider",
      URL: "https://bestprovider.com",
      price: parseUnits("100", decimalsMOR),
    }

    const buyerBalance = expected.price * 60n;
    
    const addProviderHash = await sessionRouter.write.addProvider([
      expected.address,
      expected.name,
      expected.URL,
      expected.price,
    ]);
    await publicClient.waitForTransactionReceipt({ hash: addProviderHash });
    
    const transferHash = await tokenMOR.write.transfer([buyer.account.address, buyerBalance])
    await publicClient.waitForTransactionReceipt({hash: transferHash})

    return {
      expected,
      sessionRouter,
      owner,
      provider,
      buyer,
      buyerBalance,
      publicClient,
      tokenMOR,
    }
  }

  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      const { sessionRouter, owner } = await loadFixture(
        deployMorpheusTokenAndSessionRouter
      );

      expect(await sessionRouter.read.owner()).to.equal(
        getAddress(owner.account.address)
      );
    });

    it("Should set the right token", async function () {
      const { sessionRouter, tokenMOR } = await loadFixture(
        deployMorpheusTokenAndSessionRouter
      );

      expect(await sessionRouter.read.tokenMOR()).to.equal(getAddress(tokenMOR.address));
    });
  });

  describe("Provider actions", function () {
    it("Should create a provider", async function () {
      const { sessionRouter, provider, expected } = await loadFixture(
        deployWithSingleProvider,
      );
      
      expect(await sessionRouter.read.providerList([0n])).eq(getAddress(provider.account.address));

      const providerData = await sessionRouter.read.providers([provider.account.address]);
      expect(providerData).deep.equals([
        getAddress(expected.address), 
        expected.name, 
        expected.URL, 
        expected.price
      ]);
    })
  });

  describe("Session actions", function(){
    it("Should create session", async function(){
      const { sessionRouter, provider, buyer, tokenMOR, expected, publicClient, buyerBalance } = await loadFixture(
        deployWithSingleProvider,
      );
      
      const approveHash = await tokenMOR.write.approve([sessionRouter.address, buyerBalance], {account: buyer.account})
      await publicClient.waitForTransactionReceipt({hash: approveHash})

      const {request, result: sessionId} = await sessionRouter.simulate.startSession([provider.account.address], {account: buyer.account.address})
      const startSessionHash = await sessionRouter.write.startSession([provider.account.address], request)
      const startSessionTx = await publicClient.waitForTransactionReceipt({hash: startSessionHash})
     
      const {timestamp} = await publicClient.getBlock({blockNumber: startSessionTx.blockNumber})

      // check balances
      expect(await tokenMOR.read.balanceOf([buyer.account.address])).eq(0n);
      expect(await tokenMOR.read.balanceOf([sessionRouter.address])).eq(buyerBalance);
      
      // check session
      expect(sessionId).eq(0n);

      const session = await sessionRouter.read.sessions([sessionId])
      expect(session).to.be.deep.eq([
        sessionId, 
        getAddress(buyer.account.address), 
        getAddress(provider.account.address),
        timestamp,
        timestamp + buyerBalance / expected.price * 60n,
      ])
    })

    it.skip("should close session manually and refund", async function(){
      const { sessionRouter, provider, buyer, tokenMOR, expected, publicClient, buyerBalance } = await loadFixture(
        deployWithSingleProvider,
      );
      
      const approveHash = await tokenMOR.write.approve([sessionRouter.address, expected.price], {account: buyer.account})
      await publicClient.waitForTransactionReceipt({hash: approveHash})

      const {request, result: sessionId} = await sessionRouter.simulate.startSession([provider.account.address], {account: buyer.account.address})
      const startSessionHash = await sessionRouter.write.startSession([provider.account.address], request)
      const startSessionTx = await publicClient.waitForTransactionReceipt({hash: startSessionHash})
     
      const {timestamp} = await publicClient.getBlock({blockNumber: startSessionTx.blockNumber})

      const expectedCloseTimestamp = timestamp + BigInt(30);
      console.log("exp close", expectedCloseTimestamp)
      await time.increaseTo(expectedCloseTimestamp)
      const closeSessionHash = await sessionRouter.write.closeSession([sessionId], {account: buyer.account});
      await publicClient.waitForTransactionReceipt({hash: closeSessionHash})
      const refundValue = expected.price / 2n;

      expect(Number(await tokenMOR.read.balanceOf([buyer.account.address]))).closeTo(Number(refundValue), Number(refundValue / 100n));
      expect(await tokenMOR.read.balanceOf([sessionRouter.address])).eq(refundValue);
    })
  })

  // describe("Deployment", function () {
  //   it("Should set the right unlockTime", async function () {
  //     const { lock, unlockTime } = await loadFixture(deployOneYearLockFixture);

  //     expect(await lock.read.unlockTime()).to.equal(unlockTime);
  //   });

  //   it("Should set the right owner", async function () {
  //     const { lock, owner } = await loadFixture(deployOneYearLockFixture);

  //     expect(await lock.read.owner()).to.equal(
  //       getAddress(owner.account.address)
  //     );
  //   });

  //   it("Should receive and store the funds to lock", async function () {
  //     const { lock, lockedAmount, publicClient } = await loadFixture(
  //       deployOneYearLockFixture
  //     );

  //     expect(
  //       await publicClient.getBalance({
  //         address: lock.address,
  //       })
  //     ).to.equal(lockedAmount);
  //   });

  //   it("Should fail if the unlockTime is not in the future", async function () {
  //     // We don't use the fixture here because we want a different deployment
  //     const latestTime = BigInt(await time.latest());
  //     await expect(
  //       hre.viem.deployContract("Lock", [latestTime], {
  //         value: 1n,
  //       })
  //     ).to.be.rejectedWith("Unlock time should be in the future");
  //   });
  // });

  // describe("Withdrawals", function () {
  //   describe("Validations", function () {
  //     it("Should revert with the right error if called too soon", async function () {
  //       const { lock } = await loadFixture(deployOneYearLockFixture);

  //       await expect(lock.write.withdraw()).to.be.rejectedWith(
  //         "You can't withdraw yet"
  //       );
  //     });

  //     it("Should revert with the right error if called from another account", async function () {
  //       const { lock, unlockTime, otherAccount } = await loadFixture(
  //         deployOneYearLockFixture
  //       );

  //       // We can increase the time in Hardhat Network
  //       await time.increaseTo(unlockTime);

  //       // We retrieve the contract with a different account to send a transaction
  //       const lockAsOtherAccount = await hre.viem.getContractAt(
  //         "Lock",
  //         lock.address,
  //         { client: { wallet: otherAccount } }
  //       );
  //       await expect(lockAsOtherAccount.write.withdraw()).to.be.rejectedWith(
  //         "You aren't the owner"
  //       );
  //     });

  //     it("Shouldn't fail if the unlockTime has arrived and the owner calls it", async function () {
  //       const { lock, unlockTime } = await loadFixture(
  //         deployOneYearLockFixture
  //       );

  //       // Transactions are sent using the first signer by default
  //       await time.increaseTo(unlockTime);

  //       await expect(lock.write.withdraw()).to.be.fulfilled;
  //     });
  //   });

  //   describe("Events", function () {
  //     it("Should emit an event on withdrawals", async function () {
  //       const { lock, unlockTime, lockedAmount, publicClient } =
  //         await loadFixture(deployOneYearLockFixture);

  //       await time.increaseTo(unlockTime);

  //       const hash = await lock.write.withdraw();
  //       await publicClient.waitForTransactionReceipt({ hash });

  //       // get the withdrawal events in the latest block
  //       const withdrawalEvents = await lock.getEvents.Withdrawal();
  //       expect(withdrawalEvents).to.have.lengthOf(1);
  //       expect(withdrawalEvents[0].args.amount).to.equal(lockedAmount);
  //     });
  //   });
  // });
});
