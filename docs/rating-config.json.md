# Information about rating-config.json configuration file

This file configures rating system of proxy-router, that is responsible for provider selection. The file should be placed in the root directory of the project.

- `providerAllowlist` - the list of providers that are allowed to be used by the proxy-router. Keep it empty to allow all providers.
  - `"providerAllowlist": ["0x0000000000000000000000000000000000000000"]` will only allow the local, default model to be used
- `algorithm` - the algorithm used for rating calculation.
- `params` - algorithm parameters, like weights for different metrics. Each algorithm has its own set of parameters.

Please refer to the json schema for the full list of available fields.

```json
{
  "$schema": "./internal/rating/rating-config-schema.json",
  "algorithm": "default",
  "providerAllowlist": [],
  "params": {
    "weights": {
      "tps": 0.24,
      "ttft": 0.08,
      "duration": 0.24,
      "success": 0.32,
      "stake": 0.12
    }
  }
}
```
