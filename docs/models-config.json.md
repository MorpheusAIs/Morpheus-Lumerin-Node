# Example models config file. Local model configurations are stored in this file

- `modelId` (required) is the model id
- `modelName` (required) is the name of the model
- `apiType` (required) is the type of the model api. Currently supported values are "prodia-sd", "prodia-sdxl", "prodia-v2" and "openai"
- `apiUrl` (required) is the url of the LLM server or model API. Full url including endpoint.
- `apiKey` (optional) is the api key for the model
- `concurrentSlots` (optional) are number of available distinct chats on the llm server and used for capacity policy
- `capacityPolicy` (optional) can be one of the following: "idle_timeout", "simple"
- There maybe other variables that should be included in the model configuration. Please refer to the json-schema for descriptions and list of required and optional fields.

## Examples of models-config.json entries

This file enables the morpheus-proxy-router to route requests to the correct model API. The model API can be hosted on the same server as the morpheus-proxy-router or on an external server. Please refer to the json-schema for descriptions and list of required and optional fields.

```json
{
  "$schema": "./internal/config/models-config-schema.json",
  "models": [
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "modelName": "llama2",
      "apiType": "openai",
      "apiUrl": "http://localhost:8080/v1/chat/completions"
    },
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000001",
      "modelName": "inference.sdxl.txt2img.v1",
      "apiType": "prodia-v2",
      "apiUrl": "https://inference.prodia.com/v2/job",
      "apiKey": "FILL_ME_IN"
    },
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000002",
      "modelName": "SDXL1.0-base",
      "apiType": "hyperbolic-sd",
      "apiUrl": "https://api.hyperbolic.xyz/v1/image/generation",
      "apiKey": "FILL_ME_IN"
      "parameters": {
        "cfg_scale": "5",
        "steps": "30"
      }
    },
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000003",
      "modelName": "claude-3-5-sonnet-20241022",
      "apiType": "claudeai",
      "apiUrl": "https://api.anthropic.com/v1/messages",
      "apiKey": "FILL_ME_IN"
    },
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000004",
      "modelName": "inference.sd15.txt2img.v1",
      "apiType": "prodia-v2",
      "apiUrl": "https://inference.prodia.com/v2/job",
      "apiKey": "FILL_ME_IN"  
    },    
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000005",
      "modelName": "gpt-4o-mini",
      "apiType": "openai",
      "apiUrl": "https://api.openai.com/v1/chat/completions",
      "apiKey": "FILL_ME_IN"
    },
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000006",
      "modelName": "text-embedding-bge-m3",
      "apiType": "openai",
      "apiUrl": "https://api.venice.ai/api/v1/embeddings",
      "apiKey": "FILL_ME_IN"
    },
    {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000007",
      "modelName": "tts-kokoro",
      "apiType": "openai",
      "apiUrl": "https://api.venice.ai/api/v1/audio/speech",
      "apiKey": "FILL_ME_IN"
    },
        {
      "modelId": "0x0000000000000000000000000000000000000000000000000000000000000008",
      "modelName": "whisper-1",
      "apiType": "openai",
      "apiUrl": "https://api.openai.com/v1/audio/transcriptions",
      "apiKey": "FILL_ME_IN"
    },
  ]
}
```
