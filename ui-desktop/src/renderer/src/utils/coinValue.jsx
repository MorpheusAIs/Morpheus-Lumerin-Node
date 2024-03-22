export const lmrDecimals = 10 ** 8;
export const ethDecimals = 10 ** 18;

export const fromTokenBaseUnitsToLMR = baseUnits => baseUnits / lmrDecimals;

export const fromTokenBaseUnitsToETH = baseUnits => baseUnits / ethDecimals;
