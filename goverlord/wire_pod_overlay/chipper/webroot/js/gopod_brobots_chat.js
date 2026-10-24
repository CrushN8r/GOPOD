// goverlord/wire_pod_overlay/chipper/webroot/js/gopod_brobots_chat.js
// GOPOD's Brobots chat display for Wire-Pod's web page (index.html's LOGS
// section): Copy Logs, the Brobots Chat / Clear Chat buttons, the chat bubbles,
// the spoken-sentence highlight and the chat auto-scroll. Moved out of
// index.html 2026-09-30 (operator: "offload as much as possible to external css
// and js files"), unchanged apart from the move. Loaded by index.html at the
// same spot the inline script sat, right after the chat's own markup.

    function copyText(text) {
        const tmp = document.createElement("textarea");
        tmp.value = text;
        document.body.appendChild(tmp);
        tmp.select();
        document.execCommand("copy");
        tmp.remove();
    }

    function copyToClipboard(all) {
        // 2026-10-01 (operator: "fast track the copy to all logs"): Copy All Logs
        // copies Wire-Pod's full log (the "Show all logs" one, /api/get_debug_logs)
        // straight away - no need to close the chat or tick the box first.
        if (all) {
            fetch("/api/get_debug_logs", { cache: "no-store" })
                .then(function (r) { return r.text(); })
                .then(function (text) {
                    copyText(text);
                    alert("Copied all logs (" + text.split("\n").length + " lines).");
                })
                .catch(function () { alert("Couldn't read all logs from Wire-Pod."); });
            return;
        }
        // 2026-09-28: when the Brobots chat-bubbles window is showing, copy its
        // bubbles instead of the hidden log text - so this one button copies
        // whichever of the three displays is on screen: Logs, all logs, or bubbles.
        const bubbles = document.getElementById("brobot-chat-bubbles-container");
        if (bubbles && bubbles.classList.contains("expanded")) {
            const text = Array.prototype.map.call(
                bubbles.querySelectorAll(".brobot-bubble-text"),
                function (el) { return el.dataset.raw || el.textContent; }
            ).join("\n");
            copyText(text);
            alert("Copied the text: " + text);
            return;
        }
        const botTranscriptedTextArea = document.getElementById("botTranscriptedTextArea");
        botTranscriptedTextArea.select();
        botTranscriptedTextArea.setSelectionRange(0, 99999); // For mobile devices
        document.execCommand("copy");

        alert("Copied the text: " + botTranscriptedTextArea.value);
    }

    // 2026-08-08: chat-bubbles container. 2026-09-30: shown by default (operator:
    // "Brobots Chat Window is now ON by default. Only the button hides/unhides
    // Brobots Chat Window (with display content still intact)").
    (function () {
      var container = document.getElementById("brobot-chat-bubbles-container");
      var textarea = document.getElementById("botTranscriptedTextArea");
      // 2026-09-30: the chat keeps its own copy (operator: "survives a screen saver
      // event, even survives a page refresh, contents only cleared upon button
      // clicked"). Wire-Pod's /api/get_logs only returns its last 50 lines, so the
      // page can't rebuild history from it - every bubble is saved in this browser
      // (localStorage) and redrawn on load; a new run adds below instead of
      // wiping; only Clear Chat empties it. The log lines already turned into
      // bubbles are remembered too (as short hashes), so a reload never shows them
      // twice. Whether the chat was open is remembered as well. Any storage
      // failure (private window, storage full) falls back to the old in-page
      // behaviour instead of breaking the chat.
      var CHAT_STORE_KEY = "gopod_brobots_chat_v1";
      var CHAT_OPEN_KEY = "gopod_brobots_chat_open";
      var CHAT_MAX_BUBBLES = 3000;
      var CHAT_MAX_SEEN = 6000;
      var chatStore = { bubbles: [], seen: [] };
      var seenLines = {};
      var restoringChat = false;
      function lineHash(s) {
        var h = 5381;
        for (var i = 0; i < s.length; i++) { h = ((h * 33) ^ s.charCodeAt(i)) >>> 0; }
        return h.toString(36) + "." + s.length.toString(36);
      }
      function isSeen(key) { return !!seenLines[lineHash(key)]; }
      function markSeen(key) {
        var h = lineHash(key);
        if (seenLines[h]) { return; }
        seenLines[h] = true;
        chatStore.seen.push(h);
        if (chatStore.seen.length > CHAT_MAX_SEEN) {
          chatStore.seen.splice(0, chatStore.seen.length - CHAT_MAX_SEEN).forEach(function (old) { delete seenLines[old]; });
        }
        saveChat();
      }
      function saveChat() {
        try {
          window.localStorage.setItem(CHAT_STORE_KEY, JSON.stringify(chatStore));
        } catch (e) {
          // Storage full: keep the newer half and try once more.
          chatStore.bubbles = chatStore.bubbles.slice(Math.floor(chatStore.bubbles.length / 2));
          try { window.localStorage.setItem(CHAT_STORE_KEY, JSON.stringify(chatStore)); } catch (e2) { /* in-page only */ }
        }
      }
      function recordBubble(who, text, format, to, via) {
        var saved = { who: who, text: text, format: format || "" };
        if (to) { saved.to = to; }
        if (via) { saved.via = via; }
        chatStore.bubbles.push(saved);
        if (chatStore.bubbles.length > CHAT_MAX_BUBBLES) {
          chatStore.bubbles.splice(0, chatStore.bubbles.length - CHAT_MAX_BUBBLES);
        }
        saveChat();
      }
      function loadChat() {
        try {
          var saved = JSON.parse(window.localStorage.getItem(CHAT_STORE_KEY) || "null");
          if (saved && Array.isArray(saved.bubbles) && Array.isArray(saved.seen)) { chatStore = saved; }
        } catch (e) { /* nothing saved, or unreadable - start empty */ }
        seenLines = {};
        chatStore.seen.forEach(function (h) { seenLines[h] = true; });
      }
      function restoreChat() {
        restoringChat = true;
        chatStore.bubbles.forEach(function (b) {
          if (b.format === "ball") { showBingoBall(b.text); }
          else if (b.format === "llm_ready") { restoreWarmupBubble(b.text); }
          else if (b.format === "chapter") { showChapter(b.text); }
          else if (b.format === "score") { showScore(b.text); }
          else if (b.format === "picture") { showPictures(b.text); }
          else if (b.format === "host") { showHostLine(b.who, b.text); }
          else if (b.format === "narrator") { showNarratorLine(b.text); }
          else { appendBubble(b.who, b.text, b.format, b.to, b.via); }
        });
        restoringChat = false;
        lastOperatorBubble = null;
        // 2026-10-16 (comb-through: the marquee stuck after a reload): a restored
        // egg's exchange is over.
        popupExchange = false;
        popupWraps = {};
        // A restored chat has all been heard (2026-10-15).
        Array.prototype.forEach.call(container.querySelectorAll(".brobot-say"), function (s) { s.classList.add("spoken"); });
        container.scrollTop = container.scrollHeight;
      }
      function setChatOpen(open) {
        try { window.localStorage.setItem(CHAT_OPEN_KEY, open ? "1" : "0"); } catch (e) { /* in-page only */ }
      }
      // Shown unless the viewer hid it with the button (remembered per browser,
      // until the next Wire-Pod start - see pollFeed).
      function chatWasOpen() {
        try { return window.localStorage.getItem(CHAT_OPEN_KEY) !== "0"; } catch (e) { return true; }
      }
      // 2026-10-04 (operator: "At the bottom of the robot's chat bubble ... add a
      // line for keyboard input at the very left ... radio buttons to select robot
      // 1 or 2. And then a text input where I can type a command rather than
      // speak"): one line at the bottom of the chat window. Enter sends the line to the chosen
      // Brobot (POST /api-sdk/gopod_typed_input, gopod_typed_input.go): a voice
      // command runs, anything else is a chat question. The choice of Brobot
      // starts on All at every page load (see savedRobot below).
      // Which Brobot each robot ESN is (also used for the "zzz" bubble below).
      var ESN_TO_BROBOT = { "0dd1b9e9": "Brobot 1", "0dd1d8bf": "Brobot 2" };
      // The admin (2026-10-23, operator: "System Window for admin users only
      // (me)"): the page opened on GOPOD's own machine. Other devices are
      // visitors (an admin PIN for them comes with C5).
      // Wire-Pod says which page is the admin (gopod_main_page.go): one opened on
      // GOPOD's own machine, by localhost or its network address. Until it says,
      // no page is.
      var isAdminPage = false;
      var typedRow = document.createElement("form");
      typedRow.className = "gopod-typed-input";
      typedRow.setAttribute("autocomplete", "off");
      // 2026-10-09 (operator: "default to ALL brobots radio button"): every page
      // load starts on All; a choice holds until the page is reloaded.
      var savedRobot = "all";
      // 2026-10-04 (operator: "please add a 3rd All brobots () radio button"):
      // both Brobots get the line, as when both hear it spoken (gopod_typed_input.go).
      // 2026-10-08 (operator: "Brobots ()1 ()2 ()All | Numpad? [] | Record? [] |
      // 00:00:00"): the line's layout.
      var TYPED_TARGETS = { "0dd1b9e9": "1", "0dd1d8bf": "2", "all": "All" };
      function typedRowText(text, className) {
        var span = document.createElement("span");
        span.className = className;
        span.textContent = text;
        typedRow.appendChild(span);
      }
      typedRowText("Brobots", "gopod-typed-name");
      // 2026-10-22 (operator: "replace 'All' radio button with a small drop down
      // menu: 1/2 * default selection = '1' (brobot 1) * selection '2' = (brobot 2)
      // * template drop down menu selection with default brobot 1/'Starting with
      // Brobot 2'"): All is a small 1/2 menu - who answers first (typed_line.all_first,
      // the same rule as "Starting with Brobot 2"); choosing it is choosing All.
      var allFirstSelect = null;
      Object.keys(TYPED_TARGETS).forEach(function (esn) {
        var label = document.createElement("label");
        var radio = document.createElement("input");
        radio.type = "radio";
        radio.name = "gopod-typed-robot";
        radio.value = esn;
        radio.checked = esn === savedRobot;
        label.appendChild(radio);
        if (esn === "all") {
          radio.className = "gopod-all-radio";
          label.className = "gopod-all-pick";
          label.appendChild(document.createTextNode(TYPED_TARGETS[esn]));
          allFirstSelect = document.createElement("select");
          allFirstSelect.className = "gopod-all-first";
          allFirstSelect.title = "All: who answers first";
          ["1", "2"].forEach(function (n) {
            var o = document.createElement("option");
            o.value = n;
            o.textContent = n;
            allFirstSelect.appendChild(o);
          });
          allFirstSelect.value = "1";
          ["focus", "change", "mousedown"].forEach(function (ev) {
            allFirstSelect.addEventListener(ev, function () { radio.checked = true; });
          });
          label.appendChild(allFirstSelect);
        } else {
          label.appendChild(document.createTextNode(TYPED_TARGETS[esn]));
        }
        typedRow.appendChild(label);
      });
      // A line to All, sent with Brobot 2 first: its Operator icons in that order.
      var allSecondFirst = {};
      typedRowText("|", "gopod-typed-sep");
      // 2026-10-19 (operator: "New Operator Line: Brobots ()1 ()2 ()All [Type to the
      // Brobot, press Enter] | Record? [] | HH:MM:SS 🔇 * make Native?[] Numpad?[]
      // part of system- commands * prune Native?[] & Numpad?[] altogether"): the
      // two boxes are gone - system-native on|off and system-numpad on|off
      // (gopod_system_edit.go); the numpad's state is kept per browser as before.
      var NUMPAD_KEY = "gopod_brobots_numpad_on";
      var typedFirstSep = typedRow.lastElementChild;
      // 2026-10-04 (operator: "after 'All Brobots' checkbox add a 'Record Session'
      // checkbox"; spec gopod_notes/RECORD_SESSION_SPEC_003.md): ticked records the
      // session from the next safe point, unticked stops and keeps it zipped. The
      // box follows Wire-Pod's own state (gopod_record_session.json), so it unticks
      // itself once a "Thanks Brobots" ending is answered.
      var recordLabel = document.createElement("label");
      recordLabel.className = "gopod-record-session";
      var recordBox = document.createElement("input");
      recordBox.type = "checkbox";
      recordBox.addEventListener("change", function () {
        fetch("/api-sdk/gopod_record_session", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "on=" + (recordBox.checked ? "1" : "0")
        }).catch(function () { /* the next poll shows the real state */ });
      });
      recordLabel.appendChild(document.createTextNode("Record?"));
      recordLabel.appendChild(recordBox);
      typedRow.appendChild(recordLabel);
      typedRowText("|", "gopod-typed-sep");
      // 2026-10-07 (operator: "add a session timer after '[] Record Session'
      // 00:00:00 (hh:mm:ss)"): it counts from page load; when recording starts it
      // restarts from Wire-Pod's own start time (gopod_record_session.json
      // started_ms), red; when it stops it goes back to 0 and keeps counting
      // (showSessionTimer).
      var sessionTimer = document.createElement("span");
      sessionTimer.className = "gopod-session-timer";
      sessionTimer.textContent = "00:00:00";
      typedRow.appendChild(sessionTimer);
      // 2026-10-08 (operator: "timer should start ticking immediately"; "upon
      // recording on/off the last thing is the timer is reset to 0 and starts
      // counting again"): it counts from the page load, and starts again from 0
      // when recording starts and when it stops.
      var timerFromMs = Date.now(), recordStartedMs = 0, wasRecording = null;
      function showSessionTimer() {
        var ms = Date.now() - timerFromMs;
        var sec = Math.max(0, Math.floor(ms / 1000));
        var two = function (n) { return (n < 10 ? "0" : "") + n; };
        sessionTimer.textContent = two(Math.floor(sec / 3600)) + ":" + two(Math.floor(sec / 60) % 60) + ":" + two(sec % 60);
        sessionTimer.classList.toggle("recording", !!recordStartedMs);
      }
      setInterval(showSessionTimer, 1000);
      // 2026-10-13 (operator: "a 6s loop * I want to run the loop in the
      // background very quiet, try 10% volume * keyboard line record/session
      // clock: upon any/every reset to 00:00:00 * ensure loop runs every 6s"):
      // background_loop's file played over and over with no gap (Web Audio), at
      // its volume, from its top again at every 00:00:00. A browser plays sound
      // only after a click, so it starts at the page's first click if need be.
      // 2026-10-13 (operator: "add a speaker on/off toggle icon thingy, after the
      // keyboard input text field"; "all songs (scripted runs) * remember speaker
      // toggle * toggle speaker (loop audio) off * after run, last thing is to
      // set back to remembered toggle"): the speaker toggle (kept in this
      // browser) turns the loop on or off; a run quiets it (Wire-Pod's
      // loop_pause line) and its end sets it back to the toggle (loop_resume).
      var backgroundLoop = (function () {
        var volumeCfg = 0.14;
        var file = "", volume = 0.14, ctx = null, buffer = null, source = null, gain = null, wanted = false;
        var speakerOn = true, runQuiet = false;
        try { speakerOn = localStorage.getItem("gopodLoopSpeaker") !== "off"; } catch (e) {}
        // 2026-10-17: volume is a share of the master (chat_audio.volume), evened
        // like every sound (audioLevel).
        function level() { return speakerOn && !runQuiet ? chatLevel(volume * audioLevel(file)) : 0; }
        function applyLevel() { if (gain) { gain.gain.value = level(); } }
        function start() {
          if (!wanted || !buffer || !ctx) { return; }
          if (source) { try { source.stop(); } catch (e) {} source = null; }
          source = ctx.createBufferSource();
          source.buffer = buffer;
          source.loop = true;
          source.connect(gain);
          source.start();
        }
        function load() {
          if (!file || ctx) { return; }
          var AC = window.AudioContext || window.webkitAudioContext;
          if (!AC) { return; }
          ctx = new AC();
          gain = ctx.createGain();
          gain.gain.value = level();
          gain.connect(ctx.destination);
          fetch(normSrc("/gopod_reporter_audio/" + file))
            .then(function (r) { return r.ok ? r.arrayBuffer() : Promise.reject(r.status); })
            .then(function (data) { return ctx.decodeAudioData(data); })
            .then(function (decoded) { buffer = decoded; start(); })
            .catch(function () { /* no loop file */ });
        }
        document.addEventListener("pointerdown", function () {
          if (!file) { return; }
          load();
          if (ctx && ctx.state === "suspended") { ctx.resume(); }
        }, true);
        return {
          // 2026-10-20 (operator: "upon mouseover, display a vertical volume bar
          // + knob to adjust volume * range = 0% to 100% * knob position set to
          // current default"): the loop's share, set from the speaker's bar.
          volume: function () { return volume; },
          setVolume: function (v) {
            volume = Math.max(0, Math.min(1, v));
            // Review: kept with the settings' value it overrode - a new value in
            // the settings (system-set) wins again.
            applyLevel();
          },
          setup: function (cfg) {
            file = String(cfg.file || "").trim();
            volume = volumeShare(cfg.volume);
            volumeCfg = volume;
            applyLevel();
            wanted = !!file;
            if (wanted && pageClicked) { load(); }
          },
          restart: function () { start(); },
          speaker: function () { return speakerOn; },
          setSpeaker: function (on) {
            speakerOn = !!on;
            try { localStorage.setItem("gopodLoopSpeaker", speakerOn ? "on" : "off"); } catch (e) {}
            applyLevel();
          },
          runQuiet: function (quiet) { runQuiet = !!quiet; applyLevel(); },
          // A run is going (its loop_pause came, its loop_resume not yet).
          inRun: function () { return runQuiet; }
        };
      })();
      function pollRecordSession() {
        fetch("/gopod_record_session.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (state) {
            if (!state) { return; }
            recordBox.checked = !!state.on;
            recordLabel.classList.toggle("recording", !!state.recording);
            var recording = !!state.recording;
            // 2026-10-11 (operator: "while Record? is ticked * 00:00:00 timer runs
            // red * that's the difference"): the timer is the session's, from its
            // 00:00:00 (gopod_session_log.go), back to 0 when a new session starts.
            if (state.session_started_ms) {
              if (state.session_started_ms !== timerFromMs) { backgroundLoop.restart(); }
              timerFromMs = state.session_started_ms;
            }
            wasRecording = recording;
            recordStartedMs = recording ? (state.started_ms || timerFromMs) : 0;
            showSessionTimer();
          }).catch(function () { /* keep the box as it is */ });
      }
      var typedText = document.createElement("input");
      typedText.type = "text";
      typedText.className = "gopod-typed-text";
      var typedPlaceholder = "Type to the Brobot, press Enter";
      typedText.placeholder = typedPlaceholder;
      typedText.setAttribute("aria-label", "Type to the chosen Brobot");
      typedRow.insertBefore(typedText, typedFirstSep);
      // The loop's speaker toggle, at the end of the line.
      var speakerToggle = document.createElement("button");
      speakerToggle.type = "button";
      speakerToggle.className = "gopod-loop-speaker";
      function showSpeaker() {
        var on = backgroundLoop.speaker();
        speakerToggle.textContent = on ? "\uD83D\uDD0A" : "\uD83D\uDD07";
        speakerToggle.setAttribute("aria-pressed", on ? "true" : "false");
        speakerToggle.title = on ? "Background loop on - click to turn it off" : "Background loop off - click to turn it on";
        speakerToggle.setAttribute("aria-label", speakerToggle.title);
      }
      speakerToggle.addEventListener("click", function () {
        backgroundLoop.setSpeaker(!backgroundLoop.speaker());
        showSpeaker();
        // 2026-10-19 (operator: "🔇 opens it * 🔊 closes it"): the system menu
        // opens with the loop, and closes with it - for the admin only
        // (2026-10-23, operator: "for visitors, detach System Window from 🔇 *
        // 🔇 + System Window for admin users only (me)"): a visitor's 🔇 is the
        // loop alone.
        if (!isAdminPage) { return; }
        if (backgroundLoop.speaker()) { systemMenuOpen(); } else { systemMenuClose(); }
      });
      showSpeaker();
      // 2026-10-23 (operator: "| 00:56:42 |🔇"): a bar before the speaker too.
      typedRowText("|", "gopod-typed-sep");
      typedRow.appendChild(speakerToggle);
      // The speaker's volume bar: shown while the mouse is over the speaker (or the bar).
      var volumeWrap = document.createElement("div");
      volumeWrap.className = "gopod-loop-volume";
      var volumeBar = document.createElement("input");
      volumeBar.type = "range";
      volumeBar.min = "0";
      volumeBar.max = "100";
      volumeBar.step = "1";
      volumeBar.setAttribute("aria-label", "Background loop volume");
      var volumeLabel = document.createElement("span");
      volumeLabel.className = "gopod-loop-volume-label";
      volumeWrap.appendChild(volumeLabel);
      volumeWrap.appendChild(volumeBar);
      typedRow.appendChild(volumeWrap);
      function showVolume() {
        var pct = Math.round(backgroundLoop.volume() * 100);
        volumeBar.value = String(pct);
        volumeLabel.textContent = pct + "%";
      }
      // 2026-10-21 (operator: "background loop volume setting to remember over full
      // system restarts"): saved to the settings (background_loop.volume) when let go.
      var volumeSave = null;
      volumeBar.addEventListener("input", function () {
        backgroundLoop.setVolume(parseInt(volumeBar.value, 10) / 100);
        volumeLabel.textContent = volumeBar.value + "%";
        clearTimeout(volumeSave);
        volumeSave = setTimeout(function () {
          fetch("/api-sdk/gopod_loop_volume", {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: "v=" + encodeURIComponent(volumeBar.value)
          }).catch(function () { /* Wire-Pod not up */ });
        }, 600);
      });
      var volumeHide = null;
      function volumeOpen() { clearTimeout(volumeHide); showVolume(); volumeWrap.classList.add("open"); }
      function volumeClose() { clearTimeout(volumeHide); volumeHide = setTimeout(function () { volumeWrap.classList.remove("open"); }, 400); }
      [speakerToggle, volumeWrap].forEach(function (el) {
        el.addEventListener("mouseenter", volumeOpen);
        el.addEventListener("mouseleave", volumeClose);
      });
      // 2026-10-09: Esc leaves the text box (2026-10-10: the numpad's own keys
      // are push-to-talk even in the box, when Numpad? is ticked).
      typedText.addEventListener("keydown", function (ev) {
        if (ev.key === "Escape") { typedText.blur(); }
      });
      document.addEventListener("keydown", function (ev) {
        if (ev.key === "Escape" && !typingInBox()) { skipReporters("Esc"); }
      });
      typedRow.addEventListener("submit", function (ev) {
        ev.preventDefault();
        var text = typedText.value.trim();
        var chosen = typedRow.querySelector("input[name=gopod-typed-robot]:checked");
        if (!text || !chosen) { return; }
        typedText.value = "";
        // 2026-10-20 (operator: "Goverlord should know by context what a skip
        // command means * upon 'skip' input, obey same as 'Esc' key press"): a
        // line that's only "skip" (or system-skip) skips - it's never a question.
        if (/^(system-)?skip[.!]?$/i.test(text)) { skipReporters("typed"); return; }
        // system-numpad on|off: the numpad is this page's - switched here too.
        var numpadCmd = /^system-numpad\s+(on|off)\b/i.exec(text);
        if (numpadCmd) { setNumpadOn(numpadCmd[1].toLowerCase() === "on"); }
        sendTypedLine(chosen.value, text);
      });
      function sendTypedLine(serial, text) {
        var first = serial === "all" && allFirstSelect ? allFirstSelect.value : "";
        if (first === "2") { allSecondFirst[text] = true; }
        fetch("/api-sdk/gopod_typed_input", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "serial=" + encodeURIComponent(serial) + "&text=" + encodeURIComponent(text) + (first ? "&first=" + first : "") + "&visitor=" + encodeURIComponent(visitorId)
        }).then(function (r) {
          if (!r.ok) { typedText.value = text; return null; }
          return r.text();
        }).then(function (body) {
          // 2026-10-24 (gopod_visitors.go): a visitor's dropped line says why -
          // "Waiting for approval" or "Muted" - in the text field.
          var note = "";
          try { note = body && body.charAt(0) === "{" ? (JSON.parse(body).note || "") : ""; } catch (err) { note = ""; }
          typedText.placeholder = note || typedPlaceholder;
          typedText.classList.toggle("gopod-typed-note", !!note);
        }).catch(function () { typedText.value = text; });
      }
      // 2026-10-04 (operator: "the keyboard input line would be inside within the
      // robots chat bubble"): inside the chat window, kept on its bottom edge
      // (CSS: last in the window's order, sticky to its bottom) while bubbles are
      // added above it.
      container.appendChild(typedRow);
      // 2026-10-23 (operator: "New responsive Brobots-Input-Row single line can
      // adjust to 3 lines"): one line while everything fits beside a usable text
      // field, else three (.stacked, gopod_responsive.css).
      var TYPED_TEXT_MIN_PX = 200;
      function fitTypedRow() {
        var need = TYPED_TEXT_MIN_PX, gap = parseFloat(getComputedStyle(typedRow).columnGap) || 0;
        Array.prototype.forEach.call(typedRow.children, function (el) {
          if (el === typedText || getComputedStyle(el).position === "absolute" || el.offsetParent === null) { return; }
          need += el.offsetWidth + gap;
        });
        var room = typedRow.clientWidth - (parseFloat(getComputedStyle(typedRow).paddingLeft) || 0) - (parseFloat(getComputedStyle(typedRow).paddingRight) || 0);
        typedRow.classList.toggle("stacked", room > 0 && need > room);
      }
      // The row's height changes (one line to two, a visitor's shorter row): a
      // chat at its end stays at its end, its newest bubble just above the row.
      var typedRowHeight = 0;
      function typedRowResized() {
        fitTypedRow();
        var h = typedRow.offsetHeight;
        if (h === typedRowHeight) { return; }
        var atEnd = container.scrollTop >= chatMaxScroll() - Math.abs(h - typedRowHeight) - 8;
        typedRowHeight = h;
        if (atEnd && !pendingReveal.length) { scrollChatTo(chatContentEnd()); }
      }
      if (typeof ResizeObserver === "function") { new ResizeObserver(typedRowResized).observe(typedRow); }
      window.addEventListener("resize", fitTypedRow);

      // 2026-10-14 (operator: "When I refresh & click log button, I get Brobots Log
      // correctly * ensure bubbles window also scrolls to bottom position"): the
      // chat is restored while the Log section is closed (no height, so no scroll);
      // the moment the window gets its height, it goes straight to its end.
      if (typeof ResizeObserver === "function") {
        var chatHadHeight = container.clientHeight > 0;
        new ResizeObserver(function () {
          var has = container.clientHeight > 0;
          if (has && !chatHadHeight) {
            if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
            container.scrollTop = chatMaxScroll();
          }
          chatHadHeight = has;
        }).observe(container);
      }

      function showChat(open) {
        container.classList.toggle("expanded", open);
        textarea.style.display = open ? "none" : "";
      }

      // Brobots Chat button: the only show/hide. Never touches the chat's content.
      window.showBrobotChatBubbles = function () {
        var open = !container.classList.contains("expanded");
        setChatOpen(open);
        showChat(open);
        // 2026-10-08 (operator: "If I click the brobots chat button (off/on)
        // scrolling doesn't reach full bottom position"): shown again, the chat
        // goes straight to its end, once the window has its size back.
        if (open) {
          requestAnimationFrame(function () {
            if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
            container.scrollTop = chatMaxScroll();
          });
        }
      };

      // GOLDEN chat-only view (2026-10-23, FINISH_PLAN_001.md C2; operator: "C2
      // please, but from the wire-pod page 'Log' button tap"): on a visitor's
      // device (a phone - "phone = visitor"), tapping Log opens the chat alone -
      // the Brobots chat and the typing row, the whole screen, no Wire-Pod
      // settings; "‹ Wire-Pod" goes back. The admin's Log stays as it was.
      // "?visitor" in the address shows the visitor's view on GOPOD's machine.
      var chatBack = document.createElement("button");
      chatBack.type = "button";
      chatBack.className = "gopod-chat-back";
      chatBack.textContent = "\u2039 Wire-Pod";
      chatBack.addEventListener("click", function () { document.body.classList.remove("gopod-chat-only"); });
      // 2026-10-24 (live: the PC got the visitor's view - Log was tapped before
      // Wire-Pod had said who the page is): it waits for that answer.
      var chatOnlyPending = false;
      function chatOnlyOn() {
        if (!pageRoleKnown) { chatOnlyPending = true; return; }
        if (isAdminPage) { document.body.classList.remove("gopod-chat-only"); return; }
        var log = document.getElementById("section-log");
        if (log && chatBack.parentNode !== log) { log.insertBefore(chatBack, log.firstChild); }
        document.body.classList.add("gopod-chat-only");
        if (!container.classList.contains("expanded")) { window.showBrobotChatBubbles(); }
      }
      window.addEventListener("load", function () {
        var wirePodShowLog = window.showLog;
        if (typeof wirePodShowLog !== "function") { return; }
        window.showLog = function () {
          var r = wirePodShowLog.apply(this, arguments);
          chatOnlyOn();
          return r;
        };
      });

      // Clear Chat button: empties the bubbles now, and the saved copy. Lines
      // already shown stay marked as seen, so they don't come back on the next poll.
      // The only thing that empties the chat (2026-09-30).
      window.clearBrobotChat = function () {
        clearBubbles();
        chatStore.bubbles = [];
        saveChat();
      };

      function clearBubbles() {
        Array.prototype.slice.call(container.querySelectorAll(".brobot-bubble")).forEach(function (b) {
          b.remove();
        });
        lastOperatorBubble = null;
        typedRow.style.marginTop = "";
      }

      // 2026-09-28: rich display in the robot bubbles - the Brobot's reply as written:
      // **bold**, *italic*, and math (\[...\], $$...$$, \(...\)) rendered by KaTeX.
      // Text is HTML-escaped first; math is cut out before the markdown pass and put
      // back as KaTeX output. If KaTeX is missing, the math shows as written.
      function escapeHtml(s) {
        return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
      }
      function renderMath(tex, display) {
        if (!window.katex) { return escapeHtml(display ? "\\[" + tex + "\\]" : "\\(" + tex + "\\)"); }
        try {
          return window.katex.renderToString(tex, { displayMode: display, throwOnError: false });
        } catch (e) {
          return escapeHtml(tex);
        }
      }
      // Only for an old, unmarked bubble from saved history - a live reply or a
      // song line comes marked by Wire-Pod (gopod_highlight_pieces.go) and is
      // never split here. Same plain rule as gopod_speech_sentences.go
      // (operator 2026-10-06: sentences first, then every comma inside each;
      // no length rules). Emojis right after an ending stay with it.
      var TRAILING_EMOJI_RE = /^(\s*(?:\p{Extended_Pictographic}|\uFE0F|\u200D)+)+(?=\s|$)/u;
      // An emoji used as a sentence ending ("in each sentence 😉 You know") is one
      // too: Wire-Pod says a period there (gopodEmojiSentenceEnd), so the page splits
      // there as well - after a word, emojis, a space, then a capital.
      var EMOJI_END_RE = /^[ \t]*(?:(?:\p{Extended_Pictographic}|\uFE0F|\u200D)[ \t]*)+[ \t]+\p{Lu}/u;
      function hardEnd(text, i) {
        var c = text[i];
        if (/[\p{L}\p{N})"']/u.test(c) && /[ \t\uD800-\uDBFF\u2600-\u27BF]/.test(text[i + 1] || "") && EMOJI_END_RE.test(text.slice(i + 1))) { return true; }
        if (liveSentenceEnds.indexOf(c) < 0) { return false; }
        return !(c === "." && text[i - 1] === ".");
      }
      function softEnd(text, i) {
        var c = text[i];
        if (liveCommaEnds.indexOf(c) >= 0) { return true; }
        if (dashPause(text, i)) { return true; }
        return c === "\u2026" || (c === "." && text[i - 1] === ".");
      }
      // A dash used as a pause between words is a comma in speech
      // (gopod_spoken_dashes.go, and "—" in kgsim.go), so it can end a piece
      // here too (2026-10-05, live 18:59: "disco balls - a big ball" was said in
      // two pieces but lit as one). An em dash always; an en dash or hyphen when
      // spaced ("a - b"), hanging off a word ("are- but") or doubled
      // ("laugh--that"). A hyphen inside a word, or between numbers, is not.
      function dashPause(text, i) {
        var c = text[i];
        if (c === "\u2014") { return true; }
        if (c !== "-" && c !== "\u2013") { return false; }
        var before = text[i - 1] || "", after = text[i + 1] || "";
        var left = text.slice(0, i).replace(/[\s\-\u2013]+$/, "").slice(-1), right = text.slice(i + 1).replace(/^[\s\-\u2013]+/, "").charAt(0);
        if (/[0-9]/.test(left) && /[0-9]/.test(right)) { return false; }
        if (after === "-" || after === "\u2013") { return false; }
        if (/\s/.test(before) && /\s/.test(after)) { return true; }
        if (/\p{L}/u.test(before) && /\s/.test(after)) { return true; }
        return (before === "-" || before === "\u2013") && /\p{L}/u.test(text[i - 2] || "") && /\p{L}/u.test(after);
      }
      function cutAt(text, isEnd) {
        var out = [], cur = "";
        for (var i = 0; i < text.length; i++) {
          cur += text[i];
          if (isEnd(text, i)) {
            var emoji = text.slice(i + 1).match(TRAILING_EMOJI_RE);
            if (emoji) { cur += emoji[0]; i += emoji[0].length; }
            // An ellipsis between two words ("for the...thinks") ends a piece too:
            // Wire-Pod says it as ", " (gopodEllipsisToComma), live 2026-10-04 20:36.
            var ellipsis = text[i] === "\u2026" || (text[i] === "." && text[i - 1] === ".");
            if (text[i + 1] === " " || (ellipsis && isEnd === softEnd && /[A-Za-z0-9]/.test(text[i + 1] || "")) || (isEnd === softEnd && dashPause(text, i))) { out.push(cur); cur = ""; }
          }
        }
        if (cur.trim()) { out.push(cur); }
        // The pieces keep their own spacing, so joining them gives back the text.
        return out;
      }
      function splitSentences(line) {
        var out = [];
        cutAt(line.trim(), hardEnd).forEach(function (sentence) {
          cutAt(sentence, softEnd).forEach(function (piece) { if (piece.trim()) { out.push(piece.trim()); } });
        });
        return out;
      }
      // 2026-10-15 (the Weather Kit link in Brobot 2's weather follow-up): a
      // "[words](/page.html)" link shows as its words, a link that opens in a new
      // tab (Wire-Pod's own pages or a full web address); the robot says the words.
      function bubbleLinks(html) {
        return String(html).replace(/\[([^\]\n]+)\]\((\/[^)\s]*|https?:\/\/[^)\s]+)\)/g, function (m, words, href) {
          return "<a class=\"bubble-link\" href=\"" + href.replace(/"/g, "&quot;") + "\" target=\"_blank\" rel=\"noopener\">" + words + "</a>";
        });
      }
      // 2026-10-17 (operator: "can you replace flag-text with flag emoji"): some
      // browsers (Windows) draw a flag emoji as its two letters ("US"); a known
      // country's flag (clean_politics.countries) is drawn as its picture
      // (assets/flags, Twemoji), the emoji kept as its alt text.
      var flagCodes = {};
      function flagPictures(html) {
        return String(html).replace(/(?:\uD83C[\uDDE6-\uDDFF]){2}/g, function (flag) {
          var a = flag.codePointAt(0), b = flag.codePointAt(2);
          var code = String.fromCharCode(65 + a - 0x1F1E6, 65 + b - 0x1F1E6);
          if (!flagCodes[code]) { return flag; }
          return "<img class=\"flag-emoji\" src=\"/assets/flags/" + a.toString(16) + "-" + b.toString(16) + ".svg\" alt=\"" + flag + "\">";
        });
      }
      function renderRich(text) { return flagPictures(renderRichText(text)); }
      function renderRichText(text) {
        var math = [];
        var keep = function (tex, display) { math.push(renderMath(tex.trim(), display)); return "\u0000" + (math.length - 1) + "\u0000"; };
        var t = String(text)
          .replace(/\\\[([\s\S]*?)\\\]/g, function (m, tex) { return keep(tex, true); })
          .replace(/\$\$([\s\S]*?)\$\$/g, function (m, tex) { return keep(tex, true); })
          .replace(/\\\(([\s\S]*?)\\\)/g, function (m, tex) { return keep(tex, false); });
        // 2026-09-29: each sentence is its own span, so the one the robot is saying
        // can light up (pollSpeakingNow below).
        // GOLDEN 2026-10-06 (gopod_highlight_pieces.go): a marked bubble - every
        // live reply and every scripted song line - is shown exactly as Wire-Pod
        // cut it, piece by piece, line breaks kept inside its piece, and never
        // split again here. Its pieces are numbered in this order, and Wire-Pod
        // names the number it is saying. splitSentences below is only for an
        // old, unmarked bubble.
        if (segmentMarker && t.indexOf(segmentMarker) >= 0) {
          // 2026-10-16 (operator: "please add a line between thought bubbles
          // ... thought bubbles as (spaced) paragraphs"): a piece after a line
          // break (live_split.thought_lines: a new thought) starts a new
          // paragraph, a blank line below the last.
          var newThought = false;
          var marked = escapeHtml(t).split(segmentMarker).map(function (p) {
            var item = { text: p.trim(), gap: newThought };
            if (item.text) { newThought = /\n\s*$/.test(p); }
            return item;
          }).filter(function (item) { return item.text; }).map(function (item, i) {
            var piece = item.text;
            return (item.gap && i > 0 ? "<span class=\"thought-gap\"></span>" : "") + "<span class=\"brobot-say\">" + bubbleLinks(piece)
              .replace(/\*\*([^*\n]+)\*\*/g, "<strong>$1</strong>")
              .replace(/(^|[^*])\*([^*\n]+)\*/g, "$1<em>$2</em>")
              .replace(/\n+/g, "<br>") + "</span>";
          }).join(" ");
          return ("<div class=\"brobot-bubble-line\">" + marked + "</div>").replace(/\u0000(\d+)\u0000/g, function (m, n) { return math[+n]; });
        }
        t = escapeHtml(t).split("\n").map(function (line) {
          if (!line.trim()) { return ""; }
          var pieces = segmentMarker && line.indexOf(segmentMarker) >= 0
            ? line.split(segmentMarker).map(function (p) { return p.trim(); }).filter(Boolean)
            : splitSentences(line);
          var said = pieces.map(function (sentence) {
            return "<span class=\"brobot-say\">" + bubbleLinks(sentence)
              .replace(/\*\*([^*\n]+)\*\*/g, "<strong>$1</strong>")
              .replace(/(^|[^*])\*([^*\n]+)\*/g, "$1<em>$2</em>") + "</span>";
          }).join(" ");
          return "<div class=\"brobot-bubble-line\">" + said + "</div>";
        }).join("");
        return t.replace(/\u0000(\d+)\u0000/g, function (m, n) { return math[+n]; });
      }

      // 2026-09-30: GOPOD's own names as speech-to-text hears them ("go pod",
      // "govern lord") written the way the repo spells them (GOPOD, Goverlord) -
      // display only. Same list and rules as gopod_name_spelling.go: longest first,
      // case-insensitive, whole words (gopod_name_spellings.json).
      var nameSpellings = [];
      function spellNames(text) {
        var out = String(text);
        nameSpellings.forEach(function (e) { out = out.replace(e.pattern, e.to); });
        return out;
      }
      function loadNameSpellings() {
        return fetch("/gopod_name_spellings.json", { cache: "no-store" })
          .then(function (r) { return r.json(); })
          .then(function (data) {
            nameSpellings = (data.name_spellings || [])
              .filter(function (e) { return e.from && e.to; })
              .sort(function (a, b) { return b.from.length - a.from.length; })
              .map(function (e) {
                var escaped = e.from.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
                return { to: e.to, pattern: new RegExp("(?<![\\p{L}\\p{N}_])" + escaped + "(?![\\p{L}\\p{N}_])", "giu") };
              });
          })
          .catch(function () { /* no list - names show as heard */ });
      }

      function appendBubble(who, spoken, format, to, via, then) {
        var next = function () { if (then) { then(); } };
        voiceLitNext = null; // the next voice is this bubble's (the Operator's), not an earlier line's
        if (!restoringChat) {
          spoken = spellNames(spoken);
          if (who === "Operator" && !to) { to = operatorTarget(spoken); }
          recordBubble(who, spoken, format, to, via);
        }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble " + (who === "Operator" ? "operator" : who === "Brobot 1" ? "brobot-1" : "brobot-2");
        // 2026-09-29: Wire-Pod chat replies get the speaking highlight (gopod_chat_bubbles.css).
        if (format === "llm") { bubble.className += " brobot-bubble-live"; }
        // 2026-10-01: a scripted song line (split with the segment marker, the
        // Bingo test) lights per sentence the same way (operator: "I did not see
        // sentence highlighting in bingo bubbles").
        else if (segmentMarker && String(spoken).indexOf(segmentMarker) >= 0) { bubble.className += " brobot-bubble-scripted"; }
        if (who === "Operator" && !restoringChat && /^system-/i.test(String(spoken).trim()) && systemMenuResult(String(spoken).trim(), "ask")) { return; }
        if (who === "Operator") {
          // 2026-09-28: the operator's own words - dark grey, centered, no robot face.
          var opText = document.createElement("span");
          opText.className = "brobot-bubble-text";
          opText.dataset.raw = who + ": " + spoken;
          opText.textContent = spoken;
          bubble.appendChild(operatorHead(to || operatorTarget(spoken), via, spoken));
          bubble.appendChild(opText);
          bubble.dataset.own = "1"; // the operator's own line
          operatorSeq++;
          // A new question: an unanswered pop-up keeps only its prize. 2026-10-14
          // (live 20:24, "what's in front of you" by voice: the prize and the reply
          // came as two bubbles): a backpack line is logged once the answer is
          // done (Wire-Pod's "Intent matched"), after its own pop-up - it never
          // breaks the pop-up off its reply.
          // It goes above the pop-up it asked for, so the question reads first.
          var askedWrap = via === "Backpack" ? popupWraps[to] : null;
          if (via !== "Backpack") { popupWraps = {}; popupExchange = false; nestShells = {}; }
          // GOLDEN operator bubble (2026-10-18, operator: "Operator bubble: (not
          // manager bubble) * KP7 to CHALK / keep holding… / Recording for CHALK /
          // Got it - transcribing… / ["Generating..." + hourglass]/[transcription]"):
          // a numpad line takes over its key's status bubble - its header becomes
          // the Operator's, the status lines stay, the words go under them.
          var st = numpadOpStatus;
          numpadOpStatus = null;
          if (st && !restoringChat && st.bubble && st.bubble.isConnected && /^KP/.test(String(via || ""))) {
            st.bubble.className = bubble.className;
            st.bubble.replaceChild(bubble.firstChild, st.bubble.firstChild);
            st.textEl.className = "numpad-lines";
            st.textEl.style.display = "block";
            st.bubble.appendChild(opText);
            st.bubble.dataset.own = "1";
            if (format !== "plain") { lastOperatorBubble = st.bubble; operatorInBatch = true; }
            return;
          }
          // (popupWraps: each Brobot's first egg for this question - a late
          // backpack line goes above it.)
          // 2026-10-22 (operator: "Chronologically, this bubble should appear
          // before the response, correct?"): a backpack line goes above the
          // answer it asked for too - that Brobot's open bubble, made after it
          // was asked, with no operator line above it yet.
          var askedShell = via === "Backpack" ? nestShells[to] : null;
          if (!askedWrap && askedShell && askedShell.isConnected && !askedShell.dataset.asked) {
            askedWrap = askedShell;
            askedShell.dataset.asked = "1";
            shellOperatorSeq[to] = operatorSeq;
          }
          if (askedWrap && askedWrap.isConnected) { container.insertBefore(bubble, askedWrap); }
          else { container.appendChild(bubble); }
          // 2026-10-16 (comb-through: the marquee stuck after a voice question
          // that found an egg): a backpack line comes after its own reply - that
          // exchange is over now.
          if (via === "Backpack") { popupWraps = {}; popupExchange = false; }
          if (restoringChat) { return; }
          // 2026-10-04: an operator line that arrives during the warm-up (the
          // second robot's "Bingo Test Live") goes above the warm-up bubble.
          if (warmupBubble && warmupBubble.isConnected) { container.appendChild(warmupBubble); }
          // 2026-09-29: the operator's question (or a song's opening line, e.g.
          // "Bingo Test Live") goes to the top of the chat window; the reply below
          // it follows the spoken sentence (focusSpeaking). 2026-09-30: always,
          // even when the reply is longer than the window - except "sleep",
          // which (with its "zzz") is placed like any other bubble.
          if (format === "plain") {
            lastOperatorBubble = null;
          } else {
            lastOperatorBubble = bubble;
            operatorInBatch = true;
          }
          revealWhenPlaced(bubble);
          next();
          return;
        }
        var textEl = document.createElement("span");
        textEl.className = "brobot-bubble-text";
        textEl.dataset.raw = who + ": " + spoken;
        textEl.innerHTML = renderRich(spoken);
        // GOLDEN pop-up rule (2026-10-15, operator: '"All" = ALL brobots in
        // sequence ... applies everywhere, including pop ups'; "brobot 2 should
        // ... also have marquee lights, to indicate visually, he's part of the
        // same pop up"): every Brobot answering a pop-up's question gets his own
        // marquee bubble, his "Brobot N:" header kept; the eggs are bubbles of
        // their own above (2026-10-15) - the pop-ups and the All sequence stay
        // separate processes; this only dresses the reply.
        var inPopup = popupExchange && /^Brobot [12]$/.test(who);
        // His reply fills the bubble a manager's part opened for him (nestShell).
        var shell = nestShells[who];
        if (shell && shell.isConnected && !shell.dataset.filled && !restoringChat) {
          shell.dataset.filled = "1";
          if (feedIdNow) { shell.dataset.feed = String(feedIdNow); }
          bubble.className.split(" ").forEach(function (c) { if (c) { shell.classList.add(c); } });
          if (inPopup) { shell.classList.add("popup-winner"); }
          shell.appendChild(textEl);
          attachPendingBall(shell, spoken);
          if (inPopup) { marqueeAttach(shell); }
          addQrCodes(shell);
          // 2026-10-22: his weather line's picture in this bubble too.
          addWeatherFace(shell);
          next();
          return;
        }
        if (inPopup) { bubble.classList.add("popup-winner"); }
        if (feedIdNow && /^Brobot [12]$/.test(who)) { bubble.dataset.feed = String(feedIdNow); }
        bubble.appendChild(bubbleHead(who));
        bubble.appendChild(textEl);
        attachPendingBall(bubble, spoken);
        // A Brobot's own bubble, on the raw bubble template (built and loaded
        // hidden, then shown); the marquee lights once it's in the chat.
        placeBubble(bubble, next, function (b) { if (inPopup) { marqueeAttach(b); } });
      }

      // 2026-09-30, one rule for every new bubble - a backpack question and its
      // reply, a Bingo run's lines alike (operator: "a short intentional delay
      // before displaying the operator chat bubble with the response, positioned
      // at the top ... apply the same positioning to the bingo chat bubbles").
      // A new bubble is added hidden (it still takes its room), the chat moves to
      // where it should be, and the new bubbles appear together once the move is
      // done - never before CHAT_REPLY_DELAY_MS. Bubbles that arrive together (a
      // question and its reply are logged together) are placed and shown together.
      // Where (operator: "every operator bubble aims to go to the top, or as
      // high as it can go depending on the length of the following brobot
      // response"): a new operator line goes to the top, or as high as the
      // chat below it lets it scroll - no empty room is added. Lines that come
      // later under it (a Bingo run) keep the newest line's end in view. With
      // no operator line, the bottom. Placing never waits for a robot to stop
      // speaking (2026-09-30: song lines now light per sentence, and waiting
      // left new Bingo bubbles below the window). While a robot speaks, the
      // spoken-sentence centring also moves the chat down.
      var pendingReveal = [];
      var revealTimer = null;
      var placeAgain = false;
      var operatorInBatch = false; // the bubbles being placed include an operator line
      // 2026-10-04 (operator: "Operator bubbles not jumping to top alignment
      // correctly"): a typed line shows at once and its reply comes ~10 s later,
      // as its own batch - which was placed as "later lines", scrolling to the
      // reply's end and pushing the question off the top. The first robot bubble
      // after an operator line is its reply: placed like it came with it.
      var awaitingReply = false;
      // 2026-10-13 (operator: "when the next bubble is being rendered and audio
      // loaded, do not open scroll space for it * render/load the next bubble *
      // WHEN ITS READY then adjust scroll & display * do not go to the wedding,
      // then dress for it"): a new bubble is built out of the chat's flow
      // (hidden, taking no room), and joins it only when it's ready to show.
      function stageBubble(b) {
        if (getComputedStyle(container).position === "static") { container.style.position = "relative"; }
        var pad = getComputedStyle(container).paddingLeft;
        b.dataset.staged = "1";
        b.style.position = "absolute";
        b.style.top = "0";
        b.style.left = pad;
        b.style.right = pad;
        b.style.pointerEvents = "none";
      }
      function unstageBubble(b) {
        if (b.dataset.staged !== "1") { return; }
        delete b.dataset.staged;
        b.style.position = b.style.top = b.style.left = b.style.right = b.style.pointerEvents = "";
      }
      function revealWhenPlaced(bubble) {
        bubble.style.visibility = "hidden";
        stageBubble(bubble);
        // 2026-10-12: the voice that comes right after this bubble's line.
        if (nextBubbleVoice) { bubble._voice = nextBubbleVoice; nextBubbleVoice._bubble = bubble; nextBubbleVoice = null; }
        bubble._feedId = feedIdNow;
        pendingReveal.push(bubble);
        // Drawn hidden: its voiced lines go into the template before it shows
        // (after this event has filled it in, before anything is painted).
        Promise.resolve().then(voiceTieAhead);
        if (revealTimer === null) {
          revealTimer = setTimeout(placeNewBubbles, 0);
        } else {
          placeAgain = true; // arrived during a move: place again before showing
        }
      }
      // 2026-10-16 (comb-through: the chat stopped following the highlight while
      // a next bubble waited - that bubble is staged at the top, out of the
      // chat's flow): the newest bubble that's placed, never a staged one.
      function lastPlacedChild() {
        var el = container.lastElementChild;
        while (el && (el.dataset.staged === "1" || el === typedRow)) { el = el.previousElementSibling; }
        return el;
      }
      function chatContentEnd() {
        var newest = lastPlacedChild();
        // The typed-input line sits on the window's bottom edge: the newest
        // bubble ends above it, not behind it.
        var typedHeight = newest && newest !== typedRow ? typedRow.offsetHeight : 0;
        return newest ? contentY(newest) + newest.offsetHeight + CHAT_TOP_GAP_PX + typedHeight - container.clientHeight : 0;
      }
      // 2026-10-04 (operator: "Operator bubble not correctly being positioned at
      // the top of the window"): with a short reply (or none yet) there wasn't
      // enough chat below the operator line for it to scroll to the top. Room is
      // added above the typed-input line, just enough for the line to reach the
      // top, and taken away again as the reply fills it.
      function roomBelowOperator(questionTop) {
        var current = parseFloat(typedRow.style.marginTop) || 0;
        var need = questionTop === null ? 0 : Math.max(0, Math.ceil(questionTop + container.clientHeight - (container.scrollHeight - current)));
        if (need !== current) { typedRow.style.marginTop = need ? need + "px" : ""; }
      }
      function placeTarget() {
        // 2026-10-11 (operator: "Then the next bubble appears at the bottom pushing
        // the others up * scrolling position stays at bottom"): every bubble, the
        // operator's too, goes at the bottom; no room is added above the typing line.
        roomBelowOperator(null);
        var first = pendingReveal.length ? pendingReveal[0] : null;
        if (!first || !container.contains(first)) { return chatContentEnd(); }
        return bubbleTarget(first);
      }
      // GOLDEN rule 14, every chat bubble (operator 2026-10-10: "create a golden
      // rule about creating the full bubble hidden, then handover the baton in
      // olympic fashion"; "elevate the narrator bubble rule globally to all chat
      // bubbles"): a new bubble waits hidden in its lane, no scroll, while a
      // bubble on screen is still being said or sung (its highlight lit); then
      // the chat scrolls once and it shows. The operator's own line waits too (2026-10-11).
      // At most CHAT_BATON_MAX_MS, so a highlight left on never holds the chat.
      // 2026-10-13 (operator: "15-second cap * please fix/resolve"; rule: "every
      // chat bubble must run & finish 100% before the next bubble"): 15 s let a
      // long answer be overtaken while still being said; now only a highlight
      // stuck on for 2 minutes lets go.
      var CHAT_BATON_MAX_MS = 120000, SPEAKING_MAX_MS = 30000;
      function batonHeld() {
        if (viewerBehind()) { return false; }
        if (narratorLines.some(function (n) { return n.shown; })) { return true; }
        return Array.prototype.some.call(container.querySelectorAll(".brobot-say.speaking"), function (el) {
          var b = el.closest(".brobot-bubble");
          // 2026-10-18 (leak backstop): a line lit longer than
          // chat_window.speaking_max_ms (a robot that froze, a lost "done")
          // no longer holds the next bubble.
          // 2026-10-23: a line lit some other way gets its lit time when first seen,
          // so it times out like every other (never holds the chat for good).
          if (!el._litAt) { el._litAt = Date.now(); }
          if (Date.now() - el._litAt > SPEAKING_MAX_MS) { return false; }
          return b && !b.classList.contains("narrator") && b.style.visibility !== "hidden";
        });
      }
      // 2026-10-12 (operator: "Ensure every chat bubble is fully visually
      // rendered, audio loaded (or waiting to be loaded if is being generated) &
      // ready to play. Only then is the chat bubble measured for scrolling, to
      // insert from the bottom up ... whichever top or bottom of that chat bubble
      // reaches position first holds there ... shorter bubbles stop at bottom
      // since top-border did not reach * taller bubbles stop at top since
      // bottom-border did not reach * llm warm ups are the exception"): a new
      // bubble (or the bubbles arriving with it) that fits the window ends at its
      // bottom; one taller than the window starts at its top. The LLM warm-up
      // bubble never goes through here.
      function bubbleTarget(first) {
        var last = lastPlacedChild();
        if (!last) { last = first; }
        var typedHeight = last !== typedRow ? typedRow.offsetHeight : 0;
        var room = container.clientHeight - typedHeight - 2 * CHAT_TOP_GAP_PX;
        // 2026-10-14 (operator: "the bubble needs to be repositioned down to where
        // the, [icon] Brobot N: , header is ... aim for brobot header at top of
        // window, if bottom of bubble hits bottom of window first, that's the stop
        // scroll position, no extra space below the bubble"): a pop-up the reply
        // has joined aims its reply header ("Brobot N:" under the prize) at the top.
        // The reply's "Brobot N:" header under the prize.
        var replyHeads = first.classList.contains("popup-wrap") ? first.querySelectorAll(":scope > .popup-system ~ .brobot-bubble-head") : [];
        var replyHead = replyHeads.length ? replyHeads[replyHeads.length - 1] : null;
        var top = contentY(replyHead || first);
        if (contentY(last) + last.offsetHeight - top <= room) { return chatContentEnd(); }
        return top - CHAT_TOP_GAP_PX;
      }
      // Calls then once every picture in the bubbles has loaded and each one's
      // voice can play (at most BUBBLE_READY_MAX_MS).
      var BUBBLE_READY_MAX_MS = 6000;
      function whenBubblesReady(bubbles, then) {
        var from = performance.now();
        (function check() {
          var waiting = bubbles.some(function (b) {
            var imgs = Array.prototype.some.call(b.querySelectorAll("img"), function (img) { return !img.complete; });
            var voice = b._voice && !b._voice._failed && b._voice.readyState < 3;
            return imgs || voice;
          });
          if (waiting && performance.now() - from < BUBBLE_READY_MAX_MS) {
            // 2026-10-23 (live: every bubble after the granny photo waited until
            // F5): its own timer - revealTimer means "bubbles being placed", and
            // a picture's wait (showPictures) must never leave it set.
            setTimeout(check, 50);
            return;
          }
          then();
        })();
      }
      // 2026-10-23 (live: the interview's end held ~57 s): a bubble waiting on the
      // baton longer than 10 s logs what holds it (a lit line, its text).
      var batonNoted = 0;
      function batonHolder() {
        var held = Array.prototype.filter.call(container.querySelectorAll(".brobot-say.speaking"), function (el) {
          var b = el.closest(".brobot-bubble");
          return b && !b.classList.contains("narrator") && b.style.visibility !== "hidden" && !(el._litAt && Date.now() - el._litAt > SPEAKING_MAX_MS);
        })[0];
        if (held) { return "lit line: " + String(held.textContent).slice(0, 60) + (held._litAt ? "" : " (no lit time)"); }
        if (narratorLines.some(function (n) { return n.shown; })) { return "narrator line"; }
        return "";
      }
      function placeNewBubbles() {
        var waitFrom = performance.now();
        (function waitForBaton() {
          if (performance.now() - waitFrom > 10000 && Date.now() - batonNoted > 10000 && batonHeld()) {
            batonNoted = Date.now();
            fetch("/api-sdk/gopod_page_note", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" },
              body: "note=" + encodeURIComponent("GOPOD_CHECK baton held 10 s+ by " + batonHolder()) }).catch(function () {});
          }
          // 2026-10-11 (operator: "upon new operator bubble appearing, scrolling
          // gets messed up again"): the operator's own line waits its turn too,
          // so it never scrolls away a reply still being said.
          if (batonHeld() && performance.now() - waitFrom < CHAT_BATON_MAX_MS) {
            revealTimer = setTimeout(waitForBaton, 100);
            return;
          }
          whenBubblesReady(pendingReveal.slice(), placeNewBubblesNow);
        })();
      }
      // 2026-10-13 (operator: "\"System llm bubble\" displays first and always
      // stays at the bottom of what ever happens until LLM Ready"): a warming
      // bubble moves back under every new bubble.
      function warmupToBottom() {
        if (warmupBubble && warmupBubble.isConnected && container.lastElementChild !== warmupBubble) {
          container.appendChild(warmupBubble);
        }
      }
      // 2026-10-13 (operator: "upon last message update, the bubble resized
      // correctly, however scrolling didn't account for the resize and the bubble
      // was visually cut off"): the newest bubble on screen that grows (a new
      // message, a picture loading, a highlight) is placed again, whole - only
      // while it's in view, so reading back up isn't pulled away.
      var bubbleSizes = typeof ResizeObserver === "function" ? new ResizeObserver(function (entries) {
        entries.forEach(function (entry) {
          var b = entry.target;
          var h = Math.round(entry.contentRect.height);
          var was = b._seenHeight || 0;
          b._seenHeight = h;
          if (!was || h <= was || !b.isConnected || b.style.visibility === "hidden" || pendingReveal.length) { return; }
          if (b !== container.lastElementChild && b !== warmupBubble) { return; }
          var top = contentY(b) - container.scrollTop;
          if (top > chatViewHeight() || top + h < 0) { return; }
          scrollChatTo(bubbleTarget(b));
        });
      }) : null;
      function watchSize(b) { if (bubbleSizes && b) { bubbleSizes.observe(b); } }
      function placeNewBubblesNow() {
        pendingReveal.forEach(unstageBubble);
        warmupToBottom();
        scrollChatTo(placeTarget());
        var started = performance.now();
        (function waitForPlace() {
          if (placeAgain) {
            placeAgain = false;
            whenBubblesReady(pendingReveal.slice(), function () {
              pendingReveal.forEach(unstageBubble);
              warmupToBottom();
              scrollChatTo(placeTarget());
              waitForPlace();
            });
            return;
          }
          if (chatMove || performance.now() - started < (viewerBehind() ? 0 : CHAT_REPLY_DELAY_MS)) {
            revealTimer = setTimeout(waitForPlace, 40);
            return;
          }
          revealTimer = null;
          if (operatorInBatch) { awaitingReply = true; }
          operatorInBatch = false;
          var shown = pendingReveal;
          if (shown.some(function (b) { return !b.classList.contains("operator"); })) {
            // The reply is placed; lines after it are placed as before.
            setTimeout(function () { awaitingReply = false; }, 0);
          }
          pendingReveal = [];
          shown.forEach(function (b) { b.style.visibility = ""; watchSize(b); });
          // The baton has passed: every bubble above these has been heard.
          if (shown.length) {
            var firstShown = shown[0];
            Array.prototype.forEach.call(container.querySelectorAll(".brobot-bubble"), function (b) {
              if (b !== firstShown && (b.compareDocumentPosition(firstShown) & Node.DOCUMENT_POSITION_FOLLOWING)) { revealRest(b); }
            });
          }
        })();
      }

      // 2026-09-29: centre-focused auto-scroll (operator's ask). The operator's
      // question goes to the top; once the spoken sentence's line gets past the
      // middle of the chat window, the window scrolls so that line stays in the
      // middle, down to the end of the reply (the browser stops at the bottom).
      // It never scrolls back up on its own.
      var CHAT_REPLY_DELAY_MS = 400;
      var CHAT_TOP_GAP_PX = 12;
      var lastOperatorBubble = null;
      // 2026-09-30, v10 (operator: "a NEW revised scroller of our own with a 1s ease").
      // One controller is the only thing that moves the chat. Each move eases in and
      // out over 1 second. A new target during a move does not stop and restart it:
      // the move carries on from where it is, at the speed it has, and eases into the
      // new target over the next second (a cubic Hermite curve: start position, start
      // speed, end position, end speed 0). The controller keeps its own exact position
      // so rounding never builds up. Once the end of the chat reaches the bottom,
      // centring stops and the highlight carries on down (no empty space). Instant when
      // the viewer asks for reduced motion.
      var CHAT_SCROLL_MS = 1000;
      var chatMove = null; // { from, to, speed (px per ms at the start), start (ms), frame }
      function chatMoveAt(now) {
        var T = CHAT_SCROLL_MS;
        var t = Math.max(0, Math.min((now - chatMove.start) / T, 1));
        var t2 = t * t, t3 = t2 * t;
        var pos = (2 * t3 - 3 * t2 + 1) * chatMove.from + (t3 - 2 * t2 + t) * T * chatMove.speed + (3 * t2 - 2 * t3) * chatMove.to;
        var speed = ((6 * t2 - 6 * t) * chatMove.from + (6 * t - 6 * t2) * chatMove.to) / T + (3 * t2 - 4 * t + 1) * chatMove.speed;
        return { t: t, pos: pos, speed: speed };
      }
      function chatMaxScroll() { return Math.max(0, container.scrollHeight - container.clientHeight); }
      // 2026-10-24 (operator: "ensure bubbles scroll correct & proper to adjusted
      // brobots row-height"): the chat's visible height - above the typing row
      // (one, two or three lines, whatever it is now).
      function chatViewHeight() {
        return Math.max(1, container.clientHeight - (typedRow.isConnected && typedRow.offsetParent !== null ? typedRow.offsetHeight : 0));
      }
      function contentY(el) {
        var box = container.getBoundingClientRect();
        var rect = el.getBoundingClientRect();
        return rect.top - box.top + container.scrollTop;
      }
      function scrollChatTo(top) {
        var target = Math.max(0, Math.min(top, chatMaxScroll()));
        var reduce = false;
        try { reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches; } catch (e) {}
        if (reduce || viewerBehind()) {
          if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
          container.scrollTop = target;
          return;
        }
        var now = performance.now();
        var from = container.scrollTop, speed = 0;
        if (chatMove) {
          var at = chatMoveAt(now);
          from = at.pos;
          speed = at.speed;
        } else if (Math.abs(target - from) < 1) {
          return;
        }
        var frame = chatMove ? chatMove.frame : null;
        chatMove = { from: from, to: target, speed: speed, start: now, frame: frame };
        if (frame === null) { chatMove.frame = requestAnimationFrame(chatMoveStep); }
      }
      function chatMoveStep(now) {
        var at = chatMoveAt(now);
        if (at.t >= 1) {
          container.scrollTop = Math.min(chatMove.to, chatMaxScroll());
          chatMove = null;
          return;
        }
        container.scrollTop = Math.max(0, Math.min(at.pos, chatMaxScroll()));
        chatMove.frame = requestAnimationFrame(chatMoveStep);
      }
      // The spoken sentence: once it is below the middle of the chat window, the
      // window scrolls so the sentence's middle lands on the window's middle.
      // 2026-10-15 (operator: "highlighting scrolling aims to be vertically
      // centered within the bottom/bottom top/top rule * so from top of window,
      // highlighting continues normally * then upon reaching/passing the
      // vertical center, the bubble will move up to ensure the text being
      // highlighted is in full view"): back on, in that rule.
      var FOCUS_DEAD_ZONE = 0.25;
      function focusSpeaking(span) {
        if (!span || !container.contains(span)) { return; }
        var rect = span.getBoundingClientRect();
        var middle = contentY(span) + rect.height / 2;
        // 2026-10-15 (operator: "Add a dead zone: chat holds still until
        // highlight is ~1/4 of the window, below centre * then one smooth move to
        // re-centre, unless bottom/bottom is reached first"): no move until the
        // lit piece is FOCUS_DEAD_ZONE of the window below centre (the window's
        // own scroll included), then one move back to centre.
        var view = chatViewHeight();
        var below = middle - (container.scrollTop + view / 2);
        if (below < view * FOCUS_DEAD_ZONE && !chatMove) { return; }
        var target = middle - view / 2;
        // It stops at the chat's end.
        target = Math.min(target, chatContentEnd());
        if (target > container.scrollTop + 4) { scrollChatTo(target); }
      }

      // 2026-09-28: the operator's words come from the /api/get_logs lines
      // Wire-Pod already writes ("<time>: Intent matched: ..., transcribed text:
      // '<words>', device: ..."), and Wire-Pod's own chat replies from its
      // timestamped "LLM response for" lines - both carry a time, so each is shown
      // once. Song/tool lines come from the numbered feed (pollFeed, 2026-09-30).
      // ":" then an optional space - the log page trims a line, so a reply that
      // starts on the next line leaves "LLM response for <ESN>:" with no space
      // (live 2026-10-01 15:54, no bubble).
      var KG_REPLY_RE = /^(\d{4}\.\d{2}\.\d{2} \d{2}:\d{2}:\d{2}): LLM response for (\S+?):\s?([\s\S]*)$/;
      // The animation marker and the prompt's placeholder word come from
      // gopod_brobots_settings.json, the same file Wire-Pod reads (2026-10-01,
      // operator: one source of truth). Until it loads, nothing is stripped.
      var cueMarkerPattern = null;
      var placeholderPattern = null;
      // A scripted song line comes already split by the song engine, its pieces
      // joined with scripted_split.segment_marker (an invisible character) - the
      // same pieces the robot says, so those are what light up.
      var segmentMarker = null;
      // The chat split (live_split in the settings file, the same rule Wire-Pod
      // says a reply with): the characters that end a sentence and a comma
      // piece. The defaults hold until the settings load.
      var liveSentenceEnds = ".!?";
      var liveCommaEnds = ",";
      // The song's countdown lines (countdown.role in the settings file).
      var countdownRole = "Countdown";
      // The song's Bingo ball lines (bingo_ball in the settings file).
      var ballRole = "Bingo Ball";
      function escapeForRegExp(text) {
        return String(text).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      }
      function loadBrobotsSettings() {
        return fetch("/gopod_brobots_settings.json", { cache: "no-store" })
          .then(function (r) { return r.json(); })
          .then(function (settings) {
            var marker = String(settings.animation_marker || "");
            if (marker) {
              var last = escapeForRegExp(marker.slice(-1));
              // 2026-10-04: a cue in markdown bold ("**confused**___") is a cue too.
              cueMarkerPattern = new RegExp("\\*{0,3}[A-Za-z][A-Za-z0-9-]*\\*{0,3}[ \\t]*" + escapeForRegExp(marker) + last + "*[ \\t]*", "g");
            }
            if (settings.scripted_split && settings.scripted_split.segment_marker) {
              segmentMarker = settings.scripted_split.segment_marker;
            }
            if (settings.countdown && settings.countdown.role) {
              countdownRole = settings.countdown.role;
              countdownEndText = settings.countdown.end_text || countdownEndText;
            }
            if (settings.addressing) { addressing = settings.addressing; }
            if (settings.system_bubble && settings.system_bubble.label) { systemLabel = settings.system_bubble.label; }
            if (settings.numpad_ptt) { numpad.cfg = settings.numpad_ptt; }
            if (settings.system_menu) { systemMenuData = settings.system_menu.groups || []; systemMenuIntro = settings.system_menu.intro || ""; systemMenuLook = settings.system_menu.look || {}; }
            if (settings.math && settings.math.popup_title) { mathPopupTitle = settings.math.popup_title; }
            if (settings.digit_highlight) { digitHighlight = settings.digit_highlight; }
            if (settings.word_reveal) { wordReveal = settings.word_reveal; }
            // 2026-10-18 (operator: "ensure flexible data/info values/variables are in
            // loadable files * code is dumb flow guiding machine"): the window's
            // timings and sizes come from chat_window (the numbers here are the fallbacks).
            if (settings.chat_window) {
              var cw = settings.chat_window;
              if (typeof cw.max_bubbles === "number") { CHAT_MAX_BUBBLES = cw.max_bubbles; }
              if (typeof cw.max_seen === "number") { CHAT_MAX_SEEN = cw.max_seen; }
              if (typeof cw.baton_max_ms === "number") { CHAT_BATON_MAX_MS = cw.baton_max_ms; }
              if (typeof cw.bubble_ready_max_ms === "number") { BUBBLE_READY_MAX_MS = cw.bubble_ready_max_ms; }
              if (typeof cw.reply_delay_ms === "number") { CHAT_REPLY_DELAY_MS = cw.reply_delay_ms; }
              if (typeof cw.top_gap_px === "number") { CHAT_TOP_GAP_PX = cw.top_gap_px; }
              if (typeof cw.scroll_ms === "number") { CHAT_SCROLL_MS = cw.scroll_ms; }
              if (typeof cw.focus_dead_zone === "number") { FOCUS_DEAD_ZONE = cw.focus_dead_zone; }
              if (typeof cw.countdown_hold_ms === "number") { COUNTDOWN_HOLD_MS = cw.countdown_hold_ms; }
              if (typeof cw.countdown_hold_most_ms === "number") { COUNTDOWN_HOLD_MOST_MS = cw.countdown_hold_most_ms; }
              if (typeof cw.voice_lit_wait_ms === "number") { voiceLitWaitMs = cw.voice_lit_wait_ms; }
              if (typeof cw.marquee_spacing_px === "number") { MARQUEE_SPACING_PX = cw.marquee_spacing_px; }
              if (typeof cw.marquee_inset_px === "number") { MARQUEE_INSET_PX = cw.marquee_inset_px; }
              if (typeof cw.marquee_pattern_ms === "number") { MARQUEE_PATTERN_MS = cw.marquee_pattern_ms; }
              if (typeof cw.narrator_fade_in_s === "number") { NARRATOR_FADE_IN_S = cw.narrator_fade_in_s; }
              if (typeof cw.narrator_fade_out_max_s === "number") { NARRATOR_FADE_OUT_MAX_S = cw.narrator_fade_out_max_s; }
              if (typeof cw.narrator_fade_out_end_s === "number") { NARRATOR_FADE_OUT_END_S = cw.narrator_fade_out_end_s; }
              if (typeof cw.reveal_rest_ms === "number") { REVEAL_REST_MS = cw.reveal_rest_ms; }
              if (typeof cw.speaking_max_ms === "number") { SPEAKING_MAX_MS = cw.speaking_max_ms; }
              if (typeof cw.voice_ready_max_ms === "number") { VOICE_READY_MAX_MS = cw.voice_ready_max_ms; }
              if (settings.page_sync && typeof settings.page_sync.viewer_catch_up_events === "number") { VIEWER_CATCH_UP_EVENTS = settings.page_sync.viewer_catch_up_events; }
              if (typeof cw.typed_text_min_px === "number") { TYPED_TEXT_MIN_PX = cw.typed_text_min_px; fitTypedRow(); }
            }
            if (settings.highlight_scroll && settings.highlight_scroll.dead_zone >= 0) { FOCUS_DEAD_ZONE = Number(settings.highlight_scroll.dead_zone); }
            if (settings.live_weather && settings.live_weather.military_label) { militaryLabel = settings.live_weather.military_label; }
            if (settings.llm) {
              llmWarm.warmText = settings.llm.warmup_text || llmWarm.warmText;
              if (settings.background_loop) { backgroundLoop.setup(settings.background_loop); }
              llmWarm.waitText = settings.llm.wait_text || llmWarm.waitText;
              if (settings.llm.warmup_check_seconds) { llmWarm.checkMs = settings.llm.warmup_check_seconds * 1000; }
              llmWarm.readyText = settings.llm.ready_text || llmWarm.readyText;
              llmWarm.expectedSeconds = Number(settings.llm.expected_seconds) || llmWarm.expectedSeconds;
              llmWarm.leds = Number(settings.llm.leds) || llmWarm.leds;
            }
            if (settings.bingo_ball && settings.bingo_ball.role) {
              ballRole = settings.bingo_ball.role;
            }
            if (settings.interview_chapters && settings.interview_chapters.role) {
              chapterRole = settings.interview_chapters.role;
              chapterLabel = settings.interview_chapters.label || chapterLabel;
            }
            if (settings.value_points && settings.value_points.score_role) {
              scoreRole = settings.value_points.score_role;
            }
            if (settings.sign_off && settings.sign_off.system_role) { systemRole = settings.sign_off.system_role; }
            if (settings.system_set && settings.system_set.role) { terminalRole = settings.system_set.role; }
            if (settings.system_search) { searchSettings = settings.system_search; }
            if (settings.managers) {
              displayHost = settings.managers.display_host || "";
              managerHosts = {};
              if (displayHost) { managerHosts[displayHost] = true; }
              (settings.managers.takes || []).forEach(function (t) { if (t.host) { managerHosts[t.host] = true; } });
              Object.keys(settings.managers.personas || {}).forEach(function (k) { var h = settings.managers.personas[k].host; if (h) { managerHosts[h] = true; } });
            }
            if (settings.clean_politics && settings.clean_politics.countries) {
              Object.keys(settings.clean_politics.countries).forEach(function (k) {
                var c = String(settings.clean_politics.countries[k] || "").toUpperCase();
                if (c) { flagCodes[c] = true; }
              });
            }
            // 2026-10-17: the stings loaded once, ready to play again and again.
            if (settings.reporter_voices && settings.reporter_voices.system_stings) {
              var stings = settings.reporter_voices.system_stings;
              preloadStings(Object.keys(stings).map(function (k) { return stings[k] && stings[k].file; }).filter(Boolean));
            }
            if (settings.topic_pictures && settings.topic_pictures.role) {
              pictureRole = settings.topic_pictures.role;
            }
            if (settings.topic_pictures && settings.topic_pictures.popup_text) { popupText = settings.topic_pictures.popup_text; }
            if (settings.topic_pictures && typeof settings.topic_pictures.egg_zoom_text === "string") { eggZoomText = settings.topic_pictures.egg_zoom_text.trim(); }
            if (settings.qr_codes && settings.qr_codes.codes) { qrCodes = settings.qr_codes.codes; secondQr = settings.qr_codes.second || null; addQrCodesToAll(); }
            if (settings.weather_card && settings.weather_card.faces) { weatherCard = settings.weather_card; }
            if (settings.chat_audio && typeof settings.chat_audio.volume === "number") {
              chatVolume = Math.max(0, Math.min(1, settings.chat_audio.volume));
            }
            if (settings.topic_pictures && settings.topic_pictures.in_question) {
              picturesInQuestion = settings.topic_pictures.in_question;
            }
            if (settings.reporter_voices) {
              reporterRole = settings.reporter_voices.role || reporterRole;
              hostNames = settings.reporter_voices.hosts || hostNames;
              hostLabels = settings.reporter_voices.labels || hostLabels;
              displayNames = settings.display_names || displayNames;
              if (settings.skip_gate) { skipGate.enable_on = settings.skip_gate.enable_on || []; }
              if (settings.hourglass) { hourglassCfg = settings.hourglass; }
              if (settings.bubble_kinds) { bubbleKinds = settings.bubble_kinds.rules || []; }
              openingEndText = settings.reporter_voices.opening_end_text || openingEndText;
              heardMarker = settings.reporter_voices.heard_marker || heardMarker;
              generatingMarker = settings.reporter_voices.generating_marker || generatingMarker;
              generatingText = settings.reporter_voices.generating_text || generatingText;
              songAudioRole = settings.reporter_voices.song_audio_role || songAudioRole;
              narratorRole = settings.reporter_voices.narrator_role || narratorRole;
              narratorLead = Number(settings.reporter_voices.narrator_lead_seconds) || 0;
              if (settings.reporter_voices.gap_seconds >= 0) { voiceGapMs = settings.reporter_voices.gap_seconds * 1000; }
            }
            if (settings.live_split) {
              liveSentenceEnds = settings.live_split.sentence_ends || liveSentenceEnds;
              liveCommaEnds = settings.live_split.comma_ends || liveCommaEnds;
            }
            if (settings.placeholder_word) {
              placeholderPattern = new RegExp("^\\s*" + escapeForRegExp(settings.placeholder_word) + "\\s*\\.\\.\\.\\s*[A-Za-z_]+\\s+");
            }
          })
          .catch(function () { /* no settings - cue markers show as written */ });
      }
      function kgReplyText(raw) {
        var out = String(raw)
          .replace(/\|\|\|[01C]\|\|\|/g, "")
          .replace(/\{\{[^}]*\}\}/g, "");
        if (placeholderPattern) { out = out.replace(placeholderPattern, ""); }
        // 2026-09-30: Wire-Pod already removes the animation cue from this line;
        // a leftover marker is removed here. A leading "Ah..." is speech and stays.
        if (cueMarkerPattern) { out = out.replace(cueMarkerPattern, ""); }
        return out.trim();
      }

      // 2026-09-29: one bubble per spoken line. If the same reply shows up in two
      // log formats back to back (a "Brobot N:" line and an "LLM response for"
      // line), only the first becomes a bubble.
      var lastRobotBubble = { who: "", norm: "", format: "" };
      function robotBubbleOnce(who, text, format, then) {
        var norm = String(text).toLowerCase().replace(/[^a-z0-9]+/g, "");
        if (!norm) { if (then) { then(); } return "no words"; }
        if (lastRobotBubble.who === who && lastRobotBubble.norm === norm && lastRobotBubble.format !== format) { if (then) { then(); } return "same words as the bubble before"; }
        lastRobotBubble = { who: who, norm: norm, format: format };
        appendBubble(who, text, format, undefined, undefined, then);
        return "";
      }
      // 2026-10-17 (operator: "If a reply bubble goes missing or shows twice ...
      // can you account for to reveal this possible glitch, in all-logs?"): each
      // live reply's fate goes to Wire-Pod's log - "GOPOD_PAGE live reply <id>
      // Brobot N: shown", or "NOT shown - <why>" - under its "GOPOD_FEED <id>".
      function noteLiveReply(e, why) {
        var body = new URLSearchParams();
        body.set("note", "live reply " + e.id + " " + e.who + ": " + (why ? "NOT shown - " + why : "shown"));
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: body }).catch(function () { /* Wire-Pod gone */ });
      }

      var bubblesPolling = false;
      function pollBubbles() {
        if (bubblesPolling) { return; }
        bubblesPolling = true;
        fetch("/api/get_logs")
          .then(function (r) { return r.text(); })
          .then(function (text) {
            var lines = text.split("\n").map(function (l) { return l.trim(); }).filter(Boolean);
            // 2026-09-28: a multi-line robot reply (the Brobot's rich display) is logged as one
            // "Brobot N: ..." line followed by its remaining lines with no prefix. Fold
            // those continuation lines into the robot line above them, so the whole
            // reply is one bubble. A continuation ends at the next timestamped line,
            // "Brobot N:" line, or "Intent matched" line.
            var folded = [];
            var inReply = false;
            lines.forEach(function (line) {
              var stamped = /^\d{4}\.\d{2}\.\d{2} \d{2}:\d{2}:\d{2}: /.test(line);
              var robotLine = /^(Brobot 1|Brobot 2):/.test(line) || KG_REPLY_RE.test(line);
              if (robotLine) { folded.push(line); inReply = true; return; }
              // Any other speaker or song-marker line (Brobot 3, GOPOD_SONG_START, ...)
              // ends a reply too, so it never folds into the bubble above it.
              // 2026-10-10 (live 22:18: Brobot 2's reply came out four times, each
              // with the lines after it glued on): every chat-feed line Wire-Pod
              // logs ("Goverlord Says:", "Countdown:", "Reporter Voice:", ...) ends it too.
              var otherRole = /^(Brobot \d+|GOPOD_[A-Z_]+|Goverlord Says|Goverlord|Operator|Countdown|Reporter Voice|Bingo Ball|Narrator|Song Audio|Picture|Chapter):/.test(line);
              if (stamped || otherRole) { inReply = false; folded.push(line); return; }
              if (inReply) { folded[folded.length - 1] += "\n" + line; return; }
              folded.push(line);
            });
            lines = folded;
            lines.forEach(function (line) {
              var said = line.match(/^(\d{4}\.\d{2}\.\d{2} \d{2}:\d{2}:\d{2}): Intent matched: (.*?), transcribed text: '(.*)', device: (\S+)$/);
              if (said) {
                if (isSeen(line) || feedQueued[line] || !said[3].trim()) { return; }
                // 2026-10-07: in the one line with every other bubble, in time order.
                queueBubble(line, logTimeMs(said[1]), function () {
                  // 2026-09-30: a robot told to sleep answers "zzz" (operator's ask);
                  // both bubbles are placed like any other, not moved to the top.
                  var sleep = said[2] === "intent_system_sleep";
                  // 2026-10-08 (operator: "if I backpack brobot 1, defaults to
                  // 'Operator to Brobots'"): a line the robot heard itself goes
                  // to that Brobot, unless it opens with a Brobot's name.
                  var heardBy = operatorTarget(said[3].trim());
                  if (heardBy === "Brobots" && ESN_TO_BROBOT[said[4]]) { heardBy = ESN_TO_BROBOT[said[4]]; }
                  appendBubble("Operator", said[3].trim(), sleep ? "plain" : undefined, heardBy, "Backpack");
                  if (sleep && ESN_TO_BROBOT[said[4]]) {
                    appendBubble(ESN_TO_BROBOT[said[4]], "zzz", "brobot");
                  }
                });
                return;
              }
              // 2026-09-29: Wire-Pod's own chat replies (a plain "Hey Vector" question,
              // not a GOPOD song or go search) are logged as "<time>: LLM response for
              // <ESN>: <reply>". Shown as that robot's bubble, animation cue removed.
              // 2026-10-17 (operator: "one event list, single source of truth"):
              // Wire-Pod's own chat replies ("LLM response for <ESN>: ...") come on
              // the numbered feed now (via "llm", showFeedEvent), not from this log.
              // 2026-09-30: "Brobot N: ..." lines (songs, the interview, go search,
              // the aliases) carry no time or number here, so they are no longer
              // read from this log - they come from Wire-Pod's numbered chat feed
              // instead (pollFeed below), where a repeated line or a second run of
              // the same song still gets its own bubble.
            });
          })
          .catch(function () { /* log endpoint not reachable yet */ })
          .then(function () { bubblesPolling = false; });
      }

      // 2026-10-01 (operator: "after please wait, add a countdown timer in the
      // brobots window"): a small centred bubble counts the seconds of the
      // rattle's wait down, then goes. Never saved with the chat. A countdown
      // already over (a page opened later) is skipped.
      // 2026-10-02 (operator: "[brobot icon] Brobot N: / [chat text]"): each bubble
      // opens with a header row - the animated eyes icon and the speaker's name -
      // and the chat text gets the full bubble width below it.
      // 2026-10-04 (operator: "[icon] Operator to Brobot N [icon] ... if unnamed,
      // ... Operator to Brobots"): the operator's bubble names who it went to -
      // the Brobot the line opens with ("Brobot 1, can you hear me", the same
      // name forms Wire-Pod answers to, addressing in the settings file), the
      // Brobot a typed line went to, or the Brobots - and ends with that Brobot's
      // icon in his own colour (both for the Brobots).
      var addressing = null;
      // 2026-10-23 (operator: "When 'All' selected, whatever Brobot N is called in
      // the input first ... starting with brobot N answering first"): the Brobot
      // named first anywhere in the line (Wire-Pod's gopodFirstNamed, the same names).
      function firstNamedBrobot(text) {
        if (!addressing || !addressing.names) { return ""; }
        var words = String(text).toLowerCase().split(/[^a-z0-9']+/).filter(Boolean);
        var best = "", bestAt = words.length;
        ["brobot", "robot"].forEach(function (kind) {
          ["Brobot 1", "Brobot 2"].forEach(function (brobot) {
            ((addressing.names[brobot] || {})[kind] || []).forEach(function (form) {
              var name = String(form).toLowerCase().split(/[^a-z0-9']+/).filter(Boolean);
              for (var start = 0; name.length && start + name.length <= words.length && start < bestAt; start++) {
                if (words.slice(start, start + name.length).join(" ") === name.join(" ")) { best = brobot; bestAt = start; break; }
              }
            });
          });
        });
        return best;
      }
      function operatorTarget(text) {
        if (!addressing || !addressing.names) { return "Brobots"; }
        var words = String(text).toLowerCase().split(/[^a-z0-9']+/).filter(Boolean);
        var within = addressing.start_within_words || 2;
        var kinds = ["brobot", "robot"], brobots = ["Brobot 1", "Brobot 2"];
        for (var k = 0; k < kinds.length; k++) {
          for (var b = 0; b < brobots.length; b++) {
            var forms = (addressing.names[brobots[b]] || {})[kinds[k]] || [];
            for (var f = 0; f < forms.length; f++) {
              var name = String(forms[f]).toLowerCase().split(/[^a-z0-9']+/).filter(Boolean);
              for (var start = 0; start < within && start + name.length <= words.length; start++) {
                if (name.length && words.slice(start, start + name.length).join(" ") === name.join(" ")) { return brobots[b]; }
              }
            }
          }
        }
        return "Brobots";
      }
      // 2026-10-04 (operator: "this operator bubble however, has operator icons at
      // both ends indicating a system response, Operator to Goverlord", "same chat
      // brobot icon in grey"): an answer to Goverlord (Record Session's
      // keep-or-delete) - the grey icon at both ends of the header.
      // 2026-10-04 (operator: "label these timers as system messages ... for all
      // timers [icon] Goverlord: GOPOD Brain [icon]"): the timers (the LLM
      // warm-up, the rattle countdown) carry the same grey-icon header, labelled
      // system_bubble.label from the settings file.
      var SYSTEM_TARGET = "Goverlord";
      var systemLabel = "Goverlord: GOPOD Brain";
      function systemHead(label) {
        var sysHead = bubbleHead(label, true);
        sysHead.classList.add("system-head");
        sysHead.firstChild.classList.add("to-system");
        var endIcon = document.createElement("div");
        endIcon.className = "brobot-bubble-avatar to-system";
        sysHead.appendChild(endIcon);
        return sysHead;
      }
      // 2026-10-08: the first icon of an operator line is the operator's own
      // (from-operator, gopod_chat_bubbles.css's icon template).
      // 2026-10-10 (operator: "'Operator' now update message according to input *
      // "[input] to Brobot [N]", "KP[N] to Brobot [N]" ... "Mic to Brobot [N]",
      // "KP1 to Brobot 1", "Reporter [N] to Brobot [N]""): the same white-eyed
      // Operator bubble, its label naming the input (via) when it's known.
      // 2026-10-17 (operator: "When \"starting with brobot 2\" * [icon] [input]
      // to Brobots [icon brobot eyes 1] [icon brobot eyes 2] * switch to [icon
      // brobot eyes 2] [icon brobot eyes 1], as a visual indicator"): the same
      // words Wire-Pod takes as "Brobot 2 first" (gopodAllStartsWith).
      var BROBOT2_FIRST = /\b(?:start(?:ing)?\s+with|begin(?:ning)?\s+with)\s+(?:(?:brobot|bro\s*bo(?:ugh)?t|bro\s*but)\s*(?:2|two|too|to)|pip)\b|\b(?:(?:brobot|bro\s*bo(?:ugh)?t|bro\s*but)\s*(?:2|two|too|to)|pip)\s+(?:goes\s+|answers\s+)?first\b/i;
      function operatorHead(to, via, text) {
        if (to === SYSTEM_TARGET) {
          var toSystem = systemHead("Operator to " + to);
          toSystem.firstChild.classList.add("from-operator");
          return toSystem;
        }
        var head = bubbleHead((via || "Operator") + " to " + to, true);
        head.firstChild.classList.add("from-operator");
        var icons = to === "Brobots" ? ["brobot-1", "brobot-2"] : [to === "Brobot 2" ? "brobot-2" : "brobot-1"];
        if (to === "Brobots" && (BROBOT2_FIRST.test(String(text || "")) || firstNamedBrobot(text) === "Brobot 2" || (allSecondFirst[String(text || "")] && firstNamedBrobot(text) !== "Brobot 1"))) { icons.reverse(); }
        icons.forEach(function (cls) {
          var icon = document.createElement("div");
          icon.className = "brobot-bubble-avatar to-" + cls;
          head.appendChild(icon);
        });
        return head;
      }

      // 2026-10-13 (operator: "Reporter & Brobot bubbles: [white/transparent EYED
      // icon]"; then "correct to \"Reporter 1: Brobot 3\" follows same format as
      // \"Goverlord: GOPOD Brain\""): a reporter's header, its label as written
      // (no colon after it), white-eyed icons at both ends.
      function reporterHead(label) {
        var head = bubbleHead(label, true);
        head.classList.add("system-head");
        head.firstChild.classList.add("from-operator");
        var endIcon = document.createElement("div");
        endIcon.className = "brobot-bubble-avatar from-operator";
        head.appendChild(endIcon);
        return head;
      }
      function bubbleHead(who, noColon) {
        var head = document.createElement("div");
        head.className = "brobot-bubble-head";
        var avatar = document.createElement("div");
        avatar.className = "brobot-bubble-avatar";
        var name = document.createElement("span");
        name.className = "brobot-bubble-who";
        name.textContent = noColon ? who : who + ":";
        head.appendChild(avatar);
        head.appendChild(name);
        return head;
      }

      // 2026-10-04 (operator: "timer too fast to hear rattle audio - have visuals
      // respond to what's there"): the timer stays until the song says the rattle
      // has finished (countdown.end_text from the countdown role, see pollFeed);
      // the seconds are only a fallback, plus 10, if that never comes.
      var countdownEndText = "done";
      function endCountdowns() {
        countdownDoneAt = 0;
        Array.prototype.slice.call(container.querySelectorAll(".brobot-bubble.countdown")).forEach(function (b) { b.remove(); });
      }
      // 2026-10-11 (operator: "Timer displays and stays in place at least 3s, or
      // until ball image & "Brobot 1: Bingo ball captured ..." are ready to
      // display"): at the rattle's end the timer stays; the next bubble (the big
      // ball, or the next line) takes its place, never sooner than
      // COUNTDOWN_HOLD_MS after the end. A timer with nothing after it goes after
      // COUNTDOWN_HOLD_MOST_MS.
      var COUNTDOWN_HOLD_MS = 3000, COUNTDOWN_HOLD_MOST_MS = 20000, countdownDoneAt = 0;
      function countdownHold() {
        if (!container.querySelector(".brobot-bubble.countdown")) { countdownDoneAt = 0; return; }
        countdownDoneAt = Date.now();
        var at = countdownDoneAt;
        setTimeout(function () { if (countdownDoneAt === at) { endCountdowns(); } }, COUNTDOWN_HOLD_MOST_MS);
      }
      // 2026-10-11 (operator: "timer animation: upon display, timer stays in
      // place, does not move or hide, scrolls normally"): only the big ball
      // replaces it; before any other bubble it stays where it is, a bubble
      // like the rest, and scrolls up with the chat.
      function keepCountdowns() {
        countdownDoneAt = 0;
        Array.prototype.slice.call(container.querySelectorAll(".brobot-bubble.countdown")).forEach(function (b) {
          clearTimeout(b._gone);
          b.classList.remove("countdown");
          b.classList.add("countdown-kept");
        });
      }
      function afterCountdown(then, ball) {
        if (!countdownDoneAt) { then(); return; }
        if (!ball) { keepCountdowns(); then(); return; }
        var wait = COUNTDOWN_HOLD_MS - (Date.now() - countdownDoneAt);
        setTimeout(function () { endCountdowns(); then(); }, Math.max(0, wait));
      }
      function showCountdown(seconds, sentMs) {
        if (!(seconds > 0)) { return; }
        var leftMs = (seconds + 10) * 1000 - Math.max(0, Date.now() - (sentMs || Date.now()));
        if (leftMs <= 0) { return; }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator countdown";
        bubble.appendChild(systemHead(systemLabel));
        var row = document.createElement("div");
        row.className = "warmup-timer";
        bubble.appendChild(row);
        addJuggleTimer(bubble, row);
        bubble._gone = setTimeout(function () { bubble.remove(); }, leftMs);
        container.appendChild(bubble);
        revealWhenPlaced(bubble);
      }

      // The countdown's moving part - the juggled balls and the rocking hourglass
      // - put into box; it runs until bubble leaves the page. Used by the rattle
      // countdown and the LLM warm-up bubble.
      function addJuggleTimer(bubble, box) {
        // 2026-10-02 (operator: "give me something for 'Juggling Big Shiny Bingo
        // Balls. Please Wait.'"): five small Bingo balls, B I N G O, juggled
        // beside the count (B, N, O in Brobot 1's colour; I, G in Brobot 2's).
        var juggle = document.createElement("div");
        juggle.className = "countdown-juggle";
        ["B", "I", "N", "G", "O"].forEach(function (letter) {
          var ball = document.createElement("span");
          ball.className = "juggle-ball ball-" + letter.toLowerCase();
          ball.textContent = letter;
          juggle.appendChild(ball);
        });
        box.appendChild(juggle);
        // 2026-10-04 (operator: "remove motion lines ... make the 5 balls bigger &
        // simply move them up & down ... Make balls stay in position horizontally,
        // and switch vertical between UPPER/LOWER"; earlier "mildly 'shake' the
        // hourglass at 12fps"): each ball keeps its place, B I N G O left to right,
        // half a ball apart so neighbours overlap. B, N, O and
        // I, G glide up and down opposite each other, swapping UPPER and LOWER;
        // the upper ones are drawn over the lower ones. The hourglass rocks a
        // little. No numbers: the hourglass is the wait.
        var jugBalls = juggle.querySelectorAll(".juggle-ball");
        var glass = document.createElement("span");
        glass.className = "countdown-hourglass";
        glass.textContent = "\u23F3";
        box.appendChild(glass);
        // 2026-10-04 (operator: "make balls swap start first at 6 fps; start 12fps
        // hourglass to start, offset by +3"): the balls move 6 frames a second and
        // start at once; the hourglass rocks 12 frames a second, starting 3 of its
        // frames (0.25 s) later.
        var jugFrame = 0;
        var jugTimer = setInterval(function () {
          if (!bubble.isConnected && jugFrame > 0) { clearInterval(jugTimer); return; }
          // Places fit the juggle box, so they scale with the phone layout.
          var ballSize = jugBalls.length ? jugBalls[0].offsetWidth : 56;
          var step = Math.max(0, juggle.clientWidth - ballSize) / Math.max(1, jugBalls.length - 1);
          var travel = Math.max(0, juggle.clientHeight - ballSize);
          // 0 = UPPER, 1 = LOWER, gliding over 9 frames (1.5 s) each way.
          var swing = (1 - Math.cos(jugFrame / 9 * Math.PI)) / 2;
          for (var b = 0; b < jugBalls.length; b++) {
            var low = b % 2 === 0 ? swing : 1 - swing;  // B N O vs I G
            jugBalls[b].style.transform = "translate(" + (b * step).toFixed(0) + "px," + (-(1 - low) * travel).toFixed(0) + "px)";
            jugBalls[b].style.zIndex = low < 0.5 ? 2 : 1;
          }
          jugFrame += 1;
        }, 1000 / 6);
        rockHourglass(bubble, glass, 3 * 1000 / 12);
      }

      // The hourglass rocks a little, 12 frames a second, starting after
      // delayMs, until bubble leaves the page. Used by the rattle countdown and
      // the LLM warm-up bubble.
      function rockHourglass(bubble, glass, delayMs) {
        var glassFrame = 0;
        setTimeout(function () {
          var glassTimer = setInterval(function () {
            if (!bubble.isConnected) { clearInterval(glassTimer); return; }
            if (bubble.classList.contains("ready")) { glass.style.transform = ""; clearInterval(glassTimer); return; }
            var rock = [0, -7, 0, 7][glassFrame % 4];
            glass.style.transform = "rotate(" + rock + "deg) translateX(" + (rock / 7) + "px)";
            glassFrame += 1;
          }, 1000 / 12);
        }, delayMs);
      }

      // 2026-10-03 (operator: "I prefer a pre-warm up and then have it stay
      // loaded", "the bubble disappears", "Use the timer as well as part of a
      // visual cue during the warm up"): while Wire-Pod loads the Brobots' LLM
      // after a start (gopod_llm_warmup.go), one operator bubble shows the
      // warm-up text (llm.warmup_text in the settings file) with the countdown's
      // juggled balls and hourglass, and goes when the warm-up is done. Read from
      // its own file, never saved with the chat, never in the chat memory.
      var warmupBubble = null;
      // 2026-10-12 (operator: "The opening llm system bubble completed before I
      // started, so the bubble stayed on top rather than under"): a warm-up
      // bubble that finished before the reporters' opening still moves under
      // each host line, until the opening ends.
      var warmupBeforeOpening = null, openingDone = false;
      var warmupPolling = false;
      // 2026-10-04 (operator: "make it like a horizontal battery power level
      // indicator ... a series of vertical LEDs, no bingo balls ... left to right,
      // red to green, cold llm to fully warmed up llm ... message changes, LLM
      // Warming Up to LLM Ready ... use same hour glass following LLM power
      // level"; then "it is to stay. Keep it in"): the warm-up bubble is a
      // battery of llm.leds LEDs, red to green. While warming they light one by
      // one over llm.expected_seconds from the start (never the last one); when
      // the model is loaded all are lit, the text turns to llm.ready_text and the
      // hourglass settles (sand run through). It stays in the window, saved with
      // the chat like the Bingo balls. One bubble per warm-up (Wire-Pod's start,
      // each Bingo test), known by its start time.
      var llmWarm = { warmText: "LLM Warming Up", readyText: "LLM Ready", waitText: "Please wait.", checkMs: 3000, expectedSeconds: 120, leds: 10 };
      function buildWarmupBubble(startedMs) {
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator warmup";
        bubble.dataset.started = String(startedMs);
        bubble.appendChild(systemHead(systemLabel));
        // 2026-10-04 (operator: "ReOrder to (center-aligned): [grey icon] Goverlord:
        // GOPOD Brain [grey icon] / [llm power level] ⌛ / LLM Warming/Ready"): the
        // header, the battery row, then the text under it.
        var text = document.createElement("span");
        text.className = "brobot-bubble-text";
        var row = document.createElement("div");
        row.className = "warmup-timer";
        var battery = document.createElement("div");
        battery.className = "llm-battery";
        battery.setAttribute("role", "img");
        for (var i = 0; i < llmWarm.leds; i++) {
          var led = document.createElement("span");
          led.className = "llm-led";
          // Red (hue 0) on the left to green (hue 120) on the right.
          led.style.setProperty("--led-color", "hsl(" + Math.round(120 * i / Math.max(1, llmWarm.leds - 1)) + ", 90%, 50%)");
          battery.appendChild(led);
        }
        row.appendChild(battery);
        var glass = document.createElement("span");
        glass.className = "countdown-hourglass warmup-hourglass";
        glass.innerHTML = hourglassSprite(0);
        row.appendChild(glass);
        bubble.appendChild(row);
        bubble.appendChild(text);
        container.appendChild(bubble);
        if (!restoringChat) {
          rockHourglass(bubble, glass, 0);
          revealWhenPlaced(bubble);
        }
        return bubble;
      }
      // 2026-10-09 (operator: "maybe this time say LLM ready for Bingo?"): a song's
      // warm-up names its own ready text (its knobs.json llm_ready_text), else
      // llm.ready_text.
      function showWarmupLevel(bubble, lit, ready, readyText, warmingText) {
        // 2026-10-13: while warming, Wire-Pod's message ("Voices are being
        // generated. Please wait.", "LLM Warming Up. Please wait.", "Please wait.").
        var label = ready ? (readyText || llmWarm.readyText) : (warmingText || llmWarm.warmText);
        var labelEl = bubble.querySelector(".brobot-bubble-text");
        // 2026-10-23: the same words keep their voice's lit line (voiceForLine).
        if (voiceLineWords(labelEl.textContent) !== voiceLineWords(label)) {
          labelEl.textContent = label;
          // Its lines already said stay said; the next line is tied before it shows.
          Object.keys(bubble._spoken || {}).forEach(function (t) {
            var said = voiceLineSpan(labelEl, t);
            if (said) { said.classList.add("spoken"); }
          });
          voiceTieAhead();
        }
        // 2026-10-16 (operator: "NEW system message ... lines"): one line each.
        labelEl.style.whiteSpace = "pre-line";
        var leds = bubble.querySelectorAll(".llm-led");
        for (var i = 0; i < leds.length; i++) { leds[i].classList.toggle("lit", i < lit); }
        bubble.querySelector(".llm-battery").setAttribute("aria-label", label + ", " + lit + " of " + leds.length);
        bubble.classList.toggle("ready", !!ready);
        // 2026-10-06 (operator: "make the (rocking) hourglass animation empty the
        // sand from top full to bottom full over the range of the warming level
        // ... 10 LEDS, so 5 hourglass sprites?"): two LEDs a sprite.
        var sprites = HOURGLASS_SPRITES - 1;
        var stage = ready ? sprites : Math.min(sprites, Math.floor(lit * HOURGLASS_SPRITES / Math.max(1, leds.length)));
        var glass = bubble.querySelector(".warmup-hourglass");
        if (glass && glass.dataset.stage !== String(stage)) {
          glass.dataset.stage = String(stage);
          glass.innerHTML = hourglassSprite(stage);
        }
      }
      // The warm-up hourglass, sprite stage 0 (top full) to 4 (bottom full): the
      // sand in the top bulb drains to the neck, the bottom bulb fills from its
      // base, with a thin stream between while it runs.
      var HOURGLASS_SPRITES = 5;
      var hourglassCfg = {};
      function hourglassSprite(stage, steps) {
        var f = stage / ((steps || HOURGLASS_SPRITES) - 1); // share of the sand at the bottom
        var top = 1 - f;
        var p = function (x, y) { return x.toFixed(2) + "," + y.toFixed(2); };
        var topSand = top > 0 ? "<polygon class=\"hg-sand\" points=\"" + p(12 - 8 * top, 16 - 12 * top) + " " + p(12 + 8 * top, 16 - 12 * top) + " " + p(12, 16) + "\"/>" : "";
        var ys = 28 - 12 * f, hw = 8 * (1 - f);
        var bottomSand = f > 0 ? "<polygon class=\"hg-sand\" points=\"" + p(4, 28) + " " + p(20, 28) + " " + p(12 + hw, ys) + " " + p(12 - hw, ys) + "\"/>" : "";
        var stream = f > 0 && top > 0 ? "<line class=\"hg-stream\" x1=\"12\" y1=\"16\" x2=\"12\" y2=\"" + ys.toFixed(2) + "\"/>" : "";
        return "<svg class=\"hg\" viewBox=\"0 0 24 32\" aria-hidden=\"true\">" +
          "<rect class=\"hg-frame\" x=\"2\" y=\"1\" width=\"20\" height=\"3\" rx=\"1\"/>" +
          "<rect class=\"hg-frame\" x=\"2\" y=\"28\" width=\"20\" height=\"3\" rx=\"1\"/>" +
          topSand + bottomSand + stream +
          "<path class=\"hg-glass\" d=\"M4 4 L20 4 L12.8 16 L20 28 L4 28 L11.2 16 Z\"/>" +
          "</svg>";
      }
      // A saved ready bubble, redrawn on page load (restoreChat).
      function restoreWarmupBubble(saved) {
        var parts = String(saved).split("|");
        showWarmupLevel(buildWarmupBubble(parts[0]), llmWarm.leds, true, parts.slice(1).join("|"));
      }
      // 2026-10-13 (operator: "chat bubbles must run in an orderly sequence so
      // Goverlord's orchestration is presented correctly"; "reporter lines ...
      // contained in between must appear before llm system bubble"): the
      // warm-up file only drives the warming LEDs. The bubble turns ready (or,
      // after a warm start with no warming, first appears, ready) at the feed's
      // "warmup_ready|<started_ms>|<ready text>" line - in the chat's order,
      // after every line before it.
      function warmupReadyLine(text) {
        var rest = String(text).slice("warmup_ready|".length);
        var cut = rest.indexOf("|");
        var ms = cut < 0 ? rest.trim() : rest.slice(0, cut).trim();
        var readyText = cut < 0 ? "" : rest.slice(cut + 1);
        var key = "warm|" + ms;
        if (!ms || isSeen(key)) { return; }
        if (warmupBubble && !warmupBubble.isConnected) { warmupBubble = null; }
        if (!warmupBubble || warmupBubble.dataset.started !== ms) {
          warmupBubble = buildWarmupBubble(ms);
        }
        showWarmupLevel(warmupBubble, llmWarm.leds, true, readyText);
        markSeen(key);
        recordBubble("Operator", ms + (readyText ? "|" + readyText : ""), "llm_ready");
        if (!openingDone) { warmupBeforeOpening = warmupBubble; }
        warmupBubble = null;
      }
      function warmupTextLine(text) {
        var rest = String(text).slice("warmup_text|".length);
        var cut = rest.indexOf("|");
        var ms = cut < 0 ? rest.trim() : rest.slice(0, cut).trim();
        var message = cut < 0 ? "" : rest.slice(cut + 1);
        skipGateCheck(message);
        if (!ms || isSeen("warm|" + ms)) { return false; }
        if (warmupBubble && !warmupBubble.isConnected) { warmupBubble = null; }
        if (!warmupBubble || warmupBubble.dataset.started !== ms) { warmupBubble = buildWarmupBubble(ms); }
        warmupBubble.dataset.text = message;
        var lit = warmupBubble.querySelectorAll(".llm-led.lit").length;
        showWarmupLevel(warmupBubble, Math.max(1, lit), false, "", message);
        // "Please wait." alone (the lines are next) isn't held.
        return message.trim() !== llmWarm.waitText.trim();
      }
      function pollWarmup() {
        if (warmupPolling) { return; }
        warmupPolling = true;
        fetch("/gopod_llm_warmup.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (state) {
            if (!state || !state.started_ms) { return; }
            // A restart clears the chat (clearBubbles), this bubble with it.
            if (warmupBubble && !warmupBubble.isConnected) { warmupBubble = null; }
            // Already finished and kept in the chat (or restored from it).
            if (isSeen("warm|" + state.started_ms)) { return; }
            var mine = warmupBubble && warmupBubble.dataset.started === String(state.started_ms);
            if (!state.active) {
              // Warm now; its ready line in the feed turns it ready.
              if (mine) { showWarmupLevel(warmupBubble, llmWarm.leds - 1, false, "", warmupBubble.dataset.text); }
              return;
            }
            if (!mine) { warmupBubble = buildWarmupBubble(state.started_ms); }
            var level = Math.max(0, Date.now() - state.started_ms) / (llmWarm.expectedSeconds * 1000);
            // 2026-10-17 (operator: "Goverlord awaits your return... + hourglass,
            // and waits, nothing plays"): while Wire-Pod waits for this page, the
            // bubble shows its awaits text; its own message comes back after.
            var shownText = state.awaits && state.text ? state.text : warmupBubble.dataset.text;
            showWarmupLevel(warmupBubble, Math.min(llmWarm.leds - 1, 1 + Math.floor(level * llmWarm.leds)), false, "", shownText);
          })
          .catch(function () { /* no warm-up file (older Wire-Pod) */ })
          .then(function () { warmupPolling = false; });
      }

      // 2026-10-02 (operator: "make a really big bingo ball number display for
      // B-1, I-9 & O-0"): a big ball, centred, coloured by its letter, for the
      // called number ("B-1"). 2026-10-04 (operator: "Please keep the bingo balls
      // displayed in the chat bubbles window"): it stays, is saved with the page's
      // chat and comes back after a refresh; only Clear Chat removes it. Never in
      // the Brobots' chat memory.
      // 2026-10-06 (operator: "think of it like a long interview video on youtube,
      // segmented with chapters"): a chapter title, in a system bubble
      // (interview_chapters in the settings file, gopod_interview_chapters.go).
      var chapterRole = "Chapter", chapterLabel = "Chapter: {title}";
      function showChapter(title) {
        if (!restoringChat) { recordBubble(chapterRole, String(title), "chapter"); }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator system-status chapter";
        bubble.appendChild(systemHead(chapterLabel.replace("{title}", String(title))));
        container.appendChild(bubble);
        if (restoringChat) { return; }
        revealWhenPlaced(bubble);
      }
      // 2026-10-06 (operator: "Yes, please build this way, with score bubble"):
      // the session's value-point score at "Thanks Brobots", in a Goverlord
      // system bubble (value_points, gopod_value_points.go).
      var scoreRole = "Score";
      // 2026-10-08 (operator: "Operator bubble: system message (... GOPOD
      // Brain) with \"Draft run completed.\""): sign_off.system_role's lines
      // show in the same system bubble.
      var systemRole = "Goverlord";
      // 2026-10-17 (operator: "terminal-results are sent back to chat bubbles
      // window * In terminal-results chat bubble, make terminal-bubble-text
      // font-size 50% of current bubble-text font-size"): system-set's answers,
      // in a system bubble with small text (not kept in the saved chat).
      var terminalRole = "Terminal";
      // 2026-10-18 (operator: "CHALK (display) now is to take ownership of:
      // search-results box (same smaller size, no audio) * system-status and the
      // other system bubbles"): these bubbles carry CHALK's header and board
      // (managers.display_host); with none, Goverlord's as before.
      var displayHost = "";
      function displayHead(bubble) { return hostHeadFor(bubble, displayHost); }
      // A manager's header and colour on a bubble (CHALK's display, PLAYHEAD's search).
      function hostHeadFor(bubble, who) {
        if (!who) { return systemHead(systemLabel); }
        bubble.classList.add("host-" + who.toLowerCase().replace(/[^a-z0-9]+/g, "-"));
        return reporterHead(hostHeader(who));
      }
      // GOLDEN raw bubble template (2026-10-21, operator: "before bubble types
      // there is the raw bubble scaffolding * build up bubble types from the raw
      // basic bubble boiler template * whatever processes/rules apply to a bubble
      // must complete before the next bubble"; "every bubble ... must fully render
      // & pre-load hidden * somewhere so as not to affect live scrolling"). Every
      // bubble type is built on these, one type at a time:
      //   rawBubble(spec)        the bubble, whole: kind, classes, header, lines
      //   bubbleTask(b, promise) a job it must finish before it shows
      //   placeBubble(b, then)   built in the hidden staging area (outside the
      //                          chat - no scroll), its tasks (and its pictures
      //                          loading) done, then into the chat and shown;
      //                          "then" - the next event - only after that.
      // Types on it so far: system/terminal answers; reporter and manager lines;
      // Goverlord's lines; Brobot replies. Exceptions (their own rules): the chat window, the
      // Operator bubble (the one input type), the system window (menu).
      var bubbleStaging = null, BUBBLE_TASK_MAX_MS = 10000;
      function bubbleStage() {
        if (!bubbleStaging || !bubbleStaging.isConnected) {
          bubbleStaging = document.createElement("div");
          bubbleStaging.className = "bubble-staging";
          bubbleStaging.setAttribute("aria-hidden", "true");
          document.body.appendChild(bubbleStaging);
        }
        return bubbleStaging;
      }
      function rawBubble(spec) {
        var b = document.createElement("div");
        b.className = "brobot-bubble " + (spec.classes || "");
        if (spec.kind) { b.dataset.kind = spec.kind; }
        if (spec.head) { b.appendChild(spec.head); }
        (spec.lines || (spec.text != null ? [spec.text] : [])).forEach(function (t) {
          var line = document.createElement("span");
          line.className = "brobot-bubble-text" + (spec.lineClass ? " " + spec.lineClass : "");
          line.textContent = String(t);
          b.appendChild(line);
        });
        b._tasks = [];
        return b;
      }
      function bubbleTask(b, promise) { b._tasks.push(promise); return promise; }
      function placeBubble(b, then, onPlaced) {
        var finished = false;
        function finish() {
          if (finished) { return; }
          finished = true;
          container.appendChild(b);
          if (onPlaced) { onPlaced(b); }
          if (!restoringChat) { revealWhenPlaced(b); }
          if (then) { then(); }
        }
        if (restoringChat) { finish(); return; }
        var stage = bubbleStage();
        stage.style.width = container.clientWidth + "px";
        stage.appendChild(b);
        Array.prototype.forEach.call(b.querySelectorAll("img"), function (img) {
          if (!img.complete) {
            bubbleTask(b, new Promise(function (ok) { img.addEventListener("load", ok, { once: true }); img.addEventListener("error", ok, { once: true }); }));
          }
        });
        var tasks = (b._tasks || []).map(function (t) { return Promise.resolve(t).catch(function () {}); });
        Promise.race([Promise.all(tasks), new Promise(function (ok) { setTimeout(ok, BUBBLE_TASK_MAX_MS); })]).then(finish);
      }
      function showTerminal(text, then) {
        // While the system menu is open, its answers go in its results (systemMenuResult).
        if (systemMenuResult(String(text), "answer")) { if (then) { then(); } return; }
        // 2026-10-21: a system answer is Goverlord's (the Captain's).
        placeBubble(rawBubble({ kind: "terminal", classes: "operator system-status score terminal",
          head: systemHead(systemLabel), text: text, lineClass: "terminal-text" }), then);
      }
      // 2026-10-17 (operator: "Don't speak or highlight search results * put a
      // centered rounded-rectangle bubble around the search results * search
      // results font @ 50% size * do a fast character by character reveal * then
      // brobot responses"): the results of a "search ..." line, never spoken,
      // typed in a character at a time (system_search.reveal_ms_per_char, at
      // most reveal_max_seconds).
      var searchSettings = { results_role: "Search Results" };
      function showSearchResults(text, to) {
        var bubble = document.createElement("div");
        // 2026-10-21: a "search ..." to a Brobot - PLAYHEAD's part at the top of his bubble.
        var nestTo = /^Brobot [12]$/.test(String(to || "")) && !restoringChat ? to : "";
        bubble.className = "brobot-bubble search-results";
        // 2026-10-19 (operator: "system-search = location = PLAYHEAD, not CHALK"):
        // system_search.results_host draws them (else the display host).
        var resultsHost = searchSettings.results_host || displayHost;
        if (resultsHost) { bubble.appendChild(hostHeadFor(bubble, resultsHost)); }
        var line = document.createElement("span");
        line.className = "brobot-bubble-text search-results-text";
        // Built whole (golden: placed at its full size), every character
        // hidden until typed in.
        var chars = [];
        Array.from(String(text)).forEach(function (c) {
          if (c === "\n") { line.appendChild(document.createElement("br")); return; }
          var ch = document.createElement("span");
          ch.className = "search-char";
          ch.textContent = c;
          line.appendChild(ch);
          chars.push(ch);
        });
        bubble.appendChild(line);
        var inMenu = !restoringChat && !nestTo && systemMenuResults();
        if (nestTo) {
          generatingGone();
          var shell = nestShell(nestTo);
          bubble.classList.add("nested-host");
          bubble.dataset.host = resultsHost;
          var reply = shell.querySelector(":scope > .brobot-bubble-text");
          if (reply) { shell.insertBefore(bubble, reply); } else { shell.appendChild(bubble); }
        }
        else if (inMenu) { inMenu.appendChild(bubble); inMenu.scrollTop = inMenu.scrollHeight; }
        else { container.appendChild(bubble); }
        if (restoringChat) { bubble.classList.add("typed"); return; }
        if (!inMenu && !nestTo) { revealWhenPlaced(bubble); }
        var per = searchSettings.reveal_ms_per_char || 6;
        var most = (searchSettings.reveal_max_seconds || 2) * 1000;
        if (per * chars.length > most) { per = most / Math.max(1, chars.length); }
        var shown = 0, started = null;
        function step(now) {
          if (started === null) { started = now; }
          var want = Math.min(chars.length, Math.floor((now - started) / per) + 1);
          while (shown < want) { chars[shown++].classList.add("on"); }
          if (shown < chars.length) { requestAnimationFrame(step); } else { bubble.classList.add("typed"); }
        }
        // Typed in once the bubble shows.
        whenShown(bubble, function () { requestAnimationFrame(step); });
      }
      // 2026-10-12 (operator: Reporter 0, "Reporter 0, Goverlord"): Goverlord's
      // own lines are headed with its reporter label (reporter_voices.labels).
      function showScore(text, label, then) {
        if (!restoringChat) { recordBubble(scoreRole, String(text), "score"); }
        // On the raw bubble template (built and loaded hidden, then shown).
        var bubble = rawBubble({ kind: "system", classes: "operator system-status score", head: systemHead(label || systemLabel) });
        var line = document.createElement("span");
        line.className = "brobot-bubble-text";
        bubble.appendChild(line);
        voiceLitLine(bubble, line, text);
        placeBubble(bubble, then);
      }
      // GOLDEN, one slot for a voice being made (2026-10-18, operator: "where a
      // voice is being generated displays the generating + hourglass 'message'
      // until the voice is generated & the 'message' is updated * my worry is
      // tiny adjustments like these aren't properly implemented as a
      // non-thinking sub-processes that just work everywhere"): Wire-Pod's
      // generating_marker+"on" opens a slot (generating_text and the
      // hourglass); the next voiced line - Goverlord's or a reporter's - fills
      // that same bubble with its own header and words, in place (no swap, no
      // jump). The Operator's own bubble does the same in itself ("op").
      // GOLDEN highlight for voices the page plays (2026-10-18, operator:
      // "Highlighting mechanism is awesome * can you apply this independent
      // working mechanism to system and reporter voices?"): a reporter's or
      // Goverlord's line is one piece, its words hidden until said, revealed
      // word by word across its voice's own length (the Brobots' word reveal,
      // scaled like an SRT line), then 85%. Its voice is the next one played
      // after it that isn't a sting (playVoiceThen); with no voice, its words
      // show anyway once it has been up voiceLitWaitMs.
      var voiceLitNext = null, voiceLitWaitMs = 15000;
      // GOLDEN global message variable (2026-10-18, operator: "decouple generating
      // + hourglass into its own 'global message variable' available to all
      // bubbles * [message0]/[message1] * message0: "Generating..." + hourglass *
      // message1: generated playable voice wav file + [transcription]"): every
      // voiced line - Goverlord's, a reporter's, a manager's, the Operator's own
      // (slotWaiting) - waits in its own bubble with generatingText and the
      // hourglass (generatingGlass), and turns into its words as its voice plays.
      var lineWaits = false, lineVoice = "", voiceWaits = [], voiceWaitFor = {};
      // THE one Generating template (2026-10-21, operator: "A single robust
      // bulletproof clean boring deterministic 'Generating...' mechanism template"):
      // voiceWaitOn puts generatingText + the hourglass just before an element
      // (a voiced line's words, or the Operator's own words), keyed by the voice
      // file it waits for; voiceWaitEnd takes it away and shows the words - when
      // that voice starts (voiceLitStart), or for all on "off" (voiceWaitsGone).
      function voiceWaitOn(bubble, el, file) {
        var wait = document.createElement("span");
        wait.className = "voice-wait";
        // 2026-10-22 (operator: "generating + hourglass, not hourglass + generating").
        wait.appendChild(document.createTextNode(generatingText));
        wait.appendChild(generatingGlass(bubble));
        el.parentNode.insertBefore(wait, el);
        el._wait = wait;
        if (file) { voiceWaitFor[file] = el; }
        voiceWaits.push(el);
      }
      function voiceWaitEnd(say) {
        if (say && say._wait) { say._wait.remove(); say._wait = null; }
        voiceWaits = voiceWaits.filter(function (x) { return x !== say; });
      }
      // "off": the voice failed - the waiting lines show their words.
      function voiceWaitsGone() {
        voiceWaits.slice().forEach(function (say) { voiceWaitEnd(say); say.classList.add("spoken"); });
      }
      function voiceLitLine(bubble, line, text) {
        line.textContent = "";
        var say = document.createElement("span");
        say.className = "brobot-say";
        // 2026-10-18 (operator, CHALK's weather line: "no flag"): flags drawn
        // as their pictures, like a Brobot's bubble.
        say.innerHTML = flagPictures(escapeHtml(String(text)));
        line.appendChild(say);
        var waits = lineWaits;
        lineWaits = false;
        if (lineVoice) { voiceWaitFor[lineVoice] = say; lineVoice = ""; }
        if (restoringChat) { say.classList.add("spoken"); return; }
        // Skipping the reporters' lines: shown whole, not waited on.
        if (skippingReporters) { say.classList.add("spoken"); bubble.classList.add("voice-lit"); return; }
        bubble.classList.add("voice-lit");
        voiceLitNext = say;
        // message0 until its voice plays: "Generating..." and the hourglass.
        if (waits) { voiceWaitOn(bubble, say, typeof waits === "string" ? waits : ""); }
        var shownAt = 0;
        (function showAnyway() {
          if (!bubble.isConnected || say.classList.contains("speaking") || say.classList.contains("spoken")) { return; }
          if (say._wait && say._wait.isConnected) { setTimeout(showAnyway, 500); return; }
          if (bubble.style.visibility !== "hidden" && bubble.dataset.staged !== "1") {
            if (!shownAt) { shownAt = Date.now(); }
            if (Date.now() - shownAt >= voiceLitWaitMs) {
              say.classList.add("spoken");
              if (voiceLitNext === say) { voiceLitNext = null; }
              return;
            }
          }
          setTimeout(showAnyway, 500);
        })();
      }
      // Its voice starts: the words revealed across seconds, then said.
      function voiceLitStart(seconds, words, src) {
        // The line waiting for this very voice file, else the line shown last.
        var path = String(src || "").replace(/#.*$/, "").replace(/^https?:\/\/[^/]+/, "");
        var own = path && voiceWaitFor[path];
        // A voice-only line not on the page when its voice was named: tied now.
        if (!own && path && voiceForText[path]) { own = voiceLineTie(path, voiceForText[path], false); }
        if (path) { delete voiceForText[path]; }
        if (own) { delete voiceWaitFor[path]; }
        var say = own && own.isConnected ? own : (voiceLitNext && voiceLitNext.isConnected ? voiceLitNext : null);
        if (!own || say === voiceLitNext) { voiceLitNext = null; }
        voiceWaitEnd(say);
        if (!say || say.classList.contains("spoken")) { return null; }
        if (!(seconds > 0)) { seconds = Math.max(1, sayWords(say.textContent).length * (wordReveal.seconds_per_word || 0.38)); }
        say.dataset.at = "0";
        say.dataset.end = String(seconds);
        say._spokenWords = words || null;
        lightSpan(say);
        // 2026-10-23: a line said last in the same voice, shown apart (the egg's
        // "MouseOver to Zoom In") - lit word by word at its own words' times.
        var tail = say._zoom;
        if (tail && tail.isConnected) {
          var n = sayWords(tail.textContent).length;
          tail._spokenWords = words && words.length >= n ? words.slice(words.length - n) : null;
          tail.dataset.at = "0";
          tail.dataset.end = String(seconds);
          lightSpan(tail);
        }
        return function () {
          say.classList.remove("speaking");
          say.classList.add("spoken");
          var saidIn = say.closest(".brobot-bubble");
          if (saidIn) { (saidIn._spoken = saidIn._spoken || {})[voiceLineWords(say.textContent)] = true; }
          if (tail) { tail.classList.remove("speaking"); tail.classList.add("spoken"); }
        };
      }
      // 2026-10-06 (operator: "Pictures in the chat"): a topic's pictures, one
      // row in a system bubble; the line is their paths, "|" between them.
      var pictureRole = "Picture", picturesInQuestion = [];
      var eggZoomText = "MouseOver to Zoom In";
      // A Brobot's pop-up wrapper waiting for his reply (showPictures).
      var popupWraps = {};
      // A pop-up was found for the question being answered (reset by the next
      // operator line): every Brobot's answer to it wears the marquee.
      var popupExchange = false;
      // 2026-10-13 (operator: "Make these live Pop Ups appear like a winner's
      // chat bubble * system message from Goverlord, congratulations finding
      // this pop-up"; "use system chat bubbles"; "I still want the zoom
      // feature"): a live topic's pictures ("popup:" on the line) in a Goverlord
      // system bubble, popup_text over them, zoom as ever.
      var popupText = "Congratulations, Master CrushN8r! You found a pop-up.";
      var mathPopupTitle = "Explain the Math";
      function showPictures(list) {
        voiceLitNext = null; // the next voice is this egg's, not an earlier line's
        if (!restoringChat) { recordBubble(pictureRole, String(list), "picture"); }
        var popup = String(list).indexOf("popup:") === 0;
        if (popup) { list = String(list).slice("popup:".length); }
        // 2026-10-14 (operator: "combine a brobot's response bubble with a pop-up
        // wrapper ... like more of a surprise prize"): a pop-up that names its
        // Brobot ("Brobot 2::<pictures>") is his next reply bubble - marquee,
        // "Brobot 2:", Goverlord's congratulations with the prize, then his
        // reply joins it below (appendBubble).
        var popWho = "";
        var whoMatch = popup ? /^(Brobot [12])::/.exec(String(list)) : null;
        if (whoMatch) { popWho = whoMatch[1]; list = String(list).slice(whoMatch[0].length); }
        // 2026-10-18 (operator: "'weather' can still be an easter egg, but with a
        // chalkboard color background"): a host's egg ("CHALK::") - its own board,
        // not a Brobot's wrapper.
        var hostEgg = "";
        var hostMatch = popup && !whoMatch ? /^([A-Za-z][A-Za-z0-9 ]*)::/.exec(String(list)) : null;
        if (hostMatch && hostMatch[1] !== "msg=" && hostNames[hostMatch[1]]) { hostEgg = hostMatch[1]; list = String(list).slice(hostMatch[0].length); }
        // 2026-10-16 (operator: '"Congratulations, Master CrushN8r! You found the
        // easter egg, [easter egg name/trigger]"'): the egg's own congratulations.
        var eggText = popupText;
        var msgMatch = popup ? /^msg=([^]*?)::/.exec(String(list)) : null;
        if (msgMatch) { eggText = msgMatch[1]; list = String(list).slice(msgMatch[0].length); }
        // 2026-10-17 (operator: "Congratulations! / Easter Egg '[Title-of-Easter]'
        // found. / [images] / MouseOver to Zoom In * MouseOver to Zoom In goes
        // below the images"): the zoom line goes under the pictures.
        var zoomLine = "";
        if (eggZoomText && eggText.trim().slice(-eggZoomText.length) === eggZoomText) {
          zoomLine = eggZoomText;
          eggText = eggText.trim().slice(0, -eggZoomText.length).trim();
        }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator system-status pictures" + (popup ? " popup-winner" : "") + (hostEgg ? " host-" + hostEgg.toLowerCase() : "");
        var holder = bubble;
        if (popup && popWho) {
          bubble.className = "brobot-bubble " + (popWho === "Brobot 1" ? "brobot-1" : "brobot-2") + " popup-winner popup-wrap";
          // GOLDEN pop-up rule (2026-10-15, operator: 'Correct: "no more
          // top-most header" only'): no "Brobot N:" above Goverlord's block.
          popupExchange = true;
          holder = document.createElement("div");
          holder.className = "popup-system";
          holder.appendChild(systemHead(systemLabel));
          var prize = document.createElement("span");
          prize.className = "popup-congrats";
          // 2026-10-16: "MouseOver to Zoom View" on its own line.
          prize.style.whiteSpace = "pre-line";
          holder.appendChild(prize);
          // 2026-10-18 (step 5, eggs on the bubble machine): Goverlord's
          // congratulations revealed as it's said, like every voice (voiceLitLine).
          // 2026-10-23 (operator: "no generating + hourglass while voice is
          // generating * I prefer to see a single generating + hourglass for both
          // lines"): the congratulations wait as one line (the Generating template).
          lineWaits = true;
          voiceLitLine(bubble, prize, eggText);
          bubble.appendChild(holder);
          // 2026-10-15 (operator: "Whichever Brobot triggers however many easter
          // eggs in a single input * each is displayed like this newest granny
          // bubble (brobot bubble + marquee containing congrats message) * make
          // all easter eggs that way"): every egg is its own bubble; the reply
          // comes after in his own marquee bubble (popupExchange).
          if (!popupWraps[popWho]) { popupWraps[popWho] = bubble; }
        } else if (popup) {
          bubble.appendChild(systemHead(systemLabel));
          var congrats = document.createElement("span");
          congrats.className = "brobot-bubble-text";
          congrats.style.whiteSpace = "pre-line";
          bubble.appendChild(congrats);
          lineWaits = true;
          voiceLitLine(bubble, congrats, eggText);
        }
        var row = document.createElement("div");
        row.className = "topic-pictures";
        // 2026-10-12 (operator: "move granny photo into reporter bubble, one line
        // down, still centered + mouse-over-zoom"): topic_pictures.in_question
        // pictures go inside the question bubble just above, under its words.
        var files = String(list).split("|").filter(function (src) { return src.trim(); });
        var question = container.lastElementChild;
        if (warmupBubble && question === warmupBubble) { question = question.previousElementSibling; }
        var inQuestion = files.length && files.every(function (src) {
          return picturesInQuestion.indexOf(src.trim().split("/").pop()) >= 0;
        }) && !popup && question && question.classList.contains("operator") && question.dataset.own === "1";
        var weatherSound = "";
        String(list).split("|").forEach(function (src) {
          if (!src.trim()) { return; }
          // 2026-10-16: a card prize ("card:<words>", the roast battle's) - its
          // words in the egg, bold, where a picture would be.
          if (src.trim().indexOf("card:") === 0) {
            var cardBox = document.createElement("div");
            cardBox.className = "brobot-bubble-text";
            var cardWords = document.createElement("strong");
            cardWords.textContent = src.trim().slice("card:".length);
            cardWords.style.whiteSpace = "pre-line"; // its own lines (2026-10-16)
            cardBox.appendChild(cardWords);
            holder.appendChild(cardBox);
            return;
          }
          // 2026-10-14: the "Explain the Math" pop-up ("math:<statement>"):
          // math.popup_title over the challenge; both Brobots' answers join it.
          if (src.trim().indexOf("math:") === 0) {
            var mathBox = document.createElement("div");
            mathBox.className = "brobot-bubble-text";
            var mathTitle = document.createElement("strong");
            mathTitle.textContent = mathPopupTitle;
            mathBox.appendChild(mathTitle);
            mathBox.appendChild(document.createElement("br"));
            mathBox.appendChild(document.createTextNode(src.trim().slice("math:".length)));
            holder.appendChild(mathBox);
            return;
          }
          // 2026-10-14: the weather pop-up ("weather:<words>@<hour>"):
          // Vector's weather face for it, and its sound as the pop-up shows.
          if (src.trim().indexOf("weather:") === 0) {
            var wx = /^weather:(.*)@(\d+)$/.exec(src.trim());
            var wface = weatherCard && wx ? weatherFaceFor(wx[1].toLowerCase(), parseInt(wx[2], 10)) : null;
            if (wface) {
              row.appendChild(weatherScreen(wface.name, function () { return row; }));
              weatherSound = wface.sound;
            }
            return;
          }
          var img = document.createElement("img");
          img.src = src.trim();
          img.alt = "";
          // 2026-10-09: grows to the bubble's width under the mouse (one at a
          // time; it shrinks when the mouse leaves the pictures, so the grown
          // picture moving under the mouse doesn't flicker); a tap toggles it
          // (gopod_chat_bubbles.css .zoomed).
          magnify(row, img);
          row.appendChild(img);
        });
        if (inQuestion) {
          row.classList.add("in-question");
          question.appendChild(row);
          if (restoringChat) { return; }
          // 2026-10-12 (operator: "the bubble stopped at the end of text and the
          // photo was cut off from view"): the question is placed again, whole,
          // once its picture has loaded.
          if (question.style.visibility === "hidden") { placeAgain = true; return; }
          whenBubblesReady([question], function () { scrollChatTo(bubbleTarget(question)); });
          return;
        }
        if (row.children.length) { holder.appendChild(row); }
        if (zoomLine && row.children.length) {
          var zoomEl = document.createElement("span");
          zoomEl.className = popup && popWho ? "popup-congrats popup-zoom-line" : "brobot-bubble-text popup-zoom-line";
          // 2026-10-23 (operator: "same with highlighting on, mouseover to zoom in
          // message, no highlighting?"): Goverlord says it last, in the same voice
          // as the congratulations - so it lights with that voice too: lit when the
          // congratulations' last word is said, shown said when the voice ends.
          var zoomSay = document.createElement("span");
          zoomSay.className = "brobot-say";
          zoomSay.textContent = zoomLine;
          zoomEl.appendChild(zoomSay);
          holder.appendChild(zoomEl);
          // Lit word by word from the congratulations' own voice (its last words,
          // voiceLitStart), like every voiced line.
          var litHost = popup && popWho ? prize : (popup ? congrats : null);
          var congratsSay = litHost ? litHost.querySelector(".brobot-say") : null;
          if (!congratsSay || restoringChat || skippingReporters || congratsSay.classList.contains("spoken")) {
            zoomSay.classList.add("spoken");
          } else {
            congratsSay._zoom = zoomSay;
          }
        }
        container.appendChild(bubble);
        if (popup) { marqueeAttach(bubble); }
        if (weatherSound) { weatherSoundOnShow(bubble, weatherSound); }
        if (restoringChat) { return; }
        revealWhenPlaced(bubble);
      }
      // 2026-10-13 (operator: "For marquee lights, please grab golden guts from
      // this unrelated response and apply to live pop ups * how to handle bubble
      // resizing?"): real bulbs placed evenly along the pop-up's rounded border
      // (straight sides and corner arcs), cycling Chase, Groups, Pairs and
      // Twinkle every MARQUEE_PATTERN_MS. The bubble's size and corner radius are
      // read live, and a ResizeObserver lays the bulbs out again whenever the
      // bubble changes size (text wrapping, a picture loading, the window), the
      // count following the border's length so the spacing stays even.
      var MARQUEE_SPACING_PX = 22, MARQUEE_INSET_PX = 9, MARQUEE_PATTERN_MS = 5000;
      var marquees = [], marqueeFrame = null, marqueePattern = 0, marqueeSwitch = 0;
      var marqueeSizes = typeof ResizeObserver === "function" ? new ResizeObserver(function (entries) {
        entries.forEach(function (entry) {
          marquees.forEach(function (m) { if (m.bubble === entry.target) { marqueeLayout(m); } });
        });
      }) : null;
      var marqueePatterns = [
        function chase(t, i, n) { var pos = (t * 8) % n; return Math.abs(i - pos) < 1.2 || Math.abs(i - pos - n) < 1.2; },
        function groups(t, i) { return Math.floor((t * 6 + i) / 3) % 2 === 0; },
        function pairs(t, i, n) { var pos = (t * 10) % n; return Math.min(Math.abs(i - pos), Math.abs(i - pos - n), Math.abs(i - pos + n)) < 2.5; },
        function twinkle(t, i) { return Math.sin(t * 3 + i * 1.7) * 0.5 + 0.5 > 0.65; }
      ];
      function marqueeAttach(bubble) {
        var layer = document.createElement("div");
        layer.className = "marquee-lights";
        layer.setAttribute("aria-hidden", "true");
        bubble.appendChild(layer);
        var m = { bubble: bubble, layer: layer, lights: [] };
        marquees.push(m);
        if (marqueeSizes) { marqueeSizes.observe(bubble); }
        marqueeLayout(m);
        if (marqueeFrame === null) { marqueeFrame = requestAnimationFrame(marqueeTick); }
      }
      // The point d along a rounded rectangle's border (w x h, corner radius r),
      // clockwise from the top side's left end.
      function marqueePoint(d, w, h, r, per) {
        d = ((d % per) + per) % per;
        var q = Math.PI * r / 2, top = w - 2 * r, side = h - 2 * r, a;
        if (d < top) { return { x: r + d, y: 0 }; } d -= top;
        if (d < q) { a = -Math.PI / 2 + d / Math.max(r, 0.001); return { x: w - r + Math.cos(a) * r, y: r + Math.sin(a) * r }; } d -= q;
        if (d < side) { return { x: w, y: r + d }; } d -= side;
        if (d < q) { a = d / Math.max(r, 0.001); return { x: w - r + Math.cos(a) * r, y: h - r + Math.sin(a) * r }; } d -= q;
        if (d < top) { return { x: w - r - d, y: h }; } d -= top;
        if (d < q) { a = Math.PI / 2 + d / Math.max(r, 0.001); return { x: r + Math.cos(a) * r, y: h - r + Math.sin(a) * r }; } d -= q;
        if (d < side) { return { x: 0, y: h - r - d }; } d -= side;
        a = Math.PI + d / Math.max(r, 0.001);
        return { x: r + Math.cos(a) * r, y: r + Math.sin(a) * r };
      }
      function marqueeLayout(m) {
        var W = m.bubble.clientWidth, H = m.bubble.clientHeight;
        if (!W || !H) { return; }
        var w = W - 2 * MARQUEE_INSET_PX, h = H - 2 * MARQUEE_INSET_PX;
        var radius = parseFloat(getComputedStyle(m.bubble).borderTopLeftRadius) || 0;
        var r = Math.max(0, Math.min(radius - MARQUEE_INSET_PX / 2, w / 2, h / 2));
        var per = 2 * (w - 2 * r) + 2 * (h - 2 * r) + 2 * Math.PI * r;
        var n = Math.max(8, Math.round(per / MARQUEE_SPACING_PX));
        while (m.lights.length < n) {
          var el = document.createElement("span");
          el.className = "marquee-light";
          m.layer.appendChild(el);
          m.lights.push(el);
        }
        while (m.lights.length > n) { m.layer.removeChild(m.lights.pop()); }
        for (var i = 0; i < n; i++) {
          var pt = marqueePoint(per * i / n, w, h, r, per);
          m.lights[i].style.left = (pt.x + MARQUEE_INSET_PX) + "px";
          m.lights[i].style.top = (pt.y + MARQUEE_INSET_PX) + "px";
        }
      }
      function marqueeTick(now) {
        marquees = marquees.filter(function (m) {
          if (m.bubble.isConnected) { return true; }
          if (marqueeSizes) { marqueeSizes.unobserve(m.bubble); }
          return false;
        });
        if (!marquees.length) { marqueeFrame = null; return; }
        // 2026-10-13 (operator: "marquee lights look good, not they're not
        // animating"): they always animate, the PC's reduce-motion setting too.
        var still = false;
        if (now - marqueeSwitch > MARQUEE_PATTERN_MS) {
          marqueePattern = (marqueePattern + 1) % marqueePatterns.length;
          marqueeSwitch = now;
        }
        var t = now / 1000, fn = marqueePatterns[marqueePattern];
        marquees.forEach(function (m) {
          var n = m.lights.length;
          for (var i = 0; i < n; i++) {
            // Reduced motion: every other bulb lit, still.
            var on = still ? i % 2 === 0 : fn(t, i, n);
            m.lights[i].classList.toggle("on", on);
            m.lights[i].classList.toggle("on-warm", on && !still && marqueePattern === 1);
            m.lights[i].classList.toggle("on-cool", on && !still && marqueePattern === 3);
          }
        });
        marqueeFrame = requestAnimationFrame(marqueeTick);
      }
      // 2026-10-14 (operator: "QR codes for crushn8r.ca & newsletter * add them
      // in whenever crushn8r.ca or newsletter is in the chat bubble,
      // aligned-top-right"): every bubble, however it was made, is checked as it
      // joins the chat (qr_codes, gopod_brobots_settings.json). The pictures go
      // in while the bubble is still hidden, so it shows with them.
      var qrCodes = [];
      function addQrCodes(bubble) {
        if (!qrCodes.length || !bubble.classList || !bubble.classList.contains("brobot-bubble") || bubble.querySelector("img.gopod-qr")) { return; }
        var textEl = bubble.querySelector(".brobot-bubble-text");
        if (!textEl) { return; }
        var said = String(textEl.textContent || "").toLowerCase();
        var found = qrCodes.filter(function (code) {
          var text = said;
          (code.not_words || []).forEach(function (w) { text = text.split(String(w).toLowerCase()).join(" "); });
          return (code.words || []).some(function (w) { return text.indexOf(String(w).toLowerCase()) >= 0; });
        });
        // 2026-10-18 (operator: "QR codes handled by CHALK (system stuff)"): the
        // QR codes (and the call to action's, qr_codes.second) go in CHALK's own
        // board bubble, after the Brobot's, not in his corner.
        if (displayHost) { chalkQrs(bubble, found); return; }
        if (!found.length) { return; }
        var box = cornerBox(bubble, textEl, found.length * 70);
        found.forEach(function (code) {
          var img = document.createElement("img");
          img.className = "gopod-qr";
          img.src = code.image;
          img.alt = code.title || code.url || "";
          img.title = code.title || code.url || "";
          magnify(box, img);
          box.appendChild(img);
        });
        addSecondQr(bubble);
      }
      // 2026-10-17 (operator: "2 QR CODES * QR1 is in the top-right corner * QR2
      // is added after the call to action, invisible, revealed as highlighted
      // last"; "Wanna know more? Get CREW'd by subscribing to our newsletter!
      // [QR2]"): the call to action's own QR code goes right after its words,
      // in its piece, the last step of its word reveal (qr_codes.second).
      // CHALK's QR bubble: each code with its title, the magnifier on each.
      function chalkQrs(bubble, found) {
        var codes = found.slice();
        chalkCta(bubble);
        if (!codes.length) { return; }
        // In a Brobot's bubble: CHALK's part at its bottom, after his answer.
        var inBrobot = (bubble.classList.contains("brobot-1") || bubble.classList.contains("brobot-2")) && !bubble.classList.contains("nested-host");
        var board = document.createElement("div");
        board.className = "brobot-bubble operator system-status chalk-qr" + (inBrobot ? " nested-host" : "");
        board.appendChild(displayHead(board));
        var row = document.createElement("div");
        row.className = "topic-pictures";
        codes.forEach(function (code) {
          var fig = document.createElement("figure");
          fig.className = "chalk-qr-code";
          var img = document.createElement("img");
          img.className = "gopod-qr";
          img.src = code.image;
          img.alt = img.title = code.title || code.url || "";
          magnify(row, img);
          fig.appendChild(img);
          var cap = document.createElement("figcaption");
          cap.textContent = code.title || code.url || "";
          fig.appendChild(cap);
          row.appendChild(fig);
        });
        board.appendChild(row);
        if (inBrobot) { bubble.appendChild(board); return; }
        container.appendChild(board);
        if (!restoringChat) { revealWhenPlaced(board); }
      }
      // 2026-10-18 (operator: "please put: * CHALK [QR code] Newsletter:
      // crushn8r.ca/subscribe * before wanna know more CTA"): CHALK's part, with
      // the newsletter QR and its title, just before the call to action's piece.
      function chalkCta(bubble) {
        if (!secondQr || !secondQr.after || bubble.querySelector(".chalk-cta")) { return; }
        var after = String(secondQr.after).toLowerCase(), hit = null;
        var says = ownSays(bubble);
        says.forEach(function (say) {
          if (!hit && String(say.textContent || "").toLowerCase().indexOf(after) >= 0) { hit = say; }
        });
        // 2026-10-20 (live: "Join the CRUSHN8R CREW'd newsletter at crushn8r.ca!"
        // opened an answer, in the model's own words): any newsletter line.
        if (!hit) { says.forEach(function (say) { if (!hit && /newsletter/i.test(say.textContent || "")) { hit = say; } }); }
        if (!hit) { return; }
        var ctaEnd = hit;
        // The call to action starts at "Wanna know more?" - the piece before, when it's that.
        var prev = hit.previousElementSibling;
        if (prev && prev.classList.contains("brobot-say") && /know more/i.test(prev.textContent || "") && !/know more/i.test(hit.textContent || "")) { hit = prev; }
        var box = document.createElement("div");
        box.className = "brobot-bubble operator system-status chalk-qr nested-host chalk-cta";
        box.appendChild(displayHead(box));
        var row = document.createElement("div");
        row.className = "topic-pictures";
        var fig = document.createElement("figure");
        fig.className = "chalk-qr-code";
        var img = document.createElement("img");
        img.className = "gopod-qr";
        img.src = secondQr.image;
        img.alt = img.title = secondQr.title || "";
        magnify(row, img);
        fig.appendChild(img);
        var cap = document.createElement("figcaption");
        cap.textContent = secondQr.title || "";
        fig.appendChild(cap);
        row.appendChild(fig);
        box.appendChild(row);
        // 2026-10-20 (operator: "upon opening with newsletter, add the QRs bubble
        // right after ... upon closing with newsletter, insert the QRs bubble
        // just before ... as if manager bubbles prefer to stay away (phobic) from
        // top/bottom of brobots response window"): on the inner side of the
        // newsletter line - after it when it opens the answer, before it otherwise.
        if (hit === says[0] && ctaEnd !== says[says.length - 1]) { ctaEnd.parentNode.insertBefore(box, ctaEnd.nextSibling); }
        else { hit.parentNode.insertBefore(box, hit); }
      }
      var secondQr = null;
      function addSecondQr(bubble) {
        if (displayHost) { return; } // CHALK shows it (chalkQrs)
        if (!secondQr || !secondQr.after || bubble.querySelector("img.gopod-qr2")) { return; }
        var after = String(secondQr.after).toLowerCase(), hit = null;
        Array.prototype.forEach.call(bubble.querySelectorAll(".brobot-say"), function (say) {
          if (String(say.textContent || "").toLowerCase().indexOf(after) >= 0) { hit = say; }
        });
        if (!hit) { return; }
        var img = document.createElement("img");
        img.className = "gopod-qr gopod-qr2";
        img.src = secondQr.image;
        img.alt = img.title = secondQr.title || "";
        magnify(hit, img);
        hit.appendChild(document.createTextNode(" "));
        hit.appendChild(img);
      }
      // 2026-10-17 (operator: "in a tall response bubble, QR code gets scrolled
      // up & out of view * RULE: upon the bottom of QR1 passing the top of the
      // chat window, the QR code jumps to reposition at the bottom-right,
      // respecting the same inner-width changes for marquee lights"): a bubble
      // whose top-right QR codes are scrolled above the chat's top has them at
      // its bottom-right instead, back at the top once its top shows again.
      function placeCornerQrs() {
        if (!container) { return; }
        var top = container.getBoundingClientRect().top;
        Array.prototype.forEach.call(container.querySelectorAll(".brobot-bubble.has-corner"), function (bubble) {
          var box = bubble.querySelector(".gopod-corner");
          if (!box || !box.querySelector("img.gopod-qr")) { return; }
          var b = bubble.getBoundingClientRect();
          var high = box.offsetTop + box.offsetHeight;
          if (bubble.classList.contains("corner-low")) { high = 10 + box.offsetHeight; }
          var gone = b.top + high < top && b.bottom > top;
          bubble.classList.toggle("corner-low", gone);
        });
      }
      if (container) {
        container.addEventListener("scroll", placeCornerQrs, { passive: true });
        window.addEventListener("scroll", placeCornerQrs, { passive: true });
        window.addEventListener("resize", placeCornerQrs);
      }
      // The bubble's top-right corner, shared by its QR codes and weather face:
      // in the header row's corner (the header keeps room for it); a bubble with
      // no header has them on their own line, at the right.
      function cornerBox(bubble, textEl, width) {
        var box = bubble.querySelector(".gopod-corner");
        var head = bubble.querySelector(".brobot-bubble-head");
        if (!box) {
          box = document.createElement("div");
          box.className = "gopod-corner";
          if (head) {
            bubble.classList.add("has-corner");
            bubble.appendChild(box);
          } else {
            box.classList.add("no-head");
            bubble.insertBefore(box, textEl);
          }
        }
        if (head) { head.style.paddingRight = ((parseInt(head.style.paddingRight, 10) || 0) + width) + "px"; }
        return box;
      }
      // 2026-10-14 (operator: "In chat bubbles, use the animations where
      // possible (cold has no animation) * pre-load related audio to play @ 10 %
      // volume * use same rule: aligned top-right * display 50% of actual
      // size"): a bubble holding the live weather line (its "degrees Celsius")
      // gets Vector's own weather face for it (webroot/weather/, the weather
      // kit), in the speaking Brobot's eye colour, and that weather's sound
      // plays once as the bubble shows (weather_card, gopod_brobots_settings.json).
      var weatherCard = null;
      // Vector's weather face for these weather words ("light rain") at this
      // hour: {name, sound}, or null (weather_card.faces; clear at night is the
      // stars).
      function weatherFaceFor(words, hour) {
        var face = null;
        (weatherCard.faces || []).some(function (f) {
          var hit = (f.words || []).some(function (w) { return words.indexOf(String(w).toLowerCase()) >= 0; });
          if (hit) { face = f; }
          return hit;
        });
        if (!face) { return null; }
        var picked = { name: face.face, sound: face.sound };
        if (face.night_face && hour !== null && (hour >= (weatherCard.night_from || 20) || hour < (weatherCard.night_to || 6))) {
          picked = { name: face.night_face, sound: face.night_sound || face.sound };
        }
        return picked;
      }
      // Vector's face screen with the weather face, the magnifier as every picture.
      function weatherScreen(name, row) {
        var screen = document.createElement("div");
        screen.className = "gopod-weather-face";
        screen.title = name;
        var tint = document.createElement("div");
        tint.className = "tint";
        tint.style.webkitMaskImage = tint.style.maskImage = "url(" + weatherCard.folder + "/faces/" + name + ".webp)";
        screen.appendChild(tint);
        magnify(row, screen);
        return screen;
      }
      // His weather line is said: the newest waiting weather picture of his shows.
      function weatherShowFor(brobot) {
        var cls = brobot === "Brobot 2" ? "brobot-2" : "brobot-1";
        var boxes = container.querySelectorAll(".brobot-bubble." + cls + " .gopod-weather-after:not(.shown)");
        var box = boxes.length ? boxes[boxes.length - 1] : null;
        if (box && box._show) { box._show(); }
      }
      // The weather sound, loaded now (played by its picture's showing).
      function weatherSoundLoad(sound) {
        if (!sound) { return null; }
        var audio = new Audio(normSrc(weatherCard.folder + "/sounds/" + sound + ".mp3"));
        audio.preload = "auto";
        audio.addEventListener("ended", function () { audio._playedOnce = true; });
        try { audio.load(); } catch (err) { /* no audio */ }
        return audio;
      }
      // The weather sound, loaded now, played once as bubble shows (weather_card.volume).
      function weatherSoundOnShow(bubble, sound) {
        if (restoringChat || !sound) { return null; }
        var audio = new Audio(normSrc(weatherCard.folder + "/sounds/" + sound + ".mp3"));
        audio.preload = "auto";
        audio.addEventListener("ended", function () { audio._playedOnce = true; });
        try { audio.load(); } catch (err) { /* no audio */ }
        whenShown(bubble, function () {
          // 2026-10-17: a share of the master, like every volume knob.
          audio.volume = chatLevel(volumeShare(weatherCard.volume));
          audio.play().catch(function () { audio._playedOnce = true; /* the page hasn't been clicked yet */ });
        }, 60000);
        return audio;
      }
      // A Brobot bubble with the weather line, during a run: the face in its
      // corner. 2026-10-14 (operator: "during live, can you treat 'weather'
      // requests like live pop-up bubbles?"): in the live chat the weather is
      // the pop-up's instead (showPictures), so it shows once.
      function addWeatherFace(bubble) {
        // 2026-10-17 (operator: "no longer easter eggs: Weather"): live too.
        if (!weatherCard || !bubble.classList || !bubble.classList.contains("brobot-bubble") || bubble.querySelector(".gopod-weather-face")) { return; }
        // His own words - not a notice's or a manager's part nested in it.
        var textEl = bubble.querySelector(":scope > .brobot-bubble-text");
        if (!textEl) { return; }
        var said = String(textEl.textContent || "").toLowerCase();
        // Only the weather's own words ("... degrees Fahrenheit, and light rain,
        // as of ..."), never the place's ("Windsor" holds "wind").
        // 2026-10-22: the Weather Report form too ("68° Fahrenheit and clear sky as of").
        var what = /(?:degrees|°) (?:celsius|fahrenheit),? and (.+?),? as of /.exec(said);
        if (!what) { return; }
        var hour = null;
        var at = /as of (\d{1,2}):?\d{2}( ?[ap]m)?/.exec(said);
        if (at) {
          hour = parseInt(at[1], 10) % (at[2] ? 12 : 24);
          if (at[2] && at[2].indexOf("p") >= 0) { hour += 12; }
        }
        var face = weatherFaceFor(what[1], hour);
        if (!face) { return; }
        // 2026-10-22 (operator: "pin weather image for later * remove weather image
        // from response for now, too much complexity * keep the 1st weather audio
        // played"): with weather_card.picture off, only its sound, once, as the
        // bubble shows.
        if (weatherCard.picture === false) {
          if (!bubble._weatherSound) { bubble._weatherSound = true; weatherSoundOnShow(bubble, face.sound); }
          return;
        }
        // 2026-10-22 (operator: "current position isn't a good spot for it * add
        // weather image centered below, after weather report fully spoken *
        // weather audio still plays 1st without the weather image * then the rest
        // of the response"): the picture goes centred right under the weather
        // line, hidden until that line has been said - then it shows and its
        // sound plays, together. (A bubble without pieces: the corner, as before.)
        var pieces = Array.prototype.filter.call(textEl.querySelectorAll(".brobot-say"), function (x) { return !x.closest(".nested-host"); });
        var endPiece = null, seenAsOf = false;
        for (var p = 0; p < pieces.length && !endPiece; p++) {
          var pt = String(pieces[p].textContent || "").toLowerCase();
          if (pt.indexOf("as of ") >= 0) { seenAsOf = true; }
          if (seenAsOf && /[.!?]["'”]?\s*$/.test(pt.trim())) { endPiece = pieces[p]; }
        }
        bubble.classList.add("has-weather");
        var box;
        if (endPiece) {
          box = document.createElement("div");
          box.className = "gopod-weather-after";
          endPiece.parentNode.insertBefore(box, endPiece.nextSibling);
        } else {
          box = cornerBox(bubble, textEl, 190);
        }
        var screen = weatherScreen(face.name, function () { return box; });
        box.appendChild(screen);
        // 2026-10-22 (operator: "keep as-was, play audio (less the image) once
        // only upon bubble display * then weather image centered underneath + 3 s
        // wait"): the sound once as the bubble shows (a background effect); the
        // picture once his weather line is said.
        var sound = weatherSoundOnShow(bubble, face.sound);
        function showNow() { box.classList.add("shown"); }
        // Shown when Wire-Pod says his weather line is said ("weathershow:",
        // weatherShowFor), or once the next piece lights, whichever comes first.
        if (!endPiece || restoringChat || endPiece.classList.contains("spoken")) {
          if (endPiece) { box.classList.add("shown"); if (sound) { sound._playedOnce = true; } } else { showNow(); }
        } else {
          var once = false;
          box._show = function () { if (once) { return; } once = true; watch.disconnect(); showNow(); };
          var watch = new MutationObserver(function () {
            if (endPiece.classList.contains("spoken")) { box._show(); }
          });
          watch.observe(endPiece, { attributes: true, attributeFilter: ["class"] });
        }
        // 2026-10-22 (operator: "after weather audio has finished playing the 1st
        // time, upon zoomed mouseclick, play the image's audio again * can only
        // play once at a time no matter how many image clicks"): after the first
        // play, a click plays it again - never twice at once.
        if (sound) {
          screen.addEventListener("click", function () {
            // Any click (the hover already zooms it; a click may unzoom it).
            if (!sound._playedOnce || !sound.paused) { return; }
            sound.currentTime = 0;
            sound.volume = chatLevel(volumeShare(weatherCard.volume));
            sound.play().catch(function () {});
          });
        }
      }
      // 2026-10-22 (operator: "Goverlord orchestrates * Gomads do most of the
      // thinking * Gominions do with little to no thinking * There should be 5
      // bubbles essentially now"): every bubble in the chat gets its kind
      // (bubble_kinds.rules, first match: operator, gominion, other, gomad,
      // goverlord) as data-tier; one with none is logged as a loose chain.
      var bubbleKinds = [], untaggedSeen = {};
      function tagBubbleKind(node) {
        if (!node || !node.classList || !node.classList.contains("brobot-bubble") || !bubbleKinds.length) { return; }
        var kind = "";
        for (var i = 0; i < bubbleKinds.length && !kind; i++) {
          var r = bubbleKinds[i];
          (r.has || []).forEach(function (c) {
            if (kind) { return; }
            if (c === "[own]" ? node.dataset.own === "1" : node.classList.contains(c)) { kind = r.kind; }
          });
        }
        if (kind) { node.dataset.tier = kind; return; }
        var sig = node.className;
        if (untaggedSeen[sig]) { return; }
        untaggedSeen[sig] = true;
        fetch("/api-sdk/gopod_page_note", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "note=" + encodeURIComponent("GOPOD_CHECK bubble without a kind (loose chain): " + sig) }).catch(function () {});
      }
      function addQrCodesToAll() {
        Array.prototype.forEach.call(container.querySelectorAll(".brobot-bubble"), addQrCodes);
      }
      if (container) {
        new MutationObserver(function (changes) {
          changes.forEach(function (change) {
            Array.prototype.forEach.call(change.addedNodes, function (node) { addQrCodes(node); addWeatherFace(node); tagBubbleKind(node); });
          });
        }).observe(container, { childList: true });
      }
      // GOLDEN magnifier, one helper (2026-10-18, operator: "whatever other
      // processes are the same but scattered, share from one shared slot"):
      // every picture, QR code and weather face - the one under the mouse grows
      // (the others in its row shrink), a tap toggles it, leaving the row
      // shrinks it. row may be a function giving the row (the weather face's).
      function magnify(row, el) {
        var rowOf = function () { return typeof row === "function" ? row() : row; };
        el.addEventListener("mouseenter", function () { zoomOnly(rowOf(), el); });
        el.addEventListener("click", function () {
          if (el.classList.contains("zoomed")) { el.classList.remove("zoomed"); } else { zoomOnly(rowOf(), el); }
        });
        var r = rowOf();
        if (r && !r._magnified) {
          r._magnified = true;
          r.addEventListener("mouseleave", function () { zoomOnly(r, null); });
        }
      }
      // GOLDEN "once it shows", one helper: fn runs once the bubble is shown
      // (placed, not hidden, not staged) - at once if it already is.
      function whenShown(bubble, fn, mostMs) {
        var ran = false;
        function tryRun() {
          if (ran || !bubble.isConnected || bubble.style.visibility === "hidden" || bubble.dataset.staged === "1") { return false; }
          ran = true;
          fn();
          return true;
        }
        if (tryRun()) { return; }
        var watch = new MutationObserver(function () { if (tryRun()) { watch.disconnect(); } });
        watch.observe(bubble, { attributes: true, attributeFilter: ["style", "data-staged"] });
        if (mostMs) { setTimeout(function () { watch.disconnect(); }, mostMs); }
      }
      function zoomOnly(row, img) {
        Array.prototype.forEach.call(row.querySelectorAll("img, .gopod-weather-face"), function (other) {
          other.classList.toggle("zoomed", other === img);
        });
      }
      // 2026-10-06 (operator: "have the female/male voices speak operator
      // bubble"; "can you channel the audio to play through the wire-pod web
      // page?"): the reporter hosts (Brobots 3 and 4) - their opening and closing
      // lines in a system bubble under the host's name, and every voice file
      // Wire-Pod makes played here (reporter_voices, gopod_reporter_voices.go).
      // A browser plays sound only after the page has been clicked or typed in.
      // 2026-10-06 (operator: "Reporter 1, Brobot 3:" / "Reporter 2, Brobot 4:"):
      // the header is the host's label (reporter_voices.labels).
      var reporterRole = "Reporter Voice", hostNames = {}, hostLabels = {};
      // 2026-10-22 (operator: "a bubble's header in a scripted run should have
      // what values? default (CHALK), fallback (Brobot 7), remembered edits
      // (CHALK BORED)? * editable"): the header template (display_names) - a
      // remembered edit, else the default (the GOPOD persona), else the fallback
      // from the numpad map ("Brobot 7": the key whose persona it is), else its label.
      var displayNames = {};
      // 2026-10-22 (operator: "numpad mapping sounds like a good template to form
      // from"): the default is the numpad map's persona - a scripted voice's own
      // persona (persona_reporters: "Brobot 3" -> CHALK) or the persona itself -
      // unless display_names.default names one.
      function numpadPersona(who) {
        var on = numpad.map && numpad.map.numlock_on, name = (numpad.personaReporters || {})[who] || who;
        for (var k in (on || {})) {
          if (on[k] && String(on[k].persona || "").toUpperCase() === String(name).toUpperCase()) { return { name: on[k].persona, key: k }; }
        }
        return null;
      }
      function hostHeader(who) {
        var np = numpadPersona(who);
        var dn = displayNames || {}, edits = dn.edits || {}, def = (dn["default"] || {})[who] || (np ? np.name : "");
        if (def && edits[def]) { return edits[def]; }
        if (edits[who]) { return edits[who]; }
        if (def) { return def; }
        if (np && dn.fallback) { return String(dn.fallback).replace("{n}", np.key.replace(/^KP/, "")); }
        return hostLabels[who] || who;
      }
      // 2026-10-09 (the nap, a lyric video): a song's soundtrack plays in the
      // background, and its captions show as the narrator's bubbles.
      var songAudioRole = "Song Audio", narratorRole = "Narrator", songAudio = null;
      // GOLDEN nesting (2026-10-18, operator: "If I ask brobot N = weather trigger *
      // PLAYHEAD takes over in PLAYHEAD bubble * PLAYHEAD bubble displays inside
      // brobot N bubble * at top before brobot N 'Got it'"; QR: "CHALK bubble
      // displays inside brobot N bubble * at bottom after brobot N response"): a
      // manager's line for Brobot N opens his bubble (a shell, his header) with
      // the manager's part at the top; his reply fills the same bubble below it
      // (appendBubble); CHALK's QR codes go in at the bottom (chalkQrs).
      var nestShells = {};
      // A Brobot's bubble made after an operator line already has its question
      // above it (dataset.asked); one made before (a spoken question, shown when
      // the answer is done) takes the backpack line above it.
      var operatorSeq = 0, shellOperatorSeq = {};
      function nestShell(brobot) {
        var shell = nestShells[brobot];
        if (shell && shell.isConnected && !shell.dataset.filled) { return shell; }
        shell = document.createElement("div");
        if (operatorSeq > (shellOperatorSeq[brobot] || 0)) { shell.dataset.asked = "1"; }
        shellOperatorSeq[brobot] = operatorSeq;
        shell.className = "brobot-bubble " + (brobot === "Brobot 2" ? "brobot-2" : "brobot-1") + " nest-shell";
        shell.dataset.nest = brobot;
        shell.appendChild(bubbleHead(brobot));
        container.appendChild(shell);
        nestShells[brobot] = shell;
        if (!restoringChat) { revealWhenPlaced(shell); }
        return shell;
      }
      // A manager's part (its header and colour) inside a Brobot's bubble.
      function nestBox(shell, who, atTop) {
        var box = document.createElement("div");
        box.className = "brobot-bubble operator system-status host nested-host host-" + String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-");
        box.dataset.host = who;
        box.appendChild(reporterHead(hostHeader(who)));
        // 2026-10-22 (operator's notice log, line 3): a manager's part at the top
        // goes in his notice, after its lines - its header and line on one row.
        var notice = atTop ? shell.querySelector(":scope > .bubble-notice") : null;
        if (notice) {
          box.classList.add("in-notice");
          notice.appendChild(box);
          return box;
        }
        var reply = shell.querySelector(":scope > .brobot-bubble-text");
        if (atTop && reply) { shell.insertBefore(box, reply); } else { shell.appendChild(box); }
        return box;
      }
      // PLAYHEAD's QR ("qr:<picture>|<url>"): the code and its address, shown, not said.
      function hostQr(box, text) {
        var parts = String(text).slice(3).split("|");
        var fig = document.createElement("figure");
        fig.className = "chalk-qr-code";
        var img = document.createElement("img");
        img.className = "manager-qr";
        img.src = parts[0];
        img.alt = img.title = parts[1] || "";
        var row = document.createElement("div");
        row.className = "topic-pictures";
        magnify(row, img);
        fig.appendChild(img);
        var cap = document.createElement("figcaption");
        cap.textContent = String(parts[1] || "").replace(/^https?:\/\//, "");
        fig.appendChild(cap);
        row.appendChild(fig);
        box.appendChild(row);
      }
      // CHALK's part for Brobot N (2026-10-18, operator: "CHALK BUBBLE === I can
      // help with that. QR code displayed." after "Brobot 1: [response]"): at
      // the bottom of his newest bubble, after his answer.
      function chalkBottom(who, text, to) {
        var cls = to === "Brobot 2" ? "brobot-2" : "brobot-1", target = null;
        Array.prototype.forEach.call(container.querySelectorAll(":scope > .brobot-bubble." + cls), function (b) { target = b; });
        if (!target) { return false; }
        var boxes = target.querySelectorAll(":scope > .nested-host.host-" + String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-"));
        if (String(text).indexOf("qr:") === 0) {
          if (!boxes.length) { return false; }
          hostQr(boxes[boxes.length - 1], text);
          return true;
        }
        var box = nestBox(target, who, false);
        var nline = document.createElement("span");
        nline.className = "brobot-bubble-text";
        box.appendChild(nline);
        voiceLitLine(target, nline, text);
        return true;
      }
      // GOLDEN manager bubble (2026-10-18, operator: "PLAYHEAD ["Generating..." +
      // hourglass]/[I'll help you with that.] ["Generating..." + hourglass]/[weather]
      // * all [message0]/[message1] in the same bubble * still do a generated
      // file checks * instantly toggle/update upon voice file generated & ready
      // to play"): a manager's next line goes in its bubble from this exchange
      // (its own, or its part in a Brobot's), each line waiting with the
      // hourglass while its voice is made, its words when it's ready.
      var managerHosts = {};
      function hostSlug(who) { return String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-"); }
      // The manager's bubble a next line joins: the newest in the chat, nothing after it.
      function openHostBox() {
        var last = container.lastElementChild;
        while (last && (last === warmupBubble || last === warmupBeforeOpening || last.classList.contains("generating-slot"))) { last = last.previousElementSibling; }
        if (!last || !last.dataset) { return null; }
        if (last.dataset.host && !last.classList.contains("nested-host")) { return { box: last, outer: last, who: last.dataset.host }; }
        if (last.classList.contains("nest-shell") && !last.dataset.filled) {
          var boxes = last.querySelectorAll(":scope > .nested-host[data-host]");
          if (boxes.length) { var b = boxes[boxes.length - 1]; return { box: b, outer: last, who: b.dataset.host }; }
        }
        return null;
      }
      // The line's words in its manager's bubble, a new line in it.
      function hostLineInto(who, text) {
        var at = openHostBox();
        if (!at || at.who !== who || !managerHosts[who]) { return null; }
        var row = document.createElement("div");
        row.className = "host-line";
        var line = document.createElement("span");
        line.className = "brobot-bubble-text";
        row.appendChild(line);
        at.box.appendChild(row);
        voiceLitLine(at.outer, line, text);
        return row;
      }
      function showHostLine(who, text, to, then) {
        var next = function () { if (then) { then(); } };
        var hostClass = "operator system-status host host-" + String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-");
        if (String(text).indexOf("qr:") !== 0) {
          var joined = hostLineInto(who, text);
          if (joined) {
            if (!restoringChat && !/^Brobot [12]$/.test(String(to || ""))) { recordBubble(who, String(text), "host"); }
            if (/^Brobot [12]$/.test(String(to || ""))) { addWeatherFace(joined.closest(".nested-host") || joined); }
            next();
            return;
          }
        }
        if (/^Brobot [12]$/.test(String(to || ""))) {
          if (who === displayHost && chalkBottom(who, text, to)) { next(); return; }
          // Inside Brobot N's bubble, at its top.
          generatingGone();
          var shell = nestShell(to);
          var boxes = shell.querySelectorAll(".nested-host.host-" + String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-"));
          var box = boxes.length && String(text).indexOf("qr:") === 0 ? boxes[boxes.length - 1] : nestBox(shell, who, true);
          if (String(text).indexOf("qr:") === 0) { hostQr(box, text); next(); return; }
          var nline = document.createElement("span");
          nline.className = "brobot-bubble-text";
          box.appendChild(nline);
          voiceLitLine(shell, nline, text);
          addWeatherFace(box);
          next();
          return;
        }
        // Its own bubble, on the raw bubble template (built and loaded hidden, then shown).
        var bubble = rawBubble({ kind: "host", classes: hostClass, head: reporterHead(hostHeader(who)) });
        bubble.dataset.host = who;
        if (String(text).indexOf("qr:") === 0) {
          hostQr(bubble, text);
          placeBubble(bubble, next);
          return;
        }
        if (!restoringChat) { recordBubble(who, String(text), "host"); }
        var line = document.createElement("span");
        line.className = "brobot-bubble-text";
        bubble.appendChild(line);
        voiceLitLine(bubble, line, text);
        placeBubble(bubble, next, function () {
          // 2026-10-06 (operator: "ensure this timer is always 'after' all
          // reporter bubbles as they appear"): the warm-up timer moves back under it.
          if (warmupBubble && warmupBubble.isConnected) { container.appendChild(warmupBubble); }
          else if (warmupBeforeOpening && warmupBeforeOpening.isConnected) { container.appendChild(warmupBeforeOpening); }
        });
      }

      // 2026-10-09 (operator: "instead of line by line, do several phrases in
      // fewer bigger chat bubbles * use same highlighting per narration
      // segment"): the song engine sends one narrator line per poem section, its
      // captions joined by the segment marker, each "@start-end words". The
      // bubble shows the words; each phrase lights at its seconds on the
      // soundtrack's own clock (songAudio), or, with no soundtrack playing, from
      // when the bubble arrived.
      // GOLDEN chat bubble rule (operator 2026-10-09: "scrolling sucks with
      // narrator bubbles, fix like the others & set a golden rule"; "draw out
      // full narrator bubbles off screen or while invisible * then display/make
      // visible the full narrator chat bubble at actual speech time"): the song
      // sends the bubble reporter_voices.narrator_lead_seconds early; it is built
      // whole while invisible (taking its room), the chat scrolls once so all of
      // it will be in view, and it shows at its first line's second. Its phrases
      // light in place - nothing scrolls again until the next bubble.
      // 2026-10-10 (operator: "the bubble scrolls up and highlighting stops"):
      // the next bubble arrives narrator_lead_seconds early, while this one is
      // still being sung. It waits hidden, takes its room below without
      // scrolling, and the chat scrolls for it only once the bubble above is
      // done (or CHAT_SCROLL_MS before its own first line, whichever is later);
      // each bubble lights its own phrases to its own end.
      var NARRATOR_PIECE_RE = /^@(\d+(?:\.\d+)?)-(\d+(?:\.\d+)?)\s+(?:~([A-Za-z0-9+\/=]+)\s+)?/;
      // One narrator piece's span: its words, its seconds, and - when the song
      // has them (forced alignment, 2026-10-18) - each word's time from its start.
      function narratorSpan(piece) {
        var m = NARRATOR_PIECE_RE.exec(piece.trim());
        var span = document.createElement("span");
        span.className = "brobot-say";
        span.textContent = m ? piece.trim().slice(m[0].length) : piece.trim();
        if (m) { span.dataset.at = m[1]; span.dataset.end = m[2]; }
        if (m && m[3]) {
          try {
            var rel = JSON.parse(decodeURIComponent(escape(atob(m[3]))));
            span._spokenWords = rel.map(function (w) { return { t: w[0], s: w[1], e: w[2] }; });
          } catch (err) { /* no times: the scaled reveal */ }
        }
        return span;
      }
      var narratorLines = [], narratorTimer = null, narratorLead = 0;
      function showNarratorLine(text) {
        if (!restoringChat) { recordBubble(narratorRole, String(text), "narrator"); }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator system-status host narrator";
        bubble.appendChild(systemHead((hostLabels[narratorRole] || narratorRole) + ":"));
        var line = document.createElement("div");
        line.className = "brobot-bubble-text brobot-bubble-line";
        var pieces = segmentMarker ? String(text).split(segmentMarker) : [String(text)];
        // 2026-10-15 (operator: "one image for each narrator bubble * image is
        // inside the narrator bubble under highlighted text"): a "picture:<path>"
        // piece (the song's section_pictures) shows under the words, magnifier
        // and all, like every chat picture.
        var narratorPictures = pieces.filter(function (p) { return p.trim().indexOf("picture:") === 0; });
        var narratorSingle = pieces.some(function (p) { return p.trim() === "narrator:single"; });
        var narratorLast = pieces.some(function (p) { return p.trim() === "narrator:last"; });
        pieces = pieces.filter(function (p) { return p.trim().indexOf("picture:") !== 0 && p.trim().indexOf("narrator:") !== 0; });
        if (narratorSingle) { showNarratorSingle(pieces, narratorPictures, narratorLast); return; }
        pieces.forEach(function (piece, i) {
          var span = narratorSpan(piece);
          if (i) { line.appendChild(document.createTextNode(" ")); }
          line.appendChild(span);
        });
        bubble.appendChild(line);
        if (narratorPictures.length) {
          var picRow = document.createElement("div");
          picRow.className = "topic-pictures";
          narratorPictures.forEach(function (p) {
            var img = document.createElement("img");
            img.src = p.trim().slice("picture:".length);
            img.alt = "";
            magnify(picRow, img);
            picRow.appendChild(img);
          });
          bubble.appendChild(picRow);
        }
        container.appendChild(bubble);
        if (warmupBubble && warmupBubble.isConnected) { container.appendChild(warmupBubble); }
        if (restoringChat) { return; }
        var first = line.querySelector(".brobot-say[data-at]");
        if (!first) { revealWhenPlaced(bubble); return; }
        bubble.style.visibility = "hidden";
        // 2026-10-13 (operator: "good catch"): built out of the chat's flow too,
        // taking no room until it's placed.
        stageBubble(bubble);
        // A narrator line is never the reply an operator line waits for (live
        // 2026-10-09: after "Start Baby Brobots" the chat kept aiming at that
        // line, stuck near the top, as the narrator bubbles grew below it).
        awaitingReply = false;
        operatorInBatch = false;
        narratorLines.push({ bubble: bubble, firstAt: parseFloat(first.dataset.at), shownMs: Date.now(), lit: null, shown: false, placed: false });
        if (narratorTimer === null) { narratorTimer = setInterval(lightNarrator, 100); }
        lightNarrator();
      }
      // 2026-10-15 (the nap's single-bubble copy, operator: "use only a single
      // narrator chat bubble for all 5 narrations"; "A single narration bubble
      // stays the size of the image * 5 narrations are run in sequence";
      // "narration take replaces previous take"; "same applies to static
      // background images"): one narrator bubble, its picture's own shape at the
      // bubble's full width (the magnifier's size). Each section is a part: its
      // picture behind its words, fading in to 50% as the section starts and out
      // after its last words; the next section's part then replaces it.
      var narratorStage = null;
      var NARRATOR_FADE_IN_S = 1, NARRATOR_FADE_OUT_MAX_S = 1.5, NARRATOR_FADE_OUT_END_S = 2;
      function showNarratorSingle(pieces, pictures, last) {
        var line = document.createElement("div");
        line.className = "brobot-bubble-text brobot-bubble-line";
        pieces.forEach(function (piece, i) {
          var span = narratorSpan(piece);
          if (i) { line.appendChild(document.createTextNode(" ")); }
          line.appendChild(span);
        });
        var first = line.querySelector(".brobot-say[data-at]");
        var firstAt = first ? parseFloat(first.dataset.at) : 0;
        // A new run starts over (its first section comes earlier than the last one).
        var reuse = narratorStage && narratorStage.isConnected && firstAt > narratorStage._lastAt;
        var bubble, stage;
        if (reuse) {
          bubble = narratorStage;
          stage = bubble.querySelector(".narrator-stage");
        } else {
          bubble = document.createElement("div");
          // 2026-10-15 (operator: "remove narrator header"): the picture is the bubble.
          bubble.className = "brobot-bubble operator system-status host narrator narrator-single";
          stage = document.createElement("div");
          stage.className = "narrator-stage";
          bubble.appendChild(stage);
          container.appendChild(bubble);
          if (warmupBubble && warmupBubble.isConnected) { container.appendChild(warmupBubble); }
          narratorStage = bubble;
        }
        bubble._lastAt = firstAt;
        var part = document.createElement("div");
        part.className = "narrator-part";
        var bg = null;
        if (pictures.length) {
          bg = document.createElement("img");
          bg.className = "narrator-bg";
          bg.alt = "";
          bg.src = pictures[0].trim().slice("picture:".length);
          bg.addEventListener("load", function () {
            if (!stage.style.aspectRatio && bg.naturalWidth) { stage.style.aspectRatio = bg.naturalWidth + " / " + bg.naturalHeight; }
          });
          part.appendChild(bg);
        }
        part.appendChild(line);
        stage.appendChild(part);
        if (restoringChat || !first) {
          narratorShowPart(stage, part, 0);
          Array.prototype.forEach.call(part.querySelectorAll(".brobot-say"), function (s) { s.classList.add("spoken"); });
          if (!restoringChat && !reuse) { revealWhenPlaced(bubble); }
          return;
        }
        if (!reuse) {
          bubble.style.visibility = "hidden";
          stageBubble(bubble);
        }
        awaitingReply = false;
        operatorInBatch = false;
        narratorLines.push({ bubble: bubble, part: part, bg: bg, last: last, firstAt: firstAt, shownMs: Date.now(), lit: null, shown: false, placed: !!reuse });
        if (narratorTimer === null) { narratorTimer = setInterval(lightNarrator, 100); }
        lightNarrator();
      }
      // This section's part replaces the last; its picture fades in over fadeIn s.
      function narratorShowPart(stage, part, fadeIn) {
        Array.prototype.forEach.call(stage.querySelectorAll(".narrator-part.current"), function (p) { p.classList.remove("current"); });
        part.classList.add("current");
        var bg = part.querySelector(".narrator-bg");
        if (bg) {
          bg.style.transitionDuration = fadeIn + "s";
          void bg.offsetWidth;
          bg.classList.add("on");
        }
      }
      function narratorSeconds(n) {
        if (songAudio && !songAudio.paused && !songAudio.ended && songAudio.currentTime > 0) { return songAudio.currentTime; }
        // No soundtrack playing yet (the nap's first bubble comes before it): the
        // bubble was sent narrator_lead_seconds early, so it waits that long.
        return n.firstAt - narratorLead + (Date.now() - n.shownMs) / 1000;
      }
      function lightNarrator() {
        narratorLines = narratorLines.filter(function (n) { return n.bubble.isConnected; });
        if (!narratorLines.length) {
          clearInterval(narratorTimer);
          narratorTimer = null;
          return;
        }
        narratorLines.slice().forEach(function (n, k) {
          var t = narratorSeconds(n);
          if (!n.placed) {
            // Scroll for it once the bubble above is done, a scroll's length
            // ahead of its first line at the earliest.
            if (k > 0 || t < n.firstAt - CHAT_SCROLL_MS / 1000 || batonHeld()) { return; }
            n.placed = true;
            unstageBubble(n.bubble);
            scrollChatTo(bubbleTarget(n.bubble));
          }
          if (!n.shown) {
            if (t < n.firstAt) { return; }
            n.shown = true;
            n.bubble.style.visibility = "";
            if (n.part) { narratorShowPart(n.part.parentNode, n.part, NARRATOR_FADE_IN_S); }
          }
          var spans = (n.part || n.bubble).querySelectorAll(".brobot-say[data-at]");
          var pick = null;
          for (var i = 0; i < spans.length; i++) {
            var next = spans[i + 1];
            var until = next ? parseFloat(next.dataset.at) : parseFloat(spans[i].dataset.end);
            if (t >= parseFloat(spans[i].dataset.at) && t < until) { pick = spans[i]; break; }
          }
          // The same phrase: keep it lit (a robot's highlight clears all of them).
          if (pick === n.lit) { if (pick) { pick.classList.add("speaking"); } return; }
          if (n.lit) { n.lit.classList.remove("speaking"); n.lit.classList.add("spoken"); }
          n.lit = pick;
          if (pick) {
            lightSpan(pick);
          } else if (t >= parseFloat(spans[spans.length - 1].dataset.end)) {
            // The section is over; the next bubble scrolls in.
            revealRest(n.part || n.bubble);
            // Single bubble: its picture fades out before the next section
            // (over the gap, at most NARRATOR_FADE_OUT_MAX_S; the last one slower).
            // The last section's picture stays at 50% (operator 2026-10-15).
            if (n.bg && !n.last) {
              var after = narratorLines[narratorLines.indexOf(n) + 1];
              var gap = after ? after.firstAt - parseFloat(spans[spans.length - 1].dataset.end) : NARRATOR_FADE_OUT_END_S;
              n.bg.style.transitionDuration = Math.max(0.5, Math.min(after ? NARRATOR_FADE_OUT_MAX_S : NARRATOR_FADE_OUT_END_S, gap)) + "s";
              n.bg.classList.remove("on");
            }
            narratorLines.splice(narratorLines.indexOf(n), 1);
          }
        });
      }

      // 2026-10-12 (operator: '"Bingo ball captured. [ball number]" * upon
      // saying/highlighting "B-1" = [ball number] image appears underneath'): the
      // called ball waits for the next Brobot bubble that says it, goes in under
      // its words (its room kept while hidden), and shows when that piece lights.
      var pendingBall = null;
      function showBingoBall(call) {
        if (!restoringChat) { recordBubble(ballRole, String(call).trim(), "ball"); }
        var under = document.createElement("div");
        under.className = "bingo-ball-under";
        under.appendChild(buildBingoBall(call));
        pendingBall = { call: String(call).trim(), el: under };
      }
      function attachPendingBall(bubble, spoken) {
        if (!pendingBall || String(spoken).indexOf(pendingBall.call) < 0) { return; }
        var ball = pendingBall;
        pendingBall = null;
        bubble.appendChild(ball.el);
        if (restoringChat) { return; }
        ball.el.classList.add("waiting");
        var show = function () {
          ball.el.classList.remove("waiting");
          if (watch) { watch.disconnect(); }
          clearTimeout(most);
        };
        var watch = new MutationObserver(function () {
          var lit = bubble.querySelectorAll(".speaking");
          for (var i = 0; i < lit.length; i++) {
            if (lit[i].textContent.indexOf(ball.call) >= 0) { show(); return; }
          }
        });
        watch.observe(bubble, { subtree: true, attributes: true, attributeFilter: ["class"] });
        var most = setTimeout(show, 30000); // never left hidden
      }
      function buildBingoBall(call) {
        var parts = String(call).trim().split("-");
        var letter = parts.length > 1 ? parts[0].toUpperCase() : "";
        var number = parts.length > 1 ? parts.slice(1).join("-") : String(call).trim();
        var bubble = document.createElement("div");
        bubble.className = "bingo-ball ball-" + letter.toLowerCase();
        var face = document.createElement("div");
        face.className = "bingo-ball-face";
        var inner = document.createElement("div");
        inner.className = "bingo-ball-inner";
        var l = document.createElement("span");
        l.className = "bingo-ball-letter";
        l.textContent = letter;
        var n = document.createElement("span");
        n.className = "bingo-ball-number";
        n.textContent = number;
        inner.appendChild(l);
        inner.appendChild(n);
        face.appendChild(inner);
        bubble.appendChild(face);
        return bubble;
      }

      // 2026-09-30: Wire-Pod's numbered feed of chat-page lines
      // (webroot/gopod_chat_feed.json, gopod_chat_feed.go): every display line a
      // song or tool sends, each with its own number. Each number shows once, and
      // numbers restart with a new Wire-Pod start (boot), so the key is boot + id.
      var feedPolling = false;
      function pollFeed() {
        if (feedPolling) { return; }
        feedPolling = true;
        fetch("/gopod_chat_feed.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (feed) {
            if (!feed || !Array.isArray(feed.events)) { return; }
            // A new Wire-Pod start (a different boot stamp) clears the chat
            // (operator: "upon wpr 2, clear out brobots chat window contents").
            // The very first boot this browser sees is only remembered.
            if (feed.boot && chatStore.boot !== feed.boot) {
              var hadBoot = !!chatStore.boot;
              feedQueue = [];
              feedQueued = {};
              if (chatStore.boot) {
                clearBubbles();
                chatStore.bubbles = [];
                chatStore.seen = [];
                seenLines = {};
              }
              chatStore.boot = feed.boot;
              saveChat();
              // 2026-10-10 (operator: "upon wpr 2 web page defaults are * Brobots
              // =All * Native? = checked * Numpad? = checked * Record?= not
              // checked"): Numpad? is ticked again at every new start (Brobots is
              // All at every load; Native? and Record? come from Wire-Pod's start).
              try { window.localStorage.setItem(NUMPAD_KEY, "1"); } catch (e) { /* in-page only */ }
              if (!hadBoot) { setNumpadOn(true); }
              // And the chat window is shown again, even if it was hidden with
              // the button (operator 2026-10-01: "Upon restart of wire-pod,
              // ensure brobots chat window is ON/displayed").
              setChatOpen(true);
              showChat(true);
              // 2026-10-19 (operator: "restart default value = background loop
              // icon OFF * restart OFF = no sound, no menu until icon clicked").
              backgroundLoop.setSpeaker(false);
              showSpeaker();
              // 2026-10-06: and the page reloads itself once, so it always runs
              // the chat code Wire-Pod started with (live 15:45: a page loaded
              // before an update played the hosts' opening all at once). The
              // new boot is saved first, so it reloads only once.
              if (hadBoot) {
                window.location.reload();
                return;
              }
            }
            if (feed.boot && feed.boot !== playedBoot) { playedBoot = feed.boot; playedId = 0; playedSent = 0; }
            feed.events.forEach(function (e) {
              var key = "feed|" + feed.boot + "|" + e.id;
              if (isSeen(key) || feedQueued[key] || !e.who || !e.text) { return; }
              queueBubble(key, e.time_ms || Date.now(), null, e);
            });
            runFeedQueue();
          })
          .catch(function () { /* no feed yet (older Wire-Pod, or nothing said) */ })
          .then(function () { feedPolling = false; });
      }

      // 2026-10-06 (operator: "upon restart, all is logged and queued to play
      // upon wire-pod Log button-click ... Just pause the chat bubbles from
      // running until its clicked to and viewable ... upon viewable, start the
      // sequence running whatever's been queued then continue"): feed lines wait
      // in order until the Log section shows the chat and the page has been
      // clicked (a browser plays sound only after a click); then they run one at
      // a time, each reporter voice line to its end (reporter_voices.gap_seconds
      // between). The opening_end_text line tells Wire-Pod the hosts' opening was
      // heard, and only then is it LLM Ready.
      var feedQueue = [], feedQueued = {}, feedRunning = false;
      // 2026-10-17: the page's side of the dialogue - each second, how far it has
      // played the feed goes back to Wire-Pod (/api-sdk/gopod_page_played), which
      // writes "GOPOD_PAGE played <id> ..." under its own "GOPOD_FEED <id> ...".
      var playedId = 0, playedSent = 0, playedBoot = 0, heldMost = { id: 0, ms: 0 };
      // 2026-10-24: told within a quarter second, so a visitor follows closely.
      setInterval(function () {
        if (!playedBoot || playedId <= playedSent) { return; }
        var body = new URLSearchParams();
        body.set("boot", String(playedBoot));
        body.set("id", String(playedId));
        body.set("page", pageId);
        // 2026-10-18 (step 3): the longest a line waited here before it showed.
        if (heldMost.id) { body.set("held_id", String(heldMost.id)); body.set("held_ms", String(heldMost.ms)); heldMost = { id: 0, ms: 0 }; }
        playedSent = playedId;
        fetch("/api-sdk/gopod_page_played", { method: "POST", body: body }).catch(function () { playedSent = 0; });
      }, 250);
      var pageClicked = false, openingEndText = "opening_end", voiceGapMs = 300;
      var heardMarker = "";
      // 2026-10-14 (operator: "rule: Display 'generating...' during the lag time *
      // upon voice generated, update to message and play"): Wire-Pod puts
      // generating_marker+"on" in the feed while a shown line's voice is made; a
      // grey bubble says generating_text until the line arrives - a Goverlord
      // line fills it, any other bubble takes its place ("off": the voice failed).
      var generatingMarker = "", generatingText = "Generating...", generatingBubble = null;
      function generatingGone() {
        if (generatingBubble) { generatingBubble.remove(); generatingBubble = null; }
      }
      // 2026-10-16 (operator: "Operator bubble (as voice file is being generated)
      // * update message to \"Voice file generating.\" + hourglass * update
      // message upon voice wav play"): Wire-Pod sends generating_marker+"op"
      // while the Operator line's voice is made - the Operator bubble shows
      // generating_text with the warm-up hourglass, and gets its words back when
      // its voice starts to play ("off": the voice failed - its words come back).
      function generatingGlass(bubble) {
        var row = document.createElement("div");
        row.className = "warmup-timer generating-timer";
        var glass = document.createElement("span");
        glass.className = "countdown-hourglass warmup-hourglass";
        glass.innerHTML = hourglassSprite(0);
        row.appendChild(glass);
        rockHourglass(bubble, glass, 0);
        // 2026-10-23 (operator: "animated hourglass: make sand empty from top/fill
        // the bottom, then loop from top"): a smooth drain, a short hold full at
        // the bottom, then from the top again (hourglass.steps, step_ms, hold_ms).
        var steps = hourglassCfg.steps || 16, stepMs = hourglassCfg.step_ms || 150, holdMs = hourglassCfg.hold_ms || 400;
        (function drain(stage) {
          if (!bubble.isConnected || bubble.classList.contains("ready") || !row.isConnected) { return; }
          glass.innerHTML = hourglassSprite(stage, steps);
          var last = stage >= steps - 1;
          setTimeout(function () { drain(last ? 0 : stage + 1); }, last ? holdMs : stepMs);
        })(0);
        return row;
      }
      // GOLDEN slot, the two halves every "Generating" uses (2026-10-18,
      // operator: "finish step 2 * whatever other processes are the same but
      // scattered, share from one shared slot"): slotWaiting puts
      // generating_text and the hourglass in a bubble (its own words kept);
      // slotReady takes them out and gives it its words (its own, or new ones).
      function slotWaiting(bubble) {
        var text = bubble.querySelector(".brobot-bubble-text");
        if (!text || bubble.querySelector(".generating-timer")) { return false; }
        bubble.dataset.words = text.textContent;
        // 2026-10-22: the one Generating look - its text, then the hourglass
        // (the text's height, on its line; .voice-wait).
        text.textContent = generatingText;
        text.classList.add("voice-wait");
        text.appendChild(generatingGlass(bubble));
        return true;
      }
      function slotReady(bubble, words) {
        var row = bubble.querySelector(".generating-timer");
        if (row) { row.remove(); }
        bubble.classList.add("ready");
        var text = bubble.querySelector(".brobot-bubble-text");
        if (text) { text.classList.remove("voice-wait"); }
        if (text && words == null && bubble.dataset.words != null) { text.textContent = bubble.dataset.words; }
        delete bubble.dataset.words;
        return text;
      }
      // GOLDEN voice template, for a line shown some other way (2026-10-23,
      // operator: "RULE: Every voiced line gets generating+hourglass/voice
      // per-word highlighting * when any voiced line is created, the fucking
      // template applies!!"): Wire-Pod names the voice and its words
      // ("forwait:" being made, "forlit:" made); the newest shown line with those
      // words - a warm-up or opener line, the Operator's own line - becomes the
      // voice's line: "Generating..." and the hourglass while it's made, then lit
      // word by word as it plays. A line already in the template is left as is.
      // Not on the page yet: tied when its voice plays (voiceLitStart).
      var voiceForText = {};
      function voiceLineWords(t) { return String(t).replace(/\s+/g, " ").trim(); }
      function voiceLineSpan(el, t) {
        var walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT), n;
        while ((n = walker.nextNode())) {
          if (n.parentNode && n.parentNode.closest && n.parentNode.closest(".brobot-say")) { continue; }
          var at = n.nodeValue.indexOf(t);
          if (at < 0) { continue; }
          var mid = at > 0 ? n.splitText(at) : n;
          if (mid.nodeValue.length > t.length) { mid.splitText(t.length); }
          var span = document.createElement("span");
          span.className = "brobot-say";
          mid.parentNode.insertBefore(span, mid);
          span.appendChild(mid);
          return span;
        }
        if (voiceLineWords(el.textContent) !== t || el.querySelector(".brobot-say")) { return null; }
        var whole = document.createElement("span");
        whole.className = "brobot-say";
        while (el.firstChild) { whole.appendChild(el.firstChild); }
        el.appendChild(whole);
        return whole;
      }
      function voiceLineTie(file, t, waits) {
        var els = container.querySelectorAll(".brobot-bubble-text, .brobot-bubble-line");
        for (var i = els.length - 1; i >= 0 && i >= els.length - 40; i--) {
          var el = els[i], bubble = el.closest(".brobot-bubble");
          if (!bubble || voiceLineWords(el.textContent).indexOf(t) < 0) { continue; }
          // Already in the template (its own bubble's line, or tied ahead): that line.
          if (el.closest(".brobot-say") && !el.closest(".brobot-say").classList.contains("spoken")) { return el.closest(".brobot-say"); }
          var mine = Array.prototype.filter.call(el.querySelectorAll(".brobot-say"), function (x) {
            return voiceLineWords(x.textContent) === t && !x.classList.contains("spoken");
          })[0];
          if (mine) { return mine; }
          var say = voiceLineSpan(el, t);
          if (!say) { continue; }
          if (skippingReporters) { say.classList.add("spoken"); return say; }
          bubble.classList.add("voice-lit");
          if (waits) { voiceWaitOn(bubble, say, file); }
          var shownAt = 0;
          (function showAnyway() {
            if (!bubble.isConnected || say.classList.contains("speaking") || say.classList.contains("spoken")) { return; }
            if (say._wait && say._wait.isConnected) { setTimeout(showAnyway, 500); return; }
            if (bubble.style.visibility !== "hidden" && bubble.dataset.staged !== "1") {
              if (!shownAt) { shownAt = Date.now(); }
              if (Date.now() - shownAt >= voiceLitWaitMs) { say.classList.add("spoken"); return; }
            }
            setTimeout(showAnyway, 500);
          })();
          return say;
        }
        return null;
      }
      // GOLDEN, from the ground up (2026-10-23, operator: "a message text appears
      // solid, then vanishes then appears as highlighted * TEMPLATE!! ...
      // applies everywhere!!"): a voiced line is in the template from the moment
      // it's drawn, never after - when a line is drawn, any voice named for its
      // words still waiting in the queue ties to it right then, before it shows.
      function voiceTieAhead() {
        if (!generatingMarker || restoringChat) { return; }
        feedQueue.forEach(function (item) {
          var text = item.e && String(item.e.text || "");
          if (!text || text.indexOf(generatingMarker) !== 0) { return; }
          var kind = text.slice(generatingMarker.length);
          var waits = kind.indexOf("forwait:") === 0;
          if (!waits && kind.indexOf("forlit:") !== 0) { return; }
          var rest = kind.slice(waits ? 8 : 7), bar = rest.indexOf("|");
          if (bar <= 0) { return; }
          var file = rest.slice(0, bar), t = voiceLineWords(rest.slice(bar + 1));
          if (!t || voiceWaitFor[file]) { return; }
          var say = voiceLineTie(file, t, waits);
          if (say) { voiceWaitFor[file] = say; }
        });
      }
      function voiceForLine(file, text, waits) {
        var t = voiceLineWords(text);
        if (!t || !file || restoringChat) { return; }
        var say = voiceLineTie(file, t, waits);
        if (say) { voiceWaitFor[file] = say; } else { voiceForText[file] = t; }
      }
      document.addEventListener("pointerdown", function () { pageClicked = true; setTimeout(function () { reportShown(false); }, 0); }, true);
      document.addEventListener("keydown", function () { pageClicked = true; setTimeout(function () { reportShown(false); }, 0); }, true);
      function chatViewable() {
        var log = document.getElementById("section-log");
        // 2026-10-17 (operator: "Fix collapsed chat > \"here\""): a collapsed chat
        // (the Brobots Chat button) is away too - nothing plays into it.
        return pageClicked && document.visibilityState === "visible" &&
          (!log || log.style.display !== "none") && container.offsetParent !== null &&
          container.classList.contains("expanded");
      }
      // 2026-10-07 (operator: "one bubble must complete before the next
      // appears"): the Brobots' replies from the log join this same line, in time
      // order - before, they skipped it, and showed ahead of the questions and
      // reporter lines still waiting (a dead speaker, or the tab away).
      function queueBubble(key, timeMs, run, e) {
        feedQueued[key] = true;
        var item = { key: key, timeMs: timeMs, run: run, e: e, at: Date.now() };
        var at = feedQueue.length;
        while (at > 0 && feedQueue[at - 1].timeMs > timeMs) { at--; }
        feedQueue.splice(at, 0, item);
        // 2026-10-23 (operator: "audio files don't appear to be pre-loading for
        // instant playability upon bubble display"): every voice starts loading
        // the moment it reaches the page, not when its turn comes.
        if (item.e && isVoiceEvent(item.e) && !restoringChat && !stingFiles[String(item.e.text).replace(/#vol=[0-9.]+$/, "").split("/").pop()]) { voiceAudio(item.e.text); }
        runFeedQueue();
      }
      // A log line's time ("2026.10.07 20:17:19"), at the end of its second, so
      // it lines up after a feed line from the same second.
      function logTimeMs(stamp) {
        var m = String(stamp).match(/^(\d{4})\.(\d{2})\.(\d{2}) (\d{2}):(\d{2}):(\d{2})$/);
        if (!m) { return Date.now(); }
        return new Date(+m[1], +m[2] - 1, +m[3], +m[4], +m[5], +m[6]).getTime() + 999;
      }
      function runFeedQueue() {
        if (feedRunning || !feedQueue.length || !chatViewable() || !pageRoleKnown) { return; }
        // GOLDEN, a visitor mirrors the admin (2026-10-24, operator: "GOPOD phone
        // page does nothing but zombie-follows GOPOD admin page display, wherever
        // the fuck admin page display is at"): a visitor never shows a line the
        // main page hasn't reached - at most the one it's on now (played + 1).
        // When the main page moves on, this queue moves on (reportShown).
        if (!isMainPage && feedQueue[0].e && feedQueue[0].e.id > mainPlayed + 1) { return; }
        feedRunning = true;
        var item = feedQueue.shift();
        delete feedQueued[item.key];
        // 2026-10-18 (step 3, "the engine conducts"): how long this line waited
        // here (the bubble above, a countdown, the warm-up) - told to Wire-Pod.
        if (item.e && item.e.id && Date.now() - item.at > heldMost.ms) { heldMost = { id: item.e.id, ms: Date.now() - item.at }; }
        markSeen(item.key);
        var next = function () {
          // 2026-10-17 (operator: "the web page and terminal must have some sort
          // of coordinated dialogue"): how far this page has played the feed.
          if (item.e && item.e.id > playedId) { playedId = item.e.id; }
          feedRunning = false;
          runFeedQueue();
        };
        if (item.run) { item.run(); next(); return; }
        // 2026-10-12: the voice right behind this line loads with its bubble.
        nextBubbleVoice = !isVoiceEvent(item.e) && feedQueue.length && isVoiceEvent(feedQueue[0].e) ? voiceAudio(feedQueue[0].e.text) : null;
        showFeedEvent(item.e, next);
        nextBubbleVoice = null;
      }
      // 2026-10-07 (operator: "make the run pause on web-page not the focus, and
      // continue upon focus again"): Wire-Pod is told when this page is showing
      // the chat or not (another tab, minimised, another section); a playing
      // draft waits before its next question while it isn't
      // (/api-sdk/gopod_page_shown, gopod_interview_script.go).
      var lastShown = null;
      // GOLDEN, one page owns the show (2026-10-23, FINISH_PLAN_001.md C1,
      // gopod_main_page.go): Wire-Pod names one main page - it plays the sound,
      // and its shown/played/heard reports time the run. Every other open page
      // is a viewer: the same bubbles, silent (chatLevel 0), its reports
      // ignored. The page's id lives with its tab, so F5 keeps its role; "?view"
      // in the address makes it a viewer by choice. Nothing plays until its
      // role is known.
      var pageId = "", pageViewOnly = /[?&]view(=|&|$)/.test(location.search);
      var isMainPage = false, pageRoleKnown = false;
      // A viewer follows the main page (2026-10-23, operator: "timing between the
      // phone & pc appear to be in sync then fall out of sync due to slower
      // scrolling on the phone"): more than page_sync.viewer_catch_up_events
      // lines behind it, the viewer catches up - scrolls at once, no waits, its
      // voices passed (it's silent anyway) - until it's back in step.
      var mainPlayed = 0, VIEWER_CATCH_UP_EVENTS = 2;
      function viewerBehind() {
        return pageRoleKnown && !isMainPage && mainPlayed > playedId + VIEWER_CATCH_UP_EVENTS;
      }
      try { pageId = sessionStorage.getItem("gopod_page_id") || ""; } catch (err) { /* no storage */ }
      if (!pageId) {
        pageId = Date.now().toString(36) + Math.random().toString(36).slice(2, 8);
        try { sessionStorage.setItem("gopod_page_id", pageId); } catch (err) { /* no storage */ }
      }
      // A visitor's own ID (2026-10-24, operator: "Save the visitor profile in case
      // of a disconnect reconnect"): kept in this browser, so Wire-Pod knows the
      // same phone again after a reconnect, even on a new network address.
      var visitorId = "";
      try { visitorId = localStorage.getItem("gopod_visitor_id") || ""; } catch (err) { /* no storage */ }
      if (!visitorId) {
        visitorId = "v" + Date.now().toString(36) + Math.random().toString(36).slice(2, 10);
        try { localStorage.setItem("gopod_visitor_id", visitorId); } catch (err) { /* no storage */ }
      }
      function setPageRole(main) {
        var changed = !pageRoleKnown || main !== isMainPage;
        isMainPage = main;
        pageRoleKnown = true;
        if (!changed) { return; }
        backgroundLoop.setVolume(backgroundLoop.volume());
        var body = new URLSearchParams();
        body.set("note", "page " + pageId + (main ? " is the main page (plays the sound)" : " is a viewer (silent)"));
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: body }).catch(function () { /* Wire-Pod gone */ });
        runFeedQueue();
      }
      function reportShown(force) {
        var log = document.getElementById("section-log");
        // 2026-10-17 (survey E1: the page said "shown" before it could play - it
        // plays nothing until it's been clicked): here only once it can play.
        var shown = chatViewable();
        if (!force && shown === lastShown) { return; }
        lastShown = shown;
        fetch("/api-sdk/gopod_page_shown", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "shown=" + (shown ? "1" : "0") + "&page=" + encodeURIComponent(pageId) + "&visitor=" + encodeURIComponent(visitorId) + (pageViewOnly || /[?&]visitor(=|&|$)/.test(location.search) ? "&viewer=1" : "")
        }).then(function (r) { return r.ok ? r.json() : null; })
          .then(function (d) {
            if (!d || typeof d.main !== "boolean") { return; }
            isAdminPage = d.admin === true;
            // A visitor's typing row is two lines (2026-10-24, operator: "remove
            // 3rd line: | Record? [] | 00:25:46 | 🔇"): the Brobots, then the text.
            if (document.body.classList.contains("gopod-visitor") === isAdminPage) {
              document.body.classList.toggle("gopod-visitor", !isAdminPage);
              fitTypedRow();
            }
            if (isAdminPage) { document.body.classList.remove("gopod-chat-only"); }
            if (chatOnlyPending) { chatOnlyPending = false; setTimeout(chatOnlyOn, 0); }
            setPageRole(d.main);
            if (typeof d.played === "number" && d.boot === playedBoot) { mainPlayed = d.played; runFeedQueue(); }
          })
          .catch(function () { /* Wire-Pod not up yet */ });
      }
      document.addEventListener("visibilitychange", function () { reportShown(false); runFeedQueue(); });
      // A viewer reports in twice a second, to hear how far the main page is.
      setInterval(function () { reportShown(pageRoleKnown && !isMainPage); }, 500);
      // 2026-10-17: every 5 s, so Wire-Pod knows the page is here soon after a
      // restart (gopod_page_presence.go).
      setInterval(function () { reportShown(true); }, 5000);
      // 2026-10-12: a line's voice is loaded while its bubble is placed, and
      // plays once the bubble shows.
      var nextBubbleVoice = null, voiceCache = {};
      // 2026-10-14 (operator: "Rule: Please make all chat bubble audio default
      // to volume variable to play at 10% volume ... all chat bubbles, system
      // bubbles, reporter/narrator bubbles, brobots' chat bubbles, etc... song
      // sequences as well"): every sound this page plays for a bubble or a song
      // plays at chat_audio.volume; a sound's own volume ("#vol=", a sting's or a
      // song step's) is its share of it. The background loop keeps its own.
      var chatVolume = 0.35;
      function chatLevel(share) {
        // A viewer page is silent (C1): the main page plays the sound.
        if (!isMainPage) { return 0; }
        var v = typeof share === "number" && !isNaN(share) ? share : 1;
        return Math.max(0, Math.min(1, chatVolume * v));
      }
      // 2026-10-17 (operator: "One master default knobs setting all other knobs
      // refers to * other story.md knobs are either \"\" for default or
      // \"0-100%\""): a volume knob as a share of chat_audio.volume.
      function volumeShare(v) {
        if (typeof v === "number" && !isNaN(v)) { return Math.max(0, v); }
        var t = String(v == null ? "" : v).trim();
        if (!t || t.toLowerCase() === "default") { return 1; }
        var n = parseFloat(t);
        if (isNaN(n) || n < 0) { return 1; }
        return /%$/.test(t) ? n / 100 : n;
      }
      // 2026-10-17 (operator: "normalize all volumes to -3db ... before
      // \"default\" knob setting"): each sound's offset to chat_audio.normalize_db
      // (Wire-Pod measures them: gopod_audio_levels.json), by file name.
      var audioLevels = {};
      // 2026-10-17 (operator: "Can you do what I do"): a sound's evened copy
      // (gopod_audio_levels.json "made": its page path -> the copy's), played
      // instead of it; made voices are evened in place.
      var audioMade = {};
      function loadAudioLevels() {
        fetch("/gopod_audio_levels.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (d) {
            if (!d) { return; }
            audioLevels = d.files || {};
            var made = d.made || {};
            if (JSON.stringify(made) !== JSON.stringify(audioMade)) {
              audioMade = made;
              // Sounds loaded before the list came play their evened copy next time.
              // Stings are loaded again at once (their evened copy), ready to play.
              Object.keys(stingPool).forEach(function (k) { if (audioMade[k]) { delete stingPool[k]; preloadStings([k]); } });
              Object.keys(voiceCache).forEach(function (k) { if (audioMade[k.replace(/#.*$/, "")]) { delete voiceCache[k]; } });
            }
          })
          .catch(function () { /* none yet */ });
      }
      // The page path of a sound's evened copy, or the sound itself.
      function normSrc(src) {
        var s = String(src || ""), hash = "";
        var at = s.indexOf("#");
        if (at >= 0) { hash = s.slice(at); s = s.slice(0, at); }
        var path = s.replace(/^https?:\/\/[^/]+/, "");
        if (path.charAt(0) !== "/") { path = "/" + path; }
        return audioMade[path] ? audioMade[path] + hash : String(src);
      }
      loadAudioLevels();
      setInterval(loadAudioLevels, 30000);
      function audioLevel(src) {
        var name = String(src || "").replace(/#.*$/, "").split("/").pop();
        var v = audioLevels[name];
        return typeof v === "number" && v > 0 ? v : 1;
      }
      var VOICE_READY_MAX_MS = 3000;
      function voiceAudio(src) {
        var key = String(src);
        if (!voiceCache[key]) {
          var vol = /#vol=([0-9.]+)$/.exec(key);
          var a = new Audio(normSrc(key.replace(/#vol=[0-9.]+$/, "")));
          a._share = vol ? parseFloat(vol[1]) : 1;
          a.preload = "auto";
          a.addEventListener("error", function () { a._failed = true; });
          try { a.load(); } catch (err) { a._failed = true; }
          // 2026-10-18 (forced alignment): a made voice's word times, beside it.
          var plain = key.replace(/#.*$/, "");
          a._wordsDone = true;
          if (/\/gopod_reporter_audio\/[a-z]{2}_[a-z]+_(s\d+_)?[0-9a-f]{16}\.wav$/.test(plain)) {
            // 2026-10-22 (live: PLAYHEAD's fresh weather line lit out of step):
            // the voice waits for these (playWhenShown) - a just-made voice
            // started before they came in, and lit by the estimate.
            a._wordsDone = false;
            fetch(plain.replace(/\.wav$/, ".words.json"), { cache: "no-store" })
              .then(function (r) { return r.ok ? r.json() : null; })
              .then(function (d) { if (d && d.words && d.words.length) { a._words = d.words; } })
              .catch(function () { /* none: the timed estimate */ })
              .then(function () { a._wordsDone = true; });
          }
          voiceCache[key] = a;
        }
        return voiceCache[key];
      }
      // The system stings (reporter_voices.system_stings, the key sound): one
      // loaded copy each, kept and replayed; a second copy only while one plays.
      var stingFiles = {}, stingPool = {};
      function stingFromPool(src) {
        var key = String(src).replace(/#vol=[0-9.]+$/, "");
        var name = key.split("/").pop();
        if (!stingFiles[name]) { return null; }
        var vol = /#vol=([0-9.]+)$/.exec(String(src));
        var a = stingPool[key];
        if (!a || a._failed || !a.paused) {
          a = new Audio(normSrc(key));
          a.preload = "auto";
          a.addEventListener("error", function () { a._failed = true; });
          try { a.load(); } catch (err) { a._failed = true; }
          if (!stingPool[key] || stingPool[key]._failed) { stingPool[key] = a; }
        } else {
          try { a.currentTime = 0; } catch (err) { /* not loaded yet */ }
        }
        a._share = vol ? parseFloat(vol[1]) : 1;
        return a;
      }
      function preloadStings(files) {
        files.forEach(function (path) {
          var name = String(path).split("/").pop();
          if (!name) { return; }
          stingFiles[name] = true;
          var key = String(path).indexOf("/") === 0 ? String(path) : "/gopod_reporter_audio/" + name;
          if (!stingPool[key]) {
            var a = new Audio(normSrc(key));
            a.preload = "auto";
            try { a.load(); } catch (err) { /* later */ }
            stingPool[key] = a;
          }
        });
      }
      function isVoiceEvent(e) {
        return e && e.who === reporterRole && /\.wav(#vol=[0-9.]+)?$/.test(String(e.text)) &&
          String(e.text).trim() !== openingEndText && !(heardMarker && String(e.text).indexOf(heardMarker) === 0);
      }
      // 2026-10-17 (survey E3: "heard" meant passed, not played): voices that
      // didn't play (missing, blocked) since the last heard marker - told to
      // Wire-Pod with that marker, so the journal says so.
      var voicesSkipped = 0;
      function playVoiceThen(src, done) {
        // A viewer catching up passes the voice: its line shown said, at once.
        if (viewerBehind()) {
          delete voiceCache[String(src)];
          var passed = voiceLitStart(0.1, null, src);
          if (passed) { passed(); }
          setTimeout(done, 0);
          return;
        }
        var finished = false, litDone = null;
        var finish = function () {
          if (finished) { return; }
          finished = true;
          if (voicePlaying && voicePlaying.finish === finish) { voicePlaying = null; }
          // 2026-10-18 (leak: a voice that errored or hit its cap left its line
          // "being said", and the baton held every next bubble up to 2 min):
          // however a voice ends, its line is said.
          if (litDone) { litDone(); litDone = null; }
          setTimeout(done, voiceGapMs);
        };
        var skip = function () {
          if (!finished) { voicesSkipped++; }
          finish();
        };
        try {
          // 2026-10-11: a song sound's volume rides on its path ("#vol=0.8").
          var audio = voiceAudio(src);
          var voiceFeedId = feedIdNow;
          delete voiceCache[String(src)];
          // 2026-10-17 (operator: "If system audio can be preloaded to play
          // multiple times, preload and play as needed * grunge, djent"): a sting
          // is loaded once and played again from the start each time.
          var pooled = stingFromPool(src);
          if (pooled) { audio = pooled; }
          audio.addEventListener("ended", finish, { once: true });
          audio.addEventListener("error", skip, { once: true });
          voicePlaying = { audio: audio, finish: finish };
          var from = Date.now();
          (function playWhenShown() {
            var b = audio._bubble;
            // 2026-10-16 (comb-through: a voice could play before its bubble
            // showed - its link lost, or the 20 s wait shorter than the baton's):
            // a voice waits while its bubble, or any bubble before it, is still
            // waiting to show - as long as the baton can be held.
            // 2026-10-23 (live: the interview's end held 4.5 min): only bubbles that
            // came BEFORE this voice hold it - a later one (the grand prize) waits
            // for this line to be said, so it must never hold the voice back.
            var waiting = (b && b.isConnected && b.style.visibility === "hidden") ||
              pendingReveal.some(function (p) { return !(voiceFeedId && p._feedId > voiceFeedId); });
            if (waiting && Date.now() - from < CHAT_BATON_MAX_MS + BUBBLE_READY_MAX_MS) {
              setTimeout(playWhenShown, 50);
              return;
            }
            // 2026-10-22: GOLDEN voice start - shown, then its length and word
            // times in (up to VOICE_READY_MAX_MS), then it plays and lights.
            if (!audio._readyFrom) { audio._readyFrom = Date.now(); }
            if ((!audio._wordsDone || !(audio.duration > 0)) && !audio._failed && Date.now() - audio._readyFrom < VOICE_READY_MAX_MS) {
              setTimeout(playWhenShown, 50);
              return;
            }
            audio._readyFrom = 0;
            audio.volume = chatLevel(audio._share * audioLevel(audio.src));
            // The Operator line's voice: its words come back as it plays (voiceLitStart).
            // 2026-10-17 (survey E2: a 60 s cap, set before the up-to-126 s wait
            // for its bubble, let the queue move on while this voice could still
            // start later, out of order): the cap starts when it plays - its own
            // length plus 5 s (60 s when the length isn't known), so a line is
            // never cut and never stuck.
            var capMs = isFinite(audio.duration) && audio.duration > 0 ? audio.duration * 1000 + 5000 : 60000;
            setTimeout(finish, capMs);
            // A reporter's or Goverlord's line lights with its voice (voiceLitStart).
            if (!pooled && (voiceLitNext || voiceWaitFor[String(src).replace(/#.*$/, "")])) { litDone = voiceLitStart(audio.duration, audio._words, src); }
            audio.play().catch(skip);
          })();
        } catch (err) { skip(); }
      }
      // 2026-10-20 (operator: "every time I do a wpr 2 ... I have to wait for the
      // reporters to finish their spoken lines * how to add keyboard 'Esc' key
      // press to skip to the end of the in-between run"): Esc (or system-skip)
      // stops the reporter line being said and drops the reporters' queued
      // lines and voices, up to the next line that isn't theirs (Goverlord's
      // "ready", a song's start) or the opening's end. Heard markers still pass,
      // so Wire-Pod and the songs go on as if they'd played. The managers
      // (CHALK, PLAYHEAD) and the Brobots are never skipped.
      var skippingReporters = false, voicePlaying = null;
      // 2026-10-22 (operator: "skip is set to disabled * only upon voices ready
      // can skip be enabled * upon skipped, 'skip' skip is set to disabled"):
      // the skip gate (skip_gate) - off until a line starting with one of
      // enable_on shows ("Voices ready."), off again once used.
      var skipGate = { enable_on: [], on: false };
      function skipGateCheck(message) {
        if (restoringChat) { return; }
        var m = String(message || "").trim();
        (skipGate.enable_on || []).forEach(function (w) {
          if (w && m.indexOf(w) === 0) { skipGate.on = true; }
        });
      }
      function skipReporters(how) {
        if (!skipGate.on) {
          fetch("/api-sdk/gopod_page_note", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: "note=" + encodeURIComponent("skip ignored (" + (how || "Esc") + "): not enabled") }).catch(function () {});
          return;
        }
        skipGate.on = false;
        // 2026-10-21 (operator: "add to logging * skipped [input] (Esc, numpad, mic, typed)").
        fetch("/api-sdk/gopod_page_note", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "note=" + encodeURIComponent("skipped (" + (how || "Esc") + ")")
        }).catch(function () { /* Wire-Pod not up */ });
        skippingReporters = true;
        if (voicePlaying) { try { voicePlaying.audio.pause(); } catch (err) { /* gone */ } voicePlaying.finish(); }
      }
      function skippedEvent(e) {
        if (!skippingReporters) { return false; }
        var t = String(e.text);
        if (e.who === reporterRole) {
          if (t.indexOf("skip:") === 0) { return false; }
          if (heardMarker && t.indexOf(heardMarker) === 0) { return false; }
          if (t.trim() === openingEndText) { skippingReporters = false; skipEndScroll(); return false; }
          return /\.wav(#vol=[0-9.]+)?$/.test(t); // a voice (or sting) in the stretch
        }
        // 2026-10-20 (operator: "upon 'Esc' or skip * display all bubbles (in
        // the opener it'd be showing all the reporter lines) first, then llm
        // ready"): a skipped reporter line still shows, all its words at once.
        if (hostNames[e.who] && !managerHosts[e.who]) { return false; }
        skippingReporters = false;
        skipEndScroll();
        return false;
      }
      // 2026-10-22 (operator: "upon voices ready skips * update template so chat
      // bubble also scrolls to bottom position"): a skip's end takes the chat to
      // its bottom, once the shown bubbles are placed.
      function skipEndScroll() {
        [300, 1200].forEach(function (ms) { setTimeout(function () { scrollChatTo(chatMaxScroll()); }, ms); });
      }
      var feedIdNow = 0; // the chat event being shown (a Brobot bubble keeps it: data-feed)
      function showFeedEvent(e, done) {
        feedIdNow = e.id || 0;
        // A "skip" said into the mic (Wire-Pod's skip marker): logged there already.
        if (e.who === reporterRole && String(e.text).indexOf("skip:") === 0) {
          // The mic's "skip" (Wire-Pod's skip: line) goes through the skip gate too.
          if (!skipGate.on) { done(); return; }
          skipGate.on = false;
          skippingReporters = true; if (voicePlaying) { try { voicePlaying.audio.pause(); } catch (err) { /* gone */ } voicePlaying.finish(); } done(); return; }
        if (skippedEvent(e)) { done(); return; }
        // A held rattle timer gives its place to the next bubble (voices aren't bubbles).
        if (countdownDoneAt && e.who !== reporterRole && e.who !== countdownRole && e.who !== songAudioRole) {
          afterCountdown(function () { showFeedEvent(e, done); }, e.who === ballRole);
          return;
        }
        if (e.who === songAudioRole) {
          // 2026-10-15 (operator: "Stop run didn't work during song run"): a
          // stopped timed song - its soundtrack stops, narrator bubbles not yet
          // shown go, the ones on screen show their words and their pictures fade.
          if (String(e.text).trim() === "stop") {
            if (songAudio) { songAudio.pause(); songAudio = null; }
            narratorLines.forEach(function (n) {
              if (!n.shown) {
                if (n.part && n.placed) { n.part.remove(); } else { n.bubble.remove(); }
              } else {
                revealRest(n.part || n.bubble);
                if (n.bg) { n.bg.style.transitionDuration = "1s"; n.bg.classList.remove("on"); }
              }
            });
            narratorLines = [];
            done();
            return;
          }
          try {
            if (songAudio) { songAudio.pause(); }
            songAudio = new Audio(normSrc(String(e.text)));
            songAudio.volume = chatLevel(audioLevel(e.text));
            songAudio.play().catch(function () { /* the page hasn't been clicked yet */ });
          } catch (err) { /* no audio */ }
          done();
          return;
        }
        if (e.who === narratorRole) {
          showNarratorLine(e.text);
          done();
          return;
        }
        if (e.who === reporterRole) {
          // GOLDEN, locked (2026-10-13): the System llm bubble's message, in the
          // chat's order; a step's message stays up warmup_check_seconds before
          // the next line, so every step is seen.
          if (String(e.text).trim() === "loop_pause" || String(e.text).trim() === "loop_resume") {
            backgroundLoop.runQuiet(String(e.text).trim() === "loop_pause");
            done();
            return;
          }
          if (String(e.text).indexOf("warmup_text|") === 0) {
            // 2026-10-15 (operator: '"LLM Warming Up. Please Wait." play after
            // its displayed * bubble should be ready to play upon being displayed
            // as per rule'): its voice (the sting and the words right behind it)
            // is loaded first, then the message shows and the voice plays at once.
            var warmVoices = [];
            for (var q = 0; q < feedQueue.length && isVoiceEvent(feedQueue[q].e); q++) { warmVoices.push(voiceAudio(feedQueue[q].e.text)); }
            var warmFrom = performance.now();
            (function whenWarmVoiceReady() {
              var loading = warmVoices.some(function (a) { return !a._failed && a.readyState < 3; });
              if (loading && performance.now() - warmFrom < BUBBLE_READY_MAX_MS) { setTimeout(whenWarmVoiceReady, 50); return; }
              var held = warmupTextLine(e.text);
              // Its spoken line right behind it plays now, as the hold.
              setTimeout(done, held && !warmVoices.length ? llmWarm.checkMs : 0);
            })();
            return;
          }
          if (generatingMarker && String(e.text).indexOf(generatingMarker) === 0) {
            var genKind = String(e.text).slice(generatingMarker.length);
            // GOLDEN: a voice whose words are shown some other way (gopodSayForShownLine).
            var forKind = genKind.indexOf("forwait:") === 0 ? "forwait:" : (genKind.indexOf("forlit:") === 0 ? "forlit:" : "");
            if (forKind) {
              var forRest = genKind.slice(forKind.length), forBar = forRest.indexOf("|");
              if (forBar > 0) { voiceForLine(forRest.slice(0, forBar), forRest.slice(forBar + 1), forKind === "forwait:"); }
              done();
              return;
            }
            if (genKind === "off") { voiceWaitsGone(); }
            generatingGone();
            done();
            return;
          }
          if (String(e.text).indexOf("warmup_ready|") === 0) {
            warmupReadyLine(e.text);
            done();
            return;
          }
          if (String(e.text).trim() === openingEndText) {
            openingDone = true;
            warmupBeforeOpening = null;
            var openingBody = new URLSearchParams();
            openingBody.set("page", pageId);
            fetch("/api-sdk/gopod_opening_heard", { method: "POST", body: openingBody }).catch(function () { /* Wire-Pod gone */ });
            done();
            return;
          }
          // GOLDEN (gopod_say_heard.go): the voice lines before this marker have
          // played - Wire-Pod's song waits for this before it goes on.
          if (heardMarker && String(e.text).indexOf(heardMarker) === 0) {
            var heardBody = new URLSearchParams();
            heardBody.set("token", String(e.text).slice(heardMarker.length));
            heardBody.set("skipped", String(voicesSkipped));
            heardBody.set("page", pageId);
            voicesSkipped = 0;
            fetch("/api-sdk/gopod_voice_heard", { method: "POST", body: heardBody }).catch(function () { /* Wire-Pod gone */ });
            done();
            return;
          }
          playVoiceThen(e.text, done);
          return;
        }
        // GOLDEN global message: a line whose voice is being made ("wait:") shows
        // in its own bubble with "Generating..." and the hourglass (voiceLitLine).
        lineWaits = false;
        lineVoice = "";
        // "lit:<voice file>|text" (2026-10-21): a made voice's line, tied to its file.
        if (generatingMarker && String(e.text).indexOf(generatingMarker + "lit:") === 0) {
          e.text = String(e.text).slice((generatingMarker + "lit:").length);
          var litBar = e.text.indexOf("|");
          if (litBar > 0 && e.text.charAt(0) === "/") { lineVoice = e.text.slice(0, litBar); e.text = e.text.slice(litBar + 1); }
        }
        if (generatingMarker && String(e.text).indexOf(generatingMarker + "wait:") === 0) {
          e.text = String(e.text).slice((generatingMarker + "wait:").length);
          lineWaits = true;
          // "wait:<voice file>|text": this line's own voice (2026-10-21).
          var bar = e.text.indexOf("|");
          if (bar > 0 && e.text.charAt(0) === "/") { lineWaits = e.text.slice(0, bar); e.text = e.text.slice(bar + 1); }
        }
        if (e.who !== systemRole && e.who !== scoreRole && !hostNames[e.who]) { generatingGone(); }
        if (e.who === countdownRole) {
          if (String(e.text).trim() === countdownEndText) { countdownHold(); }
          else { showCountdown(parseInt(e.text, 10), e.time_ms); }
        }
        else if (e.who === ballRole) { showBingoBall(e.text); }
        else if (e.who === chapterRole) { showChapter(e.text); }
        else if (e.who === scoreRole) { showScore(e.text, undefined, done); return; }
        // 2026-10-13 (operator: "all system messages out = \"Goverlord: GOPOD
        // Brain\" * \"Reporter 0\" is a decoration label"): one header for them all.
        else if (e.who === systemRole) { showScore(e.text, undefined, done); return; }
        else if (e.who === terminalRole) { showTerminal(e.text, done); return; }
        else if (searchSettings.results_role && e.who === searchSettings.results_role) { showSearchResults(e.text, e.to); }
        else if (e.who === pictureRole && String(e.text).indexOf("notice:") === 0) { bubbleNotice(e.to, String(e.text).slice(7)); }
        // 2026-10-23 (operator: "'Thanks for your help, PLAYHEAD.' not in highlighting
        // pipeline"): a notice line he says - lit while he says it, said after.
        else if (e.who === pictureRole && String(e.text).indexOf("noticesay:") === 0) { bubbleNotice(e.to, String(e.text).slice(10), true); }
        else if (e.who === pictureRole && String(e.text).indexOf("noticesaid:") === 0) { bubbleNoticeSaid(e.to); }
        else if (e.who === pictureRole && String(e.text).indexOf("weathershow:") === 0) { weatherShowFor(e.to); }
        else if (e.who === pictureRole) { if (!menuPictures(e.text)) { showPictures(String(e.text).replace(/^menu:/, "")); } }
        else if (hostNames[e.who]) { showHostLine(e.who, e.text, e.to, done); return; }
        // 2026-10-08: a draft's own question shows as the reporter who reads it.
        // A draft question its reporter reads: an Operator bubble, "Reporter 1 to
        // Brobot 1" (operator 2026-10-10: "assigned reporter bubbles fall under
        // 'Operator' = white eyed icon to brobot(s)").
        else if (e.who === "Operator" && e.by && hostNames[e.by]) {
          appendBubble("Operator", e.text, undefined, e.to, hostHeader(e.by));
        }
        else if (e.who === "Operator") { appendBubble("Operator", e.text, undefined, e.to, e.via); }
        // 2026-10-17 (operator: "one event list, single source of truth"): a live
        // reply comes on the feed (via "llm"), in order with every other line.
        else if (e.via === "llm") { noteLiveReply(e, robotBubbleOnce(e.who, kgReplyText(e.text), "llm", done)); return; }
        else { robotBubbleOnce(e.who, e.text, "brobot", done); return; }
        done();
      }

      // 2026-09-01: bubble background follows each robot's own real eye
      // color instead of a fixed pair - _write_brobot_eye_colors_state()
      // (run_section1_full_live_001.py) writes this file fresh once per
      // run start (Robots.__init__), so a color change on the robot's own
      // app is picked up by the NEXT run, never mid-run (operator
      // direction). A small JSON file polled on a 500ms cadence - reapplying an unchanged value
      // every tick is harmless.
      var lastEyeColors = null;
      function pollEyeColors() {
        fetch("/gopod_brobot_eye_colors.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (colors) {
            if (!colors) { return; }
            var serialized = JSON.stringify(colors);
            if (serialized === lastEyeColors) { return; }
            lastEyeColors = serialized;
            var root = document.documentElement.style;
            if (colors.brobot_1) { root.setProperty("--brobot-1-color", colors.brobot_1); }
            if (colors.brobot_2) { root.setProperty("--brobot-2-color", colors.brobot_2); }
          })
          .catch(function () { /* color file not written yet - CSS fallback holds */ });
      }

      // 2026-09-29: light up the sentence a robot is saying right now. Wire-Pod writes
      // the sentence it is about to say to /gopod_speaking_now.json (active=false when
      // the reply ends; interrupted=true on a touch interrupt); the sentence in that robot's newest bubble that shares
      // the most words with it gets the "speaking" class (solid white). Words are compared, not
      // exact text, because the spoken copy is the pronounced one ("Gowp-awd", no emoji).
      var ESN_TO_BUBBLE_CLASS = { "0dd1b9e9": "brobot-1", "0dd1d8bf": "brobot-2" };
      // 2026-10-15 (operator: "revealed as it is heard, not ahead of robot
      // speech"): a piece shows when it's lit (.speaking) and stays at 85%
      // (.spoken); the pieces before it count as heard too (a missed highlight
      // never leaves a hole). CSS hides the rest (gopod_chat_bubbles.css).
      function lightSpan(span) {
        span.classList.add("speaking");
        span._litAt = Date.now();
        // 2026-10-16 (comb-through: plain bubbles flickered): the word reveal is
        // for the bubbles shown piece by piece (live, scripted, narrator) only.
        if (span.closest(".brobot-bubble-live, .brobot-bubble-scripted, .narrator, .voice-lit")) { digitSequence(span); }
        var bubble = span.closest(".brobot-bubble");
        if (!bubble) { return; }
        var all = span.closest(".nested-host") ? span.closest(".nested-host").querySelectorAll(".brobot-say") : ownSays(bubble);
        for (var i = 0; i < all.length && all[i] !== span; i++) { all[i].classList.add("spoken"); }
      }
      // 2026-10-15 (operator: "as 142857 is spoken as a sequence of single
      // numbers, each number highlights sequentially * not precisely timed with
      // audio, but give me your best guess"; "Apply same new sequential
      // highlighting to military time, so with 0904 each number highlights
      // sequentially"): in the piece being said, a run of single digits ("1 4 2 8
      // 5 7") and a military time ("0904 Hours Military Time", in its spoken
      // parts: 0 9 0 4; 18 14; 18 00) light one number at a time - the words
      // before them at digit_highlight.seconds_per_word, then one number every
      // seconds_per_number. Like every piece: hidden until said, solid white
      // while said, 85% after.
      var digitHighlight = { seconds_per_word: 0.3, seconds_per_number: 0.45 };
      var militaryLabel = "Hours Military Time";
      // GOLDEN, word at a time (operator 2026-10-16: "extend that same process to
      // a sentence a word at a time * sentence is displayed at 0% solid * then use
      // timed, word-at-a-time, reveals to kinda match brobot speech timing"; his
      // rule, from "as of 1918 Hours Military Time or 7:18 PM.": the piece starts
      // at 0%, then "as of", then 1918 (19, 18), then "Hours Military Time", then
      // "or", then "7:18 PM" - the one being said 100% solid white, the ones said
      // 85%, as every piece). In the piece being said every word is its own step,
      // at word_reveal.seconds_per_word; a run of single digits one digit at a
      // time, a military time one spoken part at a time ("13", "16") and a clock
      // time as said ("1", ":16", "P", "M") at seconds_per_number; the label's
      // words one each ("Hours", "Military", "Time."); a math formula is one step. A best guess, not tied
      // to the audio; when the piece ends the rest shows (85%).
      var wordReveal = { seconds_per_word: 0.38 };
      function digitUnits(span) {
        if (span._digitUnits) { return span._digitUnits; }
        var units = [];
        var label = militaryLabel.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        var military = new RegExp("\\b([01]\\d|2[0-3])([0-5]\\d)(?=\\s+" + label + ")", "gi");
        var walker = document.createTreeWalker(span, NodeFilter.SHOW_TEXT, null);
        var nodes = [];
        while (walker.nextNode()) { nodes.push(walker.currentNode); }
        var formulas = [];
        nodes.forEach(function (node) {
          var parent = node.parentNode;
          var formula = parent && parent.closest ? parent.closest(".katex") : null;
          if (formula) {
            if (formulas.indexOf(formula) < 0) {
              formulas.push(formula);
              formula.classList.add("seq-unit");
              formula._seqWeight = digitHighlight.seconds_per_number * 2;
              formula._seqAt = node;
              units.push(formula);
            }
            return;
          }
          var text = node.nodeValue, cuts = [];
          var m, re = /\b\d(?: \d)+\b/g;
          while ((m = re.exec(text))) {
            for (var k = 0; k < m[0].length; k += 2) { cuts.push([m.index + k, m.index + k + 1, digitHighlight.seconds_per_number]); }
          }
          while ((m = military.exec(text))) {
            var h = parseInt(m[1], 10), mi = parseInt(m[2], 10), at = m.index, n = digitHighlight.seconds_per_number;
            if (h === 0) { cuts.push([at, at + 2, n]); }
            else if (h < 10) { cuts.push([at, at + 1, n]); cuts.push([at + 1, at + 2, n]); }
            else { cuts.push([at, at + 2, n]); }
            if (mi === 0 || mi >= 10) { cuts.push([at + 2, at + 4, n]); }
            else { cuts.push([at + 2, at + 3, n]); cuts.push([at + 3, at + 4, n]); }
            // 2026-10-16 (operator: "1316 Hours Military Time ... globally
            // reveal as per: 13 / 16 / Hours / Military / Time."): the label's
            // words follow, one step each, like every word.
          }
          // A clock time (operator 2026-10-17: "reveal: clock as \"1\" \":\"
          // \"43\" in succession"): its hour, its colon, its minutes, then each
          // letter of AM/PM.
          var clock = /\b(\d{1,2})(:\d{2})(?:(\s*)([AaPp])(\.?)([Mm]))?/g;
          while ((m = clock.exec(text))) {
            var c0 = m.index, n2 = digitHighlight.seconds_per_number, h1 = c0 + m[1].length;
            cuts.push([c0, h1, n2]);
            cuts.push([h1, h1 + 1, wordReveal.seconds_per_mark === undefined ? 0.15 : wordReveal.seconds_per_mark]);
            cuts.push([h1 + 1, h1 + m[2].length, n2]);
            if (m[4]) {
              var p0 = c0 + m[1].length + m[2].length + m[3].length;
              cuts.push([p0, p0 + 1, wordReveal.seconds_per_word]);
              var q0 = p0 + 1 + m[5].length;
              cuts.push([q0, c0 + m[0].length, wordReveal.seconds_per_word]);
            }
          }
          // A year after its date (operator 2026-10-17: "say year similar as 20
          // 26"; "October / 17 / , / 20 / 26"): its two halves, one step each.
          var year = /\b(?:January|February|March|April|May|June|July|August|September|October|November|December) \d{1,2}, (\d{2})(\d{2})\b/g;
          while ((m = year.exec(text))) {
            var y0 = m.index + m[0].length - 4;
            cuts.push([y0, y0 + 2, wordReveal.seconds_per_word]);
            cuts.push([y0 + 2, y0 + 4, wordReveal.seconds_per_word]);
          }
          cuts.sort(function (a, b) { return a[0] - b[0]; });
          // Every other word (a run of non-spaces) is a step of its own.
          var all = [], last = 0;
          cuts.forEach(function (c) {
            if (c[0] < last) { return; }
            all.push.apply(all, wordCuts(text, last, c[0]));
            all.push(c);
            last = c[1];
          });
          all.push.apply(all, wordCuts(text, last, text.length));
          if (!all.length) { return; }
          var frag = document.createDocumentFragment();
          last = 0;
          all.forEach(function (c) {
            frag.appendChild(document.createTextNode(text.slice(last, c[0])));
            var u = document.createElement("span");
            u.className = "seq-unit";
            u.textContent = text.slice(c[0], c[1]);
            u._seqWeight = c[2];
            frag.appendChild(u);
            units.push(u);
            last = c[1];
          });
          frag.appendChild(document.createTextNode(text.slice(last)));
          node.parentNode.replaceChild(frag, node);
        });
        // A country flag (its picture) is a step of its own, said as its country.
        Array.prototype.forEach.call(span.querySelectorAll("img.flag-emoji, img.gopod-qr2"), function (img) {
          img.classList.add("seq-unit");
          img._seqWeight = wordReveal.seconds_per_word;
          units.push(img);
        });
        units.sort(function (a, b) {
          var pa = a._seqAt || a, pb = b._seqAt || b;
          return pa === pb ? 0 : (pa.compareDocumentPosition(pb) & Node.DOCUMENT_POSITION_FOLLOWING ? -1 : 1);
        });
        span._digitUnits = units;
        return units;
      }
      // The words between from and to: [start, end, seconds] each. 2026-10-17
      // (operator: "Split away punctuation from words * , ... : - ! ? etc *
      // reveal the word, then the punctuation, then emojis"): a word's marks
      // before and after it are steps of their own ("Hey" then ","; "me" then
      // "..."; "11" then "°"), word_reveal.seconds_per_mark each; marks inside a
      // word stay with it (Let's, chalk-o-latteric, crushn8r.ca).
      var MARKS = ".,:;!?\u2026'\"\u201c\u201d\u2018\u2019()\\[\\]{}\\-\u2013\u2014/*";
      var markRunRe = new RegExp("[" + MARKS + "]+|[^" + MARKS + "]+", "g");
      var formatOnlyRe = /^[\u200B-\u200F\u2060-\u2064\uFE0E\uFE0F]+$/;
      function wordCuts(text, from, to) {
        var out = [], re = /\S+/g, m;
        re.lastIndex = from;
        while ((m = re.exec(text)) && m.index < to) {
          var end = Math.min(m.index + m[0].length, to);
          var run = text.slice(m.index, end);
          var first = run.search(/[A-Za-z0-9]/);
          if (first < 0) { markCuts(text, m.index, end, out); continue; }
          var last = run.length - 1;
          while (!/[A-Za-z0-9]/.test(run.charAt(last))) { last--; }
          markCuts(text, m.index, m.index + first, out);
          out.push([m.index + first, m.index + last + 1, wordReveal.seconds_per_word]);
          markCuts(text, m.index + last + 1, end, out);
        }
        return out;
      }
      // Marks (punctuation) and symbols (emoji, °) between from and to: one
      // step per run of the same kind.
      function markCuts(text, from, to, out) {
        if (from >= to) { return; }
        var part = text.slice(from, to), m, mark = wordReveal.seconds_per_mark;
        if (mark === undefined) { mark = 0.15; }
        markRunRe.lastIndex = 0;
        while ((m = markRunRe.exec(part))) {
          if (formatOnlyRe.test(m[0])) { continue; }
          out.push([from + m.index, from + m.index + m[0].length, mark]);
        }
      }
      function digitSequence(span) {
        var units = digitUnits(span);
        if (!units.length || span._digitRunning) { return; }
        span._digitRunning = true;
        // 2026-10-16 (operator: "it appears to start a moment before the brobot
        // speaks * add a short delay about the length of 2 timed words"):
        // word_reveal.start_delay_words.
        var at = (wordReveal.start_delay_words || 0) * (wordReveal.seconds_per_word || 0);
        // 2026-10-16 (operator: "Baby brobots song is timed to an srt file * please
        // ensure the new per-word-reveal in highlighting, matches srt end to end *
        // scale per-word timed reveals to fit into that spoken segment"): a line
        // with its own SRT times (data-at / data-end, a narrator's) starts at once
        // and its steps are scaled to fill it exactly, first word to last.
        var srtAt = parseFloat(span.dataset.at), srtEnd = parseFloat(span.dataset.end);
        var scale = 1;
        if (!isNaN(srtAt) && !isNaN(srtEnd) && srtEnd > srtAt) {
          var total = 0;
          units.forEach(function (u) { total += u._seqWeight || 0; });
          if (total > 0) { scale = (srtEnd - srtAt) / total; }
          at = 0;
        }
        // 2026-10-18 (operator: "Forced alignment * word for word * punctuation
        // highlights at a silent point threshold, after the ending word"): a
        // voice with its word times (the voice worker's .words.json) reveals
        // each word when it's said, and its marks when that word ends.
        var aligned = span._spokenWords ? alignUnits(units, span._spokenWords) : null;
        units.forEach(function (u, i) {
          setTimeout(function () {
            if (!span.classList.contains("speaking")) { return; }
            if (i > 0) { units[i - 1].classList.remove("now"); units[i - 1].classList.add("said"); }
            u.classList.add("now");
          }, (aligned ? aligned[i] : at) * 1000);
          at += (u._seqWeight || 0) * scale;
        });
        if (aligned) { at = Math.max(at, aligned[aligned.length - 1] || 0); }
        setTimeout(function () {
          units.forEach(function (u) { u.classList.remove("now"); u.classList.add("said"); });
        }, at * 1000);
      }
      // Each unit's second in the voice: a word at the start of the spoken word
      // it matches (a part of one, like "20" and "26" of "2026", at its share
      // of it); a mark or emoji when the word before it ends (the silence
      // after it); a word not found (said another way: "crushin-ate-a") between
      // its neighbours.
      function alignUnits(units, words) {
        var norm = function (t) { return String(t || "").toLowerCase().replace(/[^a-z0-9]+/g, ""); };
        var toks = words.map(function (w) { return { n: norm(w.t), s: +w.s || 0, e: +w.e || 0, used: 0 }; });
        var times = [], at = 0, lastEnd = 0;
        units.forEach(function (u) {
          var n = norm(u.tagName === "IMG" ? "" : u.textContent);
          if (!n) { times.push(lastEnd > 0 ? lastEnd : null); return; }
          var hit = null, part = 0;
          for (var j = at; j < Math.min(at + 6, toks.length); j++) {
            var t = toks[j];
            if (!t.n) { continue; }
            var rest = t.n.slice(t.used);
            if (rest === n || rest.indexOf(n) === 0 || (t.used === 0 && n.indexOf(t.n) === 0)) {
              hit = j; part = t.used / t.n.length;
              t.used = Math.min(t.n.length, t.used + n.length);
              break;
            }
          }
          if (hit === null) { times.push(null); return; }
          var tk = toks[hit];
          times.push(tk.s + (tk.e - tk.s) * part);
          lastEnd = tk.e;
          at = tk.used >= tk.n.length ? hit + 1 : hit;
        });
        // Not found: between the times around it.
        for (var i = 0; i < times.length; i++) {
          if (times[i] !== null) { continue; }
          var k = i;
          while (k < times.length && times[k] === null) { k++; }
          var from = i > 0 ? times[i - 1] : 0, to = k < times.length ? times[k] : (words.length ? +words[words.length - 1].e : from);
          for (var m = i; m < k; m++) { times[m] = from + (to - from) * (m - i + 1) / (k - i + 1); }
          i = k - 1;
        }
        return times;
      }
      // 2026-10-18: a Brobot bubble's own pieces - never a manager's part nested
      // in it (PLAYHEAD at its top, CHALK at its bottom), so piece numbers stay his.
      function ownSays(bubble) {
        return Array.prototype.filter.call(bubble.querySelectorAll(".brobot-say"), function (x) { return !x.closest(".nested-host"); });
      }
      function revealRest(bubble) {
        Array.prototype.forEach.call(ownSays(bubble).filter(function (x) { return !x.classList.contains("spoken"); }), function (s) {
          if (!s.classList.contains("speaking")) { s.classList.add("spoken"); }
        });
      }
      // The robot stopped speaking: if nothing more is lit in that bubble within
      // REVEAL_REST_MS (his reply is over - a highlight missed at the end), the
      // rest shows. A pause between his sentences is shorter.
      var REVEAL_REST_MS = 10000;
      function revealRestLater(bubble) {
        var seq = lastSpeakingSeq;
        setTimeout(function () {
          if (lastSpeakingSeq === seq && !bubble.querySelector(".brobot-say.speaking")) { revealRest(bubble); }
        }, REVEAL_REST_MS);
      }
      var lastSpeakingSeq = -1;
      var lastSpeakingBubble = null;
      function sayWords(text) {
        var t = String(text).toLowerCase()
          .replace(/gowp-awd/g, "gopod").replace(/ay-eye/g, "ai").replace(/goof-ball/g, "goofball");
        return t.split(/[^a-z0-9']+/).filter(Boolean);
      }
      function clearSpeaking() {
        Array.prototype.slice.call(container.querySelectorAll(".brobot-say.speaking")).forEach(function (el) {
          el.classList.remove("speaking");
          el.classList.add("spoken");
        });
      }
      // The spoken sentence and the bubble's sentences can split differently (the
      // robot's copy turns "..." into "."), so every bubble sentence mostly made of
      // the spoken words lights up; if none is, the one sharing the most words does.
      var speakingPolling = false;
      function pollSpeakingNow() {
        if (speakingPolling) { return; }
        speakingPolling = true;
        fetch("/gopod_speaking_now.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (now) {
            if (!now) { return; }
            var bubbles = ESN_TO_BUBBLE_CLASS[now.esn] ? container.querySelectorAll(".brobot-bubble." + ESN_TO_BUBBLE_CLASS[now.esn]) : [];
            // 2026-10-16 (comb-through: the highlight jumped to a second bubble):
            // his newest reply bubble - never an egg - and while he's still in a
            // reply (past its first piece) the bubble he's been lighting.
            var newest = null;
            // 2026-10-21: Wire-Pod names the bubble (its chat event) - that one, exactly.
            var named = now.feed ? container.querySelector('.brobot-bubble[data-feed="' + now.feed + '"]') : null;
            if (named) { newest = named; }
            for (var b = bubbles.length - 1; !named && b >= 0; b--) {
              if (!bubbles[b].classList.contains("popup-wrap") && ownSays(bubbles[b]).length) { newest = bubbles[b]; break; }
            }
            if (!named && now.active && typeof now.piece === "number" && now.piece > 0 && lastSpeakingBubble && lastSpeakingBubble.isConnected &&
                newest !== lastSpeakingBubble && ownSays(lastSpeakingBubble)[now.piece]) {
              newest = lastSpeakingBubble;
            }
            // Same sentence and same bubble: nothing to do. A song line's bubble can
            // arrive just after its first sentence starts - then it is lit again on
            // the new bubble (2026-09-30).
            if (now.seq === lastSpeakingSeq && newest === lastSpeakingBubble) { return; }
            lastSpeakingSeq = now.seq;
            lastSpeakingBubble = newest;
            clearSpeaking();
            if (!now.active || !newest) {
              if (!now.active && newest) { revealRestLater(newest); }
              return;
            }
            var spans = ownSays(newest);
            // GOLDEN (gopod_highlight_pieces.go): Wire-Pod names the piece by its
            // number - light exactly that one, no guessing.
            if (typeof now.piece === "number" && now.piece >= 0) {
              if (spans[now.piece]) { lightSpan(spans[now.piece]); focusSpeaking(spans[now.piece]); }
              return;
            }
            var spokenList = sayWords(now.text);
            if (!spokenList.length) { return; }
            var spoken = {};
            spokenList.forEach(function (w) { spoken[w] = true; });
            var picked = [], best = null, bestScore = 0;
            Array.prototype.forEach.call(spans, function (span) {
              var words = sayWords(span.textContent);
              if (!words.length) { return; }
              var inSpoken = words.filter(function (w) { return spoken[w]; }).length;
              if (words.length >= 3 && inSpoken / words.length >= 0.6) { picked.push(span); }
              var seen = {};
              words.forEach(function (w) { seen[w] = true; });
              var score = spokenList.filter(function (w) { return seen[w]; }).length / spokenList.length;
              if (score > bestScore) { bestScore = score; best = span; }
            });
            if (!picked.length && best && bestScore >= 0.5) { picked.push(best); }
            picked.forEach(lightSpan);
            if (picked.length) { focusSpeaking(picked[0]); }
          })
          .catch(function () { /* file not written yet - nothing to light up */ })
          .then(function () { speakingPolling = false; });
      }

      // 2026-10-01 (operator): clicking the Log button scrolls the page down to
      // the bottom of the log section, so the whole log window is in view - the
      // page itself to its end, the same place scrolling down by hand reaches
      // (scrollIntoView also moved the log's own inner scroll, operator 15:2x).
      var logButton = document.querySelector('a[onclick^="showLog"]');
      if (logButton) {
        logButton.addEventListener("click", function () {
          window.requestAnimationFrame(function () {
            window.scrollTo(0, document.documentElement.scrollHeight);
          });
        });
      }

      // 2026-10-04 (operator: "brobot 2 lists first ... robot ...b9e9 is brobot
      // 1"): the robots' battery icons at the top of the page show Brobot 1 first.
      // Wire-Pod's own battery.js adds them in its bot-list order and updates
      // each by its place in the page, so they are reordered by CSS order only
      // (the page order stays as battery.js made it). Each icon is known by the
      // serial in its tooltip.
      function orderBotStats() {
        var boxes = document.querySelectorAll("#botStats .batteryContainer");
        Array.prototype.forEach.call(boxes, function (box) {
          var serial = box.querySelector(".tooltip b");
          var place = serial ? Object.keys(ESN_TO_BROBOT).indexOf(serial.textContent.trim()) : -1;
          if (place >= 0 && box.style.order !== String(place)) { box.style.order = String(place); }
        });
      }
      setInterval(orderBotStats, 1000);

      // 2026-10-04 (operator: "Does my recording via my mic from the web page
      // work? ... please get it to work"; "grey Goverlord system bubbles in the
      // chat ... key press/hold + pre-record buffer, record after 1s press/hold
      // ... there's a buffer after key release ... handle 3x "000" as-is ... 3x
      // Numlock"): numpad push-to-talk from any keyboard this page sees, with the
      // golden numpad's rules (gopod_ptt_chat_writer_013.py) - numpad_ptt in the
      // settings file. With NumLock on, a live key of the golden persona map held
      // down records from the moment it goes down (the mic service keeps that
      // audio); it counts once held min_hold_seconds, keeps release_tail_seconds
      // after the key comes up (Wire-Pod waits that out), stops at
      // max_hold_seconds, and the words go to that key's Brobot. 000 (KP0) or
      // NumLock tapped exit_taps times within exit_window_seconds turns it off
      // until the page is reloaded. Only while the chat window is open; a page
      // can't take the numpad over, so the keys work while this page is the
      // window in use. Every step shows in a grey Goverlord status bubble (never
      // saved with the chat).
      var numpad = {
        on: (function () { try { return window.localStorage.getItem(NUMPAD_KEY) === "1"; } catch (e) { return false; } })(), map: null, bodyToBrobot: {}, active: null, pendingUp: null, taps: { Numpad0: [], NumLock: [] },
        cfg: { min_hold_seconds: 1, max_hold_seconds: 30, exit_taps: 3, exit_window_seconds: 2, texts: {} }
      };
      fetch("/api-sdk/gopod_numpad_map", { cache: "no-store" })
        .then(function (r) { return r.ok ? r.json() : null; })
        .then(function (data) {
          if (data && data.map) { numpad.map = data.map; numpad.bodyToBrobot = data.body_to_brobot || {}; numpad.personaHosts = data.persona_hosts || {}; numpad.personaReporters = data.persona_reporters || {}; }
        }).catch(function () { /* no map - push-to-talk stays quiet */ });
      // 2026-10-18 (operator: "live KPN = no grunge audio as this is not an error *
      // \"Too short - hold KP1 for at least 1 s\" = error = grunge audio plays * not
      // live yet key press = error = grunge audio plays"): a numpad error message
      // (numpad_ptt.error_texts) plays error_sound; a key that works plays nothing.
      function numpadText(name, values) {
        if ((numpad.cfg.error_texts || []).indexOf(name) >= 0) { playCue(numpad.cfg.error_sound); }
        var t = (numpad.cfg.texts || {})[name] || name;
        return String(t).replace(/\{(\w+)\}/g, function (m, k) { return values && values[k] != null ? values[k] : m; });
      }
      // A grey Goverlord bubble that adds each new message under the last.
      // 2026-10-19 (operator: "update header as per template version to,
      // '[input] to ...' = 'Numpad to ...', to match bubble template & other
      // bubbles"; "no other bubble when system menu is open * message ... appears
      // in the lower results window"): a numpad status is an Operator bubble
      // ("Numpad to <who>"); while the system menu is open it's a line in the
      // menu's results instead.
      function systemStatus(text, goneAfterMs, numpadTo) {
        if (systemMenuResults()) {
          systemMenuResult(text, "answer");
          return { bubble: null, textEl: null, set: function (t) { systemMenuResult(t, "answer"); }, done: function () {} };
        }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator system-status";
        bubble.appendChild(numpadTo ? operatorHead(numpadTo, "Numpad", "") : systemHead(systemLabel));
        var textEl = document.createElement("span");
        textEl.className = "brobot-bubble-text";
        textEl.textContent = text;
        bubble.appendChild(textEl);
        container.appendChild(bubble);
        revealWhenPlaced(bubble);
        // 2026-10-14 (operator: "When I key press I get the correct messages to
        // record & transcribe, etc.. * can you leave those in the chat window? (no
        // audio)"): each new message is added on its own line under the last, and
        // the bubble stays in the chat. Never spoken.
        var api = {
          bubble: bubble, textEl: textEl,
          set: function (t) {
            if (textEl.textContent === t) { return; }
            textEl.appendChild(document.createElement("br"));
            textEl.appendChild(document.createTextNode(t));
            // 2026-10-22 (operator: "scrolling isn't handling these bubbles
            // well"): a status line that grows the newest bubble keeps it in view.
            if (lastPlacedChild() === bubble) { scrollChatTo(Math.max(container.scrollTop, chatContentEnd())); }
          },
          done: function () { /* stays in the chat */ }
        };
        if (goneAfterMs) { api.done(goneAfterMs); }
        return api;
      }
      var numpadOpStatus = null;
      // GOPOD VISUAL CONTROL - the TPU +/- rotation (2026-10-20, operator: "Focus
      // on adding the tpu +/- rotation"; the retired 8011 demo's ring): KP+ / KP-
      // turn to the next / previous feed (visual_sources: Doc, Pip, the USB
      // webcam); Wire-Pod grabs its frame and runs the Edge TPU (gopod_visual.go).
      // One visual board shows the feed: FEED n/N, its name, LIVE or CACHED, the
      // TPU's objects, the frame with their boxes (magnifier). On CHALK's board
      // (the display host), or in the system menu's results while it's open;
      // the same board is turned in place while it's the newest thing in the chat.
      var visualBoard = null, visualBusy = false;
      function visualTurn(dir) {
        if (visualBusy) { return; }
        visualBusy = true;
        // 2026-10-20 (operator: "no multiple feeds to rotate = not applicable"):
        // one feed there - nothing turns, nothing shows.
        fetch("/api-sdk/gopod_visual?dir=" + encodeURIComponent(dir), { cache: "no-store" })
          .then(function (r) { return r.json(); })
          .then(function (v) { if (!v.single) { visualShow(visualBoardFor(), v); } })
          .catch(function () { /* Wire-Pod didn't answer */ })
          .then(function () { visualBusy = false; });
      }
      function visualBoardFor() {
        var results = systemMenuResults();
        var inPlace = visualBoard && visualBoard.isConnected &&
          (results ? visualBoard.parentNode === results : visualBoard === container.lastElementChild || visualBoard.nextElementSibling === typedRow);
        if (inPlace) { return visualBoard; }
        var board = document.createElement("div");
        board.className = results ? "system-menu-result visual-board" : "brobot-bubble operator system-status visual-board";
        if (!results) { board.appendChild(displayHead(board)); }
        var line = document.createElement("div");
        line.className = "visual-line";
        board.appendChild(line);
        var row = document.createElement("div");
        row.className = "topic-pictures visual-frame";
        board.appendChild(row);
        if (results) { results.appendChild(board); results.scrollTop = results.scrollHeight; }
        else { container.appendChild(board); revealWhenPlaced(board); }
        visualBoard = board;
        return board;
      }
      function visualShow(board, v) {
        var objects = v.objects && v.objects.length ? v.objects.join(", ") : "-";
        board.querySelector(".visual-line").textContent = "FEED " + v.index + "/" + v.count + " " + (v.label || v.id) +
          " " + (v.live ? "LIVE" : "CACHED") + " - TPU: " + objects + (v.status && v.status !== "Edge TPU: ok" ? " - " + v.status : "");
        var row = board.querySelector(".visual-frame");
        row.innerHTML = "";
        // 2026-10-20: photo, enhancement, TPU boxes (one feed); every feed's
        // photo then the TPU boxes (all feeds).
        (v.frames && v.frames.length ? v.frames : (v.frame ? [v.frame] : [])).forEach(function (src) {
          var img = document.createElement("img");
          img.src = src;
          img.alt = img.title = v.label || v.id;
          magnify(row, img);
          row.appendChild(img);
        });

      }
      // GOLDEN system menu (2026-10-19, operator: "New system- commands
      // drop-down-menu system * start from system-restart * then group & order by
      // semantic relevance * 🔇 opens it * 🔊 closes it * similar to manager
      // bubbles * system bubble contains drop-down-menu bubble * full inner
      // width/height for max window space * like after restart, where do I
      // start? what's next? * menu system should lead me like a tour guide"): a
      // Goverlord system bubble holding the menu bubble - system_menu's groups in
      // order, one open at a time, each with its guide line and what's next. A
      // command that needs nothing more runs on click; one that needs words goes
      // in the text box for you to finish and press Enter.
      var systemMenu = null, systemMenuData = [], systemMenuIntro = "", systemMenuLook = {};
      // 2026-10-19 (operator: "Make [menu system- bubble] font size 50% for more
      // room below * |menu system- bubble| |nested scroll bubble| * nested scroll
      // bubble underneath = "" or [results]"): under the menu, its results - each
      // system- line sent while it's open, and its answer.
      function systemMenuResults() {
        return systemMenu && systemMenu.isConnected ? systemMenu.querySelector(".system-menu-results") : null;
      }
      function systemMenuResult(text, kind) {
        var box = systemMenuResults();
        if (!box) { return false; }
        var row = document.createElement("div");
        row.className = "system-menu-result " + kind;
        row.textContent = kind === "ask" ? "\u203a " + text : text;
        box.appendChild(row);
        box.scrollTop = box.scrollHeight;
        return true;
      }
      function systemMenuOpen() {
        if (systemMenu && systemMenu.isConnected) { return; }
        var outer = document.createElement("div");
        outer.className = "brobot-bubble operator system-status system-menu-host";
        // Its look from system_menu.look (system-set system_menu.look.<name> <value>, F5).
        var look = { font_size: "--menu-font", background: "--menu-bg", code_background: "--menu-code-bg", code_color: "--menu-code-fg", results_background: "--menu-results-bg" };
        Object.keys(look).forEach(function (k) { if (systemMenuLook[k]) { outer.style.setProperty(look[k], systemMenuLook[k]); } });
        // 2026-10-19 (operator: "No header ... to recover window space"): no header.
        var menu = document.createElement("div");
        menu.className = "system-menu";
        if (systemMenuIntro) {
          var intro = document.createElement("div");
          intro.className = "system-menu-intro";
          intro.textContent = systemMenuIntro;
          menu.appendChild(intro);
        }
        systemMenuData.forEach(function (g, gi) {
          var group = document.createElement("details");
          group.className = "system-menu-group";
          if (gi === 0) { group.open = true; }
          var title = document.createElement("summary");
          title.textContent = g.title || "";
          group.appendChild(title);
          if (g.guide) {
            var guide = document.createElement("div");
            guide.className = "system-menu-guide";
            guide.textContent = g.guide;
            group.appendChild(guide);
          }
          (g.items || []).forEach(function (it) {
            var item = document.createElement("button");
            item.type = "button";
            item.className = "system-menu-item";
            var cmd = document.createElement("code");
            cmd.textContent = it.cmd;
            item.appendChild(cmd);
            var about = document.createElement("span");
            about.textContent = it.about || "";
            item.appendChild(about);
            item.addEventListener("click", function () {
              if (it.page === "usb") { usbPanel(); return; }
              if (it.page === "skip") { skipReporters("menu"); return; }
              if (it.run) {
                var numpadCmd = /^system-numpad\s+(on|off)\b/i.exec(it.cmd);
                if (numpadCmd) { setNumpadOn(numpadCmd[1].toLowerCase() === "on"); }
                var chosen = typedRow.querySelector("input[name=gopod-typed-robot]:checked");
                sendTypedLine(chosen ? chosen.value : "all", it.cmd);
              } else {
                typedText.value = it.fill != null ? it.fill : it.cmd + " ";
                typedText.focus();
              }
            });
            group.appendChild(item);
          });
          if (g.next) {
            var next = document.createElement("div");
            next.className = "system-menu-next";
            next.textContent = g.next;
            group.appendChild(next);
          }
          group.addEventListener("toggle", function () {
            if (!group.open) { return; }
            Array.prototype.forEach.call(menu.querySelectorAll("details"), function (d) { if (d !== group) { d.open = false; } });
          });
          menu.appendChild(group);
        });
        menu.appendChild(visitorsGroup(menu));
        outer.appendChild(menu);
        // 2026-10-19 (operator: "system window has 2 windows, system- commands &
        // results * upon mouseover of the horizontal space between them, cursor
        // changes to a space grabber, and I can resize by dragging the grabbed
        // space up & down"): the bar between them.
        var split = document.createElement("div");
        split.className = "system-menu-split";
        split.title = "Drag to resize";
        outer.appendChild(split);
        // 2026-10-19 (operator: "both start at 50% window height * vertical
        // adjustments work like this: system menu/results, 50%/50% 10%/90%,
        // 90%/10%"): the menu's share of the height, 10% to 90%; the results
        // take the rest.
        menu.style.flex = "0 0 70%"; // 2026-10-21 (operator: "system pane vs results pane to 70%/30%")
        split.addEventListener("pointerdown", function (ev) {
          ev.preventDefault();
          split.setPointerCapture(ev.pointerId);
          function move(e) {
            var box = outer.getBoundingClientRect();
            var share = Math.max(0.1, Math.min(0.9, (e.clientY - box.top) / box.height));
            menu.style.flex = "0 0 " + (share * 100).toFixed(1) + "%";
          }
          function up() {
            split.removeEventListener("pointermove", move);
            split.removeEventListener("pointerup", up);
          }
          split.addEventListener("pointermove", move);
          split.addEventListener("pointerup", up);
        });
        var results = document.createElement("div");
        results.className = "system-menu-results";
        outer.appendChild(results);
        container.appendChild(outer);
        systemMenu = outer;
        revealWhenPlaced(outer);
        // 2026-10-19 (operator: "scroll-lock system menu window, to top chat
        // window scroll position"): while it's open the chat stays at its top.
        (function pin() {
          if (systemMenu !== outer || !outer.isConnected) { return; }
          // 2026-10-23 (operator: "when open, this causes the results window to
          // scroll down past the brobots input row * If I close 1. the system
          // window resizes correctly just above the brobots row"): the window
          // is exactly the room above the typing row (one line or three),
          // whichever group is open.
          var room = Math.max(160, container.clientHeight - 12 - (typedRow.isConnected ? typedRow.offsetHeight : 0) - 8);
          if (outer.style.height !== room + "px") { outer.style.height = room + "px"; }
          if (outer.style.visibility !== "hidden") {
            var want = Math.max(0, outer.offsetTop - 12);
            if (Math.abs(container.scrollTop - want) > 2) {
              if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
              container.scrollTop = want;
            }
          }
          requestAnimationFrame(pin);
        })();
      }
      // 2026-10-20 (operator: "add in a dynamic USB panel of (hot-swappable)
      // devices"): in the menu's results, refreshed every 3 s while the menu is
      // open; a device just plugged in is marked "+ plugged in", one pulled out
      // "- pulled out" (for a few refreshes).
      var usbTimer = null;
      function usbPanel() {
        var results = systemMenuResults();
        if (!results) { return; }
        var panel = results.querySelector(".usb-panel");
        if (!panel) {
          panel = document.createElement("div");
          panel.className = "system-menu-result usb-panel";
          results.appendChild(panel);
        }
        var known = null, gone = {};
        clearInterval(usbTimer);
        function poll() {
          if (!panel.isConnected) { clearInterval(usbTimer); return; }
          fetch("/api-sdk/gopod_usb", { cache: "no-store" }).then(function (r) { return r.json(); }).then(function (list) {
            list = list || [];
            var now = {}, lines = ["USB devices (" + list.length + ") - live:"];
            list.forEach(function (d) {
              var k = d.id + " " + d.name;
              now[k] = true;
              lines.push("  " + d.id + "  " + d.name + (known && !known[k] ? "   + plugged in" : ""));
            });
            if (known) { Object.keys(known).forEach(function (k) { if (!now[k]) { gone[k] = 4; } }); }
            Object.keys(gone).forEach(function (k) {
              if (now[k] || gone[k]-- <= 0) { delete gone[k]; return; }
              lines.push("  " + k + "   - pulled out");
            });
            known = now;
            panel.textContent = lines.join("\n");
          }).catch(function () { panel.textContent = "USB devices: Wire-Pod didn't answer"; });
        }
        poll();
        usbTimer = setInterval(poll, 3000);
        results.scrollTop = results.scrollHeight;
      }
      // 2026-10-20 (operator: "brobot feed displaying outside results window"):
      // a picture row (not an egg) goes in the menu's results while it's open.
      // Review: only a system command's row ("menu:", gopodMenuPictureMark) goes
      // there - a song's or a topic's pictures stay in the chat.
      // 2026-10-21 (operator: "upon 1st reveal, easter egg bubble outside brobots
      // bubble + summary inside brobots window * upon more same egg reveals,
      // summary in brobots bubble only"): a found egg's summary - its title and a
      // line - at the top of the answering Brobot's bubble, small type.
      // 2026-10-21 (operator: "Same as message inside a brobot chat bubble.
      // Template it"): the bubble notice template - a titled notice at the top of
      // a Brobot's bubble ("notice:<title>|<line>"): an egg's summary, a native
      // Wire-Pod intent.
      function bubbleNoticeSaid(to) {
        // Any of his notices' lit lines (his shell may already be a new one).
        var cls = to === "Brobot 2" ? "brobot-2" : "brobot-1";
        Array.prototype.forEach.call(container.querySelectorAll(".brobot-bubble." + cls + " .bubble-notice .brobot-say.speaking"), function (x) {
          x.classList.remove("speaking");
          x.classList.add("spoken");
        });
      }
      // A notice line's words, as plain text or (said) as a lit piece.
      function noticeLine(line, said) {
        var el = document.createElement("span");
        el.className = "brobot-bubble-text egg-summary-text";
        if (said && !restoringChat) {
          var say = document.createElement("span");
          say.className = "brobot-say speaking";
          say._litAt = Date.now(); // a lit line that never hears "said" times out (batonHeld)
          say.textContent = line;
          el.appendChild(say);
        } else {
          el.textContent = line;
        }
        return el;
      }
      function bubbleNotice(to, text, said) {
        if (!/^Brobot [12]$/.test(String(to || "")) || restoringChat) { return; }
        var bar = text.indexOf("|");
        var title = bar >= 0 ? text.slice(0, bar) : text, line = bar >= 0 ? text.slice(bar + 1) : "";
        var shell = nestShell(to);
        // 2026-10-22 (operator: "in the same notice-bubble, if more info incoming
        // add it in underneath"): a notice with the same title takes the line.
        var same = Array.prototype.filter.call(shell.querySelectorAll(":scope > .bubble-notice"), function (n) { return n.dataset.title === title; })[0];
        if (same) {
          same.appendChild(noticeLine(line, said));
          return;
        }
        var box = document.createElement("div");
        box.className = "brobot-bubble operator system-status nested-host egg-summary bubble-notice";
        box.dataset.title = title;
        box.appendChild(systemHead(title));
        var words = document.createElement("span");
        words.className = "brobot-bubble-text egg-summary-text";
        words.textContent = line;
        box.appendChild(words);
        var reply = shell.querySelector(":scope > .brobot-bubble-text");
        if (reply) { shell.insertBefore(box, reply); } else { shell.appendChild(box); }
      }
      function menuPictures(list) {
        var results = systemMenuResults();
        if (!results || String(list).indexOf("menu:") !== 0) { return false; }
        list = String(list).slice("menu:".length);
        var row = document.createElement("div");
        row.className = "system-menu-result topic-pictures";
        String(list).split("|").forEach(function (src) {
          if (!src) { return; }
          var img = document.createElement("img");
          img.src = src;
          img.alt = img.title = src.split("/").pop();
          magnify(row, img);
          row.appendChild(img);
        });
        results.appendChild(row);
        results.scrollTop = results.scrollHeight;
        return true;
      }
      // The visitors slot (2026-10-24, operator: "In my system window, add an
      // additional slot for local network visitors * every detectable visitor
      // listed as nameless-visitor * whatever the 1st input is from a visitor,
      // that's their username request * I approve or clear ... While chat
      // participation is ON * Visitors are listed by their approved usernames *
      // add a checkbox to username as my mute switch"): the participation switch
      // (starts OFF), then each visitor - a nameless one by its network address,
      // a name request with Approve / Clear, an approved one with its box
      // (unticked = muted) and Clear. Refreshed every 2 s while the window is open.
      var visitorsTimer = null;
      function visitorsGroup(menu) {
        var group = document.createElement("details");
        group.className = "system-menu-group system-menu-visitors";
        var title = document.createElement("summary");
        title.textContent = "Local network visitors";
        group.appendChild(title);
        var body = document.createElement("div");
        body.className = "system-visitors";
        group.appendChild(body);
        group.addEventListener("toggle", function () {
          if (!group.open) { return; }
          Array.prototype.forEach.call(menu.querySelectorAll("details"), function (d) { if (d !== group) { d.open = false; } });
        });
        function act(action, key, on) {
          var form = new URLSearchParams();
          form.set("action", action);
          if (key) { form.set("key", key); }
          form.set("on", on ? "1" : "0");
          fetch("/api-sdk/gopod_visitors", { method: "POST", body: form }).then(function (r) { return r.ok ? r.json() : null; }).then(draw).catch(function () {});
        }
        function button(label, onClick) {
          var b = document.createElement("button");
          b.type = "button";
          b.className = "system-visitor-button";
          b.textContent = label;
          b.addEventListener("click", onClick);
          return b;
        }
        function draw(d) {
          if (!d) { return; }
          body.textContent = "";
          var on = document.createElement("label");
          on.className = "system-visitors-on";
          var onBox = document.createElement("input");
          onBox.type = "checkbox";
          onBox.checked = !!d.participation;
          onBox.addEventListener("change", function () { act("participation", "", onBox.checked); });
          on.appendChild(onBox);
          on.appendChild(document.createTextNode(" Visitor chat participation " + (d.participation ? "ON" : "OFF")));
          body.appendChild(on);
          var list = d.visitors || [];
          title.textContent = "Local network visitors (" + list.length + ")";
          if (!list.length) {
            var none = document.createElement("div");
            none.className = "system-visitor";
            none.textContent = "No visitors.";
            body.appendChild(none);
          }
          list.forEach(function (v) {
            var row = document.createElement("div");
            row.className = "system-visitor";
            var who = document.createElement("span");
            who.className = "system-visitor-name";
            if (v.approved) {
              var box = document.createElement("input");
              box.type = "checkbox";
              box.checked = !!v.allowed;
              box.title = v.allowed ? "Talking - untick to mute" : "Muted - tick to let them talk";
              box.addEventListener("change", function () { act("allow", v.key, box.checked); });
              row.appendChild(box);
              who.textContent = v.name;
            } else if (v.name) {
              who.textContent = "asks for: " + v.name;
            } else {
              who.textContent = "nameless-visitor";
            }
            row.appendChild(who);
            var meta = document.createElement("span");
            meta.className = "system-visitor-meta";
            meta.textContent = " " + v.addr + (v.device ? " · " + v.device : "");
            row.appendChild(meta);
            if (v.name && !v.approved) { row.appendChild(button("Approve", function () { act("approve", v.key, true); })); }
            if (v.name) { row.appendChild(button("Clear", function () { act("clear", v.key, false); })); }
            body.appendChild(row);
          });
        }
        function poll() {
          if (!group.isConnected) { clearInterval(visitorsTimer); return; }
          fetch("/api-sdk/gopod_visitors", { cache: "no-store" }).then(function (r) { return r.ok ? r.json() : null; }).then(draw).catch(function () {});
        }
        clearInterval(visitorsTimer);
        visitorsTimer = setInterval(poll, 2000);
        setTimeout(poll, 0);
        return group;
      }
      function systemMenuClose() {
        clearInterval(visitorsTimer);
        clearInterval(usbTimer);
        if (systemMenu) { systemMenu.remove(); systemMenu = null; }
      }
      function numpadPost(action, key) {
        return fetch("/api-sdk/gopod_numpad_ptt", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "action=" + action + "&key=" + encodeURIComponent(key)
        }).then(function (r) { return r.json(); })
          .catch(function (e) { return { ok: false, error: String(e) }; });
      }
      function numpadFailed(status, r) {
        status.set(r.error === "no_mic" || r.error === "not_ready" || r.error === "wait_message" ? numpadText(r.error) : numpadText("error", { error: r.error || "unknown" }));
        status.done(5000);
      }
      function numpadStart(key, brobot) {
        var a = { key: key, brobot: brobot, at: performance.now(), failed: false };
        numpad.active = a;
        a.status = systemStatus(numpadText("hold", { key: key, brobot: brobot }), 0, brobot);
        a.started = numpadPost("start", key).then(function (r) {
          if (!r.ok) { a.failed = true; numpadFailed(a.status, r); }
          a.micOff = !!r.mic_disabled;
          return r;
        });
        a.holdTimer = setTimeout(function () {
          // 2026-10-20: with the mic missing, "Mic disabled" in place of "Recording".
          if (numpad.active === a && !a.failed) { a.status.set(numpadText(a.micOff ? "mic_disabled" : "recording", { brobot: brobot })); }
        }, numpad.cfg.min_hold_seconds * 1000);
        a.maxTimer = setTimeout(numpadRelease, numpad.cfg.max_hold_seconds * 1000);
      }
      function numpadRelease() {
        var a = numpad.active;
        if (!a) { return; }
        numpad.active = null;
        clearTimeout(a.holdTimer);
        clearTimeout(a.maxTimer);
        a.started.then(function (r) {
          if (!r.ok) { return; }
          var held = (performance.now() - a.at) / 1000;
          if (held >= numpad.cfg.min_hold_seconds) { a.status.set(numpadText("transcribing")); }
          // Its words may reach the feed before this answer does.
          numpadOpStatus = a.status;
          numpadPost("stop", a.key).then(function (out) {
            if (out.status !== "ok" && numpadOpStatus === a.status) { numpadOpStatus = null; }
            if (!out.ok) { numpadFailed(a.status, out); return; }
            if (out.status === "ok") { a.status.done(0); return; } // the words show as the operator's line, in this bubble
            if (out.status === "too_short") {
              a.status.set(numpadText("too_short", { key: a.key, min: numpad.cfg.min_hold_seconds }));
            } else {
              a.status.set(numpadText("nothing_heard"));
            }
            a.status.done(4000);
          });
        });
      }
      // 2026-10-14 (operator: "Numpad push-to-talk off (3x NumLock) ... * these
      // can be spoken"): Goverlord says it - its bubble comes with the voice.
      function numpadSay(text) {
        fetch("/api-sdk/gopod_system_say", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "text=" + encodeURIComponent(text)
        }).then(function (r) { if (!r.ok) { systemStatus(text); } })
          .catch(function () { systemStatus(text); });
      }
      // 2026-10-16 (operator: "For numpad enabled key presses * as an audio cue,
      // per key press, play grunge wav"): numpad_ptt.key_sound, at its share of
      // chat_audio.volume.
      function numpadKeySound() { playCue(numpad.cfg.key_sound); }
      // One short sound cue ({file, volume}): loaded once, replayed each time.
      function playCue(ks) {
        if (!ks || !ks.file) { return; }
        try {
          // 2026-10-17: loaded once, replayed each press.
          if (!stingFiles[String(ks.file).split("/").pop()]) { preloadStings([ks.file]); }
          var a = stingFromPool(ks.file) || new Audio(normSrc(ks.file));
          a.volume = chatLevel(volumeShare(ks.volume) * audioLevel(ks.file));
          var played = a.play();
          if (played && played.catch) { played.catch(function () { /* blocked until the page is used */ }); }
        } catch (e) { /* no sound */ }
      }
      function numpadExitTap(code) {
        var now = performance.now(), span = numpad.cfg.exit_window_seconds * 1000;
        var taps = numpad.taps[code] = numpad.taps[code].filter(function (t) { return now - t <= span; });
        taps.push(now);
        if (taps.length < numpad.cfg.exit_taps) { return; }
        setNumpadOn(false);
        numpadSay(numpadText("exit", { how: numpad.cfg.exit_taps + "\u00d7 " + (code === "NumLock" ? "NumLock" : "000") }));
      }
      function setNumpadOn(on) {
        // 2026-10-08 (operator: "add to log numpad checkbox clicks").
        fetch("/api-sdk/gopod_page_note", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "note=" + encodeURIComponent("Numpad " + (on ? "on" : "off"))
        }).catch(function () { /* Wire-Pod not up */ });
        numpad.on = on;
        try { window.localStorage.setItem(NUMPAD_KEY, on ? "1" : "0"); } catch (e) { /* in-page only */ }
        if (!on) { numpadRelease(); }
      }
      function chatShown() { return container.classList.contains("expanded") && container.offsetParent !== null; }
      // 2026-10-09 (operator: "upon numpad checked, I cannot type normally * I have
      // a keyboard + numpad in one"): while a text box has the focus, every key
      // types as usual; push-to-talk is for when it doesn't.
      function typingInBox() {
        var el = document.activeElement;
        return !!el && (el.tagName === "INPUT" && /^(text|search)?$/i.test(el.type || "") || el.tagName === "TEXTAREA" || el.isContentEditable);
      }
      document.addEventListener("keydown", function (ev) {
        // 2026-10-10 (live 07:38: "numpad checkbox didn't work" - after typing
        // "Party" the typing line kept the focus, so every numpad key was taken
        // as typing): with Numpad? ticked, the numpad's own keys are always
        // push-to-talk, even in a text box; every other key types as usual.
        var numpadKey = /^Numpad[0-9]$/.test(ev.code) || ev.code === "NumLock";
        if (!numpad.on || !numpad.map || !chatShown() || (typingInBox() && !numpadKey)) { return; }
        if (numpadKey && !ev.repeat) { numpadKeySound(); }
        if (ev.code === "NumLock") {
          if (!ev.repeat) { numpadExitTap(ev.code); }
          return;
        }
        // 2026-10-20 (operator: "upon numpad's '*' input, obey same as 'Esc' key press").
        if (ev.code === "NumpadMultiply" && !typingInBox()) {
          ev.preventDefault();
          if (!ev.repeat) { skipReporters("numpad *"); }
          return;
        }
        // GOPOD VISUAL CONTROL (2026-10-20): KP+ next feed, KP- previous (visualTurn).
        if ((ev.code === "NumpadAdd" || ev.code === "NumpadSubtract") && !typingInBox()) {
          ev.preventDefault();
          if (!ev.repeat) { visualTurn(ev.code === "NumpadAdd" ? "next" : "prev"); }
          return;
        }
        // 2026-10-16 (operator: "can you extend this detection, so a long '0'
        // press acts as the mic version of 'keyboard's 'All' radio button
        // option?"): 0 held all_hold_seconds is push-to-talk to all the Brobots
        // (Brobot 1 answers, then Brobot 2, like the All button); a shorter press
        // is one of the exit taps, as before.
        if (ev.code === "Numpad0") {
          ev.preventDefault();
          if (ev.repeat) { return; }
          if (numpad.pendingUp && numpad.pendingUp.key === "KP0") {
            clearTimeout(numpad.pendingUp.timer);
            numpad.pendingUp = null;
            return;
          }
          if (numpad.active || numpad.zeroDown) { return; }
          var downAt = performance.now();
          numpad.zeroDown = {
            timer: setTimeout(function () {
              numpad.zeroDown = null;
              if (numpad.active) { return; }
              if (!ev.getModifierState("NumLock")) { systemStatus(numpadText("numlock_off", { key: "KP0" }), 3000, "Brobots"); return; }
              numpadStart("KP0", numpad.cfg.all_label || "All");
              numpad.active.at = downAt;
            }, (numpad.cfg.all_hold_seconds || 0.5) * 1000)
          };
          return;
        }
        var m = /^Numpad([1-9])$/.exec(ev.code);
        if (!m) { return; }
        ev.preventDefault();
        if (ev.repeat) { return; }
        var key = "KP" + m[1];
        // A key-up and key-down pair a few ms apart is the key's auto-repeat
        // (some remote desktops send it that way), not a new press.
        if (numpad.pendingUp && numpad.pendingUp.key === key) {
          clearTimeout(numpad.pendingUp.timer);
          numpad.pendingUp = null;
          return;
        }
        if (numpad.active) { return; }
        if (!ev.getModifierState("NumLock")) { systemStatus(numpadText("numlock_off", { key: key }), 3000, "Brobots"); return; }
        var entry = (numpad.map.numlock_on || {})[key];
        // 2026-10-18: a voice-only persona (KP7 CHALK, KP8 PLAYHEAD) talks to its host.
        var personaHost = entry && entry.status === "live" ? (numpad.personaHosts || {})[String(entry.persona || "").toUpperCase()] : "";
        if (personaHost) { numpadStart(key, personaHost); return; }
        if (!entry || entry.status !== "live" || !numpad.bodyToBrobot[entry.body]) {
          systemStatus(numpadText("not_live", { key: key, persona: entry ? entry.persona : "-", body: entry ? entry.body : "-" }), 3000, entry && entry.persona ? String(entry.persona) : key);
          return;
        }
        numpadStart(key, numpad.bodyToBrobot[entry.body]);
      });
      document.addEventListener("keyup", function (ev) {
        if (ev.code === "Numpad0" && numpad.zeroDown) {
          clearTimeout(numpad.zeroDown.timer);
          numpad.zeroDown = null;
          numpadExitTap("Numpad0");
          return;
        }
        var m = /^Numpad([0-9])$/.exec(ev.code);
        if (!m || !numpad.active || numpad.active.key !== "KP" + m[1]) { return; }
        ev.preventDefault();
        var key = "KP" + m[1];
        numpad.pendingUp = {
          key: key,
          timer: setTimeout(function () { numpad.pendingUp = null; numpadRelease(); }, 60)
        };
      });

      loadChat();
      showChat(chatWasOpen());
      setInterval(pollSpeakingNow, 250);
      setInterval(pollEyeColors, 500);
      pollEyeColors();
      // The name list and the settings load before the first bubbles, so none show
      // the misheard spelling or a cue marker.
      Promise.all([loadNameSpellings(), loadBrobotsSettings()]).then(function () {
        restoreChat();
        setInterval(pollBubbles, 500);
        setInterval(pollFeed, 500);
        setInterval(pollRecordSession, 1000);
        setInterval(pollWarmup, 500);
        pollBubbles();
        pollFeed();
        pollWarmup();
      });
    })();

// ## EOF
