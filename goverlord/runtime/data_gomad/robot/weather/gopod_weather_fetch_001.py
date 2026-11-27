#!/usr/bin/env python3
"""GOPOD's own independent weather fetch for Windsor, Ontario.

Isolated on purpose - no import of, or routing through, Wire-Pod's own
pkg/wirepod/ttr/weather.go. Reads apiConfig.json live at call time (never
cached, never mirrored into this repo - read-only Wire-Pod Go-runtime
property, same rule as always). Makes its own two-step OpenWeatherMap call
(geocode -> current weather by lat/lon), the same shape weather.go itself
uses, written independently.

Fetch once, in one neutral unit (metric/Celsius). Per-robot unit and clock
format are read LIVE from that robot's own Wire-Pod jdocs RobotSettings
(jdocs.json, keyed by serial - see load_robot_format_from_jdocs() below),
never hardcoded here - format_for_robot() takes them as plain arguments.
Changed 2026-08-13 (WEATHER_LIVE_JDOCS_SOURCED_001.md): this used to read a
hand-authored static file (robot_weather_format.json) that happened to
match each robot's real settings - a silent-drift seam, since nothing kept
the two in sync. One source of truth now: the robot's own jdocs.

Fails loudly - no silent fallback, no cached stale answer - if the key is
missing, the provider isn't the one this call shape was written for, the
robot's own RobotSettings jdoc can't be found, or the API call itself
fails.
"""
import json
import os
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

API_CONFIG_PATH = Path(
    os.getenv("GOPOD_WIREPOD_CHIPPER_ROOT", "/home/goverlord/wire-pod/chipper")
) / "apiConfig.json"
JDOCS_PATH = Path(
    os.getenv("GOPOD_WIREPOD_CHIPPER_ROOT", "/home/goverlord/wire-pod/chipper")
) / "jdocs" / "jdocs.json"
LOCATION = "Windsor,ON,CA"  # weather.go's own documented geocode format: city,state code,country code
EXPECTED_PROVIDER = "openweathermap.org"


def _read_weather_config():
    with API_CONFIG_PATH.open("r", encoding="utf-8") as handle:
        config = json.load(handle)
    weather = config.get("weather", {})
    provider = weather.get("provider", "")
    key = weather.get("key", "")
    if provider != EXPECTED_PROVIDER:
        raise RuntimeError(
            f"BLOCKED: weather provider in {API_CONFIG_PATH} is {provider!r}, "
            f"expected {EXPECTED_PROVIDER!r} - this call shape was written for "
            "openweathermap.org only, refusing to guess a different one"
        )
    if not key:
        raise RuntimeError(f"BLOCKED: no weather API key found in {API_CONFIG_PATH}")
    return provider, key


def _http_get_json(url, timeout=10):
    request = urllib.request.Request(url, headers={"User-Agent": "gopod-weather-fetch/001"})
    with urllib.request.urlopen(request, timeout=timeout) as response:
        body = response.read().decode("utf-8", errors="replace")
    return json.loads(body), body


def fetch_windsor_weather(location=None):
    """location=None (every existing caller) preserves the exact original
    behavior - the module's own LOCATION constant, "Windsor,ON,CA" - the
    optional param (2026-08-15, PHCAL_NOTES_WEATHER_WAKE_FIXES_001.md) only
    lets phcal's own bench weather test target somewhere else on request.
    Single direct geocode query, same as always - no candidate-fallback
    retry logic (that lives in phcal's own now-unreachable
    cmd_weather_test/_weather_geocode_candidates, not ported here to avoid
    reimplementing this function). Name kept as-is (not renamed to a
    generic fetch_weather()) to avoid rippling through every import site
    for a phcal-scoped addition."""
    provider, key = _read_weather_config()
    query_location = location if location else LOCATION

    geo_url = (
        "http://api.openweathermap.org/geo/1.0/direct?q="
        + urllib.parse.quote(query_location)
        + "&limit=1&appid="
        + key
    )
    geo_data, geo_raw = _http_get_json(geo_url)
    if not geo_data:
        raise RuntimeError(
            f"BLOCKED: geocoding returned no results for {query_location!r}. Raw response: {geo_raw}"
        )
    lat = geo_data[0]["lat"]
    lon = geo_data[0]["lon"]

    weather_url = (
        "https://api.openweathermap.org/data/2.5/weather"
        f"?lat={lat}&lon={lon}&units=metric&appid={key}"
    )
    weather_data, weather_raw = _http_get_json(weather_url)
    if "weather" not in weather_data or "main" not in weather_data:
        raise RuntimeError(f"BLOCKED: unexpected weather response shape. Raw response: {weather_raw}")

    return {
        "provider": provider,
        "queried_location": query_location,
        "resolved_name": geo_data[0].get("name", ""),
        "resolved_state": geo_data[0].get("state", ""),
        "resolved_country": geo_data[0].get("country", ""),
        "lat": lat,
        "lon": lon,
        "condition_main": weather_data["weather"][0]["main"],
        "condition_description": weather_data["weather"][0]["description"],
        "temp_c": weather_data["main"]["temp"],
        "feels_like_c": weather_data["main"]["feels_like"],
        "humidity": weather_data["main"]["humidity"],
        "wind_speed_mps": weather_data.get("wind", {}).get("speed"),
        "dt_utc": weather_data["dt"],
        "timezone_offset_seconds": weather_data.get("timezone", 0),
    }


