
package database

import (
	"time"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
	"github.com/lib/pq"

	"github.com/SamtendoNetwork/BetterWSC/globals"
)

// InitializeObjectByPreparePostParam creates a new, not yet uploaded, object
func InitializeObjectByPreparePostParam(ownerPID types.PID, param datastore_types.DataStorePreparePostParam) (uint64, *nex.Error) {
	now := time.Now().UTC()

	tx, err := globals.Postgres.Begin()
	if err != nil {
		return 0, unknownError(err)
	}
	defer tx.Rollback()

	slot := int64(param.PersistenceInitParam.PersistenceSlotID)

	if bool(param.PersistenceInitParam.DeleteLastObject) && slot != persistenceSlotNone {
		_, err = tx.Exec(`
			UPDATE datastore.objects
			SET deleted = true, update_date = $4
			WHERE owner = $1 AND persistence_slot_id = $2 AND game = $3 AND deleted = false`,
			int64(ownerPID),
			slot,
			"SN_WSC",
			now,
		)
		if err != nil {
			return 0, unknownError(err)
		}
	}

	var dataID int64

	err = tx.QueryRow(`
		INSERT INTO datastore.objects (
			game,
			upload_completed,
			deleted,
			owner,
			size,
			name,
			data_type,
			meta_binary,
			permission,
			permission_recipients,
			delete_permission,
			delete_permission_recipients,
			flag,
			period,
			refer_data_id,
			tags,
			persistence_slot_id,
			extra_data,
			access_password,
			update_password,
			creation_date,
			update_date
		) VALUES (
			$1, false, false, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, 0, 0, $17, $17
		) RETURNING data_id`,
		"SN_WSC",
		int64(ownerPID),
		int64(param.Size),
		string(param.Name),
		int64(param.DataType),
		[]byte(param.MetaBinary),
		int64(param.Permission.Permission),
		pq.Array(fromPIDList(param.Permission.RecipientIDs)),
		int64(param.DelPermission.Permission),
		pq.Array(fromPIDList(param.DelPermission.RecipientIDs)),
		int64(param.Flag),
		int64(param.Period),
		int64(param.ReferDataID),
		pq.Array(fromStringList(param.Tags)),
		slot,
		pq.Array(fromStringList(param.ExtraData)),
		now,
	).Scan(&dataID)
	if err != nil {
		return 0, unknownError(err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, unknownError(err)
	}

	return uint64(dataID), nil
}