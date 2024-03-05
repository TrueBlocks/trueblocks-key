package database

import (
	"context"
	"fmt"
	"log"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/jackc/pgx/v5"
)

const hardFetchLimit = 1000

type Appearance struct {
	BlockNumber      uint32 `json:"blockNumber"`
	TransactionIndex uint32 `json:"transactionIndex"`
}

func FetchAppearancesFirstPage(ctx context.Context, c *Connection, earliest bool, address string, lastBlock uint, limit uint) (results []Appearance, err error) {
	if limit > hardFetchLimit {
		log.Printf("database/FetchAppearances: limit too large (%d),setting it to %d\n", limit, hardFetchLimit)
	}

	var sqlString string
	if earliest {
		sqlString = sql.SelectAppearancesEarliestPage(c.AppearancesTableName(), c.AddressesTableName())
	} else {
		sqlString = sql.SelectAppearancesFirstPage(c.AppearancesTableName(), c.AddressesTableName())
	}

	rows, err := c.conn.Query(
		ctx,
		sqlString,
		pgx.NamedArgs{
			"address":   address,
			"lastBlock": lastBlock,
			"pageSize":  limit,
		},
	)
	if err != nil {
		return
	}

	log.Println("address =", address, "limit =", limit, "lastBlock =", lastBlock)

	results, err = pgx.CollectRows[Appearance](rows, pgx.RowToStructByPos[Appearance])

	return
}

func FetchAppearancesPage(ctx context.Context, c *Connection, nextPage bool, address string, lastBlock uint, limit uint, appBlockNumber uint, appTransactionIndex uint) (results []Appearance, err error) {
	if limit > hardFetchLimit {
		log.Printf("database/FetchAppearances: limit too large (%d),setting it to %d\n", limit, hardFetchLimit)
	}

	var sqlString string
	if nextPage {
		sqlString = sql.SelectAppearancesNextPage(c.AppearancesTableName(), c.AddressesTableName())
	} else {
		sqlString = sql.SelectAppearancesPreviousPage(c.AppearancesTableName(), c.AddressesTableName())
	}
	rows, err := c.conn.Query(
		ctx,
		sqlString,
		pgx.NamedArgs{
			"address":             address,
			"lastBlock":           lastBlock,
			"pageSize":            limit,
			"appBlockNumber":      appBlockNumber,
			"appTransactionIndex": appTransactionIndex,
		},
	)
	if err != nil {
		return
	}

	log.Println("address =", address, "limit =", limit, "lastBlock =", lastBlock, "next?", nextPage)

	results, err = pgx.CollectRows[Appearance](rows, pgx.RowToStructByPos[Appearance])

	return
}

type AppearancesDatasetBounds struct {
	Latest   Appearance `json:"latest"`
	Earliest Appearance `json:"earliest"`
}

func (a *AppearancesDatasetBounds) IsLatest(appearance *Appearance) bool {
	return appearance.BlockNumber == a.Latest.BlockNumber && appearance.TransactionIndex == a.Latest.TransactionIndex
}

func (a *AppearancesDatasetBounds) IsEarliest(appearance *Appearance) bool {
	return appearance.BlockNumber == a.Earliest.BlockNumber && appearance.TransactionIndex == a.Earliest.TransactionIndex
}

func FetchAppearancesDatasetBounds(ctx context.Context, c *Connection, address string, lastBlock uint) (bounds AppearancesDatasetBounds, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.SelectAppearancesDatasetBounds(c.AppearancesTableName(), c.AddressesTableName()),
		pgx.NamedArgs{
			"address":   address,
			"lastBlock": lastBlock,
		},
	)
	if err != nil {
		return
	}

	raw, err := pgx.CollectRows[Appearance](rows, pgx.RowToStructByPos[Appearance])
	if err != nil {
		return
	}

	if l := len(raw); l != 2 {
		err = fmt.Errorf("expected bounds result length == 2, but got %d", l)
		return
	}

	bounds.Latest = raw[0]
	bounds.Earliest = raw[1]

	return
}

func InsertAppearanceBatch(ctx context.Context, c *Connection, apps []queueItem.Appearance) (err error) {
	batch := &pgx.Batch{}

	for _, app := range apps {
		batch.Queue(
			sql.InsertAppearance(c.AppearancesTableName(), c.AddressesTableName()),
			app.Address,
			app.BlockNumber,
			app.TransactionIndex,
		)
	}

	return c.conn.SendBatch(ctx, batch).Close()
}

func (a *Appearance) Insert(ctx context.Context, c *Connection, address string) (err error) {
	_, err = c.conn.Exec(ctx,
		sql.InsertAppearance(c.AppearancesTableName(), c.AddressesTableName()),
		address,
		a.BlockNumber,
		a.TransactionIndex,
	)

	return
}
