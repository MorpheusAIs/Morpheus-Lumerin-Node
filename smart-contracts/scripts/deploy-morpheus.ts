import hre from "hardhat";

async function main() {
  console.log("Deploying morpheus contract ...");
  const mor = await hre.viem.deployContract("MorpheusToken");
  console.log("Deployed to:", mor.address);

  console.log("Verifying morpheus contract ...");
  await hre.run("verify", {
    address: mor.address,
    constructorArgsParams: [],
  });
  console.log("Mor contract verified");
}

main();
