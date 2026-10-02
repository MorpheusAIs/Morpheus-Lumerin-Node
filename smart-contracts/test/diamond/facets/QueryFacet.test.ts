import { expect } from 'chai';
import { ethers } from 'hardhat';

describe('QueryFacet (RFP H5)', () => {
  it('should compile and deploy QueryFacet successfully', async () => {
    const QueryFacetFactory = await ethers.getContractFactory('QueryFacet');
    const queryFacet = await QueryFacetFactory.deploy();
    await queryFacet.waitForDeployment();

    expect(await queryFacet.getAddress()).to.properAddress;
  });

  it('should return empty structs for non-existent session/bid views safely without reverting', async () => {
    const QueryFacetFactory = await ethers.getContractFactory('QueryFacet');
    const queryFacet = await QueryFacetFactory.deploy();
    await queryFacet.waitForDeployment();

    const emptyBytes32 = ethers.ZeroHash;
    const sessionView = await queryFacet.getSessionView(emptyBytes32);
    expect(sessionView.session.user).to.eq(ethers.ZeroAddress);
    expect(sessionView.bid.provider).to.eq(ethers.ZeroAddress);

    const bidView = await queryFacet.getBidView(emptyBytes32);
    expect(bidView.bid.provider).to.eq(ethers.ZeroAddress);
    expect(bidView.provider.endpoint).to.eq('');

    const [views, total] = await queryFacet.getModelBidsEnriched(emptyBytes32, 0, 10);
    expect(total).to.eq(0n);
    expect(views.length).to.eq(0);
  });
});
