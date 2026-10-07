package nex_datastore_nintendo_badge_arcade

import (
	"fmt"

	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/database"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/utility"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-go/datastore"
	datastore_nintendo_badge_arcade "github.com/PretendoNetwork/nex-protocols-go/datastore/nintendo-badge-arcade"
)

func GetMetaByOwnerID(err error, client *nex.Client, callID uint32, param *datastore_nintendo_badge_arcade.DataStoreGetMetaByOwnerIDParam) {
	pMetaInfo := make([]*datastore.DataStoreMetaInfo, 0)
	var pHasNext bool = false // Unknown

	if len(param.OwnerIDs) != len(param.DataTypes) {
		// Not sure if this is possible in the first place
		fmt.Println("WARNING: Owner ID and DataType length mismatch")
	}

	for i, _ := range param.OwnerIDs {
		switch param.DataTypes[i] {
		case 100: // Free Play Data
			lookupID := param.OwnerIDs[i]
			if lookupID == client.PID() {
				lookupID = database.CanonicalOwnerID(lookupID)
			}
			freePlayDataMetaInfo := utility.FreePlayDataToDataStoreMetaInfo(lookupID, param.DataTypes[i])
			if freePlayDataMetaInfo != nil {
				freePlayDataMetaInfo.OwnerID = param.OwnerIDs[i]
				pMetaInfo = append(pMetaInfo, freePlayDataMetaInfo)
			}
		default:
			fmt.Println("WARNING: Unknown DataType: %d", param.DataTypes[i])
		}
	}

	rmcResponseStream := nex.NewStreamOut(globals.NEXServer)

	rmcResponseStream.WriteListStructure(pMetaInfo)
	rmcResponseStream.WriteBool(pHasNext)

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCResponse(datastore_nintendo_badge_arcade.ProtocolID, callID)
	rmcResponse.SetSuccess(datastore_nintendo_badge_arcade.MethodGetMetaByOwnerID, rmcResponseBody)

	rmcResponseBytes := rmcResponse.Bytes()

	responsePacket, _ := nex.NewPacketV1(client, nil)

	responsePacket.SetVersion(1)
	responsePacket.SetSource(0xA1)
	responsePacket.SetDestination(0xAF)
	responsePacket.SetType(nex.DataPacket)
	responsePacket.SetPayload(rmcResponseBytes)

	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.AddFlag(nex.FlagReliable)

	globals.NEXServer.Send(responsePacket)
}
