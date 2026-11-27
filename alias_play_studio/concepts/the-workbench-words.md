# The workbench words
- **words:** golden, wip, lock it in, commit, push, desk start, lsusb, the rig
- **ask:** What do golden, WIP and the desk start mean?
---

| Say | It means |
|---|---|
| **Golden** | Proven live and locked. Changed only on the owner's go. |
| **WIP** | Being worked on. |
| **Lock it in** | Mark it golden and commit it. |
| **Commit / push** | Save to the repo / send it to GitHub. Two separate go's. |
| **F5** | Reload the chat page (after `wpr 2`). |
| **Desk start** | `source ~/.bashrc; gorepo; lsusb; wpr 2`, then F5 (below). |
| **Aliases** | Every terminal alias has one card, its one source: [Start and Wire-Pod](../start-and-wire-pod/README.md) · [Brobots settings](../brobots-settings/README.md) · [Songs](../songs/README.md) · [Air hockey and the stage](../air-hockey/README.md). |

## Desk start

Your terminal routine: `source ~/.bashrc; gorepo; lsusb; wpr 2`, then F5.

- `source` reads only the first file it's given (a second becomes an argument). `~/.bashrc`
  already loads `~/.bash_aliases`, so `source ~/.bashrc` alone does both.
- `gorepo`: `cd ~/crushn8r_git/GOPOD`.
- `lsusb` - the rig's parts, by name (bus/device numbers drift, names don't): **TTGK
  Technology USB Audio** (3302:12e9, the USB mic), **China Resource Semico ... USB Keyboard**
  (1a2c:4324, the numpad), **Novatel Wireless M3000** (the mobile hotspot), Realtek Bluetooth
  Radio and 802.11ac NIC (Wi-Fi), the VIA / Huasheng USB hubs. Mic or numpad missing here = the
  chat's mic or numpad won't work.
- `wpr 2`: sets the audio routing (`AUDIO_ROUTING_SET speakers=Built-in Audio Analog Stereo
  mic=USB Audio Mono`), then restarts Wire-Pod (asks for sudo; `WPR_RESULT status=FORCED`,
  "wire-pod restarted <time>"). A deploy made after that time needs another `wpr 2`.
- F5 on the chat page: the page clears the chat for the new start and reloads once by itself.
- Then click **Log** on Wire-Pod's web page. Nothing plays until you do (a browser plays sound
  only after a click).

[Concepts](README.md) · [ALIAS-PLAY-STUDIO](../../ALIAS-PLAY-STUDIO.md)
