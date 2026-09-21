#!/bin/bash
set -euo pipefail

# Creates the database shared by the API and shell on first boot of the MariaDB volume.
# NOTE: if the LuckPerms/Plan/Flectone tables live in the Minecraft network's own
# MariaDB, point DATABASE_HOST in .env at that instance instead of using this local one.

mysql -u root -p"${MARIADB_ROOT_PASSWORD}" <<-EOSQL
  CREATE DATABASE IF NOT EXISTS \`${DATABASE_NAME}\`;
EOSQL