_DIGIT_WORDS = ["zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"]
_TEEN_WORDS = {
    10: "ten", 11: "eleven", 12: "twelve", 13: "thirteen", 14: "fourteen",
    15: "fifteen", 16: "sixteen", 17: "seventeen", 18: "eighteen", 19: "nineteen",
}
_TENS_WORDS = {2: "twenty", 3: "thirty", 4: "forty", 5: "fifty"}


def _two_digit_word(n):
    if n < 20:
        return _TEEN_WORDS[n]
    tens_digit, ones_digit = divmod(n, 10)
    word = _TENS_WORDS[tens_digit]
    if ones_digit:
        word += "-" + _DIGIT_WORDS[ones_digit]
    return word


def _military_hour_word(hour):
    if hour == 0:
        return "zero"
    if hour < 10:
        return f"zero {_DIGIT_WORDS[hour]}"
    return _two_digit_word(hour)


def _military_minute_word(minute):
    if minute < 10:
        return f"zero {_DIGIT_WORDS[minute]}"
    return _two_digit_word(minute)


def _military_time_spoken(hour, minute):
    if minute == 0:
        return f"{_military_hour_word(hour)} hundred"
    return f"{_military_hour_word(hour)} {_military_minute_word(minute)}"


def _twelve_hour_word(hour):
    if hour < 10:
        return _DIGIT_WORDS[hour]
    return _TEEN_WORDS[hour]


def _twelve_hour_time_spoken(hour_24, minute):
    hour = hour_24 % 12 or 12
    hour_word = _twelve_hour_word(hour)
    if minute == 0:
        minute_word = "o'clock"
    elif minute < 10:
        minute_word = f"oh {_DIGIT_WORDS[minute]}"
    else:
        minute_word = _two_digit_word(minute)
    suffix = "'a' m" if hour_24 < 12 else "pm"
    return f"{hour_word} {minute_word} {suffix}"


def format_for_robot(facts, unit, clock, include_location_date=False):
    c_value = round(facts["temp_c"])
    f_value = round(facts["temp_c"] * 9 / 5 + 32)
    c_symbol, c_words = f"{c_value}°C", f"{c_value} degrees Celsius"
    f_symbol, f_words = f"{f_value}°F", f"{f_value} degrees Fahrenheit"
    if unit == "C":
        primary_symbol, primary_words = c_symbol, c_words
        secondary_symbol, secondary_words = f_symbol, f_words
    elif unit == "F":
        primary_symbol, primary_words = f_symbol, f_words
        secondary_symbol, secondary_words = c_symbol, c_words
    else:
        raise RuntimeError(f"BLOCKED: unknown unit {unit!r}, must be 'C' or 'F'")

    local_dt = datetime.fromtimestamp(
        facts["dt_utc"] + facts["timezone_offset_seconds"], tz=timezone.utc
    )
    military_digit = local_dt.strftime("%H:%M")
    military_spoken = _military_time_spoken(local_dt.hour, local_dt.minute)
    twelve_hour = local_dt.hour % 12 or 12
    twelve_suffix = "am" if local_dt.hour < 12 else "pm"
    twelve_digit = f"{twelve_hour}:{local_dt.minute:02d} {twelve_suffix}"
    twelve_spoken = _twelve_hour_time_spoken(local_dt.hour, local_dt.minute)
    if clock == "24":
        time_str = military_digit
        primary_time_digit, primary_time_spoken = military_digit, military_spoken
        secondary_time_digit, secondary_time_spoken = twelve_digit, twelve_spoken
    elif clock == "12":
        time_str = twelve_digit
        primary_time_digit, primary_time_spoken = twelve_digit, twelve_spoken
        secondary_time_digit, secondary_time_spoken = military_digit, military_spoken
    else:
        raise RuntimeError(f"BLOCKED: unknown clock {clock!r}, must be '24' or '12'")

    if include_location_date:
        location_words = facts.get("resolved_name", "")
        state_words = facts.get("resolved_state", "")
        if state_words:
            location_words = f"{location_words} {state_words}"
        date_words = local_dt.strftime("%B %-d")
        year_words = local_dt.strftime("%Y")
        spoken = (
            f"{location_words}, {date_words}, {year_words}, "
            f"{primary_words} or {secondary_words}, and {facts['condition_description']}, "
            f"as of {primary_time_spoken} or {secondary_time_spoken}."
        )
        display_text = (
            f"{location_words}, {date_words}, {year_words}, "
            f"{primary_symbol} or {secondary_symbol}, and {facts['condition_description']}, "
            f"as of {primary_time_digit} or {secondary_time_digit}"
        )
    else:
        spoken = (
            f"{primary_words} or {secondary_words}, and {facts['condition_description']}, "
            f"as of {primary_time_spoken} or {secondary_time_spoken}."
        )
        display_text = (
            f"{primary_symbol} or {secondary_symbol}, and {facts['condition_description']}, "
            f"as of {primary_time_digit} or {secondary_time_digit}"
        )

    return {"spoken": spoken, "display_text": display_text, "local_time": time_str}


