
### MORPHEUS CREATING PROVIDERS:
**Diamond contract:** `0x8e19288d908b2d9F8D7C539c74C899808AC3dE45`
**Needed information:**
* Provider/Owner:   `0x9E26Fea97F7d644BAf62d0e20e4d4b8F836C166c` # Your ERC-20 Wallet with saMOR & saETH
* Endpoint:         `server.domain.com:1234` # Internet publicly accessible server/node access point 
* Model ID:         `0xe1e6e3e77148d140065ef2cd4fba7f4ae59c90e1639184b6df5c84` # Random 32byte/hex that you generate 
* ipfcCID:          `0xc2d3a5e4f9b7c1a2c8f0b1d5e6c78492fa7bcd34e9a3b9c9e18f25d3be47a1f6` # Another 32byte/hex random for future use
* Model Name:       `CapybaraHermes-v2.5-Mistral-7b` # Human Readable name for the model
* Bid Cost:         `200000000000` (1*10^18 or ~7MOR) # What will the model cost per second to use


**1. Authorize Diamond Contract to spend on the Provider's behalf**
* https://sepolia.arbiscan.io/address/0xc1664f994fd3991f98ae944bc16b9aed673ef5fd#writeContract 
* Connect to Web3 (connect Provider wallet) 
* Click Approve 
* Spender Address = Diamond Contract 
* Authorized Amount = 3*10^18 (3 full MOR) 
* Write and approve via MetaMask 
* The Diamond Contract is now authorized to spend MOR on provider's behalf 

**2. Create Provider**
* https://louper.dev/diamond/0x8e19288d908b2d9F8D7C539c74C899808AC3dE45?network=arbitrumSepolia#write 
* Connect Wallet (approve via MM) 
* `ProviderRegistry/providerRegister` Facet 
    * addr = Provider address 
    * addStake = Amount of stake for provider to risk - Stake can be 0 now 
    * Endpoint = Publicly accessible endpoint for provider (ip:port or fqdn:port … no protocol)

**3. Create Model**
* Refresh page and reconnect wallet
* `ModelRegistry/modelRegister` facet 
    * modelId: random 32byte/hex that will uniquely identify model (uuid):
    * ipfsCID: another random for future use (model library): 
    * Fee: fee for the model usage - 42 for now 
    * addStake: stake for model usage - 0 for now 
    * Owner: Provider eth Address 
    * name: Human Readable model name 
    * tags: comma delimited tags for the model 

**4. Offer Model Bid**
* Refresh page and reconnect wallet
* `Marketplace/postModelBid` facet 
    * providerAddr: Provider / Owner 
    * modelID: Model ID Created in last step: 
    * pricePerSecond: (this is in 1*10^18 format) 
