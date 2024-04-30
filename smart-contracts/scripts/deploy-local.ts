import * as fixtures from "../test/fixtures";

async function main() {
  const data = await fixtures.deployDiamond();
  console.log(`
    MOR token deployed to         ${data.tokenMOR.address}
    ProviderRegistry deployed to  ${data.providerRegistry.address}
    ModelRegistry deployed to     ${data.modelRegistry.address}
    Marketplace deployed to       ${data.marketplace.address}
    Session router deployed to    ... not yet ...
  `);
}

main();
