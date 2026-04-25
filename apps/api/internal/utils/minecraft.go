package utils

import (
	"crypto/md5"
	"fmt"

	"github.com/google/uuid"
)

const OfflinePlayerUUIDFormat = "OfflinePlayer:%s"

// Generate Minecraft offline player UUID from username
func GetOfflinePlayerUUID(username string) (uuid.UUID, error) {
	data := []byte(fmt.Sprintf(OfflinePlayerUUIDFormat, username))
	hash := md5.Sum(data)

	// Modify version bits (UUID v3)
	hash[6] = (hash[6] & 0x0f) | 0x30 // Version 3
	hash[8] = (hash[8] & 0x3f) | 0x80 // Variant 1

	return uuid.FromBytes(hash[:])
}
