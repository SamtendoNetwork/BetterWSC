package nex

import (
	"slices"
	"time"

	commonmmdb "github.com/Whatsauser/nex-protocols-common-go/v2/match-making/database"
	commonmatchmaking "github.com/Whatsauser/nex-protocols-common-go/v2/match-making"
	commonmatchmakingext "github.com/Whatsauser/nex-protocols-common-go/v2/match-making-ext"
	commonmatchmakeextension "github.com/Whatsauser/nex-protocols-common-go/v2/matchmake-extension"
	commonnattraversal "github.com/Whatsauser/nex-protocols-common-go/v2/nat-traversal"
	commonsecure "github.com/Whatsauser/nex-protocols-common-go/v2/secure-connection"
	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/nex-go/v2"
	common_globals "github.com/Whatsauser/nex-protocols-common-go/v2/globals"
	commonutility "github.com/Whatsauser/nex-protocols-common-go/v2/utility"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
	commondatastore "github.com/Whatsauser/nex-protocols-common-go/v2/datastore"
	matchmaking "github.com/PretendoNetwork/nex-protocols-go/v2/match-making"
	matchmakingext "github.com/PretendoNetwork/nex-protocols-go/v2/match-making-ext"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	nattraversal "github.com/PretendoNetwork/nex-protocols-go/v2/nat-traversal"
	match_making_types "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	secure "github.com/PretendoNetwork/nex-protocols-go/v2/secure-connection"
	serviceitem "github.com/PretendoNetwork/nex-protocols-go/v2/service-item/wii-sports-club"
	serviceitemtypes "github.com/PretendoNetwork/nex-protocols-go/v2/service-item/wii-sports-club/types"
	utility "github.com/PretendoNetwork/nex-protocols-go/v2/utility"
	"github.com/SamtendoNetwork/BetterWSC/globals"
	database "github.com/SamtendoNetwork/BetterWSC/datastore"
)

