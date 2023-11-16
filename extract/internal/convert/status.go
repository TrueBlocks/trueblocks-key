package convert

import (
	"fmt"
	"os"
	"path"
)

var statusFile *os.File

func init() {
	sf, err := os.Create(path.Join(
		os.TempDir(),
		"key_convert_status.txt",
	))
	if err != nil {
		panic(fmt.Errorf("opening status file: %w", err))
	}
	statusFile = sf
}

type StatusValue string

const (
	StatusDone     StatusValue = "done"
	StatusError                = "error"
	StatusAppError             = "appearance-error"
)

func SaveStatus(chunk string, status StatusValue) error {
	_, err := statusFile.WriteString(fmt.Sprintf(
		"%s\t%s",
		status, chunk,
	))
	return err
}

func CloseStatusFile() {
	statusFile.Close()
}
