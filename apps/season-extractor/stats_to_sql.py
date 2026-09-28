#!/usr/bin/env python3
"""Convert stats.py JSON output into SQL that fills `profile_season_stats`
(playtime, deaths, mob_kills) of profiles that already exist, for example
seasons imported with to_sql.py where some playtime was lost.

The UUIDs in the worlds are the ones the players had then. The site may have
moved a profile to another UUID since (premium rekey to the Mojang UUID,
nickname change, profile merge), and profile_season_stats follows
profiles.mc_uuid. So the SQL resolves every world UUID to the profile holding
it now, in this order:
  1. profiles.mc_uuid          (unchanged profile)
  2. profiles.legacy_mc_uuid   (offline UUID before the premium rekey)
  3. profile_former_uuids      (UUID left by a merge or a nickname change)
Counts of world UUIDs resolving to the same profile and season are summed, as
the profile merge sums the season stats. Missing rows are inserted; existing
rows only grow (GREATEST of the stored and the world value), so re-running is
safe and data from other sources is never lowered. The first SELECT lists the
rows that change (stored vs world values); the last one lists unresolved UUIDs.

The world path -> season id mapping is the --season-map file of to_sql.py.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from to_sql import load_season_map, sql_string

INSERT_BATCH_SIZE = 500

HEADER = """\
CREATE TEMPORARY TABLE season_stats_import (
    world_mc_uuid uuid NOT NULL,
    mc_username tinytext NOT NULL,
    season_id uuid NOT NULL,
    playtime bigint NOT NULL,
    deaths bigint NOT NULL,
    mob_kills bigint NOT NULL,
    mc_uuid uuid NULL
);
"""

# Fills mc_uuid of an import table with the profile holding world_mc_uuid now.
RESOLVE_UUIDS = """\
UPDATE {table} i
SET i.mc_uuid = COALESCE(
    (SELECT p.mc_uuid FROM profiles p WHERE p.mc_uuid = i.world_mc_uuid),
    (SELECT p.mc_uuid FROM profiles p WHERE p.legacy_mc_uuid = i.world_mc_uuid LIMIT 1),
    (SELECT p.mc_uuid FROM profile_former_uuids f JOIN profiles p ON p.id = f.profile_id
        WHERE f.mc_uuid = i.world_mc_uuid)
);
"""

FOOTER = RESOLVE_UUIDS.format(table="season_stats_import") + """
CREATE TEMPORARY TABLE season_stats_import_totals AS
SELECT mc_uuid, season_id, SUM(playtime) AS playtime, SUM(deaths) AS deaths, SUM(mob_kills) AS mob_kills
FROM season_stats_import
WHERE mc_uuid IS NOT NULL
GROUP BY mc_uuid, season_id;

SELECT p.mc_username, t.season_id,
    pss.playtime AS stored_playtime, t.playtime AS world_playtime,
    pss.deaths AS stored_deaths, t.deaths AS world_deaths,
    pss.mob_kills AS stored_mob_kills, t.mob_kills AS world_mob_kills
FROM season_stats_import_totals t
JOIN profiles p ON p.mc_uuid = t.mc_uuid
LEFT JOIN profile_season_stats pss ON pss.mc_uuid = t.mc_uuid AND pss.season_id = t.season_id
WHERE pss.mc_uuid IS NULL OR pss.playtime < t.playtime OR pss.deaths < t.deaths OR pss.mob_kills < t.mob_kills
ORDER BY p.mc_username, t.season_id;

INSERT INTO profile_season_stats (mc_uuid, season_id, playtime, deaths, mob_kills)
SELECT mc_uuid, season_id, playtime, deaths, mob_kills FROM season_stats_import_totals
ON DUPLICATE KEY UPDATE
    profile_season_stats.playtime = GREATEST(profile_season_stats.playtime, VALUES(playtime)),
    profile_season_stats.deaths = GREATEST(profile_season_stats.deaths, VALUES(deaths)),
    profile_season_stats.mob_kills = GREATEST(profile_season_stats.mob_kills, VALUES(mob_kills));

SELECT world_mc_uuid, mc_username, season_id, playtime, deaths, mob_kills
FROM season_stats_import
WHERE mc_uuid IS NULL
ORDER BY mc_username;
"""


def build_rows(players: list[dict], season_map: dict[str, str]) -> list[str]:
    """One VALUES tuple per player and season; worlds mapped to the same season are summed."""
    rows = []
    for player in players:
        counts_by_season: dict[str, list[int]] = {}
        for entry in player.get("world_stats", []):
            counts = counts_by_season.setdefault(season_map[entry["world"]], [0, 0, 0])
            counts[0] += entry["playtime"]
            counts[1] += entry["deaths"]
            counts[2] += entry["mob_kills"]
        for season_id, (playtime, deaths, mob_kills) in counts_by_season.items():
            if playtime == 0 and deaths == 0 and mob_kills == 0:
                continue
            rows.append(
                f"({sql_string(player['uuid'])}, {sql_string(player['username'])}, "
                f"{sql_string(season_id)}, {playtime}, {deaths}, {mob_kills})"
            )
    return rows


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("input", type=Path, help="JSON file produced by stats.py")
    parser.add_argument("--output", type=Path, default=None, help="Write SQL to file instead of stdout")
    parser.add_argument(
        "--season-map",
        type=Path,
        required=True,
        help="JSON file mapping world path (as in the stats.py output) -> season id",
    )
    args = parser.parse_args()

    try:
        players = json.loads(args.input.read_text(encoding="utf-8"))
        season_map = load_season_map(args.season_map)
    except (OSError, ValueError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    worlds = {entry["world"] for player in players for entry in player.get("world_stats", [])}
    unmapped = sorted(worlds - season_map.keys())
    if unmapped:
        print("error: worlds missing from --season-map: " + ", ".join(unmapped), file=sys.stderr)
        return 1

    rows = build_rows(players, season_map)
    parts = [HEADER]
    for start in range(0, len(rows), INSERT_BATCH_SIZE):
        batch = rows[start : start + INSERT_BATCH_SIZE]
        parts.append(
            "INSERT INTO season_stats_import (world_mc_uuid, mc_username, season_id, playtime, deaths, mob_kills) VALUES\n"
            + ",\n".join(batch)
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
