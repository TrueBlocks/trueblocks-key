package convert

import (
	"context"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/appearance"
)

type DatabaseReceiver struct {
	DbConn *database.Connection
}

func (d *DatabaseReceiver) SendBatch(batch []appearance.Appearance) error {
	return database.InsertAppearanceBatch(context.TODO(), d.DbConn, batch)
}
