package queue

import (
	"fmt"
	"os"

	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
)

// FileQueue is only meant for local testing, as it uses
// a file rather than queue implementation
type FileQueue struct {
	path string
	file *os.File
}

func NewFileQueue(path string) *FileQueue {
	return &FileQueue{
		path: path,
	}
}

func (f *FileQueue) Init() (err error) {
	f.file, err = os.OpenFile(f.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)

	return err
}

func (f *FileQueue) Add(itemType queueItem.ItemType, item any) (msgId string, err error) {
	var content string
	switch itemType {
	case queueItem.ItemTypeAppearance:
		content = item.(*queueItem.Appearance).String()
	case queueItem.ItemTypeChunk:
		chunk := item.(*queueItem.Chunk)
		content = fmt.Sprintf("%s: %s (%s)", chunk.Range, chunk.Cid, chunk.Author)
	default:
		return "", fmt.Errorf("unsupported queue item type: %s", itemType)
	}
	_, err = f.file.WriteString(content + "\n")
	return
}

func (f *FileQueue) AddAppearanceBatch(items []*queueItem.Appearance) (err error) {
	for _, app := range items {
		if _, err = f.Add(queueItem.ItemTypeAppearance, app); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileQueue) AddChunkBatch(items []*queueItem.Chunk) (err error) {
	for _, app := range items {
		if _, err = f.Add(queueItem.ItemTypeChunk, app); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileQueue) Close() error {
	return f.file.Close()
}
