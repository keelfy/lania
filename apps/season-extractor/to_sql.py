#!/usr/bin/env python3
"""Convert extractor.py JSON output into SQL INSERT statements for `profiles`
and `profile_playtimes`.

Filled columns: mc_uuid, mc_username, first_seen_at, last_seen_at, is_slim,
role ('player'). All other columns use their DDL defaults.

is_slim is resolved through the Mojang API by username (same flow as the API's
MojangService): username -> Mojang UUID -> skin model. When the username is not
known to Mojang, the model comes from the client's default skin for the
player's UUID (Steve/Alex/... picked by UUID hash).

profile_playtimes rows come from each player's "world_playtimes". The world
path -> season id mapping is a JSON object passed with --season-map, e.g.
{"worlds/season1": "<season uuid>", "worlds/season2": "<season uuid>"}. Keys
must match the world paths in the extractor output exactly. Worlds mapped to
the same season are summed. Playtime is stored in milliseconds. The INSERTs
use ON DUPLICATE KEY UPDATE on (mc_uuid, season_id), so re-running is safe.
"""

from __future__ import annotations

import argparse
import base64
import json
import sys
import time
import uuid as uuid_module
import urllib.error
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

MOJANG_UUID_URL = "https://api.mojang.com/users/profiles/minecraft/{}"
MOJANG_PROFILE_URL = "https://sessionserver.mojang.com/session/minecraft/profile/{}"
REQUEST_DELAY_SECONDS = 0.3
MAX_RETRIES = 3
MILLIS_PER_HOUR = 3_600_000


def http_get_json(url: str) -> dict | None:
    """Return parsed JSON, or None on 204/404. Retries on 429."""
    for attempt in range(MAX_RETRIES):
        request = urllib.request.Request(url, headers={"User-Agent": "lania-season-extractor"})
        try:
            with urllib.request.urlopen(request, timeout=15) as response:
                if response.status == 204:
                    return None
                return json.load(response)
        except urllib.error.HTTPError as exc:
            if exc.code == 404:
                return None
            if exc.code == 429 and attempt < MAX_RETRIES - 1:
                time.sleep(2 ** (attempt + 2))
                continue
            raise
    return None


def is_slim_model(username: str) -> bool | None:
    """True/False from the Mojang skin model, None if the player is not found."""
    profile_ref = http_get_json(MOJANG_UUID_URL.format(username))
    if not profile_ref or "id" not in profile_ref:
        return None

    time.sleep(REQUEST_DELAY_SECONDS)
    profile = http_get_json(MOJANG_PROFILE_URL.format(profile_ref["id"]))
    if not profile:
        return None

    for prop in profile.get("properties", []):
        if prop.get("name") == "textures":
            textures = json.loads(base64.b64decode(prop["value"]))
            model = textures.get("textures", {}).get("SKIN", {}).get("metadata", {}).get("model", "")
            return model.lower() == "slim"
    return False


def default_skin_is_slim(player_uuid: str) -> bool:
    """Mirror the 1.19.3+ client: DEFAULT_SKINS[floorMod(uuid.hashCode(), 18)],
    where indices 0-8 are the slim skins and 9-17 the wide ones."""
    value = uuid_module.UUID(player_uuid).int
    msb, lsb = value >> 64, value & 0xFFFFFFFFFFFFFFFF
    hilo = msb ^ lsb
    java_hash = ((hilo >> 32) ^ hilo) & 0xFFFFFFFF
    if java_hash >= 0x80000000:
        java_hash -= 0x100000000
    return java_hash % 18 < 9


