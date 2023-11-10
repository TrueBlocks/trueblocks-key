package convert

import database "trueblocks.io/database/pkg"

type DatabaseReceiver struct {
	DbConn *database.Connection
}

func (d *DatabaseReceiver) SendBatch(batch []*database.Appearance) error {
	err := d.DbConn.Db().Create(batch).Error
	return err
}
