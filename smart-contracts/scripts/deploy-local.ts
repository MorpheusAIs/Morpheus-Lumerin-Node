import * as fixtures from "../test/fixtures";

async function main() {
  const data = await fixtures.deploySingleBid();
  console.log(`
    MOR token deployed to         ${data.tokenMOR.address}
    Diamond deployed to           ${data.marketplace.address}
  `);
}

main();
