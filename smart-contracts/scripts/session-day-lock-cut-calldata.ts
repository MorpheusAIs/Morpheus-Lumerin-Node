import { Fragment, Interface } from 'ethers';
import { ethers } from 'hardhat';

import { parseConfig } from '../deploy/helpers/config-parser';

import {
  ISessionRouter__factory,
  IStatsStorage__factory,
  LumerinDiamond__factory,
} from '@/generated-types/ethers';
import { FacetAction } from '@/test/helpers/deployers';

/**
 * Step 2 of the two-step SessionRouter day-lock upgrade (#830): builds the
 * exact `diamondCut` the diamond owner must sign, from a fresh loupe read.
 * Read-only; sends nothing.
 *
 * Usage (after migration 5):
 *   NEW_SESSION_ROUTER_FACET=0x... \
 *     npx hardhat run scripts/session-day-lock-cut-calldata.ts --network base_sepolia
 *
 * Env:
 *   NEW_SESSION_ROUTER_FACET  (required) SessionRouter facet from migration 5
 *   DIAMOND                   (optional) defaults to config_base_sepolia.json;
 *                             MUST be set explicitly for mainnet
 *                             (0x6aBE1d282f72B474E54527D93b979A4f64d3030a)
 *   REMOVED_SELECTORS_ALLOWLIST (optional) comma-separated selectors this cut
 *                             intentionally drops; anything else aborts
 *
 * Same ceremony on Sepolia and mainnet — see
 * smart-contracts/docs/session-day-lock-upgrade-runbook.md
 */

function requireAddress(name: string): string {
  const value = process.env[name];
  if (!value || !ethers.isAddress(value)) {
    throw new Error(`Set ${name} to a valid address (got: ${value ?? 'unset'})`);
  }
  return ethers.getAddress(value);
}

function interfaceSelectors(iface: Interface): string[] {
  return iface.fragments.filter(Fragment.isFunction).map((f) => f.selector);
}

function selectorNames(): Map<string, string> {
  const names = new Map<string, string>();
  for (const [label, iface] of [
    ['ISessionRouter', ISessionRouter__factory.createInterface()],
    ['IStatsStorage', IStatsStorage__factory.createInterface()],
  ] as const) {
    for (const fragment of iface.fragments.filter(Fragment.isFunction)) {
      names.set(fragment.selector, `${label}.${fragment.format('sighash')}`);
    }
  }
  return names;
}

