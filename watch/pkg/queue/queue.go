package queue

import (
	"errors"

	"gorm.io/gorm"
	database "trueblocks.io/database/pkg"
)

var ErrQueueEmpty = errors.New("queue empty")

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
	item.SetStatus(database.StatusWaiting)

	return q.queueConn.Db().Create(item).Error
}

func (q *Queue) Process() (queueErr error, destErr error) {
	var items []database.QueueItem

	queueErr = q.queueConn.Db().Transaction(func(tx *gorm.DB) error {
		query := tx.Where("not (status like ? or status like ?)", database.StatusSuccess, database.StatusUploading)
		query.Limit(100)
		if err := query.Find(&items).Error; err != nil {
			return err
		}

		return setItemsStatus(tx, database.StatusUploading, items)
	})
	if queueErr != nil {
		return
	}

	if len(items) == 0 {
		queueErr = ErrQueueEmpty
		return
	}

	insert := q.destConn.Db().CreateInBatches(items, len(items))
	if destErr = insert.Error; destErr != nil {
		// If `setItemsFail` fails, we don't have to report the error back.
		// Live database doesn't allow us to insert duplicates, so only item
		// statuses in queue will be updated next time.
		_ = setItemsStatus(q.destConn.Db(), database.StatusFail, items)
		return
	}

	// Again, if this fail we will update statuses next time
	_ = setItemsStatus(q.destConn.Db(), database.StatusSuccess, items)
	return
}

func setItemsStatus(tx *gorm.DB, status database.QueueItemStatus, items []database.QueueItem) (err error) {
	for _, item := range items {
		item.SetStatus(status)
		if err = tx.Save(&item).Error; err != nil {
			return
		}
	}
	return
}
