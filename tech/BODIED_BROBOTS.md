# BODIED_BROBOTS

> Where the robot bodies live.
> Three sets of robots. One honest proof standard. No overclaiming which ones actually move yet.

The **brobots** scaffolding hosts robot *bodies*. This doc is the body half of that story — how each
physical robot connects, how far its control is actually proven, and what's still open. It is about
bodies only. Personas (Doc, Pip, the rest — the GOPOD layer) **link to a brobot profile** to borrow a
body, or stay bodyless. That link is the only place personas touch this page. Everything below is
hardware and honesty.

---

## The proof ladder

A body isn't "working" because it answered a ping. It's working when it *moves on command and can be
stopped.* Every robot on this page is rated against the same four rungs:

```
L0  DISCOVERED       the robot / its network can be found
L1  CONNECTED        a session or protocol handshake is established
L2  CONTROLLED       a command actually reaches the robot
L3  PHYSICAL PROOF   the robot physically responds — and can be stopped
```

The rule that keeps this page honest:

> **Connectivity is not control. Control is not physical proof. Physical proof is not recovery.**

Nothing below **L3** is called "fully operational." A link that passes L0–L2 but has never turned a
wheel is exactly that — a link, not a performance.

---

## The Robot Link Contract

Every body — Vector, Cozmo, Scout, and anything added later — answers the same nine questions. This is
the checklist a new robot has to pass before it's trusted on stage:

```
1. IDENTITY    which physical robot is this, by a stable label
2. NETWORK     how it joins the local fabric
3. TRANSPORT   the middleware / protocol underneath
4. HANDSHAKE   how a session is established
5. COMMAND     the control endpoint that reaches it
6. RESPONSE    proof it acted (telemetry / physical motion)
7. STOP        a deterministic hard-stop, independent of any LLM
8. RECOVERY    how it comes back after a drop, reboot, or restart
9. PROOF       saved evidence the whole chain worked, not "worked-ish"
```

The point of the contract is that "it connected once" and "it's a trustworthy body" are different
claims, and this page never lets them blur.

---

## The common layer (envisioned, not built)

The architectural payoff, once more than one robot type is real:

```
              RobotNode
           /      |       \
       Vector   Scout    Cozmo
      Wire-Pod   ROS      SDK
```

One shared orchestration layer on top; robot-specific transport underneath. A persona (or the GOPOD
layer) talks to the common layer, never to a robot's raw protocol. **No raw LLM-generated commands
reach any robot** — everything passes through an approved adapter. This layer is a direction, not a
delivered thing.

---

## 1. Vectors — Brobots 1 & 2 · **L3, live-proven**

The golden reference. Everything else on this page is aiming at what the Vectors already do.

- **Identity.** Two Anki Vectors. **Brobot 1 = ESN `0dd1b9e9`**, **Brobot 2 = ESN `0dd1d8bf`**. ESN is
  the device serial — the stable key the whole system uses to tell them apart.
- **Transport.** Wire-Pod on the Jetson. Each Vector is keyed by its ESN; Wire-Pod resolves that ESN to
  a Vector SDK connection.
- **Command.** Driven through the alias play studio — `phcal` (the calibration bench, one primitive at a
  time) and `pha0b` (the performance front door, whole songs with a playhead). Deterministic: the same
  song plays the same way every time.
- **Network.** Local-first, offline. Robots join a local 2.4 GHz fabric; no internet required after
  setup.
- **Proof.** Discovered, connected, controlled, physically responding, and stoppable — all confirmed,
  live, on real hardware. **L3.**

This is the one set that has cleared every rung. It's the standard the Cozmos and Scouts are measured
against, and the reference the future common layer is built to match.

*One honest boundary:* Wire-Pod's robot resolution builds a **Vector-specific** connection. It isn't a
key you can widen to a non-Vector robot — it's the connection type itself. That's why the other two sets
below don't route through Wire-Pod; they get their own adapter paths. (There's already a working
precedent in the repo for talking to hardware directly, outside Wire-Pod's API.)

