# Example models config file.  Local model configurations are stored in this file
* `rooot_key` (required) is the model id
* `modelName` (required) is the name of the model
* `apiType` (required) is the type of the model api.  Currently supported values are "prodia" and "openai"
* `apiUrl` (required) is the url of the LLM server or model API
* `apiKey` (optional) is the api key for the model
* `cononcurrentSlots` (optional) are number of available distinct chats on the llm server and used for capacity policy
* `capacityPolicy` (optional) can be one of the following: "idle_timeout", "simple" 

## Examples of models-config.json entries 
* The first key (`0x6a...9018`) is the model id of the local default model (llama2)
* The middle two keys are examples of externally hosted and owned models where the morpheus-proxy-router enables proxying requests to the external model API
* The last key is an example of a model hosted on a server owned by the morpheus-proxy-router operator

```bash
{
    "0x6a4813e866a48da528c533e706344ea853a1d3f21e37b4c8e7ffd5ff25779018": {
        "modelName": "llama2",
        "apiType": "openai"
    }, 
    "0x60d5900b10534de1a668fd990bd790fa3fe04df8177e740f249c750023a680fb": {
        "modelName": "v1-5-pruned-emaonly.safetensors [d7049739]",
        "apiType": "prodia",
        "apiUrl": "https://api.prodia.com/v1/sd/generate",
        "apiKey": "ed53950-852a-45a7-bf07-47b89bb492e38"
    },
    "0x0d90cf8ca0a811a5cd2347148171e9a2401e9fbc32f683b648c5c1381df91ff7": {
        "modelName": "animagineXLV3_v30.safetensors [75f2f05b]",
        "apiType": "prodia",
        "apiUrl": "https://api.prodia.com/v1/sdxl/generate",
        "apiKey": "ed53950-852a-45a7-bf07-47b89bb492e38"
    },
    "0x06c7b502d0b7f14a96bb4fda5f2ba941068601d5f6b90804c9330e96b093b0ce": {
        "modelName": "LMR-Collective Cognition Mistral 7B",
        "apiType": "openai",
        "apiUrl": "http://llmserver.domain.io:8080/v1",
        "concurrentSlots": 8,
        "capacityPolicy": "simple"
    }
}
```