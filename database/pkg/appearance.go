package database

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

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
			"address":   strings.ToLower(address),
			"lastBlock": lastBlock,
			"pageSize":  limit,
		},
	)
	if err != nil {
		return
	}

	log.Println("address =", strings.ToLower(address), "limit =", limit, "lastBlock =", lastBlock)

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
			"address":             strings.ToLower(address),
			"lastBlock":           lastBlock,
			"pageSize":            limit,
			"appBlockNumber":      appBlockNumber,
			"appTransactionIndex": appTransactionIndex,
		},
	)
	if err != nil {
		return
	}

	log.Println("address =", strings.ToLower(address), "limit =", limit, "lastBlock =", lastBlock, "next?", nextPage)

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
			"address":   strings.ToLower(address),
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

type PublicAppearancesDatasetBounds struct {
	Latest   PublicAppearance `json:"latest"`
	Earliest PublicAppearance `json:"earliest"`
}

func PublicBounds(bounds *AppearancesDatasetBounds) *PublicAppearancesDatasetBounds {
	return &PublicAppearancesDatasetBounds{
		Latest:   *AppearanceToPublic(&bounds.Latest),
		Earliest: *AppearanceToPublic(&bounds.Earliest),
	}
}

func InsertAppearanceBatch(ctx context.Context, c *Connection, apps []queueItem.Appearance) (err error) {
	batch := &pgx.Batch{}

	for _, app := range apps {
		batch.Queue(
			sql.InsertAppearance(c.AppearancesTableName(), c.AddressesTableName()),
			strings.ToLower(app.Address),
			app.BlockNumber,
			app.TransactionIndex,
		)
	}

	return c.conn.SendBatch(ctx, batch).Close()
}

func (a *Appearance) Insert(ctx context.Context, c *Connection, address string) (err error) {
	_, err = c.conn.Exec(ctx,
		sql.InsertAppearance(c.AppearancesTableName(), c.AddressesTableName()),
		strings.ToLower(address),
		a.BlockNumber,
		a.TransactionIndex,
	)

	return
}

type PublicAppearance struct {
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex string `json:"transactionIndex"`
}

func AppearanceToPublic(a *Appearance) *PublicAppearance {
	return &PublicAppearance{
		BlockNumber:      strconv.FormatUint(uint64(a.BlockNumber), 10),
		TransactionIndex: strconv.FormatUint(uint64(a.TransactionIndex), 10),
	}
}

func AppearanceSliceToPublicSlice(slice []Appearance) []PublicAppearance {
	result := make([]PublicAppearance, 0, len(slice))
	for _, appearance := range slice {
		result = append(result, *AppearanceToPublic(&appearance))
	}
	return result
}

func (p *PublicAppearance) Appearance() (appearance *Appearance, err error) {
	appearance = &Appearance{}

	blockNumber, err := strconv.ParseUint(p.BlockNumber, 0, 32)
	if err != nil {
		return
	}
	transactionIndex, err := strconv.ParseUint(p.TransactionIndex, 0, 32)
	if err != nil {
		return
	}
	appearance.BlockNumber = uint32(blockNumber)
	appearance.TransactionIndex = uint32(transactionIndex)
	return
}
