# Morpheus Lumerin Node
![Simple-Overview](docs/images/simple.png)
The purpose of this software is to enable interaction with distributed, decentralized LLMs on the Morpheus network through a desktop chat experience.

The steps listed below are for both the Consumer and Provider to get started with the Morpheus Lumerin Node. As the software is developed, both onboarding & configuration of the provider and consumer roles will be simplified, automated and more transparent to the end user.

# **NOTE: ARBITRUM SEPOLIA TESTNET ONLY at this time - DEVELOPER PREVIEW ONLY**

**Components that are included in this repository are:**
* Local `Llama.cpp` and tinyllama model to run locally for demonstration purposes only
* Lumerin `proxy-router` is a background process that monitors sepcific blockchain contract events, 
manages secure sessions between consumers and providers and routes prompts and responses between them
* Lumerin `ui-desktop` is the front end UI to interact with LLMs and the Morpheus network via the proxy-router as a consumer
* Lumerin `cli` is the cli client to interact with LLMs and the Morpheus network via the proxy-router as a consumer

## Tokens and Contract Information 
* Morpheus saMOR Token: `0xc1664f994fd3991f98ae944bc16b9aed673ef5fd` 
* Lumerin Morpheus Smart Contract : `0x8e19288d908b2d9F8D7C539c74C899808AC3dE45`
    * Interact with the Morpheus Contract: https://louper.dev/diamond/0x8e19288d908b2d9F8D7C539c74C899808AC3dE45?network=arbitrumSepolia#write
* Blockchain Explorer: `https://sepolia.arbiscan.io/`
* Swagger API: `http://localhost:8082/swagger/index.html`

## Funds
* **WALLET:** For testing as a provider or consumer, you will need both `saMOR` and `saETH` tokens in your wallet. You should be able to get either of these from the usual Sepolia Arbitrum testnet faucets.
    * `saMOR` is the token used to pay for the model provider staking and consumer usage
    * `saETH` is the token used to pay for the gas fees on the network  

## Installation & Operation 
* [00-Overview](docs/00-overview.md) - This provides a comprehensive picture of the Provider, Blockchain and Consumer environments and how they interact. This will also link to other documents for more advanced setup and configuration.

* [04-Consumer-Setup](docs/04-consumer-setup.md) - This is the simplest way to get started with the Morpheus Lumerin Node as a Consumer.  This will allow you to interact with the Morpheus network and the models offered on the network as a consumer running from packaged releases.

* [02-Provider-Setup](docs/02-provider-setup.md) - This is the simplest way to get started with the Morpheus Lumerin Node as a Provider.  This will allow you to connect your existing AI-Model to the Morpheus network and offer it for use by consumers.