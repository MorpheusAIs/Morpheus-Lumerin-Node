export const buildLocalModelsConfig = (modelName: string, apiType: string, apiUrl: string) => {
  return {
    $schema:
      'https://raw.githubusercontent.com/MorpheusAIs/Morpheus-Lumerin-Node/a719073670adb17de6282b12d1852d39d629cb6e/proxy-router/internal/config/models-config-schema.json',
    models: [
      {
        modelId: '0x0000000000000000000000000000000000000000000000000000000000000000',
        modelName: modelName,
        apiType: apiType,
        apiUrl: apiUrl
      }
    ]
  }
}

export const buildLocalRatingConfig = () => {
  return {
    $schema:
      'https://raw.githubusercontent.com/MorpheusAIs/Morpheus-Lumerin-Node/a719073670adb17de6282b12d1852d39d629cb6e/proxy-router/internal/rating/rating-config-schema.json',
    algorithm: 'default',
    providerAllowlist: [],
    params: {
      weights: {
        tps: 0.24,
        ttft: 0.08,
        duration: 0.24,
        success: 0.32,
        stake: 0.12
      }
    }
  }
}
