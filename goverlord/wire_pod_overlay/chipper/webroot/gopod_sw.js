// goverlord/wire_pod_overlay/chipper/webroot/gopod_sw.js
// GOPOD as a phone app (2026-10-24, C3): the small background script a phone
// needs before it offers "Install app". It keeps nothing - every page, voice
// and picture still comes live from GOPOD, so the app never shows an old show.
self.addEventListener("install", function () { self.skipWaiting(); });
self.addEventListener("activate", function (event) { event.waitUntil(self.clients.claim()); });
self.addEventListener("fetch", function (event) {
  event.respondWith(fetch(event.request));
});
