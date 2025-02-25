export const lmrDecimals = 10 ** 18;
export const ethDecimals = 10 ** 18;

export const fromTokenBaseUnitsToLMR = (baseUnits) => baseUnits / lmrDecimals;

export const fromTokenBaseUnitsToETH = (baseUnits) => baseUnits / ethDecimals;

export const formatValue = (valueWithDecimals: number, decimals = 18) => {
  const value = valueWithDecimals / 10 ** decimals;
  return value.toFixed(2);
};
