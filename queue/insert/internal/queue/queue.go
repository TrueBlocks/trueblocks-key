package queue

import queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"

type RemoteQueuer interface {
	Init() error
	Add(itemType queueItem.ItemType, item any) (string, error)
	AddAppearanceBatch(items []*queueItem.Appearance) (err error)
	AddChunkBatch(items []*queueItem.Chunk) (err error)
}

type Queue struct {
	remote RemoteQueuer
}

func NewQueue(remoteQueue RemoteQueuer) (q *Queue, err error) {
	err = remoteQueue.Init()
	q = &Queue{
		remote: remoteQueue,
	}
	return
}

func (q *Queue) AddAppearance(app *queueItem.Appearance) (msgId string, err error) {
	return q.remote.Add(queueItem.ItemTypeAppearance, app)
}

func (q *Queue) AddAppearanceBatch(apps []*queueItem.Appearance) (err error) {
	return q.remote.AddAppearanceBatch(apps)
}

func (q *Queue) AddChunk(chunk *queueItem.Chunk) (msgId string, err error) {
	return q.remote.Add(queueItem.ItemTypeChunk, chunk)
}

func (q *Queue) AddChunkBatch(chunks []*queueItem.Chunk) (err error) {
	return q.remote.AddChunkBatch(chunks)
}
