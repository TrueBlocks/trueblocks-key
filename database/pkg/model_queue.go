package database

const (
	StatusWaiting   = "waiting"
	StatusUploading = "uploading"
	StatusSuccess   = "success"
	StatusFail      = "fail"
	StatusWillRetry = "will_retry"
)

type QueueItem struct {
	*Appearance
	Status string
}

func (q *QueueItem) SetWaiting() {
	q.Status = StatusWaiting
}

func (q *QueueItem) SetUploading() {
	q.Status = StatusUploading
}

func (q *QueueItem) SetSuccess() {
	q.Status = StatusSuccess
}

func (q *QueueItem) SetFail() {
	q.Status = StatusFail
}

func (q *QueueItem) SetWillRetry() {
	q.Status = StatusWillRetry
}
