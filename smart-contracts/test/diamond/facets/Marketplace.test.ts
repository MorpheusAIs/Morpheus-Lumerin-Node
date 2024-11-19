import {
  DelegateRegistry,
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
  ProviderRegistry,
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

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
import { setNextTime } from '@/utils/block-helper';

describe('Marketplace', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let SECOND: SignerWithAddress;
  let PROVIDER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let marketplace: Marketplace;
  let modelRegistry: ModelRegistry;
  let providerRegistry: ProviderRegistry;

  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;

  const baseModelId1 = getHex(Buffer.from('1'));
  const baseModelId2 = getHex(Buffer.from('2'));
  let modelId1 = getHex(Buffer.from(''));
  let modelId2 = getHex(Buffer.from(''));

  before(async () => {
    [OWNER, SECOND, PROVIDER] = await ethers.getSigners();

    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);
    [providerRegistry, modelRegistry, , marketplace] = await Promise.all([
      deployFacetProviderRegistry(diamond),
      deployFacetModelRegistry(diamond),
      deployFacetSessionRouter(diamond, OWNER),
      deployFacetMarketplace(diamond, token, wei(0.0001), wei(900)),
      deployFacetDelegation(diamond, delegateRegistry),
    ]);

    await token.transfer(SECOND, wei(1000));
    await token.connect(SECOND).approve(providerRegistry, wei(1000));
    await token.approve(providerRegistry, wei(1000));
    await token.connect(SECOND).approve(modelRegistry, wei(1000));
    await token.approve(modelRegistry, wei(1000));
    await token.connect(SECOND).approve(marketplace, wei(1000));
    await token.approve(marketplace, wei(1000));

    const ipfsCID = getHex(Buffer.from('ipfs://ipfsaddress'));
    await providerRegistry.connect(SECOND).providerRegister(SECOND, wei(100), 'test');
    await modelRegistry.connect(SECOND).modelRegister(SECOND, baseModelId1, ipfsCID, 0, wei(100), 'name', ['tag_1']);
    await modelRegistry.connect(SECOND).modelRegister(SECOND, baseModelId2, ipfsCID, 0, wei(100), 'name', ['tag_1']);

    modelId1 = await modelRegistry.getModelId(SECOND, baseModelId1);
    modelId2 = await modelRegistry.getModelId(SECOND, baseModelId2);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('#__Marketplace_init', () => {
    it('should set correct data after creation', async () => {
      expect(await marketplace.getToken()).to.eq(await token.getAddress());
    });
    it('should revert if try to call init function twice', async () => {
      await expect(marketplace.__Marketplace_init(token, wei(0.001), wei(0.002))).to.be.rejectedWith(
        'Initializable: contract is already initialized',
      );
    });
  });

  describe('#setMarketplaceBidFee', async () => {
    it('should set marketplace bid fee', async () => {
      const fee = wei(100);

      await expect(marketplace.setMarketplaceBidFee(fee)).to.emit(marketplace, 'MaretplaceFeeUpdated').withArgs(fee);

      expect(await marketplace.getBidFee()).eq(fee);
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(marketplace.connect(SECOND).setMarketplaceBidFee(100)).to.be.revertedWithCustomError(
        diamond,
        'OwnableUnauthorizedAccount',
      );
    });
  });

  describe('#setMinMaxBidPricePerSecond', async () => {
    it('should set min and max price per second', async () => {
      await expect(marketplace.setMinMaxBidPricePerSecond(wei(1), wei(2)))
        .to.emit(marketplace, 'MarketplaceBidMinMaxPriceUpdated')
        .withArgs(wei(1), wei(2));

      expect(await marketplace.getMinMaxBidPricePerSecond()).deep.eq([wei(1), wei(2)]);
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(
        marketplace.connect(SECOND).setMinMaxBidPricePerSecond(wei(1), wei(2)),
      ).to.be.revertedWithCustomError(diamond, 'OwnableUnauthorizedAccount');
    });
    it('should throw error when min price is zero', async () => {
      await expect(marketplace.setMinMaxBidPricePerSecond(wei(0), wei(2))).to.be.revertedWithCustomError(
        marketplace,
        'MarketplaceBidMinPricePerSecondIsZero',
      );
    });
    it('should throw error when min price greater then max price', async () => {
      await expect(marketplace.setMinMaxBidPricePerSecond(wei(3), wei(2))).to.be.revertedWithCustomError(
        marketplace,
        'MarketplaceBidMinPricePerSecondIsInvalid',
      );
    });
  });

  describe('#postModelBid', async () => {
    beforeEach(async () => {
      await marketplace.setMarketplaceBidFee(wei(1));
    });

    it('should post a model bid', async () => {
      await setNextTime(300);
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));

      const bidId = await marketplace.getBidId(SECOND, modelId1, 0);
      const data = await marketplace.getBid(bidId);
      expect(data.provider).to.eq(SECOND);
      expect(data.modelId).to.eq(modelId1);
      expect(data.pricePerSecond).to.eq(wei(10));
      expect(data.nonce).to.eq(0);
      expect(data.createdAt).to.eq(300);
      expect(data.deletedAt).to.eq(0);

      expect(await token.balanceOf(marketplace)).to.eq(wei(301));
      expect(await token.balanceOf(SECOND)).to.eq(wei(699));

      expect(await marketplace.getProviderBids(SECOND, 0, 10)).deep.eq([[bidId], 1n]);
      expect(await marketplace.getModelBids(modelId1, 0, 10)).deep.eq([[bidId], 1n]);
      expect(await marketplace.getProviderActiveBids(SECOND, 0, 10)).deep.eq([[bidId], 1n]);
      expect(await marketplace.getModelActiveBids(modelId1, 0, 10)).deep.eq([[bidId], 1n]);
    });
    it('should post a model bid from the delegatee address', async () => {
      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_MARKETPLACE(), true);

      await setNextTime(300);
      await marketplace.connect(OWNER).postModelBid(SECOND, modelId1, wei(10));

      const bidId = await marketplace.getBidId(SECOND, modelId1, 0);
      const data = await marketplace.getBid(bidId);
      expect(data.provider).to.eq(SECOND);
      expect(data.modelId).to.eq(modelId1);
      expect(data.pricePerSecond).to.eq(wei(10));
      expect(data.nonce).to.eq(0);
      expect(data.createdAt).to.eq(300);
      expect(data.deletedAt).to.eq(0);

      expect(await token.balanceOf(marketplace)).to.eq(wei(301));
      expect(await token.balanceOf(SECOND)).to.eq(wei(699));

      expect(await marketplace.getProviderBids(SECOND, 0, 10)).deep.eq([[bidId], 1n]);
      expect(await marketplace.getModelBids(modelId1, 0, 10)).deep.eq([[bidId], 1n]);
      expect(await marketplace.getProviderActiveBids(SECOND, 0, 10)).deep.eq([[bidId], 1n]);
      expect(await marketplace.getModelActiveBids(modelId1, 0, 10)).deep.eq([[bidId], 1n]);
    });
    it('should post few model bids', async () => {
      await setNextTime(300);
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId2, wei(20));

      const bidId1 = await marketplace.getBidId(SECOND, modelId1, 0);
      let data = await marketplace.getBid(bidId1);
      expect(data.provider).to.eq(SECOND);
      expect(data.modelId).to.eq(modelId1);
      expect(data.pricePerSecond).to.eq(wei(10));
      expect(data.nonce).to.eq(0);
      expect(data.createdAt).to.eq(300);
      expect(data.deletedAt).to.eq(0);

      const bidId2 = await marketplace.getBidId(SECOND, modelId2, 0);
      data = await marketplace.getBid(bidId2);
      expect(data.provider).to.eq(SECOND);
      expect(data.modelId).to.eq(modelId2);
      expect(data.pricePerSecond).to.eq(wei(20));
      expect(data.nonce).to.eq(0);
      expect(data.createdAt).to.eq(301);
      expect(data.deletedAt).to.eq(0);

      expect(await token.balanceOf(marketplace)).to.eq(wei(302));
      expect(await token.balanceOf(SECOND)).to.eq(wei(698));

      expect(await marketplace.getProviderBids(SECOND, 0, 10)).deep.eq([[bidId1, bidId2], 2n]);
      expect(await marketplace.getModelBids(modelId1, 0, 10)).deep.eq([[bidId1], 1n]);
      expect(await marketplace.getModelBids(modelId2, 0, 10)).deep.eq([[bidId2], 1n]);
      expect(await marketplace.getProviderActiveBids(SECOND, 0, 10)).deep.eq([[bidId1, bidId2], 2n]);
      expect(await marketplace.getModelActiveBids(modelId1, 0, 10)).deep.eq([[bidId1], 1n]);
      expect(await marketplace.getModelActiveBids(modelId2, 0, 10)).deep.eq([[bidId2], 1n]);
    });
    it('should post a new model bid and delete an old bid when an old bid is active', async () => {
      await setNextTime(300);
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(20));

      const bidId1 = await marketplace.getBidId(SECOND, modelId1, 0);
      let data = await marketplace.getBid(bidId1);
      expect(data.deletedAt).to.eq(301);

      const bidId2 = await marketplace.getBidId(SECOND, modelId1, 1);
      data = await marketplace.getBid(bidId2);
      expect(data.provider).to.eq(SECOND);
      expect(data.modelId).to.eq(modelId1);
      expect(data.pricePerSecond).to.eq(wei(20));
      expect(data.nonce).to.eq(1);
      expect(data.createdAt).to.eq(301);
      expect(data.deletedAt).to.eq(0);

      expect(await token.balanceOf(marketplace)).to.eq(wei(302));
      expect(await token.balanceOf(SECOND)).to.eq(wei(698));

      expect(await marketplace.getProviderBids(SECOND, 0, 10)).deep.eq([[bidId1, bidId2], 2]);
      expect(await marketplace.getModelBids(modelId1, 0, 10)).deep.eq([[bidId1, bidId2], 2]);
      expect(await marketplace.getProviderActiveBids(SECOND, 0, 10)).deep.eq([[bidId2], 1]);
      expect(await marketplace.getModelActiveBids(modelId1, 0, 10)).deep.eq([[bidId2], 1]);
    });
    it('should post a new model bid and skip the old bid delete', async () => {
      await setNextTime(300);
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));

      const bidId1 = await marketplace.getBidId(SECOND, modelId1, 0);
      await marketplace.connect(SECOND).deleteModelBid(bidId1);
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(20));
    });
    it('should throw error when the provider is deregistered', async () => {
      await providerRegistry.connect(SECOND).providerDeregister(SECOND);
      await expect(marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10))).to.be.revertedWithCustomError(
        marketplace,
        'MarketplaceProviderNotFound',
      );
    });
    it('should throw error when the model is deregistered', async () => {
      await modelRegistry.connect(SECOND).modelDeregister(modelId1);
      await expect(marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10))).to.be.revertedWithCustomError(
        marketplace,
        'MarketplaceModelNotFound',
      );
    });
    it('should throw error when the bid price is invalid #1', async () => {
      await expect(
        marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(99999)),
      ).to.be.revertedWithCustomError(marketplace, 'MarketplaceBidPricePerSecondInvalid');
    });
    it('should throw error when the bid price is invalid #2', async () => {
      await expect(marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(0))).to.be.revertedWithCustomError(
        marketplace,
        'MarketplaceBidPricePerSecondInvalid',
      );
    });
    it('should throw error when model not found', async () => {
      await expect(
        marketplace.connect(SECOND).postModelBid(SECOND, getHex(Buffer.from('123')), wei(1)),
      ).to.be.revertedWithCustomError(marketplace, 'MarketplaceModelNotFound');
    });
    it('should throw error when provider not found', async () => {
      await expect(
        marketplace.connect(OWNER).postModelBid(OWNER, getHex(Buffer.from('123')), wei(1)),
      ).to.be.revertedWithCustomError(marketplace, 'MarketplaceProviderNotFound');
    });
    it('should throw error when post a bid without delegation or with incorrect rules', async () => {
      await expect(marketplace.connect(OWNER).postModelBid(SECOND, modelId1, wei(10))).to.be.revertedWithCustomError(
        providerRegistry,
        'InsufficientRightsForOperation',
      );

      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, getHex(Buffer.from('123')), true);
      await expect(marketplace.connect(OWNER).postModelBid(SECOND, modelId1, wei(10))).to.be.revertedWithCustomError(
        providerRegistry,
        'InsufficientRightsForOperation',
      );
    });
  });

  describe('#deleteModelBid', async () => {
    it('should delete a bid', async () => {
      await setNextTime(300);
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));

      const bidId1 = await marketplace.getBidId(SECOND, modelId1, 0);
      await marketplace.connect(SECOND).deleteModelBid(bidId1);

      const data = await marketplace.getBid(bidId1);
      expect(data.deletedAt).to.eq(301);
      expect(await marketplace.isBidActive(bidId1)).to.eq(false);
    });
    it('should delete a bid from the delegatee address', async () => {
      await setNextTime(300);
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));

      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_MARKETPLACE(), true);

      const bidId1 = await marketplace.getBidId(SECOND, modelId1, 0);
      await marketplace.connect(OWNER).deleteModelBid(bidId1);

      const data = await marketplace.getBid(bidId1);
      expect(data.deletedAt).to.eq(302);
      expect(await marketplace.isBidActive(bidId1)).to.eq(false);
    });
    it('should throw error when caller is not an owner', async () => {
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));

      const bidId1 = await marketplace.getBidId(SECOND, modelId1, 0);
      await expect(marketplace.connect(PROVIDER).deleteModelBid(bidId1)).to.be.revertedWithCustomError(
        marketplace,
        'InsufficientRightsForOperation',
      );
    });
    it('should throw error when bid already deleted', async () => {
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));

      const bidId1 = await marketplace.getBidId(SECOND, modelId1, 0);
      await marketplace.connect(SECOND).deleteModelBid(bidId1);
      await expect(marketplace.connect(SECOND).deleteModelBid(bidId1)).to.be.revertedWithCustomError(
        marketplace,
        'MarketplaceActiveBidNotFound',
      );
    });
  });

  describe('#withdraw', async () => {
    beforeEach(async () => {
      await marketplace.setMarketplaceBidFee(wei(1));
    });

    it('should withdraw fee, all fee balance', async () => {
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));
      expect(await marketplace.getFeeBalance()).to.eq(wei(1));

      await marketplace.withdrawFee(PROVIDER, wei(999));

      expect(await marketplace.getFeeBalance()).to.eq(wei(0));
      expect(await token.balanceOf(marketplace)).to.eq(wei(300));
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(1));
    });
    it('should withdraw fee, part of fee balance', async () => {
      await marketplace.connect(SECOND).postModelBid(SECOND, modelId1, wei(10));
      expect(await marketplace.getFeeBalance()).to.eq(wei(1));

      await marketplace.withdrawFee(PROVIDER, wei(0.1));

      expect(await marketplace.getFeeBalance()).to.eq(wei(0.9));
      expect(await token.balanceOf(marketplace)).to.eq(wei(300.9));
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(0.1));
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(marketplace.connect(SECOND).withdrawFee(PROVIDER, wei(1))).to.be.revertedWithCustomError(
        diamond,
        'OwnableUnauthorizedAccount',
      );
    });
    it('should throw error when withdraw amount is zero', async () => {
      await expect(marketplace.withdrawFee(PROVIDER, wei(1))).to.be.revertedWithCustomError(
        marketplace,
        'MarketplaceFeeAmountIsZero',
      );
    });
  });
});

// npm run generate-types && npx hardhat test "test/diamond/facets/Marketplace.test.ts"
// npx hardhat coverage --solcoverjs ./.solcover.ts --testfiles "test/diamond/facets/Marketplace.test.ts"
