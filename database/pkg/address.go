package database

import (
	"context"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	"github.com/jackc/pgx/v5"
)

func FetchAddressesInTx(ctx context.Context, c *Connection, blockNumber int, transactionIndex int) (results []string, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.SelectAddressesInTx(c.AppearancesTableName(), c.AddressesTableName()),
		pgx.NamedArgs{
			"blockNumber":      blockNumber,
			"transactionIndex": transactionIndex,
		},
	)
	if err != nil {
		return
	}

	results, err = pgx.CollectRows[string](rows, pgx.RowTo[string])

	return
}

func FetchAddressesInBlock(ctx context.Context, c *Connection, blockNumber int) (results []string, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.SelectAddressesInBlock(c.AppearancesTableName(), c.AddressesTableName()),
		pgx.NamedArgs{
			"blockNumber": blockNumber,
		},
	)
	if err != nil {
		return
	}

	results, err = pgx.CollectRows[string](rows, pgx.RowTo[string])

	return
}
