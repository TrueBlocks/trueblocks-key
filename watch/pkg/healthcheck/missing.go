package healthcheck

import (
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	database "trueblocks.io/database/pkg"
)

func MissingAppearances(conn *database.Connection, blockRange base.BlockRange, expected uint) (diff int, err error) {
	query := conn.Db().Model(&database.Appearance{})
	query.Where(
		&database.Appearance{
			BlockRangeStart: blockRange.First,
			BlockRangeEnd:   blockRange.Last,
		},
	)
	var count int64
	err = query.Count(&count).Error
	diff = int(count - int64(expected))
	return
}