---

## 2. Cozmos — **staged, below L3** · chat-first, body-later

Cozmo groundwork is mapped, but no live body control is proven. Honest status: **staged (SDK import and
connectivity approach mapped), live control NOT proven.**

- **Not a plain IP robot.** Cozmo's control path runs **phone/tablet ↔ Cozmo App / Engine ↔ robot** —
  there's an app/SDK dependency in the middle, not a direct socket. Any real integration has to account
  for that app/engine hop, not treat Cozmo as a bare network device.
- **The duo USB WiFi adapter gotcha (known-hard).** Two USB WiFi adapters were added for Cozmo comms.
  They are **OS-identical** — same chip (**Realtek RTL8723BU**), same **VID:PID `0bda:b720`**, same
  product string. `lsusb` alone *cannot tell them apart*, and the bus/device numbers it shows are not
  stable — they renumber across reboots and replugs. The only stable way to distinguish two identical
  adapters is the **physical USB port path** (walk `/sys`, capture each adapter's port path, then map
  port path → current interface). The commands to do this exist, but have **not yet been run or confirmed
  live** — and the port-path → which-Cozmo mapping is the step after that.
- **Rung status.** Connectivity approach understood; the physical-response and recovery rungs are open.
  **Below L3.**

Cozmo's intended near-term role is chat-first (a voice/persona participant) with the physical body
following later — which is why the not-yet-proven body control is a known, accepted gap rather than a
broken promise.

---

## 3. Moorebot Scouts — **staged, below L3** · likely-first next body

The most likely next body to bring up, because Scout is a *different platform* — which means building
fresh alongside Vector rather than bending Vector-only code to fit.

- **Its own platform.** Scout is ROS-based. The pieces that matter: **ROS master ownership**, a
  **`/cmd_vel` publisher/subscriber** pair for drive, and a **camera / RTSP path kept separate from
  drive control** (seeing and moving are two different lanes, proven independently).
- **Routes around Wire-Pod.** Because Wire-Pod's transport is Vector-specific, Scout doesn't go through
  it at all — it gets its own adapter path, following the same direct-to-hardware precedent that already
  exists in the repo. This is *additive*, not a refactor of the Vector path.
- **Known-open, kept separate: motion calibration.** Drive-straightness / right-drift is tracked as its
  own issue, deliberately not tangled with connectivity — a Scout can have a passing link and an open
  straightness problem at the same time. Keeping them separate stops "the link works" from being mistaken
  for "the robot drives straight."
- **Rung status.** Connectivity approach mapped; live control and physical proof open. **Below L3.**

Scout being a clean, separate platform is exactly why it's the gentler on-ramp: no fighting Wire-Pod's
Vector plumbing, and a chat-participant role is more forgiving than full-body performance.

---

## Why this page is thin on addresses

This is a public document. Network identifiers — IP addresses, MAC addresses, hostnames, interface
names, ROS master URIs, network names — live in private notes and never appear here. What *does* appear
is safe by nature: **ESNs** (Vector device serials, already public across the repo) and **USB chip IDs**
(the `RTL8723BU` / `0bda:b720` hardware identifiers). Everything on this page describes the *approach* —
how a body connects and gets proven — not the private addresses that would let someone reach into the
local fabric. When a detail couldn't be stated without a network identifier, it was stated generically
or left out.

---

## Current standing, one glance

| Robot set | Role | Proof |
|---|---|---|
| **Vectors** (Brobot 1 · Brobot 2) | Built, performing | **L3 — live-proven** |
| **Cozmos** | Chat-first, body-later | Staged — below L3 (duo-adapter ID open) |
| **Moorebot Scouts** | Likely-first next body | Staged — below L3 (straightness open) |

One set on stage. Two staged, honestly. The scaffolding that hosts them is real; the bodies that aren't
proven yet say so.

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](README.md) for everything else in this folder, or [the root map](../README.md) for the rest of GOPOD.
