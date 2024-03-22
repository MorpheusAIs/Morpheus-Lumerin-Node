/**
 * Converts a stringified JSON array to a JS array
 * @param {string} str 
 * @returns {string[]}
 */
function parseJSONArray(str) {
  let parsed;
  try {
    parsed = JSON.parse(str);
    if (!Array.isArray(parsed)) {
      throw null
    }
  } catch (err) {
    throw new Error('not a valid JSON array');
  }
  return parsed;
}

module.exports = {
  parseJSONArray,
};