#!/usr/bin/env python3
"""Extract username, playtime, first-join and last-played dates for every
player in one or more Minecraft world folders. Records are combined when
they share a UUID (username taken from the latest last played) or a username
(offline-mode UUID wins). Combined records sum playtime, take the earliest
first join and the latest last played.

Reads:
  - <world>/playerdata/<uuid>.dat  (NBT: "bukkit.lastKnownName",
                                    "bukkit.firstPlayed", "bukkit.lastPlayed")
  - <world>/stats/<uuid>.json      ("minecraft:play_time" in ticks)
  - usercache.json (optional, --usercache): UUID -> username fallback for
    players whose .dat has no "bukkit.lastKnownName"

No third-party dependencies; NBT is parsed with a small built-in reader.
"""

from __future__ import annotations

import argparse
import csv
import gzip
import hashlib
import io
import json
import struct
import sys
import uuid as uuid_module
from dataclasses import dataclass, asdict
from datetime import datetime, timezone
from pathlib import Path

TICKS_PER_SECOND = 20

TAG_END = 0
TAG_BYTE = 1
TAG_SHORT = 2
TAG_INT = 3
TAG_LONG = 4
TAG_FLOAT = 5
TAG_DOUBLE = 6
TAG_BYTE_ARRAY = 7
TAG_STRING = 8
TAG_LIST = 9
TAG_COMPOUND = 10
TAG_INT_ARRAY = 11
TAG_LONG_ARRAY = 12


def _read_string(buf: io.BufferedReader) -> str:
    (length,) = struct.unpack(">H", buf.read(2))
    return buf.read(length).decode("utf-8", errors="replace")


def _read_payload(buf: io.BufferedReader, tag_type: int):
    if tag_type == TAG_BYTE:
        return struct.unpack(">b", buf.read(1))[0]
    if tag_type == TAG_SHORT:
        return struct.unpack(">h", buf.read(2))[0]
    if tag_type == TAG_INT:
        return struct.unpack(">i", buf.read(4))[0]
    if tag_type == TAG_LONG:
        return struct.unpack(">q", buf.read(8))[0]
    if tag_type == TAG_FLOAT:
        return struct.unpack(">f", buf.read(4))[0]
    if tag_type == TAG_DOUBLE:
        return struct.unpack(">d", buf.read(8))[0]
    if tag_type == TAG_BYTE_ARRAY:
        (length,) = struct.unpack(">i", buf.read(4))
        return buf.read(length)
    if tag_type == TAG_STRING:
        return _read_string(buf)
    if tag_type == TAG_LIST:
        (elem_type,) = struct.unpack(">b", buf.read(1))
        (length,) = struct.unpack(">i", buf.read(4))
        return [_read_payload(buf, elem_type) for _ in range(length)]
    if tag_type == TAG_COMPOUND:
        result = {}
        while True:
            (child_type,) = struct.unpack(">b", buf.read(1))
            if child_type == TAG_END:
                break
            name = _read_string(buf)
            result[name] = _read_payload(buf, child_type)
        return result
    if tag_type == TAG_INT_ARRAY:
        (length,) = struct.unpack(">i", buf.read(4))
        return struct.unpack(f">{length}i", buf.read(4 * length))
    if tag_type == TAG_LONG_ARRAY:
        (length,) = struct.unpack(">i", buf.read(4))
        return struct.unpack(f">{length}q", buf.read(8 * length))
    raise ValueError(f"unknown NBT tag type: {tag_type}")


def parse_nbt(raw: bytes) -> dict:
    buf = io.BytesIO(gzip.decompress(raw))
    (root_type,) = struct.unpack(">b", buf.read(1))
    if root_type != TAG_COMPOUND:
        raise ValueError("NBT root is not a compound tag")
    _read_string(buf)  # root name, usually empty
    return _read_payload(buf, TAG_COMPOUND)


@dataclass
class PlayerRecord:
    uuid: str
    username: str
    playtime_hours: float
    first_join: str | None
    last_played: str | None


def read_playtime_hours(stats_path: Path) -> float | None:
    if not stats_path.is_file():
        return None
    data = json.loads(stats_path.read_text(encoding="utf-8"))
    custom = data.get("stats", {}).get("minecraft:custom", {})
    ticks = custom.get("minecraft:play_time", custom.get("minecraft:play_one_minute"))
    if ticks is None:
        return None
    return round(ticks / TICKS_PER_SECOND / 3600, 2)


def _millis_to_iso(millis: int | None) -> str | None:
    if not millis:
        return None
    return datetime.fromtimestamp(millis / 1000, tz=timezone.utc).isoformat()


