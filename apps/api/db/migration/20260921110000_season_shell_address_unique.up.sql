-- A shell writes one LuckPerms prefix node per player with no server context, so two seasons
-- cannot share a shell and still show different cosmetics. NULL stays allowed for seasons without a server.
-- Fix seasons that share an address by hand before applying:
--   SELECT shell_address, COUNT(*) FROM seasons WHERE shell_address IS NOT NULL GROUP BY shell_address HAVING COUNT(*) > 1;
CREATE UNIQUE INDEX IF NOT EXISTS idx_seasons_shell_address ON seasons(shell_address);
