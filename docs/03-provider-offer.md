
# Creating Provider, Model and Bid on the Blockchain:
**Diamond contract:** `0x208eaeD75A12C35625708140c99A614FC45bf780`

**Needed information:**
* Provider/Owner:   `0x9E26Fea97F7d644BAf62d0e20e4d4b8F836C166c` # Your ERC-20 Wallet with saMOR & saETH
* Endpoint:         `server.domain.com:3333` # Internet publicly accessible server/node access point 
* Model ID:         `0xe1e6e3e77148d140065ef2cd4fba7f4ae59c90e1639184b6df5c84` # Random 32byte/hex that you generate 
* ipfcCID:          `0xc2d3a5e4f9b7c1a2c8f0b1d5e6c78492fa7bcd34e9a3b9c9e18f25d3be47a1f6` # Another 32byte/hex random for future use
* Model Name:       `CapybaraHermes-v2.5-Mistral-7b` # Human Readable name for the model
* Bid Cost:         `200000000000` (1*10^18 or ~7MOR) # What will the model cost per second to use

## Steps
 1. WEB3/Arbiscan/Metamask: Authorize Diamond Contract to spend on the Provider's behalf 
       1. https://sepolia.arbiscan.io/address/0xc1664f994fd3991f98ae944bc16b9aed673ef5fd#writeContract 
       1. Connect to Web3 (connect Provider wallet) 
       1. Click Approve 
       1. Spender Address = Diamond Contract 
       1. Authorized Amount = remember that this is in the form 1*10^18 so make sure there's enough MOR 1ranted to cover the contract fees 
       1. The Diamond Contract is now authorized to spend MOR on provider's behalf 

1. Create Provider in the Diamond contract via swagger api:
    1. Start proxy-router 
    1. http://localhost:8082/swagger/index.html#/providers/post_blockchain_providers
    1. Enter required fields:   
        1. addStake = Amount of stake for provider to risk - Stake can be 0 now 
        1. Endpoint = Your publicly accessible endpoint for the proxy-router provider (ip:port or fqdn:port no protocol) eg: `mycoolmornode.domain.com:3333`

1. Create Model in the contract:
    1. Go to http://localhost:8082/swagger/index.html#/models/post_blockchain_models and enter
        1. modelId: random 32byte/hex that will uniquely identify model (uuid)
        1. ipfsCID: another random32byte/hex for future use (model library)
        1. Fee: fee for the model usage - 0 for now
        1. addStake: stake for model usage - 0 for now 
        1. Owner: Provider Wallet Address 
        1. name: Human Readable model like "Llama 2.0" or "Mistral 2.5" or "Collective Cognition 1.1" 
        1. tags: array of tag strings for the model 
        1. Capture the `modelID` from the JSON response

1. Offer Model Bid in the contract: 
    1. Navigate to http://localhost:8082/swagger/index.html#/bids/post_blockchain_bids and enter
        1. modelID: Model ID Created in last step: 
        1. pricePerSecond: this is in 1*10^18 format so 100000000000 should make 5 minutes for the session 1round 37.5 saMOR 
        1. Click Execute 

----------------