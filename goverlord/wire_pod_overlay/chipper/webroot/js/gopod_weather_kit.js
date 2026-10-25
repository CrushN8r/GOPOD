// goverlord/wire_pod_overlay/chipper/webroot/js/gopod_weather_kit.js - the Weather Kit page's script (moved out of the page 2026-10-25).
// One card per weather. temp_c is a sample; the live card takes the real one.
var KIT = [
  { key: "sunny", words: "Sunny", face: "sunny", sound: "weather_sunny", temp_c: 24 },
  { key: "cloudy", words: "Cloudy", face: "cloudy", sound: "weather_cloudy", temp_c: 15 },
  { key: "rain", words: "Rain", face: "rain", sound: "weather_rain", temp_c: 11 },
  { key: "snow", words: "Snow", face: "snow", sound: "weather_snow", temp_c: -3 },
  { key: "thunderstorms", words: "Thunderstorms", face: "thunderstorms", sound: "weather_thunder", temp_c: 19 },
  { key: "windy", words: "Windy", face: "windy", sound: "weather_windy", temp_c: 9 },
  { key: "stars", words: "Clear night", face: "stars", sound: "weather_stars", temp_c: 6 },
  { key: "cold", words: "Cold and clear", face: "", sound: "weather_cold", temp_c: -18 }
];
var eyes = { vector: "#00e5ff" }, units = "C", place = "Windsor";
var list = document.getElementById("kitList");

function maskSpan(src, width) {
  var s = document.createElement("span");
  s.style.width = width + "px";
  s.style.webkitMaskImage = s.style.maskImage = "url(" + src + ")";
  return s;
}
// Vector's temperature layouts: digits 45 wide, the degree sign 29, minus first.
function tempScreen(t) {
  var screen = document.createElement("div");
  screen.className = "vector-screen";
  var row = document.createElement("div");
  row.className = "temp-row";
  if (t < 0) { row.appendChild(maskSpan("weather/digits/negative.png", 43)); }
  String(Math.abs(t)).split("").forEach(function (d) { row.appendChild(maskSpan("weather/digits/temp_" + d + ".png", 68)); });
  row.appendChild(maskSpan("weather/digits/" + (units === "C" ? "celsius" : "fahrenheit") + ".png", 43));
  screen.appendChild(row);
  return screen;
}
function faceScreen(face) {
  var screen = document.createElement("div");
  screen.className = "vector-screen";
  var tint = document.createElement("div");
  tint.className = "tint";
  tint.style.webkitMaskImage = tint.style.maskImage = "url(weather/faces/" + face + ".webp)";
  screen.appendChild(tint);
  return screen;
}
function card(item) {
  var bubble = document.createElement("div");
  bubble.className = "brobot-bubble operator system-status";
  var head = document.createElement("div");
  head.className = "brobot-bubble-head system-head";
  head.innerHTML = '<div class="brobot-bubble-avatar to-system"></div><span class="brobot-bubble-who">Goverlord: GOPOD Brain</span><div class="brobot-bubble-avatar to-system"></div>';
  bubble.appendChild(head);
  var c = document.createElement("div");
  c.className = "weather-card";
  var t = units === "C" ? item.temp_c : Math.round(item.temp_c * 9 / 5 + 32);
  if (item.face) { c.appendChild(faceScreen(item.face)); }
  c.appendChild(tempScreen(t));
  var words = document.createElement("div");
  words.className = "weather-words brobot-bubble-text";
  words.innerHTML = place + ": " + item.words + ", " + t + "°" + units + "<small>sound: weather/sounds/" + item.sound + ".mp3</small>";
  c.appendChild(words);
  var audio = new Audio("weather/sounds/" + item.sound + ".mp3");
  c.addEventListener("click", function () { audio.currentTime = 0; audio.play(); });
  bubble.appendChild(c);
  return bubble;
}
function render() {
  list.innerHTML = "";
  KIT.forEach(function (item) { list.appendChild(card(item)); });
}
function pick(eye) {
  document.documentElement.style.setProperty("--eye", eyes[eye] || eyes.vector);
  document.querySelectorAll("[data-eye]").forEach(function (b) { b.setAttribute("aria-pressed", b.dataset.eye === eye ? "true" : "false"); });
}
document.querySelectorAll("[data-eye]").forEach(function (b) { b.addEventListener("click", function () { pick(b.dataset.eye); }); });
document.querySelectorAll("[data-units]").forEach(function (b) {
  b.addEventListener("click", function () {
    units = b.dataset.units;
    document.querySelectorAll("[data-units]").forEach(function (o) { o.setAttribute("aria-pressed", o === b ? "true" : "false"); });
    render();
  });
});
fetch("gopod_brobot_eye_colors.json", { cache: "no-store" }).then(function (r) { return r.json(); }).then(function (e) {
  eyes.brobot_1 = e.brobot_1; eyes.brobot_2 = e.brobot_2;
}).catch(function () {});
document.querySelector('[data-units="C"]').setAttribute("aria-pressed", "true");
pick("vector");
render();
