// 2026-11-16: the lingo improver page (gopod_lingo_improver.html) - its voices and
// words from Wire-Pod (/api-sdk/gopod_lingo), nothing written here.
(function () {
  var $ = function (id) { return document.getElementById(id); };
  var texts = {}, voices = [];
  function t(k, kv) { var s = texts[k] || ""; Object.keys(kv || {}).forEach(function (x) { s = s.replace("{" + x + "}", kv[x]); }); return s; }
  function voice() { return voices.filter(function (v) { return v.voice === $("liVoice").value; })[0] || {}; }
  function keepLabel() { $("liKeep").textContent = t("keep", { lingo: voice().lingo || "" }); }
  function post(body) {
    var p = new URLSearchParams(); Object.keys(body).forEach(function (k) { p.set(k, body[k]); });
    return fetch("/api-sdk/gopod_lingo", { method: "POST", body: p }).then(function (r) { return r.json(); });
  }
  fetch("/api-sdk/gopod_lingo", { cache: "no-store" }).then(function (r) { return r.json(); }).then(function (d) {
    texts = d.texts || {}; voices = d.voices || [];
    document.title = "GOPOD " + t("title");
    $("liTitle").textContent = t("title"); $("liIntro").textContent = t("intro");
    $("liVoiceLabel").textContent = t("voice"); $("liTextLabel").textContent = t("text");
    $("liText").placeholder = t("placeholder"); $("liSay").textContent = t("say");
    $("liSaidLabel").textContent = t("as_said") + ":"; $("liWordLabel").textContent = t("word");
    $("liSpelledLabel").textContent = t("spelled");
    voices.forEach(function (v) {
      var o = document.createElement("option"); o.value = v.voice; o.textContent = v.label + " (" + v.lingo + ")";
      $("liVoice").appendChild(o);
    });
    keepLabel();
  });
  $("liVoice").addEventListener("change", keepLabel);
  function say() {
    $("liNote").textContent = "";
    post({ action: "try", voice: $("liVoice").value, text: $("liText").value }).then(function (d) {
      $("liSaid").textContent = d.said || "";
      $("liNote").textContent = d.note || "";
      if (d.audio) { $("liAudio").src = d.audio; $("liAudio").play().catch(function () {}); }
    });
  }
  $("liSay").addEventListener("click", say);
  $("liText").addEventListener("keydown", function (e) { if (e.key === "Enter") { say(); } });
  $("liKeep").addEventListener("click", function () {
    post({ action: "keep", voice: $("liVoice").value, word: $("liWord").value, said: $("liSpelled").value }).then(function (d) {
      $("liNote").textContent = d.note || "";
    });
  });
})();
