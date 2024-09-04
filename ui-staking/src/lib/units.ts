export const decimalsMOR = 18;
export const decimalsLMR = 8;

export function formatMOR(num: bigint): string {
  return `${formatUnits(num, decimalsMOR)} MOR`;
}

export function formatLMR(num: bigint): string {
  return `${formatUnits(num, decimalsLMR)} LMR`;
}

const thousandsSeparator = ",";

export function formatUnits(value: bigint, decimals: number, significantDigits = 4) {
  let display = value.toString();

  const negative = display.startsWith("-");
  if (negative) display = display.slice(1);

  display = display.padStart(decimals, "0");

  let [integer, fraction] = [
    display.slice(0, display.length - decimals),
    display.slice(display.length - decimals),
  ];

  const integerSignificantDigits = integer.length;

  // split the integer part into groups of 3 digits
  for (let i = integer.length - 3; i > 0; i -= 3) {
    integer = `${integer.slice(0, i)}${thousandsSeparator}${integer.slice(i)}`;
  }

  const fractionSignificantDigits = significantDigits - integerSignificantDigits;

  // round the fraction part to thousands

  if (fractionSignificantDigits <= 0) {
    fraction = "";
  } else {
    // limit the number of significant digits in the fraction
    let isZero = true;

    for (let i = 0; i < fraction.length; i++) {
      if (fraction[i] !== "0") {
        let digits = i + fractionSignificantDigits;

        // round number of digits to the nearest multiple of 3, thousands
        const remainder = digits % 3;
        if (remainder !== 0) {
          digits = digits + 3 - remainder;
        }

        fraction = fraction.slice(0, digits);
        isZero = false;
        break;
      }
    }

    if (isZero) {
      fraction = "";
    }
    // remove trailing zeros
    // fraction = fraction.replace(/(0+)$/, "");
  }

  // split the fraction part into groups of 3 digits
  const groups: string[] = [];
  for (let i = 0; i < fraction.length; i += 3) {
    groups.push(fraction.slice(i, i + 3));
  }

  // remove trailing zeros groups
  for (let i = groups.length - 1; i > 0; i--) {
    if (groups[i] === "000") {
      groups.pop();
    } else {
      break;
    }
  }

  fraction = groups.join(thousandsSeparator);

  return `${negative ? "-" : ""}${integer || "0"}${fraction ? `.${fraction}` : ""}`;
}
