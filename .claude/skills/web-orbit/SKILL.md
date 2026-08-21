---
name: web-orbit
description: Use when a task touches GOPOD's web orbit — the CRUSHN8R domain network, hosting, newsletter, shop, social accounts, or YouTube — or when any task risks touching a credential, SSH key, or secure-local material. States the security boundary (nothing from the local secure lane ever enters a session's output, ever), carries the public-safe account roster and YouTube channel/playlist structure, and points at where deeper web-orbit detail lives (gopod_notes, outside this repo) for anything not already summarized here. Promoted 2026-08-13 from boundary-only to boundary-plus-procedures.
---

# Web orbit

## A. Identity

GOPOD's web orbit — the CRUSHN8R domain network, hosting, newsletter, shop — is
**support infrastructure**, not the project. Operator doctrine, verbatim: "Websites
should document and route GOPOD proof, not become the project." Per the campaign's
launch conditions (see `niche-buzz` §8), web-orbit build-out sits downstream of
ignition — sessions do not start website work unprompted.

## B. The hard fence — security boundary, absolute

- A local secure lane exists outside this repo holding credentials, SSH keys, and
  launch-prep material. Sessions NEVER read, list, copy, quote, summarize, commit, or
  stage anything from it — not into a chat, a prompt, a report, or a file.
- SSH keys, passwords, and tokens never appear in any session output, ever, even
  partially, even redacted, even "just the filename list."
- If a task appears to require a credential, STOP and hand the step to the operator —
  the operator runs credentialed actions themselves.
- Anything under a path containing "secure" is treated as fenced by default.

## C. Where web truth lives (pointers, not content)

- The web-orbit knowledge base — hosting, domains, baseline, automation notes — lives
  in `gopod_notes` (outside the public repo).
- The campaign-level web map (domain roles, the water flow) lives in the `niche-buzz`
  skill.

Sessions needing web context read those two. This skill deliberately holds none of
that detail itself — `.claude/skills/` is public, tracked, MIT-licensed, and headed for
a public launch, so no domain-by-domain list, hostname, port, username, key filename,
or secure-folder path belongs here, ever.

## D. Account roster

Public-safe only — domain names and platform roles, no handles, no logins, no hosting
detail. Pulled from `gopod_notes/older_notes/GOPOD_SOCIAL_MEDIA_WEBSITES_MAP_001.md`
(domain roles) and the social-account starter-registry note (platform roles) — do not
invent an account not already named on disk. As of this promotion, no specific social
account handle is confirmed anywhere in this repo's notes; the registry itself says so
("Operator confirmation needed: account existence, handle, login readiness..."). This
table names *roles*, not live accounts.

**Ownership role-split:** `@CrushN8r` = the GOPOD universe (creative/lore/product/
community); `@Flare4com` = all non-GOPOD ops (web/infra/tech/business) — clean role
wall, `@Flare4com` never overrides GOPOD.

