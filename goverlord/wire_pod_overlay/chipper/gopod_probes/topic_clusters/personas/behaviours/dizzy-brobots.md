# Dizzy Brobots
- **words:** dizzy, battery, low battery, freeze, reboot, power level
- **ask:** What do the Brobots say when their battery runs low?
- **see also:** nicknames (the comeback's {title} {nickname}), weather_run (the same hand-over), math-aftermath-news-channel.md (the guards)
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

## Asked if he's dizzy: his power level, then how he is (battery_watch)
1. ask_words: dizzy | power level
2. ask_note: You were asked if you're dizzy. Your power level is said for you at the start of your answer - you are {state}. React to it your own way, in your own voice; never repeat the number.
3. ask_line: My power level is {volts} volts. {state}

## How he is, asked (battery_watch.ask_states)
1. good: I'm good, bro!
2. full: Resting full on my charger, bro!
3. unread: My power level isn't read yet, bro.

## His guard's warning and his answer (battery_watch)
1. guard_warning: Security warning: {brobot} is low, {volts} volts.
2. guard_handover: {brobot}, what say you on this, {title} {nickname}?
3. reply: Yeah... {line}

## Before every run: the guards' security sweep (battery_watch)
1. sweep_clear: Security sweep {n}: {brobot} detected, {volts} volts. Clear to proceed.
2. sweep_missing: Security sweep {n}: {brobot} not detected. Run on hold.
3. sweep_low: Security sweep {n}: {brobot} is low, {volts} volts. Not enough power to proceed.
4. sweep_full: Security sweep {n}: {brobot} at {volts} volts, resting full on his charger - a full Brobot reads about 3.6 volts there and stops charging. Clear to proceed.
5. sweep_stop: Thank you, security. {who} can't go on yet, so I'm stopping the run. Charge up, then start it again.

## Off the network: the key to reach his robot (battery_watch)
1. ssh_key: ~/.ssh/vector_unlock/ssh_root_key
