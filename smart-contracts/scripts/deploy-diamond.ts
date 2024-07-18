import * as fixtures from "../test/fixtures";
import hre from "hardhat";

async function main() {
  const morAddress = process.env.MOR_TOKEN_ADDRESS;
  if (!morAddress) {
    throw new Error("MOR_TOKEN_ADDRESS is not set");
  }
  console.log("Deploying diamond...");
  console.log("MOR_TOKEN_ADDRESS: ", morAddress);

  const [owner] = await hre.viem.getWalletClients();
  const data = await fixtures.onlyDeployDiamond(morAddress, owner);
  console.log(`
    Diamond deployed to           ${data.diamond.address}
  `);

  try {
    // TODO: correct constructor arguments
    //
    // console.log("Verifying diamond...");
    // await hre.run("verify", {
    //   address: data.diamond.address,
    //   constructorArguments: data.constructorArgs,
    // });

    for (const facet of data.facets) {
      console.log(`Verifying facet ${facet.name} at address: ${facet.address}`);
      await hre.run("verify", { address: facet.address });
    }
  } catch (e) {
    console.error("Error verifying diamond: ", e);
    console.log("Facets", data.facets);
    console.log("Constructor args", data.constructorArgs);
  }
}

main();
