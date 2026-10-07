package nex_datastore

import (
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/database"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"
	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-go/datastore"
)

func PrepareUpdateObject(err error, client *nex.Client, callID uint32, param *datastore.DataStorePrepareUpdateParam) {
	dataID := uint32(param.DataID)
	version := database.GetVersionByDataID(dataID) + 1
	key := objectKey(dataID, version)
	postURL, formFields, storeErr := makeObjectUploadInfo(key)
	if storeErr != nil {
		globals.Logger.Error(storeErr.Error())
	}

	updateInfo := datastore.NewDataStoreReqUpdateInfo()
	updateInfo.Version = version
	updateInfo.Url = postURL
	updateInfo.RequestHeaders = []*datastore.DataStoreKeyValue{}
	updateInfo.FormFields = formFields
	updateInfo.RootCaCert = []byte{}

	responseStream := nex.NewStreamOut(globals.NEXServer)
	responseStream.WriteStructure(updateInfo)
	response := nex.NewRMCResponse(datastore.ProtocolID, callID)
	response.SetSuccess(datastore.MethodPrepareUpdateObject, responseStream.Bytes())

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
