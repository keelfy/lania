#!/usr/bin/env python3
"""Combine several extractor.py outputs (JSON or CSV) into one file.

Uses the same merge logic as extractor.py between worlds: records sharing a
UUID or a username are combined (playtime summed, earliest first join, latest
last played; the offline-mode UUID wins for a shared username).
"""

from __future__ import annotations

import argparse
import csv
import json
import sys
from pathlib import Path

from extractor import PlayerRecord, merge_records, write_output


def _record_from_dict(row: dict) -> PlayerRecord:
    return PlayerRecord(
        uuid=row["uuid"],
        username=row["username"],
        playtime_hours=float(row.get("playtime_hours") or 0.0),
        first_join=row.get("first_join") or None,
        last_played=row.get("last_played") or None,
    )


def load_records(path: Path) -> list[PlayerRecord]:
    if path.suffix.lower() == ".csv":
        with path.open(newline="", encoding="utf-8") as f:
            rows = list(csv.DictReader(f))
    else:
        rows = json.loads(path.read_text(encoding="utf-8"))
    return [_record_from_dict(row) for row in rows]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "inputs",
        type=Path,
        nargs="+",
        help="extractor.py output files (.json or .csv)",
    )
    parser.add_argument("--output", type=Path, default=None, help="Write result to file instead of stdout")
    parser.add_argument("--format", choices=["json", "csv"], default="json")
    args = parser.parse_args()

    records = []
    try:
        for path in args.inputs:
            records.extend(load_records(path))
    except (OSError, ValueError, KeyError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    write_output(merge_records(records), args.output, args.format)
    return 0


if __name__ == "__main__":
    sys.exit(main())
