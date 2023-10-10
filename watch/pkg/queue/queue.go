package queue

import (
	"gorm.io/gorm"
	database "trueblocks.io/database/pkg"
)

type Queue struct {
	destConn  *database.Connection
	queueConn *database.Connection
}

func NewQueue(destConn, queueConn *database.Connection) (q *Queue, err error) {
	q = &Queue{
		destConn,
		queueConn,
	}

	err = q.destConn.Db().AutoMigrate(
		&database.QueueItem{},
	)
	return
}

func (q *Queue) Add(appearance *database.Appearance) error {
	// Make sure we have unique appearance ID
	appearance.SetAppearanceId()

	// Save item to queue database
	item := &database.QueueItem{
		Appearance: appearance,
	}
	item.SetWaiting()

	return q.queueConn.Db().Create(item).Error
}

func (q *Queue) Process() (err error) {
	var items []database.QueueItem

	// query := q.queueConn.Db().Where("not (status like ? or status like ?)", database.StatusSuccess, database.StatusUploading)
	// query.Limit(100)
	// dbtx := query.Find(&items)
	// if err = dbtx.Error; err != nil {
	// 	return
	// }

	err = q.queueConn.Db().Transaction(func(tx *gorm.DB) error {
		query := tx.Where("not (status like ? or status like ?)", database.StatusSuccess, database.StatusUploading)
		query.Limit(100)
		if err := query.Find(&items).Error; err != nil {
			return err
		}

		return setItemsUploading(tx, items)
	})
	if err != nil {
		return
	}

	insert := q.destConn.Db().CreateInBatches(items, len(items))
	if err = insert.Error; err != nil {
		updErr := setItemsFail(q.destConn.Db(), items)
		if updErr != nil {
			// TODO: report stuck items
		}
		return
	}

	// TODO: The problem here is that this function returns single error for 2 different sources:
	// out database write and queue status update. It should report 2 errors
	return setItemsSuccess(q.destConn.Db(), items)
}

// TODO: get rid of repeated code in functions below

func setItemsUploading(tx *gorm.DB, items []database.QueueItem) (err error) {
	for _, item := range items {
		item.SetUploading()
		if err = tx.Save(&item).Error; err != nil {
			return
		}
	}
	return
}

func setItemsFail(tx *gorm.DB, items []database.QueueItem) (err error) {
	for _, item := range items {
		item.SetFail()
		if err = tx.Save(&item).Error; err != nil {
			return
		}
	}
	return
}

func setItemsSuccess(tx *gorm.DB, items []database.QueueItem) (err error) {
	for _, item := range items {
		item.SetSuccess()
		if err = tx.Save(&item).Error; err != nil {
			return
		}
	}
	return
}
