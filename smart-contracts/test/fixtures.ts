import hre from "hardhat";
import { getAddress, padBytes, padHex, parseUnits } from "viem/utils";
import { getHex, getTxTimestamp } from "./utils";
import crypto from "crypto";

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
  const { tokenMOR, decimalsMOR } = await deployMORtoken();

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
    await deployProviderRegistry();
  const expected = {
    address: provider.account.address,
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
  };
}

export async function deployModelRegistry() {
  const [owner, provider] = await hre.viem.getWalletClients();
  const { tokenMOR, decimalsMOR } = await deployMORtoken();

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
  const { modelRegistry, owner, provider, publicClient, tokenMOR } = await deployModelRegistry();
  const bytes = crypto.randomBytes(32);
  const modelId = getHex(bytes);
  const expected = {
    modelId,
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
