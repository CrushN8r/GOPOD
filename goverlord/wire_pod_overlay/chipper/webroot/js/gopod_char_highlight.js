// goverlord/wire_pod_overlay/chipper/webroot/js/gopod_char_highlight.js
//
// The per-character highlight (2026-11-02, operator: "lock in the highlighting
// mechanism * a single robust independent add-on process * per-character within
// a timed per-word highlight * the whole word appears 100% solid but each
// character turns 85% solid in sequence, within the timing of the word (word is
// 8 characters, spoken over 2 seconds ... 0.25s per character)").
//
// One job, on its own: the page lights a word (its .seq-unit gets "now", 100%
// solid - gopod_brobots_chat.js digitSequence) and hands it here with how long
// it's said for. Its letters then turn 85% one at a time, left to right, the last
// one as the word ends - so when the next word lights, this one is already the
// 85% every said word is. Pictures and formulas are left whole.
//
// Turn it off: GopodCharHighlight.settings.on = false (the words still light
// one at a time, as before). The look: css/gopod_char_highlight.css.

(function () {
  "use strict";

  var settings = {
    on: true,
    // A word said longer than this (a pause after it) spreads its letters over
    // this many seconds at most, so they don't crawl through the silence.
    max_word_seconds: 2.5
  };

  // The unit's letters as their own spans, made once. Grapheme-safe enough for
  // GOPOD's text: Array.from keeps an emoji or accented letter whole.
  function letters(unit) {
    if (unit._chLetters) { return unit._chLetters; }
    unit._chLetters = [];
    if (unit.tagName === "IMG" || unit.classList.contains("katex") || unit.children.length) { return unit._chLetters; }
    var text = unit.textContent || "";
    if (!text) { return unit._chLetters; }
    var frag = document.createDocumentFragment();
    Array.from(text).forEach(function (c) {
      var s = document.createElement("span");
      s.className = "ch";
      s.textContent = c;
      frag.appendChild(s);
      unit._chLetters.push(s);
    });
    unit.textContent = "";
    unit.appendChild(frag);
    return unit._chLetters;
  }

  function stop(unit) {
    (unit._chTimers || []).forEach(clearTimeout);
    unit._chTimers = [];
  }

  // play: the word is lit now, said for seconds - its letters turn 85% in turn.
  function play(unit, seconds) {
    if (!settings.on || !unit) { return; }
    var ls = letters(unit);
    if (!ls.length) { return; }
    stop(unit);
    ls.forEach(function (l) { l.classList.remove("ch-said"); });
    var total = Math.min(Math.max(+seconds || 0, 0), settings.max_word_seconds);
    var step = total / ls.length;
    unit._chTimers = ls.map(function (l, i) {
      return setTimeout(function () { l.classList.add("ch-said"); }, step * (i + 1) * 1000);
    });
  }

  window.GopodCharHighlight = { settings: settings, play: play, stop: stop };
})();
// ## EOF
