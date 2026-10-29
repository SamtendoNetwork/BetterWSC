package nex

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"github.com/PretendoNetwork/nex-go/v2"
	common_globals "github.com/Whatsauser/nex-protocols-common-go/v2/globals"
	"github.com/SamtendoNetwork/BetterWSC/globals"
	"github.com/SamtendoNetwork/BetterWSC/datastore"
)

func StartSecureServer() {
	globals.SecureServer = nex.NewPRUDPServer()

	globals.SecureEndpoint = nex.NewPRUDPEndPoint(1)
	globals.SecureEndpoint.IsSecureEndPoint = true
	globals.SecureEndpoint.ServerAccount = globals.SecureServerAccount
	globals.SecureEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	globals.SecureEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername

	globals.SecureServer.BindPRUDPEndPoint(globals.SecureEndpoint)

	globals.SecureServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(3, 4, 7))
	globals.SecureServer.AccessKey = "4d324052"

	globals.SecureEndpoint.OnData(func(packet nex.PacketInterface) {
		request := packet.RMCMessage()
		protocolName, methodName := globals.RMCNames(request.ProtocolID, request.MethodID)

		fmt.Printf("== Wii Sports Club - Secure ==\n")
		fmt.Printf("User: %d\n", packet.Sender().PID())
		fmt.Printf("Protocol: %d (%s)\n", request.ProtocolID, protocolName)
		fmt.Printf("Method: %d (%s)\n", request.MethodID, methodName)
		if !strings.EqualFold(os.Getenv("SN_WSC_DEBUG_MODE"), "true") {
			fmt.Printf("Paramaters: %s\n", hex.EncodeToString(request.Parameters))
		}
		fmt.Println("===============")
	})

	globals.SecureEndpoint.OnError(func(err *nex.Error) {
		globals.Logger.Errorf("Secure: %v", err)
	})
	globals.MatchmakingManager = common_globals.NewMatchmakingManager(globals.SecureEndpoint, globals.Postgres)
	globals.MatchmakingManager.GetUserFriendPIDs = globals.GetUserFriendPIDs

	registerCommonSecureServerProtocols()
	database.InitDataStore()
	globals.SecureServer.Listen(globals.SecureServerPort)
}
