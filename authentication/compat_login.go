package main

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	nex "github.com/PretendoNetwork/nex-go"
	nexAuthentication "github.com/PretendoNetwork/nex-protocols-go/authentication"
)

func respondToLogin(err error, client *nex.Client, callID uint32, username string, token string, methodID uint32, secureStationURL *nex.StationURL) {
	response := nex.NewRMCResponse(nexAuthentication.ProtocolID, callID)
	if err != nil {
		response.SetError(nex.Errors.Core.Unknown)
		sendLoginResponse(client, response)
		return
	}

	username = strings.TrimRight(username, "\x00")
	var userPID uint32
	var userPassword string
	if username == "guest" {
		userPID = 100
		userPassword = "MMQea3n!fsik"
	} else {
		parsedPID, parseErr := strconv.ParseUint(username, 10, 32)
		if parseErr != nil {
			response.SetError(nex.Errors.RendezVous.InvalidUsername)
			sendLoginResponse(client, response)
			return
		}
		userPID = uint32(parsedPID)

		// A signed proxy token means the console holds the password of that
		// proxy account, even though it logs in with its own stored PID.
		passwordPID := userPID
		if proxyPID, ok := pidFromNEXToken(strings.TrimRight(token, "\x00")); ok {
			if !bindConsolePID(userPID, proxyPID) {
				logger.Error(fmt.Sprintf("Console PID %d is bound to a different proxy account than %d", userPID, proxyPID))
				response.SetError(nex.Errors.RendezVous.InvalidUsername)
				sendLoginResponse(client, response)
				return
			}
			passwordPID = proxyPID
		}

		var errorCode uint32
		userPassword, errorCode = passwordFromPID(passwordPID)
		if errorCode != 0 {
			response.SetError(errorCode)
			sendLoginResponse(client, response)
			return
		}
	}

	ticket, ticketErr := createLoginTicket(userPID, userPassword)
	if ticketErr != nil {
		response.SetError(nex.Errors.Core.Unknown)
		sendLoginResponse(client, response)
		return
	}

	rememberLogin(client, userPID, userPassword)

	connectionData := nex.NewRVConnectionData()
	connectionData.SetStationURL(secureStationURL.EncodeToString())
	connectionData.SetSpecialProtocols([]byte{})
	connectionData.SetStationURLSpecialProtocols("prudp:/")
	connectionData.SetTime(nex.NewDateTime(0).UTC())

	responseBody := nex.NewStreamOut(nexServer)
	responseBody.WriteResult(nex.NewResultSuccess(nex.Errors.Core.Unknown))
	responseBody.WriteUInt32LE(userPID)
	responseBody.WriteBuffer(ticket)
	responseBody.WriteStructure(connectionData)
	responseBody.WriteString("Badge Arcade Auth")
	response.SetSuccess(methodID, responseBody.Bytes())
	sendLoginResponse(client, response)
}

func createLoginTicket(userPID uint32, userPassword string) ([]byte, error) {
	const targetPID uint32 = 2

	sessionKey := make([]byte, nexServer.KerberosKeySize())
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, err
	}

	userKey := nex.DeriveKerberosKey(userPID, []byte(userPassword))
	targetKey := nex.DeriveKerberosKey(targetPID, []byte(nexServer.KerberosPassword()))

	ticketData := nex.NewKerberosTicketInternalData()
	ticketData.SetTimestamp(nex.NewDateTime(nex.NewDateTime(0).Now()))
	ticketData.SetUserPID(userPID)
	ticketData.SetSessionKey(sessionKey)
	encryptedTicketData := ticketData.Encrypt(targetKey, nex.NewStreamOut(nexServer))

	ticket := nex.NewKerberosTicket()
	ticket.SetSessionKey(sessionKey)
	ticket.SetTargetPID(targetPID)
	ticket.SetInternalData(encryptedTicketData)
	return ticket.Encrypt(userKey, nex.NewStreamOut(nexServer)), nil
}

func sendLoginResponse(client *nex.Client, response nex.RMCResponse) {
	packet, _ := nex.NewPacketV1(client, nil)
	packet.SetVersion(1)
	packet.SetSource(0xA1)
	packet.SetDestination(0xAF)
	packet.SetType(nex.DataPacket)
	packet.SetPayload(response.Bytes())
	packet.AddFlag(nex.FlagNeedsAck)
	packet.AddFlag(nex.FlagReliable)
	nexServer.Send(packet)
}

// A 3DS asks for further tickets on the connection it logged in with, and
// can only decrypt them with the password used for its login ticket. That
// password may belong to its proxy account rather than its login PID, so
// remember it per connection.
type loginCredentials struct {
	pid      uint32
	password string
	loggedIn time.Time
}

var (
	loginCredentialsMutex    sync.Mutex
	loginCredentialsByClient = map[*nex.Client]loginCredentials{}
)

func rememberLogin(client *nex.Client, pid uint32, password string) {
	loginCredentialsMutex.Lock()
	defer loginCredentialsMutex.Unlock()

	for knownClient, credentials := range loginCredentialsByClient {
		if time.Since(credentials.loggedIn) > nexTokenMaxAge {
			delete(loginCredentialsByClient, knownClient)
		}
	}
	loginCredentialsByClient[client] = loginCredentials{pid: pid, password: password, loggedIn: time.Now()}
}

func respondToRequestTicket(err error, client *nex.Client, callID uint32, userPID uint32, targetPID uint32) {
	response := nex.NewRMCResponse(nexAuthentication.ProtocolID, callID)

	loginCredentialsMutex.Lock()
	credentials, ok := loginCredentialsByClient[client]
	loginCredentialsMutex.Unlock()

	if err != nil || !ok || credentials.pid != userPID || targetPID != 2 {
		response.SetError(nex.Errors.Core.AccessDenied)
		sendLoginResponse(client, response)
		return
	}

	ticket, ticketErr := createLoginTicket(userPID, credentials.password)
	if ticketErr != nil {
		response.SetError(nex.Errors.Core.Unknown)
		sendLoginResponse(client, response)
		return
	}

	responseBody := nex.NewStreamOut(nexServer)
	responseBody.WriteResult(nex.NewResultSuccess(nex.Errors.Core.Unknown))
	responseBody.WriteBuffer(ticket)
	response.SetSuccess(nexAuthentication.MethodRequestTicket, responseBody.Bytes())
	sendLoginResponse(client, response)
}
