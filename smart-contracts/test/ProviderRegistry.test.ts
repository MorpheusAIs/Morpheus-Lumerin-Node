import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { getAddress } from "viem";
import { deployProviderRegistry, deploySingleProvider } from "./fixtures";
import { expectError, getTxTimestamp } from "./utils";

describe("Provider registry", function () {
  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      const { providerRegistry, owner } = await loadFixture(deployProviderRegistry);

      expect(await providerRegistry.read.owner()).to.equal(getAddress(owner.account.address));
    });

    it("Should set the right token", async function () {
      const { providerRegistry, tokenMOR } = await loadFixture(deployProviderRegistry);

      expect(await providerRegistry.read.token()).to.equal(getAddress(tokenMOR.address));
    });

    it("Should count correctly", async function () {
      const { providerRegistry } = await loadFixture(deployProviderRegistry);

      expect(await providerRegistry.read.getCount()).eq(0n);
    });
  });

  describe("Actions", function () {
    it("Should register", async function () {
      const { providerRegistry, provider, expected } = await loadFixture(deploySingleProvider);
      const data = await providerRegistry.read.map([provider.account.address]);
      const events = await providerRegistry.getEvents.RegisteredUpdated({
        provider: getAddress(provider.account.address),
      });

      expect(await providerRegistry.read.getCount()).eq(1n);
      expect(await providerRegistry.read.providers([0n])).eq(getAddress(provider.account.address));
      expect(data).deep.equal([expected.endpoint, expected.stake, expected.timestamp, false]);
      expect(events.length).eq(1);
    });

    it("Should error when registering with insufficient stake", async function () {
      const { providerRegistry, provider } = await loadFixture(deployProviderRegistry);
      const minStake = 100n;
      await providerRegistry.write.setMinStake([minStake]);
      try {
        await providerRegistry.simulate.register([
          provider.account.address,
          minStake - 1n,
          "endpoint",
        ]);
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, providerRegistry.abi, "StakeTooLow");
      }
    });

    it("Should error when registering with insufficient allowance", async function () {
      const { providerRegistry, provider, tokenMOR } = await loadFixture(deployProviderRegistry);
      try {
        await providerRegistry.simulate.register([provider.account.address, 100n, "endpoint"]);
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, tokenMOR.abi, "ERC20InsufficientAllowance");
      }
    });

    it("Should error when register account doesnt match sender account", async function () {
      const [, , user] = await hre.viem.getWalletClients();

      const { providerRegistry, provider, tokenMOR, owner } = await loadFixture(
        deployProviderRegistry
      );
      try {
        await providerRegistry.simulate.register([user.account.address, 100n, "endpoint"], {
          account: provider.account.address,
        });
        expect.fail("Expected error");
      } catch (e) {
        expectError(e, providerRegistry.abi, "NotSenderOrOwner");
      }
    });

    it("Should deregister by provider", async function () {
      const { providerRegistry, provider, expected } = await loadFixture(deploySingleProvider);

      await providerRegistry.write.deregister([provider.account.address], {
        account: provider.account,
      });
      const events = await providerRegistry.getEvents.Deregistered({
        provider: getAddress(provider.account.address),
      });

      expect(await providerRegistry.read.getCount()).eq(0n);
      expect((await providerRegistry.read.map([provider.account.address]))[3]).to.equal(true);
      expect(events.length).eq(1);
      await expect(providerRegistry.read.getByIndex([0n])).rejectedWith(
        /.*reverted with panic code 0x32 (Array accessed at an out-of-bounds or negative index)*/
      );
      expect(await providerRegistry.read.providers([0n])).equals(
        getAddress(provider.account.address)
      );
    });

    it("Should deregister by admin", async function () {
      const { providerRegistry, provider, owner } = await loadFixture(deploySingleProvider);

      await providerRegistry.write.deregister([provider.account.address], {
        account: owner.account,
      });
      const events = await providerRegistry.getEvents.Deregistered({
        provider: getAddress(provider.account.address),
      });

      expect(await providerRegistry.read.getCount()).eq(0n);
      expect(events.length).eq(1);
    });

    it("Should return stake on deregister", async function () {
      const { providerRegistry, provider, tokenMOR, expected } = await loadFixture(
        deploySingleProvider
      );

      const balanceBefore = await tokenMOR.read.balanceOf([provider.account.address]);
      await providerRegistry.write.deregister([provider.account.address]);
      const balanceAfter = await tokenMOR.read.balanceOf([provider.account.address]);

      expect(balanceAfter - balanceBefore).eq(expected.stake);
    });

    it("Should update stake and url", async function () {
      const { providerRegistry, provider, tokenMOR, expected, publicClient } = await loadFixture(
        deploySingleProvider
      );
      const updates = {
        addStake: expected.stake * 2n,
        endpoint: "new-endpoint",
      };
      await tokenMOR.write.approve([providerRegistry.address, updates.addStake], {
        account: provider.account,
      });

      const txHash = await providerRegistry.write.register(
        [provider.account.address, updates.addStake, updates.endpoint],
        { account: provider.account }
      );
      const timestamp = await getTxTimestamp(publicClient, txHash);
      const providerData = await providerRegistry.read.map([provider.account.address]);

      expect(providerData).deep.equal([
        updates.endpoint,
        expected.stake + updates.addStake,
        timestamp,
        expected.isDeleted,
      ]);
    });

    it("Should emit event on update", async function () {
      const { providerRegistry, provider, tokenMOR, expected } = await loadFixture(
        deploySingleProvider
      );
      const updates = {
        addStake: expected.stake * 2n,
        endpoint: "new-endpoint",
      };

      await tokenMOR.write.approve([providerRegistry.address, updates.addStake], {
        account: provider.account,
      });
      await providerRegistry.write.register(
        [provider.account.address, updates.addStake, updates.endpoint],
        { account: provider.account }
      );

      const events = await providerRegistry.getEvents.RegisteredUpdated({
        provider: getAddress(provider.account.address),
      });
      expect(events.length).eq(1);
    });
  });

  describe("Getters", function () {
    it("Should get by index", async function () {
      const { providerRegistry, provider, expected } = await loadFixture(deploySingleProvider);
      const [address, providerData] = await providerRegistry.read.getByIndex([0n]);

      expect(address).eq(getAddress(provider.account.address));
      expect(providerData).deep.equal({
        endpoint: expected.endpoint,
        stake: expected.stake,
        timestamp: expected.timestamp,
        isDeleted: expected.isDeleted,
      });
    });

    it("Should get by address", async function () {
      const { providerRegistry, provider, expected } = await loadFixture(deploySingleProvider);

      const providerData = await providerRegistry.read.map([provider.account.address]);
      expect(providerData).deep.equal([
        expected.endpoint,
        expected.stake,
        expected.timestamp,
        expected.isDeleted,
      ]);
    });
  });

  describe("Views", function () {
    it("should get all", async function () {
      const { providerRegistry, expected } = await loadFixture(deploySingleProvider);
      const [ids, providers] = await providerRegistry.read.getAll();

      expect(providers.length).eq(1);
      expect(ids.length).eq(1);
      expect(ids[0]).eq(expected.address);
      expect(providers[0]).deep.equal({
        endpoint: expected.endpoint,
        stake: expected.stake,
        timestamp: expected.timestamp,
        isDeleted: expected.isDeleted,
      });
    });
  });

  describe("Min stake", function () {
    it("Should set min stake", async function () {
      const { providerRegistry, owner } = await loadFixture(deployProviderRegistry);
      const minStake = 100n;

      await providerRegistry.write.setMinStake([minStake], { account: owner.account });
      const events = await providerRegistry.getEvents.MinStakeUpdated();
      expect(await providerRegistry.read.minStake()).eq(minStake);
      expect(events.length).eq(1);
      expect(events[0].args.newStake).eq(minStake);
    });

    it("Should error when not owner is setting min stake", async function () {
      const { providerRegistry, provider } = await loadFixture(deploySingleProvider);

      await expect(
        providerRegistry.write.setMinStake([100n], { account: provider.account })
      ).to.be.rejectedWith("Ownable: caller is not the owner");
    });
  });
});
