package nex_datastore

import (
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/database"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"
	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-go/datastore"
)

func PreparePostObject(err error, client *nex.Client, callID uint32, param *datastore.DataStorePreparePostParam) {
	pid := database.CanonicalOwnerID(client.PID())
	const slot uint16 = 0
	dataID := database.GetDataStorePersistenceInfo(pid, slot)
	key := objectKey(dataID, 1)
	postURL, formFields, storeErr := makeObjectUploadInfo(key)
	if storeErr != nil {
		globals.Logger.Error(storeErr.Error())
	}

	postInfo := datastore.NewDataStoreReqPostInfo()
	postInfo.DataID = uint64(dataID)
	postInfo.URL = postURL
	postInfo.RequestHeaders = []*datastore.DataStoreKeyValue{}
	postInfo.FormFields = formFields
	postInfo.RootCACert = []byte{}

	responseStream := nex.NewStreamOut(globals.NEXServer)
	responseStream.WriteStructure(postInfo)
	response := nex.NewRMCResponse(datastore.ProtocolID, callID)
	response.SetSuccess(datastore.MethodPreparePostObject, responseStream.Bytes())

	packet, _ := nex.NewPacketV1(client, nil)
	packet.SetVersion(1)
	packet.SetSource(0xA1)
	packet.SetDestination(0xAF)
	packet.SetType(nex.DataPacket)
	packet.SetPayload(response.Bytes())
	packet.AddFlag(nex.FlagNeedsAck)
	packet.AddFlag(nex.FlagReliable)
	globals.NEXServer.Send(packet)
}
