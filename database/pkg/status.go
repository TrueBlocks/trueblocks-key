package database

import (
	"context"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	"github.com/jackc/pgx/v5"
)

type Status struct {
	LastIndexedBlock uint `json:"lastIndexedBlock"`
}

func (s *Status) HasLastIndexedBlock() bool {
	return s.LastIndexedBlock > 0
}

func FetchStatus(ctx context.Context, c *Connection) (result Status, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.SelectStatus(c.AppearancesTableName(), c.AddressesTableName()),
	)
	if err != nil {
		return
	}

	result, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[Status])

	return
}
