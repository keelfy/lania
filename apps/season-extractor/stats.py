#!/usr/bin/env python3
"""Extract playtime, deaths and mob kills for every player in one or more Minecraft
world folders. Players are identified and combined with the same rules as
extractor.py (shared UUID or username; the offline-mode UUID wins), so the
uuids match extractor.py output for the same worlds and usercache files.
Counts are summed over combined records; each record also lists its counts
per world folder ("world_stats"), summed when worlds are combined.

Reads:
  - <world>/playerdata/<uuid>.dat  (player identity, as in extractor.py)
  - <world>/stats/<uuid>.json      ("minecraft:play_time" in ticks, "minecraft:deaths",
                                    "minecraft:mob_kills")
  - usercache.json (optional, --usercache): as in extractor.py
"""

from __future__ import annotations

import argparse
import csv
import io
import json
import sys
from dataclasses import dataclass, asdict
from pathlib import Path

from extractor import extract_world, group_records, load_usercache

CSV_FIELDS = ["uuid", "username", "playtime", "deaths", "mob_kills", "world_stats"]
MILLIS_PER_TICK = 50


@dataclass
class WorldStats:
    world: str  # world folder path exactly as passed on the command line
    playtime: int  # milliseconds
    deaths: int
    mob_kills: int


@dataclass
class PlayerStats:
    uuid: str
    username: str
    playtime: int  # milliseconds
    deaths: int
    mob_kills: int
    world_stats: list[WorldStats]


def read_stats(stats_path: Path) -> tuple[int, int, int]:
    """(playtime in milliseconds, deaths, mob kills) from a stats file; zeros when the file is missing."""
    if not stats_path.is_file():
        return 0, 0, 0
    data = json.loads(stats_path.read_text(encoding="utf-8"))
    custom = data.get("stats", {}).get("minecraft:custom", {})
    ticks = custom.get("minecraft:play_time", custom.get("minecraft:play_one_minute", 0))
    return ticks * MILLIS_PER_TICK, custom.get("minecraft:deaths", 0), custom.get("minecraft:mob_kills", 0)


def extract(world_dirs: list[Path], usercache_paths: list[Path] | None = None) -> list[PlayerStats]:
    usercache = {}
    for path in usercache_paths or []:
        usercache.update(load_usercache(path))

    records = []
    counts: dict[tuple[str, str], tuple[int, int, int]] = {}  # (world, uuid) -> (playtime, deaths, mob kills)
    for world_dir in world_dirs:
        for record in extract_world(world_dir, usercache):
            records.append(record)
            counts[(str(world_dir), record.uuid)] = read_stats(world_dir / "stats" / f"{record.uuid}.json")

    result = []
    for identity, members in group_records(records):
        per_world: dict[str, list[int]] = {}
        for member in members:
            world = member.world_playtimes[0].world
            totals = per_world.setdefault(world, [0, 0, 0])
            for i, count in enumerate(counts[(world, member.uuid)]):
                totals[i] += count
        world_stats = [WorldStats(world, *totals) for world, totals in per_world.items()]
        result.append(
            PlayerStats(
                uuid=identity.uuid,
                username=identity.username,
                playtime=sum(w.playtime for w in world_stats),
                deaths=sum(w.deaths for w in world_stats),
                mob_kills=sum(w.mob_kills for w in world_stats),
                world_stats=world_stats,
            )
        )
    result.sort(key=lambda s: s.username.lower())
    return result


def write_output(stats: list[PlayerStats], output: Path | None, fmt: str) -> None:
    if fmt == "json":
        text = json.dumps([asdict(s) for s in stats], indent=2)
    else:
        buf = io.StringIO()
        writer = csv.DictWriter(buf, fieldnames=CSV_FIELDS)
        writer.writeheader()
        for s in stats:
            row = asdict(s)
            row["world_stats"] = json.dumps(row["world_stats"])  # JSON string in one cell
            writer.writerow(row)
        text = buf.getvalue()

    if output is not None:
        output.write_text(text, encoding="utf-8")
    else:
        print(text)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "worlds",
        type=Path,
        nargs="+",
        help="One or more world folders (each contains playerdata/, stats/)",
    )
    parser.add_argument(
        "--usercache",
        type=Path,
        action="append",
        default=None,
        help="usercache.json used to resolve usernames for UUIDs without a lastKnownName (repeatable)",
    )
    parser.add_argument("--output", type=Path, default=None, help="Write result to file instead of stdout")
    parser.add_argument("--format", choices=["json", "csv"], default="json")
    args = parser.parse_args()

    try:
        stats = extract(args.worlds, args.usercache)
    except (FileNotFoundError, json.JSONDecodeError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    write_output(stats, args.output, args.format)
    return 0


if __name__ == "__main__":
    sys.exit(main())
