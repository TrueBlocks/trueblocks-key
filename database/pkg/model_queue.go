package database

import "fmt"

type QueueItemStatus int

const (
	StatusWaiting QueueItemStatus = iota
	StatusUploading
	StatusSuccess
	StatusFail
	StatusWillRetry
)

func (s QueueItemStatus) String() string {
	switch s {
	case StatusWaiting:
		return "waiting"
	case StatusUploading:
		return "uploading"
	case StatusSuccess:
		return "success"
	case StatusFail:
		return "fail"
	case StatusWillRetry:
		return "will_retry"
	default:
		panic(fmt.Errorf("unsupported Status value %d", s))
	}
}

type QueueItem struct {
	*Appearance
	Status string
}

func (q *QueueItem) SetStatus(newStatus QueueItemStatus) {
	q.Status = newStatus.String()
}
