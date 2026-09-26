package config

import (
	"os"
)

/** SYSTEM */

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "9090"
	}
	return port
}

func IsDebug() bool {
	return os.Getenv("DEBUG") == "true"
}

/** AUTH */

// GetToken returns the shared secret clients must send as a bearer token.
func GetToken() string {
	return os.Getenv("SHELL_TOKEN")
}

/** RCON */

// GetRconAddress returns host:port of the Minecraft server RCON. Empty disables permission sync and fails whitelist changes.
func GetRconAddress() string {
	return os.Getenv("RCON_ADDRESS")
}

func GetRconPassword() string {
	return os.Getenv("RCON_PASSWORD")
}

// GetLuckpermsSyncCommand returns the console command that makes the running
// server reload permissions from the database. Use "lp networksync" when
// LuckPerms messaging is set up across several servers.
func GetLuckpermsSyncCommand() string {
	command := os.Getenv("LUCKPERMS_SYNC_COMMAND")
	if command == "" {
		return "lp sync"
	}
	return command
}

// GetWhitelistAddCommand returns the console command template that whitelists
// a player. {username} and {uuid} are replaced with the player values.
// Override it when a whitelist plugin replaces the vanilla whitelist.
func GetWhitelistAddCommand() string {
	command := os.Getenv("WHITELIST_ADD_COMMAND")
	if command == "" {
		return "whitelist add {username}"
	}
	return command
}

// GetWhitelistRemoveCommand is the counterpart of GetWhitelistAddCommand.
func GetWhitelistRemoveCommand() string {
	command := os.Getenv("WHITELIST_REMOVE_COMMAND")
	if command == "" {
		return "whitelist remove {username}"
	}
	return command
}

/** DATABASE */

func GetDatabaseHost() string {
	return os.Getenv("DATABASE_HOST")
}

func GetDatabasePort() string {
	return os.Getenv("DATABASE_PORT")
}

func GetDatabaseUser() string {
	return os.Getenv("DATABASE_USER")
}

func GetDatabasePassword() string {
	return os.Getenv("DATABASE_PASSWORD")
}

// GetDatabaseName returns the single database holding the LuckPerms, Plan and Flectone tables.
func GetDatabaseName() string {
	return os.Getenv("DATABASE_NAME")
}

/** TABLES */

// Plugin tables share one database, so each plugin's tables carry its own prefix.
// The defaults match the plugins' default prefixes; override them when the plugins are configured differently.

func GetLuckpermsUserPermissionsTableName() string {
	return getEnvOrDefault("LUCKPERMS_USER_PERMISSIONS_TABLE_NAME", "luckperms_user_permissions")
}

func GetLuckpermsPlayersTableName() string {
	return getEnvOrDefault("LUCKPERMS_PLAYERS_TABLE_NAME", "luckperms_players")
}

func GetPlanUsersTableName() string {
	return getEnvOrDefault("PLAN_USERS_TABLE_NAME", "plan_users")
}

func GetPlanSessionsTableName() string {
	return getEnvOrDefault("PLAN_SESSIONS_TABLE_NAME", "plan_sessions")
}

func GetPlanServersTableName() string {
	return getEnvOrDefault("PLAN_SERVERS_TABLE_NAME", "plan_servers")
}

func GetFlectonePlayerTableName() string {
	return getEnvOrDefault("FLECTONE_PLAYER_TABLE_NAME", "player")
}

func getEnvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
