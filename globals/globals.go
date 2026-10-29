package globals

import (
	"database/sql"

	pbfriends "github.com/PretendoNetwork/grpc/go/friends"
	"github.com/PretendoNetwork/nex-go/v2"
	common_globals "github.com/Whatsauser/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/plogger-go"
	"os"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

)
var Postgres *sql.DB
var MatchmakingManager *common_globals.MatchmakingManager
var Logger *plogger.Logger
var KerberosPassword = os.Getenv("SN_WSC_KERBEROS_PASSWORD")

var AuthenticationServerPort int
var SecureServerHost string
var SecureServerPort int

var AuthenticationServer *nex.PRUDPServer
var AuthenticationEndpoint *nex.PRUDPEndPoint

var SecureServer *nex.PRUDPServer
var SecureEndpoint *nex.PRUDPEndPoint

var GRPCFriendsClientConnection *grpc.ClientConn
var GRPCFriendsClient pbfriends.FriendsClient
var GRPCFriendsCommonMetadata metadata.MD
