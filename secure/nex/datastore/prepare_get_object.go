package nex_datastore

import (
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/database"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"
	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-go/datastore"
)

func PrepareGetObject(err error, client *nex.Client, callID uint32, param *datastore.DataStorePrepareGetParam) {
	dataID := uint32(param.DataID)
	version := database.GetVersionByDataID(dataID)
	key := objectKey(dataID, version)
	objectURL, dataSize, storeErr := makeObjectDownloadURL(key)
	if storeErr != nil {
		globals.Logger.Error(storeErr.Error())
	}

	getInfo := datastore.NewDataStoreReqGetInfo()
	getInfo.URL = objectURL
	getInfo.RequestHeaders = []*datastore.DataStoreKeyValue{}
	getInfo.Size = uint32(dataSize)
	getInfo.RootCA = []byte{}
	getInfo.DataID = param.DataID

	responseStream := nex.NewStreamOut(globals.NEXServer)
	responseStream.WriteStructure(getInfo)
	response := nex.NewRMCResponse(datastore.ProtocolID, callID)
	response.SetSuccess(datastore.MethodPrepareGetObject, responseStream.Bytes())

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
