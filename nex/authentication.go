package nex

import (
	"encoding/hex"
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2"

	"github.com/SamtendoNetwork/BetterWSC/globals"
)

func StartAuthenticationServer() {
	globals.AuthenticationServer = nex.NewPRUDPServer()

	globals.AuthenticationEndpoint = nex.NewPRUDPEndPoint(1)
	globals.AuthenticationEndpoint.ServerAccount = globals.AuthenticationServerAccount
	globals.AuthenticationEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	globals.AuthenticationEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername

	globals.AuthenticationServer.BindPRUDPEndPoint(globals.AuthenticationEndpoint)

	globals.AuthenticationServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(3, 4, 7))
	globals.AuthenticationServer.AccessKey = "4d324052"

	globals.AuthenticationEndpoint.OnData(func(packet nex.PacketInterface) {
		request := packet.RMCMessage()
		protocolName, methodName := globals.RMCNames(request.ProtocolID, request.MethodID)

		fmt.Printf("== Wii Sports Club - Auth ==\n")
		fmt.Printf("User: %d\n", packet.Sender().PID())
		fmt.Printf("Protocol: %d (%s)\n", request.ProtocolID, protocolName)
		fmt.Printf("Method: %d (%s)\n", request.MethodID, methodName)
		fmt.Printf("Paramaters: %s\n", hex.EncodeToString(request.Parameters))
		fmt.Println("===============")
	})

	globals.AuthenticationEndpoint.OnError(func(err *nex.Error) {
		globals.Logger.Errorf("Auth: %v", err)
	})

	registerCommonAuthenticationServerProtocols()

	globals.AuthenticationServer.Listen(globals.AuthenticationServerPort)
}
