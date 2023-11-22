package queue

import (
	"github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/appearance"
)

type RemoteQueuer interface {
	Init() error
	Add(app *appearance.Appearance) (string, error)
	AddBatch(apps []*appearance.Appearance) (err error)
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

func (q *Queue) Add(app *appearance.Appearance) (msgId string, err error) {
	app.SetAppearanceId()
	return q.remote.Add(app)
}

func (q *Queue) AddBatch(apps []*appearance.Appearance) (err error) {
	for _, item := range apps {
		item := item
		item.SetAppearanceId()
	}
	return q.remote.AddBatch(apps)
}
