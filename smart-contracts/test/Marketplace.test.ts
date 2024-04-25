import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { erc20Abi, getAddress, parseEventLogs } from "viem";
import { deployMarketplace } from "./fixtures";
import { expectError, getTxTimestamp } from "./utils";

describe("Marketplace", function () {
  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      const { sessionRouter, owner } = await loadFixture(deployMarketplace);

      expect(await sessionRouter.read.owner()).to.equal(getAddress(owner.account.address));
    });

    it("Should set the right token", async function () {
      const { sessionRouter, tokenMOR } = await loadFixture(deployMarketplace);

      expect(await sessionRouter.read.token()).to.equal(getAddress(tokenMOR.address));
    });

    it("Should set registries correctly", async function () {
      const { sessionRouter, modelRegistry, providerRegistry } = await loadFixture(
        deployMarketplace
      );

      expect(await sessionRouter.read.modelRegistry()).eq(getAddress(modelRegistry.address));
      expect(await sessionRouter.read.providerRegistry()).eq(getAddress(providerRegistry.address));
    });
  });

  describe("bid actions", function () {
    it("Should create a bid and query by id", async function () {
      const { sessionRouter, expectedBid } = await loadFixture(deployMarketplace);
      const data = await sessionRouter.read.bidMap([expectedBid.id]);

      expect(data).to.be.deep.equal([
        expectedBid.providerAddr,
        expectedBid.modelId,
        expectedBid.amount,
        expectedBid.nonce,
        expectedBid.createdAt,
        expectedBid.deletedAt,
      ]);
    });

    it("Should create a bid and query by id", async function () {
      const { sessionRouter, expectedBid } = await loadFixture(deployMarketplace);
      const data = await sessionRouter.read.bidMap([expectedBid.id]);

      expect(data).to.be.deep.equal([
        expectedBid.providerAddr,
        expectedBid.modelId,
        expectedBid.amount,
        expectedBid.nonce,
        expectedBid.createdAt,
        expectedBid.deletedAt,
      ]);
    });

    it("Should create second bid", async function () {
      const {
        sessionRouter,
        expectedBid: expBid,
        publicClient,
      } = await loadFixture(deployMarketplace);

      // create new bid with same provider and modelId
      const client = await hre.viem.getWalletClient(expBid.providerAddr);
      const postModelBid = await sessionRouter.simulate.postModelBid(
        [expBid.providerAddr, expBid.modelId, expBid.amount],
        { account: expBid.providerAddr }
      );
      const txHash = await client.writeContract(postModelBid.request);
      const timestamp = await getTxTimestamp(publicClient, txHash);

      // check indexes are updated
      const newBids1 = await sessionRouter.read.getActiveBidsByProvider([expBid.providerAddr]);
      const newBids2 = await sessionRouter.read.getActiveBidsByModelAgent([expBid.modelId]);

      expect(newBids1).to.be.deep.equal(newBids2);
      expect(newBids1.length).to.be.equal(1);
      expect(newBids1[0]).to.be.deep.equal({
        provider: expBid.providerAddr,
        modelAgentId: expBid.modelId,
        amount: expBid.amount,
        nonce: expBid.nonce + 1n,
        createdAt: timestamp,
        deletedAt: expBid.deletedAt,
      });

      // check old bid is deleted
      const oldBid = await sessionRouter.read.bidMap([expBid.id]);
      expect(oldBid).to.be.deep.equal([
        expBid.providerAddr,
        expBid.modelId,
        expBid.amount,
        expBid.nonce,
        expBid.createdAt,
        timestamp,
      ]);

      // check old bid is still queried
      const oldBids1 = await sessionRouter.read.getBidsByProvider([expBid.providerAddr, 0n, 100]);
      const oldBids2 = await sessionRouter.read.getBidsByModelAgent([expBid.modelId, 0n, 100]);
      expect(oldBids1).to.be.deep.equal(oldBids2);
      expect(oldBids1.length).to.be.equal(2);
      expect(oldBids1[1]).to.be.deep.equal({
        provider: expBid.providerAddr,
        modelAgentId: expBid.modelId,
        amount: expBid.amount,
        nonce: expBid.nonce,
        createdAt: expBid.createdAt,
        deletedAt: timestamp,
      });
    });

    it("Should query by provider", async function () {
      const { sessionRouter, expectedBid } = await loadFixture(deployMarketplace);
      const data = await sessionRouter.read.getActiveBidsByProvider([expectedBid.providerAddr]);

      expect(data.length).to.equal(1);
      expect(data[0]).to.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        amount: expectedBid.amount,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: expectedBid.deletedAt,
      });
    });

    it("Should query by provider with pagination", async function () {
      const { sessionRouter, expectedBid } = await loadFixture(deployMarketplace);
      const data = await sessionRouter.read.getActiveBidsByProvider([expectedBid.providerAddr]);

      expect(data.length).to.equal(1);
      expect(data[0]).to.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        amount: expectedBid.amount,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: expectedBid.deletedAt,
      });
    });

    it("Should query by modelId", async function () {
      const { sessionRouter, expectedBid } = await loadFixture(deployMarketplace);
      const data = await sessionRouter.read.getActiveBidsByModelAgent([expectedBid.modelId]);

      expect(data.length).to.equal(1);
      expect(data[0]).to.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        amount: expectedBid.amount,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: expectedBid.deletedAt,
      });
    });
  });

  describe("bid fee", function () {
    it("should set bid fee", async function () {
      const { sessionRouter, owner, publicClient } = await loadFixture(deployMarketplace);
      const newFee = 100n;
      const txHash = await sessionRouter.write.setBidFee([newFee, newFee], {
        account: owner.account.address,
      });
      const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
      const events = parseEventLogs({
        abi: sessionRouter.abi,
        logs: receipt.logs,
        eventName: "FeeUpdated",
      });
      expect(events.length).to.be.equal(1);
      expect(events[0].args).to.be.deep.equal({ modelFee: newFee, agentFee: newFee });

      const modelBidFee = await sessionRouter.read.modelBidFee();
      const agentBidFee = await sessionRouter.read.agentBidFee();
      expect(modelBidFee).to.be.equal(newFee);
      expect(agentBidFee).to.be.equal(newFee);
    });

    it("should collect bid fee and withdraw", async function () {
      const { sessionRouter, owner, expectedBid, publicClient, provider, tokenMOR } =
        await loadFixture(deployMarketplace);
      const newFee = 100n;
      await sessionRouter.write.setBidFee([newFee, newFee], {
        account: owner.account.address,
      });

      // check balance before
      const balanceBefore = await tokenMOR.read.balanceOf([sessionRouter.address]);

      // add bid
      await tokenMOR.write.approve([sessionRouter.address, expectedBid.amount + newFee], {
        account: expectedBid.providerAddr,
      });
      const postModelBid = await sessionRouter.simulate.postModelBid(
        [expectedBid.providerAddr, expectedBid.modelId, expectedBid.amount],
        { account: expectedBid.providerAddr }
      );
      const txHash = await provider.writeContract(postModelBid.request);
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      // check balance after
      const balanceAfter = await tokenMOR.read.balanceOf([sessionRouter.address]);
      expect(balanceAfter - balanceBefore).to.be.equal(newFee);
    });

    it("should allow withdrawal by owner", async function () {
      const { sessionRouter, owner, expectedBid, publicClient, provider, tokenMOR } =
        await loadFixture(deployMarketplace);
      const newFee = 100n;
      await sessionRouter.write.setBidFee([newFee, newFee], {
        account: owner.account.address,
      });

      // add bid
      await tokenMOR.write.approve([sessionRouter.address, expectedBid.amount + newFee], {
        account: expectedBid.providerAddr,
      });
      const postModelBid = await sessionRouter.simulate.postModelBid(
        [expectedBid.providerAddr, expectedBid.modelId, expectedBid.amount],
        { account: expectedBid.providerAddr }
      );
      const txHash = await provider.writeContract(postModelBid.request);
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      // check balance after
      const balanceBefore = await tokenMOR.read.balanceOf([owner.account.address]);
      await sessionRouter.write.withdraw([owner.account.address, newFee], {
        account: owner.account.address,
      });
      const balanceAfter = await tokenMOR.read.balanceOf([owner.account.address]);

      expect(balanceAfter - balanceBefore).to.be.equal(newFee);
    });

    it("should not allow withdrawal by any other account except owner", async function () {
      const { sessionRouter, owner, expectedBid, publicClient, provider, tokenMOR } =
        await loadFixture(deployMarketplace);
      const newFee = 100n;
      await sessionRouter.write.setBidFee([newFee, newFee], {
        account: owner.account.address,
      });

      // add bid
      await tokenMOR.write.approve([sessionRouter.address, expectedBid.amount + newFee], {
        account: expectedBid.providerAddr,
      });
      const postModelBid = await sessionRouter.simulate.postModelBid(
        [expectedBid.providerAddr, expectedBid.modelId, expectedBid.amount],
        { account: expectedBid.providerAddr }
      );
      const txHash = await provider.writeContract(postModelBid.request);
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      // check balance after
      try {
        await sessionRouter.write.withdraw([expectedBid.providerAddr, newFee], {
          account: expectedBid.providerAddr,
        });
        expect.fail("Should have thrown an error");
      } catch (e) {
        expect((e as Error).message).includes("Ownable: caller is not the owner");
      }
    });

    it("should not allow withdrawal if not enough balance", async function () {
      const { sessionRouter, owner, tokenMOR } = await loadFixture(deployMarketplace);

      try {
        await sessionRouter.write.withdraw([owner.account.address, 100000000n], {
          account: owner.account.address,
        });
        expect.fail("Should have thrown an error");
      } catch (e) {
        expectError(e, sessionRouter.abi, "NotEnoughBalance");
      }
    });
  });
});
