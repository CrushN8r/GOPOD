
(function () {
  "use strict";

  var settings = {
    on: true,
    max_word_seconds: 2.5
  };

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
