
# Provider Hosting (Local LLM to offer, Proxy-Router running as background/service): 

## Assumptions: 
* Your AI model has been configured, started and made available to the proxy-router server via a private endpoint (IP:PORT or DNS:PORT) eg: `http://mycoolaimodel.domain.com:8080`
    * Optional
        * You can use the provided `llama.cpp` and `tinyllama` model to test locally
        * If your local model is listening on a different port locally, you will need to modify the `OPENAI_BASE_URL` in the .env file to match the correct port
* You have an existing funded wallet with saMOR and saETH and also have the `private key` for the wallet (this will be needed for the .env file configuration)
* You have created an Alchemy or Infura free account and have a private API key for the Arbitrum Sepolia testnet (wss://arb-sepolia.g.alchemy.com/v2/<your_private_alchemy_api_key>)
* Your proxy-router must have a publicly accessible endpoint for the provider (ip:port or fqdn:port no protocol) eg: `mycoolmornode.domain.com:3333` - this will be used when creating the provider on the blockchain

## Installation & Configuration Steps:
1. Download latest release for your operating system: https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node/releases

1. Extract the zip to a local folder (examples)
    * Windows: `(%USERPROFILE%)/Downloads/morpheus)` 
    * Linux & MacOS: `~/Downloads/morpheus`
    * On MacOS you may need to execute `xattr -c proxy-router` in a command window to remove the quarantine flag on MacOS

1.  Edit the `.env` file following the guide below [proxy-router .ENV Variables](#proxy-router-env-variables) 

1. **(OPTIONAL)** - External Provider or Pass through 
    * In some cases you will want to leverage external or existing AI Providers in the network via their own, private API
    * Dependencies: 
        * `model-config.json` file in the proxy-router directory
        * proxy-router .env file for proxy-router must also be updated to include `MODELS_CONFIG_PATH=<path_to_proxy-router>/models-config.json`
    * Once your provider is up and running, deploy a new model and model bid via the diamond contract (you will need the `model_ID` for the configuration)
    * Edit the model-config.json to the following json format ... with 
    * The JSON ID will be the ModelID that you created above, modelName, apiTYpe, apiURL and apiKey are from the external provider and specific to their offered models 
    * Once the model-config.json file is updated, the morpheus node will need to be restarted to pick up the new configuration (not all models (eg: image generation can be utilized via the UI-Desktop, but API integration is possible)
    * Example model-config.json file for external providers
```
#model-config.json 
{
    "0x4b5d6c2d3e4f5a6b7c8de7f89a0b19e07f4a6e1f2c3a3c28d9d5e6": {
        "modelName": "v1-5-specialmodel.modelversion [externalmodel]",
        "apiType": "provider_api_type",
        "apiUrl": "https://api.externalmodel.com/v1/xyz/generate",
        "apiKey": "api-key-from-external-provider"
    },
    "0xb2c8a6b2c1d9ed7f0e9a3b4c2d6e5f14f9b8c3a7e5d6a1a0b9c7d8e4f30f4a7b": {
        "modelName": "v1-7-specialmodel2.modelversion [externalmodel]",
        "apiType": "provider_api_type",
        "apiUrl": "https://api.externalmodel.com/v1/abc/generate",
        "apiKey": "api-key-from-external-provider"
    }
}
```

## Start the Proxy Router 
1. On your server, launch the proxy-router with the modified .env file shown above
    * Windows: Double click the `proxy-router.exe` (You will need to tell Windows Defender this is ok to run)  
    * Linux & MacOS: Open a terminal and navigate to the folder and run `./proxy-router`from the morpheus/proxy-router folder
1.  This will start the proxy-router and begin monitoring the blockchain for events
    
## Validating Steps:
1. Once the proxy-router is running, you can navigate to the Swagger API Interface (http://localhost:8082/swagger/index.html as example) to validate that the proxy-router is running and listening for blockchain events
1. You can also check the logs in the `./data` directory for any errors or issues that may have occurred during startup
1. Once validated, you can move on and create your provider, model and bid on the blockchain [03-provider-offer.md](03-provider-offer.md)


----------------
### proxy-router .ENV Variables 
Key Values in the .env file are (there are others, but these are primarly responsible for connecting to the blockchain, the provider AI model and listening for incoming traffic): 
- `WALLET_PRIVATE_KEY=` 
    - Private Key from your wallet needed for the proxy-router to sign transactions and respond to provided prompts (this is why the proxy router must be secured and the API endpoint protected)
- `ETH_NODE_ADDRESS=wss://arb-sepolia.g.alchemy.com/v2/<your_private_alchemy_api_key>`
    - Ethereum Node Address for the Arbitrum blockchain (via Alchemy or Infura)
    - This websocket (wss) address is key for the proxy-router to listen and post to the blockchain
    - We recommend using your own private ETH Node Address for better performance (free account setup via Alchemy or Infura)
- `DIAMOND_CONTRACT_ADDRESS=0x208eaeD75A12C35625708140c99A614FC45bf780`
    - This is the key Lumerin Smart Contract (currently Sepolia Arbitrum testnet)
    - This is the address of the smart contract that the proxy-router will interact with to post providers, models & bids 
    - This address will change as the smart-contract is updated and for mainnet contract interaction 
- `MOR_TOKEN_ADDRESS=0xc1664f994fd3991f98ae944bc16b9aed673ef5fd`
    - This is the Morpheus Token (saMOR) address for Sepolia Arbitrum testnet
    - This address will be different for mainnet token
- `WEB_ADDRESS=0.0.0.0:8082` 
    - This is the local listenting port for your proxy-router API (Swagger) interface
    - Based on your local needs, this may need to change (8082 is default)
- `WEB_PUBLIC_URL=localhost:8082` 
    - If you have or will be exposing your API interface to a local, PRIVATE (or VPN) network, you can change this to the DNS name or IP and port where the API will be available. The default is just on the local machine (localhost)
    - The PORT must be the same as in the `WEB_ADDRESS` setting             
- `OPENAI_BASE_URL=http://localhost:8080/v1` 
    - This is where the proxy-router should send OpenAI compatible requests to the provider model. 
    - By default (and included in the Morpheus-Lumerin software releases) this is set to `http://localhost:8080/v1` for the included llama.cpp model
    - In a real-world scenario, this would be the IP address and port of the provider model server or server farm that is hosting the AI model separately from the proxy-router 
- `PROXY_STORAGE_PATH=./data/`
    - This is the path where the proxy-router will store logs and other data
    - This path should be writable by the user running the proxy-router software
- `MODELS_CONFIG_PATH=` 
    - location of the models-config.json file that contains the models that the proxy-router will be providing. 
    - it has the capability to also (via private API) call external providers models (like Prodia)
- `PROXY_ADDRESS=0.0.0.0:3333` 
    - This is the local listening port for the proxy-router to receive prompts and inference requests from the consumer nodes
    - This is the port that the consumer nodes will send prompts to and should be available publicly and via the provider definition setup on the blockchain