package query

import (
	"context"
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

func (q *Query) Do() (results []database.Appearance, err error) {
	if q.Limit > maxLimit {
		q.Limit = maxLimit
	}

	results, err = database.FetchAppearances(context.TODO(), q.Connection, q.Address, uint(q.Limit), uint(q.Offset))
	if err != nil {
		return nil, fmt.Errorf("query.Do: executing: %w", err)
	}

	return
}