func registerCommonSecureServerProtocols() {

	// Secure
	secureProtocol := secure.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(secureProtocol)
	commonSecureProtocol := commonsecure.NewCommonProtocol(secureProtocol)
	commonSecureProtocol.EnableInsecureRegister()
	commonSecureProtocol.CreateReportDBRecord = func(_ types.PID, _ types.UInt32, _ types.QBuffer) error {
		return nil
	}
	globals.SecureEndpoint.OnConnectionEnded(func(connection *nex.PRUDPConnection) {
		go func() {
			time.Sleep(200 * time.Millisecond)
			globals.MatchmakingManager.Database.Exec(`
				UPDATE matchmaking.matchmake_sessions ms
				SET open_participation = true
				FROM matchmaking.gatherings g
				WHERE ms.id = g.id
				AND g.registered = true
				AND ms.open_participation = false
				AND COALESCE(array_length(g.participants, 1), 0) < g.max_participants`)
		}()
	})

	// NAT
	natTraversalProtocol := nattraversal.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(natTraversalProtocol)
	commonnattraversal.NewCommonProtocol(natTraversalProtocol)

	// Match Making protocols
	matchMakingProtocol := matchmaking.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMakingProtocol)
	commonmatchmaking.NewCommonProtocol(matchMakingProtocol).SetManager(globals.MatchmakingManager)
	matchMakingProtocol.SetHandlerGetSessionURLs(func(_ error, packet nex.PacketInterface, callID uint32, gid types.UInt32) (*nex.RMCMessage, *nex.Error) {
	endpoint := packet.Sender().Endpoint()
	requester := uint64(packet.Sender().PID())

	globals.MatchmakingManager.Mutex.RLock()
	_, _, participants, _, nexErr := commonmmdb.FindGatheringByID(globals.MatchmakingManager, uint32(gid))
	globals.MatchmakingManager.Mutex.RUnlock()
	if nexErr != nil {
		return nil, nexErr
	}

	if !slices.Contains(participants, requester) {
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "not a participant of this gathering")
	}

	urls := types.NewList[types.StationURL]()
	for _, pid := range participants {
		if pid == requester {
			continue
		}
		conn := globals.SecureEndpoint.FindConnectionByPID(pid)
		if conn == nil || len(conn.StationURLs) == 0 {
			continue
		}
		latest := conn.StationURLs[len(conn.StationURLs)-1]
		urls = append(urls, latest.Copy().(types.StationURL))
	}

	stream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	urls.WriteTo(stream)

	response := nex.NewRMCSuccess(endpoint, stream.Bytes())
	response.ProtocolID = matchmaking.ProtocolID
	response.MethodID = matchmaking.MethodGetSessionURLs
	response.CallID = callID
	return response, nil
	})

	// ext
	matchMakingExtProtocol := matchmakingext.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchMakingExtProtocol)
	commonmatchmakingext.NewCommonProtocol(matchMakingExtProtocol).SetManager(globals.MatchmakingManager)

	// extension
	matchmakeExtensionProtocol := matchmakeextension.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(matchmakeExtensionProtocol)

	mmExtCommon := commonmatchmakeextension.NewCommonProtocol(matchmakeExtensionProtocol)
	mmExtCommon.SetManager(globals.MatchmakingManager)
	mmExtCommon.CleanupMatchmakeSessionSearchCriterias = func(searchCriterias types.List[match_making_types.MatchmakeSessionSearchCriteria]) {}
	mmExtCommon.CleanupSearchMatchmakeSession = func(matchmakeSession *match_making_types.MatchmakeSession) {}
	mmExtCommon.OnAfterCloseParticipation = func(packet nex.PacketInterface, gid types.UInt32) {
		connection := packet.Sender().(*nex.PRUDPConnection)

		globals.MatchmakingManager.Mutex.Lock()
		defer globals.MatchmakingManager.Mutex.Unlock()

		if nexErr := commonmmdb.UnregisterGathering(globals.MatchmakingManager, connection.PID(), uint32(gid)); nexErr != nil {
			common_globals.Logger.Error(nexErr.Error())
		}
	}

	// Utility
	utilityProtocol := utility.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(utilityProtocol)
	commonUtilityProtocol := commonutility.NewCommonProtocol(utilityProtocol)
	commonUtilityProtocol.SetManager(common_globals.NewUtilityManager(globals.SecureEndpoint, globals.Postgres))

	// Service Item
	serviceItemProtocol := serviceitem.NewProtocol(globals.SecureEndpoint)
	serviceItemProtocol.GetServiceItemRightRequest = func(_ error, _ nex.PacketInterface, callID uint32, _ serviceitemtypes.ServiceItemGetServiceItemRightParam) (*nex.RMCMessage, *nex.Error) {
		response := serviceitemtypes.NewServiceItemGetServiceItemRightResponse()

		stream := nex.NewByteStreamOut(globals.SecureEndpoint.LibraryVersions(), globals.SecureEndpoint.ByteStreamSettings())
		response.WriteTo(stream)

		rmcResponse := nex.NewRMCSuccess(globals.SecureEndpoint, stream.Bytes())
		rmcResponse.ProtocolID = serviceitem.ProtocolID
		rmcResponse.MethodID = serviceitem.MethodGetServiceItemRightRequest
		rmcResponse.CallID = callID

		return rmcResponse, nil
	}
	globals.SecureEndpoint.RegisterServiceProtocol(serviceItemProtocol)

	// Datastore
	dataStoreProtocol := datastore.NewProtocol()
	globals.SecureEndpoint.RegisterServiceProtocol(dataStoreProtocol)
	commonDataStoreProtocol := commondatastore.NewCommonProtocol(dataStoreProtocol)
	commonDataStoreProtocol.GetObjectInfosByDataStoreSearchParam = database.GetObjectInfosByDataStoreSearchParam
	commonDataStoreProtocol.InitializeObjectByPreparePostParam = database.InitializeObjectByPreparePostParam
	globals.SecureEndpoint.RegisterServiceProtocol(dataStoreProtocol)

	// Cleans up old/dead sessions from db every 10 sec and checks for empty sessions that were empty for longer than a minute
	startSessionCleanup(10*time.Second, 1*time.Minute) 
}
