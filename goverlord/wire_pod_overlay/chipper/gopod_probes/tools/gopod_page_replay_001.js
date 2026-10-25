// goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_page_replay_001.js
// GOPOD's page replay (2026-10-24, operator: "if you don't need the brobots, do the
// full run without them"; "I want it kept as a tool in the repo please"): plays a
// run's feed lines through the real Brobots chat page code in a test browser
// engine (jsdom) - no robots, no Wire-Pod, no sound - and reports what the page
// drew and lit: every Easter egg bubble (its congratulations, zoom line, pictures,
// Generating), the order lines lit in, the Operator lines' voices, the prize
// lines, the bubbles by kind, and any page error, stuck line or loop.
//
//   gopod_page_replay.sh              - the built-in scenario (every egg kind, both prizes)
//   gopod_page_replay.sh <feed.json>  - a saved feed ({"boot":..,"events":[..]})
const fs = require("fs"), path = require("path");
const { JSDOM } = require("jsdom");
const W = process.argv[2], feedFile = process.argv[3];
// The built-in scenario: a question, its egg and its answer, for every egg kind;
// a photo with one picture not made (its hourglass lets go); both prizes; a
// newsletter answer (CHALK's QR).
function scenario() {
  const V = "/gopod_reporter_audio/", ev = [];
  let t = Date.now();
  const add = (who, text, more) => ev.push(Object.assign({ id: ev.length + 1, who, text, time_ms: (t += 50) }, more || {}));
  const op = (to, text) => {
    add("Operator", text, { to, via: "Keyboard" });
    const f = V + "replay_op_" + ev.length + ".wav";
    add("Reporter Voice", "generating:forlit:" + f + "|" + text);
    add("Reporter Voice", f);
    add("Reporter Voice", "heard:" + ev.length);
  };
  const egg = (who, name, prize, voice, zoom) => {
    const msg = "Congratulations!\nEaster Egg '" + name + "' found." + (zoom ? "\nMouseOver to Zoom In" : "");
    add("Picture", "popup:" + who + "::msg=" + msg + "::" + prize);
    add("Reporter Voice", "generating:forlit:" + V + voice + "|" + msg);
    add("Reporter Voice", V + voice);
    add("Reporter Voice", "heard:" + ev.length);
  };
  const reply = (who, text) => add(who, text, { via: "llm" });
  const line = (text, voice) => { add("Goverlord", "generating:lit:" + V + voice + "|" + text); add("Reporter Voice", V + voice); add("Reporter Voice", "heard:" + ev.length); };
  op("Brobot 1", "Brobot 1, tell me about bingo for seniors.");
  egg("Brobot 1", "Bingo and foodmath", "/pictures/chocolate_bingo_kit.jpg|/pictures/chocolate_bingo_card.jpg", "replay_egg_bingo.wav", true);
  reply("Brobot 1", "Sure, \u2063let's play! \u2063Say start Bingo test!");
  op("Brobots", "convert 0.272727 repeating to a fraction, explain the math");
  egg("Brobot 2", "AI Wordplay Explain the Math", "math:convert 0.272727 repeating to a fraction", "replay_egg_math.wav", false);
  reply("Brobot 2", "Let's do it! \u2063\u2022 Let x = 0.27 repeating.");
  reply("Brobot 1", "My bro! \u2063\u2022 So 0.27 repeating = 3\u204411.");
  op("Brobot 2", "what do you see?");
  egg("Brobot 2", "Get Photo", "/pictures/chocolate_bingo_kit.jpg|/pictures/chocolate_bingo_card.jpg|/photos/replay_not_made_yet_tpu.jpg", "replay_egg_photo.wav", true);
  reply("Brobot 2", "I see a person and a chair.");
  op("Brobot 1", "roast battle!");
  egg("Brobot 1", "Roast Battle", "card:\ud83d\udd25 Roast Battle \ud83d\udd25\nParty Brobots vs Any and All takers", "replay_egg_roast.wav", false);
  reply("Brobot 1", "Alright, bro! \u2063Let's get this party started!");
  line("Grand prize! Every value point revealed. Now GOPOD Yourself!", "replay_grand.wav");
  line("Master Grand Prize! Every value point and every Easter egg found, hidden ones too. Now GOPOD Yourself!", "replay_master.wav");
  op("Brobot 2", "Brobot 2, how do I keep up with you guys?");
  reply("Brobot 2", "Join the CRUSHN8R CREW'd newsletter at crushn8r.ca! \u2063the CREW'd newsletter! \u2063Get CREW'd.");
  return { boot: 1, events: ev };
}

