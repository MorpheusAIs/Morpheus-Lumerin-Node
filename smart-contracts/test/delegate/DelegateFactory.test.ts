import {
  DelegateFactory,
  LumerinDiamond,
  MorpheusToken,
  ProvidersDelegate,
  ProvidersDelegate__factory,
  UUPSMock,
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { wei } from '@/scripts/utils/utils';
import {
  deployDelegateFactory,
  deployFacetMarketplace,
  deployFacetProviderRegistry,
  deployLumerinDiamond,
  deployMORToken,
} from '@/test/helpers/deployers';
import { Reverter } from '@/test/helpers/reverter';

describe('DelegateFactory', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let KYLE: SignerWithAddress;
  let SHEV: SignerWithAddress;

  let diamond: LumerinDiamond;
  let delegatorFactory: DelegateFactory;

  let token: MorpheusToken;

  before(async () => {
    [OWNER, KYLE, SHEV] = await ethers.getSigners();

    [diamond, token] = await Promise.all([deployLumerinDiamond(), deployMORToken()]);
    await Promise.all([
      deployFacetProviderRegistry(diamond),
      deployFacetMarketplace(diamond, token, wei(0.0001), wei(900)),
    ]);

    delegatorFactory = await deployDelegateFactory(diamond, 3600);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('UUPS', () => {
    describe('#DelegateFactory_init', () => {
      it('should revert if try to call init function twice', async () => {
        await expect(delegatorFactory.DelegateFactory_init(OWNER, OWNER, 3600)).to.be.rejectedWith(
          'Initializable: contract is already initialized',
        );
      });
    });
    describe('#version', () => {
      it('should return correct version', async () => {
        expect(await delegatorFactory.version()).to.eq(1);
      });
    });
    describe('#upgradeTo', () => {
      it('should upgrade to the new version', async () => {
        const factory = await ethers.getContractFactory('UUPSMock');
        const newImpl = await factory.deploy();

        await delegatorFactory.upgradeTo(newImpl);
        const newDelegatorFactory = newImpl.attach(delegatorFactory) as UUPSMock;

        expect(await newDelegatorFactory.version()).to.eq(999);
      });
      it('should throw error when caller is not an owner', async () => {
        await expect(delegatorFactory.connect(KYLE).upgradeTo(KYLE)).to.be.revertedWith(
          'Ownable: caller is not the owner',
        );
      });
    });
  });

  describe('#setMinDeregistrationTimeout', () => {
    it('should set the provider name', async () => {
      expect(await delegatorFactory.minDeregistrationTimeout()).eq(3600);

      await delegatorFactory.setMinDeregistrationTimeout(100);

      expect(await delegatorFactory.minDeregistrationTimeout()).eq(100);
    });
    it('should throw error when name is zero', async () => {
      await expect(delegatorFactory.connect(SHEV).setMinDeregistrationTimeout(100)).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
  });

  describe('#deployProxy', () => {
    let providersDelegateFactory: ProvidersDelegate__factory;

    before(async () => {
      providersDelegateFactory = await ethers.getContractFactory('ProvidersDelegate');
    });

    it('should deploy a new proxy', async () => {
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);

      const proxy = providersDelegateFactory.attach(await delegatorFactory.proxies(SHEV, 0)) as ProvidersDelegate;

      expect(await proxy.owner()).to.eq(SHEV);
      expect(await proxy.fee()).to.eq(wei(0.1, 25));
      expect(await proxy.feeTreasury()).to.eq(KYLE);
      expect(await proxy.name()).to.eq('name');
      expect(await proxy.endpoint()).to.eq('endpoint');
      expect(await proxy.deregistrationOpenAt()).to.eq(9998887771);
    });
    it('should deploy new proxies', async () => {
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name1', 'endpoint1', 9998887771);
      await delegatorFactory.connect(SHEV).deployProxy(SHEV, wei(0.2, 25), 'name2', 'endpoint2', 9998887772);
      await delegatorFactory.connect(KYLE).deployProxy(SHEV, wei(0.3, 25), 'name3', 'endpoint3', 9998887773);

      let proxy = providersDelegateFactory.attach(await delegatorFactory.proxies(SHEV, 1)) as ProvidersDelegate;
      expect(await proxy.owner()).to.eq(SHEV);
      expect(await proxy.fee()).to.eq(wei(0.2, 25));
      expect(await proxy.feeTreasury()).to.eq(SHEV);
      expect(await proxy.name()).to.eq('name2');
      expect(await proxy.endpoint()).to.eq('endpoint2');
      expect(await proxy.deregistrationOpenAt()).to.eq(9998887772);

      proxy = providersDelegateFactory.attach(await delegatorFactory.proxies(KYLE, 0)) as ProvidersDelegate;
      expect(await proxy.owner()).to.eq(KYLE);
      expect(await proxy.fee()).to.eq(wei(0.3, 25));
      expect(await proxy.feeTreasury()).to.eq(SHEV);
      expect(await proxy.name()).to.eq('name3');
      expect(await proxy.endpoint()).to.eq('endpoint3');
      expect(await proxy.deregistrationOpenAt()).to.eq(9998887773);
    });
    it('should throw error when fee is invalid', async () => {
      await expect(
        delegatorFactory.deployProxy(KYLE, wei(0.1, 25), 'name1', 'endpoint1', 1),
      ).to.be.revertedWithCustomError(delegatorFactory, 'InvalidDeregistrationOpenAt');
    });
  });

  describe('#pause, #unpause', () => {
    it('should revert when paused and not after the unpause', async () => {
      await delegatorFactory.pause();
      await expect(
        delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name1', 'endpoint1', 9998887771),
      ).to.be.rejectedWith('Pausable: paused');

      await delegatorFactory.unpause();
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name1', 'endpoint1', 9998887771);
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(delegatorFactory.connect(KYLE).pause()).to.be.revertedWith('Ownable: caller is not the owner');
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(delegatorFactory.connect(KYLE).unpause()).to.be.revertedWith('Ownable: caller is not the owner');
    });
  });

  describe('#predictProxyAddress', () => {
    it('should predict a proxy address', async () => {
      const predictedProxyAddress = await delegatorFactory.predictProxyAddress(SHEV);
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);

      const proxyAddress = await delegatorFactory.proxies(SHEV, 0);

      expect(proxyAddress).to.eq(predictedProxyAddress);
    });
    it('should predict proxy addresses', async () => {
      let predictedProxyAddress = await delegatorFactory.predictProxyAddress(SHEV);
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);
      expect(await delegatorFactory.proxies(SHEV, 0)).to.eq(predictedProxyAddress);

      predictedProxyAddress = await delegatorFactory.predictProxyAddress(SHEV);
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);
      expect(await delegatorFactory.proxies(SHEV, 1)).to.eq(predictedProxyAddress);

      predictedProxyAddress = await delegatorFactory.predictProxyAddress(KYLE);
      await delegatorFactory.connect(KYLE).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);
      expect(await delegatorFactory.proxies(KYLE, 0)).to.eq(predictedProxyAddress);

      predictedProxyAddress = await delegatorFactory.predictProxyAddress(SHEV);
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);
      expect(await delegatorFactory.proxies(SHEV, 2)).to.eq(predictedProxyAddress);
    });
  });

  describe('#updateImplementation', () => {
    it('should update proxies implementation', async () => {
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);
      await delegatorFactory.connect(SHEV).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);
      await delegatorFactory.connect(KYLE).deployProxy(KYLE, wei(0.1, 25), 'name', 'endpoint', 9998887771);

      const factory = await ethers.getContractFactory('UUPSMock');
      const newImpl = await factory.deploy();

      let proxy = factory.attach(await delegatorFactory.proxies(SHEV, 0)) as ProvidersDelegate;
      expect(await proxy.version()).to.eq(1);
      proxy = factory.attach(await delegatorFactory.proxies(SHEV, 1)) as ProvidersDelegate;
      expect(await proxy.version()).to.eq(1);
      proxy = factory.attach(await delegatorFactory.proxies(KYLE, 0)) as ProvidersDelegate;
      expect(await proxy.version()).to.eq(1);

      await delegatorFactory.updateImplementation(newImpl);

      proxy = factory.attach(await delegatorFactory.proxies(SHEV, 0)) as ProvidersDelegate;
      expect(await proxy.version()).to.eq(999);
      proxy = factory.attach(await delegatorFactory.proxies(SHEV, 1)) as ProvidersDelegate;
      expect(await proxy.version()).to.eq(999);
      proxy = factory.attach(await delegatorFactory.proxies(KYLE, 0)) as ProvidersDelegate;
      expect(await proxy.version()).to.eq(999);
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(delegatorFactory.connect(KYLE).updateImplementation(KYLE)).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
  });
});

// npm run generate-types && npx hardhat test "test/delegate/DelegateFactory.test.ts"
// npx hardhat coverage --solcoverjs ./.solcover.ts --testfiles "test/delegate/DelegateFactory.test.ts"
