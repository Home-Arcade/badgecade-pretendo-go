package nex

import (
	"github.com/PretendoNetwork/nex-go"
	datastore_nintendo_badge_arcade "github.com/PretendoNetwork/nex-protocols-go/datastore/nintendo-badge-arcade"
	nexprotoglobals "github.com/PretendoNetwork/nex-protocols-go/globals"
	secure_connection_nintendo_badge_arcade "github.com/PretendoNetwork/nex-protocols-go/secure-connection/nintendo-badge-arcade"
	shop_nintendo_badge_arcade "github.com/PretendoNetwork/nex-protocols-go/shop/nintendo-badge-arcade"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"
	nex_datastore "github.com/PretendoNetwork/nintendo-badge-arcade-secure/nex/datastore"
	nex_datastore_nintendo_badge_arcade "github.com/PretendoNetwork/nintendo-badge-arcade-secure/nex/datastore/nintendo-badge-arcade"
	nex_secure_connection "github.com/PretendoNetwork/nintendo-badge-arcade-secure/nex/secure-connection"
	nex_secure_connection_nintendo_badge_arcade "github.com/PretendoNetwork/nintendo-badge-arcade-secure/nex/secure-connection/nintendo-badge-arcade"
	nex_shop_nintendo_badge_arcade "github.com/PretendoNetwork/nintendo-badge-arcade-secure/nex/shop/nintendo-badge-arcade"
)

func registerNEXProtocols() {

	secureConnectionProtocol := secure_connection_nintendo_badge_arcade.NewSecureConnectionNintendoBadgeArcadeProtocol(globals.NEXServer)

	secureConnectionProtocol.Register(nex_secure_connection.Register)
	secureConnectionProtocol.GetMaintenanceStatus(nex_secure_connection_nintendo_badge_arcade.GetMaintenanceStatus)

	dataStoreNintendoBadgeArcadeProtocol := datastore_nintendo_badge_arcade.NewDataStoreNintendoBadgeArcadeProtocol(globals.NEXServer)

	dataStoreNintendoBadgeArcadeProtocol.GetPersistenceInfo(nex_datastore.GetPersistenceInfo)
	dataStoreNintendoBadgeArcadeProtocol.PostMetaBinary(nex_datastore.PostMetaBinary)
	dataStoreNintendoBadgeArcadeProtocol.PreparePostObject(nex_datastore.PreparePostObject)
	dataStoreNintendoBadgeArcadeProtocol.CompletePostObject(nex_datastore.CompletePostObject)
	dataStoreNintendoBadgeArcadeProtocol.PrepareGetObject(nex_datastore.PrepareGetObject)
	dataStoreNintendoBadgeArcadeProtocol.GetMetaByOwnerID(nex_datastore_nintendo_badge_arcade.GetMetaByOwnerID)
	dataStoreNintendoBadgeArcadeProtocol.ChangeMeta(nex_datastore.ChangeMeta)
	dataStoreNintendoBadgeArcadeProtocol.PrepareUpdateObject(nex_datastore.PrepareUpdateObject)
	dataStoreNintendoBadgeArcadeProtocol.CompleteUpdateObject(nex_datastore.CompleteUpdateObject)

	// nex-protocols-go v1.0.25 never routes GetRivToken (method 1), so the
	// game got NotImplemented (006-0103) when buying plays. Build the protocol
	// without its Setup and route both methods here instead.
	shopNintendoBadgeArcadePrococol := &shop_nintendo_badge_arcade.ShopNintendoBadgeArcadeProtocol{Server: globals.NEXServer}
	shopNintendoBadgeArcadePrococol.ShopProtocol.Server = globals.NEXServer

	shopNintendoBadgeArcadePrococol.PostPlayLog(nex_shop_nintendo_badge_arcade.PostPlayLog)
	shopNintendoBadgeArcadePrococol.GetRivToken(nex_shop_nintendo_badge_arcade.GetRivToken)

	globals.NEXServer.On("Data", func(packet nex.PacketInterface) {
		request := packet.RMCRequest()
		if request.ProtocolID() != shop_nintendo_badge_arcade.ProtocolID || request.CustomID() != shop_nintendo_badge_arcade.CustomProtocolID {
			return
		}
		switch request.MethodID() {
		case shop_nintendo_badge_arcade.MethodGetRivToken:
			go shopNintendoBadgeArcadePrococol.HandleGetRivToken(packet)
		case shop_nintendo_badge_arcade.MethodPostPlayLog:
			go shopNintendoBadgeArcadePrococol.HandlePostPlayLog(packet)
		default:
			go nexprotoglobals.RespondNotImplementedCustom(packet, shop_nintendo_badge_arcade.CustomProtocolID)
		}
	})
}
