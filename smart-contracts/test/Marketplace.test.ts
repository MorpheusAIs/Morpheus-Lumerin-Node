import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { parseEventLogs } from "viem";
import { deploySingleBid } from "./fixtures";
import {
  catchError,
  getTxTimestamp,
  randomAddress,
  randomBytes32,
} from "./utils";

describe("Marketplace", function () {
  describe("bid actions", function () {
    it("Should create a bid and query by id", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const data = await marketplace.read.bidMap([expectedBid.id]);

      expect(data).to.be.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        pricePerSecond: expectedBid.pricePerSecond,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: expectedBid.deletedAt,
      });
    });

    it("Should error if provider doesn't exist", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const unknownProvider = randomAddress();

      await catchError(marketplace.abi, "ProviderNotFound", async () => {
        await marketplace.simulate.postModelBid(
          [unknownProvider, expectedBid.modelId, expectedBid.pricePerSecond],
          { account: unknownProvider },
        );
      });
    });

    it("Should error if model doesn't exist", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const unknownModel = randomBytes32();

      await catchError(marketplace.abi, "ModelOrAgentNotFound", async () => {
        await marketplace.simulate.postModelBid(
          [expectedBid.providerAddr, unknownModel, expectedBid.pricePerSecond],
          { account: expectedBid.providerAddr },
        );
      });
    });

    it("Should create second bid", async function () {
      const {
        marketplace,
        expectedBid: expBid,
        publicClient,
      } = await loadFixture(deploySingleBid);

      // create new bid with same provider and modelId
      const client = await hre.viem.getWalletClient(expBid.providerAddr);
      const postModelBid = await marketplace.simulate.postModelBid(
        [expBid.providerAddr, expBid.modelId, expBid.pricePerSecond],
        { account: expBid.providerAddr },
      );
      const txHash = await client.writeContract(postModelBid.request);
      const timestamp = await getTxTimestamp(publicClient, txHash);

      // check indexes are updated
      const [, newBids1] = await marketplace.read.getActiveBidsByProvider([
        expBid.providerAddr,
      ]);
      const [, newBids2] = await marketplace.read.getActiveBidsByModelAgent([
        expBid.modelId,
      ]);

      expect(newBids1).to.be.deep.equal(newBids2);
      expect(newBids1.length).to.be.equal(1);
      expect(newBids1[0]).to.be.deep.equal({
        provider: expBid.providerAddr,
        modelAgentId: expBid.modelId,
        pricePerSecond: expBid.pricePerSecond,
        nonce: expBid.nonce + 1n,
        createdAt: timestamp,
        deletedAt: expBid.deletedAt,
      });

      // check old bid is deleted
      const oldBid = await marketplace.read.bidMap([expBid.id]);
      expect(oldBid).to.be.deep.equal({
        provider: expBid.providerAddr,
        modelAgentId: expBid.modelId,
        pricePerSecond: expBid.pricePerSecond,
        nonce: expBid.nonce,
        createdAt: expBid.createdAt,
        deletedAt: timestamp,
      });

      // check old bid is still queried
      const [, oldBids1] = await marketplace.read.getBidsByProvider([
        expBid.providerAddr,
        0n,
        100,
      ]);
      const [, oldBids2] = await marketplace.read.getBidsByModelAgent([
        expBid.modelId,
        0n,
        100,
      ]);
      expect(oldBids1).to.be.deep.equal(oldBids2);
      expect(oldBids1.length).to.be.equal(2);
      expect(oldBids1[1]).to.be.deep.equal({
        provider: expBid.providerAddr,
        modelAgentId: expBid.modelId,
        pricePerSecond: expBid.pricePerSecond,
        nonce: expBid.nonce,
        createdAt: expBid.createdAt,
        deletedAt: timestamp,
      });
    });

    it("Should query by provider", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const [bidIds, bids] = await marketplace.read.getActiveBidsByProvider([
        expectedBid.providerAddr,
      ]);

      expect(bidIds.length).to.equal(1);
      expect(bids.length).to.equal(1);
      expect(bidIds[0]).to.equal(expectedBid.id);
      expect(bids[0]).to.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        pricePerSecond: expectedBid.pricePerSecond,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: expectedBid.deletedAt,
      });
    });

    it("Should query by provider with pagination", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const [_, bids] = await marketplace.read.getActiveBidsByProvider([
        expectedBid.providerAddr,
      ]);

      expect(bids.length).to.equal(1);
      expect(bids[0]).to.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        pricePerSecond: expectedBid.pricePerSecond,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: expectedBid.deletedAt,
      });
    });

    it("Should query by modelId", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const [bidIds, bids] = await marketplace.read.getActiveBidsByModelAgent([
        expectedBid.modelId,
      ]);

      expect(bids.length).to.equal(1);
      expect(bids[0]).to.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        pricePerSecond: expectedBid.pricePerSecond,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: expectedBid.deletedAt,
      });
    });
  });

  describe("delete bid", function () {
    it("Should delete a bid", async function () {
      const { marketplace, expectedBid, publicClient, provider } =
        await loadFixture(deploySingleBid);

      // delete bid
      const deleteBid = await marketplace.simulate.deleteModelAgentBid(
        [expectedBid.id],
        { account: expectedBid.providerAddr },
      );
      const txHash = await provider.writeContract(deleteBid.request);
      await publicClient.waitForTransactionReceipt({
        hash: txHash,
      });

      // check indexes are updated
      const newBids1 = await marketplace.read.getActiveBidsByProvider([
        expectedBid.providerAddr,
      ]);
      const newBids2 = await marketplace.read.getActiveBidsByModelAgent([
        expectedBid.modelId,
      ]);

      expect(newBids1).to.be.deep.equal(newBids2);
      expect(newBids1.length).to.be.equal(0);

      // check bid is deleted
      const data = await marketplace.read.bidMap([expectedBid.id]);
      expect(data).to.be.deep.equal({
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        pricePerSecond: expectedBid.pricePerSecond,
        nonce: expectedBid.nonce,
        createdAt: expectedBid.createdAt,
        deletedAt: await getTxTimestamp(publicClient, txHash),
      });
    });

    it("Should error if bid doesn't exist", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const unknownBid = randomBytes32();

      await catchError(marketplace.abi, "ActiveBidNotFound", async () => {
        await marketplace.simulate.deleteModelAgentBid([unknownBid], {
          account: expectedBid.providerAddr,
        });
      });
    });

    it("Should error if not onwer", async function () {
      const { marketplace, expectedBid } = await loadFixture(deploySingleBid);
      const unknownProvider = randomAddress();

      await catchError(marketplace.abi, "NotSenderOrOwner", async () => {
        await marketplace.simulate.deleteModelAgentBid([expectedBid.id], {
          account: unknownProvider,
        });
      });
    });

    it("Should allow bid owner to delete bid", async function () {
      const { marketplace, expectedBid, publicClient, owner, provider } =
        await loadFixture(deploySingleBid);

      // delete bid
      const deleteBid = await marketplace.simulate.deleteModelAgentBid(
        [expectedBid.id],
        { account: expectedBid.providerAddr },
      );
      const txHash = await provider.writeContract(deleteBid.request);
      await publicClient.waitForTransactionReceipt({
        hash: txHash,
      });
    });

    it("Should allow contract owner to delete bid", async function () {
      const { marketplace, expectedBid, publicClient, owner, provider } =
        await loadFixture(deploySingleBid);

      // delete bid
      const deleteBid = await marketplace.simulate.deleteModelAgentBid(
        [expectedBid.id],
        { account: owner.account.address },
      );
      const txHash = await provider.writeContract(deleteBid.request);
      await publicClient.waitForTransactionReceipt({
        hash: txHash,
      });
    });

    it("Should allow to create bid after it was deleted [H-1]", async function () {
      const { marketplace, expectedBid, publicClient, provider } =
        await loadFixture(deploySingleBid);

      // delete bid
      const deleteBid = await marketplace.simulate.deleteModelAgentBid(
        [expectedBid.id],
        { account: expectedBid.providerAddr },
      );
      const txHash = await provider.writeContract(deleteBid.request);
      await publicClient.waitForTransactionReceipt({
        hash: txHash,
      });

      // create new bid with same provider and modelId
      const postModelBid = await marketplace.simulate.postModelBid(
        [
          expectedBid.providerAddr,
          expectedBid.modelId,
          expectedBid.pricePerSecond,
        ],
        { account: expectedBid.providerAddr },
      );
      const newTxHash = await provider.writeContract(postModelBid.request);
      await publicClient.waitForTransactionReceipt({
        hash: newTxHash,
      });
    });
  });

  describe("bid fee", function () {
    it("should set bid fee", async function () {
      const { marketplace, owner, publicClient } =
        await loadFixture(deploySingleBid);
      const newFee = 100n;
      const txHash = await marketplace.write.setBidFee([newFee], {
        account: owner.account.address,
      });
      const receipt = await publicClient.waitForTransactionReceipt({
        hash: txHash,
      });
      const events = parseEventLogs({
        abi: marketplace.abi,
        logs: receipt.logs,
        eventName: "FeeUpdated",
      });
      expect(events.length).to.be.equal(1);
      expect(events[0].args).to.be.deep.equal({ bidFee: newFee });

      const modelBidFee = await marketplace.read.bidFee();

      expect(modelBidFee).to.be.equal(newFee);
    });

    it("should collect bid fee", async function () {
      const {
        marketplace,
        owner,
        expectedBid,
        publicClient,
        provider,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      const newFee = 100n;
      await marketplace.write.setBidFee([newFee], {
        account: owner.account.address,
      });

      await tokenMOR.write.transfer([expectedBid.providerAddr, 100n], {
        account: owner.account,
      });

      // check balance before
      const balanceBefore = await tokenMOR.read.balanceOf([
        marketplace.address,
      ]);

      // add bid
      await tokenMOR.write.approve(
        [marketplace.address, expectedBid.pricePerSecond + newFee],
        {
          account: expectedBid.providerAddr,
        },
      );
      const postModelBid = await marketplace.simulate.postModelBid(
        [
          expectedBid.providerAddr,
          expectedBid.modelId,
          expectedBid.pricePerSecond,
        ],
        { account: expectedBid.providerAddr },
      );
      const txHash = await provider.writeContract(postModelBid.request);
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      // check balance after
      const balanceAfter = await tokenMOR.read.balanceOf([marketplace.address]);
      expect(balanceAfter - balanceBefore).to.be.equal(newFee);
    });

    it("should allow withdrawal by owner", async function () {
      const {
        marketplace,
        owner,
        expectedBid,
        publicClient,
        provider,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      const newFee = 100n;
      await marketplace.write.setBidFee([newFee], {
        account: owner.account.address,
      });
      await tokenMOR.write.transfer([expectedBid.providerAddr, 100n], {
        account: owner.account,
      });

      // add bid
      await tokenMOR.write.approve(
        [marketplace.address, expectedBid.pricePerSecond + newFee],
        {
          account: expectedBid.providerAddr,
        },
      );
      const postModelBid = await marketplace.simulate.postModelBid(
        [
          expectedBid.providerAddr,
          expectedBid.modelId,
          expectedBid.pricePerSecond,
        ],
        { account: expectedBid.providerAddr },
      );
      const txHash = await provider.writeContract(postModelBid.request);
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      // check balance after
      const balanceBefore = await tokenMOR.read.balanceOf([
        owner.account.address,
      ]);
      await marketplace.write.withdraw([owner.account.address, newFee], {
        account: owner.account.address,
      });
      const balanceAfter = await tokenMOR.read.balanceOf([
        owner.account.address,
      ]);

      expect(balanceAfter - balanceBefore).to.be.equal(newFee);
    });

    it("should not allow withdrawal by any other account except owner", async function () {
      const {
        marketplace,
        owner,
        expectedBid,
        publicClient,
        provider,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      const newFee = 100n;
      await marketplace.write.setBidFee([newFee], {
        account: owner.account.address,
      });

      await tokenMOR.write.transfer([expectedBid.providerAddr, 100n], {
        account: owner.account,
      });

      // add bid
      await tokenMOR.write.approve(
        [marketplace.address, expectedBid.pricePerSecond + newFee],
        {
          account: expectedBid.providerAddr,
        },
      );
      const postModelBid = await marketplace.simulate.postModelBid(
        [
          expectedBid.providerAddr,
          expectedBid.modelId,
          expectedBid.pricePerSecond,
        ],
        { account: expectedBid.providerAddr },
      );
      const txHash = await provider.writeContract(postModelBid.request);
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      // check balance after
      await catchError(
        (await hre.artifacts.readArtifact("OwnershipFacet")).abi,
        "NotContractOwner",
        async () => {
          await marketplace.write.withdraw([expectedBid.providerAddr, newFee], {
            account: expectedBid.providerAddr,
          });
        },
      );
    });

    it("should not allow withdrawal if not enough balance", async function () {
      const { marketplace, owner, tokenMOR } =
        await loadFixture(deploySingleBid);

      await catchError(marketplace.abi, "NotEnoughBalance", async () => {
        await marketplace.write.withdraw([owner.account.address, 100000000n], {
          account: owner.account.address,
        });
      });
    });
  });
});
