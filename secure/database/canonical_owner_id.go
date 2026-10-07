package database

import (
	"context"
	"time"

	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CanonicalOwnerID maps a console PID to its Badgecade account (from
// consolebindings) so the save is the same in both Nimbus modes
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
