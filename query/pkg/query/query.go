package query

import (
	"fmt"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

// TODO make it configurable
const maxLimit = 1000

type Query struct {
	Limit      int
	Offset     int
	Address    string
	Connection *database.Connection
}

type Item struct {
	Address       string
	BlockNumber   uint32
	TransactionId uint32
}

func (q *Query) Do() (results []Item, err error) {
	if q.Limit > maxLimit {
		q.Limit = maxLimit
	}

	if err := q.Connection.Connect(); err != nil {
		return nil, fmt.Errorf("query.Do: connecting: %w", err)
	}

	results = make([]Item, 0, q.Limit)
	query := q.Connection.Db().Where("address like ?", q.Address)
	query.Model(&database.Appearance{})
	query.Limit(q.Limit)
	query.Offset(q.Offset)
	query.Order("block_number desc")
	dbtx := query.Find(&results)
	if err = dbtx.Error; err != nil {
		return nil, fmt.Errorf("query.Do: executing: %w", err)
	}

	return
}
