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

func GetDatabasePlanName() string {
	return os.Getenv("DATABASE_PLAN_NAME")
}

func GetDatabaseFlectoneName() string {
	return os.Getenv("DATABASE_FLECTONE_NAME")
}

func GetDatabaseLuckpermsName() string {
	return os.Getenv("DATABASE_LUCKPERMS_NAME")
}

func GetLuckpermsUserPermissionsTableName() string {
	tableName := os.Getenv("LUCKPERMS_USER_PERMISSIONS_TABLE_NAME")
	if tableName == "" {
		return "luckperms_user_permissions"
	}
	return tableName
}
