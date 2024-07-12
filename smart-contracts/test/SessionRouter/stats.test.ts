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

describe("Session router - stats tests", function () {
  it("should update provider-model stats", async function () {
    const {
      sessionRouter,
      expectedSession: exp,
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

    const stats = await marketplace.read.getActiveBidsRatingByModelAgent([
      exp.modelAgentId,
    ]);

    console.dir(stats, { depth: 5 });
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
  const { msg, signature } = await getProviderApproval(provider, bidID);
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
