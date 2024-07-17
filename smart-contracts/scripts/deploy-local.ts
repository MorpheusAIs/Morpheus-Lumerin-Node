import * as fixtures from "../test/fixtures";

async function main() {
  const data = await fixtures.deploySingleBid();
  console.log(`
    MOR token       ${data.tokenMOR.address}
    Diamond         ${data.marketplace.address}

    Owner:          ${data.owner.account.address}
    Provider:       ${data.provider.account.address}
    User:           ${data.user.account.address}
  `);
}

main();
