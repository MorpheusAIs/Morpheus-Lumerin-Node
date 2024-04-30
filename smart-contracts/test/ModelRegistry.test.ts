import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { getAddress } from "viem";
import { deployDiamond, deploySingleModel } from "./fixtures";
import { expectError, getHex, getTxTimestamp, randomAddress, randomBytes32 } from "./utils";

describe("Model registry", function () {
  describe("Actions", function () {
    it("Should count correctly", async function () {
      const { modelRegistry } = await loadFixture(deployDiamond);

      expect(await modelRegistry.read.modelGetCount()).eq(0n);
    });

    it("Should register", async function () {
      const { modelRegistry, expectedModel } = await loadFixture(deploySingleModel);
      const data = await modelRegistry.read.modelMap([expectedModel.modelId]);
      const events = await modelRegistry.getEvents.ModelRegisteredUpdated({
        modelId: expectedModel.modelId,
        owner: expectedModel.owner,
      });

      expect(await modelRegistry.read.modelGetCount()).eq(1n);
      expect(await modelRegistry.read.models([0n])).eq(expectedModel.modelId);
      expect(data).deep.equal([
        expectedModel.ipfsCID,
        expectedModel.fee,
        expectedModel.stake,
        getAddress(expectedModel.owner),
        expectedModel.name,
        // expectedModel.tags,
        expectedModel.timestamp,
        expectedModel.isDeleted,
      ]);
      expect(events.length).eq(1);
    });

    it("Should error when registering with insufficient stake", async function () {
      const { modelRegistry, owner } = await loadFixture(deployDiamond);
      const minStake = 100n;
      await modelRegistry.write.modelSetMinStake([minStake]);
      try {
        await modelRegistry.simulate.modelRegister([
          randomBytes32(),
          randomBytes32(),
          0n,
          0n,
          owner.account.address,
          "a",
          [],
        ]);
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, modelRegistry.abi, "StakeTooLow");
      }
    });

    it("Should error when registering with insufficient allowance", async function () {
      const { modelRegistry, owner, tokenMOR } = await loadFixture(deployDiamond);
      try {
        await modelRegistry.simulate.modelRegister([
          randomBytes32(),
          randomBytes32(),
          0n,
          100n,
          owner.account.address,
          "a",
          [],
        ]);
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, tokenMOR.abi, "ERC20InsufficientAllowance");
      }
    });

    it("Should error when register account doesnt match sender account", async function () {
      const [, , user] = await hre.viem.getWalletClients();

      const { modelRegistry, tokenMOR } = await loadFixture(deployDiamond);
      await tokenMOR.write.approve([modelRegistry.address, 100n]);
      try {
        await modelRegistry.simulate.modelRegister(
          [randomBytes32(), randomBytes32(), 0n, 100n, randomAddress(), "a", []],
          {
            account: user.account.address,
          }
        );
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, modelRegistry.abi, "NotSenderOrOwner");
      }
    });

    it("Should deregister by owner", async function () {
      const { modelRegistry, owner, expectedModel } = await loadFixture(deploySingleModel);

      await modelRegistry.write.modelDeregister([expectedModel.modelId], {
        account: owner.account,
      });
      const events = await modelRegistry.getEvents.ModelDeregistered({
        modelId: expectedModel.modelId,
        owner: expectedModel.owner,
      });

      expect(await modelRegistry.read.modelGetCount()).eq(0n);
      expect((await modelRegistry.read.modelMap([expectedModel.modelId]))[6]).to.equal(true); // 7 is index of isDeleted field
      expect(events.length).eq(1);
      await expect(modelRegistry.read.modelGetByIndex([0n])).rejectedWith(
        /.*reverted with panic code 0x32 (Array accessed at an out-of-bounds or negative index)*/
      );
      expect(await modelRegistry.read.models([0n])).equals(expectedModel.modelId);
    });

    it("Should deregister by admin", async function () {
      const { modelRegistry, owner, expectedModel } = await loadFixture(deploySingleModel);

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
      try {
        await modelRegistry.write.modelDeregister([randomBytes32()], {
          account: owner.account,
        });
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, modelRegistry.abi, "KeyNotFound");
      }
    });

    it("Should return stake on deregister", async function () {
      const { modelRegistry, tokenMOR, expectedModel } = await loadFixture(deploySingleModel);

      const balanceBefore = await tokenMOR.read.balanceOf([expectedModel.owner]);
      await modelRegistry.write.modelDeregister([expectedModel.modelId]);
      const balanceAfter = await tokenMOR.read.balanceOf([expectedModel.owner]);

      expect(balanceAfter - balanceBefore).eq(expectedModel.stake);
    });

    it("Should update existing model", async function () {
      const { modelRegistry, provider, tokenMOR, expectedModel, publicClient, owner } =
        await loadFixture(deploySingleModel);
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
      const providerData = await modelRegistry.read.modelMap([expectedModel.modelId]);

      expect(providerData).deep.equal([
        updates.ipfsCID,
        updates.fee,
        expectedModel.stake + updates.addStake,
        getAddress(updates.owner),
        updates.name,
        // expectedModel.tags,
        timestamp,
        expectedModel.isDeleted,
      ]);
    });

    it("Should emit event on update", async function () {
      const { modelRegistry, provider, tokenMOR, expectedModel, owner } = await loadFixture(
        deploySingleModel
      );
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
      const { modelRegistry, provider, expectedModel } = await loadFixture(deploySingleModel);
      const [modelId, providerData] = await modelRegistry.read.modelGetByIndex([0n]);

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
      const { modelRegistry, provider, expectedModel } = await loadFixture(deploySingleModel);

      const providerData = await modelRegistry.read.modelMap([expectedModel.modelId]);
      expect(providerData).deep.equal([
        expectedModel.ipfsCID,
        expectedModel.fee,
        expectedModel.stake,
        getAddress(expectedModel.owner),
        expectedModel.name,
        // expectedModel.tags,
        expectedModel.timestamp,
        expectedModel.isDeleted,
      ]);
    });
  });

  describe("Min stake", function () {
    it("Should set min stake", async function () {
      const { modelRegistry, owner } = await loadFixture(deployDiamond);
      const minStake = 100n;

      await modelRegistry.write.modelSetMinStake([minStake], { account: owner.account });
      const events = await modelRegistry.getEvents.ModelMinStakeUpdated();
      expect(await modelRegistry.read.modelMinStake()).eq(minStake);
      expect(events.length).eq(1);
      expect(events[0].args.newStake).eq(minStake);
    });

    it("Should error when not owner is setting min stake", async function () {
      const { modelRegistry, provider } = await loadFixture(deploySingleModel);
      try {
        await modelRegistry.write.modelSetMinStake([100n], { account: provider.account });
        expect.fail("Expected error");
      } catch (e) {
        expectError(
          e,
          (await hre.artifacts.readArtifact("OwnershipFacet")).abi,
          "NotContractOwner"
        );
      }
    });
  });
});
