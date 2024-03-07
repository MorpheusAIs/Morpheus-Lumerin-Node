/**
 * Converts comma-separated decimal values into point-separated decimal values
 * Useful to deal with locales with comma-separated decimals
 *
 * @param {string} amount - The string to sanitize
 */
export function sanitize(amount = '') {
  return amount.replace(',', '.');
}

/**
 * Removes extra spaces and converts to lowercase
 * Useful for sanitizing user input before recovering a wallet.
 *
 * @param {string} str - The string to sanitize
 */
export function sanitizeMnemonic(str) {
  return str
    .replace(/\s+/gi, ' ')
    .trim()
    .toLowerCase();
}

/**
 * Removes extra spaces, tabs and line breaks.
 * Useful for sanitizing pasted texts that may contain
 * invalid breaks.
 *
 * @param {string} str - The string to sanitize
 */
export function sanitizeInput(str) {
  return typeof str === 'string' ? str.replace(/\s+|\n+|\t+/g, '') : str;
}
