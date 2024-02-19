package database

import (
	"context"
	"log"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	"github.com/jackc/pgx/v5"
)

func FetchAddressesInTx(ctx context.Context, c *Connection, blockNumber int, transactionIndex int, limit uint, offset uint) (results []string, err error) {
	if limit > hardFetchLimit {
		log.Printf("database/FetchAddressesInTx: limit too large (%d),setting it to %d\n", limit, hardFetchLimit)
	}
	rows, err := c.conn.Query(
		ctx,
		sql.SelectAddressesInTx(c.AppearancesTableName(), c.AddressesTableName()),
		blockNumber,
		transactionIndex,
		limit,
		offset,
	)
	if err != nil {
		return
	}

	results, err = pgx.CollectRows[string](rows, pgx.RowTo[string])

	return
}
