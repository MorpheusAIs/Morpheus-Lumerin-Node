/**
 * One exact-snippet replacement inside a text file.
 *
 * Rewriting a whole file to change three lines is how a model loses the other
 * nine hundred: it has to reproduce from memory everything it did not mean to
 * touch, and a single dropped block is silent. Sending only the snippet that
 * changes removes that risk, and it removes the 2 MB write ceiling as a reason
 * a large file cannot be edited at all.
 *
 * The snippet has to match exactly once. A snippet that matches nowhere is a
 * model working from a stale read, and a snippet that matches twice is a model
 * that has not said which one it means; guessing either way edits the wrong
 * lines and reports success. Both are refused with the count, which is enough
 * for the model to read again or widen the snippet.
 */

export interface CoworkFileEditResult {
  /** The whole file after the replacement, ready to be written back. */
  content: string
  /** 1-based line where the replaced snippet began. */
  startLine: number
  /** Lines the snippet spanned before the edit. */
  removedLines: number
  /** Lines the replacement spans after the edit. */
  addedLines: number
  /**
   * True when the file uses CRLF and the submitted snippet used bare newlines.
   * The snippet was matched and the replacement written in the file's own line
   * endings, so an edit never silently converts half a file to LF.
   */
  normalizedLineEndings: boolean
}

const countOccurrences = (haystack: string, needle: string): number => {
  let count = 0
  let index = haystack.indexOf(needle)
  while (index !== -1) {
    count += 1
    index = haystack.indexOf(needle, index + needle.length)
  }
  return count
}

const lineOf = (source: string, index: number): number =>
  source.slice(0, index).split('\n').length

/**
 * Replaces the single occurrence of `oldText` with `newText`, or throws with a
 * message the model can act on. Pure, so the rules are tested without touching
 * a disk or an Electron app object.
 */
export function applyCoworkFileEdit(
  source: string,
  oldText: string,
  newText: string
): CoworkFileEditResult {
  if (!oldText) {
    throw new Error('edit_file requires oldText. Use write_file to create a file from nothing.')
  }
  if (oldText === newText) {
    throw new Error('edit_file received identical oldText and newText, so there is nothing to do.')
  }

  let searchFor = oldText
  let replaceWith = newText
  let normalizedLineEndings = false
  let occurrences = countOccurrences(source, searchFor)

  // A model reading a CRLF file through read_file sees the lines, not the
  // carriage returns, and sends the snippet back with bare newlines. Matching
  // in the file's own endings is the difference between an edit that works and
  // an unexplained "snippet not found" on every Windows-authored file.
  if (occurrences === 0 && source.includes('\r\n') && !oldText.includes('\r\n')) {
    const crlfOld = oldText.replaceAll('\n', '\r\n')
    const crlfOccurrences = countOccurrences(source, crlfOld)
    if (crlfOccurrences > 0) {
      searchFor = crlfOld
      replaceWith = newText.replaceAll('\n', '\r\n')
      occurrences = crlfOccurrences
      normalizedLineEndings = true
    }
  }

  if (occurrences === 0) {
    throw new Error(
      'edit_file found no match for oldText. Read the file again and copy the snippet exactly, including indentation.'
    )
  }
  if (occurrences > 1) {
    throw new Error(
      `edit_file found ${occurrences} matches for oldText and will not guess which one you mean. ` +
        'Include more surrounding lines so the snippet appears exactly once.'
    )
  }

  const index = source.indexOf(searchFor)
  return {
    content: source.slice(0, index) + replaceWith + source.slice(index + searchFor.length),
    startLine: lineOf(source, index),
    removedLines: searchFor.split('\n').length,
    addedLines: replaceWith ? replaceWith.split('\n').length : 0,
    normalizedLineEndings
  }
}
