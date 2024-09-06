import hre from "hardhat";

async function main() {
  console.log("Deploying lumerin contract ...");
  const lmr = await hre.viem.deployContract("LumerinToken");
  console.log("Deployed to:", lmr.address);

  console.log("Verifying contract ...");
  await hre.run("verify", {
    address: lmr.address,
    constructorArgsParams: [],
  });
  console.log("Contract verified");
}

main();
