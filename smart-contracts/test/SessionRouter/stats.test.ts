import hre from "hardhat";
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import {
  deploySingleBid,
  getProviderApproval,
  getReport,
  getStake,
} from "../fixtures";
import { ArtifactsMap } from "hardhat/types/artifacts";
import { GetContractReturnType } from "@nomicfoundation/hardhat-viem/types";
import { getSessionId } from "../utils";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { HOUR, SECOND } from "../../utils/time";
import { expect } from "chai";

describe("Session router - stats tests", function () {
  it("should update provider-model stats", async function () {
    const {
      sessionRouter,
      expectedSession: exp,
      expectedBid,
      marketplace,
      tokenMOR,
    } = await loadFixture(deploySingleBid);

    await openCloseSession(
      hre,
      tokenMOR,
      sessionRouter,
      exp.bidID,
      BigInt((1 * HOUR) / SECOND),
      exp.pricePerSecond,
      100,
      1000,
      true,
    );

    await openCloseSession(
      hre,
      tokenMOR,
      sessionRouter,
      exp.bidID,
      BigInt((1 * HOUR) / SECOND),
      exp.pricePerSecond,
      150,
      2000,
      true,
    );

    const [bidIds, bids, stats] =
      await marketplace.read.getActiveBidsRatingByModelAgent([
        exp.modelAgentId,
        0n,
        100,
      ]);

    expect(bidIds).to.deep.equal([exp.bidID]);
    expect(bids[0]).to.deep.equal({
      modelAgentId: expectedBid.modelId,
      nonce: expectedBid.nonce,
      pricePerSecond: expectedBid.pricePerSecond,
      provider: expectedBid.providerAddr,
      createdAt: expectedBid.createdAt,
      deletedAt: expectedBid.deletedAt,
    });
    expect(stats[0].successCount).to.equal(2);
    expect(stats[0].totalCount).to.equal(2);
    expect(Number(stats[0].tpsScaled1000.mean)).to.greaterThan(0);
    expect(Number(stats[0].ttftMs.mean)).to.greaterThan(0);
  });
});

async function openCloseSession(
  hre: HardhatRuntimeEnvironment,
  mor: GetContractReturnType<ArtifactsMap["MorpheusToken"]["abi"]>,
  sr: GetContractReturnType<ArtifactsMap["SessionRouter"]["abi"]>,
  bidID: `0x${string}`,
  durationSeconds: bigint,
  pricePerSecond: bigint,
  tps: number,
  ttft: number,
  success = true,
) {
  const [owner, provider, user] = await hre.viem.getWalletClients();
  const publicClient = await hre.viem.getPublicClient();

  // open session
  const { msg, signature } = await getProviderApproval(
    provider,
    user.account.address,
    bidID,
  );
  const stake = await getStake(sr, durationSeconds, pricePerSecond);

  await mor.write.transfer([user.account.address, stake], {
    account: owner.account.address,
  });
  await mor.write.approve([sr.address, stake], {
    account: user.account.address,
  });
  const openTx = await sr.write.openSession([stake, msg, signature], {
    account: user.account.address,
  });
  const sessionId = await getSessionId(publicClient, hre, openTx);

  // wait till end of the session
  await time.increase(durationSeconds);

  // close session
  const signer = success ? provider : user;
  const report = await getReport(signer, sessionId, tps, ttft);
  await sr.write.closeSession([report.msg, report.sig], {
    account: user.account,
  });
}
