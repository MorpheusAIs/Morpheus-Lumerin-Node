import { zxcvbn } from 'zxcvbn3';

export const MaxScore = 4;
const defaultSuggestion = 'Add another word or two. Uncommon words are better.';

const scoreMaxGuessesMap = new Map([
  [0, 10 ** 4],
  [1, 10 ** 8],
  [2, 10 ** 12],
  [3, 10 ** 16],
  [4, Infinity]
]);

/**
 *
 * @param {number} guesses
 * @returns {number}
 */
const mapGuessesToScore = guesses => {
  for (const [score, maxGuesses] of scoreMaxGuessesMap) {
    if (guesses < maxGuesses) {
      return score;
    }
  }
};

/**
 * Returns password strength data
 * @param {string} password password to test
 * @returns {ScoreResult}
 *
 * @typedef {Object} ScoreResult
 * @property {number} score from 0 to 4
 * @property {boolean} isStrong true for strong password
 * @property {string[]} suggestions array of password suggestions
 */
export const GetPasswordStrength = password => {
  if (!password) {
    password = '';
  }
  const res = zxcvbn(password);
  const score = res?.guesses ? mapGuessesToScore(res.guesses) : undefined;
  const suggestions = res?.feedback?.suggestions || [];
  const useDefaultSuggestion = suggestions.length === 0 && score < MaxScore;

  return {
    score: score,
    suggestions: useDefaultSuggestion ? [defaultSuggestion] : suggestions,
    isStrong: score === MaxScore
  };
};

/**
 *
 * @param {string} password
 * @returns {boolean}
 */
export const IsPasswordStrong = password =>
  GetPasswordStrength(password).isStrong;
