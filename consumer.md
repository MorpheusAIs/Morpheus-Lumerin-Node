# Consumer Node Setup (Draft - 2024-07-23)
This document provides a step-by-step guide to setting up a Consumer Node for the Morepheus Network so that you can setup session and interact with the remote providers.

## Pre-requisites:
* Create or use an existing ERC-20 wallet that has saMOR and saETH (Sepolia Arbitrum) tokens - you can use Metamask (new wallet..not derived) or any other ERC-20 wallet.  You will need to have access to the wallet's private key **NEVER SHARE THIS WITH ANYONE** for steps below to authorize the contract to spend on your behalf.

## TL;DR
* Install and Configure the proxy-router node (once)
* Authorize the contract to spend on your behalf (once)  
* Query the blockchain for various models / providers & get the ModelID `Id` (every session)
* Create a session with the provider using the ModelID `Id` (every session)
* Interact with the provider by sending the prompt (every session)

## Detail:
==================================
### A. Proxy-Router CLI Setup  
#### 1. Install OS-Specific Dependencies
* git (https://git-scm.com/)
* go (https://golang.org/)

#### 2. Clone & Navigate to the Repository
```bash
git clone https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node.git #this will get you the DEV branch (daily updates)
cd Morpheus-Lumerin-Node/proxy-router
```
#### 3. Edit the .env configuration file
```bash 
cp .env.example .env
vi .env 
```
Modify the following entries to match your configuration for wallet and ethereum node: 
* `WALLET_PRIVATE_KEY=` # Private Key from your Wallet as Consumer or Provider (needed for the proxy-router to sign transactions)
* `ETH_NODE_ADDRESS=wss://` # Recommend using your own private ETH Node Address for better performance (via Alchemy or Infura)
Save the .env file and exit the editor 

#### 4. Build and start the proxy-router 
```bash 
./build.sh
go run cmd/main.go
```
After the iniial setup, you can execute `git pull` to get the latest updates and re-run the `./build.sh` and `go run cmd/main.go` to update the proxy-router with the latest changes.

#### 5. Confirm that the build is successful and console should show similar to below after started (and listening on specified ports 8082 for Swagger API and 3333 for the proxy-router):

You can also test http://localhost:8082/swagger/index.html to confirm the API is running and accessible.

```
Loaded config: {AIEngine:{OpenAIBaseURL: OpenAIKey:} Blockchain:{EthNodeAddress: EthLegacyTx:false ExplorerApiUrl:} Environment:development Marketplace:{DiamondContractAddress:0x8e19288d908b2d9F8D7C539c74C899808AC3dE45 MorTokenAddress:0xc1664f994Fd3991f98aE944bC16B9aED673eF5fD WalletPrivateKey:<nil>} Log:{Color:true FolderPath: IsProd:false JSON:false LevelApp:info LevelConnection:info LevelProxy:info LevelScheduler:info LevelContract:} Proxy:{Address:0.0.0.0:3333 MaxCachedDests:5 StoragePath:} System:{Enable:false LocalPortRange:1024 65535 NetdevMaxBacklog:100000 RlimitHard:524288 RlimitSoft:524288 Somaxconn:100000 TcpMaxSynBacklog:100000} Web:{Address:0.0.0.0:8082 PublicUrl:localhost:8082}}
2024-07-23T12:58:04.560735	INFO	APP	proxy-router TO BE SET AT BUILD TIME
2024-07-23T12:58:08.249559	INFO	APP	connected to ethereum node: wss://arb-sepolia.g.alchemy.com/v2/<masked>, chainID: 421614
2024-07-23T12:58:08.278792	INFO	BADGER	All 0 tables opened in 0s
2024-07-23T12:58:08.28444	INFO	BADGER	Discard stats nextEmptySlot: 0
2024-07-23T12:58:08.284515	INFO	BADGER	Set nextTxnTs to 0
2024-07-23T12:58:08.284769	WARN	Using env wallet. Private key persistance unavailable
2024-07-23T12:58:08.290268	INFO	proxy state: running
2024-07-23T12:58:08.290507	INFO	HTTP	http server is listening: 0.0.0.0:8082
2024-07-23T12:58:08.290631	INFO	Wallet address: <masked>
2024-07-23T12:58:08.290841	INFO	started watching events, address 0x8e19288d908b2d9F8D7C539c74C899808AC3dE45
2024-07-23T12:58:08.290866	INFO	TCP	tcp server is listening: 0.0.0.0:3333
```
==================================
### B. Authorize the contract to spend on your behalf
Either via the swagger interface http://localhost:8082/swagger/index.html#/wallet/post_blockchain_allowance or following CLI, you can authorize the contract to spend on your behalf. **This only needs to be done once per wallet, or when funds have been depleted.**
`curl -X 'POST' 'http://localhost:8082/blockchain/approve?spender=0x8e19288d908b2d9F8D7C539c74C899808AC3dE45&amount=3' -H 'accept: application/json' -d ''` # Approve the contract to spend 3 saMOR tokens on your behalf

### C. Query the blockchain for various models / providers (Get ModelID)
You can query the blockchain for various models and providers to get the ModelID. This can be done via the swagger interface http://localhost:8082/swagger/index.html#/marketplace/get_marketplace_models or following CLI:
* `curl -X 'GET' 'http://localhost:8082/wallet' -H 'accept: application/json'` # Returns the wallet ID (confirm that it matches your wallet)
* `curl -X 'GET' 'http://localhost:8082/blockchain/models' -H 'accept: application/json'` # Returns the list of models and providers
    * The first model in the list is the default model that you can use for testing purposes...see example below
    * `Id` is the ID of the model that you will use to create a session with the provider.
    * `Name` is the type of model offered 

```
{
  "models": [
    {
      "Id": "0x6a4813e866a48da528c533e706344ea853a1d3f21e37b4c8e7ffd5ff25779018",
      "IpfsCID": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "Fee": 0,
      "Stake": 0,
      "Owner": "0x0eb467381abbc5b71f275df0c8a4e0ed8561f46f",
      "Name": "llama2:7b",
      "Tags": [],
      "CreatedAt": 1721220139,
      "IsDeleted": false
    },
    {
      "Id": "0x72eb5a6a575cdfb59e650994240961db2b1d915dbaa7c009b53b20fe8b9d2d7c",
      "IpfsCID": "0x019ae5515ec6259cf835639fd645620811fe951f54c55ae5c85c1bb101cdcc3a",
      "Fee": 42,
      "Stake": 0,
      "Owner": "0x65bbb982d9b0afe9aed13e999b79c56ddf9e04fc",
      "Name": "CollectiveCognition-v1.1-Mistral-7b",
      "Tags": [],
      "CreatedAt": 1721222411,
      "IsDeleted": false
    },
    {
      "Id": "0x84b6df5c84e1e6ae59c90e1639e3e77148d140065ef2cd4fba7f41cc7440e2c5",
      "IpfsCID": "0xa3b9c9e18f25d3be47a2c8f0b1d5e6c78492fa7bcd34e9a1f6c2d3a5e4f9b7c1",
      "Fee": 42,
      "Stake": 0,
      "Owner": "0xb8f836c167d60e20e44baf62d4d46c9e26fea97f",
      "Name": "CapybaraHermes-v2.5-Mistral-7b",
      "Tags": [],
      "CreatedAt": 1721224046,
      "IsDeleted": false
    }
  ]
}
```

### D. Create a session with the provider 
```bash 
curl -s -X 'POST' 'http://localhost:8082/blockchain/models/<Id_from_model_query_above>/session' \
-H 'accept:application/json' \
-H 'Content-Type: application/json' \
-d '{"sessionDuration": 600}'
```
Now that the session is open, you can send inference queries to the provider and process responses in usual OpenAI format. 
Your Wallet (on https://sepolia.arbiscan.io/address/<wallet_id>) should show the transaction for the session creation.

### E. Interact with the provider
* Send the prompt (Standard OpenAI format) with session_id in the header to interact.  
* Minimally, you need to provide the `content` and `role` in the messages block and (currently) `"stream":true` for the remote prompt to work.
```bash
curl -X 'POST' \
  'http://localhost:8082/v1/chat/completions' \
  -H 'accept: application/json' \
  -H 'session_id: <sessionId_returned_from_session_open>' \
  -H 'Content-Type: application/json' \
  -d '{
  "messages": [
    {
      "content": "tell me a joke",
      "role": "user"
    }
  ],
  "stream": true
  }'
```


### Quick and Dirty Sample:
`curl -s -X 'GET' 'http://localhost:8082/wallet' -H 'accept: application/json' | jq .address` 
    # returns the wallet ID (confirm that it matches your wallet)

`curl -s -X 'GET' 'http://localhost:8082/blockchain/models' -H 'accept: application/json' | jq -r '.models[] | "\(.Id), \(.Name)"'`
    # returns model ID and Name ... copy the ID for the next step `0x84b6df5c84e1e6ae59c90e1639e3e77148d140065ef2cd4fba7f41cc7440e2c5`

`curl -s -X 'POST' 'http://localhost:8082/blockchain/models/0x84b6df5c84e1e6ae59c90e1639e3e77148d140065ef2cd4fba7f41cc7440e2c5/session' -H 'accept:application/json' -H 'Content-Type: application/json' -d '{"sessionDuration": 600}' | jq .sessionId` 
    # returns the session ID ... copy the ID for the next step `0x089111479fa2847106b4f7b17eace2e9b37e0d3c0db331b4e01a6e24de827477`

```bash
curl -X 'POST' \
  'http://localhost:8082/v1/chat/completions' \
  -H 'accept: application/json' \
  -H 'session_id: 0x089111479fa2847106b4f7b17eace2e9b37e0d3c0db331b4e01a6e24de827477' \
  -H 'Content-Type: application/json' \
  -d '{
  "messages": [
    {
      "content": "tell me a joke",
      "role": "user"
    }
  ],
  "stream": true
  }'
```

