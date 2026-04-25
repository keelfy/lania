package domain

import (
	"github.com/google/uuid"
)

type LuckpermsPermissionServer string

const (
	LuckpermsPermissionServerGlobal LuckpermsPermissionServer = "global"
)

type LuckpermsPermissionWorld string

const (
	LuckpermsPermissionWorldGlobal LuckpermsPermissionWorld = "global"
)

type LuckpermsPermissionValue int

const (
	LuckpermsPermissionValueEnabled  LuckpermsPermissionValue = 1
	LuckpermsPermissionValueDisabled LuckpermsPermissionValue = 0
)

type LuckpermsUserPermission struct {
	ID         int64
	UUID       uuid.UUID
	Permission string
	Value      LuckpermsPermissionValue
	Server     LuckpermsPermissionServer
	World      LuckpermsPermissionWorld
	Expiry     int64
	Contexts   string
}
