package hashHelper

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashHMAC(data string, key string) string {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}
