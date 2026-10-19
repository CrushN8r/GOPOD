# DEMO / TEMPLATE for gopod_audio_config.sh. Copy this file to
# gopod_audio_config.sh, in the SAME directory as your own brobots.sh (e.g.
# ~/.gopod_alias_lib/gopod_audio_config.sh - never commit your real one), and
# fill in your own machine's real device names. The two values below are
# obviously-placeholder examples - replace them.
#
# To find your own real device names:
#   pactl list short sources     (mic - use the NAME column, 2nd column)
#   pactl list short sinks       (speakers - same, NAME column)
# Use the stable NAME string, not the leading numeric index - indexes drift
# on reboot/replug, names don't.
#
# Missing this file entirely is fine - _gopod_check_audio_routing() in
# brobots.sh falls back to its own hardcoded default (whatever machine that
# copy of brobots.sh was originally built on) if gopod_audio_config.sh isn't
# present.

GOPOD_EXPECTED_AUDIO_SINK="alsa_output.your-speaker-device-name-here"
GOPOD_EXPECTED_AUDIO_SOURCE="alsa_input.your-mic-device-name-here"
