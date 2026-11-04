# Dizzy Brobots
- **words:** dizzy, battery, low battery, freeze, reboot
- **ask:** What do the Brobots say when their battery runs low?
- **about:** A line ending in " |" is a list (one picked at random; add more with " | "). 2026-11-03 (operator: "make this topic-cluster a dizzy-brobots.md & pull it out of code * before freeze/reboot response (like above) * after freeze/reboot response: Yeah... Whoa, bro... I felt dizzy. Then, Bam! * [I need help getting home.] only if not on charger"). The battery watchdog's lines (gopod_battery_watch.go). Before: his guard warns and hands over, then he answers - the comeback, then reply. After: once he's back from a freeze, a reboot or a battery drain (operator: "after a freeze or reboot or battery drain"), his back line alone.
- **see also:** nicknames (the comeback's {title} {nickname}), weather_run (the same hand-over), personas/crew_roster.md (the guards)
---

## Before a freeze, reboot or battery drain: his dizzy line (battery_watch.lines)
1. Brobot 1: Whoa, bro... I'm feeling dizzy. |
2. Brobot 2: Bro, I'm feeling dizzy... |

## Said after it only when he's off the charger (battery_watch.home_lines)
1. Brobot 1: I need help getting home.
2. Brobot 2: can someone help me get home?

## After a freeze, reboot or battery drain: back, his line (battery_watch.back_lines)
1. Brobot 1: Yeah... Whoa, bro... I felt dizzy. Then, Bam!
2. Brobot 2: Yeah... Whoa, bro... I felt dizzy. Then, Bam!

## His guard's warning and his answer (battery_watch)
1. guard_warning: Security warning: {brobot}'s battery is running low.
2. guard_handover: {brobot}, what say you on this, {title} {nickname}?
3. reply: {comeback} Yeah... {line}