async function main() {
  const network = await ethers.provider.getNetwork();
  const diamondAddr = process.env.DIAMOND
    ? ethers.getAddress(process.env.DIAMOND)
    : ethers.getAddress(parseConfig().lumerinProtocol);
  const newSessionRouterFacet = requireAddress('NEW_SESSION_ROUTER_FACET');

  if ((await ethers.provider.getCode(newSessionRouterFacet)) === '0x') {
    throw new Error(
      `NEW_SESSION_ROUTER_FACET ${newSessionRouterFacet} has no code on chain ${network.chainId} — run migration 5 first`,
    );
  }

  const diamond = LumerinDiamond__factory.connect(diamondAddr, ethers.provider);
  const owner = await diamond.owner();
  const ownerCode = await ethers.provider.getCode(owner);

  const closeSessionSelector = ISessionRouter__factory.createInterface().getFunction('closeSession').selector;
  const oldSessionRouterFacet = await diamond.facetAddress(closeSessionSelector);
  if (oldSessionRouterFacet === ethers.ZeroAddress) {
    throw new Error('closeSession selector not registered — is DIAMOND pointing at the right contract?');
  }
  const oldSelectors = [...(await diamond.facetFunctionSelectors(oldSessionRouterFacet))];

  const sessionRouterSelectors = interfaceSelectors(ISessionRouter__factory.createInterface());
  const statsSelectors = interfaceSelectors(IStatsStorage__factory.createInterface());

  for (const selector of [...sessionRouterSelectors, ...statsSelectors]) {
    const current = await diamond.facetAddress(selector);
    if (current !== ethers.ZeroAddress && current !== oldSessionRouterFacet) {
      throw new Error(`Selector ${selector} is already served by unrelated facet ${current} — aborting`);
    }
  }

  const allowedDrops = new Set(
    (process.env.REMOVED_SELECTORS_ALLOWLIST ?? '')
      .split(',')
      .map((s) => s.trim().toLowerCase())
      .filter(Boolean),
  );
  const reAdded = new Set([...sessionRouterSelectors, ...statsSelectors].map((s) => s.toLowerCase()));
  const dropped = oldSelectors.filter((s) => !reAdded.has(s.toLowerCase()));
  const unexpectedDrops = dropped.filter((s) => !allowedDrops.has(s.toLowerCase()));
  if (unexpectedDrops.length > 0) {
    throw new Error(
      `This cut would drop ${unexpectedDrops.length} selector(s) the live facet serves: ` +
        `${unexpectedDrops.join(', ')}. If intentional, list them in REMOVED_SELECTORS_ALLOWLIST.`,
    );
  }

  const registeredFacets = new Set((await diamond.facetAddresses()).map((a: string) => a.toLowerCase()));
  if (registeredFacets.has(newSessionRouterFacet.toLowerCase())) {
    throw new Error(
      `NEW_SESSION_ROUTER_FACET ${newSessionRouterFacet} is already registered — upgrades are ` +
        `forward-only (fresh bytecode via migration 5); re-adding an old facet is prohibited.`,
    );
  }

  const cuts = [
    {
      facetAddress: oldSessionRouterFacet,
      action: FacetAction.Remove,
      functionSelectors: oldSelectors,
    },
    {
      facetAddress: newSessionRouterFacet,
      action: FacetAction.Add,
      functionSelectors: sessionRouterSelectors,
    },
    {
      facetAddress: newSessionRouterFacet,
      action: FacetAction.Add,
      functionSelectors: statsSelectors,
    },
  ];

  const calldata = LumerinDiamond__factory.createInterface().encodeFunctionData(
    'diamondCut((address,uint8,bytes4[])[])',
    [cuts],
  );

  const names = selectorNames();
  const describe = (selector: string) => names.get(selector) ?? '(not in current ABI — removed function)';
  const actionLabel = ['Add', 'Replace', 'Remove'];

  console.log('=== diamondCut payload: SessionRouter stipend day-lock (#830) ===');
  console.log(`network:            ${network.name} (chainId ${network.chainId})`);
  console.log(`diamond:            ${diamondAddr}`);
  console.log(
    `diamond owner:      ${owner} (${ownerCode === '0x' ? 'EOA — signs directly' : 'contract — likely a Safe, propose there'})`,
  );
  console.log(`old SessionRouter:  ${oldSessionRouterFacet} (${oldSelectors.length} selectors removed)`);
  console.log(`new SessionRouter:  ${newSessionRouterFacet}`);
  console.log('');

  for (const cut of cuts) {
    console.log(`--- ${actionLabel[cut.action]} @ ${cut.facetAddress} (${cut.functionSelectors.length} selectors) ---`);
    for (const selector of cut.functionSelectors) {
      console.log(`  ${selector}  ${describe(selector)}`);
    }
  }

  const brandNew = [...sessionRouterSelectors, ...statsSelectors].filter((s) => !oldSelectors.includes(s));
  console.log('');
  console.log(
    `selectors carried over unchanged: ${sessionRouterSelectors.length + statsSelectors.length - brandNew.length}`,
  );
  console.log(`brand-new selectors: ${brandNew.map((s) => `${s} ${describe(s)}`).join(', ') || '(none — ABI unchanged)'}`);

  console.log('');
  console.log('=== transaction for the diamond owner (single atomic tx) ===');
  console.log(`to:       ${diamondAddr}`);
  console.log(`value:    0`);
  console.log(`function: diamondCut((address,uint8,bytes4[])[])`);
  console.log(`data:     ${calldata}`);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
