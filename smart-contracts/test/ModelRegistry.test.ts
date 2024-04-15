import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { getAddress } from "viem";
import { deployModelRegistry, deploySingleModel } from "./fixtures";
import { expectError, getHex, getTxTimestamp, randomAddress, randomBytes32 } from "./utils";

describe("Model registry", function () {
  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      const { modelRegistry, owner } = await loadFixture(deployModelRegistry);

      expect(await modelRegistry.read.owner()).to.equal(getAddress(owner.account.address));
    });

    it("Should set the right token", async function () {
      const { modelRegistry, tokenMOR } = await loadFixture(deployModelRegistry);

      expect(await modelRegistry.read.token()).to.equal(getAddress(tokenMOR.address));
    });

    it("Should count correctly", async function () {
      const { modelRegistry } = await loadFixture(deployModelRegistry);

      expect(await modelRegistry.read.getCount()).eq(0n);
    });
  });

  describe("Actions", function () {
    it("Should register", async function () {
      const { modelRegistry, expected } = await loadFixture(deploySingleModel);
      const data = await modelRegistry.read.map([expected.modelId]);
      const events = await modelRegistry.getEvents.RegisteredUpdated({
        modelId: expected.modelId,
        owner: expected.owner,
      });

      expect(await modelRegistry.read.getCount()).eq(1n);
      expect(await modelRegistry.read.models([0n])).eq(expected.modelId);
      expect(data).deep.equal([
        expected.ipfsCID,
        expected.fee,
        expected.stake,
        getAddress(expected.owner),
        expected.name,
        // expected.tags,
        expected.timestamp,
        expected.isDeleted,
      ]);
      expect(events.length).eq(1);
    });

    it("Should error when registering with insufficient stake", async function () {
      const { modelRegistry, owner } = await loadFixture(deployModelRegistry);
      const minStake = 100n;
      await modelRegistry.write.setMinStake([minStake]);
      try {
        await modelRegistry.simulate.register([
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
      const { modelRegistry, owner, tokenMOR } = await loadFixture(deployModelRegistry);
      try {
        await modelRegistry.simulate.register([
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

      const { modelRegistry, tokenMOR } = await loadFixture(deployModelRegistry);
      await tokenMOR.write.approve([modelRegistry.address, 100n]);
      try {
        await modelRegistry.simulate.register(
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
      const { modelRegistry, owner, expected } = await loadFixture(deploySingleModel);

      await modelRegistry.write.deregister([expected.modelId], {
        account: owner.account,
      });
      const events = await modelRegistry.getEvents.Deregistered({
        modelId: expected.modelId,
        owner: expected.owner,
      });

      expect(await modelRegistry.read.getCount()).eq(0n);
      expect((await modelRegistry.read.map([expected.modelId]))[6]).to.equal(true); // 7 is index of isDeleted field
      expect(events.length).eq(1);
      await expect(modelRegistry.read.getByIndex([0n])).rejectedWith(
        /.*reverted with panic code 0x32 (Array accessed at an out-of-bounds or negative index)*/
      );
      expect(await modelRegistry.read.models([0n])).equals(expected.modelId);
    });

    it("Should deregister by admin", async function () {
      const { modelRegistry, owner, expected } = await loadFixture(deploySingleModel);

      await modelRegistry.write.deregister([expected.modelId], {
        account: owner.account,
      });
      const events = await modelRegistry.getEvents.Deregistered({
        modelId: expected.modelId,
      });

      expect(await modelRegistry.read.getCount()).eq(0n);
      expect(events.length).eq(1);
    });

    it("Should error if model not known by admin", async function () {
      const { modelRegistry, owner } = await loadFixture(deploySingleModel);
      try {
        await modelRegistry.write.deregister([randomBytes32()], {
          account: owner.account,
        });
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, modelRegistry.abi, "KeyNotFound");
      }
    });

    it("Should return stake on deregister", async function () {
      const { modelRegistry, tokenMOR, expected } = await loadFixture(deploySingleModel);

      const balanceBefore = await tokenMOR.read.balanceOf([expected.owner]);
      await modelRegistry.write.deregister([expected.modelId]);
      const balanceAfter = await tokenMOR.read.balanceOf([expected.owner]);

      expect(balanceAfter - balanceBefore).eq(expected.stake);
    });

    it("Should update existing model", async function () {
      const { modelRegistry, provider, tokenMOR, expected, publicClient, owner } =
        await loadFixture(deploySingleModel);
      const updates = {
        ipfsCID: getHex(Buffer.from("ipfs://new-ipfsaddress")),
        fee: expected.fee * 2n,
        addStake: expected.stake * 2n,
        owner: provider.account.address,
        name: "Llama 3.0",
        tags: ["llama", "smart", "angry"],
      };
      await tokenMOR.write.approve([modelRegistry.address, updates.addStake], {
        account: owner.account,
      });

      const txHash = await modelRegistry.write.register([
        expected.modelId,
        updates.ipfsCID,
        updates.fee,
        updates.addStake,
        updates.owner,
        updates.name,
        updates.tags,
      ]);
      const timestamp = await getTxTimestamp(publicClient, txHash);
      const providerData = await modelRegistry.read.map([expected.modelId]);

      expect(providerData).deep.equal([
        updates.ipfsCID,
        updates.fee,
        expected.stake + updates.addStake,
        getAddress(updates.owner),
        updates.name,
        // expected.tags,
        timestamp,
        expected.isDeleted,
      ]);
    });

    it("Should emit event on update", async function () {
      const { modelRegistry, provider, tokenMOR, expected, owner } = await loadFixture(
        deploySingleModel
      );
      const updates = {
        ipfsCID: getHex(Buffer.from("ipfs://new-ipfsaddress")),
        fee: expected.fee * 2n,
        addStake: expected.stake * 2n,
        owner: provider.account.address,
        name: "Llama 3.0",
        tags: ["llama", "smart", "angry"],
      };

      await tokenMOR.write.approve([modelRegistry.address, updates.addStake]);
      await modelRegistry.write.register([
        expected.modelId,
        updates.ipfsCID,
        updates.fee,
        updates.addStake,
        updates.owner,
        updates.name,
        updates.tags,
      ]);

      const events = await modelRegistry.getEvents.RegisteredUpdated({
        modelId: expected.modelId,
        owner: getAddress(provider.account.address),
      });
      expect(events.length).eq(1);
    });
  });

  describe("Getters", function () {
    it("Should get by index", async function () {
      const { modelRegistry, provider, expected } = await loadFixture(deploySingleModel);
      const [modelId, providerData] = await modelRegistry.read.getByIndex([0n]);

      expect(modelId).eq(expected.modelId);
      expect(providerData).deep.equal({
        ipfsCID: expected.ipfsCID,
        fee: expected.fee,
        stake: expected.stake,
        owner: getAddress(expected.owner),
        name: expected.name,
        tags: expected.tags,
        timestamp: expected.timestamp,
        isDeleted: expected.isDeleted,
      });
    });

    it("Should get by address", async function () {
      const { modelRegistry, provider, expected } = await loadFixture(deploySingleModel);

      const providerData = await modelRegistry.read.map([expected.modelId]);
      expect(providerData).deep.equal([
        expected.ipfsCID,
        expected.fee,
        expected.stake,
        getAddress(expected.owner),
        expected.name,
        // expected.tags,
        expected.timestamp,
        expected.isDeleted,
      ]);
    });
  });

  describe("Min stake", function () {
    it("Should set min stake", async function () {
      const { modelRegistry, owner } = await loadFixture(deployModelRegistry);
      const minStake = 100n;

      await modelRegistry.write.setMinStake([minStake], { account: owner.account });
      const events = await modelRegistry.getEvents.MinStakeUpdated();
      expect(await modelRegistry.read.minStake()).eq(minStake);
      expect(events.length).eq(1);
      expect(events[0].args.newStake).eq(minStake);
    });

    it("Should error when not owner is setting min stake", async function () {
      const { modelRegistry, provider } = await loadFixture(deploySingleModel);

      await expect(
        modelRegistry.write.setMinStake([100n], { account: provider.account })
      ).to.be.rejectedWith("Ownable: caller is not the owner");
    });
  });
});
