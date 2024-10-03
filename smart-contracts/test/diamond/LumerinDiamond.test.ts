import { LumerinDiamond } from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { Reverter } from '../helpers/reverter';

describe('LumerinDiamond', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;

  let diamond: LumerinDiamond;

  before('setup', async () => {
    [OWNER] = await ethers.getSigners();

    const [LumerinDiamond] = await Promise.all([ethers.getContractFactory('LumerinDiamond')]);

    [diamond] = await Promise.all([LumerinDiamond.deploy()]);

    await diamond.__LumerinDiamond_init();

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('Diamond functionality', () => {
    describe('#__LumerinDiamond_init', () => {
      it('should set correct data after creation', async () => {
        expect(await diamond.owner()).to.eq(await OWNER.getAddress());
      });
      it('should revert if try to call init function twice', async () => {
        const reason = 'Initializable: contract is already initialized';

        await expect(diamond.__LumerinDiamond_init()).to.be.rejectedWith(reason);
      });
    });
  });
});
