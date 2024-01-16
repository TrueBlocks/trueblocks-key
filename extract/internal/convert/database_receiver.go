package convert

import (
	"context"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
)

type DatabaseReceiver struct {
	DbConn *database.Connection
}

func (d *DatabaseReceiver) SendBatch(batch []queueItem.Appearance) error {
	return database.InsertAppearanceBatch(context.TODO(), d.DbConn, batch)
}
