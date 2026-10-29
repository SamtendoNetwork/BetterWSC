package database

import (
	"fmt"
	"math"
	"strings"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore_constants "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/constants"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
	"github.com/lib/pq"

	"github.com/SamtendoNetwork/BetterWSC/globals"
)

func GetObjectInfosByDataStoreSearchParam(param datastore_types.DataStoreSearchParam, _ types.PID) ([]datastore_types.DataStoreMetaInfo, uint32, *nex.Error) {
	args := []any{"SN_WSC"}
	conditions := []string{
		`o.game = $1`,
		`o.deleted = false`,
		`o.upload_completed = true`,
	}

	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if len(param.DataTypes) > 0 {
		dataTypes := make([]int64, 0, len(param.DataTypes))
		for _, dataType := range param.DataTypes {
			dataTypes = append(dataTypes, int64(dataType))
		}

		conditions = append(conditions, fmt.Sprintf(`COALESCE(o.data_type, 0) = ANY(%s::bigint[])`, addArg(pq.Array(dataTypes))))
	} else if uint16(param.DataType) != dataTypeAny {
		conditions = append(conditions, fmt.Sprintf(`COALESCE(o.data_type, 0) = %s`, addArg(int64(param.DataType))))
	}

	if len(param.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf(`o.tags && %s::text[]`, addArg(pq.Array(fromStringList(param.Tags)))))
	}

	if len(param.OwnerIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf(`COALESCE(o.owner, 0) = ANY(%s::bigint[])`, addArg(pq.Array(fromPIDList(param.OwnerIDs)))))
	}

	if param.ReferDataID != 0 {
		conditions = append(conditions, fmt.Sprintf(`COALESCE(o.refer_data_id, 0) = %s`, addArg(int64(param.ReferDataID))))
	}

	if param.CreatedAfter != 0 {
		conditions = append(conditions, fmt.Sprintf(`o.creation_date >= %s`, addArg(param.CreatedAfter.Standard())))
	}

	if param.CreatedBefore != 0 {
		conditions = append(conditions, fmt.Sprintf(`o.creation_date <= %s`, addArg(param.CreatedBefore.Standard())))
	}

	if param.UpdatedAfter != 0 {
		conditions = append(conditions, fmt.Sprintf(`o.update_date >= %s`, addArg(param.UpdatedAfter.Standard())))
	}

	if param.UpdatedBefore != 0 {
		conditions = append(conditions, fmt.Sprintf(`o.update_date <= %s`, addArg(param.UpdatedBefore.Standard())))
	}

	where := strings.Join(conditions, " AND ")

	var total int64

	err := globals.Postgres.QueryRow(`SELECT COUNT(*) FROM datastore.objects o WHERE `+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, unknownError(err)
	}

	direction := "ASC"
	if param.ResultOrder == datastore_constants.SearchSortOrderDesc {
		direction = "DESC"
	}

	length := uint32(param.ResultRange.Length)
	if length == 0 || length > 1000 {
		length = 100
	}

	offset := uint32(param.ResultRange.Offset)
	if offset == math.MaxUint32 {
		offset = 0
	}

	where += fmt.Sprintf(
		` ORDER BY %s %s, o.data_id %s LIMIT %s OFFSET %s`,
		searchOrderExpression(param.ResultOrderColumn),
		direction,
		direction,
		addArg(int64(length)),
		addArg(int64(offset)),
	)

	objects, err := queryObjects(where, args...)
	if err != nil {
		return nil, 0, unknownError(err)
	}

	return objects, uint32(total), nil
}