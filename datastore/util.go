package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore_constants "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/constants"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
	"github.com/lib/pq"

	"github.com/SamtendoNetwork/BetterWSC/globals"
)

const (
	dataTypeAny         = 0xFFFF
	persistenceSlotNone = 0xFFFF
)

const objectColumns = `
	o.data_id,
	o.upload_completed,
	o.deleted,
	COALESCE(o.owner, 0),
	COALESCE(o.size, 0),
	COALESCE(o.name, ''),
	COALESCE(o.data_type, 0),
	o.meta_binary,
	COALESCE(o.permission, 0),
	o.permission_recipients,
	COALESCE(o.delete_permission, 0),
	o.delete_permission_recipients,
	COALESCE(o.flag, 0),
	COALESCE(o.period, 0),
	COALESCE(o.refer_data_id, 0),
	o.tags,
	COALESCE(o.persistence_slot_id, 0),
	o.extra_data,
	COALESCE(o.creation_date, to_timestamp(0)),
	COALESCE(o.update_date, to_timestamp(0))`

// initialize sql
func InitDataStore() error {
	statements := []string{
		`CREATE SCHEMA IF NOT EXISTS datastore`,
		`CREATE SEQUENCE IF NOT EXISTS datastore.object_data_id_seq
			INCREMENT 1
			MINVALUE 1
			MAXVALUE 281474976710656
			START 1
			CACHE 1`,
		`CREATE TABLE IF NOT EXISTS datastore.objects (
			data_id bigint NOT NULL DEFAULT nextval('datastore.object_data_id_seq') PRIMARY KEY,
			game text NOT NULL DEFAULT '',
			upload_completed boolean NOT NULL DEFAULT FALSE,
			deleted boolean NOT NULL DEFAULT FALSE,
			owner bigint,
			size bigint,
			name text,
			data_type int,
			meta_binary bytea,
			permission int,
			permission_recipients bigint[],
			delete_permission int,
			delete_permission_recipients bigint[],
			flag bigint,
			period int,
			refer_data_id bigint,
			tags text[],
			persistence_slot_id int,
			extra_data text[],
			access_password bigint NOT NULL DEFAULT 0,
			update_password bigint NOT NULL DEFAULT 0,
			creation_date timestamp,
			update_date timestamp
		)`,
		`ALTER TABLE datastore.objects ADD COLUMN IF NOT EXISTS game text NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS objects_game_owner_slot_idx ON datastore.objects (game, owner, persistence_slot_id)`,
		`CREATE INDEX IF NOT EXISTS objects_tags_idx ON datastore.objects USING gin (tags)`,
		`CREATE TABLE IF NOT EXISTS datastore.object_ratings (
			data_id bigint NOT NULL,
			slot smallint NOT NULL,
			flag smallint NOT NULL DEFAULT 0,
			internal_flag smallint NOT NULL DEFAULT 0,
			lock_type smallint NOT NULL DEFAULT 0,
			initial_value bigint NOT NULL DEFAULT 0,
			range_min int NOT NULL DEFAULT 0,
			range_max int NOT NULL DEFAULT 0,
			period_hour smallint NOT NULL DEFAULT 0,
			period_duration smallint NOT NULL DEFAULT 0,
			total_value bigint NOT NULL DEFAULT 0,
			rating_count bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (data_id, slot)
		)`,
	}

	for _, statement := range statements {
		_, err := globals.Postgres.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func unknownError(err error) *nex.Error {
	globals.Logger.Error(err.Error())
	return nex.NewError(nex.ResultCodes.DataStore.Unknown, "database error")
}

func toPIDList(values []int64) types.List[types.PID] {
	list := types.NewList[types.PID]()

	for _, value := range values {
		list = append(list, types.PID(value))
	}

	return list
}

func toStringList(values []string) types.List[types.String] {
	list := types.NewList[types.String]()

	for _, value := range values {
		list = append(list, types.String(value))
	}

	return list
}

func fromPIDList(list types.List[types.PID]) []int64 {
	values := make([]int64, 0, len(list))

	for _, pid := range list {
		values = append(values, int64(pid))
	}

	return values
}

func fromStringList(list types.List[types.String]) []string {
	values := make([]string, 0, len(list))

	for _, value := range list {
		values = append(values, string(value))
	}

	return values
}

func searchOrderExpression(column datastore_constants.SearchSortColumn) string {
	switch column {
	case datastore_constants.SearchSortColumnDataID:
		return "o.data_id"
	case datastore_constants.SearchSortColumnSize:
		return "o.size"
	case datastore_constants.SearchSortColumnNameAlphabetical:
		return "o.name"
	case datastore_constants.SearchSortColumnDataType:
		return "o.data_type"
	case datastore_constants.SearchSortColumnCreatedTime:
		return "o.creation_date"
	case datastore_constants.SearchSortColumnUpdatedTime:
		return "o.update_date"
	}

	if column >= datastore_constants.SearchSortColumnRating0 && column <= datastore_constants.SearchSortColumnRating7 {
		slot := int(column - datastore_constants.SearchSortColumnRating0)

		return fmt.Sprintf(`COALESCE((SELECT r.total_value FROM datastore.object_ratings r WHERE r.data_id = o.data_id AND r.slot = %d), 0)`, slot)
	}

	return "o.update_date"
}

func scanObjects(rows *sql.Rows) ([]datastore_types.DataStoreMetaInfo, error) {
	defer rows.Close()

	objects := make([]datastore_types.DataStoreMetaInfo, 0)

	for rows.Next() {
		var dataID, owner, size, name_, dataType int64
		_ = name_
		var (
			uploadCompleted         bool
			deleted                 bool
			name                    string
			metaBinary              []byte
			permission              int64
			permissionRecipients    []int64
			delPermission           int64
			delPermissionRecipients []int64
			flag                    int64
			period                  int64
			referDataID             int64
			tags                    []string
			persistenceSlotID       int64
			extraData               []string
			created                 time.Time
			updated                 time.Time
		)

		err := rows.Scan(
			&dataID,
			&uploadCompleted,
			&deleted,
			&owner,
			&size,
			&name,
			&dataType,
			&metaBinary,
			&permission,
			pq.Array(&permissionRecipients),
			&delPermission,
			pq.Array(&delPermissionRecipients),
			&flag,
			&period,
			&referDataID,
			pq.Array(&tags),
			&persistenceSlotID,
			pq.Array(&extraData),
			&created,
			&updated,
		)
		if err != nil {
			return nil, err
		}

		var createdTime types.DateTime
		createdTime = createdTime.FromTimestamp(created)

		var updatedTime types.DateTime
		updatedTime = updatedTime.FromTimestamp(updated)

		info := datastore_types.NewDataStoreMetaInfo()
		info.DataID = types.UInt64(dataID)
		info.OwnerID = types.PID(owner)
		info.Size = types.UInt32(size)
		info.Name = types.String(name)
		info.DataType = types.UInt16(dataType)
		info.MetaBinary = types.QBuffer(metaBinary)
		info.Permission.Permission = datastore_constants.Permission(permission)
		info.Permission.RecipientIDs = toPIDList(permissionRecipients)
		info.DelPermission.Permission = datastore_constants.Permission(delPermission)
		info.DelPermission.RecipientIDs = toPIDList(delPermissionRecipients)
		info.CreatedTime = createdTime
		info.UpdatedTime = updatedTime
		info.Period = types.UInt16(period)
		info.Status = datastore_constants.DataStatusNone
		info.ReferDataID = types.UInt32(referDataID)
		info.Flag = datastore_constants.DataFlag(flag)
		info.ReferredTime = createdTime
		info.ExpireTime = types.NewDateTime(0x9C3F3E0000) // * 9999-12-31T00:00:00.000Z
		info.Tags = toStringList(tags)
		info.Ratings = types.NewList[datastore_types.DataStoreRatingInfoWithSlot]()

		objects = append(objects, info)
	}

	return objects, rows.Err()
}

func attachRatings(objects []datastore_types.DataStoreMetaInfo) error {
	if len(objects) == 0 {
		return nil
	}

	ids := make([]int64, len(objects))
	index := make(map[int64]int, len(objects))

	for i, object := range objects {
		id := int64(object.DataID)
		ids[i] = id
		index[id] = i
	}

	rows, err := globals.Postgres.Query(`
		SELECT data_id, slot, total_value, rating_count, initial_value
		FROM datastore.object_ratings
		WHERE data_id = ANY($1::bigint[])
		ORDER BY data_id, slot`,
		pq.Array(ids),
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var dataID, total, count, initial int64
		var slot int16

		err = rows.Scan(&dataID, &slot, &total, &count, &initial)
		if err != nil {
			return err
		}

		i, ok := index[dataID]
		if !ok {
			continue
		}

		rating := datastore_types.NewDataStoreRatingInfoWithSlot()
		rating.Slot = types.Int8(slot)
		rating.Rating.TotalValue = types.Int64(total)
		rating.Rating.Count = types.UInt32(count)
		rating.Rating.InitialValue = types.Int64(initial)

		objects[i].Ratings = append(objects[i].Ratings, rating)
	}

	return rows.Err()
}

func queryObjects(where string, args ...any) ([]datastore_types.DataStoreMetaInfo, error) {
	rows, err := globals.Postgres.Query(`SELECT `+objectColumns+` FROM datastore.objects o WHERE `+where, args...)
	if err != nil {
		return nil, err
	}

	objects, err := scanObjects(rows)
	if err != nil {
		return nil, err
	}

	err = attachRatings(objects)
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func queryObject(where string, args ...any) (datastore_types.DataStoreMetaInfo, *nex.Error) {
	objects, err := queryObjects(where+` LIMIT 1`, args...)
	if err != nil {
		return datastore_types.NewDataStoreMetaInfo(), unknownError(err)
	}

	if len(objects) == 0 {
		return datastore_types.NewDataStoreMetaInfo(), nex.NewError(nex.ResultCodes.DataStore.NotFound, "not found")
	}

	return objects[0], nil
}