def load_usercache(path: Path) -> dict[str, str]:
    """Read a server usercache.json into a {uuid: username} map."""
    entries = json.loads(path.read_text(encoding="utf-8"))
    names = {}
    for entry in entries:
        try:
            key = str(uuid_module.UUID(entry["uuid"]))
        except (KeyError, ValueError):
            continue
        if entry.get("name"):
            names[key] = entry["name"]
    return names


def extract_world(world_dir: Path, usercache: dict[str, str] | None = None) -> list[PlayerRecord]:
    playerdata_dir = world_dir / "playerdata"
    stats_dir = world_dir / "stats"
    if not playerdata_dir.is_dir():
        raise FileNotFoundError(f"playerdata folder not found: {playerdata_dir}")

    usercache = usercache or {}
    records = []
    for dat_file in sorted(playerdata_dir.glob("*.dat")):
        uuid = dat_file.stem
        bukkit = parse_nbt(dat_file.read_bytes()).get("bukkit", {})
        username = bukkit.get("lastKnownName") or usercache.get(uuid.lower(), uuid)
        playtime_hours = read_playtime_hours(stats_dir / f"{uuid}.json")
        records.append(
            PlayerRecord(
                uuid=uuid,
                username=username,
                playtime_hours=playtime_hours if playtime_hours is not None else 0.0,
                first_join=_millis_to_iso(bukkit.get("firstPlayed")),
                last_played=_millis_to_iso(bukkit.get("lastPlayed")),
            )
        )

    return records


def offline_uuid(name: str) -> str:
    """Reproduce Java's UUID.nameUUIDFromBytes("OfflinePlayer:<name>")."""
    digest = bytearray(hashlib.md5(b"OfflinePlayer:" + name.encode("utf-8")).digest())
    digest[6] = (digest[6] & 0x0F) | 0x30  # version 3
    digest[8] = (digest[8] & 0x3F) | 0x80  # variant RFC 4122
    return str(uuid_module.UUID(bytes=bytes(digest)))


def _combine(group: list[PlayerRecord], identity: PlayerRecord) -> PlayerRecord:
    """Merge records; uuid and username come from `identity`."""
    first_joins = [r.first_join for r in group if r.first_join]
    last_played = [r.last_played for r in group if r.last_played]
    return PlayerRecord(
        uuid=identity.uuid,
        username=identity.username,
        playtime_hours=round(sum(r.playtime_hours for r in group), 2),
        first_join=min(first_joins) if first_joins else None,
        last_played=max(last_played) if last_played else None,
    )


def _latest(group: list[PlayerRecord]) -> PlayerRecord:
    # ISO-8601 UTC strings compare chronologically.
    return max(group, key=lambda r: r.last_played or "")


def merge_records(records: list[PlayerRecord]) -> list[PlayerRecord]:
    by_uuid: dict[str, list[PlayerRecord]] = {}
    for record in records:
        by_uuid.setdefault(record.uuid, []).append(record)
    # Same UUID: username comes from the most recently played record.
    uuid_merged = [_combine(group, _latest(group)) for group in by_uuid.values()]

    by_name: dict[str, list[PlayerRecord]] = {}
    for record in uuid_merged:
        by_name.setdefault(record.username.lower(), []).append(record)

    result = []
    for group in by_name.values():
        if len(group) == 1:
            result.append(group[0])
            continue
        # Same username, different UUIDs: the offline-mode UUID record wins.
        offline = [r for r in group if r.uuid == offline_uuid(r.username)]
        identity = _latest(offline) if offline else _latest(group)
        result.append(_combine(group, identity))

    result.sort(key=lambda r: (r.first_join is None, r.first_join))
    return result


def extract(world_dirs: list[Path], usercache_paths: list[Path] | None = None) -> list[PlayerRecord]:
    usercache = {}
    for path in usercache_paths or []:
        usercache.update(load_usercache(path))
    records = []
    for world_dir in world_dirs:
        records.extend(extract_world(world_dir, usercache))
    return merge_records(records)


def write_output(records: list[PlayerRecord], output: Path | None, fmt: str) -> None:
    if fmt == "json":
        text = json.dumps([asdict(r) for r in records], indent=2)
    else:
        buf = io.StringIO()
        writer = csv.DictWriter(
            buf, fieldnames=["uuid", "username", "playtime_hours", "first_join", "last_played"]
        )
        writer.writeheader()
        for r in records:
            writer.writerow(asdict(r))
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
        records = extract(args.worlds, args.usercache)
    except (FileNotFoundError, json.JSONDecodeError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    write_output(records, args.output, args.format)
    return 0


if __name__ == "__main__":
    sys.exit(main())