let html = fs.readFileSync(W + "/index.html", "utf8").replace(/<script[\s\S]*?<\/script>/g, "");
const feed = feedFile ? JSON.parse(fs.readFileSync(feedFile, "utf8")) : scenario();
const upto = feed.events[feed.events.length - 1].id;
const t0 = Date.now(), log = [];
const stamp = () => ((Date.now() - t0) / 1000).toFixed(1) + "s";
let ticks = 0;
setInterval(() => { ticks = 0; }, 20);
const local = p => { const f = path.join(W, decodeURIComponent(p.split("?")[0])); return fs.existsSync(f) && fs.statSync(f).isFile() ? f : null; };
let played = 0;
// The display mirror's store - what Wire-Pod keeps (gopod_display.go).
const display = { v: 0, body: null };
const js = fs.readFileSync(W + "/js/gopod_brobots_chat.js", "utf8").replace(/\bwhile \(/g, "while (__g(), ");
// One page: "main" (plays the feed, sends the display) or "visitor" (paints it).
function makePage(role) {
  const dom = new JSDOM(html, { url: "http://localhost:8080/" + (role === "visitor" ? "?visitor" : ""), runScripts: "outside-only", pretendToBeVisual: true });
  const w = dom.window, d = w.document, notes = [];
  w.__g = function () { if (++ticks > 3e6) { console.log("LOOP (" + role + ")", new Error().stack.split("\n").slice(2, 8).join("\n")); process.exit(3); } };
  Object.defineProperty(w.HTMLElement.prototype, "offsetParent", { get() { return this.parentNode ? d.body : null; } });
  Object.defineProperty(w.HTMLImageElement.prototype, "complete", { get() { return true; } });
  Object.defineProperty(w.HTMLImageElement.prototype, "naturalWidth", { get() { return local(new URL(this.src, "http://x").pathname) ? 100 : 0; } });
  w.HTMLMediaElement.prototype.play = function () { const a = this; setTimeout(() => { a.dispatchEvent(new w.Event("ended")); }, 2000); return Promise.resolve(); };
  w.HTMLMediaElement.prototype.pause = function () {}; w.HTMLMediaElement.prototype.load = function () {};
  Object.defineProperty(w.HTMLMediaElement.prototype, "duration", { get() { return 2; } });
  Object.defineProperty(w.HTMLMediaElement.prototype, "readyState", { get() { return 4; } });
  w.Element.prototype.scrollTo = function () {}; w.scrollTo = function () {}; w.Element.prototype.scrollIntoView = function () {};
  w.matchMedia = () => ({ matches: false, addListener() {}, addEventListener() {} });
  w.fetch = function (url, opt) {
    const full = String(url).replace(/^https?:\/\/[^/]+/, ""), p = full.split("?")[0];
    let body = null, ok = true, status = 200;
    if (p === "/gopod_chat_feed.json") { body = JSON.stringify(feed); }
    else if (p === "/api-sdk/gopod_page_shown") { body = JSON.stringify(role === "main" ? { main: true, admin: true } : { main: false, admin: false, visitor: { state: "on", name: "visitor" } }); }
    else if (p === "/api-sdk/gopod_page_note") { notes.push(decodeURIComponent(String(opt && opt.body)).replace(/\+/g, " ")); body = "ok"; }
    else if (p === "/api-sdk/gopod_page_played") { const m = /id=(\d+)/.exec(String(opt && opt.body)); if (m && role === "main") { played = Math.max(played, +m[1]); } body = "ok"; }
    else if (p === "/api-sdk/gopod_display") {
      if (opt && opt.method === "POST") { if (role === "main") { const j = JSON.parse(opt.body); j.v = ++display.v; display.body = JSON.stringify(j); } body = "ok"; }
      else { const since = +((/since=(\d+)/.exec(full) || [, 0])[1]); if (!display.body || display.v === since) { status = 204; body = ""; } else { body = display.body; } }
    }
    else if (local(p)) { body = fs.readFileSync(local(p), "utf8"); }
    else if (/\.json$/.test(p)) { ok = false; status = 404; body = "{}"; }
    else { body = "{}"; }
    return Promise.resolve({ ok, status, json: () => Promise.resolve(JSON.parse(body)), text: () => Promise.resolve(body) });
  };
  // Pictures: loaded or failed as on disk, again on every new src.
  new w.MutationObserver(ch => ch.forEach(m => {
    const imgs = m.type === "attributes" ? [m.target] : Array.from(m.addedNodes).flatMap(n => n.querySelectorAll ? [n].concat(Array.from(n.querySelectorAll("img"))) : []);
    imgs.filter(i => i.tagName === "IMG").forEach(i => setTimeout(() => i.dispatchEvent(new w.Event(local(new URL(i.src, "http://x").pathname) ? "load" : "error")), 30));
  })).observe(d, { subtree: true, childList: true, attributes: true, attributeFilter: ["src"] });
  if (role === "main") {
    // Lighting: when each voiced part starts and ends.
    new w.MutationObserver(ch => ch.forEach(m => {
      const el = m.target;
      if (!el.classList || !el.classList.contains("brobot-say")) { return; }
      const b = el.closest(".brobot-bubble"); if (!b || b.dataset.kind !== "egg") { return; }
      const state = el.classList.contains("speaking") ? "lit" : el.classList.contains("spoken") ? "said" : "";
      if (state && el._last !== state) { el._last = state; log.push(stamp() + " egg '" + (b.textContent.match(/Easter Egg '([^']*)'/) || [, "?"])[1] + "' " + (el.closest(".popup-zoom-line") ? "zoom line" : "congratulations") + " " + state); }
    })).observe(d, { subtree: true, attributes: true, attributeFilter: ["class"] });
  }
  w.addEventListener("error", e => notes.push("PAGE ERROR (" + role + ") " + e.message));
  try { w.eval(js); } catch (e) { console.log("LOAD ERROR (" + role + ")", e.stack.split("\n").slice(0, 4).join("\n")); process.exit(2); }
  d.dispatchEvent(new w.Event("DOMContentLoaded"));
  const sec = d.getElementById("section-log"); if (sec) { sec.style.display = ""; }
  const c = d.getElementById("brobot-chat-bubbles-container"); if (c) { c.classList.add("expanded"); }
  setTimeout(() => { d.dispatchEvent(new w.Event("pointerdown", { bubbles: true })); }, 300);
  return { w, d, c, notes };
}
const main = makePage("main"), visitor = makePage("visitor");
const c = main.c, notes = main.notes;
// The mirror, checked every quarter second: the visitor's bubbles against the
// main page's (the ones it sends) - their words and what's lit. A difference
// that lasts over mirror_lag_ms is noted once.
const MIRROR_LAG_MS = 1500;
const shown = (cont, visitorSide) => Array.from(cont.children).filter(el => el.classList && el.classList.contains("brobot-bubble") &&
  (visitorSide || (el.dataset.staged !== "1" && el.style.visibility !== "hidden" && !el.classList.contains("gopod-admin-only")))).slice(-12);
const look = el => el.textContent.replace(/\s+/g, " ").trim().slice(0, 70) + " [" + Array.from(el.querySelectorAll(".brobot-say")).map(s => s.classList.contains("speaking") ? "L" : s.classList.contains("spoken") ? "s" : "-").join("") + "]";
let mirrorChecks = 0, mirrorOff = 0, offSince = 0, offNoted = {}, lagMost = 0;
const mirrorProblems = [];
setInterval(() => {
  const a = shown(c, false).map(look), b = shown(visitor.c, true).map(look);
  mirrorChecks++;
  const same = a.length === b.length && a.every((x, i) => x === b[i]);
  if (same) { if (offSince) { lagMost = Math.max(lagMost, Date.now() - offSince); } offSince = 0; return; }
  mirrorOff++;
  if (!offSince) { offSince = Date.now(); }
  if (Date.now() - offSince > MIRROR_LAG_MS) {
    const i = a.findIndex((x, k) => x !== b[k]), key = i + "|" + (a[i] || "");
    if (!offNoted[key]) { offNoted[key] = true; mirrorProblems.push(stamp() + " bubble " + (i + 1) + " of " + a.length + ": main \"" + (a[i] || "(none)") + "\" / visitor \"" + (b[i] || "(none)") + "\""); }
  }
}, 250);
let last = 0, since = Date.now();
setInterval(() => {
  if (played !== last) { last = played; since = Date.now(); }
  const done = played >= upto && Date.now() - since > 6000;
  if (!done && Date.now() - since < 120000) { return; }
  console.log("PLAYED " + played + " of " + upto + (played >= upto ? "" : "  <-- STOPPED"));
  console.log("\nEGGS");
  c.querySelectorAll('[data-kind="egg"]').forEach(b => {
    const congrats = b.querySelector(".popup-congrats .brobot-say, .brobot-bubble-text .brobot-say");
    const zoom = b.querySelector(".popup-zoom-line .brobot-say");
    console.log("  " + (b.textContent.match(/Easter Egg '([^']*)'/) || [, "?"])[1] + ": " + b.className.replace("brobot-bubble ", "") +
      " | congratulations " + (congrats ? congrats.className.replace("brobot-say", "").trim() || "dormant" : "MISSING") +
      " | zoom " + (zoom ? zoom.className.replace("brobot-say", "").trim() || "dormant" : "-") +
      " | pictures " + b.querySelectorAll("img:not(.picture-waiting)").length + " shown, " + b.querySelectorAll(".picture-wait").length + " waiting" +
      " | Generating left " + b.querySelectorAll(".voice-wait").length +
      " | reply joined " + (b.querySelector(":scope > .brobot-bubble-text") ? "yes" : "no"));
  });
  console.log("\nLIGHTING ORDER"); log.forEach(l => console.log("  " + l));
  console.log("\nOPERATOR LINES (their own voice)");
  c.querySelectorAll(".brobot-bubble.operator[data-own='1']").forEach(b => { const s = b.querySelector(".brobot-say"); console.log("  " + (b.querySelector(".brobot-bubble-text") || b).textContent.trim().slice(0, 60) + " | " + (s ? (s.className.replace("brobot-say", "").trim() || "dormant") : "NOT TIED to its voice")); });
  console.log("\nGOVERLORD LINES");
  Array.from(c.querySelectorAll(".brobot-bubble")).filter(b => /Grand Prize/i.test(b.textContent) && b.dataset.kind !== "egg").forEach(b => console.log("  " + b.textContent.replace(/\s+/g, " ").trim().slice(0, 90) + " | " + Array.from(b.querySelectorAll(".brobot-say")).map(s => s.className.replace("brobot-say", "").trim() || "dormant").join(",")));
  console.log("\nBUBBLES BY KIND"); const k = {}; c.querySelectorAll(".brobot-bubble").forEach(b => { k[b.dataset.kind || b.className.split(" ").slice(1, 3).join(".")] = (k[b.dataset.kind || b.className.split(" ").slice(1, 3).join(".")] || 0) + 1; }); console.log("  " + JSON.stringify(k));
  console.log("\nVISITOR MIRROR"); console.log("  " + mirrorChecks + " checks, in step " + Math.round(100 * (mirrorChecks - mirrorOff) / Math.max(1, mirrorChecks)) + "%, longest catch-up " + (lagMost / 1000).toFixed(1) + " s" + (offSince ? ", OUT OF STEP at the end" : ""));
  mirrorProblems.slice(0, 8).forEach(m => console.log("  behind " + (MIRROR_LAG_MS / 1000) + " s+: " + m));
  console.log("\nPAGE NOTES (errors, stuck, baton)"); notes.concat(visitor.notes).filter(n => /error|stuck|baton|NOT shown/i.test(n)).forEach(n => console.log("  " + n.slice(0, 200)));
  process.exit(0);
}, 500);
