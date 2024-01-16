package database

import (
	"context"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	"github.com/jackc/pgx/v5"
)

func FetchAppearancesCount(ctx context.Context, c *Connection) (result int, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.SelectAppearancesCount(c.AppearancesTableName()),
	)
	if err != nil {
		return
	}

	return pgx.CollectOneRow[int](rows, pgx.RowTo[int])
}
