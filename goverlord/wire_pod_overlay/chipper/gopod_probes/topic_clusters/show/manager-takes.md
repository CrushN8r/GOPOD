# Manager takes
- **words:** takes, hand-over, weather, time and place, usb, qr
- **ask:** Which manager takes which job?
- **chain:** Gomads and interns - a take hands a question to a manager or his intern (PLAYHEAD to Cache PYC for the weather)
- **see also:** ../personas/playhead.md, ../personas/cache-pyc.md, ../personas/chalk.md, ../personas/git-repo.md, weather_run
- **about:** 2026-11-03 (step 7): 02_cast.json managers.takes, a section each, in order (the code tries them in this order). why is the take's own note.
---

## Take 1: weather (CACHE PYC) (managers.takes.0)
1. host: {CACHE PYC.call}
2. via: {PLAYHEAD.call}
3. from: {GIT REPO.call}
4. kind: weather
5. words: weather |
6. take_text: ``
7. brobot_says: true
8. why: 2026-10-27 (operator: "what's the weather like right now * [Gomad] * [assigned-intern] will help you with that. * Thanks for your help, [intern]."): PLAYHEAD hands it to Cache PYC (was PLAYHEAD's own "I'll help you with that."); the Brobot thanks her and says the weather.

## Take 2: url_qr (PLAYHEAD) (managers.takes.1)
1. host: {PLAYHEAD.call}
2. kind: url_qr
3. take_text: I can help with that. QR code generated.
4. words: []

## Take 3: usb (GIT REPO) (managers.takes.2)
1. host: {GIT REPO.call}
2. via: {PLAYHEAD.call}
3. kind: usb
4. words: usb devices | what's plugged in | what is plugged in | lsusb
5. take_text: I'll check that.

## Take 4: time_place (PLAYHEAD) (managers.takes.3)
1. host: {PLAYHEAD.call}
2. kind: time_place
3. words: what time | time is it | where are we | wawn | what day is it | what's the date | what is the date | today's date
4. take_text: I'll help you with that.

## Managers (managers)
1. qr_display_text: I can help with that. QR code displayed.
2. qr_note: A QR code for {url} is shown in your bubble - you can tell them to scan the QR code.
3. time_place_text: {place}, {weekday}, {date}, {clock}.
4. usb_text: I see {count} USB devices. The mic is {mic}, and the numpad is {numpad}.
5. who_words: who are you | who r u | who is this | your name
6. wip_text: Bits you're still working on - try one only when the moment fits, never in place of your job: {bits}.
