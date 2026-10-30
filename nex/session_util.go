package nex
import (
	"time"

	pqextended "github.com/PretendoNetwork/pq-extended"
	"github.com/SamtendoNetwork/BetterWSC/globals"
)

func startSessionCleanup(interval, emptyMaxAge time.Duration) {
	manager := globals.MatchmakingManager

	manager.Mutex.Lock()
	manager.Database.Exec(`UPDATE matchmaking.gatherings SET registered=false WHERE registered=true`)
	manager.Mutex.Unlock()

	go func() {
		for range time.Tick(interval) {
			cleanupSessions(emptyMaxAge)
		}
	}()
}

// Used when the server starts up do the db is fresh for new sessions instantly
func ClearSessions() {
	manager := globals.MatchmakingManager

	manager.Mutex.Lock()
	manager.Database.Exec(`TRUNCATE TABLE matchmaking.matchmake_sessions, matchmaking.gatherings CASCADE`)
	manager.Mutex.Unlock()
}

func cleanupSessions(emptyMaxAge time.Duration) {
	manager := globals.MatchmakingManager

	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()
	// TODO: Do we need any more conditions to clear?
	// Gets rid of sessions that have been open for too long or are just bad to prevent matchmaking issues
	manager.Database.Exec(`
		UPDATE matchmaking.gatherings
		SET registered=false
		WHERE registered=true
		  AND COALESCE(array_length(participants, 1), 0) = 0
		  AND started_time < $1`,
		time.Now().Add(-emptyMaxAge),
	)

	rows, err := manager.Database.Query(`
		SELECT id, participants
		FROM matchmaking.gatherings
		WHERE registered=true
		  AND COALESCE(array_length(participants, 1), 0) > 0`)
	if err != nil {
		return
	}

	var stale []uint32
	for rows.Next() {
		var id uint32
		var pids []uint64
		if err := rows.Scan(&id, pqextended.Array(&pids)); err != nil {
			continue
		}

		connected := false
		for _, pid := range pids {
			if globals.SecureEndpoint.FindConnectionByPID(pid) != nil {
				connected = true
				break
			}
		}
		if !connected {
			stale = append(stale, id)
		}
	}
	rows.Close()

	for _, id := range stale {
		manager.Database.Exec(`UPDATE matchmaking.gatherings SET registered=false WHERE id=$1`, id)
	}
}
