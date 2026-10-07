package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

func passwordFromPID(pid uint32) (string, uint32) {
	// 3DS consoles log in with their friends-account PID and the password the
	// console generated for it, not the proxy-issued credentials. Those players
	// need a pretendo.nexaccounts record holding the console's real password.
	if account := getNEXAccountByPID(pid); account != nil {
		if password, ok := account["password"].(string); ok && password != "" {
			return password, 0
		}
	}

	// The HTTPS proxy gives the 3DS this exact per-player password in its
	// nex_token response. Derive it here as well so Auth can issue a ticket
	// without relying on an account record that the proxy never creates.
	secret := os.Getenv("NEX_ACCOUNT_SECRET")
	if len(secret) != 64 {
		return "", 0x80010001
	}

	message := fmt.Sprintf("badgecade-nex-account:%d", pid)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)), 0
}
