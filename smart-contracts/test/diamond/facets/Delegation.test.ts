import { Delegation, LumerinDiamond, MorpheusToken } from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { DelegateRegistry } from '@/generated-types/ethers/contracts/mock/delegate-registry/src';
import { getHex, wei } from '@/scripts/utils/utils';
import {
  deployDelegateRegistry,
  deployFacetDelegation,
  deployFacetMarketplace,
  deployFacetModelRegistry,
  deployFacetProviderRegistry,
  deployFacetSessionRouter,
  deployLumerinDiamond,
  deployMORToken,
} from '@/test/helpers/deployers';
import { Reverter } from '@/test/helpers/reverter';

describe('Delegation', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let PROVIDER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let delegation: Delegation;

  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;

  before(async () => {
    [OWNER, PROVIDER] = await ethers.getSigners();

    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);

    [, , , , delegation] = await Promise.all([
      deployFacetProviderRegistry(diamond),
      deployFacetModelRegistry(diamond),
      deployFacetSessionRouter(diamond, OWNER),
      deployFacetMarketplace(diamond, token, wei(0.0001), wei(900)),
      deployFacetDelegation(diamond, delegateRegistry),
    ]);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('#__Delegation_init', () => {
    it('should revert if try to call init function twice', async () => {
      await expect(delegation.__Delegation_init(OWNER)).to.be.rejectedWith(
        'Initializable: contract is already initialized',
      );
    });
  });

  describe('#setRegistry', async () => {
    it('should set new registry', async () => {
      await expect(delegation.setRegistry(PROVIDER))
        .to.emit(delegation, 'DelegationRegistryUpdated')
        .withArgs(PROVIDER);

      expect(await delegation.getRegistry()).eq(PROVIDER);
    });

    it('should throw error when caller is not an owner', async () => {
      await expect(delegation.connect(PROVIDER).setRegistry(PROVIDER)).to.be.revertedWithCustomError(
        diamond,
        'OwnableUnauthorizedAccount',
      );
    });
  });

  describe('#isRightsDelegated', async () => {
    it('should return `false` when rights not delegated', async () => {
      const rights = await delegation.DELEGATION_RULES_PROVIDER();

      expect(await delegation.isRightsDelegated(OWNER, PROVIDER, rights)).eq(false);
    });
    it('should return `false` when correct rights not delegated', async () => {
      const rights = await delegation.DELEGATION_RULES_PROVIDER();
      await delegateRegistry.connect(PROVIDER).delegateContract(OWNER, delegation, rights, true);

      expect(await delegation.isRightsDelegated(OWNER, PROVIDER, await delegation.DELEGATION_RULES_SESSION())).eq(
        false,
      );
    });
    it('should return `true` when rights delegated for the specific rule', async () => {
      const rights = await delegation.DELEGATION_RULES_PROVIDER();
      await delegateRegistry.connect(PROVIDER).delegateContract(OWNER, delegation, rights, true);

      expect(await delegation.isRightsDelegated(OWNER, PROVIDER, rights)).eq(true);
    });
    it('should return `true` when rights delegated for the all rules', async () => {
      const rights = await delegation.DELEGATION_RULES_PROVIDER();
      await delegateRegistry.connect(PROVIDER).delegateContract(OWNER, delegation, getHex(Buffer.from('')), true);

      expect(await delegation.isRightsDelegated(OWNER, PROVIDER, rights)).eq(true);
    });
  });
});

// npm run generate-types && npx hardhat test "test/diamond/facets/Delegation.test.ts"
// npx hardhat coverage --solcoverjs ./.solcover.ts --testfiles "test/diamond/facets/Delegation.test.ts"
