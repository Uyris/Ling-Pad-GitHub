'use strict';

const editor       = document.getElementById('editor');
const statChars    = document.getElementById('stat-chars');
const statWords    = document.getElementById('stat-words');
const statSentences = document.getElementById('stat-sentences');
const statAvgWord  = document.getElementById('stat-avg-word');
const topWordsList = document.getElementById('top-words');
const btnClear     = document.getElementById('btn-clear');
const btnExport    = document.getElementById('btn-export');

/**
 * Return an array of lowercase words extracted from a string,
 * stripping punctuation.
 * @param {string} text
 * @returns {string[]}
 */
function extractWords(text) {
  return text
    .toLowerCase()
    .replace(/[^\w\s'-]/g, '')
    .split(/\s+/)
    .filter(Boolean);
}

/**
 * Count sentences using basic punctuation heuristic.
 * @param {string} text
 * @returns {number}
 */
function countSentences(text) {
  const trimmed = text.trim();
  if (!trimmed) return 0;
  const matches = trimmed.match(/[^.!?]*[.!?]+/g);
  const counted = matches ? matches.length : 0;
  // Count a trailing fragment that has no closing punctuation
  const lastPunct = trimmed.search(/[.!?][^.!?]*$/);
  const hasTrailing = lastPunct === -1
    ? counted === 0  // no punctuation at all → 1 sentence
    : !/[.!?]$/.test(trimmed);
  return counted + (hasTrailing ? 1 : 0);
}

/**
 * Return the top N most frequent words as an array of [word, count] pairs.
 * @param {string[]} words
 * @param {number} n
 * @returns {[string, number][]}
 */
function topWords(words, n = 5) {
  const freq = {};
  for (const w of words) {
    freq[w] = (freq[w] || 0) + 1;
  }
  return Object.entries(freq)
    .sort((a, b) => b[1] - a[1])
    .slice(0, n);
}

/** Update all statistics displayed in the sidebar. */
function updateStats() {
  const text  = editor.value;
  const words = extractWords(text);

  statChars.textContent    = text.length;
  statWords.textContent    = words.length;
  statSentences.textContent = countSentences(text);

  const avgLen = words.length
    ? (words.reduce((sum, w) => sum + w.length, 0) / words.length).toFixed(1)
    : '0';
  statAvgWord.textContent = avgLen;

  const top = topWords(words);
  topWordsList.innerHTML = top
    .map(([word, count]) => `<li>${word} <em>(${count})</em></li>`)
    .join('');
}

/** Clear the editor after confirmation. */
function clearEditor() {
  if (editor.value && !confirm('Clear all text?')) return;
  editor.value = '';
  updateStats();
  editor.focus();
}

/** Download the current text as a UTF-8 .txt file. */
function exportText() {
  const blob = new Blob([editor.value], { type: 'text/plain;charset=utf-8' });
  const url  = URL.createObjectURL(blob);
  const a    = document.createElement('a');
  a.href     = url;
  a.download = 'ling-pad-export.txt';
  a.click();
  URL.revokeObjectURL(url);
}

editor.addEventListener('input', updateStats);
btnClear.addEventListener('click', clearEditor);
btnExport.addEventListener('click', exportText);

// Initialise stats on load (in case of browser-restored content)
updateStats();
