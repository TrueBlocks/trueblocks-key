package queue

import (
	"os"

	"trueblocks.io/queue/consume/pkg/appearance"
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

func (f *FileQueue) Add(app *appearance.Appearance) (msgId string, err error) {
	app.SetAppearanceId()
	_, err = f.file.WriteString(app.AppearanceId + "\n")
	return
}

func (f *FileQueue) Close() error {
	return f.file.Close()
}
