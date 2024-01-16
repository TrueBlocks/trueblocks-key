package queueItem

type ItemType string

const (
	ItemTypeAppearance ItemType = "appearance"
	ItemTypeChunk      ItemType = "chunk"
)

type itemPayload interface {
	Appearance | Chunk
}

type QueueItem[T itemPayload] struct {
	Type    ItemType
	Payload T
}

func (q *QueueItem[T]) IsAppearance() bool {
	return q.Type == ItemTypeAppearance
}

func (q *QueueItem[T]) IsChunk() bool {
	return q.Type == ItemTypeChunk
}
