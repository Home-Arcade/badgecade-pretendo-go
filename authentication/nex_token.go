package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const nexTokenMaxAge = 24 * time.Hour

// pidFromNEXToken returns the proxy PID signed into a LoginEx token by the
// HTTPS proxy's nex_token response. 3DS consoles log in with their locally
// stored PID but use the password from that response, so this PID decides
// which password the ticket must be encrypted with.
func pidFromNEXToken(token string) (uint32, bool) {
	secret := os.Getenv("NEX_ACCOUNT_SECRET")
	if len(secret) != 64 {
		return 0, false
	}

	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil || len(raw) != 24 {
		return 0, false
	}

	body, signature := raw[:8], raw[8:]
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("badgecade-nex-token:"))
	_, _ = mac.Write(body)
	if !hmac.Equal(signature, mac.Sum(nil)[:16]) {
		return 0, false
	}

	issued := time.Unix(int64(binary.BigEndian.Uint32(body[4:8])), 0)
	if age := time.Since(issued); age > nexTokenMaxAge || age < -5*time.Minute {
		return 0, false
	}

	return binary.BigEndian.Uint32(body[0:4]), true
}

// bindConsolePID ties a console's login PID to the proxy account that first
// used it, so one player cannot log in as another player's console PID.
func bindConsolePID(consolePID uint32, proxyPID uint32) bool {
	bindings := mongoDatabase.Collection("consolebindings")
	filter := bson.D{{Key: "pid", Value: consolePID}}
	update := bson.D{{Key: "$setOnInsert", Value: bson.D{
		{Key: "pid", Value: consolePID},
		{Key: "proxy_pid", Value: proxyPID},
		{Key: "created_at", Value: time.Now().UTC()},
	}}}

	var binding bson.M
	err := bindings.FindOneAndUpdate(
		context.TODO(),
		filter,
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&binding)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Error(err.Error())
		}
		return false
	}

	switch bound := binding["proxy_pid"].(type) {
	case int64:
		return bound == int64(proxyPID)
	case int32:
		return bound == int32(proxyPID)
	}
	return false
}
