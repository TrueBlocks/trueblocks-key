package convertNew

import (
	"bytes"
	"encoding/binary"
	"io"
)

func readBytes(r io.ReaderAt, offset int64, size int, dest any) (err error) {
	p := make([]byte, size)
	if _, err = r.ReadAt(p, offset); err != nil {
		return
	}
	pr := bytes.NewReader(p)
	return binary.Read(pr, binary.LittleEndian, dest)
}
