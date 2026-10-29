# Bingo reactor

`bingo_reactor_001.py` watches the chat file the Bingo game writes to. Each time a number is
drawn, it plays an angry animation (`anim_rtpickup_loop_10`) on Brobot 2 (`ESN02`), and it
stops itself about 2 seconds after the game ends.

- **Runs with the game:** `gobingo` starts it in the background and stops it when Bingo exits.
- **Hands off the game:** it only reads the chat file. If it isn't running, Bingo plays the
  same, just without Brobot 2's reaction.
- **Best effort:** a failed connect, animation or read is logged
  (`GOPOD_BINGO_REACTOR_*_FAIL`) and the watcher keeps going.
- **Another Brobot 2:** pass `--brobot2-serial <serial>`. The robot's address comes from the
  `anki_vector` SDK's `~/.anki_vector/sdk_config.ini`.
- **Needs:** Python 3 and the `anki_vector` package from Wire-Pod's
  [wirepod-vector-python-sdk](https://github.com/kercre123/wirepod-vector-python-sdk)
  (Apache-2.0), in a venv here (`requirements.txt`).
