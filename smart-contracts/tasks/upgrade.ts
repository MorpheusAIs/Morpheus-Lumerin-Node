import { task } from "hardhat/config";
import { ArtifactsMap, HardhatRuntimeEnvironment } from "hardhat/types";
import {
  FacetCutAction,
  getSelectors,
  getSelectorsWithFunctions,
} from "../libraries/diamond";
import { zeroAddress, zeroHash } from "viem";

interface Args {
  facet: string;
  diamond: string;
}

function validateAddress(address: string): address is `0x${string}` {
  return address.match(/^0x[0-9a-fA-F]{40}$/) !== null;
}

function validateArtifact(
  name: string,
  hre: HardhatRuntimeEnvironment,
): name is keyof ArtifactsMap {
  try {
    const a = hre.artifacts.readArtifactSync(name);
    return true;
  } catch (e) {
    return false;
  }
}

task<Args>("upgrade", "Upgrades contract's faucet")
  .addParam("facet", "facet name")
  .addParam("diamond", "diamond address")
  .setAction(async (args: Args, hre) => {
    if (!validateAddress(args.diamond)) {
      throw new Error("Invalid diamond address");
    }

    if (!validateArtifact(args.facet, hre)) {
      throw new Error("Facet does not exist");
    }

    const facetContract = await hre.viem.deployContract(
      args.facet as string,
      [],
      {},
    );
    console.log(`Facet ${args.facet} deployed at ${facetContract.address}`);

    const fns = getSelectorsWithFunctions(facetContract.abi);

    console.log(
      `Updating function names:\n${fns.map(({ hash, signature }) => {
        return `${hash} - ${signature}\n`;
      })}\n`,
    );

    const facetCut = {
      facetAddress: facetContract.address,
      action: FacetCutAction.Replace,
      functionSelectors: fns.map((s) => s.hash),
    };

    const diamondCut = await hre.viem.getContractAt(
      "DiamondCutFacet",
      args.diamond,
    );
    const [owner] = await hre.viem.getWalletClients();
    const publicClient = await hre.viem.getPublicClient();

    const req = await diamondCut.simulate.diamondCut([
      [facetCut],
      zeroAddress,
      "0x",
    ]);
    const hash = await owner.writeContract(req.request);
    await publicClient.waitForTransactionReceipt({ hash });

    console.log("Diamond upgraded successfully!");
    console.log("Txhash:", hash);
  });
