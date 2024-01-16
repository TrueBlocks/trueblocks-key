package database

import (
	"context"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/jackc/pgx/v5"
)

type Chunk struct {
	Cid    string `json:"cid"`
	Range  string `json:"range"`
	Author string `json:"author"`
}

func InsertChunkBatch(ctx context.Context, c *Connection, chunks []queueItem.Chunk) (err error) {
	batch := &pgx.Batch{}

	for _, chunk := range chunks {
		batch.Queue(
			sql.InsertChunk(c.ChunksTableName()),
			chunk.Cid,
			chunk.Range,
			chunk.Author,
		)
	}

	return c.conn.SendBatch(ctx, batch).Close()
}

func FetchDuplicatedChunks(ctx context.Context, c *Connection) (results []string, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.SelectDuplicatedChunks(c.ChunksTableName()),
	)
	if err != nil {
		return
	}

	return pgx.CollectRows[string](rows, pgx.RowTo[string])
}

func CountChunks(ctx context.Context, c *Connection) (result int, err error) {
	rows, err := c.conn.Query(
		ctx,
		sql.CountChunks(c.ChunksTableName()),
	)
	if err != nil {
		return
	}

	return pgx.CollectOneRow[int](rows, pgx.RowTo[int])
}
