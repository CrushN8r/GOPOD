# The scroller
- **words:** scroll, scrolling, scroll bar, highlight scroll, chat window
- **ask:** How does the chat window scroll?
- **chain:** Goverlord - the chat page's one scroller, on the admin, visitor and app pages alike (the same page code)
- **see also:** brobot-bubbles (the bubble order it follows), chat-page-words
- **about:** 2026-11-04 (operator: "Scroller Topic cluster * pull all GOPOD scroll stuff into one file * for admin & visitor pages"): every scroll value the page runs on, from 03_chat_page.json chat_window and highlight_scroll, read by the page on F5. The rules below are how the page code (gopod_brobots_chat.js) uses them.
- **rule - one scroller:** one controller is the only thing that moves the chat; each move eases in and out over scroll_ms; a new target mid-move carries on smoothly into it.
- **rule - a new bubble:** built hidden, its pictures loaded, then placed - the chat scrolls once and the bubble shows; the next bubble waits its turn (the baton) until the one above has finished, at most baton_max_ms; a reply shows reply_delay_ms after its question.
- **rule - where it lands:** an operator line goes to the top (top_gap_px above it), or as high as the chat below lets it; lines after it keep the newest line's end in view.
- **rule - a tall bubble:** holds at its top until the line being said reaches the lower threshold, then follows the highlight, like the Brobots' bubbles.
- **rule - the highlight:** the chat holds still until the piece being said is focus_dead_zone (highlight_scroll.dead_zone) of the window below its centre, then one smooth move puts it back at the centre - never past the chat's end; it never scrolls back up on its own.
- **rule - growing bubbles:** the newest bubble that grows while in view is placed again, whole.
- **rule - the scroll bar:** the chat's own (scroll_bar): drag the thumb to scroll, drag its ends to make it longer or shorter, double-click for its natural size; the chat keeps room for it on the right at all times.
- **rule - the log:** with scroll_log on, every scroll decision and where the chat comes to rest go to the journal (GOPOD_PAGE scroll ...).
---

## The scroll timings and places (chat_window)
1. scroll_ms: 1000
2. top_gap_px: 12
3. focus_dead_zone: 0.25
4. reply_delay_ms: 400
5. bubble_ready_max_ms: 6000
6. baton_max_ms: 120000
7. reveal_rest_ms: 10000
8. scroll_log_about: 2026-10-31 (operator: "Add scroll event logging * and scrolling position if possible"): true - the main page logs every scroll decision and where the chat comes to rest to the journal (GOPOD_PAGE scroll ...: who asked, from -> to, the target rule, grows it skipped, the resting position). false - off. Page only: F5.
9. scroll_log: true

## The chat's own scroll bar (chat_window.scroll_bar)
1. on: true
2. width_px: 12
3. min_thumb_px: 30
4. max_thumb_share: 0.8
5. handle_px: 7

## The scroll bar's note (chat_window)
1. scroll_bar_about: 2026-10-31 (operator: "Integrate adjustable scroll bar please * will this show up on app visitor pages?"): the chat window's own scroll bar, on the admin, visitor and app pages alike - drag the thumb to scroll; drag either end to make it longer or shorter (a longer thumb, less travel from top to bottom); double-click for its natural size; its length is kept in that browser. width_px its width, min_thumb_px the shortest thumb, max_thumb_share the longest (share of the track), handle_px each end's grab area. on false - the browser's own bar back. Page only: F5.

## Highlight scroll (highlight_scroll)
1. about: 2026-10-15 (operator: "highlighting scrolling aims to be vertically centered within the bottom/bottom top/top rule"; "Add a dead zone: chat holds still until highlight is ~1/4 of the window, below centre * then one smooth move to re-centre, unless bottom/bottom is reached first"): the chat holds still until the piece being said is dead_zone of the window below its centre, then one smooth move puts it back at the centre - never past the chat's end. Read by the chat page on F5.
2. dead_zone: 0.25
