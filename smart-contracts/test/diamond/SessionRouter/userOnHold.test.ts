import { MAX_UINT8 } from "@/scripts/utils/constants";
import { getHex, randomBytes32, wei } from "@/scripts/utils/utils";
import { FacetAction } from "@/test/helpers/enums";
import { getDefaultPools } from "@/test/helpers/pool-helper";
import { Reverter } from "@/test/helpers/reverter";
import { getCurrentBlockTime, setTime } from "@/utils/block-helper";
import { getProviderApproval, getReport } from "@/utils/provider-helper";
import { DAY, HOUR, YEAR } from "@/utils/time";
import {
  IBidStorage,
  IMarketplace__factory,
  IModelRegistry__factory,
  IModelStorage,
  IProviderRegistry__factory,
  IProviderStorage,
  ISessionRouter__factory,
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
  ProviderRegistry,
  SessionRouter,
} from "@ethers-v6";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { expect } from "chai";
import { Addressable, Fragment } from "ethers";
import { ethers } from "hardhat";

describe("User on hold tests", () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let SECOND: SignerWithAddress;
  let THIRD: SignerWithAddress;
  let PROVIDER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let marketplace: Marketplace;
  let modelRegistry: ModelRegistry;
  let providerRegistry: ProviderRegistry;
  let sessionRouter: SessionRouter;

  let MOR: MorpheusToken;

  async function deployProvider(): Promise<
    IProviderStorage.ProviderStruct & {
      address: Addressable;
    }
  > {
    const provider = {
      endpoint: "localhost:3334",
      stake: wei(100),
      createdAt: 0n,
      limitPeriodEnd: 0n,
      limitPeriodEarned: 0n,
      isDeleted: false,
      address: PROVIDER,
    };

    await MOR.transfer(PROVIDER, provider.stake * 100n);
    await MOR.connect(PROVIDER).approve(sessionRouter, provider.stake);

    await providerRegistry
      .connect(PROVIDER)
      .providerRegister(provider.address, provider.stake, provider.endpoint);
    provider.createdAt = await getCurrentBlockTime();
    provider.limitPeriodEnd = provider.createdAt + YEAR;

    return provider;
  }

  async function deployModel(): Promise<
    IModelStorage.ModelStruct & {
      modelId: string;
    }
  > {
    const model = {
      modelId: randomBytes32(),
      ipfsCID: getHex(Buffer.from("ipfs://ipfsaddress")),
      fee: 100,
      stake: 100,
      owner: OWNER,
      name: "Llama 2.0",
      tags: ["llama", "animal", "cute"],
      createdAt: 0n,
      isDeleted: false,
    };

    await MOR.approve(modelRegistry, model.stake);

    await modelRegistry.modelRegister(
      model.modelId,
      model.ipfsCID,
      model.fee,
      model.stake,
      model.owner,
      model.name,
      model.tags,
    );
    model.createdAt = await getCurrentBlockTime();

    return model;
  }

  async function deployBid(model: any): Promise<
    [
      IBidStorage.BidStruct & {
        id: string;
        modelId: string;
      },
      {
        durationSeconds: bigint;
        totalCost: bigint;
        pricePerSecond: bigint;
        user: SignerWithAddress;
        provider: SignerWithAddress;
        modelAgentId: any;
        bidID: string;
        stake: bigint;
      },
    ]
  > {
    let bid = {
      id: "",
      modelId: model.modelId,
      pricePerSecond: wei(0.0001),
      nonce: 0,
      createdAt: 0n,
      deletedAt: 0,
      provider: PROVIDER,
      modelAgentId: model.modelId,
    };

    await MOR.approve(modelRegistry, 10000n * 10n ** 18n);

    bid.id = await marketplace
      .connect(PROVIDER)
      .postModelBid.staticCall(bid.provider, bid.modelId, bid.pricePerSecond);
    await marketplace
      .connect(PROVIDER)
      .postModelBid(bid.provider, bid.modelId, bid.pricePerSecond);

    bid.createdAt = await getCurrentBlockTime();

    // generating data for sample session
    const durationSeconds = HOUR;
    const totalCost = bid.pricePerSecond * durationSeconds;
    const totalSupply = await sessionRouter.totalMORSupply(
      await getCurrentBlockTime(),
    );
    const todaysBudget = await sessionRouter.getTodaysBudget(
      await getCurrentBlockTime(),
    );

    const expectedSession = {
      durationSeconds,
      totalCost,
      pricePerSecond: bid.pricePerSecond,
      user: SECOND,
      provider: bid.provider,
      modelAgentId: bid.modelId,
      bidID: bid.id,
      stake: (totalCost * totalSupply) / todaysBudget,
    };

    await MOR.transfer(SECOND, expectedSession.stake);
    await MOR.connect(SECOND).approve(modelRegistry, expectedSession.stake);

    return [bid, expectedSession];
  }

  async function openSession(session: any) {
    // open session
    const { msg, signature } = await getProviderApproval(
      PROVIDER,
      await SECOND.getAddress(),
      session.bidID,
    );
    const sessionId = await sessionRouter
      .connect(SECOND)
      .openSession.staticCall(session.stake, msg, signature);
    await sessionRouter
      .connect(SECOND)
      .openSession(session.stake, msg, signature);

    return sessionId;
  }

  async function openEarlyCloseSession(session: any, sessionId: string) {
    await setTime(
      Number(await getCurrentBlockTime()) + Number(session.durationSeconds) / 2,
    );

    // close session
    const report = await getReport(PROVIDER, sessionId, 10, 10);
    await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

    return session.stake / 2n;
  }
  before("setup", async () => {
    [OWNER, SECOND, THIRD, PROVIDER] = await ethers.getSigners();

    const LinearDistributionIntervalDecrease = await ethers.getContractFactory(
      "LinearDistributionIntervalDecrease",
    );
    const linearDistributionIntervalDecrease =
      await LinearDistributionIntervalDecrease.deploy();

    const [
      LumerinDiamond,
      Marketplace,
      ModelRegistry,
      ProviderRegistry,
      SessionRouter,
      MorpheusToken,
    ] = await Promise.all([
      ethers.getContractFactory("LumerinDiamond"),
      ethers.getContractFactory("Marketplace"),
      ethers.getContractFactory("ModelRegistry"),
      ethers.getContractFactory("ProviderRegistry"),
      ethers.getContractFactory("SessionRouter", {
        libraries: {
          LinearDistributionIntervalDecrease:
            linearDistributionIntervalDecrease,
        },
      }),
      ethers.getContractFactory("MorpheusToken"),
    ]);

    [
      diamond,
      marketplace,
      modelRegistry,
      providerRegistry,
      sessionRouter,
      MOR,
    ] = await Promise.all([
      LumerinDiamond.deploy(),
      Marketplace.deploy(),
      ModelRegistry.deploy(),
      ProviderRegistry.deploy(),
      SessionRouter.deploy(),
      MorpheusToken.deploy(),
    ]);

    await diamond.__LumerinDiamond_init();

    await diamond["diamondCut((address,uint8,bytes4[])[])"]([
      {
        facetAddress: marketplace,
        action: FacetAction.Add,
        functionSelectors: IMarketplace__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
      {
        facetAddress: providerRegistry,
        action: FacetAction.Add,
        functionSelectors: IProviderRegistry__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
      {
        facetAddress: sessionRouter,
        action: FacetAction.Add,
        functionSelectors: ISessionRouter__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
      {
        facetAddress: modelRegistry,
        action: FacetAction.Add,
        functionSelectors: IModelRegistry__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
    ]);

    marketplace = marketplace.attach(diamond.target) as Marketplace;
    providerRegistry = providerRegistry.attach(
      diamond.target,
    ) as ProviderRegistry;
    modelRegistry = modelRegistry.attach(diamond.target) as ModelRegistry;
    sessionRouter = sessionRouter.attach(diamond.target) as SessionRouter;

    await marketplace.__Marketplace_init(MOR);
    await modelRegistry.__ModelRegistry_init();
    await providerRegistry.__ProviderRegistry_init();

    await sessionRouter.__SessionRouter_init(OWNER, getDefaultPools());

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe("Actions", () => {
    let provider: IProviderStorage.ProviderStruct;
    let model: IModelStorage.ModelStruct & {
      modelId: string;
    };
    let bid: IBidStorage.BidStruct & {
      id: string;
      modelId: string;
    };
    let session: {
      durationSeconds: bigint;
      totalCost: bigint;
      pricePerSecond: bigint;
      user: SignerWithAddress;
      provider: SignerWithAddress;
      modelAgentId: any;
      bidID: string;
      stake: bigint;
    };
    let sessionId: string;
    let expectedOnHold: bigint;

    beforeEach(async () => {
      provider = await deployProvider();
      model = await deployModel();
      [bid, session] = await deployBid(model);
      sessionId = await openSession(session);
      expectedOnHold = await openEarlyCloseSession(session, sessionId);
    });

    it("user stake should be locked right after closeout", async () => {
      // right after closeout
      const [available, onHold] = await sessionRouter.withdrawableUserStake(
        SECOND,
        MAX_UINT8,
      );
      expect(available).to.equal(0n);
      expect(onHold).to.closeTo(
        expectedOnHold,
        BigInt(0.01 * Number(expectedOnHold)),
      );
    });

    it("user stake should be locked before the next day", async () => {
      // before next day
      await setTime(
        startOfTomorrow(Number(await getCurrentBlockTime())) - Number(HOUR),
      );
      const [available3, onHold3] = await sessionRouter.withdrawableUserStake(
        SECOND,
        Number(MAX_UINT8),
      );
      expect(available3).to.equal(0n);
      expect(onHold3).to.closeTo(
        expectedOnHold,
        BigInt(0.01 * Number(expectedOnHold)),
      );
    });

    it("user stake should be available on the next day and withdrawable", async () => {
      await setTime(startOfTomorrow(Number(await getCurrentBlockTime())));
      const [available2, onHold2] = await sessionRouter.withdrawableUserStake(
        SECOND,
        Number(MAX_UINT8),
      );
      expect(available2).to.closeTo(
        expectedOnHold,
        BigInt(0.01 * Number(expectedOnHold)),
      );
      expect(onHold2).to.equal(0n);

      const balanceBefore = await MOR.balanceOf(SECOND);
      await sessionRouter
        .connect(SECOND)
        .withdrawUserStake(available2, Number(MAX_UINT8));
      const balanceAfter = await MOR.balanceOf(SECOND);
      const balanceDelta = balanceAfter - balanceBefore;
      expect(balanceDelta).to.closeTo(
        expectedOnHold,
        BigInt(0.01 * Number(expectedOnHold)),
      );
    });

    it("user shouldn't be able to withdraw more than there is available stake", async () => {
      await setTime(startOfTomorrow(Number(await getCurrentBlockTime())));
      const [available2] = await sessionRouter.withdrawableUserStake(
        SECOND,
        Number(MAX_UINT8),
      );

      await expect(
        sessionRouter
          .connect(SECOND)
          .withdrawUserStake(available2 * 2n, Number(MAX_UINT8)),
      ).to.be.revertedWithCustomError(
        sessionRouter,
        "NotEnoughWithdrawableBalance",
      );
    });
  });
});

function startOfTomorrow(epochSeconds: number): number {
  const startOfToday = epochSeconds - (epochSeconds % Number(DAY));
  return startOfToday + Number(DAY);
}
