import hre from "hardhat";
import { encodePacked, getAddress, parseUnits } from "viem/utils";
import { getHex, getTxTimestamp, randomBytes32 } from "./utils";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";

export async function deployMORtoken() {
  const [owner] = await hre.viem.getWalletClients();

  // Contracts are deployed using the first signer/account by default
  const tokenMOR = await hre.viem.deployContract("MorpheusToken", []);
  const decimalsMOR = await tokenMOR.read.decimals();

  return {
    tokenMOR,
    decimalsMOR,
    owner,
  };
}

export async function deployProviderRegistry() {
  const [owner, provider] = await hre.viem.getWalletClients();
  const { tokenMOR, decimalsMOR } = await loadFixture(deployMORtoken);

  // Contracts are deployed using the first signer/account by default
  const providerRegistry = await hre.viem.deployContract("ProviderRegistry", [], {});
  await providerRegistry.write.initialize([tokenMOR.address]);

  const publicClient = await hre.viem.getPublicClient();

  return {
    tokenMOR,
    decimalsMOR,
    providerRegistry,
    owner,
    provider,
    publicClient,
  };
}

export async function deploySingleProvider() {
  const { providerRegistry, owner, provider, publicClient, decimalsMOR, tokenMOR } =
    await loadFixture(deployProviderRegistry);
  const expected = {
    address: getAddress(provider.account.address),
    stake: parseUnits("100", decimalsMOR),
    endpoint: "https://bestprovider.com",
    timestamp: 0n,
    isDeleted: false,
  };

  await tokenMOR.write.transfer([provider.account.address, expected.stake * 100n], {
    account: owner.account,
  });

  await tokenMOR.write.approve([providerRegistry.address, expected.stake], {
    account: provider.account,
  });

  const addProviderHash = await providerRegistry.write.register(
    [expected.address, expected.stake, expected.endpoint],
    { account: provider.account }
  );

  expected.timestamp = await getTxTimestamp(publicClient, addProviderHash);

  return {
    expected,
    providerRegistry,
    owner,
    provider,
    publicClient,
    tokenMOR,
    decimalsMOR,
  };
}

export async function deployModelRegistry() {
  const [owner, provider] = await hre.viem.getWalletClients();
  const { tokenMOR, decimalsMOR } = await loadFixture(deployMORtoken);

  // Contracts are deployed using the first signer/account by default
  const modelRegistry = await hre.viem.deployContract("ModelRegistry", [], {});
  await modelRegistry.write.initialize([tokenMOR.address]);

  const publicClient = await hre.viem.getPublicClient();

  return {
    tokenMOR,
    decimalsMOR,
    modelRegistry,
    owner,
    provider,
    publicClient,
  };
}

export async function deploySingleModel() {
  const { modelRegistry, owner, provider, publicClient, tokenMOR } = await loadFixture(
    deployModelRegistry
  );
  const expected = {
    modelId: randomBytes32(),
    ipfsCID: getHex(Buffer.from("ipfs://ipfsaddress")),
    fee: 100n,
    stake: 100n,
    owner: owner.account.address,
    name: "Llama 2.0",
    timestamp: 0n,
    tags: ["llama", "animal", "cute"],
    isDeleted: false,
  };

  await tokenMOR.write.approve([modelRegistry.address, expected.stake]);

  const addProviderHash = await modelRegistry.write.register([
    expected.modelId,
    expected.ipfsCID,
    expected.fee,
    expected.stake,
    expected.owner,
    expected.name,
    expected.tags,
  ]);

  expected.timestamp = await getTxTimestamp(publicClient, addProviderHash);

  return {
    expected,
    modelRegistry,
    owner,
    provider,
    publicClient,
    tokenMOR,
  };
}

export async function deployMarketplace() {
  // deploy provider registry and deps
  const {
    tokenMOR,
    publicClient,
    owner,
    expected: expectedProvider,
    provider,
    providerRegistry,
    decimalsMOR,
  } = await loadFixture(deploySingleProvider);

  // deploy model registry
  const modelRegistry = await hre.viem.deployContract("ModelRegistry", [], {});
  await modelRegistry.write.initialize([tokenMOR.address]);

  // add single model
  const expectedModel = {
    modelId: randomBytes32(),
    ipfsCID: getHex(Buffer.from("ipfs://ipfsaddress")),
    fee: 100n,
    stake: 100n,
    owner: owner.account.address,
    name: "Llama 2.0",
    timestamp: 0n,
    tags: ["llama", "animal", "cute"],
    isDeleted: false,
  };

  await tokenMOR.write.approve([modelRegistry.address, expectedModel.stake]);
  const addProviderHash = await modelRegistry.write.register([
    expectedModel.modelId,
    expectedModel.ipfsCID,
    expectedModel.fee,
    expectedModel.stake,
    expectedModel.owner,
    expectedModel.name,
    expectedModel.tags,
  ]);

  expectedModel.timestamp = await getTxTimestamp(publicClient, addProviderHash);

  // deploy staking
  const [, , user] = await hre.viem.getWalletClients();

  // deploy session router
  const sessionRouter = await hre.viem.deployContract("SessionRouter", [], {});
  await sessionRouter.write.initialize([
    tokenMOR.address,
    owner.account.address,
    modelRegistry.address,
    providerRegistry.address,
  ]);

  await tokenMOR.write.approve([sessionRouter.address, 10000n * 10n ** 18n]);
  await tokenMOR.write.transfer([user.account.address, 10000n * 10n ** 18n]);

  const expectedStake = {
    stakeAmount: 1000n * 10n ** 18n,
    account: user.account.address,
    transferTo: user.account.address,
    expectedStipend: 0n,
    spendAmount: 0n,
  };

  const computeBalance = await sessionRouter.read.getComputeBalance();
  const totalSupply = await tokenMOR.read.totalSupply();

  expectedStake.expectedStipend =
    ((computeBalance / 100n) * expectedStake.stakeAmount) / totalSupply;
  expectedStake.spendAmount = expectedStake.expectedStipend / 4n;

  await tokenMOR.write.approve([sessionRouter.address, 10000n * 10n ** 18n], {
    account: expectedStake.account,
  });
  await sessionRouter.write.stake([expectedStake.account, expectedStake.stakeAmount], {
    account: expectedStake.account,
  });

  expectedStake.transferTo = sessionRouter.address;

  // expected bid
  const expectedBid = {
    id: "" as `0x${string}`,
    providerAddr: getAddress(expectedProvider.address),
    modelId: expectedModel.modelId,
    amount: 10n,
    nonce: 0n,
    createdAt: 0n,
    deletedAt: 0n,
  };

  // add single bid
  const postBidtx = await sessionRouter.simulate.postModelBid(
    [expectedBid.providerAddr, expectedBid.modelId, expectedBid.amount],
    { account: provider.account.address }
  );
  const client = await hre.viem.getWalletClient(provider.account.address);
  const txHash = await client.writeContract(postBidtx.request);

  expectedBid.id = postBidtx.result;
  expectedBid.createdAt = await getTxTimestamp(publicClient, txHash);

  return {
    tokenMOR,
    decimalsMOR,
    sessionRouter,
    owner,
    user,
    provider,
    expectedBid,
    expectedModel,
    expectedProvider,
    expectedStake,
    publicClient,
    modelRegistry,
    providerRegistry,
  };
}

export const providerReport = {
  ips: 128,
  timestamp: 10000,
};
export const reportAbi = ["uint32", "uint32"] as const;
export const encodedReport = encodePacked(reportAbi, [
  providerReport.ips,
  providerReport.timestamp,
]);