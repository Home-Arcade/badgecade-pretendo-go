package main

import (
	"fmt"
	"os"

	nex "github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-common-go/authentication"
	nexAuthentication "github.com/PretendoNetwork/nex-protocols-go/authentication"
)

var nexServer *nex.Server

func main() {
	nexServer = nex.NewServer()
	nexServer.SetPRUDPVersion(1)
	nexServer.SetPRUDPProtocolMinorVersion(3)
	nexServer.SetDefaultNEXVersion(&nex.NEXVersion{
		Major: 3,
		Minor: 7,
		Patch: 16,
	})
	nexServer.SetKerberosPassword(os.Getenv("KERBEROS_PASSWORD"))
	nexServer.SetAccessKey("82d5962d")

	nexServer.On("Data", func(packet *nex.PacketV1) {
		request := packet.RMCRequest()

		fmt.Println("==Badge Arcade - Auth==")
		fmt.Printf("Protocol ID: %#v\n", request.ProtocolID())
		fmt.Printf("Method ID: %#v\n", request.MethodID())
		fmt.Println("====================")
	})

	authenticationProtocol := authentication.NewCommonAuthenticationProtocol(nexServer)

	secureStationURL := nex.NewStationURL("")
	secureStationURL.SetScheme("prudps")
	secureStationURL.SetAddress(os.Getenv("SECURE_SERVER_LOCATION"))
	secureStationURL.SetPort(os.Getenv("SECURE_SERVER_PORT"))
	secureStationURL.SetCID("1")
	secureStationURL.SetPID("2")
	secureStationURL.SetSID("1")
	secureStationURL.SetStream("10")
	secureStationURL.SetType("2")

	authenticationProtocol.SetSecureStationURL(secureStationURL)
	authenticationProtocol.SetBuildName("Badge Arcade Auth")
	authenticationProtocol.SetPasswordFromPIDFunction(passwordFromPID)
	authenticationProtocol.Login(func(err error, client *nex.Client, callID uint32, username string) {
		respondToLogin(err, client, callID, username, "", nexAuthentication.MethodLogin, secureStationURL)
	})
	authenticationProtocol.LoginEx(func(err error, client *nex.Client, callID uint32, username string, authenticationInfo *nexAuthentication.AuthenticationInfo) {
		token := ""
		if authenticationInfo != nil {
			token = authenticationInfo.Token
		}
		respondToLogin(err, client, callID, username, token, nexAuthentication.MethodLoginEx, secureStationURL)
	})

	authenticationProtocol.RequestTicket(respondToRequestTicket)

	nexServer.Listen(":59400")
}
