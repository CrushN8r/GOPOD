# Start and Wire-Pod
- **words:** goinit, gorepo, wpr, wpr 2, wps, wpq, wpu, start a session, restart wire-pod, stop wire-pod, start wire-pod, update wire-pod
- **ask:** What are the start and Wire-Pod aliases?
- **about:** the terminal's start and Wire-Pod aliases (~/.bash_aliases and ~/.gopod_alias_lib/brobots.sh); in the chat, system-restart does what wpr 2 does
---

## Tour
1. A new terminal? Load the aliases once: “source ~/.bashrc ~/.bash_aliases”.
2. Then “goinit” - the session's start in one go: the repo, the Brobots on balanced power, their frame rate shown, Wire-Pod restarted (it asks for the sudo password).
3. Wire-Pod acting up mid-session? “wpr 2” restarts it; “wpr 1” only restarts it if it isn't healthy.
4. Done for the day? “wpq” stops Wire-Pod; “wps” starts it again.

## Commandments
1. goinit, a session's start: gorepo, gopower balanced, gofps, wpr 2
2. gorepo, to the GOPOD repo folder
3. wpr, restart Wire-Pod: 1 (only if it isn't healthy) | 2 (forced, sudo password every time); sets the speakers and USB mic first
4. wps, start Wire-Pod (skips if it's already running)
5. wpq, stop Wire-Pod
6. wpu, update Wire-Pod from its own upstream (stop, update.sh, setup.sh daemon-enable, start) - it resets ~/wire-pod and builds plain Wire-Pod, replacing GOPOD's: run the deploy after it

## Next
1. The Brobots set for the show? Go to brobots-settings.

[ALIAS-PLAY-STUDIO](../../ALIAS-PLAY-STUDIO.md) · [Brobots settings](../brobots-settings/README.md)