def sql_string(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace("'", "''")
    return f"'{escaped}'"


def sql_timestamp(iso: str | None) -> str:
    if not iso:
        return "NULL"
    utc = datetime.fromisoformat(iso).astimezone(timezone.utc)
    return f"'{utc.strftime('%Y-%m-%d %H:%M:%S')}'"


def build_insert(player: dict, slim: bool) -> str:
    return (
        "INSERT INTO profiles "
        "(mc_uuid, mc_username, first_seen_at, last_seen_at, role, is_slim) VALUES ("
        f"{sql_string(player['uuid'])}, "
        f"{sql_string(player['username'])}, "
        f"{sql_timestamp(player.get('first_join'))}, "
        f"{sql_timestamp(player.get('last_played'))}, "
        "'player', "
        f"{'TRUE' if slim else 'FALSE'});"
    )


def load_season_map(path: Path) -> dict[str, str]:
    """Read the world path -> season id JSON file, validating the season ids."""
    mapping = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(mapping, dict):
        raise ValueError(f"{path}: expected a JSON object of world path -> season id")
    for world, season_id in mapping.items():
        try:
            uuid_module.UUID(season_id)
        except (TypeError, ValueError):
            raise ValueError(f"{path}: invalid season id for {world!r}: {season_id!r}") from None
    return mapping


def find_unmapped_worlds(players: list[dict], season_map: dict[str, str]) -> list[str]:
    worlds = {entry["world"] for player in players for entry in player.get("world_playtimes", [])}
    return sorted(worlds - season_map.keys())


def build_playtime_inserts(player: dict, season_map: dict[str, str]) -> list[str]:
    """One INSERT per season; worlds mapped to the same season are summed."""
    millis_by_season: dict[str, int] = {}
    for entry in player.get("world_playtimes", []):
        season_id = season_map[entry["world"]]
        millis = round(entry["playtime_hours"] * MILLIS_PER_HOUR)
        millis_by_season[season_id] = millis_by_season.get(season_id, 0) + millis

    return [
        "INSERT INTO profile_playtimes (mc_uuid, season_id, playtime) VALUES ("
        f"{sql_string(player['uuid'])}, {sql_string(season_id)}, {millis}) "
        "ON DUPLICATE KEY UPDATE playtime = VALUES(playtime);"
        for season_id, millis in millis_by_season.items()
        if millis > 0
    ]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("input", type=Path, help="JSON file produced by extractor.py")
    parser.add_argument("--output", type=Path, default=None, help="Write SQL to file instead of stdout")
    parser.add_argument(
        "--skip-mojang",
        action="store_true",
        help="Do not call the Mojang API; derive is_slim from the default skin for the UUID",
    )
    parser.add_argument(
        "--season-map",
        type=Path,
        required=True,
        help="JSON file mapping world path (as in the extractor output) -> season id",
    )
    args = parser.parse_args()

    players = json.loads(args.input.read_text(encoding="utf-8"))
    try:
        season_map = load_season_map(args.season_map)
    except (OSError, ValueError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    unmapped = find_unmapped_worlds(players, season_map)
    if unmapped:
        print("error: worlds missing from --season-map: " + ", ".join(unmapped), file=sys.stderr)
        return 1

    statements = []
    for index, player in enumerate(players, start=1):
        username = player["username"]
        if username == player["uuid"]:
            print(f"warning: no username for {username}, skipped", file=sys.stderr)
            continue

        slim = default_skin_is_slim(player["uuid"])
        if not args.skip_mojang:
            try:
                result = is_slim_model(username)
            except (urllib.error.URLError, ValueError) as exc:
                print(f"warning: Mojang lookup failed for {username}: {exc}; using default skin", file=sys.stderr)
                result = None
            source = "mojang"
            if result is None:
                source = "default skin"
            else:
                slim = result
            print(f"[{index}/{len(players)}] {username}: slim={slim} ({source})", file=sys.stderr)
            time.sleep(REQUEST_DELAY_SECONDS)

        statements.append(build_insert(player, slim))
        statements.extend(build_playtime_inserts(player, season_map))

    sql = "\n".join(statements) + "\n"
    if args.output is not None:
        args.output.write_text(sql, encoding="utf-8")
    else:
        print(sql, end="")
    return 0


if __name__ == "__main__":
    sys.exit(main())
