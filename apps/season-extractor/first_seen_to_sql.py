#!/usr/bin/env python3
"""Convert extractor.py JSON output into SQL that moves `profiles.first_seen_at`
of existing profiles back to the earliest first join found in the worlds
(the earliest of "bukkit.firstPlayed" and the advancement timestamps).

World UUIDs are resolved to the profile holding them now exactly as in
stats_to_sql.py (mc_uuid, then legacy_mc_uuid, then profile_former_uuids);
UUIDs resolving to the same profile take the earliest date. first_seen_at is
only set when it is empty or later than the world date, so re-running is safe
and an earlier date from another source is kept. The first SELECT lists the
profiles that change (stored vs world date); the last one lists unresolved
UUIDs. Dates are written in UTC, as in to_sql.py.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from stats_to_sql import INSERT_BATCH_SIZE, RESOLVE_UUIDS
from to_sql import sql_string, sql_timestamp

HEADER = """\
CREATE TEMPORARY TABLE first_seen_import (
    world_mc_uuid uuid NOT NULL,
    mc_username tinytext NOT NULL,
    first_seen_at timestamp NOT NULL,
    mc_uuid uuid NULL
);
"""

FOOTER = RESOLVE_UUIDS.format(table="first_seen_import") + """
CREATE TEMPORARY TABLE first_seen_import_totals AS
SELECT mc_uuid, MIN(first_seen_at) AS first_seen_at
FROM first_seen_import
WHERE mc_uuid IS NOT NULL
GROUP BY mc_uuid;

SELECT p.mc_username, p.first_seen_at AS stored_first_seen_at, t.first_seen_at AS world_first_seen_at
FROM first_seen_import_totals t
JOIN profiles p ON p.mc_uuid = t.mc_uuid
WHERE p.first_seen_at IS NULL OR p.first_seen_at > t.first_seen_at
ORDER BY p.mc_username;

UPDATE profiles p
JOIN first_seen_import_totals t ON t.mc_uuid = p.mc_uuid
SET p.first_seen_at = t.first_seen_at
WHERE p.first_seen_at IS NULL OR p.first_seen_at > t.first_seen_at;

SELECT world_mc_uuid, mc_username, first_seen_at
FROM first_seen_import
WHERE mc_uuid IS NULL
ORDER BY mc_username;
"""


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("input", type=Path, help="JSON file produced by extractor.py")
    parser.add_argument("--output", type=Path, default=None, help="Write SQL to file instead of stdout")
    args = parser.parse_args()

    try:
        players = json.loads(args.input.read_text(encoding="utf-8"))
    except (OSError, ValueError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    rows = [
        f"({sql_string(player['uuid'])}, {sql_string(player['username'])}, {sql_timestamp(player['first_join'])})"
        for player in players
        if player.get("first_join")
    ]
    parts = [HEADER]
    for start in range(0, len(rows), INSERT_BATCH_SIZE):
        parts.append(
            "INSERT INTO first_seen_import (world_mc_uuid, mc_username, first_seen_at) VALUES\n"
            + ",\n".join(rows[start : start + INSERT_BATCH_SIZE])
            + ";\n"
        )
    parts.append(FOOTER)

    sql = "\n".join(parts)
    if args.output is not None:
        args.output.write_text(sql, encoding="utf-8")
    else:
        print(sql, end="")
    return 0


if __name__ == "__main__":
    sys.exit(main())
