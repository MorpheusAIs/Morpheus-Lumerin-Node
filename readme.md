# Morpheus Lumerin Node
### Take part in the Lumerin coding weight rewards!! [stake.mor.lumerin.io](https://stake.mor.lumerin.io/)
![Simple-Overview](docs/images/simple.png)
The purpose of this software is to enable interaction with distributed, decentralized LLMs on the Morpheus network through a desktop chat experience.

0. PreRequisites: BASE Layer 2 Blockchain, MOR and ETH on BASE for staking and bidding
1. Existing, Hosted AI model that is available for inference via the Morpheus network
2. The proxy-router talks to and listens to the blockchain, routes prompts and inference between the providers’ models and the consumers that purchase and use the models
3. Providers register their models via bids on the blockchain
4. The consumer node is the “client” that will be purchasing bids from the blockchain, sending prompts via the proxy-router and receiving inference back from the provider’s models
5. Consumers purchase the bid and stake MOR for their session time
6. Once the bid has been purchased, prompt and inference (ChatGPT-like) can start


**Components that are included in this repository are:**
* Local `Llama.cpp` and tinyllama model to run locally for demonstration purposes only
* Lumerin `proxy-router` is a background process that monitors sepcific blockchain contract events, 
manages secure sessions between consumers and providers and routes prompts and responses between them
* Lumerin `MorpheusUI` is the front end UI to interact with LLMs and the Morpheus network via the proxy-router as a consumer
* Lumerin `cli` is the cli client to interact with LLMs and the Morpheus network via the proxy-router as a consumer
* Kubo `ipfs` is the ipfs client to interact with the ipfs network to store and retrieve model/agent files

## Tokens and Contract Information (updated 12/17/2025)
### MainNet: (MAIN Branch and MAIN-* Releases)
* Blockchain: BASE Mainnet (ChainID: `8453`)
* Morpheus MOR Token: `0x7431aDa8a591C955a994a21710752EF9b882b8e3` 
* Diamond MarketPlace Contract: `0x6aBE1d282f72B474E54527D93b979A4f64d3030a` 
* Blockchain Explorer: `https://base.blockscout.com/`
* GitHub Source: https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/tree/main

### TestNet (TEST Branch and TEST-* Releases)
* Blockchain: BASE Sepolia (ChainID: `84532`)
* Morpheus MOR Token: `0x5C80Ddd187054E1E4aBBfFCD750498e81d34FfA3` 
* Diamond MarketPlace Contract: `0x6e4d0B775E3C3b02683A6F277Ac80240C4aFF930`
* Blockchain Explorer: `https://base-sepolia.blockscout.com/`
* GitHub Source: https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/tree/test 

## Funds
* **WALLET:** For testing as a provider or consumer, you will need both `MOR` and `ETH` tokens in your wallet. 
    * `MOR` is the token used to pay for the model provider staking and consumer usage
    * `ETH` is the token used to pay for the gas fees on the network  

## Installation & Operation 
* [00-Overview](docs/00-overview.md) - This provides a comprehensive picture of the Provider, Blockchain and Consumer environments and how they interact. This will also link to other documents for more advanced setup and configuration.

* [04-Consumer-Setup](docs/04-consumer-setup.md) - This is the simplest way to get started with the Morpheus Lumerin Node as a Consumer.  This will allow you to interact with the Morpheus network and the models offered on the network as a consumer running from packaged releases.

* [02-Provider-Setup](docs/02-provider-setup.md) - This is the simplest way to get started with the Morpheus Lumerin Node as a Provider.  This will allow you to connect your existing AI-Model to the Morpheus network and offer it for use by consumers.
    * [02.1-Proxy-Router-Docker](docs/02.1-proxy-router-docker.md) - Fast start using existing Docker image and proxy-router configuration