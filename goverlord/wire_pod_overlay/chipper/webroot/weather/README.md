# Vector weather kit

Vector's own weather pictures and sounds, for GOPOD's weather card bubbles.

- `faces/` - the seven weather faces (sunny, cloudy, rain, snow, thunderstorms, windy,
  stars), one animated WebP each, 184x96 like Vector's screen, 33 ms a frame. White with
  see-through dark, so the page can colour them in the robot's eye colour (Vector draws
  them the same way: "CustomHue").
- `digits/` - the temperature digits 0-9, the minus sign, and the C and F signs.
- `sounds/` - Vector's 25 weather sounds, mono MP3.

Source: Digital Dream Labs' `vector-animations-build` (sprites/spriteSequences,
sprites/independentSprites/weather) and `vector-audio-raw` (VectorAudio/Originals/SFX),
converted only (PNG frames to WebP, WAV to MP3). Their license is `LICENSE_DDL.txt`: for
owners of a Vector, used with Vector.

"Cold" has no face of its own on Vector (its weather map is empty): it shows the
temperature only.