| Item | Type | Role | Audience |
|---|---|---|---|
| `crushn8r.com` (communications) | Domain | Main incoming-traffic redirector / comms hub — UTM, affiliate codes, redirected out to affiliate product/location. Routes to GOPOD, AccessMath, Math Aftermath, shop, contact, newsletter. | General public entry point. |
| `crushn8r.ca` (contact) | Domain | Trust layer — physical location, contact form + newsletter (Listmonk) capture. Home of the CRUSHN8R CREW'd Newsletter. | Regional visitors, credibility surface. |
| `crushn8r.net` (network) | Domain | The linktree — link-tree "YOU ARE HERE" mall map; parent of the niche-pillar subdomains. | General public, first landing from the wake-phrase funnel. |
| `shop.crushn8r.com` | Domain | **LIVE, WIP** — confirmed live 2026-08-15: bare WordPress/WooCommerce default scaffold, sample pages, no real products yet; not live commerce. Intended purpose (operator-confirmed, not yet built): (1) backend persona-management via WooCommerce product management (products = persona/object handles, not retail goods), (2) landing point for `crushn8r.com` redirects → instant affiliate purchases, (3) can be set up for WDTM merch routing. | Commerce (future). |
| `mathaftermath.crushn8r.net` | Domain | GOPOD's own News Channel — recaps/clips/quote cards. | Public-safe GOPOD followers. |
| 8 niche-pillar subdomains (accessmath/foodmath/moneymath/sportsmath/fashionmath/languagemath/personalitymath/survivalmath, all `.crushn8r.net`) | Domain | Everyday-math content pillars, the niche pillar micro-sites, one lane each — see the websites map for per-pillar detail. | Pillar-specific readers (educators, cooks, budget-minded, etc). |
| YouTube | Social | Video sparks, longer proof cuts, Shorts; descriptions route to a real blog/pillar root. Not for pre-proof launch/shop/newsletter claims. | Viewers following video proof. |
| TikTok | Social | Fast clips, hooks, character moments, trend-fit posts — routes back to the right live root, never the authority source itself. | Short-form/trend audience. |
| Instagram | Social | Reels, carousels, quote cards, CHALK visual cards, mobile proof. No product tags/shop CTAs before shop proof exists. | Visual/mobile audience. |
| Facebook | Social | Local/trust updates, longer captions, video shares, community sharing. No contact/booking/shop claims without proof. | Community/local audience. |
| X / Twitter | Social | Short sparks, threads, status notes, proof snippets — not a live-status broadcaster. | Fast-moving/status-following audience. |
| LinkedIn | Social | Professional proof, accessibility/session framing, build updates, authority summaries. No overclaiming availability. | Professional/institutional audience. |
| Pinterest | Social | Durable visual search — FashionMath/FoodMath/SurvivalMath, visual explainers, worksheet-style pins. No fake products/pages. | Visual-search audience. |
| Reddit | Social | Careful community-fit discussion and listening only, not broad promotion — planning-state discipline. | Community/niche-forum audience. |
| Email newsletter (CRUSHN8R CREW'd, off `crushn8r.ca`) | Owned rail (not social) | Weekly updates, segment digests, video-spark digests, shop/product updates — only after capture/list proof exists. | Opted-in subscriber list. |

**Website intro order** (operator-confirmed): the CRUSHN8R brand first, then the three
root domains in role order — `crushn8r.ca` (contact/trust) → `crushn8r.com`
(communications/redirector) → `crushn8r.net` (network/linktree) — then the niche-pillar
subdomains.

`gopod_notes/older_notes/websites_gomad/30_automation/social_routing.md` (cross-site
platform routing/back-link map) is status `PLANNED, NOT WIRED` — nothing here is
automated yet; every posting decision is still a manual, proof-gated call.

**Live-status discipline (read before citing any status)**: the operator has confirmed
the sites are live for PLANNING purposes this session — that confirmation is real and
usable for planning. It is **not** the same as the per-domain disk record. The
`websites_gomad` tree's own `manifest.md` status fields (hosting, SSL, WordPress
install, etc.) stay `UNKNOWN` on disk by design until the operator's own direct
observation fills them in (per that GOMAD's own "UNKNOWN-by-default status discipline").
State it this way, every time: *operator-confirmed for planning; per-domain technical
status still tracked in the GOMAD, filled only by the operator's own eyes.* Never write
"all sites live" as a disk fact.

## E. YouTube channel

**Channel identity** — confirmed, verbatim, do not reword:

> CrushN8r — "Crazy Off-Course Crash-Courses!"
> Two robots, Doc and Pip, explaining the math behind the wordplay. Songs,
> interviews, and live sessions from the workbench. When you're ready, just
> GOPOD Yourself!!
> 🔗 crushn8r.com

The channel is near-empty today (a handful of subscribers, one video) — everything
below describes launch-forward **structure**, not existing content. No overclaiming
what's already posted.

**Playlist plan** — maps the niche-buzz funnel (`niche-buzz` §3: BAIT → MAIN ACT → NET
video → UPSELLS) onto upload/playlist order:

1. **BAIT hooks** — the short intro flash(es) that open the funnel (`00_brobots_awaken`,
   ~90-second news-flash shape). First thing a new viewer should hit.
2. **MAIN ACT / NET videos (two)** — the flagship, split 2026-08-19 into a 2-video
   playlist (`gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`, operator's own framing call,
   `GOPOLISHER_FIXES_001.md`): `01_brobots_interview_vamp` (video 1, the pre-show
   banter) uploads first, `02_brobots_interview_run` (video 2, the scripted Brobot
   1/Brobot 2 interview, seven exchanges) follows. Together the pair that gets
   cross-posted to spark the wider push.
3. **UPSELLS, Bingo first** — `101_brobots_bingo_test` (the scripted upsell capture song)
   leads the upsell playlist; `105_brobots_nap` (formerly `104_brobots_baby_robots_sleep`,
   "Do Baby Robots Dream?") follows, held for later per standing operator direction
   (`niche-buzz` §4). The live
   Bingo *game* itself (`102_brobots_bingo_game`, "Chocolate Bingo") is a different piece
   — see `tech/alias_play_studio/SONG_102_BROBOTS_1_2_BINGO_GAME.md` for the live game's own
   doc (and `SONG_101_BROBOTS_1_2_BINGO.md` for the scripted song's);
   don't conflate the scripted upsell video with the live game when populating this
   playlist.

**Song-video → channel mapping**:

| Song | Funnel role | Playlist slot |
|---|---|---|
| `00_brobots_awaken` | BAIT | Opening hook(s) |
| `01_brobots_interview_vamp` | NET video 1 / MAIN ACT | Flagship video 1, pre-show, cross-posted |
| `02_brobots_interview_run` | NET video 2 / MAIN ACT | Flagship video 2, seven exchanges, cross-posted |
| `101_brobots_bingo_test` | UPSELL 1 | Upsells, first |
| `105_brobots_nap` (was `104_brobots_baby_robots_sleep`) | UPSELL 3 | Upsells, held for later |

(`102_brobots_cross_persona`, formerly UPSELL 2, is archived per `niche-buzz` §4 —
superseded once the live `103_gopod_is_that_you` demo proved the same bit for real; not a
current playlist candidate.)

**Guardrail, kept intact**: YouTube is an outlet, not the strategy owner. Social
platforms receive assets *produced* by the session — never the other way around. No
YouTube-first treadmill, no requirement that GOPOD become a full-time media company;
reusable-asset flow beats brute-force content grind.

## F. Posting rules

- **Proof before any launch/shop/newsletter claim.** No platform post claims a site,
  shop path, or newsletter signup is live before that specific thing has actually been
  proven — matches every per-platform rule in §D's roster (e.g. "no shop CTAs before
  shop proof exists," "no contact/booking claims without proof") and the GOMAD's own
  UNKNOWN-by-default discipline.
- **The two-clocks rule.** Web-orbit work never mutates the runtime (robot/song) lane,
  and never waits on it either — they run on separate clocks. The web/social lane only
  needs one proof loop captured to move forward; it does not block on the robots being
  "finished," and robot/song work is never blocked waiting on web-orbit either.
- **Evidence-and-boundary review before social distribution.** Clips/captions/quote
  cards/recaps reach social platforms only after the evidence and public/private
  boundary review — same review discipline this repo already applies to public-facing
  material generally.

## Scope

- Boundary is still absolute (§B) — this promotion adds public-safe procedures, it does
  not loosen the hard fence in any way. No credentials, no hostnames, no ports, no
  usernames, no secure-lane paths, ever.
- §D/§E/§F hold only what's already confirmed on disk or given directly by the operator
  (the YouTube channel identity block) — no invented accounts, no invented handles, no
  invented live-status claims. Deeper domain-by-domain technical detail (hosting, SSL,
  DNS) still lives only in `gopod_notes` (§C), not duplicated here.
- Public-safe by construction — every addition to this file gets the same test before it
  lands: would this be fine for a Wire-Pod visitor to read on launch day?