def gopod_robot_serial(label):
    """A robot's label (ESN01, ESN02) as its serial, for a direct robot connection:
    who it is this Wire-Pod start (gopod_brobots_now.json), else the operator's
    private ~/.gopod_esns.json; anything else as it is (2026-10-29)."""
    key = {"ESN01": "brobot_1", "ESN02": "brobot_2"}.get(str(label).upper())
    if not key:
        return label
    now = "/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_brobots_now.json"
    for path, k in ((now, key), (os.path.expanduser("~/.gopod_esns.json"), str(label).upper())):
        try:
            with open(path) as f:
                value = json.load(f).get(k)
            if value and value != "no-brobot-2":
                return value
        except (OSError, ValueError):
            pass
    return label


def load_robot_format_from_jdocs(serial, jdocs_path=None):
    """Read the target robot's LIVE vic.RobotSettings jdoc (jdocs.json,
    read-only Wire-Pod Go-runtime property, never mirrored - same rule as
    apiConfig.json above) and derive unit/clock ordering from it:
    temp_is_fahrenheit -> lead unit, clock_24_hour -> lead clock. One
    source of truth: the robot's own real settings. Fails loudly - no
    silent fallback - if this robot's own RobotSettings entry is missing or
    unreadable, matching this module's existing no-silent-fallback rule."""
    jdocs_path = Path(jdocs_path) if jdocs_path else JDOCS_PATH
    with jdocs_path.open("r", encoding="utf-8") as handle:
        entries = json.load(handle)
    thing = f"vic:{serial}"
    for entry in entries:
        if entry.get("thing") == thing and entry.get("name") == "vic.RobotSettings":
            settings = json.loads(entry["jdoc"]["json_doc"])
            unit = "F" if settings.get("temp_is_fahrenheit") else "C"
            clock = "24" if settings.get("clock_24_hour") else "12"
            return {"unit": unit, "clock": clock}
    raise RuntimeError(
        f"BLOCKED: no vic.RobotSettings jdoc found for {thing!r} in {jdocs_path} - "
        "cannot derive live per-robot weather unit/clock ordering"
    )


def main():
    facts = fetch_windsor_weather()
    print("RAW FACTS:")
    print(json.dumps(facts, indent=2))
    demo_serials = {
        "brobot_1": gopod_robot_serial(os.getenv("GOPOD_BROBOT_1_SERIAL", "ESN01")),
        "brobot_2": gopod_robot_serial(os.getenv("GOPOD_BROBOT_2_SERIAL", "ESN02")),
    }
    for robot_name, serial in demo_serials.items():
        fmt = load_robot_format_from_jdocs(serial)
        formatted = format_for_robot(facts, fmt["unit"], fmt["clock"])
        print(f"\n{robot_name} ({serial}, {fmt['unit']}, {fmt['clock']}h):")
        print(json.dumps(formatted, indent=2))


if __name__ == "__main__":
    main()
