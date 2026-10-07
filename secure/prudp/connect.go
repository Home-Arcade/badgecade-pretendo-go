package prudp

import (
	"encoding/binary"
	"fmt"

	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/globals"

	"github.com/PretendoNetwork/nex-go"
)

func Connect(packet nex.PacketInterface) {
	ticketData, requestData, ok := readConnectBuffers(packet.Payload())
	if !ok || len(ticketData) < 16 || len(requestData) < 16 {
		globals.Logger.Error("[Secure] Invalid PRUDP CONNECT payload")
		return
	}

	serverKey := nex.DeriveKerberosKey(2, []byte(globals.NEXServer.KerberosPassword()))
	if !nex.NewKerberosEncryption(serverKey).Validate(ticketData) {
		globals.Logger.Error("[Secure] CONNECT ticket HMAC validation failed")
		return
	}

	ticket := nex.NewKerberosTicketInternalData()
	ticket.Decrypt(nex.NewStreamIn(ticketData, globals.NEXServer), serverKey)

	// TODO: Check timestamp here

	sessionKey := ticket.SessionKey()
	if len(sessionKey) != globals.NEXServer.KerberosKeySize() {
		globals.Logger.Error("[Secure] CONNECT ticket has an invalid session key")
		return
	}
	kerberos := nex.NewKerberosEncryption(sessionKey)
	if !kerberos.Validate(requestData) {
		globals.Logger.Error("[Secure] CONNECT request HMAC validation failed")
		return
	}

	decryptedRequestData := kerberos.Decrypt(requestData)
	if len(decryptedRequestData) != 12 {
		globals.Logger.Error("[Secure] CONNECT request has an invalid length")
		return
	}
	checkDataStream := nex.NewStreamIn(decryptedRequestData, globals.NEXServer)

	userPID := checkDataStream.ReadUInt32LE()
	if userPID != ticket.UserPID() {
		globals.Logger.Error("[Secure] CONNECT request PID does not match ticket PID")
		return
	}
	_ = checkDataStream.ReadUInt32LE() //CID of secure server station url
	responseCheck := checkDataStream.ReadUInt32LE()

	responseValueStream := nex.NewStreamOut(globals.NEXServer)
	responseValueStream.WriteUInt32LE(responseCheck + 1)

	responseValueBufferStream := nex.NewStreamOut(globals.NEXServer)
	responseValueBufferStream.WriteBuffer(responseValueStream.Bytes())

	globals.NEXServer.AcknowledgePacket(packet, responseValueBufferStream.Bytes())

	packet.Sender().UpdateRC4Key(sessionKey)
	packet.Sender().SetSessionKey(sessionKey)

	packet.Sender().SetPID(userPID)
	globals.Logger.Success(fmt.Sprintf("[Secure] PRUDP connection established for PID %d", userPID))
}

func readConnectBuffers(payload []byte) ([]byte, []byte, bool) {
	readBuffer := func(data []byte) ([]byte, []byte, bool) {
		if len(data) < 4 {
			return nil, nil, false
		}

		length := binary.LittleEndian.Uint32(data[:4])
		if uint64(length) > uint64(len(data)-4) {
			return nil, nil, false
		}

		end := 4 + int(length)
		return data[4:end], data[end:], true
	}

	ticketData, remaining, ok := readBuffer(payload)
	if !ok {
		return nil, nil, false
	}
	requestData, remaining, ok := readBuffer(remaining)
	if !ok || len(remaining) != 0 {
		return nil, nil, false
	}

	return ticketData, requestData, true
}
