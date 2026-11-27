
    function copyText(text) {
        const tmp = document.createElement("textarea");
        tmp.value = text;
        document.body.appendChild(tmp);
        tmp.select();
        document.execCommand("copy");
        tmp.remove();
    }

    function visibleChatText(el) {
        var out = [];
        var line = "";
        function flush() {
            var t = line.replace(/[\u2063\u200b\u00ad]/g, "").replace(/[ \t]+/g, " ").trim();
            if (t) { out.push(t); }
            line = "";
        }
        function walk(node) {
            if (node.nodeType === 3) { line += node.nodeValue; return; }
            if (node.nodeType !== 1) { return; }
            var cs = window.getComputedStyle(node);
            if (cs.display === "none" || cs.visibility === "hidden" || Number(cs.opacity) === 0 || node.dataset.staged === "1" || node.getAttribute("aria-hidden") === "true") { return; }
            if (node.tagName === "FORM" || node.classList.contains("gopod-typed-input") || /^(SELECT|BUTTON|INPUT|TEXTAREA)$/.test(node.tagName)) { return; }
            if (node.tagName === "BR") { flush(); return; }
            if (node.tagName === "IMG") { if (node.alt && /\p{Extended_Pictographic}|\p{Regional_Indicator}/u.test(node.alt)) { line += node.alt; } return; }
            var block = !/^inline/.test(cs.display);
            if (block) { flush(); }
            for (var c = node.firstChild; c; c = c.nextSibling) { walk(c); }
            if (block) { flush(); }
        }
        walk(el);
        flush();
        return out.join("\n");
    }

    function copyToClipboard(all) {
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
        const bubbles = document.getElementById("brobot-chat-bubbles-container");
        if (bubbles && bubbles.classList.contains("expanded")) {
            const text = visibleChatText(bubbles);
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

    (function () {
      var container = document.getElementById("brobot-chat-bubbles-container");
      var textarea = document.getElementById("botTranscriptedTextArea");
      var CHAT_STORE_KEY = "gopod_brobots_chat_v1";
      var CHAT_OPEN_KEY = "gopod_brobots_chat_open";
      var CHAT_MAX_BUBBLES;
      var CHAT_MAX_SEEN;
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
          else if (b.format === "notice") { bubbleNotice(b.who, b.text); }
          else { appendBubble(b.who, b.text, b.format, b.to, b.via); }
        });
        restoringChat = false;
        lastOperatorBubble = null;
        popupExchange = false;
        popupWraps = {};
        Array.prototype.forEach.call(container.querySelectorAll(".brobot-say"), function (s) { s.classList.add("spoken"); });
        container.scrollTop = container.scrollHeight;
      }
      function setChatOpen(open) {
        try { window.localStorage.setItem(CHAT_OPEN_KEY, open ? "1" : "0"); } catch (e) { /* in-page only */ }
      }
      function chatWasOpen() {
        try { return window.localStorage.getItem(CHAT_OPEN_KEY) !== "0"; } catch (e) { return true; }
      }
      var ESN_TO_BROBOT = { "ESN01": "Brobot 1", "ESN02": "Brobot 2" };
      var isAdminPage = false;
      var typedRow = document.createElement("form");
      typedRow.className = "gopod-typed-input";
      typedRow.setAttribute("autocomplete", "off");
      var savedRobot = "all";
      function typedRowText(text, className) {
        var span = document.createElement("span");
        span.className = className;
        span.textContent = text;
        typedRow.appendChild(span);
      }
      typedRowText("Brobots", "gopod-typed-name");
      var allFirstSelect = null;
      var numpadLayoutSelect = null, NUMPAD_LAYOUT_KEY = "gopod_numpad_layout", numpadMapLoad = null;
      function numpadLayoutNow() { return numpadLayoutSelect && numpadLayoutSelect.dataset.picked === "1" ? numpadLayoutSelect.value : ""; }
      var PICKER_KEYS = ["+*", "+.", "+0", "+1", "+2", "+3", "+4", "+5", "+6", "+7", "+8", "+9",
        "-*", "-0", "-1", "-2", "-3", "-4", "-5", "-6", "-7", "-8", "-9"];
      function pickerKey(v) {
        var n = String(v).slice(1), off = String(v).charAt(0) === "-";
        if (n === ".") { return "KPdot"; }
        return n === "*" ? "KPmul" : "KP" + n + (off ? "off" : "");
      }
      function pickerWords() { return ((typeof numpad !== "undefined" && numpad.cfg) || {}).picker || {}; }
      function relabelPicker() {
        if (!allFirstSelect) { return; }
        var targets = (typeof numpad !== "undefined" && numpad.targets) || {};
        var map = (typeof numpad !== "undefined" && numpad.map) || {}, w = pickerWords();
        Array.prototype.forEach.call(allFirstSelect.options, function (o) {
          var n = o.value.slice(1), off = o.value.charAt(0) === "-";
          if (n === "*") { o.textContent = (off ? w.all_second : w.all_first) || o.value; return; }
          if (n === ".") { o.textContent = targets.KPdot || o.value; o.title = w.any || ""; o.disabled = !targets.KPdot; return; }
          var key = pickerKey(o.value), who = targets[key] || "";
          var entry = ((off ? map.numlock_off : map.numlock_on) || {})["KP" + n] || {};
          var persona = entry.persona && entry.persona !== "-" ? entry.persona : "";
          o.title = String((off ? w.off : w.on) || "{n} {persona}").replace("{n}", n).replace("{persona}", persona).trim();
          o.textContent = who || (w.unassigned || "unassigned");
          o.disabled = !who;
          if (o.classList.contains("gopod-brain")) { o.disabled = o.disabled || !isAdminPage; o.hidden = !isAdminPage; }
        });
      }
      (function () {
        var label = document.createElement("label");
        label.className = "gopod-all-pick";
        allFirstSelect = document.createElement("select");
        allFirstSelect.className = "gopod-all-first";
        allFirstSelect.title = "";
        PICKER_KEYS.forEach(function (n) {
          var o = document.createElement("option");
          o.value = n;
          o.textContent = n;
          if (n === "-0") { o.className = "gopod-brain"; o.hidden = true; o.disabled = true; }
          allFirstSelect.appendChild(o);
        });
        allFirstSelect.value = savedRobot === "all" ? "+*" : "+1";
        label.appendChild(allFirstSelect);
        typedRow.appendChild(label);
        numpadLayoutSelect = document.createElement("select");
        numpadLayoutSelect.className = "gopod-all-first gopod-numpad-layout";
        numpadLayoutSelect.title = "";
        [["1", "pc"], ["2", "usb"]].forEach(function (p) {
          var o = document.createElement("option");
          o.value = p[1];
          o.textContent = p[0];
          numpadLayoutSelect.appendChild(o);
        });
        try { var keptLayout = window.localStorage.getItem(NUMPAD_LAYOUT_KEY); if (keptLayout) { numpadLayoutSelect.value = keptLayout; numpadLayoutSelect.dataset.picked = "1"; } } catch (e) { /* no storage */ }
        numpadLayoutSelect.addEventListener("change", function () {
          numpadLayoutSelect.dataset.picked = "1";
          try { window.localStorage.setItem(NUMPAD_LAYOUT_KEY, numpadLayoutSelect.value); } catch (e) { /* no storage */ }
          if (typeof numpadMapLoad === "function") { numpadMapLoad(); }
        });
        var layoutPick = document.createElement("label");
        layoutPick.className = "gopod-layout-pick";
        layoutPick.appendChild(numpadLayoutSelect);
        typedRow.appendChild(layoutPick);
        var typedSubmit = document.createElement("button");
        typedSubmit.type = "submit";
        typedSubmit.className = "gopod-typed-submit";
        typedSubmit.textContent = "";
        typedRow.appendChild(typedSubmit);
      })();
      function typedChoice() {
        var v = allFirstSelect ? allFirstSelect.value : "+*";
        if (v === "+*") { return { serial: "all", first: "1" }; }
        if (v === "-*") { return { serial: "all", first: "2" }; }
        if (v === "-0") { return { serial: "system", first: "" }; }
        if (v === "+.") { return { serial: "kp:dot", first: "" }; }
        var n = v.slice(1) + (v.charAt(0) === "-" ? "off" : "");
        return { serial: "kp:" + n, first: "" };
      }
      var allSecondFirst = {};
      typedRowText("|", "gopod-typed-sep");
      var NUMPAD_KEY = "gopod_brobots_numpad_on";
      var typedFirstSep = typedRow.lastElementChild;
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
      var sessionTimer = document.createElement("span");
      sessionTimer.className = "gopod-session-timer";
      sessionTimer.textContent = "00:00:00";
      typedRow.appendChild(sessionTimer);
      var timerFromMs = Date.now(), recordStartedMs = 0, wasRecording = null;
      function showSessionTimer() {
        var ms = Date.now() - timerFromMs;
        var sec = Math.max(0, Math.floor(ms / 1000));
        var two = function (n) { return (n < 10 ? "0" : "") + n; };
        sessionTimer.textContent = two(Math.floor(sec / 3600)) + ":" + two(Math.floor(sec / 60) % 60) + ":" + two(sec % 60);
        sessionTimer.classList.toggle("recording", !!recordStartedMs);
      }
      setInterval(showSessionTimer, 1000);
      var backgroundLoop = (function () {
        var volumeCfg = 0.14;
        var file = "", volume = 0.14, ctx = null, buffer = null, source = null, gain = null, wanted = false, setupDone = false;
        var speakerOn = false, runQuiet = false;
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
          volume: function () { return volume; },
          setVolume: function (v) {
            volume = Math.max(0, Math.min(1, v));
            applyLevel();
          },
          setup: function (cfg) {
            file = String(cfg.file || "").trim();
            volume = volumeShare(cfg.volume);
            volumeCfg = volume;
            if (!setupDone) { speakerOn = cfg.starts_on === true; setupDone = true; if (typeof showSpeaker === "function") { showSpeaker(); } }
            applyLevel();
            wanted = !!file;
            if (wanted && pageClicked) { load(); }
          },
          restart: function () { start(); },
          speaker: function () { return speakerOn; },
          setSpeaker: function (on) {
            speakerOn = !!on;
            applyLevel();
          },
          runQuiet: function (quiet) { runQuiet = !!quiet; applyLevel(); },
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
      var typedPlaceholder = "";
      typedText.placeholder = typedPlaceholder;
      typedText.setAttribute("aria-label", "Type to the chosen Brobot");
      typedRow.insertBefore(typedText, typedFirstSep);
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
        if (!isAdminPage) { return; }
        if (!backgroundLoop.speaker()) { volumeWrap.classList.remove("open"); }
        if (backgroundLoop.speaker()) { systemMenuOpen(); } else { systemMenuClose(); }
      });
      showSpeaker();
      typedRowText("|", "gopod-typed-sep");
      typedRow.appendChild(speakerToggle);
      var masterToggle = document.createElement("button");
      masterToggle.type = "button";
      masterToggle.className = "gopod-loop-speaker";
      try { masterMuted = localStorage.getItem("gopod_master_mute") === "1"; } catch (err) { masterMuted = false; }
      function showMasterMute() {
        masterToggle.textContent = masterMuted ? "\uD83D\uDD07" : "\uD83D\uDD0A";
        masterToggle.setAttribute("aria-pressed", masterMuted ? "true" : "false");
        masterToggle.title = masterMuted ? "All sound muted - click to unmute" : "All sound on - click to mute";
        masterToggle.setAttribute("aria-label", masterToggle.title);
      }
      masterToggle.addEventListener("click", function () {
        masterMuted = !masterMuted;
        try { localStorage.setItem("gopod_master_mute", masterMuted ? "1" : "0"); } catch (err) { /* not kept */ }
        backgroundLoop.setVolume(backgroundLoop.volume());
        showMasterMute();
      });
      showMasterMute();
      typedRow.appendChild(masterToggle);
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
      function volumeOpen() { clearTimeout(volumeHide); if (!backgroundLoop.speaker()) { volumeWrap.classList.remove("open"); return; } showVolume(); volumeWrap.classList.add("open"); }
      function volumeClose() { clearTimeout(volumeHide); volumeHide = setTimeout(function () { volumeWrap.classList.remove("open"); }, 400); }
      [speakerToggle, volumeWrap].forEach(function (el) {
        el.addEventListener("mouseenter", volumeOpen);
        el.addEventListener("mouseleave", volumeClose);
      });
      typedText.addEventListener("keydown", function (ev) {
        if (ev.key === "Escape") { typedText.blur(); }
      });
      document.addEventListener("keydown", function (ev) {
        if (ev.key === "Escape" && !typingInBox()) { escapeRun(); }
      });
      typedRow.addEventListener("submit", function (ev) {
        ev.preventDefault();
        var text = typedText.value.trim();
        if (!text) { return; }
        typedText.value = "";
        if (/^(system-)?skip[.!]?$/i.test(text)) { skipReporters("typed"); return; }
        var numpadCmd = /^system-numpad\s+(on|off)\b/i.exec(text);
        if (numpadCmd) { setNumpadOn(numpadCmd[1].toLowerCase() === "on"); }
        sendTypedLine("", text);
      });
      function sendTypedLine(serial, text) {
        var c = serial ? { serial: serial, first: "" } : typedChoice();
        serial = c.serial;
        var first = c.first;
        if (first === "2") { allSecondFirst[text] = true; }
        fetch("/api-sdk/gopod_typed_input", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "serial=" + encodeURIComponent(serial) + "&text=" + encodeURIComponent(text) + (first ? "&first=" + first : "") + "&visitor=" + encodeURIComponent(visitorId)
        }).then(function (r) {
          if (!r.ok) { typedText.value = text; return null; }
          return r.text();
        }).then(function (body) {
          var note = "";
          try { note = body && body.charAt(0) === "{" ? (JSON.parse(body).note || "") : ""; } catch (err) { note = ""; }
          typedText.placeholder = note || typedPlaceholder;
          typedText.classList.toggle("gopod-typed-note", !!note);
        }).catch(function () { typedText.value = text; });
      }
      container.appendChild(typedRow);
      var TYPED_TEXT_MIN_PX;
      function fitTypedRow() {
        if (TYPED_TEXT_MIN_PX == null) { return; } // until the settings load
        var need = TYPED_TEXT_MIN_PX, gap = parseFloat(getComputedStyle(typedRow).columnGap) || 0;
        Array.prototype.forEach.call(typedRow.children, function (el) {
          if (el === typedText || getComputedStyle(el).position === "absolute" || el.offsetParent === null) { return; }
          need += el.offsetWidth + gap;
        });
        var room = typedRow.clientWidth - (parseFloat(getComputedStyle(typedRow).paddingLeft) || 0) - (parseFloat(getComputedStyle(typedRow).paddingRight) || 0);
        typedRow.classList.toggle("stacked", room > 0 && need > room);
      }
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

      window.showBrobotChatBubbles = function () {
        var open = !container.classList.contains("expanded");
        setChatOpen(open);
        showChat(open);
        if (open) {
          requestAnimationFrame(function () {
            if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
            container.scrollTop = chatMaxScroll();
          });
        }
      };

      var chatBack = document.createElement("button");
      chatBack.type = "button";
      chatBack.className = "gopod-chat-back";
      chatBack.textContent = "\u2039 Wire-Pod";
      chatBack.addEventListener("click", function () { document.body.classList.remove("gopod-chat-only"); });
      var chatOnlyPending = false;
      function chatOnlyOn() {
        if (!pageRoleKnown) { chatOnlyPending = true; return; }
        if (isAdminPage) { document.body.classList.remove("gopod-chat-only"); return; }
        var log = document.getElementById("section-log");
        if (chatBack.parentNode) { chatBack.remove(); }
        document.body.classList.add("gopod-chat-only");
        if (!container.classList.contains("expanded")) { window.showBrobotChatBubbles(); }
      }
      if ("serviceWorker" in navigator && window.isSecureContext) {
        navigator.serviceWorker.register("/gopod_sw.js").catch(function () { /* no app - the page works as ever */ });
      }
      var appStart = /[?&]app(=|&|$)/.test(location.search);
      function visitorOpenChat() {
        if (typeof window.toggleVisibility === "function") {
          window.toggleVisibility(["section-intents", "section-log", "section-botauth", "section-version", "section-uicustomizer"], "section-log", "icon-Logs");
        } else {
          var log = document.getElementById("section-log");
          if (log) { log.style.display = ""; }
        }
        chatOnlyOn();
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
      var TRAILING_EMOJI_RE = /^(\s*(?:\p{Extended_Pictographic}|\uFE0F|\u200D)+)+(?=\s|$)/u;
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
            var ellipsis = text[i] === "\u2026" || (text[i] === "." && text[i - 1] === ".");
            if (text[i + 1] === " " || (ellipsis && isEnd === softEnd && /[A-Za-z0-9]/.test(text[i + 1] || "")) || (isEnd === softEnd && dashPause(text, i))) { out.push(cur); cur = ""; }
          }
        }
        if (cur.trim()) { out.push(cur); }
        return out;
      }
      function splitSentences(line) {
        var out = [];
        cutAt(line.trim(), hardEnd).forEach(function (sentence) {
          cutAt(sentence, softEnd).forEach(function (piece) { if (piece.trim()) { out.push(piece.trim()); } });
        });
        return out;
      }
      function bubbleLinks(html) {
        return String(html).replace(/\[([^\]\n]+)\]\((\/[^)\s]*|https?:\/\/[^)\s]+)\)/g, function (m, words, href) {
          return "<a class=\"bubble-link\" href=\"" + href.replace(/"/g, "&quot;") + "\" target=\"_blank\" rel=\"noopener\">" + words + "</a>";
        });
      }
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
        if (segmentMarker && t.indexOf(segmentMarker) >= 0) {
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

      function appendBubble(who, spoken, format, to, via, then, byWho) {
        var next = function () { if (then) { then(); } };
        voiceLitNext = null; // the next voice is this bubble's (the Operator's), not an earlier line's
        if (!restoringChat) {
          spoken = spellNames(spoken);
          if (who === "Operator" && !to) { to = operatorTarget(spoken); }
          recordBubble(who, spoken, format, to, via);
        }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble " + (who === "Operator" ? "operator" : who === "Brobot 1" ? "brobot-1" : "brobot-2");
        if (byWho) { bubble.classList.add("host-" + personaSlug(byWho)); }
        else if (/^Brobot [12]$/.test(who)) { brobotHostClass(bubble, who, brobotBy[who]); }
        if (format === "llm") { bubble.className += " brobot-bubble-live"; }
        else if (segmentMarker && String(spoken).indexOf(segmentMarker) >= 0) { bubble.className += " brobot-bubble-scripted"; }
        if (who === "Operator" && !restoringChat && /^system-/i.test(String(spoken).trim()) && systemMenuResult(String(spoken).trim(), "ask")) { next(); return; }
        if (who === "Operator") {
          var opText = document.createElement("span");
          opText.className = "brobot-bubble-text";
          opText.dataset.raw = who + ": " + spoken;
          if (!restoringChat) {
            var opSay = document.createElement("span");
            opSay.className = "brobot-say";
            opSay.textContent = spoken;
            opText.appendChild(opSay);
            bubble.classList.add("voice-lit");
            setTimeout(function () { if (!opSay.classList.contains("speaking") && !opSay.classList.contains("spoken")) { opSay.classList.add("spoken"); } }, 30000);
          } else {
            opText.textContent = spoken;
          }
          if (!restoringChat) { voiceLineAhead(spoken, opText); }
          var opHead = operatorHead(to || operatorTarget(spoken), via, spoken);
          bubble.appendChild(opHead);
          if (opHead.dataset.sender) { bubble.classList.add("host-" + personaSlug(opHead.dataset.sender)); }
          bubble.appendChild(opText);
          bubbleSound("Operator");
          bubble.dataset.own = "1"; // the operator's own line
          popupWraps = {}; popupExchange = false; eggMarked = {};
          var shellFor = to || operatorTarget(spoken);
          if (/^Brobot [12]$/.test(shellFor)) { delete nestShells[shellFor]; }
          else if (shellFor === "Brobots" || shellFor === "All" || !shellFor) { nestShells = {}; }
          var st = numpadOpStatus;
          numpadOpStatus = null;
          if (st && !restoringChat && st.bubble && st.bubble.isConnected && /^KP/.test(String(via || ""))) {
            st.bubble.className = bubble.className;
            if (managerHosts[to]) { st.bubble.classList.add("numpad-to-gomad"); }
            st.bubble.classList.add("numpad-line");
            st.bubble.replaceChild(bubble.firstChild, st.bubble.firstChild);
            st.textEl.className = "numpad-lines";
            st.textEl.style.display = "block";
            st.bubble.appendChild(opText);
            st.bubble.dataset.own = "1";
            if (format !== "plain") { lastOperatorBubble = st.bubble; operatorInBatch = true; }
            next();
            return;
          }
          container.appendChild(bubble);
          if (restoringChat) { next(); return; }
          if (warmupBubble && warmupBubble.isConnected) { container.appendChild(warmupBubble); }
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
        var inPopup = (popupExchange || eggMarked[who]) && /^Brobot [12]$/.test(who);
        delete eggMarked[who];
        var shell = nestShells[who];
        if (shell && shell.isConnected && !shell.dataset.filled && !restoringChat) {
          shell.dataset.filled = "1";
          var shellWho = shell.querySelector(".brobot-bubble-who");
          if (shellWho && brobotBy[who]) { shellWho.textContent = brobotBy[who] + ":"; }
          if (feedIdNow) { shell.dataset.feed = String(feedIdNow); }
          bubble.className.split(" ").forEach(function (c) { if (c) { shell.classList.add(c); } });
          if (/^Brobot [12]$/.test(who)) { brobotHostClass(shell, who, brobotBy[who]); }
          bubbleNote(shell, "filled");
          if (inPopup) { shell.classList.add("popup-winner"); }
          shell.appendChild(textEl);
          attachPendingBall(shell, spoken);
          if (inPopup) { marqueeAttach(shell); }
          addQrCodes(shell);
          addWeatherFace(shell);
          next();
          return;
        }
        if (inPopup) { bubble.classList.add("popup-winner"); }
        if (feedIdNow && /^Brobot [12]$/.test(who)) { bubble.dataset.feed = String(feedIdNow); }
        bubble.appendChild(bubbleHead(who));
        bubble.appendChild(textEl);
        attachPendingBall(bubble, spoken);
        placeBubble(bubble, next, function (b) { if (inPopup) { marqueeAttach(b); } });
      }

      var pendingReveal = [];
      var revealTimer = null;
      var placeAgain = false;
      var operatorInBatch = false; // the bubbles being placed include an operator line
      var awaitingReply = false;
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
        if (nextBubbleVoice) { bubble._voice = nextBubbleVoice; nextBubbleVoice._bubble = bubble; nextBubbleVoice = null; }
        bubble._feedId = feedIdNow;
        pendingReveal.push(bubble);
        Promise.resolve().then(voiceTieAhead);
        if (revealTimer === null) {
          revealTimer = setTimeout(placeNewBubbles, 0);
        } else {
          placeAgain = true; // arrived during a move: place again before showing
        }
      }
      function lastPlacedChild() {
        var el = container.lastElementChild;
        while (el && (el.dataset.staged === "1" || el === typedRow)) { el = el.previousElementSibling; }
        return el;
      }
      function chatContentEnd() {
        var newest = lastPlacedChild();
        var typedHeight = newest && newest !== typedRow ? typedRow.offsetHeight : 0;
        return newest ? contentY(newest) + newest.offsetHeight + CHAT_TOP_GAP_PX + typedHeight - container.clientHeight : 0;
      }
      function roomBelowOperator(questionTop) {
        var current = parseFloat(typedRow.style.marginTop) || 0;
        var need = questionTop === null ? 0 : Math.max(0, Math.ceil(questionTop + container.clientHeight - (container.scrollHeight - current)));
        if (need !== current) { typedRow.style.marginTop = need ? need + "px" : ""; }
      }
      function placeTarget() {
        roomBelowOperator(null);
        var first = pendingReveal.length ? pendingReveal[0] : null;
        if (!first || !container.contains(first)) { return chatContentEnd(); }
        return bubbleTarget(first);
      }
      var CHAT_BATON_MAX_MS, SPEAKING_MAX_MS;
      function batonHeld() {
        if (narratorLines.some(function (n) { return n.shown; })) { return true; }
        return Array.prototype.some.call(container.querySelectorAll(".brobot-say.speaking"), function (el) {
          var b = el.closest(".brobot-bubble");
          if (!el._litAt) { el._litAt = Date.now(); }
          if (Date.now() - el._litAt > SPEAKING_MAX_MS) { return false; }
          return b && !b.classList.contains("narrator") && b.style.visibility !== "hidden";
        });
      }
      function bubbleTarget(first) {
        var last = lastPlacedChild();
        if (!last) { last = first; }
        var typedHeight = last !== typedRow ? typedRow.offsetHeight : 0;
        var room = container.clientHeight - typedHeight - 2 * CHAT_TOP_GAP_PX;
        var replyHeads = first.classList.contains("popup-wrap") ? first.querySelectorAll(":scope > .popup-system ~ .brobot-bubble-head") : [];
        var replyHead = replyHeads.length ? replyHeads[replyHeads.length - 1] : null;
        var top = contentY(replyHead || first);
        var tall = contentY(last) + last.offsetHeight - top;
        if (tall <= room) { scrollLog("target: fits (" + Math.round(tall) + " of " + Math.round(room) + ") - its bottom at the window's bottom"); return chatContentEnd(); }
        scrollLog("target: taller than the window (" + Math.round(tall) + " of " + Math.round(room) + ") - its top at the window's top");
        return top - CHAT_TOP_GAP_PX;
      }
      var BUBBLE_READY_MAX_MS;
      function whenBubblesReady(bubbles, then) {
        var from = performance.now();
        (function check() {
          var waiting = bubbles.some(function (b) {
            var imgs = Array.prototype.some.call(b.querySelectorAll("img"), function (img) { return !img.complete; });
            var voice = b._voice && !b._voice._failed && b._voice.readyState < 3;
            return imgs || voice;
          });
          if (waiting && performance.now() - from < BUBBLE_READY_MAX_MS) {
            setTimeout(check, 50);
            return;
          }
          then();
        })();
      }
      var batonNoted = 0;
      function batonHolder() {
        var held = Array.prototype.filter.call(container.querySelectorAll(".brobot-say.speaking"), function (el) {
          var b = el.closest(".brobot-bubble");
          return b && !b.classList.contains("narrator") && b.style.visibility !== "hidden" && !(el._litAt && Date.now() - el._litAt > SPEAKING_MAX_MS);
        })[0];
        if (held) { return "lit line: " + Array.from(String(held.textContent)).slice(0, 60).join("") + (held._litAt ? "" : " (no lit time)"); }
        if (narratorLines.some(function (n) { return n.shown; })) { return "narrator line"; }
        return "";
      }
      function placeNewBubbles() {
        var waitFrom = performance.now(), litNow = "";
        (function waitForBaton() {
          var lit = batonHolder();
          if (lit && lit !== litNow) { litNow = lit; waitFrom = performance.now(); }
          if (performance.now() - waitFrom > 10000 && Date.now() - batonNoted > 10000 && batonHeld()) {
            batonNoted = Date.now();
            var held = new URLSearchParams();
            held.set("note", "GOPOD_CHECK baton held 10 s+ by " + batonHolder());
            fetch("/api-sdk/gopod_page_note", { method: "POST", body: held }).catch(function () {});
          }
          if (batonHeld() && performance.now() - waitFrom < CHAT_BATON_MAX_MS) {
            revealTimer = setTimeout(waitForBaton, 100);
            return;
          }
          whenBubblesReady(pendingReveal.slice(), placeNewBubblesNow);
        })();
      }
      function warmupToBottom() {
        if (warmupBubble && warmupBubble.isConnected && container.lastElementChild !== warmupBubble) {
          container.appendChild(warmupBubble);
        }
      }
      var bubbleSizes = typeof ResizeObserver === "function" ? new ResizeObserver(function (entries) {
        entries.forEach(function (entry) {
          var b = entry.target;
          var h = Math.round(entry.contentRect.height);
          var was = b._seenHeight || 0;
          b._seenHeight = h;
          if (!was || h <= was || !b.isConnected || b.style.visibility === "hidden") { return; }
          if (pendingReveal.length) { scrollLog("grow skipped: " + scrollWho(b) + " (+" + (h - was) + "), " + pendingReveal.length + " bubbles waiting"); return; }
          if (b !== container.lastElementChild && b !== warmupBubble) { scrollLog("grow skipped: " + scrollWho(b) + " (+" + (h - was) + "), not the newest"); return; }
          var top = contentY(b) - container.scrollTop;
          if (top > chatViewHeight() || top + h < 0) { scrollLog("grow skipped: " + scrollWho(b) + " (+" + (h - was) + "), out of view"); return; }
          scrollLog("grow: " + scrollWho(b) + " (+" + (h - was) + ")");
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
          if (chatMove || performance.now() - started < CHAT_REPLY_DELAY_MS) {
            revealTimer = setTimeout(waitForPlace, 40);
            return;
          }
          revealTimer = null;
          if (operatorInBatch) { awaitingReply = true; }
          operatorInBatch = false;
          var shown = pendingReveal;
          if (shown.some(function (b) { return !b.classList.contains("operator"); })) {
            setTimeout(function () { awaitingReply = false; }, 0);
          }
          pendingReveal = [];
          shown.forEach(function (b) { b.style.visibility = ""; watchSize(b); });
          if (shown.length) {
            var firstShown = shown[0];
            Array.prototype.forEach.call(container.querySelectorAll(".brobot-bubble"), function (b) {
              if (b !== firstShown && (b.compareDocumentPosition(firstShown) & Node.DOCUMENT_POSITION_FOLLOWING)) { revealRest(b); }
            });
          }
        })();
      }

      var CHAT_REPLY_DELAY_MS;
      var CHAT_TOP_GAP_PX;
      var lastOperatorBubble = null;
      var CHAT_SCROLL_MS;
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
      function chatViewHeight() {
        return Math.max(1, container.clientHeight - (typedRow.isConnected && typedRow.offsetParent !== null ? typedRow.offsetHeight : 0));
      }
      function contentY(el) {
        var box = container.getBoundingClientRect();
        var rect = el.getBoundingClientRect();
        return rect.top - box.top + container.scrollTop;
      }
      var SCROLL_LOG = false;
      var scrollLogLast = "", scrollLogAt = 0;
      function scrollLog(what) {
        if (!SCROLL_LOG || restoringChat || (pageRoleKnown && !isMainPage)) { return; }
        var line = "scroll " + what;
        if (line === scrollLogLast && Date.now() - scrollLogAt < 1000) { return; }
        scrollLogLast = line; scrollLogAt = Date.now();
        var body = new URLSearchParams();
        body.set("note", line.slice(0, 200));
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: body }).catch(function () {});
      }
      function scrollCaller() {
        var lines = String(new Error().stack || "").split("\n");
        for (var i = 3; i < lines.length; i++) {
          var m = /at (?:Object\.)?([A-Za-z0-9_$]+)/.exec(lines[i]) || /^([A-Za-z0-9_$]+)@/.exec(lines[i]);
          if (m && m[1] !== "scrollChatTo") { return m[1]; }
        }
        return "?";
      }
      function scrollWho(el) {
        if (!el) { return "none"; }
        var head = el.querySelector(".brobot-bubble-head");
        var name = head ? head.textContent.trim().slice(0, 24) : (el.dataset.kind || el.className.split(" ")[1] || "?");
        return name + (el.style.visibility === "hidden" ? " (hidden)" : "") + " h=" + el.offsetHeight;
      }
      var scrollBar = { on: true, width_px: 12, min_thumb_px: 30, max_thumb_share: 0.8, handle_px: 7 };
      var sbTrack = null, sbThumb = null, sbThumbPx = 0, sbDrag = null;
      function sbKept() { try { return parseFloat(localStorage.getItem("gopod_scroll_thumb_px")) || 0; } catch (e) { return 0; } }
      function sbKeep(px) { try { if (px) { localStorage.setItem("gopod_scroll_thumb_px", String(Math.round(px))); } else { localStorage.removeItem("gopod_scroll_thumb_px"); } } catch (e) {} }
      function sbBuild() {
        if (sbTrack || !scrollBar.on) { return; }
        sbTrack = document.createElement("div");
        sbTrack.className = "gopod-scroll-track";
        sbThumb = document.createElement("div");
        sbThumb.className = "gopod-scroll-thumb";
        ["top", "bottom"].forEach(function (end) {
          var h = document.createElement("div");
          h.className = "gopod-scroll-handle " + end;
          h.dataset.end = end;
          sbThumb.appendChild(h);
        });
        sbTrack.appendChild(sbThumb);
        document.body.appendChild(sbTrack);
        document.documentElement.classList.add("gopod-own-scrollbar");
        document.documentElement.style.setProperty("--gopod-scroll-room", (scrollBar.width_px + 8) + "px");
        sbThumbPx = sbKept();
        sbThumb.addEventListener("pointerdown", function (ev) {
          var end = ev.target.dataset ? ev.target.dataset.end : "";
          sbDrag = { end: end || "", y: ev.clientY, top: parseFloat(sbThumb.style.top) || 0, len: sbThumb.offsetHeight };
          if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
          try { sbThumb.setPointerCapture(ev.pointerId); } catch (e) {}
          sbThumb.classList.add("dragging");
          ev.preventDefault();
          ev.stopPropagation();
        });
        sbThumb.addEventListener("pointermove", function (ev) {
          if (!sbDrag) { return; }
          var d = ev.clientY - sbDrag.y, track = sbTrack.clientHeight;
          if (sbDrag.end === "bottom") {
            sbThumbPx = Math.min(track - sbDrag.top, sbDrag.len + d);
          } else if (sbDrag.end === "top") {
            sbThumbPx = Math.min(sbDrag.top + sbDrag.len, sbDrag.len - d);
          } else {
            var travel = Math.max(1, track - sbThumb.offsetHeight);
            container.scrollTop = Math.max(0, Math.min(1, (sbDrag.top + d) / travel)) * chatMaxScroll();
          }
          sbUpdate();
        });
        var done = function () {
          if (!sbDrag) { return; }
          if (sbDrag.end) { sbKeep(sbThumbPx); }
          sbDrag = null;
          sbThumb.classList.remove("dragging");
        };
        sbThumb.addEventListener("pointerup", done);
        sbThumb.addEventListener("pointercancel", done);
        sbThumb.addEventListener("dblclick", function () { sbThumbPx = 0; sbKeep(0); sbUpdate(); });
        sbTrack.addEventListener("pointerdown", function (ev) {
          if (ev.target !== sbTrack) { return; }
          var above = ev.clientY < sbThumb.getBoundingClientRect().top;
          if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
          container.scrollTop += (above ? -1 : 1) * chatViewHeight() * 0.9;
        });
        container.addEventListener("scroll", sbUpdate, { passive: true });
        window.addEventListener("resize", sbUpdate);
        window.addEventListener("scroll", sbUpdate, { passive: true });
        if (typeof ResizeObserver === "function") { new ResizeObserver(sbUpdate).observe(container); }
        if (typeof MutationObserver === "function") { new MutationObserver(sbUpdate).observe(container, { childList: true, subtree: true }); }
      }
      function sbUpdate() {
        if (!sbTrack) { return; }
        var box = container.getBoundingClientRect();
        var max = chatMaxScroll();
        if (!scrollBar.on || box.height < 40 || max <= 0 || !container.classList.contains("expanded")) { sbTrack.style.display = "none"; return; }
        var inset = 10;
        sbTrack.style.display = "block";
        sbTrack.style.left = (box.right - scrollBar.width_px - 4) + "px";
        sbTrack.style.top = (box.top + inset) + "px";
        sbTrack.style.width = scrollBar.width_px + "px";
        sbTrack.style.height = Math.max(20, box.height - 2 * inset) + "px";
        var track = sbTrack.clientHeight;
        var natural = track * container.clientHeight / container.scrollHeight;
        var len = Math.max(scrollBar.min_thumb_px, Math.min(track * scrollBar.max_thumb_share, sbThumbPx || natural));
        if (len > track) { len = track; }
        var top = (track - len) * container.scrollTop / max;
        sbThumb.style.height = len + "px";
        sbThumb.style.top = top + "px";
        Array.prototype.forEach.call(sbThumb.children, function (h) { h.style.height = scrollBar.handle_px + "px"; });
      }
      var scrollRestTimer = null, scrollByCode = false, chatMovedAt = 0;
      container.addEventListener("scroll", function () {
        if (!SCROLL_LOG) { return; }
        if (chatMove || Date.now() - chatMovedAt < 500) { scrollByCode = true; }
        clearTimeout(scrollRestTimer);
        scrollRestTimer = setTimeout(function () {
          var max = chatMaxScroll(), at = container.scrollTop, end = chatContentEnd();
          scrollLog("position: " + Math.round(at) + " of max " + Math.round(max) +
            (max - at < 2 ? " (at the bottom)" : ", " + Math.round(max - at) + " px above the bottom") +
            ", newest bubble's end " + Math.round(end - at) + " px " + (end - at > 0 ? "below" : "in") + " view" +
            (scrollByCode ? ", moved by the chat" : ", moved by hand"));
          scrollByCode = false;
        }, 300);
      }, { passive: true });
      function scrollChatTo(top) {
        var target = Math.max(0, Math.min(top, chatMaxScroll()));
        if (SCROLL_LOG) {
          scrollLog("by " + scrollCaller() + ": " + Math.round(container.scrollTop) + " -> " + Math.round(target) +
            (target !== top ? " (asked " + Math.round(top) + ")" : "") + " of max " + Math.round(chatMaxScroll()) +
            ", view " + Math.round(chatViewHeight()) + ", newest " + scrollWho(lastPlacedChild()) + (pendingReveal.length ? ", " + pendingReveal.length + " waiting" : ""));
        }
        var reduce = false;
        try { reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches; } catch (e) {}
        if (reduce) {
          if (chatMove) { cancelAnimationFrame(chatMove.frame); chatMove = null; }
          container.scrollTop = target;
          chatMovedAt = Date.now();
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
          chatMovedAt = Date.now();
          chatMove = null;
          return;
        }
        container.scrollTop = Math.max(0, Math.min(at.pos, chatMaxScroll()));
        chatMovedAt = Date.now();
        chatMove.frame = requestAnimationFrame(chatMoveStep);
      }
      var FOCUS_DEAD_ZONE;
      function focusSpeaking(span) {
        if (!span || !container.contains(span)) { return; }
        var rect = span.getBoundingClientRect();
        var middle = contentY(span) + rect.height / 2;
        var view = chatViewHeight();
        var below = middle - (container.scrollTop + view / 2);
        if (below < view * FOCUS_DEAD_ZONE && !chatMove) { return; }
        scrollLog("highlight: lit line " + Math.round(below) + " px below the middle");
        var target = middle - view / 2;
        target = Math.min(target, chatContentEnd());
        if (target > container.scrollTop + 4) { scrollChatTo(target); }
      }

      var cueMarkerPattern = null;
      var placeholderPattern = null;
      var segmentMarker = null;
      var liveSentenceEnds = ".!?";
      var liveCommaEnds = ",";
      var countdownRole;
      var ballRole;
      function escapeForRegExp(text) {
        return String(text).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      }
      var pageTexts = {};
      function pageText(key, values) {
        var t = String(pageTexts[key] == null ? "" : pageTexts[key]);
        Object.keys(values || {}).forEach(function (k) { t = t.split("{" + k + "}").join(String(values[k])); });
        return t;
      }
      function loadBrobotsSettings() {
        return fetch("/gopod_brobots_settings.json", { cache: "no-store" })
          .then(function (r) { return r.json(); })
          .then(function (settings) {
            pageTexts = settings.page_text || {};
            typedPlaceholder = pageText("typed_placeholder");
            if (!typedText.value && !typedText.classList.contains("gopod-typed-note")) { typedText.placeholder = typedPlaceholder; }
            var allFirst = typedRow.querySelector(".gopod-all-first");
            if (allFirst) { allFirst.title = pageText("all_first_title"); }
            if (numpadLayoutSelect) { numpadLayoutSelect.title = pageText("numpad_layout_title"); }
            var submit = typedRow.querySelector(".gopod-typed-submit");
            if (submit) { submit.textContent = pageText("submit"); }
            if (settings.chat_window && typeof settings.chat_window.hourglass_sprites === "number") { HOURGLASS_SPRITES = settings.chat_window.hourglass_sprites; }
            var marker = String(settings.animation_marker || "");
            if (marker) {
              var last = escapeForRegExp(marker.slice(-1));
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
            var vp = settings.system_bubble && settings.system_bubble.voice_persona;
            var vpp = vp && (settings.host_personas || {})[vp];
            if (vpp) {
              var nameParts = String(vpp.name || "").trim().split(/\s+/);
              systemLabel = String(systemLabel)
                .replace("{first}", nameParts[0] || "").replace("{last}", nameParts.length > 1 ? nameParts[nameParts.length - 1] : "")
                .replace("{middle}", nameParts.slice(1, -1).join(" ")).replace("{name}", vpp.name || "")
                .replace("{nickname}", vpp.nickname || "").replace("{job}", vpp.job_name || "");
            }
            window.gopodSystemLabel = systemLabel;
            systemVoicePersona = vp || "";
            personaColours(settings.persona_colours || {});
            personaHeads(settings.persona_colours || {}, settings.persona_heads || {});
            personaVoiceOf = {};
            Object.keys(((settings.managers || {}).personas) || {}).forEach(function (k) {
              var p = settings.managers.personas[k] || {};
              personaVoiceOf[k] = p.voice_of || "";
            });
            if (settings.numpad_ptt) { numpad.cfg = settings.numpad_ptt; }
            if (settings.comm_check && settings.comm_check.record_seconds) { commCheckRecordSeconds = settings.comm_check.record_seconds; }
            if (settings.system_menu) { systemMenuData = settings.system_menu.groups || []; systemMenuIntro = settings.system_menu.intro || ""; systemMenuLook = settings.system_menu.look || {}; systemMenuCopy = settings.system_menu.copy || systemMenuCopy; }
            var runCards = (settings.game && settings.game.runs) || [];
            systemMenuData = systemMenuData.map(function (g) {
              if (!g.runs) { return g; }
              var copy = Object.assign({}, g);
              copy.items = runCards.filter(function (r) { return !r.pending; }).map(function (r) { return { cmd: r.start, about: r.menu_about, fill: r.start }; }).concat(g.items || []);
              return copy;
            });
            if (settings.math && settings.math.popup_title) { mathPopupTitle = settings.math.popup_title; }
            if (settings.digit_highlight) { digitHighlight = settings.digit_highlight; }
            if (settings.word_reveal) { wordReveal = settings.word_reveal; }
            if (settings.chat_window) {
              var cw = settings.chat_window;
              if (typeof cw.max_bubbles === "number") { CHAT_MAX_BUBBLES = cw.max_bubbles; }
              if (typeof cw.max_seen === "number") { CHAT_MAX_SEEN = cw.max_seen; }
              if (typeof cw.baton_max_ms === "number") { CHAT_BATON_MAX_MS = cw.baton_max_ms; }
              if (typeof cw.bubble_ready_max_ms === "number") { BUBBLE_READY_MAX_MS = cw.bubble_ready_max_ms; }
              if (typeof cw.reply_delay_ms === "number") { CHAT_REPLY_DELAY_MS = cw.reply_delay_ms; }
              if (typeof cw.top_gap_px === "number") { CHAT_TOP_GAP_PX = cw.top_gap_px; }
              if (typeof cw.scroll_ms === "number") { CHAT_SCROLL_MS = cw.scroll_ms; }
              if (typeof cw.scroll_log === "boolean") { SCROLL_LOG = cw.scroll_log; }
              if (cw.scroll_bar && typeof cw.scroll_bar === "object") {
                Object.keys(scrollBar).forEach(function (k) { if (typeof cw.scroll_bar[k] === typeof scrollBar[k]) { scrollBar[k] = cw.scroll_bar[k]; } });
                sbBuild();
                sbUpdate();
              }
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
              if (typeof cw.picture_retry_ms === "number") { pictureRetryMs = cw.picture_retry_ms; }
              if (typeof cw.picture_wait_max_ms === "number") { pictureWaitMaxMs = cw.picture_wait_max_ms; }
              if (settings.page_sync && typeof settings.page_sync.display_bubbles === "number") { DISPLAY_BUBBLES = settings.page_sync.display_bubbles; }
              if (settings.page_sync && typeof settings.page_sync.item_max_ms === "number") { feedItemMaxMs = settings.page_sync.item_max_ms; }
              if (settings.page_sync && typeof settings.page_sync.display_ms === "number") { DISPLAY_MS = settings.page_sync.display_ms; }
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
            if (settings.value_points) { scoreRole = settings.value_points.score_role || ""; }
            if (settings.value_points && settings.value_points.score_bar && settings.value_points.score_bar.done) {
              scoreBarDone = settings.value_points.score_bar.done;
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
            if (settings.bubble_sounds) { bubbleSounds = settings.bubble_sounds; }
            if (typeof settings.animation_lead_in === "string") { leadInText = settings.animation_lead_in.trim(); }
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
            if (settings.chat_moves) { chatMoves = settings.chat_moves; }
            if (settings.reporter_voices) {
              reporterRole = settings.reporter_voices.role || reporterRole;
              hostNames = settings.reporter_voices.hosts || hostNames;
              hostLabels = settings.reporter_voices.labels || hostLabels;
              var nickTpl = settings.reporter_voices.nickname_label || "";
              Object.keys(settings.host_personas || {}).forEach(function (h) {
                var hp = settings.host_personas[h] || {};
                if (nickTpl && hp.nickname) { hostLabels[h] = nickLabels[h] = nickTpl.replace("{nickname}", hp.nickname).replace("{name}", hp.name || h); }
              });
              displayNames = settings.display_names || displayNames;
              hostPersonasNow = settings.host_personas || hostPersonasNow;
              settingsRunCheck = settings.run_check || settingsRunCheck;
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
        if (cueMarkerPattern) { out = out.replace(cueMarkerPattern, ""); }
        return out.trim();
      }

      var lastRobotBubble = { who: "", norm: "", format: "" };
      function robotBubbleOnce(who, text, format, then) {
        var norm = String(text).toLowerCase().replace(/[^a-z0-9]+/g, "");
        if (!norm) { if (then) { then(); } return "no words"; }
        if (lastRobotBubble.who === who && lastRobotBubble.norm === norm && lastRobotBubble.format !== format) { if (then) { then(); } return "same words as the bubble before"; }
        lastRobotBubble = { who: who, norm: norm, format: format };
        appendBubble(who, text, format, undefined, undefined, then);
        return "";
      }
      function noteLiveReply(e, why) {
        var body = new URLSearchParams();
        body.set("note", "live reply " + e.id + " " + e.who + ": " + (why ? "NOT shown - " + why : "shown"));
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: body }).catch(function () { /* Wire-Pod gone */ });
      }

      var addressing = null;
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
      var SYSTEM_TARGET = "Goverlord";
      var systemLabel;
      function systemHead(label) {
        var sysHead = makeHead(label, label, [], true);
        sysHead.classList.add("system-head");
        return sysHead;
      }
      var BROBOT2_FIRST = /\b(?:start(?:ing)?\s+with|begin(?:ning)?\s+with)\s+(?:(?:brobot|bro\s*bo(?:ugh)?t|bro\s*but)\s*(?:2|two|too|to)|pip)\b|\b(?:(?:brobot|bro\s*bo(?:ugh)?t|bro\s*but)\s*(?:2|two|too|to)|pip)\s+(?:goes\s+|answers\s+)?first\b/i;
      function personaSlug(who) { return String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-"); }
      function headProp(name, ring, badge) {
        switch (name) {
          case "crown": return '<path d="M30 30 L36 14 L46 26 L50 10 L54 26 L64 14 L70 30 Z" fill="' + ring + '"/><circle cx="50" cy="10" r="3" fill="#fff"/>';
          case "bowtie": return '<path d="M38 82 L50 88 L62 82 L62 94 L50 88 L38 94 Z" fill="' + ring + '"/><g transform="translate(70 18) rotate(30)"><circle cx="0" cy="0" r="5" fill="none" stroke="' + ring + '" stroke-width="3"/><rect x="4" y="-1.5" width="14" height="3" fill="' + ring + '"/></g>';
          case "gem": return '<path d="M44 22 L56 22 L60 28 L50 38 L40 28 Z" fill="' + ring + '" stroke="#fff" stroke-width="1"/>';
          case "cap": return '<path d="M24 34 Q50 12 76 34 Z" fill="#11204d"/><rect x="20" y="32" width="60" height="6" rx="3" fill="#11204d"/><circle cx="50" cy="26" r="6" fill="' + ring + '"/><text x="50" y="29" font-size="8" font-family="sans-serif" font-weight="700" text-anchor="middle" fill="#fff">' + (badge || "") + '</text>';
          case "glasses": return '<g fill="none" stroke="' + ring + '" stroke-width="3"><circle cx="38" cy="52" r="10"/><circle cx="62" cy="52" r="10"/><path d="M48 52 L52 52"/></g>';
          case "partyhat": return '<path d="M38 30 L50 4 L62 30 Z" fill="' + ring + '"/><circle cx="50" cy="4" r="4" fill="#ffd54a"/><path d="M42 22 L58 22" stroke="#fff" stroke-width="2"/>';
          case "tiara": return '<path d="M34 28 L40 20 L45 26 L50 16 L55 26 L60 20 L66 28 Z" fill="' + ring + '"/>';
          case "headphones": return '<path d="M22 54 Q22 18 50 18 Q78 18 78 54" fill="none" stroke="' + ring + '" stroke-width="4"/><rect x="16" y="48" width="10" height="16" rx="4" fill="' + ring + '"/><rect x="74" y="48" width="10" height="16" rx="4" fill="' + ring + '"/>';
          case "shades": return '<g fill="' + ring + '"><path d="M28 46 L46 46 L44 58 L30 58 Z"/><path d="M54 46 L72 46 L70 58 L56 58 Z"/></g><path d="M46 49 L54 49" stroke="' + ring + '" stroke-width="3"/>';
          case "beanie": return '<path d="M26 36 Q50 6 74 36 Z" fill="#1f6feb"/><rect x="24" y="33" width="52" height="7" rx="3" fill="' + ring + '"/>';
          case "cash": return '<rect x="24" y="28" width="52" height="8" rx="4" fill="' + ring + '"/><text x="50" y="35" font-size="8" font-family="sans-serif" font-weight="700" text-anchor="middle" fill="#fff">$ $ $</text>';
          case "mic": return '<rect x="76" y="66" width="10" height="16" rx="5" fill="#d9d9d9"/><rect x="79" y="82" width="4" height="14" fill="#9a9a9a"/>';
        }
        return "";
      }
      function headSVG(c, props, badge) {
        var bg = c.background || "#333", ring = c.border || "#ccc";
        var eyes = props.indexOf("shades") >= 0 ? "" : '<circle cx="38" cy="52" r="' + (props.indexOf("glasses") >= 0 ? 3.5 : 4.5) + '" fill="' + ring + '"/><circle cx="62" cy="52" r="' + (props.indexOf("glasses") >= 0 ? 3.5 : 4.5) + '" fill="' + ring + '"/>';
        return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="54" r="42" fill="' + bg + '" stroke="' + ring + '" stroke-width="4"/>' + eyes +
          '<path d="M40 70 Q50 76 60 70" fill="none" stroke="' + ring + '" stroke-width="3" stroke-linecap="round"/>' +
          props.map(function (p) { return headProp(p, ring, badge); }).join("") + '</svg>';
      }
      function personaHeads(colours, heads) {
        var rules = [];
        Object.keys(heads).forEach(function (who) {
          var h = heads[who] || {};
          var c = colours[h.from || who] || {};
          var url = "url(\"data:image/svg+xml," + encodeURIComponent(headSVG(c, h.props || [], h.badge)) + "\")";
          var s = personaSlug(who);
          var sel = who === "system"
            ? ".brobot-bubble-avatar.to-system"
            : ".brobot-bubble.host-" + s + " > .brobot-bubble-head > .brobot-bubble-avatar:first-child:not(.from-operator):not([class*='to-brobot']), .brobot-bubble-avatar.head-" + s;
          rules.push(sel + " { background: " + url + " center / contain no-repeat; }");
        });
        var el = document.getElementById("gopod-persona-heads");
        if (!el) { el = document.createElement("style"); el.id = "gopod-persona-heads"; document.head.appendChild(el); }
        el.textContent = rules.join("\n");
        window.gopodPersonaHeads = heads;
      }
      function personaColours(colours) {
        var rules = [];
        var root = document.documentElement.style;
        var each = [];
        Object.keys(colours).forEach(function (who) {
          var c = colours[who] || {};
          each.push([who, c]);
          if (/^Brobot [12]$/.test(who)) {
            var n = who === "Brobot 2" ? "2" : "1";
            if (c.border) { root.setProperty("--brobot-" + n + "-color", c.border); }
            if (c.background) { root.setProperty("--brobot-" + n + "-bg", c.background); }
          }
        });
        each.forEach(function (pair) {
          var who = pair[0], c = pair[1];
          if (!c.background && !c.border) { return; }
          var s = personaSlug(who), b = ".brobot-bubble.host-" + s;
          var body = [];
          if (c.background) { body.push("background: " + c.background); }
          if (c.text) { body.push("color: " + c.text); }
          if (c.border) { body.push("box-shadow: 0 0 0 var(--bubble-border-width) " + c.border); }
          rules.push(b + " { " + body.join("; ") + "; }");
          if (c.border) { rules.push(".host-" + s + " .brobot-bubble-avatar:first-child, .brobot-bubble .host-" + s + " .brobot-bubble-avatar:first-child, " + b + " > .brobot-bubble-head > .brobot-bubble-avatar:first-child { --eye-fill: " + c.border + "; }"); }
          if (c.background) {
            rules.push(".suit-" + s + " { --suit-border: " + c.background + "; }");
            rules.push(".brobot-bubble.search-results.host-" + s + " { background: " + c.background + "; }");
          }
        });
        var el = document.getElementById("gopod-persona-colours");
        if (!el) { el = document.createElement("style"); el.id = "gopod-persona-colours"; document.head.appendChild(el); }
        el.textContent = rules.join("\n");
      }
      function brobotAs(who) { return brobotTo[who] || brobotBy[who] || who; }
      function brobotHostClass(shell, who, as) {
        as = as || brobotAs(who);
        Array.prototype.slice.call(shell.classList).forEach(function (c) { if (/^host-/.test(c)) { shell.classList.remove(c); } });
        shell.classList.add("host-" + personaSlug(as));
        var w = shell.querySelector(":scope > .brobot-bubble-head .brobot-bubble-who");
        if (w) { w.textContent = as + ":"; }
        var ic = shell.querySelector(":scope > .brobot-bubble-head .brobot-bubble-avatar");
        if (ic) { headIconClass(ic, as); }
      }
      function operatorHead(to, via, text) {
        var senderKey = headNameKey(via);
        var sender = senderKey ? via : HEAD_OPERATOR;
        var made = operatorHeadMade(to, via, text, sender);
        if (senderKey && !/^Brobot [12]$/.test(senderKey)) { made.dataset.sender = senderKey; }
        return made;
      }
      function operatorHeadMade(to, via, text, sender) {
        if (to === SYSTEM_TARGET) {
          return makeHead(sender, keyShown(via || "Operator") + " to " + (systemLabel || to), [HEAD_SYSTEM], true, via);
        }
        var toShown = brobotTo[to] || (to && to === systemVoicePersona ? systemLabel : to);
        var systemCmd = /^\s*system-/i.test(String(text || ""));
        if (systemCmd) { toShown = systemLabel || toShown; }
        var text0 = keyShown(via || "Operator") + " to " + toShown;
        if (systemCmd) { return makeHead(sender, text0, [HEAD_SYSTEM], true, via); }
        var receivers = to === "Brobots" ? ["Brobot 1", "Brobot 2"] : [to];
        if (to === "Brobots" && (BROBOT2_FIRST.test(String(text || "")) || firstNamedBrobot(text) === "Brobot 2" || ((allSecondFirst[String(text || "")] || (/^KP/.test(String(via || "")) && allFirstSelect && allFirstSelect.value === "-*")) && firstNamedBrobot(text) !== "Brobot 1"))) { receivers.reverse(); }
        if (receivers.length === 1 && brobotTo[to]) { receivers = [brobotTo[to]]; }
        return makeHead(sender, text0, receivers, true, via);
      }

      function reporterHead(label) {
        var head = makeHead(label, label, [], true);
        head.classList.add("system-head");
        return head;
      }
      function bubbleHead(who, noColon) {
        var shown = brobotBy[who] || ((hostPersonasNow || {})[who] || {}).call_name || who;
        return makeHead(brobotBy[who] || who, noColon ? shown : shown + ":", [], true);
      }
      var HEAD_OPERATOR = "\u0000operator", HEAD_SYSTEM = "\u0000system";
      function makeHead(from, text, to, noColon, via) {
        var head = document.createElement("div");
        head.className = "brobot-bubble-head";
        head.appendChild(headIcon(from, via));
        var name = document.createElement("span");
        name.className = "brobot-bubble-who";
        name.textContent = noColon ? String(text) : String(text) + ":";
        head.appendChild(name);
        (to || []).forEach(function (r) { head.appendChild(headIcon(r, "", true)); });
        return head;
      }
      function headIcon(name, via, receiver) {
        var icon = document.createElement("div");
        icon.className = "brobot-bubble-avatar";
        headIconClass(icon, name, via, receiver);
        return icon;
      }
      function headIconClass(icon, name, via, receiver) {
        Array.prototype.slice.call(icon.classList).forEach(function (c) { if (c !== "brobot-bubble-avatar") { icon.classList.remove(c); } });
        if (name === HEAD_OPERATOR) {
          icon.classList.add("from-operator");
          if (/^(KP|Numpad)/i.test(String(via || ""))) { icon.classList.add("via-numpad"); }
          else if (/mic|heard|voice/i.test(String(via || ""))) { icon.classList.add("via-mic"); }
          else if (/keyboard|typed/i.test(String(via || ""))) { icon.classList.add("via-keyboard"); }
          return;
        }
        if (name === HEAD_SYSTEM) { icon.classList.add("to-system"); return; }
        var m = /^Brobot ([12])$/.exec(String(name || ""));
        if (m) { icon.classList.add("to-brobot-" + m[1]); return; }
        var k = headNameKey(name);
        var mk = /^Brobot ([12])$/.exec(k);
        if (mk) { icon.classList.add("to-brobot-" + mk[1]); return; }
        if (k && (window.gopodPersonaHeads || {})[k]) { icon.classList.add("head-" + personaSlug(k)); return; }
        icon.classList.add("to-system");
      }
      function headNameKey(label) {
        var norm = function (x) { return String(x || "").toLowerCase().replace(/[\u2013\u2014]/g, "-").replace(/[^a-z0-9]+/g, " ").trim(); };
        var n = norm(label);
        if (!n) { return ""; }
        var heads = window.gopodPersonaHeads || {}, hp = hostPersonasNow || {};
        var cands = [];
        Object.keys(heads).forEach(function (k) { cands.push([norm(k), k]); });
        Object.keys(hp).forEach(function (k) { var p = hp[k] || {}; [k, p.call_name, p.nickname, p.name].forEach(function (x) { if (x) { cands.push([norm(x), k]); } }); });
        if (systemLabel && systemVoicePersona) { cands.push([norm(systemLabel), systemVoicePersona]); }
        cands.sort(function (a, b) { return b[0].length - a[0].length; });
        for (var i = 0; i < cands.length; i++) {
          var c = cands[i][0];
          if (c && (n === c || n.indexOf(c + " ") === 0)) {
            var k = cands[i][1];
            if (/\breporter\b/.test(n) && heads[k + " reporter"]) { return k + " reporter"; }
            return k;
          }
        }
        return "";
      }

      var countdownEndText = "done";
      function endCountdowns() {
        countdownDoneAt = 0;
        Array.prototype.slice.call(container.querySelectorAll(".brobot-bubble.countdown")).forEach(function (b) { b.remove(); });
      }
      var COUNTDOWN_HOLD_MS, COUNTDOWN_HOLD_MOST_MS, countdownDoneAt = 0;
      function countdownHold() {
        if (!container.querySelector(".brobot-bubble.countdown")) { countdownDoneAt = 0; return; }
        countdownDoneAt = Date.now();
        var at = countdownDoneAt;
        setTimeout(function () { if (countdownDoneAt === at) { endCountdowns(); } }, COUNTDOWN_HOLD_MOST_MS);
      }
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
      function showCountdown(seconds, sentMs, then) {
        var next = function () { if (then) { then(); } };
        if (!(seconds > 0)) { next(); return; }
        var leftMs = (seconds + 10) * 1000 - Math.max(0, Date.now() - (sentMs || Date.now()));
        if (leftMs <= 0) { next(); return; }
        var bubble = rawBubble({ kind: "countdown", classes: "operator countdown", head: systemHead(systemLabel) });
        var row = document.createElement("div");
        row.className = "warmup-timer";
        bubble.appendChild(row);
        addJuggleTimer(bubble, row);
        bubble._gone = setTimeout(function () { bubble.remove(); }, leftMs);
        placeBubble(bubble, then);
      }

      function addJuggleTimer(bubble, box) {
        var juggle = document.createElement("div");
        juggle.className = "countdown-juggle";
        ["B", "I", "N", "G", "O"].forEach(function (letter) {
          var ball = document.createElement("span");
          ball.className = "juggle-ball ball-" + letter.toLowerCase();
          ball.textContent = letter;
          juggle.appendChild(ball);
        });
        box.appendChild(juggle);
        var jugBalls = juggle.querySelectorAll(".juggle-ball");
        var glass = document.createElement("span");
        glass.className = "countdown-hourglass";
        glass.textContent = "\u23F3";
        box.appendChild(glass);
        var jugFrame = 0;
        var jugTimer = setInterval(function () {
          if (!bubble.isConnected && jugFrame > 0) { clearInterval(jugTimer); return; }
          var ballSize = jugBalls.length ? jugBalls[0].offsetWidth : 56;
          var step = Math.max(0, juggle.clientWidth - ballSize) / Math.max(1, jugBalls.length - 1);
          var travel = Math.max(0, juggle.clientHeight - ballSize);
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

      var warmupBubble = null;
      var warmupBeforeOpening = null, openingDone = false;
      var warmupPolling = false;
      var llmWarm = { warmText: "LLM Warming Up", readyText: "LLM Ready", waitText: "Please wait.", checkMs: 3000, expectedSeconds: 120, leds: 10 };
      function buildWarmupBubble(startedMs) {
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator warmup";
        bubble.dataset.started = String(startedMs);
        bubble.appendChild(systemHead(systemLabel));
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
      function showWarmupLevel(bubble, lit, ready, readyText, warmingText) {
        var label = ready ? (readyText || llmWarm.readyText) : (warmingText || llmWarm.warmText);
        var labelEl = bubble.querySelector(".brobot-bubble-text");
        if (voiceLineWords(labelEl.textContent) !== voiceLineWords(label)) {
          labelEl.textContent = label;
          String(label).split("\n").forEach(function (one) { voiceLineAhead(one, labelEl); });
          Object.keys(bubble._spoken || {}).forEach(function (t) {
            var said = voiceLineSpan(labelEl, t);
            if (said) { said.classList.add("spoken"); }
          });
          voiceTieAhead();
        }
        labelEl.style.whiteSpace = "pre-line";
        var leds = bubble.querySelectorAll(".llm-led");
        for (var i = 0; i < leds.length; i++) { leds[i].classList.toggle("lit", i < lit); }
        bubble.querySelector(".llm-battery").setAttribute("aria-label", label + ", " + lit + " of " + leds.length);
        bubble.classList.toggle("ready", !!ready);
        var sprites = HOURGLASS_SPRITES - 1;
        var stage = ready ? sprites : Math.min(sprites, Math.floor(lit * HOURGLASS_SPRITES / Math.max(1, leds.length)));
        var glass = bubble.querySelector(".warmup-hourglass");
        if (glass && glass.dataset.stage !== String(stage)) {
          glass.dataset.stage = String(stage);
          glass.innerHTML = hourglassSprite(stage);
        }
      }
      var HOURGLASS_SPRITES;
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
      function restoreWarmupBubble(saved) {
        var parts = String(saved).split("|");
        var warm = buildWarmupBubble(parts[0]);
        showWarmupLevel(warm, llmWarm.leds, true, parts.slice(1).join("|"));
        var label = warm.querySelector(".brobot-bubble-text");
        String(label.textContent).split("\n").forEach(function (one) {
          if (!one.trim()) { return; }
          var said = voiceLineSpan(label, one);
          if (said) { said.classList.add("spoken"); }
        });
      }
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
        if (!ms || isSeen("warm|" + ms)) { return false; }
        if (warmupBubble && !warmupBubble.isConnected) { warmupBubble = null; }
        if (!warmupBubble || warmupBubble.dataset.started !== ms) { warmupBubble = buildWarmupBubble(ms); }
        warmupBubble.dataset.text = message;
        var lit = warmupBubble.querySelectorAll(".llm-led.lit").length;
        showWarmupLevel(warmupBubble, Math.max(1, lit), false, "", message);
        return message.trim() !== llmWarm.waitText.trim();
      }
      function pollWarmup() {
        if (!isMainPage) { return; } // the display mirror shows a viewer's chat
        if (warmupPolling) { return; }
        warmupPolling = true;
        fetch("/gopod_llm_warmup.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (state) {
            if (!state || !state.started_ms) { return; }
            if (warmupBubble && !warmupBubble.isConnected) { warmupBubble = null; }
            if (isSeen("warm|" + state.started_ms)) { return; }
            var mine = warmupBubble && warmupBubble.dataset.started === String(state.started_ms);
            if (!state.active) {
              if (mine) { showWarmupLevel(warmupBubble, llmWarm.leds - 1, false, "", warmupBubble.dataset.text); }
              return;
            }
            if (!mine) { warmupBubble = buildWarmupBubble(state.started_ms); }
            var level = Math.max(0, Date.now() - state.started_ms) / (llmWarm.expectedSeconds * 1000);
            var shownText = state.awaits && state.text ? state.text : warmupBubble.dataset.text;
            showWarmupLevel(warmupBubble, Math.min(llmWarm.leds - 1, 1 + Math.floor(level * llmWarm.leds)), false, "", shownText);
          })
          .catch(function () { /* no warm-up file (older Wire-Pod) */ })
          .then(function () { warmupPolling = false; });
      }

      var chapterRole, chapterLabel;
      function showChapter(title, then) {
        if (!restoringChat) { recordBubble(chapterRole, String(title), "chapter"); }
        placeBubble(rawBubble({ kind: "chapter", classes: "operator system-status chapter", head: systemHead(chapterLabel.replace("{title}", String(title))) }), then);
      }
      var scoreRole;
      var systemRole;
      var terminalRole;
      var displayHost = "";
      function displayHead(bubble) { return hostHeadFor(bubble, displayHost); }
      function hostHeadFor(bubble, who) {
        if (!who) { return systemHead(systemLabel); }
        bubble.classList.add("host-" + who.toLowerCase().replace(/[^a-z0-9]+/g, "-"));
        return reporterHead(hostHeader(who));
      }
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
      var pictureRetryMs = 1000, pictureWaitMaxMs = 30000;
      function pictureWait(b, img) {
        var spot = document.createElement("span");
        spot.className = "picture-wait";
        spot.appendChild(generatingGlass(b));
        img.parentNode.insertBefore(spot, img);
        img.classList.add("picture-waiting");
        var src = img.getAttribute("src"), since = Date.now(), tries = 0;
        function there() {
          spot.remove();
          img.classList.remove("picture-waiting");
          if (b.isConnected && b === lastPlacedChild()) { scrollChatTo(bubbleTarget(b)); }
        }
        img.addEventListener("load", there);
        function again() {
          if (Date.now() - since >= pictureWaitMaxMs || !b.isConnected) {
            spot.remove();
            img.remove();
            return;
          }
          setTimeout(function () { img.src = src + (src.indexOf("?") < 0 ? "?" : "&") + "try=" + (++tries); }, pictureRetryMs);
        }
        img.addEventListener("error", again);
        if (img.complete) { if (img.naturalWidth) { there(); } else { again(); } }
      }
      function bubbleTask(b, promise) { (b._tasks = b._tasks || []).push(promise); return promise; }
      function bubbleNote(b, why) {
        if (restoringChat || (pageRoleKnown && !isMainPage) || !b || !b.classList) { return; }
        var head = b.querySelector(":scope > .brobot-bubble-head");
        var hosts = Array.prototype.filter.call(b.classList, function (c) { return /^host-/.test(c); }).join(" ");
        var cs = window.getComputedStyle(b);
        var line = "bubble " + (why || "shown") + " " + (b.dataset.feed ? "#" + b.dataset.feed + " " : "") +
          "head=\"" + (head ? head.textContent.replace(/\s+/g, " ").trim() : "") + "\" " +
          "persona=" + (hosts || "-") + " bg=" + cs.backgroundColor + " ring=" + (cs.boxShadow || "none").replace(/\s+0px/g, "");
        var body = new URLSearchParams();
        body.set("note", line.slice(0, 400));
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: body }).catch(function () {});
      }
      function placeBubble(b, then, onPlaced) {
        var finished = false;
        function finish() {
          if (finished) { return; }
          finished = true;
          container.appendChild(b);
          bubbleNote(b, "shown");
          if (onPlaced) { onPlaced(b); }
          if (!restoringChat) { revealWhenPlaced(b); }
          if (then) { then(); }
        }
        if (restoringChat) { finish(); return; }
        var stage = bubbleStage();
        stage.style.width = container.clientWidth + "px";
        stage.appendChild(b);
        Array.prototype.forEach.call(b.querySelectorAll("img"), function (img) {
          if (!img.complete || !img.naturalWidth) { pictureWait(b, img); }
        });
        var tasks = (b._tasks || []).map(function (t) { return Promise.resolve(t).catch(function () {}); });
        Promise.race([Promise.all(tasks), new Promise(function (ok) { setTimeout(ok, BUBBLE_TASK_MAX_MS); })]).then(finish);
      }
      function resultTemplate(text) {
        var out = document.createElement("div");
        out.className = "gopod-result";
        var lines = String(text).replace(/\s+$/, "").split("\n"), table = null, titled = false;
        lines.forEach(function (raw) {
          var line = raw.replace(/\s+$/, "");
          if (!line.trim()) { table = null; return; }
          if (!titled) {
            titled = true;
            var title = document.createElement("div");
            title.className = "gopod-result-title";
            title.textContent = line.trim();
            out.appendChild(title);
            return;
          }
          var cols = /^\s*(\S.*\S)\s{2,}(\S.*)$/.exec(line);
          if (cols) {
            if (!table) {
              table = document.createElement("div");
              table.className = "gopod-result-table";
              out.appendChild(table);
            }
            var cmd = document.createElement("code");
            cmd.className = "gopod-result-cmd";
            cmd.textContent = cols[1];
            var about = document.createElement("span");
            about.className = "gopod-result-about";
            about.textContent = cols[2];
            table.appendChild(cmd);
            table.appendChild(about);
            return;
          }
          table = null;
          var row = document.createElement("div");
          row.className = "gopod-result-line";
          var indent = /^\s*/.exec(line)[0].length;
          if (indent) { row.style.paddingLeft = Math.min(indent, 8) * 0.5 + "em"; }
          row.textContent = line.trim();
          out.appendChild(row);
        });
        return out;
      }
      function showTerminal(text, then) {
        if (systemMenuResult(String(text), "answer")) { if (then) { then(); } return; }
        var b = rawBubble({ kind: "terminal", classes: "operator system-status score terminal gopod-admin-only",
          head: systemHead(systemLabel), text: "", lineClass: "terminal-text" });
        var line = b.querySelector(".terminal-text");
        if (line) { line.appendChild(resultTemplate(text)); }
        placeBubble(b, then);
      }
      var searchSettings = { results_role: "Search Results" };
      function showSearchResults(text, to, then) {
        var next = function () { if (then) { then(); } };
        var bubble = document.createElement("div");
        var nestTo = /^Brobot [12]$/.test(String(to || "")) && !restoringChat ? to : "";
        bubble.className = "brobot-bubble search-results";
        var resultsHost = searchSettings.results_host || displayHost;
        if (resultsHost) { bubble.appendChild(hostHeadFor(bubble, resultsHost)); }
        var line = document.createElement("span");
        line.className = "brobot-bubble-text search-results-text";
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
        if (restoringChat) { if (!nestTo && !inMenu) { container.appendChild(bubble); } bubble.classList.add("typed"); next(); return; }
        if (!inMenu && !nestTo) { placeBubble(bubble, then); } else { next(); }
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
        whenShown(bubble, function () { requestAnimationFrame(step); });
      }
      function linesInto(bubble, text) {
          String(text).slice(6).split("\n").forEach(function (lineText) {
            var el = document.createElement("span");
            el.className = "brobot-bubble-text brobot-bubble-line";
            bubble.appendChild(el);
            lineText.split("\u2063").forEach(function (one, i) {
              if (i > 0) { el.appendChild(document.createTextNode(" ")); }
              var say = document.createElement("span");
              say.className = "brobot-say";
              say.textContent = one;
              el.appendChild(say);
              if (restoringChat || skippingReporters || !one.trim()) { say.classList.add("spoken"); return; }
              voiceLineAhead(one, say);
              showOnItsTurn(bubble, say);
            });
          });
          if (!restoringChat) { bubble.classList.add("voice-lit"); voiceTieAhead(); }
      }
      function showScore(text, label, then) {
        if (!restoringChat) { recordBubble(scoreRole, String(text), "score"); }
        var bubble = rawBubble({ kind: "system", classes: "operator system-status score", head: systemHead(label || systemLabel) });
        if (String(text).indexOf("lines:") === 0) {
          linesInto(bubble, text);
          placeBubble(bubble, then);
          return;
        }
        var line = document.createElement("span");
        line.className = "brobot-bubble-text";
        bubble.appendChild(line);
        voiceLitLine(bubble, line, text);
        placeBubble(bubble, then);
      }
      var voiceLitNext = null, voiceLitWaitMs = 15000;
      var lineWaits = false, lineVoice = "", voiceWaits = [], voiceWaitFor = {};
      function voiceWaitOn(bubble, el, file) {
        var wait = document.createElement("span");
        wait.className = "voice-wait";
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
      function voiceWaitsGone() {
        voiceWaits.slice().forEach(function (say) { voiceWaitEnd(say); saidWhole(say); });
      }
      function lineTurnCome(bubble, say) {
        var lines = bubble.querySelectorAll(".brobot-say");
        for (var i = 0; i < lines.length && lines[i] !== say; i++) {
          if (!lines[i].classList.contains("spoken")) { return false; }
        }
        return true;
      }
      function showOnItsTurn(bubble, say) {
        var shownAt = 0;
        (function showAnyway() {
          if (!bubble.isConnected || say.classList.contains("speaking") || say.classList.contains("spoken")) { return; }
          if (say._wait && say._wait.isConnected) { shownAt = 0; setTimeout(showAnyway, 500); return; }
          if (bubble.style.visibility !== "hidden" && bubble.dataset.staged !== "1" && lineTurnCome(bubble, say)) {
            if (!shownAt) { shownAt = Date.now(); }
            if (Date.now() - shownAt >= voiceLitWaitMs) {
              saidWhole(say);
              if (voiceLitNext === say) { voiceLitNext = null; }
              return;
            }
          } else {
            shownAt = 0;
          }
          setTimeout(showAnyway, 500);
        })();
      }
      function saidWhole(say) {
        say.classList.add("spoken");
        (say._parts || []).forEach(function (p) { p.classList.add("spoken"); });
      }
      function voiceLitLine(bubble, line, text) {
        line.textContent = "";
        var say = document.createElement("span");
        say.className = "brobot-say";
        say.innerHTML = flagPictures(escapeHtml(String(text)));
        line.appendChild(say);
        var waits = lineWaits;
        lineWaits = false;
        if (lineVoice) { voiceWaitFor[lineVoice] = say; lineVoice = ""; }
        if (restoringChat) { say.classList.add("spoken"); return; }
        if (skippingReporters) { say.classList.add("spoken"); bubble.classList.add("voice-lit"); return; }
        bubble.classList.add("voice-lit");
        voiceLitNext = say;
        if (waits) { voiceWaitOn(bubble, say, typeof waits === "string" ? waits : ""); }
        showOnItsTurn(bubble, say);
      }
      function voiceLitStart(seconds, words, src, audio) {
        var path = String(src || "").replace(/#.*$/, "").replace(/^https?:\/\/[^/]+/, "");
        var own = path && voiceWaitFor[path];
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
        say._voice = audio || null;
        say._voiceFrom = 0;
        var parts = (say._parts || []).filter(function (p) { return p.isConnected && !p.classList.contains("spoken"); });
        if (parts.length) {
          var counts = [say].concat(parts).map(function (p) { return Math.max(1, sayWords(p.textContent).length); });
          var total = counts.reduce(function (a, c) { return a + c; }, 0);
          var real = [];
          (words || []).forEach(function (w, k) { if (/[A-Za-z0-9]/.test(String(w.t || ""))) { real.push(k); } });
          var cut = function (n) { return n < real.length ? real[n] : (words || []).length; };
          var timed = real.length >= total;
          if (timed) { say._spokenWords = words.slice(0, cut(counts[0])); }
          else { say.dataset.end = String(seconds * counts[0] / total); }
          var from = counts[0];
          parts.forEach(function (p, i) {
            var n = counts[i + 1], startAt = seconds * from / total;
            p.dataset.at = "0";
            if (timed) {
              var mine = words.slice(cut(from), i === parts.length - 1 ? words.length : cut(from + n));
              startAt = +mine[0].s || startAt;
              p._spokenWords = mine.map(function (w) { return { t: w.t, s: (+w.s || 0) - startAt, e: (+w.e || 0) - startAt }; });
              p.dataset.end = String(Math.max(0.1, (+mine[mine.length - 1].e || 0) - startAt));
            } else {
              p.dataset.end = String(seconds * n / total);
            }
            p._voice = audio || null;
            p._voiceFrom = startAt;
            voiceAt(audio, startAt, function () { if (!p.classList.contains("spoken")) { lightSpan(p); } });
            from += n;
          });
        }
        lightSpan(say);
        return function () {
          say.classList.remove("speaking");
          say.classList.add("spoken");
          var saidIn = say.closest(".brobot-bubble");
          if (saidIn) { (saidIn._spoken = saidIn._spoken || {})[voiceLineWords(say.textContent)] = true; }
          parts.forEach(function (p) { p.classList.remove("speaking"); p.classList.add("spoken"); });
        };
      }
      var pictureRole, picturesInQuestion = [];
      var eggZoomText;
      var popupWraps = {};
      var popupExchange = false;
      var eggMarked = {}; // Brobot -> his line hit an egg (eggmark:)
      var popupText;
      var mathPopupTitle;
      function showPictures(list, then) {
        var next = function () { if (then) { then(); } };
        voiceLitNext = null; // the next voice is this egg's, not an earlier line's
        if (!restoringChat) { recordBubble(pictureRole, String(list), "picture"); }
        if (String(list).indexOf("popup:") === 0) { eggBubble(eggSpec(String(list).slice("popup:".length)), then); return; }
        var bubble = document.createElement("div");
        bubble.className = "brobot-bubble operator system-status pictures";
        var row = document.createElement("div");
        row.className = "topic-pictures";
        var files = String(list).split("|").filter(function (src) { return src.trim(); });
        var question = container.lastElementChild;
        if (warmupBubble && question === warmupBubble) { question = question.previousElementSibling; }
        var inQuestion = files.length && files.every(function (src) {
          return picturesInQuestion.indexOf(src.trim().split("/").pop()) >= 0;
        }) && question && question.classList.contains("operator") && question.dataset.own === "1";
        var weatherSound = prizeItems(list, bubble, row);
        if (inQuestion) {
          row.classList.add("in-question");
          question.appendChild(row);
          if (restoringChat) { next(); return; }
          if (question.style.visibility === "hidden") { placeAgain = true; next(); return; }
          whenBubblesReady([question], function () { scrollChatTo(bubbleTarget(question)); });
          next();
          return;
        }
        if (row.children.length) { bubble.appendChild(row); }
        placeBubble(bubble, then, function (b) {
          if (weatherSound) { weatherSoundOnShow(b, weatherSound); }
        });
      }
      function prizeItems(list, holder, row) {
        var weatherSound = "";
        String(list).split("|").forEach(function (src) {
          if (!src.trim()) { return; }
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
          magnify(row, img);
          row.appendChild(img);
        });
        return weatherSound;
      }
      function eggSpec(list) {
        var spec = { who: "", host: "", text: popupText, zoom: "", prize: list };
        var m = /^(Brobot [12])::/.exec(spec.prize);
        if (m) { spec.who = m[1]; spec.prize = spec.prize.slice(m[0].length); }
        m = !spec.who ? /^([A-Za-z][A-Za-z0-9 ]*)::/.exec(spec.prize) : null;
        if (m && m[1] !== "msg=" && hostNames[m[1]]) { spec.host = m[1]; spec.prize = spec.prize.slice(m[0].length); }
        m = /^msg=([^]*?)::/.exec(spec.prize);
        if (m) { spec.text = m[1]; spec.prize = spec.prize.slice(m[0].length); }
        if (eggZoomText && spec.text.trim().slice(-eggZoomText.length) === eggZoomText) {
          spec.zoom = eggZoomText;
          spec.text = spec.text.trim().slice(0, -eggZoomText.length).trim();
        }
        return spec;
      }
      function eggBubble(spec, then) {
        var b, holder, line;
        if (spec.who) {
          b = rawBubble({ kind: "egg", classes: (spec.who === "Brobot 1" ? "brobot-1" : "brobot-2") + " popup-winner popup-wrap" });
          holder = document.createElement("div");
          holder.className = "popup-system";
          holder.appendChild(systemHead(systemLabel));
          line = document.createElement("span");
          line.className = "popup-congrats";
          b.appendChild(holder);
          popupExchange = true;
          if (!popupWraps[spec.who]) { popupWraps[spec.who] = b; }
        } else {
          b = rawBubble({ kind: "egg", classes: "operator system-status pictures popup-winner" + (spec.host ? " host-" + spec.host.toLowerCase() : ""),
            head: systemHead(systemLabel) });
          holder = b;
          line = document.createElement("span");
          line.className = "brobot-bubble-text";
        }
        line.style.whiteSpace = "pre-line";
        holder.appendChild(line);
        lineWaits = true;
        voiceLitLine(b, line, spec.text);
        var say = line.querySelector(".brobot-say");
        var row = document.createElement("div");
        row.className = "topic-pictures";
        var weatherSound = prizeItems(spec.prize, holder, row);
        if (row.children.length) { holder.appendChild(row); }
        var words = spec.text;
        if (spec.zoom && row.children.length) {
          var zoomEl = document.createElement("span");
          zoomEl.className = spec.who ? "popup-congrats popup-zoom-line" : "brobot-bubble-text popup-zoom-line";
          var zoomSay = document.createElement("span");
          zoomSay.className = "brobot-say";
          zoomSay.textContent = spec.zoom;
          zoomEl.appendChild(zoomSay);
          holder.appendChild(zoomEl);
          if (!say || restoringChat || skippingReporters || say.classList.contains("spoken")) { zoomSay.classList.add("spoken"); }
          else { say._parts = [zoomSay]; }
          words += " " + spec.zoom;
        }
        if (say && !restoringChat) { voiceLineAhead(words, say); }
        placeBubble(b, then, function (placed) {
          marqueeAttach(placed);
          if (weatherSound) { weatherSoundOnShow(placed, weatherSound); }
        });
      }
      var MARQUEE_SPACING_PX, MARQUEE_INSET_PX, MARQUEE_PATTERN_MS;
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
        marqueeTrack(bubble, layer);
      }
      function marqueeTrack(bubble, layer) {
        var m = { bubble: bubble, layer: layer, lights: [] };
        marquees.push(m);
        if (marqueeSizes) { marqueeSizes.observe(bubble); }
        marqueeLayout(m);
        if (marqueeFrame === null) { marqueeFrame = requestAnimationFrame(marqueeTick); }
      }
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
        var still = false;
        if (now - marqueeSwitch > MARQUEE_PATTERN_MS) {
          marqueePattern = (marqueePattern + 1) % marqueePatterns.length;
          marqueeSwitch = now;
        }
        var t = now / 1000, fn = marqueePatterns[marqueePattern];
        marquees.forEach(function (m) {
          var n = m.lights.length;
          for (var i = 0; i < n; i++) {
            var on = still ? i % 2 === 0 : fn(t, i, n);
            m.lights[i].classList.toggle("on", on);
            m.lights[i].classList.toggle("on-warm", on && !still && marqueePattern === 1);
            m.lights[i].classList.toggle("on-cool", on && !still && marqueePattern === 3);
          }
        });
        marqueeFrame = requestAnimationFrame(marqueeTick);
      }
      var qrCodes = [];
      function addQrCodes(bubble) {
        if (!qrCodes.length || !bubble.classList || !bubble.classList.contains("brobot-bubble") || bubble.querySelector("img.gopod-qr")) { return; }
        if (bubble.classList.contains("operator")) { return; }
        var textEl = bubble.querySelector(".brobot-bubble-text");
        var says = ownSays(bubble);
        if (!textEl || !says.length) { return; }
        var said = says.map(function (x) { return String(x.textContent || ""); }).join(" ").toLowerCase();
        var found = qrCodes.filter(function (code) {
          var text = said;
          (code.not_words || []).forEach(function (w) { text = text.split(String(w).toLowerCase()).join(" "); });
          return (code.words || []).some(function (w) { return text.indexOf(String(w).toLowerCase()) >= 0; });
        });
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
      function chalkQrs(bubble, found) {
        var codes = found.slice();
        if (chalkCta(bubble) || bubble.querySelector(".chalk-cta")) { codes = codes.filter(function (code) { return code.image !== secondQr.image; }); }
        var cta = bubble.querySelector(".chalk-cta img.gopod-qr");
        if (cta) { codes = codes.filter(function (code) { return code.image !== cta.getAttribute("src"); }); }
        if (!codes.length) { return; }
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
      function chalkCta(bubble) {
        if (!secondQr || !secondQr.after || bubble.querySelector(".chalk-cta")) { return; }
        var after = String(secondQr.after).toLowerCase(), hit = null;
        var says = ownSays(bubble);
        says.forEach(function (say) {
          if (!hit && String(say.textContent || "").toLowerCase().indexOf(after) >= 0) { hit = say; }
        });
        if (!hit) { says.forEach(function (say) { if (!hit && /newsletter/i.test(say.textContent || "")) { hit = say; } }); }
        if (!hit) { return; }
        var ctaEnd = hit;
        var prev = hit.previousElementSibling;
        if (prev && prev.classList.contains("brobot-say") && /know more/i.test(prev.textContent || "") && !/know more/i.test(hit.textContent || "")) { hit = prev; }
        var box = nestPart("cta", "chalk-qr chalk-cta", null);
        box.insertBefore(displayHead(box), box.firstChild);
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
        if (hit === says[0] && ctaEnd !== says[says.length - 1]) { ctaEnd.parentNode.insertBefore(box, ctaEnd.nextSibling); }
        else { hit.parentNode.insertBefore(box, hit); }
        return true;
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
      var weatherCard = null;
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
      function weatherShowFor(brobot) {
        var cls = brobot === "Brobot 2" ? "brobot-2" : "brobot-1";
        var boxes = container.querySelectorAll(".brobot-bubble." + cls + " .gopod-weather-after:not(.shown)");
        var box = boxes.length ? boxes[boxes.length - 1] : null;
        if (box && box._show) { box._show(); }
      }
      function weatherSoundLoad(sound) {
        if (!sound) { return null; }
        var audio = new Audio(normSrc(weatherCard.folder + "/sounds/" + sound + ".mp3"));
        audio.preload = "auto";
        audio.addEventListener("ended", function () { audio._playedOnce = true; });
        try { audio.load(); } catch (err) { /* no audio */ }
        return audio;
      }
      function weatherSoundOnShow(bubble, sound) {
        if (restoringChat || !sound) { return null; }
        var audio = new Audio(normSrc(weatherCard.folder + "/sounds/" + sound + ".mp3"));
        audio.preload = "auto";
        audio.addEventListener("ended", function () { audio._playedOnce = true; });
        try { audio.load(); } catch (err) { /* no audio */ }
        whenShown(bubble, function () {
          audio.volume = chatLevel(volumeShare(weatherCard.volume));
          audio.play().catch(function () { audio._playedOnce = true; /* the page hasn't been clicked yet */ });
        }, 60000);
        return audio;
      }
      function weatherSoundOnLit(bubble, textEl, sound) {
        var at = String(weatherCard.sound_at || "").toLowerCase().trim();
        var piece = !at ? null : Array.prototype.filter.call(textEl.querySelectorAll(".brobot-say"), function (x) {
          return !x.closest(".nested-host") && String(x.textContent || "").toLowerCase().indexOf(at) >= 0;
        })[0];
        if (!piece) { return weatherSoundOnShow(bubble, sound); }
        if (restoringChat || !sound || piece.classList.contains("spoken")) { return null; }
        var audio = new Audio(normSrc(weatherCard.folder + "/sounds/" + sound + ".mp3"));
        audio.preload = "auto";
        try { audio.load(); } catch (err) { /* no audio */ }
        function play() {
          audio.volume = chatLevel(volumeShare(weatherCard.volume));
          audio.play().catch(function () { /* the page hasn't been clicked yet */ });
        }
        if (piece.classList.contains("speaking")) { play(); return audio; }
        var watch = new MutationObserver(function () {
          if (piece.classList.contains("speaking")) { watch.disconnect(); play(); }
          else if (piece.classList.contains("spoken") || !piece.isConnected) { watch.disconnect(); }
        });
        watch.observe(piece, { attributes: true, attributeFilter: ["class"] });
        return audio;
      }
      function addWeatherFace(bubble) {
        if (!weatherCard || !bubble.classList || !bubble.classList.contains("brobot-bubble") || bubble.querySelector(".gopod-weather-face")) { return; }
        var textEl = bubble.querySelector(":scope > .brobot-bubble-text");
        if (!textEl) { return; }
        var said = String(textEl.textContent || "").toLowerCase();
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
        if (weatherCard.picture === false) {
          if (!bubble._weatherSound) { bubble._weatherSound = true; weatherSoundOnLit(bubble, textEl, face.sound); }
          return;
        }
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
        var sound = weatherSoundOnShow(bubble, face.sound);
        function showNow() { box.classList.add("shown"); }
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
        if (sound) {
          screen.addEventListener("click", function () {
            if (!sound._playedOnce || !sound.paused) { return; }
            sound.currentTime = 0;
            sound.volume = chatLevel(volumeShare(weatherCard.volume));
            sound.play().catch(function () {});
          });
        }
      }
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
      var reporterRole, hostNames = {}, hostLabels = {}, chatMoves = {};
      function gotItOf(who) {
        var g = (chatMoves.got_it || {})[who] || {};
        return String(g.text || chatMoves.thinking_text || "").trim();
      }
      function openReceipt(who) {
        var at = openHostBox();
        return at && at.who === who && at.outer.dataset.receipt ? at.outer : null;
      }
      var displayNames = {};
      function numpadPersona(who) {
        var on = numpad.map && numpad.map.numlock_on, name = (numpad.personaReporters || {})[who] || who;
        for (var k in (on || {})) {
          if (on[k] && String(on[k].persona || "").toUpperCase() === String(name).toUpperCase()) { return { name: on[k].persona, key: k }; }
        }
        return null;
      }
      function reporterSuit(who) {
        var anchor = (numpad.personaReporters || {})[who];
        if (!anchor || String(anchor).toUpperCase() === String(who).toUpperCase()) { return ""; }
        return " reporter-suit suit-" + String(anchor).toLowerCase().replace(/[^a-z0-9]+/g, "-");
      }
      var nickLabels = {}, personaVoiceOf = {}, systemVoicePersona = "";
      function hostHeader(who) {
        if (who === systemVoicePersona || (personaVoiceOf[who] && personaVoiceOf[who] === systemRole)) { return systemLabel; }
        if (nickLabels[who]) { return nickLabels[who]; }
        var np = numpadPersona(who);
        var dn = displayNames || {}, edits = dn.edits || {}, def = (dn["default"] || {})[who] || (np ? np.name : "");
        if (def && edits[def]) { return edits[def]; }
        if (edits[who]) { return edits[who]; }
        if (hostLabels[who]) { return hostLabels[who]; }
        if (def) { return def; }
        if (np && dn.fallback) { return String(dn.fallback).replace("{n}", np.key.replace(/^KP/, "")); }
        return hostLabels[who] || who;
      }
      var songAudioRole, narratorRole, songAudio = null;
      var nestShells = {};
      function nestShell(brobot) {
        var shell = nestShells[brobot];
        if (shell && shell.isConnected && !shell.dataset.filled) { return shell; }
        shell = rawBubble({ kind: "shell", classes: (brobot === "Brobot 2" ? "brobot-2" : "brobot-1") + " nest-shell", head: bubbleHead(brobot) });
        shell.dataset.nest = brobot;
        brobotHostClass(shell, brobot);
        nestShells[brobot] = shell;
        placeBubble(shell);
        return shell;
      }
      function nestPart(kind, classes, head, line, lineClass) {
        return rawBubble({ kind: kind, classes: "operator system-status nested-host " + classes, head: head, lines: line == null ? [] : [line], lineClass: lineClass });
      }
      function nestPlace(shell, part, atTop) {
        var reply = shell.querySelector(":scope > .brobot-bubble-text");
        if (atTop && reply) { shell.insertBefore(part, reply); } else { shell.appendChild(part); }
        return part;
      }
      function nestBox(shell, who, atTop) {
        var box = nestPart("host", "host host-" + String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-") + reporterSuit(who), reporterHead(hostHeader(who)));
        box.dataset.host = who;
        var notice = atTop ? shell.querySelector(":scope > .bubble-notice") : null;
        if (notice) {
          box.classList.add("in-notice");
          notice.appendChild(box);
          brainInView(notice);
          return box;
        }
        return nestPlace(shell, box, atTop);
      }
      function brainInView(el) {
        function keep() {
          if (!el || !container.contains(el)) { return; }
          var bottom = contentY(el) + el.getBoundingClientRect().height;
          var over = bottom - (container.scrollTop + chatViewHeight()) + 8;
          if (over > 0) {
            scrollLog("brain bubble's bottom kept in view: +" + Math.round(over));
            scrollChatTo(Math.min(container.scrollTop + over, chatContentEnd()));
          }
        }
        setTimeout(keep, 80);
        setTimeout(keep, 600);
      }
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
      var managerHosts = {};
      function hostSlug(who) { return String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-"); }
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
        var hostClass = "operator system-status host host-" + String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-") + reporterSuit(who);
        if (String(text).indexOf("qr:") !== 0 && String(text).indexOf("lines:") !== 0) {
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
          generatingGone();
          var shell = nestShell(to);
          var boxes = shell.querySelectorAll(".nested-host.host-" + String(who).toLowerCase().replace(/[^a-z0-9]+/g, "-"));
          var box = boxes.length && String(text).indexOf("qr:") === 0 ? boxes[boxes.length - 1] : nestBox(shell, who, true);
          if (String(text).indexOf("qr:") !== 0) { bubbleSound("Manager"); }
          if (String(text).indexOf("qr:") === 0) { hostQr(box, text); next(); return; }
          var nline = document.createElement("span");
          nline.className = "brobot-bubble-text";
          box.appendChild(nline);
          voiceLitLine(shell, nline, text);
          addWeatherFace(box);
          next();
          return;
        }
        bubbleSound("Manager");
        var bubble = rawBubble({ kind: "host", classes: hostClass, head: reporterHead(hostHeader(who)) });
        bubble.dataset.host = who;
        if (String(text).indexOf("qr:") === 0) {
          hostQr(bubble, text);
          placeBubble(bubble, next);
          return;
        }
        if (!restoringChat) { recordBubble(who, String(text), "host"); }
        if (String(text).indexOf("lines:") === 0) {
          var receipt = openReceipt(who);
          if (receipt) {
            delete receipt.dataset.receipt;
            hostBrain(receipt, who);
            linesInto(receipt, text);
            next();
            return;
          }
          hostBrain(bubble, who);
          linesInto(bubble, text);
          placeBubble(bubble, next, function () {
            if (warmupBubble && warmupBubble.isConnected) { container.appendChild(warmupBubble); }
            else if (warmupBeforeOpening && warmupBeforeOpening.isConnected) { container.appendChild(warmupBeforeOpening); }
          });
          return;
        }
        var line = document.createElement("span");
        line.className = "brobot-bubble-text";
        if (gotItOf(who) && String(text).trim() === gotItOf(who)) {
          var opener = document.createElement("div");
          opener.className = "brobot-opener";
          opener.appendChild(line);
          bubble.appendChild(opener);
          bubble.dataset.receipt = "1";
        } else {
          bubble.appendChild(line);
        }
        hostBrain(bubble, who);
        voiceLitLine(bubble, line, text);
        placeBubble(bubble, next, function () {
          if (warmupBubble && warmupBubble.isConnected) { container.appendChild(warmupBubble); }
          else if (warmupBeforeOpening && warmupBeforeOpening.isConnected) { container.appendChild(warmupBeforeOpening); }
        });
      }

      var NARRATOR_PIECE_RE = /^@(\d+(?:\.\d+)?)-(\d+(?:\.\d+)?)\s+(?:~([A-Za-z0-9+\/=]+)\s+)?/;
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
        var bubble = rawBubble({ kind: "narrator", classes: "operator system-status host narrator", head: systemHead((hostLabels[narratorRole] || narratorRole) + ":") });
        var line = document.createElement("div");
        line.className = "brobot-bubble-text brobot-bubble-line";
        var pieces = segmentMarker ? String(text).split(segmentMarker) : [String(text)];
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
        stageBubble(bubble);
        awaitingReply = false;
        operatorInBatch = false;
        narratorLines.push({ bubble: bubble, firstAt: parseFloat(first.dataset.at), shownMs: Date.now(), lit: null, shown: false, placed: false });
        if (narratorTimer === null) { narratorTimer = setInterval(lightNarrator, 100); }
        lightNarrator();
      }
      var narratorStage = null;
      var NARRATOR_FADE_IN_S, NARRATOR_FADE_OUT_MAX_S, NARRATOR_FADE_OUT_END_S;
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
        var reuse = narratorStage && narratorStage.isConnected && firstAt > narratorStage._lastAt;
        var bubble, stage;
        if (reuse) {
          bubble = narratorStage;
          stage = bubble.querySelector(".narrator-stage");
        } else {
          bubble = rawBubble({ kind: "narrator", classes: "operator system-status host narrator narrator-single" });
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
          if (pick === n.lit) { if (pick) { pick.classList.add("speaking"); } return; }
          if (n.lit) { n.lit.classList.remove("speaking"); n.lit.classList.add("spoken"); }
          n.lit = pick;
          if (pick) {
            lightSpan(pick);
          } else if (t >= parseFloat(spans[spans.length - 1].dataset.end)) {
            revealRest(n.part || n.bubble);
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

      var feedPolling = false;
      function pollFeed() {
        if (feedPolling) { return; }
        feedPolling = true;
        fetch("/gopod_chat_feed.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (feed) {
            if (!feed || !Array.isArray(feed.events)) { return; }
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
              try { window.localStorage.setItem(NUMPAD_KEY, "1"); } catch (e) { /* in-page only */ }
              if (!hadBoot) { setNumpadOn(true); }
              setChatOpen(true);
              showChat(true);
              backgroundLoop.setSpeaker(false);
              showSpeaker();
              if (hadBoot) {
                window.location.reload();
                return;
              }
            }
            if (feed.boot && feed.boot !== playedBoot) { playedBoot = feed.boot; playedId = 0; playedSent = 0; }
            if (pageRoleKnown && !isMainPage) { return; }
            var seenSkipOn = false, seenOpeningEnd = false;
            feed.events.forEach(function (e) {
              var key = "feed|" + feed.boot + "|" + e.id;
              if (isSeen(key) && e.who === reporterRole) {
                if (String(e.text).trim() === "skip_on") { seenSkipOn = true; }
                if (String(e.text).trim() === openingEndText) { seenOpeningEnd = true; }
              }
              if (isSeen(key) && e.id > playedId && !feedQueue.length && !feedRunning) { playedId = e.id; }
              if (isSeen(key) || feedQueued[key] || !e.who || !e.text) { return; }
              if (e.who === reporterRole && String(e.text).indexOf(turnDropMark) === 0) { markSeen(key); dropTurn(Number(String(e.text).slice(turnDropMark.length))); return; }
              if (e.turn && droppedTurns[e.turn]) { markSeen(key); if (e.id > playedId && !feedQueue.length && !feedRunning) { playedId = e.id; } return; }
              queueBubble(key, e);
            });
            if (!skipGate.decided && seenSkipOn) {
              skipGate.decided = true;
              if (!seenOpeningEnd) { skipGate.on = true; }
            }
            runFeedQueue();
          })
          .catch(function () { /* no feed yet (older Wire-Pod, or nothing said) */ })
          .then(function () { feedPolling = false; });
      }

      var feedQueue = [], feedQueued = {}, feedRunning = false;
      var turnDropMark = "drop:", droppedTurns = {}, feedItemNow = null;
      function dropTurn(turn) {
        if (!turn) { return; }
        droppedTurns[turn] = true;
        var kept = [];
        feedQueue.forEach(function (item) {
          if (item.e && item.e.turn === turn) {
            delete feedQueued[item.key];
            markSeen(item.key);
            if (item.e.id > playedId) { playedId = item.e.id; }
          } else {
            kept.push(item);
          }
        });
        feedQueue = kept;
        if (feedItemNow && feedItemNow.e && feedItemNow.e.turn === turn) {
          if (voicePlaying) { try { voicePlaying.audio.pause(); } catch (err) { /* gone */ } voicePlaying.finish(); }
          skipFinishWaiting();
        }
        var note = new URLSearchParams();
        note.set("note", "turn " + turn + " dropped");
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: note }).catch(function () { /* Wire-Pod gone */ });
        runFeedQueue();
      }
      var playedId = 0, playedSent = 0, playedBoot = 0, heldMost = { id: 0, ms: 0 };
      setInterval(function () {
        if (!playedBoot || playedId <= playedSent) { return; }
        var body = new URLSearchParams();
        body.set("boot", String(playedBoot));
        body.set("id", String(playedId));
        body.set("page", pageId);
        if (heldMost.id) { body.set("held_id", String(heldMost.id)); body.set("held_ms", String(heldMost.ms)); heldMost = { id: 0, ms: 0 }; }
        playedSent = playedId;
        fetch("/api-sdk/gopod_page_played", { method: "POST", body: body })
          .then(function (r) { return r.text(); })
          .then(function (t) { if (t !== "ok") { playedSent = 0; } })
          .catch(function () { playedSent = 0; });
      }, 250);
      var pageClicked = false, openingEndText = "opening_end", voiceGapMs = 300;
      var heardMarker = "";
      var generatingMarker = "", generatingText = "Generating...", generatingBubble = null;
      function generatingGone() {
        if (generatingBubble) { generatingBubble.remove(); generatingBubble = null; }
      }
      function generatingGlass(bubble) {
        var row = document.createElement("div");
        row.className = "warmup-timer generating-timer";
        var glass = document.createElement("span");
        glass.className = "countdown-hourglass warmup-hourglass";
        glass.innerHTML = hourglassSprite(0);
        row.appendChild(glass);
        rockHourglass(bubble, glass, 0);
        var steps = hourglassCfg.steps || 16, stepMs = hourglassCfg.step_ms || 150, holdMs = hourglassCfg.hold_ms || 400;
        var placedOnce = false, madeAt = Date.now();
        (function drain(stage) {
          if (bubble.classList.contains("ready")) { return; }
          if (!row.isConnected) {
            if (placedOnce || Date.now() - madeAt > 10000) { return; }
            setTimeout(function () { drain(stage); }, stepMs);
            return;
          }
          placedOnce = true;
          glass.innerHTML = hourglassSprite(stage, steps);
          var last = stage >= steps - 1;
          setTimeout(function () { drain(last ? 0 : stage + 1); }, last ? holdMs : stepMs);
        })(0);
        return row;
      }
      function slotWaiting(bubble) {
        var text = bubble.querySelector(".brobot-bubble-text");
        if (!text || bubble.querySelector(".generating-timer")) { return false; }
        bubble.dataset.words = text.textContent;
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
      var voiceForText = {};
      function voiceLineWords(t) { return String(t).replace(/\s+/g, " ").trim(); }
      function voiceLineSpan(el, t) {
        var ready = el.querySelectorAll(".brobot-say:not(.spoken):not(.speaking)");
        for (var r = 0; r < ready.length; r++) {
          if (voiceLineKey(ready[r].textContent) === voiceLineKey(t)) { return ready[r]; }
        }
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
      var voiceLinesAhead = {};
      function voiceLineKey(t) { return String(t || "").toLowerCase().replace(/[^a-z0-9]+/g, ""); }
      function voiceLineAhead(words, el) { if (voiceLineKey(words)) { voiceLinesAhead[voiceLineKey(words)] = el; } }
      function voiceLineTieIn(el, bubble, file, t, waits, trust) {
        var mine = Array.prototype.filter.call(el.querySelectorAll(".brobot-say"), function (x) {
          return voiceLineWords(x.textContent) === t && !x.classList.contains("spoken");
        })[0];
        if (mine) { return mine; }
        var say = voiceLineSpan(el, t);
        if (!say && trust && !el.querySelector(".brobot-say") && voiceLineKey(el.textContent) === voiceLineKey(t)) {
          say = document.createElement("span");
          say.className = "brobot-say";
          while (el.firstChild) { say.appendChild(el.firstChild); }
          el.appendChild(say);
        }
        if (!say) { return null; }
        if (skippingReporters) { say.classList.add("spoken"); return say; }
        bubble.classList.add("voice-lit");
        if (waits) { voiceWaitOn(bubble, say, file); }
        showOnItsTurn(bubble, say);
        return say;
      }
      function voiceLineTie(file, t, waits) {
        var key = voiceLineKey(t), ahead = voiceLinesAhead[key];
        if (ahead) {
          delete voiceLinesAhead[key];
          if (ahead.isConnected && !ahead.classList.contains("spoken")) {
            if (ahead.classList.contains("brobot-say")) {
              var host = ahead.closest(".brobot-bubble");
              if (waits && host && !ahead._wait && !ahead.classList.contains("speaking")) { voiceWaitOn(host, ahead, file); }
              return ahead;
            }
            var named = ahead.closest(".brobot-bubble");
            var tied = named ? voiceLineTieIn(ahead, named, file, t, waits, true) : null;
            if (tied) { return tied; }
          }
        }
        var els = container.querySelectorAll(".brobot-bubble-text, .brobot-bubble-line");
        for (var i = els.length - 1; i >= 0 && i >= els.length - 40; i--) {
          var el = els[i], bubble = el.closest(".brobot-bubble");
          if (!bubble || voiceLineWords(el.textContent).indexOf(t) < 0) { continue; }
          if (el.closest(".brobot-say") && !el.closest(".brobot-say").classList.contains("spoken")) { return el.closest(".brobot-say"); }
          var say = voiceLineTieIn(el, bubble, file, t, waits, false);
          if (say) { return say; }
        }
        return null;
      }
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
        if (runCheckMuted && pageClicked && (!log || log.style.display !== "none")) { return true; }
        return pageClicked && document.visibilityState === "visible" &&
          (!log || log.style.display !== "none") && container.offsetParent !== null &&
          container.classList.contains("expanded");
      }
      function queueBubble(key, e) {
        feedQueued[key] = true;
        var item = { key: key, e: e, at: Date.now() };
        var at = feedQueue.length;
        while (at > 0 && feedQueue[at - 1].e.id > e.id) { at--; }
        feedQueue.splice(at, 0, item);
        if (isMainPage && item.e && isVoiceEvent(item.e) && !restoringChat && !stingFiles[String(item.e.text).replace(/#vol=[0-9.]+$/, "").split("/").pop()]) { voiceAudio(item.e.text); }
        runFeedQueue();
      }
      function pageError(what) {
        var body = new URLSearchParams();
        body.set("note", "error " + String(what).slice(0, 300));
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: body }).catch(function () { /* Wire-Pod gone */ });
      }
      window.addEventListener("error", function (ev) { pageError((ev.message || "error") + " at " + (ev.lineno || "?")); });
      window.addEventListener("unhandledrejection", function (ev) { pageError("promise: " + (ev.reason && ev.reason.message || ev.reason)); });
      var feedItemMaxMs = 90000;
      function runFeedQueue() {
        if (feedRunning || !feedQueue.length || !chatViewable() || !pageRoleKnown || !isMainPage) { return; }
        feedRunning = true;
        var item = feedQueue.shift();
        delete feedQueued[item.key];
        feedItemNow = item;
        if (item.e && item.e.id && Date.now() - item.at > heldMost.ms) { heldMost = { id: item.e.id, ms: Date.now() - item.at }; }
        markSeen(item.key);
        var went = false, guard = null;
        var next = function () {
          if (went) { return; }
          went = true;
          clearTimeout(guard);
          if (item.e && item.e.id > playedId) { playedId = item.e.id; }
          feedRunning = false;
          runFeedQueue();
        };
        guard = setTimeout(function () {
          pageError("stuck " + Math.round(feedItemMaxMs / 1000) + " s on feed " + (item.e && item.e.id) + " " + (item.e && item.e.who) + " (" + (batonHolder() || "no lit line") + ") - passed");
          next();
        }, feedItemMaxMs);
        nextBubbleVoice = !isVoiceEvent(item.e) && feedQueue.length && isVoiceEvent(feedQueue[0].e) ? voiceAudio(feedQueue[0].e.text) : null;
        try { showFeedEvent(item.e, next); } catch (err) {
          pageError("feed " + (item.e && item.e.id) + ": " + (err && err.message));
          next();
        }
        nextBubbleVoice = null;
      }
      var lastShown = null;
      var pageId = "", pageViewOnly = /[?&]view(=|&|$)/.test(location.search);
      var isMainPage = false, pageRoleKnown = false;
      try { pageId = sessionStorage.getItem("gopod_page_id") || ""; } catch (err) { /* no storage */ }
      if (!pageId) {
        pageId = Date.now().toString(36) + Math.random().toString(36).slice(2, 8);
        try { sessionStorage.setItem("gopod_page_id", pageId); } catch (err) { /* no storage */ }
      }
      var visitorId = "";
      try { visitorId = localStorage.getItem("gopod_visitor_id") || ""; } catch (err) { /* no storage */ }
      if (!visitorId) {
        visitorId = "v" + Date.now().toString(36) + Math.random().toString(36).slice(2, 10);
        try { localStorage.setItem("gopod_visitor_id", visitorId); } catch (err) { /* no storage */ }
      }
      var visitorLastState = "", visitorClearTimer = null, secondAdminShown = "";
      function visitorStanding(v) {
        var state = v.state || "";
        if (state === visitorLastState) { return; }
        var was = visitorLastState;
        visitorLastState = state;
        clearTimeout(visitorClearTimer);
        if (state === "waiting") {
          typedText.placeholder = v.waiting_text || typedPlaceholder;
          typedText.classList.add("gopod-typed-note");
          return;
        }
        var approved = state === "approved" || state === "on" || state === "muted";
        var told = "";
        try { told = localStorage.getItem("gopod_visitor_approved") || ""; } catch (err) { told = ""; }
        if (approved && v.name && told !== v.name && (was === "waiting" || was === "")) {
          try { localStorage.setItem("gopod_visitor_approved", v.name); } catch (err) { /* no storage */ }
          typedText.placeholder = v.approved_text || typedPlaceholder;
          typedText.classList.add("gopod-typed-note");
          visitorClearTimer = setTimeout(function () {
            typedText.placeholder = typedPlaceholder;
            typedText.classList.remove("gopod-typed-note");
          }, v.approved_show_ms || 4000);
          return;
        }
        if (state === "nameless") {
          try { localStorage.removeItem("gopod_visitor_approved"); } catch (err) { /* no storage */ }
        }
        typedText.placeholder = typedPlaceholder;
        typedText.classList.remove("gopod-typed-note");
      }
      function setPageRole(main) {
        var changed = !pageRoleKnown || main !== isMainPage;
        isMainPage = main;
        pageRoleKnown = true;
        if (!changed) { return; }
        if (!main) { feedQueue = []; feedQueued = {}; } // a viewer plays none of the feed
        backgroundLoop.setVolume(backgroundLoop.volume());
        var body = new URLSearchParams();
        body.set("note", "page " + pageId + (main ? " is the main page (plays the sound)" : " is a viewer (silent)"));
        fetch("/api-sdk/gopod_page_note", { method: "POST", body: body }).catch(function () { /* Wire-Pod gone */ });
        displayMirror();
        runFeedQueue();
      }
      var DISPLAY_BUBBLES, DISPLAY_MS;
      var displaySendTimer = null, displaySent = "", displayObserver = null, displayPollTimer = null, displaySeen = 0;
      var displayKeys = new WeakMap(), displayNextKey = 0;
      function displayBubbles() {
        return Array.prototype.filter.call(container.children, function (el) {
          return el !== typedRow && el.dataset.staged !== "1" && el.style.visibility !== "hidden" && !el.classList.contains("gopod-admin-only");
        }).slice(-DISPLAY_BUBBLES);
      }
      function displaySend() {
        displaySendTimer = null;
        if (!isMainPage) { return; }
        var items = displayBubbles().map(function (b) {
          if (!displayKeys.has(b)) { displayKeys.set(b, ++displayNextKey); }
          var h = b.outerHTML;
          if (b.querySelector(".marquee-lights")) {
            var copy = b.cloneNode(true);
            Array.prototype.forEach.call(copy.querySelectorAll(".marquee-lights"), function (l) { l.textContent = ""; });
            h = copy.outerHTML;
          }
          return { k: displayKeys.get(b), h: h };
        });
        var body = JSON.stringify({ items: items });
        if (body === displaySent) { return; }
        displaySent = body;
        displaySending = true;
        fetch("/api-sdk/gopod_display", { method: "POST", headers: { "Content-Type": "application/json" }, body: body })
          .catch(function () { displaySent = ""; })
          .then(function () {
            displaySending = false;
            if (displayPending) { displayPending = false; displayChanged(); }
          });
      }
      var displaySending = false, displayPending = false;
      function displayChanged() {
        if (!isMainPage) { return; }
        if (displaySending) { displayPending = true; return; }
        if (displaySendTimer === null) { displaySendTimer = setTimeout(displaySend, DISPLAY_MS); }
      }
      function displayMorph(a, b) {
        if (a.nodeType !== b.nodeType || a.nodeName !== b.nodeName) { a.parentNode.replaceChild(b, a); return; }
        if (a.nodeType === 3 || a.nodeType === 8) { if (a.nodeValue !== b.nodeValue) { a.nodeValue = b.nodeValue; } return; }
        if (a.nodeType !== 1) { return; }
        if (a.classList.contains("marquee-lights")) { return; } // this page's own bulbs
        Array.prototype.slice.call(a.attributes).forEach(function (at) { if (!b.hasAttribute(at.name)) { a.removeAttribute(at.name); } });
        Array.prototype.forEach.call(b.attributes, function (at) { if (a.getAttribute(at.name) !== at.value) { a.setAttribute(at.name, at.value); } });
        var bk = Array.prototype.slice.call(b.childNodes);
        bk.forEach(function (child, i) {
          if (i < a.childNodes.length) { displayMorph(a.childNodes[i], child); } else { a.appendChild(child); }
        });
        while (a.childNodes.length > bk.length) { a.removeChild(a.lastChild); }
      }
      function displayPaint(d) {
        var firstPaint = !Array.prototype.some.call(container.children, function (el) { return el !== typedRow && el.dataset.mid; });
        var have = {};
        Array.prototype.forEach.call(container.children, function (el) { if (el !== typedRow && el.dataset.mid) { have[el.dataset.mid] = el; } });
        var keep = {}, order = [], firstNew = null;
        (d.items || []).forEach(function (it) {
          var t = document.createElement("template");
          t.innerHTML = it.h;
          var fresh = t.content.firstElementChild;
          if (!fresh) { return; }
          fresh.setAttribute("data-mid", String(it.k));
          var el = have[it.k];
          if (el) { displayMorph(el, fresh); } else { el = fresh; if (!firstNew) { firstNew = el; } }
          keep[it.k] = true;
          order.push(el);
        });
        Array.prototype.slice.call(container.children).forEach(function (el) { if (el !== typedRow && !keep[el.dataset.mid]) { el.remove(); } });
        var placed = order.filter(function (el) { return el.parentNode === container; });
        var now = Array.prototype.filter.call(container.children, function (el) { return el !== typedRow; });
        var lastPlaced = placed.length ? order.indexOf(placed[placed.length - 1]) : -1;
        var inOrder = now.length === placed.length && now.every(function (el, i) { return el === placed[i]; }) &&
          order.every(function (el, i) { return el.parentNode === container || i > lastPlaced; });
        order.forEach(function (el) { if (!inOrder || el.parentNode !== container) { container.appendChild(el); } });
        Array.prototype.forEach.call(container.querySelectorAll(".marquee-lights"), function (layer) {
          if (!marquees.some(function (m) { return m.layer === layer; })) { marqueeTrack(layer.parentElement, layer); }
        });
        var speaking = container.querySelector(".brobot-say.speaking");
        if (firstPaint && firstNew) {
          var toEnd = function () { container.scrollTop = chatMaxScroll(); window.scrollTo(0, document.body.scrollHeight); };
          toEnd();
          setTimeout(toEnd, 600); setTimeout(toEnd, 1500); // pictures still loading
        }
        else if (firstNew) { scrollChatTo(bubbleTarget(firstNew)); } else if (speaking) { focusSpeaking(speaking); }
      }
      function displayPoll() {
        if (isMainPage) { return; }
        fetch("/api-sdk/gopod_display?since=" + displaySeen, { cache: "no-store" })
          .then(function (r) { return r.status === 200 ? r.json() : null; })
          .then(function (d) { if (d && !isMainPage && d.v > displaySeen) { displaySeen = d.v; displayPaint(d); } })
          .catch(function () { /* Wire-Pod gone */ });
      }
      function displayMirror() {
        clearInterval(displayPollTimer);
        displayPollTimer = null;
        if (displayObserver) { displayObserver.disconnect(); displayObserver = null; }
        if (isMainPage) {
          displayObserver = new MutationObserver(function (records) {
            var real = records.some(function (r) {
              var el = r.target.nodeType === 1 ? r.target : r.target.parentElement;
              return !(el && el.closest && el.closest(".marquee-lights"));
            });
            if (real) { displayChanged(); }
          });
          displayObserver.observe(container, { childList: true, subtree: true, attributes: true, characterData: true });
          displayChanged();
        } else {
          displayPollTimer = setInterval(displayPoll, DISPLAY_MS);
          displayPoll();
        }
      }
      function commandmentAt(x, y) {
        var node = null, at = 0;
        if (document.caretRangeFromPoint) {
          var r = document.caretRangeFromPoint(x, y);
          if (r) { node = r.startContainer; at = r.startOffset; }
        } else if (document.caretPositionFromPoint) {
          var p = document.caretPositionFromPoint(x, y);
          if (p) { node = p.offsetNode; at = p.offset; }
        }
        if (!node || node.nodeType !== 3) { return ""; }
        var t = node.textContent, a = at, b = at;
        while (a > 0 && /[A-Za-z0-9-]/.test(t.charAt(a - 1))) { a--; }
        while (b < t.length && /[A-Za-z0-9-]/.test(t.charAt(b))) { b++; }
        var word = t.slice(a, b).replace(/-+$/, "");
        return /^system-[A-Za-z0-9-]+$/i.test(word) ? word.toLowerCase() : "";
      }
      document.addEventListener("dblclick", function (ev) {
        if (!isAdminPage || ev.target === typedText) { return; }
        var cmd = commandmentAt(ev.clientX, ev.clientY);
        if (!cmd) { return; }
        try { window.getSelection().removeAllRanges(); } catch (err) { /* no selection */ }
        typedText.value = cmd;
        typedText.focus();
      }, true);
      function reportShown(force) {
        var log = document.getElementById("section-log");
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
            var zero = allFirstSelect && allFirstSelect.querySelector(".gopod-brain");
            if (zero) {
              zero.hidden = !isAdminPage;
              zero.disabled = !isAdminPage;
              if (!isAdminPage && allFirstSelect.value === "-0") { allFirstSelect.value = "+*"; }
            }
            if (document.body.classList.contains("gopod-visitor") === isAdminPage) {
              document.body.classList.toggle("gopod-visitor", !isAdminPage);
              fitTypedRow();
            }
            if (isAdminPage) { document.body.classList.remove("gopod-chat-only"); }
            if (!isAdminPage && !document.body.classList.contains("gopod-chat-only")) { visitorOpenChat(); }
            if (appStart) { appStart = false; if (isAdminPage && typeof window.showLog === "function") { window.showLog(); } }
            if (chatOnlyPending) { chatOnlyPending = false; setTimeout(chatOnlyOn, 0); }
            setPageRole(d.main);
            if (d.visitor) { visitorStanding(d.visitor); }
            if (isAdminPage) {
              var second = d.second_admin || "";
              if (second !== secondAdminShown) {
                secondAdminShown = second;
                typedText.placeholder = second || typedPlaceholder;
                typedText.classList.toggle("gopod-typed-note", !!second);
              }
            }
          })
          .catch(function () { /* Wire-Pod not up yet */ });
      }
      document.addEventListener("visibilitychange", function () { reportShown(false); runFeedQueue(); });
      setInterval(function () { reportShown(false); }, 1000);
      setInterval(function () { reportShown(true); }, 5000);
      var nextBubbleVoice = null, voiceCache = {};
      var chatVolume = 0.35;
      var runCheckMuted = false; // run_check.mute_marker / unmute_marker on the feed
      var masterMuted = false; // the master mute button (after the loop's speaker)
      function chatLevel(share) {
        if (!isMainPage) { return 0; }
        var v = typeof share === "number" && !isNaN(share) ? share : 1;
        if (runCheckMuted || masterMuted) { return 0; } // 2026-11-08: the silent run-check; 2026-11-09: the master mute
        return Math.max(0, Math.min(1, chatVolume * v));
      }
      function volumeShare(v) {
        if (typeof v === "number" && !isNaN(v)) { return Math.max(0, v); }
        var t = String(v == null ? "" : v).trim();
        if (!t || t.toLowerCase() === "default") { return 1; }
        var n = parseFloat(t);
        if (isNaN(n) || n < 0) { return 1; }
        return /%$/.test(t) ? n / 100 : n;
      }
      var audioLevels = {};
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
              Object.keys(stingPool).forEach(function (k) { if (audioMade[k]) { delete stingPool[k]; preloadStings([k]); } });
              Object.keys(voiceCache).forEach(function (k) { if (audioMade[k.replace(/#.*$/, "")]) { delete voiceCache[k]; } });
            }
          })
          .catch(function () { /* none yet */ });
      }
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
      var VOICE_READY_MAX_MS;
      function voiceAudio(src) {
        var key = String(src);
        if (!voiceCache[key]) {
          var vol = /#vol=([0-9.]+)$/.exec(key);
          var a = new Audio(normSrc(key.replace(/#vol=[0-9.]+$/, "")));
          a._share = vol ? parseFloat(vol[1]) : 1;
          a.preload = "auto";
          a.addEventListener("error", function () { a._failed = true; });
          try { a.load(); } catch (err) { a._failed = true; }
          var plain = key.replace(/#.*$/, "");
          a._wordsDone = true;
          if (/\/gopod_reporter_audio\/[a-z]{2}_[a-z]+_(s\d+_)?[0-9a-f]{16}\.wav$/.test(plain)) {
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
        a._wordsDone = true;
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
      var voicesSkipped = 0;
      function playVoiceThen(src, done) {
        var finished = false, litDone = null;
        var finish = function () {
          if (finished) { return; }
          finished = true;
          if (voicePlaying && voicePlaying.finish === finish) { voicePlaying = null; }
          if (litDone) { litDone(); litDone = null; }
          setTimeout(done, voiceGapMs);
        };
        var skip = function () {
          if (!finished) { voicesSkipped++; }
          finish();
        };
        try {
          var audio = voiceAudio(src);
          var voiceFeedId = feedIdNow;
          delete voiceCache[String(src)];
          var pooled = stingFromPool(src);
          if (pooled) { audio = pooled; }
          audio.addEventListener("ended", finish, { once: true });
          audio.addEventListener("error", skip, { once: true });
          voicePlaying = { audio: audio, finish: finish };
          var from = Date.now();
          (function playWhenShown() {
            var b = audio._bubble;
            var waiting = (b && b.isConnected && b.style.visibility === "hidden") ||
              pendingReveal.some(function (p) { return !(voiceFeedId && p._feedId > voiceFeedId); });
            if (waiting && Date.now() - from < CHAT_BATON_MAX_MS + BUBBLE_READY_MAX_MS) {
              setTimeout(playWhenShown, 50);
              return;
            }
            if (!audio._readyFrom) { audio._readyFrom = Date.now(); audio._shownAt = Date.now(); }
            if ((!audio._wordsDone || !(audio.duration > 0)) && !audio._failed && Date.now() - audio._readyFrom < VOICE_READY_MAX_MS) {
              setTimeout(playWhenShown, 50);
              return;
            }
            audio._readyFrom = 0;
            audio.volume = chatLevel(audio._share * audioLevel(audio.src));
            var capMs = isFinite(audio.duration) && audio.duration > 0 ? audio.duration * 1000 + 5000 : 60000;
            setTimeout(finish, capMs);
            if (!pooled && (voiceLitNext || voiceWaitFor[String(src).replace(/#.*$/, "")])) { litDone = voiceLitStart(audio.duration, audio._words, src, audio); }
            var startedAt = Date.now();
            if (startedAt - from >= 800) {
              var lateNote = new URLSearchParams();
              lateNote.set("note", "voice late " + ((startedAt - from) / 1000).toFixed(1) + " s: bubble " + (((audio._shownAt || startedAt) - from) / 1000).toFixed(1) + " s, file " + ((startedAt - (audio._shownAt || startedAt)) / 1000).toFixed(1) + " s - " + String(src).split("/").pop().replace(/#.*$/, ""));
              fetch("/api-sdk/gopod_page_note", { method: "POST", body: lateNote }).catch(function () {});
            }
            audio.play().catch(skip);
          })();
        } catch (err) { skip(); }
      }
      var skippingReporters = false, voicePlaying = null;
      var skipGate = { on: false, decided: false };
      function escapeRun() {
        fetch("/api-sdk/gopod_escape", { method: "POST" })
          .then(function (r) { return r.ok ? r.json() : { stopped: false }; })
          .then(function (d) { if (!d || !d.stopped) { skipReporters("Esc"); } })
          .catch(function () { skipReporters("Esc"); });
      }
      function skipReporters(how) {
        if (!skipGate.on) {
          fetch("/api-sdk/gopod_page_note", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: "note=" + encodeURIComponent("skip ignored (" + (how || "Esc") + "): not enabled") }).catch(function () {});
          return;
        }
        fetch("/api-sdk/gopod_page_note", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "note=" + encodeURIComponent("skipped (" + (how || "Esc") + ")")
        }).catch(function () { /* Wire-Pod not up */ });
        skippingReporters = true;
        if (voicePlaying) { try { voicePlaying.audio.pause(); } catch (err) { /* gone */ } voicePlaying.finish(); }
        skipFinishWaiting();
      }
      function skipEndWhenDrained() {
        var check = setInterval(function () {
          if (!skippingReporters) { clearInterval(check); return; }
          if (!feedQueue.length && !feedRunning) { skippingReporters = false; clearInterval(check); skipEndScroll(); }
        }, 500);
      }
      function skipFinishWaiting() {
        skipEndWhenDrained();
        Array.prototype.forEach.call(container.querySelectorAll(".brobot-bubble.voice-lit .brobot-say:not(.spoken)"), function (say) {
          var b = say.closest(".brobot-bubble");
          if (!b || b.classList.contains("brobot-1") || b.classList.contains("brobot-2")) { return; }
          if (say._wait && say._wait.parentNode) { say._wait.parentNode.removeChild(say._wait); }
          saidWhole(say);
        });
        voiceLitNext = null;
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
        if (((hostNames[e.who] || personaVoiceOf.hasOwnProperty(e.who)) && !managerHosts[e.who]) || e.who === systemRole) { return false; }
        skippingReporters = false;
        skipEndScroll();
        return false;
      }
      function skipEndScroll() {
        [300, 1200].forEach(function (ms) { setTimeout(function () { scrollChatTo(chatMaxScroll()); }, ms); });
      }
      var feedIdNow = 0; // the chat event being shown (a Brobot bubble keeps it: data-feed)
      var brobotBy = {}, hostPersonasNow = {}, settingsRunCheck = {}; // the name each Brobot answered as last; the served labels
      var brobotTo = {}; // 2026-11-14: who a line to a Brobot body reaches (its to_as)
      function showFeedEvent(e, done) {
        feedIdNow = e.id || 0;
        if (/^Brobot [12]$/.test(e.who)) { brobotBy[e.who] = e.by || e.who; }
        if (/^Brobot [12]$/.test(e.to || "")) { brobotTo[e.to] = e.to_as || e.to; }
        var rcs = settingsRunCheck || {};
        if (e.who === reporterRole && (e.text === rcs.mute_marker || e.text === rcs.unmute_marker)) { runCheckMuted = e.text === rcs.mute_marker; done(); return; }
        if (e.who === reporterRole && String(e.text).indexOf("skip:") === 0) {
          if (!skipGate.on) { done(); return; }
          skippingReporters = true; if (voicePlaying) { try { voicePlaying.audio.pause(); } catch (err) { /* gone */ } voicePlaying.finish(); } skipFinishWaiting(); done(); return; }
        if (skippedEvent(e)) { done(); return; }
        if (countdownDoneAt && e.who !== reporterRole && e.who !== countdownRole && e.who !== songAudioRole) {
          afterCountdown(function () { showFeedEvent(e, done); }, e.who === ballRole);
          return;
        }
        if (e.who === songAudioRole) {
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
          if (String(e.text).trim() === "loop_pause" || String(e.text).trim() === "loop_resume") {
            backgroundLoop.runQuiet(String(e.text).trim() === "loop_pause");
            done();
            return;
          }
          if (String(e.text).indexOf("numpad_sim|") === 0) {
            var sim = String(e.text).split("|"), simKey = sim[1] || "", simWho = (numpad.targets && numpad.targets[simKey]) || sim[2] || "";
            if (restoringChat) { done(); return; }
            if (/^Brobot [12]$/.test(simWho) && sim[3]) { brobotTo[simWho] = sim[3]; }
            var simSt = systemStatus(numpadText("hold", { key: simKey, brobot: simWho }), 0, simWho);
            var holdMs = ((numpad.cfg && numpad.cfg.min_hold_seconds) || 0.5) * 1000, recMs = (commCheckRecordSeconds || 2) * 1000;
            setTimeout(function () {
              simSt.set(numpadText("recording", { brobot: simWho }));
              setTimeout(function () {
                simSt.set(numpadText("transcribing"));
                numpadOpStatus = simSt;
                done();
              }, recMs);
            }, holdMs);
            return;
          }
          if (String(e.text).trim() === "skip_on") {
            if (!restoringChat) { skipGate.on = true; skipGate.decided = true; }
            done();
            return;
          }
          if (String(e.text).indexOf("warmup_text|") === 0) {
            var warmVoices = [];
            for (var q = 0; q < feedQueue.length && isVoiceEvent(feedQueue[q].e); q++) { warmVoices.push(voiceAudio(feedQueue[q].e.text)); }
            var warmFrom = performance.now();
            (function whenWarmVoiceReady() {
              var loading = warmVoices.some(function (a) { return !a._failed && a.readyState < 3; });
              if (loading && performance.now() - warmFrom < BUBBLE_READY_MAX_MS) { setTimeout(whenWarmVoiceReady, 50); return; }
              var held = warmupTextLine(e.text);
              setTimeout(done, held && !warmVoices.length ? llmWarm.checkMs : 0);
            })();
            return;
          }
          if (generatingMarker && String(e.text).indexOf(generatingMarker) === 0) {
            var genKind = String(e.text).slice(generatingMarker.length);
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
        lineWaits = false;
        lineVoice = "";
        if (generatingMarker && String(e.text).indexOf(generatingMarker + "lit:") === 0) {
          e.text = String(e.text).slice((generatingMarker + "lit:").length);
          var litBar = e.text.indexOf("|");
          if (litBar > 0 && e.text.charAt(0) === "/") { lineVoice = e.text.slice(0, litBar); e.text = e.text.slice(litBar + 1); }
        }
        if (generatingMarker && String(e.text).indexOf(generatingMarker + "wait:") === 0) {
          e.text = String(e.text).slice((generatingMarker + "wait:").length);
          lineWaits = true;
          var bar = e.text.indexOf("|");
          if (bar > 0 && e.text.charAt(0) === "/") { lineWaits = e.text.slice(0, bar); e.text = e.text.slice(bar + 1); }
        }
        if (e.who !== systemRole && e.who !== scoreRole && !hostNames[e.who] && !personaVoiceOf.hasOwnProperty(e.who)) { generatingGone(); }
        if (e.who === countdownRole) {
          if (String(e.text).trim() === countdownEndText) { countdownHold(); }
          else { showCountdown(parseInt(e.text, 10), e.time_ms, done); return; }
        }
        else if (e.who === ballRole) { showBingoBall(e.text); }
        else if (e.who === chapterRole) { showChapter(e.text, done); return; }
        else if (e.who === scoreRole) { showScore(e.text, undefined, done); return; }
        else if (e.who === systemRole) { showScore(e.text, undefined, done); return; }
        else if (e.who === terminalRole) { showTerminal(e.text, done); return; }
        else if (searchSettings.results_role && e.who === searchSettings.results_role) { showSearchResults(e.text, e.to, done); return; }
        else if (e.who === pictureRole && String(e.text).indexOf("notice:") === 0) {
          recordBubble(e.to || "", String(e.text).slice(7), "notice");
          bubbleNotice(e.to, String(e.text).slice(7));
        }
        else if (e.who === pictureRole && String(e.text).indexOf("opener:") === 0) { brobotOpener(e.to, String(e.text).slice(7), done); return; }
        else if (e.who === pictureRole && String(e.text).indexOf("noticesay:") === 0) { bubbleNotice(e.to, String(e.text).slice(10), true); }
        else if (e.who === pictureRole && String(e.text).indexOf("noticesaid:") === 0) { bubbleNoticeSaid(e.to); }
        else if (e.who === pictureRole && String(e.text).indexOf("weathershow:") === 0) { weatherShowFor(e.to); }
        else if (e.who === pictureRole && String(e.text).indexOf("eggmark:") === 0) { if (/^Brobot [12]$/.test(String(e.to || ""))) { eggMarked[e.to] = true; } }
        else if (e.who === pictureRole) {
          if (menuPictures(e.text)) { done(); return; }
          showPictures(String(e.text).replace(/^menu:/, ""), done);
          return;
        }
        else if (hostNames[e.who] || personaVoiceOf.hasOwnProperty(e.who)) { showHostLine(e.who, e.text, e.to, done); return; }
        else if (e.who === "Operator" && e.by && (hostNames[e.by] || personaVoiceOf.hasOwnProperty(e.by) || e.by === systemRole)) {
          appendBubble("Operator", e.text, undefined, e.to, hostHeader(e.by), done, e.by);
          return;
        }
        else if (e.who === "Operator") { appendBubble("Operator", e.text, undefined, e.to, e.via, done); return; }
        else if (e.via === "llm") { noteLiveReply(e, robotBubbleOnce(e.who, kgReplyText(e.text), "llm", done)); return; }
        else { robotBubbleOnce(e.who, e.text, "brobot", done); return; }
        done();
      }

      var ESN_TO_BUBBLE_CLASS = { "ESN01": "brobot-1", "ESN02": "brobot-2" };
      function lightSpan(span) {
        span.classList.add("speaking");
        span._litAt = Date.now();
        if (span.closest(".brobot-bubble-live, .brobot-bubble-scripted, .narrator, .voice-lit")) { digitSequence(span); }
        var bubble = span.closest(".brobot-bubble");
        if (!bubble) { return; }
        var all = span.closest(".nested-host") ? span.closest(".nested-host").querySelectorAll(".brobot-say") : ownSays(bubble);
        for (var i = 0; i < all.length && all[i] !== span; i++) { all[i].classList.add("spoken"); }
        focusSpeaking(span);
      }
      var digitHighlight = { seconds_per_word: 0.3, seconds_per_number: 0.45 };
      var militaryLabel;
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
            cuts.push([at, at + 2, n]);
            cuts.push([at + 2, at + 4, n]); // MM, one group too
          }
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
          var year = /\b(?:January|February|March|April|May|June|July|August|September|October|November|December) \d{1,2}, (\d{2})(\d{2})\b/g;
          while ((m = year.exec(text))) {
            var y0 = m.index + m[0].length - 4;
            cuts.push([y0, y0 + 2, wordReveal.seconds_per_word]);
            cuts.push([y0 + 2, y0 + 4, wordReveal.seconds_per_word]);
          }
          cuts.sort(function (a, b) { return a[0] - b[0]; });
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
        var at = (wordReveal.start_delay_words || 0) * (wordReveal.seconds_per_word || 0);
        if (span.dataset.leadIn === "1" && wordReveal.lead_in_start_seconds > 0) { at = wordReveal.lead_in_start_seconds; }
        var srtAt = parseFloat(span.dataset.at), srtEnd = parseFloat(span.dataset.end);
        var scale = 1;
        if (!isNaN(srtAt) && !isNaN(srtEnd) && srtEnd > srtAt) {
          var total = 0;
          units.forEach(function (u) { total += u._seqWeight || 0; });
          if (total > 0) { scale = (srtEnd - srtAt) / total; }
          at = 0;
        }
        var aligned = span._spokenWords ? alignUnits(units, span._spokenWords) : null;
        var starts = [];
        units.forEach(function (u, i) {
          starts.push(aligned ? aligned[i] : at);
          at += (u._seqWeight || 0) * scale;
        });
        if (aligned) { at = Math.max(at, aligned[aligned.length - 1] || 0); }
        var voiceEnd = aligned && span._spokenWords.length ? +span._spokenWords[span._spokenWords.length - 1].e || at : at;
        var step = function (i) {
          var u = units[i];
          var said = Math.max(0, (i + 1 < starts.length ? starts[i + 1] : Math.max(voiceEnd, at)) - starts[i]);
          if (!span.classList.contains("speaking")) { return; }
          if (i > 0) { units[i - 1].classList.remove("now"); units[i - 1].classList.add("said"); }
          u.classList.add("now");
          if (window.GopodCharHighlight) { window.GopodCharHighlight.play(u, said); }
        };
        var allSaid = function () {
          units.forEach(function (u) { u.classList.remove("now"); u.classList.add("said"); });
        };
        if (aligned && span._voice) {
          units.forEach(function (u, i) { voiceAt(span._voice, (span._voiceFrom || 0) + starts[i], function () { step(i); }); });
          voiceAt(span._voice, (span._voiceFrom || 0) + at, allSaid);
          return;
        }
        units.forEach(function (u, i) { setTimeout(function () { step(i); }, starts[i] * 1000); });
        setTimeout(allSaid, at * 1000);
      }
      function voiceAt(audio, t, fn) {
        if (!audio) { setTimeout(fn, t * 1000); return; }
        var began = Date.now();
        var tick = function () {
          if (Date.now() - began > (t + 30) * 1000 ||
              audio.ended || audio._failed || (audio.currentTime || 0) >= t || (audio.paused && audio.currentTime > 0)) { fn(); return; }
          if (document.hidden) { setTimeout(tick, 250); } else { requestAnimationFrame(tick); }
        };
        tick();
      }

      function alignUnits(units, words) {
        var norm = function (t) { return String(t || "").toLowerCase().replace(/[^a-z0-9]+/g, ""); };
        var unitText = function (v) { return v ? norm(v.tagName === "IMG" ? "" : v.textContent) : ""; };
        var a = units.map(unitText), b = words.map(function (w) { return norm(w.t); });
        var n = a.length, m = b.length, L = [], i, j;
        for (i = 0; i <= n; i++) { L.push(new Array(m + 1).fill(0)); }
        for (i = n - 1; i >= 0; i--) {
          for (j = m - 1; j >= 0; j--) {
            L[i][j] = (a[i] && a[i] === b[j]) ? L[i + 1][j + 1] + 1 : Math.max(L[i + 1][j], L[i][j + 1]);
          }
        }
        var s = new Array(n).fill(null), e = new Array(n).fill(null);
        i = 0; j = 0;
        while (i < n && j < m) {
          if (a[i] && a[i] === b[j]) { s[i] = +words[j].s || 0; e[i] = +words[j].e || 0; i++; j++; }
          else if (L[i + 1][j] >= L[i][j + 1]) { i++; }
          else { j++; }
        }
        var lastEnd = words.length ? +words[words.length - 1].e || 0 : 0;
        for (i = 0; i < n; i++) {
          if (s[i] !== null) { continue; }
          var k = i;
          while (k < n && s[k] === null) { k++; }
          var from = i > 0 ? e[i - 1] : 0, to = k < n ? s[k] : Math.max(from, lastEnd);
          var total = 0, q;
          for (q = i; q < k; q++) { total += a[q].length; }
          var at = from;
          for (q = i; q < k; q++) {
            var share = total ? (to - from) * a[q].length / total : 0;
            s[q] = at; e[q] = at + share; at += share;
          }
          i = k - 1;
        }
        return s;
      }
      function ownSays(bubble) {
        return Array.prototype.filter.call(bubble.querySelectorAll(".brobot-say"), function (x) { return !x.closest(".nested-host, .brobot-opener"); });
      }
      function revealRest(bubble) {
        Array.prototype.forEach.call(ownSays(bubble).filter(function (x) { return !x.classList.contains("spoken"); }), function (s) {
          if (!s.classList.contains("speaking")) { s.classList.add("spoken"); }
        });
      }
      var REVEAL_REST_MS;
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
      var speakingPolling = false;
      function pollSpeakingNow() {
        if (speakingPolling) { return; }
        speakingPolling = true;
        fetch("/gopod_speaking_now.json", { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (now) {
            if (!now) { return; }
            var bubbles = ESN_TO_BUBBLE_CLASS[now.esn] ? container.querySelectorAll(".brobot-bubble." + ESN_TO_BUBBLE_CLASS[now.esn]) : [];
            var newest = null;
            var named = now.feed ? container.querySelector('.brobot-bubble[data-feed="' + now.feed + '"]') : null;
            if (named) { newest = named; }
            for (var b = bubbles.length - 1; !named && b >= 0; b--) {
              if (!bubbles[b].classList.contains("popup-wrap") && ownSays(bubbles[b]).length) { newest = bubbles[b]; break; }
            }
            if (!named && now.active && typeof now.piece === "number" && now.piece > 0 && lastSpeakingBubble && lastSpeakingBubble.isConnected &&
                newest !== lastSpeakingBubble && ownSays(lastSpeakingBubble)[now.piece]) {
              newest = lastSpeakingBubble;
            }
            if (now.seq === lastSpeakingSeq && newest === lastSpeakingBubble) { return; }
            lastSpeakingSeq = now.seq;
            lastSpeakingBubble = newest;
            clearSpeaking();
            if (!now.active || !newest) {
              if (!now.active && newest) { revealRestLater(newest); }
              return;
            }
            var spans = ownSays(newest);
            if (typeof now.piece === "number" && now.piece >= 0) {
              if (spans[now.piece]) { spans[now.piece].dataset.leadIn = leadInText && String(now.text || "").trim().indexOf(leadInText) === 0 ? "1" : ""; lightSpan(spans[now.piece]); }
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

      var logButton = document.querySelector('a[onclick^="showLog"]');
      if (logButton) {
        logButton.addEventListener("click", function () {
          window.requestAnimationFrame(function () {
            window.scrollTo(0, document.documentElement.scrollHeight);
          });
        });
      }

      var botOrder = [], botOrderAsked = 0;
      function orderBotStats() {
        if (!isAdminPage) { return; }
        if (!botOrder.length && Date.now() - botOrderAsked > 10000) {
          botOrderAsked = Date.now();
          fetch("/api-sdk/gopod_brobot_order", { cache: "no-store" })
            .then(function (r) { return r.ok ? r.json() : []; })
            .then(function (list) { if (Array.isArray(list)) { botOrder = list.map(function (e) { return String(e).toLowerCase(); }); } })
            .catch(function () { /* Wire-Pod gone */ });
        }
        var boxes = document.querySelectorAll("#botStats .batteryContainer");
        Array.prototype.forEach.call(boxes, function (box) {
          var serial = box.querySelector(".tooltip b");
          var place = serial ? botOrder.indexOf(serial.textContent.trim().toLowerCase()) : -1;
          if (place >= 0 && box.style.order !== String(place)) { box.style.order = String(place); }
        });
      }
      setInterval(orderBotStats, 1000);

      var numpad = {
        on: (function () { try { return window.localStorage.getItem(NUMPAD_KEY) === "1"; } catch (e) { return false; } })(), map: null, targets: {}, active: null, pendingUp: null, taps: { Numpad0: [], NumLock: [] },
        cfg: { min_hold_seconds: 1, max_hold_seconds: 30, exit_taps: 3, exit_window_seconds: 2, texts: {} }
      };
      numpadMapLoad = function () {
        fetch("/api-sdk/gopod_numpad_map?layout=" + encodeURIComponent(numpadLayoutNow()), { cache: "no-store" })
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (data) {
            if (data && data.map) { numpad.map = data.map; numpad.targets = data.targets || {}; numpad.personaReporters = data.persona_reporters || {}; relabelPicker(); }
            if (data && data.layout && numpadLayoutSelect && !numpadLayoutNow()) { numpadLayoutSelect.value = data.layout; }
          }).catch(function () { /* no map - push-to-talk stays quiet */ });
      };
      numpadMapLoad();
      function numpadText(name, values) {
        if ((numpad.cfg.error_texts || []).indexOf(name) >= 0) { bubbleSound("Operator", "error"); numpad.errorNext = true; }
        var t = (numpad.cfg.texts || {})[name] || name;
        return String(t).replace(/\{(\w+)\}/g, function (m, k) { return values && values[k] != null ? (k === "key" ? keyShown(values[k]) : values[k]) : m; });
      }
      function keyShown(key) { return String(key).replace(/^KPmul$/i, "KP*"); }
      function systemStatus(text, goneAfterMs, numpadTo) {
        if (systemMenuResults()) {
          systemMenuResult(text, "answer");
          return { bubble: null, textEl: null, set: function (t) { systemMenuResult(t, "answer"); }, done: function () {} };
        }
        var bubble = rawBubble({ kind: "status", classes: "operator system-status", head: numpadTo ? operatorHead(numpadTo, "Numpad", "") : systemHead(systemLabel), text: text });
        if (numpadTo) { bubbleSound("Operator"); }
        var textEl = bubble.querySelector(".brobot-bubble-text");
        if (numpad && numpad.errorNext) { bubble.classList.add("is-error"); numpad.errorNext = false; }
        placeBubble(bubble);
        var api = {
          bubble: bubble, textEl: textEl,
          set: function (t) {
            if (textEl.textContent === t) { return; }
            if (numpad && numpad.errorNext) { bubble.classList.add("is-error"); numpad.errorNext = false; }
            textEl.appendChild(document.createElement("br"));
            textEl.appendChild(document.createTextNode(t));
            if (lastPlacedChild() === bubble) { scrollChatTo(Math.max(container.scrollTop, chatContentEnd())); }
          },
          done: function () { /* stays in the chat */ }
        };
        if (goneAfterMs) { api.done(goneAfterMs); }
        return api;
      }
      var numpadOpStatus = null, commCheckRecordSeconds = 2;
      var visualBoard = null, visualBusy = false;
      function visualTurn(dir) {
        if (visualBusy) { return; }
        visualBusy = true;
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
        (v.frames && v.frames.length ? v.frames : (v.frame ? [v.frame] : [])).forEach(function (src) {
          var img = document.createElement("img");
          img.src = src;
          img.alt = img.title = v.label || v.id;
          magnify(row, img);
          row.appendChild(img);
        });

      }
      var systemMenu = null, systemMenuData = [], systemMenuIntro = "", systemMenuLook = {};
      var systemMenuCopy = {}, systemMenuLastAsk = "";
      function copyButton(text, look) {
        look = look || systemMenuCopy;
        var b = document.createElement("button");
        b.type = "button";
        b.className = "gopod-result-copy";
        b.textContent = look.icon || "";
        b.title = look.title || "";
        b.addEventListener("click", function (ev) {
          ev.stopPropagation();
          var done = function () {
            b.textContent = look.copied || b.textContent;
            setTimeout(function () { b.textContent = look.icon || ""; }, look.copied_ms || 1500);
          };
          var words = typeof text === "function" ? text() : text;
          var fallback = function () {
            var area = document.createElement("textarea");
            area.value = words;
            area.style.position = "fixed";
            area.style.opacity = "0";
            document.body.appendChild(area);
            area.select();
            try { document.execCommand("copy"); done(); } catch (err) { /* no clipboard */ }
            area.remove();
          };
          if (navigator.clipboard && window.isSecureContext) {
            navigator.clipboard.writeText(words).then(done, fallback);
          } else {
            fallback();
          }
        });
        return b;
      }
      function systemMenuResults() {
        return systemMenu && systemMenu.isConnected ? systemMenu.querySelector(".system-menu-results") : null;
      }
      function systemMenuResult(text, kind) {
        var box = systemMenuResults();
        if (!box) { return false; }
        if (!box.querySelector(".gopod-results-copy-all")) {
          var bar = document.createElement("div");
          bar.className = "gopod-results-copy-all";
          bar.appendChild(copyButton(function () {
            var out = [];
            Array.prototype.forEach.call(box.querySelectorAll(":scope > .system-menu-result"), function (r) {
              if (r.classList.contains("ask") && out.length) { out.push(""); }
              out.push(r._copy != null ? r._copy : (r.innerText || r.textContent || "").trim());
            });
            return out.join("\n");
          }, systemMenuCopy.all || systemMenuCopy));
          box.insertBefore(bar, box.firstChild);
        }
        var row = document.createElement("div");
        row.className = "system-menu-result " + kind;
        row._copy = kind === "ask" ? "\u203a " + text : String(text);
        if (kind === "answer") {
          row.appendChild(resultTemplate(text));
          row.appendChild(copyButton("\u203a " + systemMenuLastAsk + "\n" + String(text)));
        } else {
          if (kind === "ask") { systemMenuLastAsk = String(text); }
          row.textContent = kind === "ask" ? "\u203a " + text : text;
        }
        box.appendChild(row);
        box.scrollTop = box.scrollHeight;
        return true;
      }
      function systemMenuOpen() {
        if (systemMenu && systemMenu.isConnected) { return; }
        var outer = document.createElement("div");
        outer.className = "brobot-bubble operator system-status system-menu-host gopod-admin-only";
        var look = { font_size: "--menu-font", background: "--menu-bg", code_background: "--menu-code-bg", code_color: "--menu-code-fg", results_background: "--menu-results-bg", texture: "--menu-texture", border: "--menu-border", accent: "--menu-accent", rule: "--menu-rule" };
        Object.keys(look).forEach(function (k) { if (systemMenuLook[k]) { outer.style.setProperty(look[k], systemMenuLook[k]); } });
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
              if (it.page === "voices") { voicesPanel(); return; }
              if (it.page === "skip") { skipReporters("menu"); return; }
              if (it.run) {
                var numpadCmd = /^system-numpad\s+(on|off)\b/i.exec(it.cmd);
                if (numpadCmd) { setNumpadOn(numpadCmd[1].toLowerCase() === "on"); }
                sendTypedLine(/^system-/i.test(it.cmd) ? "system" : "", it.cmd);
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
            if (group.open) {
              Array.prototype.forEach.call(menu.querySelectorAll("details"), function (d) { if (d !== group) { d.open = false; } });
            }
          });
          menu.appendChild(group);
        });
        menu.appendChild(visitorsGroup(menu));
        outer.appendChild(menu);
        function menuFitSlot() {
          var open = menu.querySelector("details[open]");
          requestAnimationFrame(function () {
            if (!open) {
              menu.style.flex = "";
              menu.style.height = "";
              menu.scrollTop = 0;
              return;
            }
            var pad = parseFloat(getComputedStyle(menu).paddingTop) || 0;
            var room = outer.getBoundingClientRect().height * 0.85;
            var want = Math.min(open.getBoundingClientRect().height + pad, room);
            menu.style.flex = "0 0 auto";
            menu.style.height = want + "px";
            menu.scrollTop += open.getBoundingClientRect().top - menu.getBoundingClientRect().top - pad;
          });
        }
        menu.addEventListener("toggle", menuFitSlot, true);
        var split = document.createElement("div");
        split.className = "system-menu-split";
        split.title = pageText("drag_title");
        outer.appendChild(split);
        split.addEventListener("pointerdown", function (ev) {
          ev.preventDefault();
          split.setPointerCapture(ev.pointerId);
          function move(e) {
            var box = outer.getBoundingClientRect();
            var share = Math.max(0.05, Math.min(0.95, (e.clientY - box.top) / box.height));
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
        (function pin() {
          if (systemMenu !== outer || !outer.isConnected) { return; }
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
      var usbTimer = null;
      function voicesPanel() {
        var results = systemMenuResults();
        if (!results) { return; }
        var panel = results.querySelector(".voices-panel");
        if (!panel) {
          panel = document.createElement("div");
          panel.className = "system-menu-result voices-panel";
          results.appendChild(panel);
        }
        function draw(data, note) {
          panel.textContent = "";
          var used = {};
          (data.rows || []).forEach(function (r) { if (!r.bodied && r.voice) { used[r.voice] = (used[r.voice] || 0) + 1; } });
          var title = document.createElement("div");
          title.textContent = pageText("voices_title");
          panel.appendChild(title);
          (data.rows || []).forEach(function (r) {
            var row = document.createElement("div");
            row.className = "voices-row";
            var label = document.createElement("span");
            label.textContent = r.key + " > " + (r.persona || "-") + " ";
            row.appendChild(label);
            if (r.bodied || !r.host) {
              var none = document.createElement("span");
              none.textContent = r.bodied ? "no voice required" : "-";
              row.appendChild(none);
            } else {
              var pick = document.createElement("select");
              if (r.key === "Operator") {
                var un = document.createElement("option");
                un.value = "";
                un.textContent = pageText("voices_unassigned", { host: data.unassigned_host || "-" });
                pick.appendChild(un);
              }
              (data.voices || []).forEach(function (v) {
                var o = document.createElement("option");
                o.value = v;
                o.textContent = v + (used[v] > 1 ? " (" + used[v] + ")" : "");
                if (used[v] > 1) { o.style.fontWeight = "bold"; }
                pick.appendChild(o);
              });
              pick.value = r.voice || "";
              if (r.voice && used[r.voice] > 1) { pick.classList.add("voice-twice"); }
              pick.addEventListener("change", function () {
                fetch("/api-sdk/gopod_voice_map", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" },
                  body: "key=" + encodeURIComponent(r.key) + "&voice=" + encodeURIComponent(pick.value) })
                  .then(function (x) { return x.json(); })
                  .then(function (out) { load(out.ok ? out.message : "not changed: " + (out.error || "?")); })
                  .catch(function () { load("Wire-Pod didn't answer"); });
              });
              row.appendChild(pick);
            }
            panel.appendChild(row);
          });
          if (note) {
            var n = document.createElement("div");
            n.textContent = note;
            panel.appendChild(n);
          }
          results.scrollTop = results.scrollHeight;
        }
        function load(note) {
          fetch("/api-sdk/gopod_voice_map", { cache: "no-store" }).then(function (x) { return x.json(); })
            .then(function (data) { draw(data || {}, note); })
            .catch(function () { panel.textContent = pageText("no_answer", { what: pageText("voices_title").replace(/:\s*$/, "") }); });
        }
        load("");
      }
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
          }).catch(function () { panel.textContent = pageText("no_answer", { what: pageText("usb_title") }); });
        }
        poll();
        usbTimer = setInterval(poll, 3000);
        results.scrollTop = results.scrollHeight;
      }
      function bubbleNoticeSaid(to) {
        var cls = to === "Brobot 2" ? "brobot-2" : "brobot-1";
        Array.prototype.forEach.call(container.querySelectorAll(".brobot-bubble." + cls + " .bubble-notice .brobot-say.speaking"), function (x) {
          x.classList.remove("speaking");
          x.classList.add("spoken");
        });
      }
      var scoreBarDone = "=";
      function noticeLine(line, said) {
        var el = document.createElement("span");
        el.className = "brobot-bubble-text egg-summary-text";
        var tab = String(line).indexOf("\t");
        if (tab >= 0 && !said) {
          el.classList.add("notice-score-row");
          var label = document.createElement("span");
          label.className = "notice-score-label";
          label.textContent = String(line).slice(0, tab).trim();
          var marks = String(line).slice(tab + 1).trim();
          var done = marks.split(scoreBarDone).length - 1;
          var bar = document.createElement("span");
          bar.className = "notice-score-bar";
          bar.title = marks;
          var fill = document.createElement("span");
          fill.className = "notice-score-fill";
          fill.style.width = (marks.length ? Math.min(100, 100 * done / marks.length) : 0) + "%";
          bar.appendChild(fill);
          el.appendChild(label);
          el.appendChild(bar);
          return el;
        }
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
      function brobotOpener(to, line, done) {
        if (!/^Brobot [12]$/.test(String(to || "")) || restoringChat || !line) { done(); return; }
        var shell = nestShell(to);
        var el = document.createElement("div");
        el.className = "brobot-opener";
        var say = document.createElement("span");
        say.className = "brobot-say";
        say.textContent = line;
        el.appendChild(say);
        var head = shell.querySelector(":scope > .brobot-bubble-head");
        var first = Array.prototype.filter.call(shell.children, function (c) { return c !== head && !c.classList.contains("brobot-opener"); })[0];
        if (first) { shell.insertBefore(el, first); } else { shell.appendChild(el); }
        replaySaidLine(say, done);
      }
      function replaySaidLine(say, done) {
        if (restoringChat || skippingReporters) { say.classList.add("spoken"); done(); return; }
        var seconds = Math.max(0.5, sayWords(say.textContent).length * (wordReveal.seconds_per_word || 0.38));
        var bubble = say.closest(".brobot-bubble");
        if (bubble) { bubble.classList.add("voice-lit"); }
        say.dataset.at = "0";
        say.dataset.end = String(seconds);
        var settle = (wordReveal.replay_settle_seconds || 0) * 1000;
        setTimeout(function () {
          if (!say.isConnected) { done(); return; }
          lightSpan(say);
          setTimeout(function () {
            say.classList.remove("speaking");
            say.classList.add("spoken");
            done();
          }, seconds * 1000 + 100);
        }, settle);
      }
      var hostBrains = {};
      function hostBrain(bubble, who) {
        var lines = hostBrains[who];
        delete hostBrains[who];
        if (!lines || !lines.length) { return; }
        var boxes = {};
        lines.forEach(function (n) {
          var box = boxes[n.title] || Array.prototype.filter.call(bubble.querySelectorAll(":scope > .bubble-notice"), function (x) { return x.dataset.title === n.title; })[0];
          if (!box) {
            box = boxes[n.title] = nestPart("notice", "egg-summary bubble-notice", systemHead(n.title), String(n.line).indexOf("\t") >= 0 ? null : n.line, "egg-summary-text");
            if (String(n.line).indexOf("\t") >= 0) { box.appendChild(noticeLine(n.line)); }
            box.dataset.title = n.title;
            nestPlace(bubble, box, true);
          } else {
            box.appendChild(noticeLine(n.line));
          }
        });
      }
      function bubbleNotice(to, text, said) {
        var bar = text.indexOf("|");
        var title = bar >= 0 ? text.slice(0, bar) : text, line = bar >= 0 ? text.slice(bar + 1) : "";
        if (!/^Brobot [12]$/.test(String(to || ""))) {
          if (to) { (hostBrains[to] = hostBrains[to] || []).push({ title: title, line: line }); }
          var receipt = to ? openReceipt(to) : null;
          if (receipt) { hostBrain(receipt, to); }
          return;
        }
        var shell = nestShell(to);
        var same = Array.prototype.filter.call(shell.querySelectorAll(":scope > .bubble-notice"), function (n) { return n.dataset.title === title; })[0];
        if (same) {
          same.appendChild(noticeLine(line, said));
          brainInView(same);
          return;
        }
        var box = nestPart("notice", "egg-summary bubble-notice", systemHead(title), String(line).indexOf("\t") >= 0 ? null : line, "egg-summary-text");
        if (String(line).indexOf("\t") >= 0) { box.appendChild(noticeLine(line, said)); }
        box.dataset.title = title;
        nestPlace(shell, box, true);
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
      var visitorsTimer = null;
      function visitorsGroup(menu) {
        var group = document.createElement("details");
        group.className = "system-menu-group system-menu-visitors";
        var title = document.createElement("summary");
        title.textContent = pageText("visitors_title");
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
          title.textContent = pageText("visitors_count", { n: list.length });
          if (!list.length) {
            var none = document.createElement("div");
            none.className = "system-visitor";
            none.textContent = pageText("visitors_none");
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
              who.textContent = pageText("visitor_asks", { name: v.name });
            } else {
              who.textContent = pageText("visitor_nameless");
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
          body: "action=" + action + "&key=" + encodeURIComponent(key) + "&layout=" + encodeURIComponent(numpadLayoutNow()) + (allFirstSelect && allFirstSelect.value === "-*" ? "&first=2" : "&first=1")
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
      function numpadSay(text) {
        fetch("/api-sdk/gopod_system_say", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: "text=" + encodeURIComponent(text)
        }).then(function (r) { if (!r.ok) { systemStatus(text); } })
          .catch(function () { systemStatus(text); });
      }
      function numpadKeySound() { playCue(numpad.cfg.key_sound); }
      var bubbleSounds = {}, leadInText = "";
      function bubbleSound(kind, label) {
        var cfg = bubbleSounds[kind];
        if (!cfg || restoringChat) { return; }
        if (label === "error") { if (cfg.error && cfg.error.file) { playCue(cfg.error); } return; }
        if (cfg.mute) { return; }
        playCue({ file: cfg.file || bubbleSounds.default_file, volume: cfg.volume });
      }
      function playCue(ks) {
        if (!ks || !ks.file) { return; }
        try {
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
      function typingInBox() {
        var el = document.activeElement;
        return !!el && (el.tagName === "INPUT" && /^(text|search)?$/i.test(el.type || "") || el.tagName === "TEXTAREA" || el.isContentEditable);
      }
      document.addEventListener("keydown", function (ev) {
        var dotKey = ev.code === "NumpadDecimal" || (ev.location === 3 && (ev.key === "." || ev.key === "," || ev.key === "Delete" || ev.key === "Decimal"));
        var numpadKey = /^Numpad[0-9]$/.test(ev.code) || ev.code === "NumLock" || ev.code === "NumpadMultiply" || ev.code === "NumpadDivide" || dotKey;
        if (ev.location === 3 && !numpadKey && !/^Numpad(Add|Subtract|Enter)$/.test(ev.code) && !ev.repeat) {
          var unk = new URLSearchParams(); unk.set("note", "numpad key not mapped: code=" + ev.code + " key=" + ev.key);
          fetch("/api-sdk/gopod_page_note", { method: "POST", body: unk }).catch(function () {});
        }
        if (!numpad.on || !numpad.map || !chatShown() || (typingInBox() && !numpadKey)) { return; }
        if (numpadKey && !ev.repeat) { numpadKeySound(); }
        if (ev.code === "NumLock") {
          if (!ev.repeat) { numpadExitTap(ev.code); }
          return;
        }
        if (ev.code === "NumpadDivide") {
          ev.preventDefault();
          if (!ev.repeat) { skipReporters("numpad /"); }
          return;
        }
        if (ev.code === "NumpadMultiply") {
          ev.preventDefault();
          if (ev.repeat || numpad.active) { return; }
          if (numpad.pendingUp && numpad.pendingUp.key === "KPmul") {
            clearTimeout(numpad.pendingUp.timer);
            numpad.pendingUp = null;
            return;
          }
          if (!ev.getModifierState("NumLock")) { systemStatus(numpadText("numlock_off", { key: "*" }), 3000, "Brobots"); return; }
          numpadStart("KPmul", numpad.targets.KPmul || numpad.cfg.all_label || "All");
          return;
        }
        if (dotKey) {
          ev.preventDefault();
          if (ev.repeat || numpad.active) { return; }
          if (numpad.pendingUp && numpad.pendingUp.key === "KPdot") {
            clearTimeout(numpad.pendingUp.timer);
            numpad.pendingUp = null;
            return;
          }
          if (!numpad.targets.KPdot) { systemStatus(numpadText("not_live", { key: ".", persona: "-", body: "-" }), 3000, "."); return; }
          numpadStart("KPdot", numpad.targets.KPdot);
          return;
        }
        if ((ev.code === "NumpadAdd" || ev.code === "NumpadSubtract") && !typingInBox()) {
          ev.preventDefault();
          if (!ev.repeat) { visualTurn(ev.code === "NumpadAdd" ? "next" : "prev"); }
          return;
        }
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
              if (!ev.getModifierState("NumLock")) {
                if (numpad.targets.KP0off) { numpadStart("KP0off", numpad.targets.KP0off); numpad.active.at = downAt; return; }
                systemStatus(numpadText("numlock_off", { key: "KP0" }), 3000, "Brobots"); return;
              }
              if (!numpad.targets.KP0) { systemStatus(numpadText("not_live", { key: "KP0", persona: "-", body: "-" }), 3000, "KP0"); return; }
              numpadStart("KP0", numpad.targets.KP0);
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
        if (numpad.pendingUp && numpad.pendingUp.key === key) {
          clearTimeout(numpad.pendingUp.timer);
          numpad.pendingUp = null;
          return;
        }
        if (numpad.active) { return; }
        if (!ev.getModifierState("NumLock")) {
          if (numpad.targets[key + "off"]) { numpadStart(key + "off", numpad.targets[key + "off"]); return; }
          systemStatus(numpadText("numlock_off", { key: key }), 3000, "Brobots"); return;
        }
        var target = numpad.targets[key];
        if (!target) {
          var entry = (numpad.map.numlock_on || {})[key];
          systemStatus(numpadText("not_live", { key: key, persona: entry ? entry.persona : "-", body: entry ? entry.body : "-" }), 3000, entry && entry.persona ? String(entry.persona) : key);
          return;
        }
        numpadStart(key, target);
      });
      document.addEventListener("keyup", function (ev) {
        var dotUp = ev.code === "NumpadDecimal" || (ev.location === 3 && (ev.key === "." || ev.key === "," || ev.key === "Delete" || ev.key === "Decimal"));
        if (dotUp && numpad.active && numpad.active.key === "KPdot") {
          ev.preventDefault();
          numpad.pendingUp = { key: "KPdot", timer: setTimeout(function () { numpad.pendingUp = null; numpadRelease(); }, 60) };
          return;
        }
        if (ev.code === "NumpadMultiply" && numpad.active && numpad.active.key === "KPmul") {
          ev.preventDefault();
          numpad.pendingUp = { key: "KPmul", timer: setTimeout(function () { numpad.pendingUp = null; numpadRelease(); }, 60) };
          return;
        }
        if (ev.code === "Numpad0" && numpad.zeroDown) {
          clearTimeout(numpad.zeroDown.timer);
          numpad.zeroDown = null;
          numpadExitTap("Numpad0");
          return;
        }
        var m = /^Numpad([0-9])$/.exec(ev.code);
        if (!m || !numpad.active || String(numpad.active.key).replace(/off$/, "") !== "KP" + m[1]) { return; }
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
      Promise.all([loadNameSpellings(), loadBrobotsSettings()]).then(function () {
        restoreChat();
        setInterval(pollFeed, 500);
        setInterval(pollRecordSession, 1000);
        setInterval(pollWarmup, 500);
        pollFeed();
        pollWarmup();
      });
    })();

