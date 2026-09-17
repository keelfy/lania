#!/bin/bash
set -euo pipefail

# Creates the databases the API expects on first boot of the MariaDB volume.
# NOTE: if DATABASE_FLECTONE_NAME / DATABASE_PLAN_NAME are meant to read data
# produced by the Minecraft network's Flectone/Plan plugins, point DATABASE_HOST
# in .env at that existing MariaDB instance instead of using this local one.

mysql -u root -p"${MARIADB_ROOT_PASSWORD}" <<-EOSQL
  CREATE DATABASE IF NOT EXISTS \`${DATABASE_NAME}\`;
  CREATE DATABASE IF NOT EXISTS \`${DATABASE_FLECTONE_NAME}\`;
  CREATE DATABASE IF NOT EXISTS \`${DATABASE_PLAN_NAME}\`;
EOSQL
