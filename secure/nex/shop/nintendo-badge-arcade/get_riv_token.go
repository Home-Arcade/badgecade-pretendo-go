package nex_shop_nintendo_badge_arcade

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"

	"github.com/PretendoNetwork/nex-go"
	nexproto "github.com/PretendoNetwork/nex-protocols-go/shop/nintendo-badge-arcade"
)

func GetRivToken(err error, client *nex.Client, callID uint32, itemCode string, referenceID []byte) {
	if err != nil {
		globals.Logger.Error(err.Error())
	}
	globals.Logger.Info(fmt.Sprintf("[Shop] GetRivToken pid=%d item=%q reference=%x", client.PID(), itemCode, referenceID))

	// Plays come from the fake eShop for free, so any token will do. The real
	// format is unknown; send something non-empty in case the game checks.
	token := make([]byte, 16)
	_, _ = rand.Read(token)
	rmcResponseStream := nex.NewStreamOut(globals.NEXServer)
	rmcResponseStream.WriteString(hex.EncodeToString(token))

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCResponse(nexproto.ProtocolID, callID)
	rmcResponse.SetSuccess(nexproto.MethodGetRivToken, rmcResponseBody)
	rmcResponse.SetCustomID(nexproto.CustomProtocolID)

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
