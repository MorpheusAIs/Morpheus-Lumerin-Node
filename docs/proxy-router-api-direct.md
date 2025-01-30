# Direct Consumer interaction with the proxy-router (no MorpheusUI)
This document provides a step-by-step flow to query your local proxy router and interact with a remote model using only the API. This is useful for developers or users who want to interact with the model directly without using the MorpheusUI.

## Pre-requisites:
* Create or use an existing ERC-20 wallet that has MOR and ETH (Sepolia Arbitrum) tokens - you can use Metamask (new wallet..not derived) or any other ERC-20 wallet.  
* You will need to have access to the wallet's private key **NEVER SHARE THIS WITH ANYONE** for steps below to authorize the contract to spend on your behalf.
* Install and launch the local llama.cpp and proxy-router from source (see [/docs/mac-boot-strap.md](/docs/mac-boot-strap.md) for instructions)

## TL;DR
* Authorize the contract to spend on your behalf (once)  
* Query the blockchain for various models / providers & get the ModelID `Id` (every session)
* Create a session with the provider using the ModelID `Id` (every session)
* Interact with the provider by sending the prompt (every session)

## Detail:
### A. Start the proxy-router (per instructions)
See [/docs/proxy-router-api-auth.md](/docs/proxy-router-api-auth.md) for auth instructions.

### B. Authorize the contract to spend on your behalf
Either via the swagger interface http://localhost:8082/swagger/index.html#/wallet/post_blockchain_allowance or following CLI, you can authorize the contract to spend on your behalf. **This only needs to be done once per wallet, or when funds have been depleted.**

Approve the contract to spend 3 MOR tokens on your behalf

```sh 
curl -X 'POST' 'http://localhost:8082/blockchain/approve?spender=0xb8C55cD613af947E73E262F0d3C54b7211Af16CF&amount=3' -H 'accept: application/json' -d '' 
``` 

### C. Query the blockchain for various models / providers (Get ModelID)
* You can query the blockchain for various models and providers to get the ModelID. 
* This can be done via the swagger interface http://localhost:8082/swagger/index.html#/marketplace/get_marketplace_models or following CLI:
```sh 
# Returns the wallet ID (confirm that it matches your wallet): 
curl -X 'GET' 'http://localhost:8082/wallet' -H 'accept: application/json'

# Returns the list of models and providers: 
curl -X 'GET' 'http://localhost:8082/blockchain/models' -H 'accept: application/json'
``` 
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
Your Wallet (on https://arbitrum-sepolia.blockscan.com/address/<wallet_id>) should show the transaction for the session creation.

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
`curl -X 'POST' 'http://localhost:8082/blockchain/approve?spender=0xb8C55cD613af947E73E262F0d3C54b7211Af16CF&amount=3' -H 'accept: application/json' -d ''`
    # approves the smart contract `0x8e19...dE45` to spend 3 MOR tokens on your behalf

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
#### Sample Output: 
```
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758293,"model":"llama2","choices":[{"index":0,"delta":{"content":"Why"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758293,"model":"llama2","choices":[{"index":0,"delta":{"content":" don"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758293,"model":"llama2","choices":[{"index":0,"delta":{"content":"'"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758293,"model":"llama2","choices":[{"index":0,"delta":{"content":"t"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758293,"model":"llama2","choices":[{"index":0,"delta":{"content":" scientists"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758293,"model":"llama2","choices":[{"index":0,"delta":{"content":" trust"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758293,"model":"llama2","choices":[{"index":0,"delta":{"content":" atoms"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":"?"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":"\n"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":"\n"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":"Because"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":" they"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":" make"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":" up"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":" everything"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{"content":"\u003c|im_end|\u003e"},"finish_reason":null,"content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
data: {"id":"chatcmpl-8wtgv86nUWMJneEObprWKXcWKX81PAAo","object":"chat.completion.chunk","created":1721758294,"model":"llama2","choices":[{"index":0,"delta":{},"finish_reason":"stop","content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false}}}]}
```


