package database

import (
	"context"
	"log"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/jackc/pgx/v5"
)

const hardFetchLimit = 3000

type Appearance struct {
	BlockNumber   uint32 `json:"blockNumber"`
	TransactionId uint32 `json:"transactionId"`
}

func FetchAppearances(ctx context.Context, c *Connection, address string, limit uint, offset uint) (results []Appearance, err error) {
	if limit > hardFetchLimit {
		log.Printf("database/FetchAppearances: limit too large (%d),setting it to %d\n", limit, hardFetchLimit)
	}
	rows, err := c.conn.Query(
		ctx,
		sql.SelectAppearances(c.AppearancesTableName(), c.AddressesTableName()),
		address,
		limit,
		offset,
	)
	if err != nil {
		return
	}

	log.Println("limit =", limit, "offset =", offset)

	results, err = pgx.CollectRows[Appearance](rows, pgx.RowToStructByPos[Appearance])

	return
}

func InsertAppearanceBatch(ctx context.Context, c *Connection, apps []queueItem.Appearance) (err error) {
	batch := &pgx.Batch{}

	for _, app := range apps {
		batch.Queue(
			sql.InsertAppearance(c.AppearancesTableName(), c.AddressesTableName()),
			app.Address,
			app.BlockNumber,
			app.TransactionId,
		)
	}

	return c.conn.SendBatch(ctx, batch).Close()
}

func (a *Appearance) Insert(ctx context.Context, c *Connection, address string) (err error) {
	_, err = c.conn.Exec(ctx,
		sql.InsertAppearance(c.AppearancesTableName(), c.AddressesTableName()),
		address,
		a.BlockNumber,
		a.TransactionId,
	)

	return
}

func FetchMaxBlockNumber(ctx context.Context, c *Connection) (result int, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.SelectAppearancesMaxBlockNumber(c.AppearancesTableName()),
	)
	if err != nil {
		return
	}

	return pgx.CollectOneRow[int](rows, pgx.RowTo[int])
}
