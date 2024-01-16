package queuetest

import (
	"fmt"

	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
)

type MockQueue struct {
	apps   []*queueItem.Appearance
	chunks []*queueItem.Chunk
}

func (m *MockQueue) Init() error {
	return nil
}

func (m *MockQueue) Add(itemType queueItem.ItemType, item any) (string, error) {
	switch itemType {
	case queueItem.ItemTypeAppearance:
		app := item.(*queueItem.Appearance)
		m.apps = append(m.apps, app)
		return fmt.Sprintf("%d", m.Len()), nil
	case queueItem.ItemTypeChunk:
		chunk := item.(*queueItem.Chunk)
		m.chunks = append(m.chunks, chunk)
		return fmt.Sprintf("%d", m.Len()), nil
	default:
		return "", fmt.Errorf("unsupported type: %s", itemType)
	}
}

func (m *MockQueue) AddAppearanceBatch(items []*queueItem.Appearance) (err error) {
	for _, item := range items {
		if _, err = m.Add(queueItem.ItemTypeAppearance, item); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockQueue) AddChunkBatch(items []*queueItem.Chunk) (err error) {
	for _, item := range items {
		if _, err = m.Add(queueItem.ItemTypeChunk, item); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockQueue) Len() int {
	return len(m.apps) + len(m.chunks)
}

func (m *MockQueue) GetAppearances(index int) *queueItem.Appearance {
	return m.apps[index]
}

func (m *MockQueue) GetChunks(index int) *queueItem.Chunk {
	return m.chunks[index]
}
