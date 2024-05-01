import hre from "hardhat";
import { encodeFunctionData, encodePacked, getAddress, parseUnits } from "viem/utils";
import { getHex, getTxTimestamp, randomBytes32 } from "./utils";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { FacetCutAction, getSelectors } from "../libraries/diamond";

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

export async function deployDiamond() {
  // deploy provider registry and deps
  const { tokenMOR, owner, decimalsMOR } = await loadFixture(deployMORtoken);

  const [, provider, user] = await hre.viem.getWalletClients();
  const publicClient = await hre.viem.getPublicClient();

  // 1. deploy diamont init
  const diamondInit = await hre.viem.deployContract("DiamondInit", [], {});

  // 2. deploy faucets
  const FacetNames = [
    "DiamondCutFacet",
    "DiamondLoupeFacet",
    "OwnershipFacet",
    "ModelRegistry",
    "ProviderRegistry",
    "Marketplace",
    "SessionRouter",
  ] as const;

  const facetContracts = await Promise.all(
    FacetNames.map(async (name) => {
      try {
        return await hre.viem.deployContract(name as any, [], {});
      } catch (e) {
        console.log(`error deploying ${name}`);
        throw e;
      }
    }),
  );

  // 3. deploy diamond
  const facetCuts = facetContracts.map((facetContract) => ({
    facetAddress: facetContract.address,
    action: FacetCutAction.Add,
    functionSelectors: getSelectors(facetContract.abi),
  }));

  const diamondArgs = {
    owner: owner.account.address,
    init: diamondInit.address,
    initCalldata: encodeFunctionData({
      abi: diamondInit.abi,
      functionName: "init",
      args: [tokenMOR.address, owner.account.address],
    }),
  };
  const diamond = await hre.viem.deployContract("Diamond", [facetCuts, diamondArgs]);

  const modelRegistry = await hre.viem.getContractAt(
    "contracts/facets/ModelRegistry.sol:ModelRegistry",
    diamond.address
  );
  const providerRegistry = await hre.viem.getContractAt(
    "contracts/facets/ProviderRegistry.sol:ProviderRegistry",
    diamond.address
  );
  const marketplace = await hre.viem.getContractAt(
    "contracts/facets/Marketplace.sol:Marketplace",
    diamond.address
  );
  const sessionRouter = await hre.viem.getContractAt(
    "contracts/facets/SessionRouter.sol:SessionRouter",
    diamond.address
  );

  return {
    tokenMOR,
    decimalsMOR,
    diamond,
    owner,
    user,
    provider,
    publicClient,
    modelRegistry,
    providerRegistry,
    marketplace,
    sessionRouter,
  };
}

export async function deploySingleProvider() {
  const {
    sessionRouter,
    providerRegistry,
    owner,
    provider,
    publicClient,
    decimalsMOR,
    tokenMOR,
    modelRegistry,
    user,
    marketplace,
  } = await loadFixture(deployDiamond);

  const expectedProvider = {
    address: getAddress(provider.account.address),
    stake: parseUnits("100", decimalsMOR),
    endpoint: "https://bestprovider.com",
    timestamp: 0n,
    isDeleted: false,
  };

  await tokenMOR.write.transfer(
    [provider.account.address, expectedProvider.stake * 100n],
    {
      account: owner.account,
    },
  );

  await tokenMOR.write.approve(
    [sessionRouter.address, expectedProvider.stake],
    {
      account: provider.account,
    },
  );

  const addProviderHash = await providerRegistry.write.providerRegister(
    [
      expectedProvider.address,
      expectedProvider.stake,
      expectedProvider.endpoint,
    ],
    { account: provider.account },
  );
  expectedProvider.timestamp = await getTxTimestamp(
    publicClient,
    addProviderHash,
  );

  return {
    expectedProvider,
    providerRegistry,
    modelRegistry,
    sessionRouter,
    marketplace,
    owner,
    provider,
    user,
    publicClient,
    tokenMOR,
    decimalsMOR,
  };
}

export async function deploySingleModel() {
  const { owner, provider, publicClient, tokenMOR, modelRegistry } =
    await loadFixture(deployDiamond);

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

  const addProviderHash = await modelRegistry.write.modelRegister([
    expectedModel.modelId,
    expectedModel.ipfsCID,
    expectedModel.fee,
    expectedModel.stake,
    expectedModel.owner,
    expectedModel.name,
    expectedModel.tags,
  ]);

  expectedModel.timestamp = await getTxTimestamp(publicClient, addProviderHash);

  return {
    expectedModel,
    modelRegistry,
    owner,
    provider,
    publicClient,
    tokenMOR,
  };
}
export async function deploySingleBid() {
  const {
    owner,
    provider,
    publicClient,
    tokenMOR,
    modelRegistry,
    user,
    decimalsMOR,
    marketplace,
    sessionRouter,
    expectedProvider,
  } = await loadFixture(deploySingleProvider);

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
  const addProviderHash = await modelRegistry.write.modelRegister([
    expectedModel.modelId,
    expectedModel.ipfsCID,
    expectedModel.fee,
    expectedModel.stake,
    expectedModel.owner,
    expectedModel.name,
    expectedModel.tags,
  ]);

  expectedModel.timestamp = await getTxTimestamp(publicClient, addProviderHash);

  await tokenMOR.write.approve([modelRegistry.address, 10000n * 10n ** 18n]);
  await tokenMOR.write.transfer([user.account.address, 100000n * 10n ** 18n]);
  await tokenMOR.write.approve(
    [modelRegistry.address, 10000000n * 10n ** 18n],
    {
      account: user.account,
    },
  );

  // expected bid
  const expectedBid = {
    id: "" as `0x${string}`,
    providerAddr: getAddress(expectedProvider.address),
    modelId: expectedModel.modelId,
    pricePerSecond: 1n,
    nonce: 0n,
    createdAt: 0n,
    deletedAt: 0n,
  };

  // add single bid
  const postBidtx = await marketplace.simulate.postModelBid(
    [expectedBid.providerAddr, expectedBid.modelId, expectedBid.pricePerSecond],
    { account: provider.account.address },
  );
  const txHash = await provider.writeContract(postBidtx.request);

  expectedBid.id = postBidtx.result;
  expectedBid.createdAt = await getTxTimestamp(publicClient, txHash);

  return {
    expectedBid,
    marketplace,
    owner,
    provider,
    publicClient,
    tokenMOR,
    decimalsMOR,
    sessionRouter,
    user,
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
