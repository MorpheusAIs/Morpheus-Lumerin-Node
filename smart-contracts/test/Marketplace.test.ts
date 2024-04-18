import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { getAddress } from "viem";
import { deployMarketplace, deployProviderRegistry, deploySingleProvider } from "./fixtures";
import { expectError, getTxTimestamp } from "./utils";
import exp from "constants";

describe("Marketplace", function () {
  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      const { marketplace, owner } = await loadFixture(deployMarketplace);

      expect(await marketplace.read.owner()).to.equal(getAddress(owner.account.address));
    });

    it("Should set the right token", async function () {
      const { marketplace, tokenMOR } = await loadFixture(deployMarketplace);

      expect(await marketplace.read.token()).to.equal(getAddress(tokenMOR.address));
    });

    it("Should set registries correctly", async function () {
      const { marketplace, modelRegistry, providerRegistry } = await loadFixture(deployMarketplace);

      expect(await marketplace.read.modelRegistry()).eq(getAddress(modelRegistry.address));
      expect(await marketplace.read.providerRegistry()).eq(getAddress(providerRegistry.address));
    });
  });

  describe("bid actions", function () {
    it("Should create a bid and query by id", async function () {
      const { marketplace, expectedBid } = await loadFixture(deployMarketplace);
      const data = await marketplace.read.map([expectedBid.id]);

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
      const { marketplace, expectedBid } = await loadFixture(deployMarketplace);
      const data = await marketplace.read.map([expectedBid.id]);

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
        marketplace,
        expectedBid: expBid,
        publicClient,
      } = await loadFixture(deployMarketplace);

      // create new bid with same provider and modelId
      const client = await hre.viem.getWalletClient(expBid.providerAddr);
      const postModelBid = await marketplace.simulate.postModelBid(
        [expBid.providerAddr, expBid.modelId, expBid.amount],
        { account: expBid.providerAddr }
      );
      const txHash = await client.writeContract(postModelBid.request);
      const timestamp = await getTxTimestamp(publicClient, txHash);

      // check indexes are updated
      const newBids1 = await marketplace.read.getActiveBidsByProvider([expBid.providerAddr]);
      const newBids2 = await marketplace.read.getActiveBidsByModelAgent([expBid.modelId]);

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
      const oldBid = await marketplace.read.map([expBid.id]);
      expect(oldBid).to.be.deep.equal([
        expBid.providerAddr,
        expBid.modelId,
        expBid.amount,
        expBid.nonce,
        expBid.createdAt,
        timestamp,
      ]);

      // check old bid is still queried
      const oldBids1 = await marketplace.read.getBidsByProvider([expBid.providerAddr, 0n, 100]);
      const oldBids2 = await marketplace.read.getBidsByModelAgent([expBid.modelId, 0n, 100]);
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
      const { marketplace, expectedBid } = await loadFixture(deployMarketplace);
      const data = await marketplace.read.getActiveBidsByProvider([expectedBid.providerAddr]);

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
      const { marketplace, expectedBid } = await loadFixture(deployMarketplace);
      const data = await marketplace.read.getActiveBidsByProvider([expectedBid.providerAddr]);

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
      const { marketplace, expectedBid } = await loadFixture(deployMarketplace);
      const data = await marketplace.read.getActiveBidsByModelAgent([expectedBid.modelId]);

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
});
