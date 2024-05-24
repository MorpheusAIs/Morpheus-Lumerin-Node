import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { getAddress } from "viem";
import { deployDiamond, deploySingleModel } from "./fixtures";
import {
  catchError,
  getHex,
  getTxTimestamp,
  randomAddress,
  randomBytes32,
} from "./utils";

describe("Model registry", function () {
  describe("Actions", function () {
    it("Should count correctly", async function () {
      const { modelRegistry } = await loadFixture(deployDiamond);

      expect(await modelRegistry.read.modelGetCount()).eq(0n);
    });

    it("Should register", async function () {
      const { modelRegistry, expectedModel } =
        await loadFixture(deploySingleModel);
      const data = await modelRegistry.read.modelMap([expectedModel.modelId]);
      const events = await modelRegistry.getEvents.ModelRegisteredUpdated({
        modelId: expectedModel.modelId,
        owner: expectedModel.owner,
      });

      expect(await modelRegistry.read.modelGetCount()).eq(1n);
      expect(await modelRegistry.read.models([0n])).eq(expectedModel.modelId);
      expect(data).deep.equal({
        ipfsCID: expectedModel.ipfsCID,
        fee: expectedModel.fee,
        stake: expectedModel.stake,
        owner: getAddress(expectedModel.owner),
        name: expectedModel.name,
        tags: expectedModel.tags,
        timestamp: expectedModel.timestamp,
        isDeleted: expectedModel.isDeleted,
      });
      expect(events.length).eq(1);
    });

    it("Should error when registering with insufficient stake", async function () {
      const { modelRegistry, owner } = await loadFixture(deployDiamond);
      const minStake = 100n;
      await modelRegistry.write.modelSetMinStake([minStake]);

      await catchError(modelRegistry.abi, "StakeTooLow", async () => {
        await modelRegistry.simulate.modelRegister([
          randomBytes32(),
          randomBytes32(),
          0n,
          0n,
          owner.account.address,
          "a",
          [],
        ]);
      });
    });

    it("Should error when registering with insufficient allowance", async function () {
      const { modelRegistry, owner, tokenMOR } =
        await loadFixture(deployDiamond);

      await catchError(tokenMOR.abi, "ERC20InsufficientAllowance", async () => {
        await modelRegistry.simulate.modelRegister([
          randomBytes32(),
          randomBytes32(),
          0n,
          100n,
          owner.account.address,
          "a",
          [],
        ]);
      });
    });

    it("Should error when register account doesnt match sender account", async function () {
      const [, , user] = await hre.viem.getWalletClients();

      const { modelRegistry, tokenMOR } = await loadFixture(deployDiamond);
      await tokenMOR.write.approve([modelRegistry.address, 100n]);

      await catchError(modelRegistry.abi, "NotSenderOrOwner", async () => {
        await modelRegistry.simulate.modelRegister(
          [
            randomBytes32(),
            randomBytes32(),
            0n,
            100n,
            randomAddress(),
            "a",
            [],
          ],
          {
            account: user.account.address,
          },
        );
      });
    });

    it("Should deregister by owner", async function () {
      const { modelRegistry, owner, expectedModel } =
        await loadFixture(deploySingleModel);

      await modelRegistry.write.modelDeregister([expectedModel.modelId], {
        account: owner.account,
      });
      const events = await modelRegistry.getEvents.ModelDeregistered({
        modelId: expectedModel.modelId,
        owner: expectedModel.owner,
      });

      expect(await modelRegistry.read.modelGetCount()).eq(0n);
      expect(
        (await modelRegistry.read.modelMap([expectedModel.modelId])).isDeleted,
      ).to.equal(true);
      expect(events.length).eq(1);
      await expect(modelRegistry.read.modelGetByIndex([0n])).rejectedWith(
        /.*reverted with panic code 0x32 (Array accessed at an out-of-bounds or negative index)*/,
      );
      expect(await modelRegistry.read.models([0n])).equals(
        expectedModel.modelId,
      );
    });

    it("Should deregister by admin", async function () {
      const { modelRegistry, owner, expectedModel } =
        await loadFixture(deploySingleModel);

      await modelRegistry.write.modelDeregister([expectedModel.modelId], {
        account: owner.account,
      });
      const events = await modelRegistry.getEvents.ModelDeregistered({
        modelId: expectedModel.modelId,
      });

      expect(await modelRegistry.read.modelGetCount()).eq(0n);
      expect(events.length).eq(1);
    });

    it("Should error if model not known by admin", async function () {
      const { modelRegistry, owner } = await loadFixture(deploySingleModel);
      await catchError(modelRegistry.abi, "KeyNotFound", async () => {
        await modelRegistry.write.modelDeregister([randomBytes32()], {
          account: owner.account,
        });
      });
    });

    it("Should return stake on deregister", async function () {
      const { modelRegistry, tokenMOR, expectedModel } =
        await loadFixture(deploySingleModel);

      const balanceBefore = await tokenMOR.read.balanceOf([
        expectedModel.owner,
      ]);
      await modelRegistry.write.modelDeregister([expectedModel.modelId]);
      const balanceAfter = await tokenMOR.read.balanceOf([expectedModel.owner]);

      expect(balanceAfter - balanceBefore).eq(expectedModel.stake);
    });

    it("Should update existing model", async function () {
      const {
        modelRegistry,
        provider,
        tokenMOR,
        expectedModel,
        publicClient,
        owner,
      } = await loadFixture(deploySingleModel);
      const updates = {
        ipfsCID: getHex(Buffer.from("ipfs://new-ipfsaddress")),
        fee: expectedModel.fee * 2n,
        addStake: expectedModel.stake * 2n,
        owner: provider.account.address,
        name: "Llama 3.0",
        tags: ["llama", "smart", "angry"],
      };
      await tokenMOR.write.approve([modelRegistry.address, updates.addStake], {
        account: owner.account,
      });

      const txHash = await modelRegistry.write.modelRegister([
        expectedModel.modelId,
        updates.ipfsCID,
        updates.fee,
        updates.addStake,
        updates.owner,
        updates.name,
        updates.tags,
      ]);
      const timestamp = await getTxTimestamp(publicClient, txHash);
      const providerData = await modelRegistry.read.modelMap([
        expectedModel.modelId,
      ]);

      expect(providerData).deep.equal({
        ipfsCID: updates.ipfsCID,
        fee: updates.fee,
        stake: expectedModel.stake + updates.addStake,
        owner: getAddress(updates.owner),
        name: updates.name,
        tags: updates.tags,
        timestamp: timestamp,
        isDeleted: expectedModel.isDeleted,
      });
    });

    it("Should emit event on update", async function () {
      const { modelRegistry, provider, tokenMOR, expectedModel, owner } =
        await loadFixture(deploySingleModel);
      const updates = {
        ipfsCID: getHex(Buffer.from("ipfs://new-ipfsaddress")),
        fee: expectedModel.fee * 2n,
        addStake: expectedModel.stake * 2n,
        owner: provider.account.address,
        name: "Llama 3.0",
        tags: ["llama", "smart", "angry"],
      };

      await tokenMOR.write.approve([modelRegistry.address, updates.addStake]);
      await modelRegistry.write.modelRegister([
        expectedModel.modelId,
        updates.ipfsCID,
        updates.fee,
        updates.addStake,
        updates.owner,
        updates.name,
        updates.tags,
      ]);

      const events = await modelRegistry.getEvents.ModelRegisteredUpdated({
        modelId: expectedModel.modelId,
        owner: getAddress(provider.account.address),
      });
      expect(events.length).eq(1);
    });
  });

  describe("Getters", function () {
    it("Should get by index", async function () {
      const { modelRegistry, provider, expectedModel } =
        await loadFixture(deploySingleModel);
      const [modelId, providerData] = await modelRegistry.read.modelGetByIndex([
        0n,
      ]);

      expect(modelId).eq(expectedModel.modelId);
      expect(providerData).deep.equal({
        ipfsCID: expectedModel.ipfsCID,
        fee: expectedModel.fee,
        stake: expectedModel.stake,
        owner: getAddress(expectedModel.owner),
        name: expectedModel.name,
        tags: expectedModel.tags,
        timestamp: expectedModel.timestamp,
        isDeleted: expectedModel.isDeleted,
      });
    });

    it("Should get by address", async function () {
      const { modelRegistry, provider, expectedModel } =
        await loadFixture(deploySingleModel);

      const providerData = await modelRegistry.read.modelMap([
        expectedModel.modelId,
      ]);
      expect(providerData).deep.equal({
        ipfsCID: expectedModel.ipfsCID,
        fee: expectedModel.fee,
        stake: expectedModel.stake,
        owner: getAddress(expectedModel.owner),
        name: expectedModel.name,
        tags: expectedModel.tags,
        timestamp: expectedModel.timestamp,
        isDeleted: expectedModel.isDeleted,
      });
    });
  });

  describe("Min stake", function () {
    it("Should set min stake", async function () {
      const { modelRegistry, owner } = await loadFixture(deployDiamond);
      const minStake = 100n;

      await modelRegistry.write.modelSetMinStake([minStake], {
        account: owner.account,
      });
      const events = await modelRegistry.getEvents.ModelMinStakeUpdated();
      expect(await modelRegistry.read.modelMinStake()).eq(minStake);
      expect(events.length).eq(1);
      expect(events[0].args.newStake).eq(minStake);
    });

    it("Should error when not owner is setting min stake", async function () {
      const { modelRegistry, provider } = await loadFixture(deploySingleModel);
      await catchError(modelRegistry.abi, "NotContractOwner", async () => {
        await modelRegistry.write.modelSetMinStake([100n], {
          account: provider.account,
        });
      });
    });
  });
});
