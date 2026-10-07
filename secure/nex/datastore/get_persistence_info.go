package nex_datastore

import (
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/database"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-go/datastore"
)

func GetPersistenceInfo(err error, client *nex.Client, callID uint32, ownerID uint32, persistenceSlotID uint16) {
	// A console asking for its own save gets the save of the Badgecade
	// account it is bound to; the reply still names the console's PID.
	lookupID := ownerID
	if ownerID == client.PID() {
		lookupID = database.CanonicalOwnerID(ownerID)
	}
	dataID := database.GetDataStorePersistenceInfo(lookupID, persistenceSlotID)

	rmcResponse := nex.NewRMCResponse(datastore.ProtocolID, callID)

	// Version 0 means PostMetaBinary created the record but the save upload
	// never completed. Report it as missing so the game uploads a new save
	// instead of trying to download an object that does not exist.
	if dataID != 0 && database.GetVersionByDataID(dataID) != 0 {
		pPersistenceInfo := datastore.NewDataStorePersistenceInfo()
		pPersistenceInfo.OwnerID = ownerID
		pPersistenceInfo.PersistenceSlotID = persistenceSlotID
		pPersistenceInfo.DataID = uint64(dataID)

		rmcResponseStream := nex.NewStreamOut(globals.NEXServer)

		rmcResponseStream.WriteStructure(pPersistenceInfo)

		rmcResponseBody := rmcResponseStream.Bytes()

		rmcResponse.SetSuccess(datastore.MethodGetPersistenceInfo, rmcResponseBody)
	} else {
		rmcResponse.SetError(nex.Errors.DataStore.NotFound)
	}

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
