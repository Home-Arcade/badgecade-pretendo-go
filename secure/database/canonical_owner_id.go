package database

import (
	"context"
	"time"

	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CanonicalOwnerID returns the Badgecade account that owns a console's saves.
//
// A console logs in with a different PID depending on its Nimbus mode
// (Nintendo NNID or Pretendo account), but Auth binds each login PID to the
// player's Badgecade proxy account in pretendo.consolebindings. Storing saves
// under that account lets one save follow the player across modes and
// consoles. PIDs without a binding keep owning their own saves.
func CanonicalOwnerID(pid uint32) uint32 {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var binding bson.M
	err := accountDatabase.Collection("consolebindings").FindOne(ctx, bson.D{{Key: "pid", Value: int64(pid)}}).Decode(&binding)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			globals.Logger.Error(err.Error())
		}
		return pid
	}

	switch owner := binding["proxy_pid"].(type) {
	case int64:
		if owner > 0 && owner <= 0xFFFFFFFF {
			return uint32(owner)
		}
	case int32:
		if owner > 0 {
			return uint32(owner)
		}
	}
	return pid
}
