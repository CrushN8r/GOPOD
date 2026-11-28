"""Find a song's folder by its name (song_id), wherever it sits."""

import json
import sys
from pathlib import Path

SONGS_DIR = Path(__file__).resolve().parents[1]
NOT_SONGS = {"tools", "_new_song"}


def all_songs(songs_dir=SONGS_DIR):
    """{song_id: folder} for every song folder with a knobs.json."""
    found = {}
    for knobs in sorted(Path(songs_dir).glob("*/knobs.json")):
        folder = knobs.parent
        if folder.name in NOT_SONGS or folder.name.startswith((".", "_")):
            continue
        try:
            song_id = json.loads(knobs.read_text(encoding="utf-8")).get("song_id", "")
        except (OSError, ValueError):
            continue
        if not song_id:
            continue
        if song_id in found:
            raise ValueError(f"song_id {song_id!r} is in both {found[song_id].name} and {folder.name}")
        found[song_id] = folder
    return found


def find_song_dir(song_id, songs_dir=SONGS_DIR):
    """The folder of the song with this song_id, or None."""
    return all_songs(songs_dir).get(song_id)


if __name__ == "__main__":
    songs = all_songs()
    if len(sys.argv) > 1:
        folder = songs.get(sys.argv[1])
        if folder is None:
            print(f"no song with song_id {sys.argv[1]!r} - songs: {', '.join(sorted(songs))}", file=sys.stderr)
            sys.exit(1)
        print(folder)
    else:
        for song_id, folder in songs.items():
            print(f"{song_id}\t{folder.name}")
