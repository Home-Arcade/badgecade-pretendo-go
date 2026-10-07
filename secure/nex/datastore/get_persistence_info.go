package nex_datastore

import (
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/database"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-go/datastore"
)

func GetPersistenceInfo(err error, client *nex.Client, callID uint32, ownerID uint32, persistenceSlotID uint16) {
	// look up the save under the linked Badgecade account
	lookupID := ownerID
	if ownerID == client.PID() {
		lookupID = database.CanonicalOwnerID(ownerID)
	}
	dataID := database.GetDataStorePersistenceInfo(lookupID, persistenceSlotID)

	rmcResponse := nex.NewRMCResponse(datastore.ProtocolID, callID)

	// version 0 = upload never finished, treat as no save
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
