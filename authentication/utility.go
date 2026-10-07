package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

func passwordFromPID(pid uint32) (string, uint32) {
	// use the stored password if there is one
	if account := getNEXAccountByPID(pid); account != nil {
		if password, ok := account["password"].(string); ok && password != "" {
			return password, 0
		}
	}

	// otherwise use the same password the proxy hands out in nex_token
	secret := os.Getenv("NEX_ACCOUNT_SECRET")
	if len(secret) != 64 {
		return "", 0x80010001
	}

	message := fmt.Sprintf("badgecade-nex-account:%d", pid)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)), 0
}
