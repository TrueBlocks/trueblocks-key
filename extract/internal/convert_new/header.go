package convertNew

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
)

type indexHeader struct {
	Magic           uint32
	Hash            common.Hash
	AddressCount    uint32
	AppearanceCount uint32
}

func NewHeader(r io.ReaderAt) (*indexHeader, error) {
	i := &indexHeader{}
	if err := i.Read(r); err != nil {
		return nil, fmt.Errorf("new header: %w", err)
	}
	return i, nil
}

func (i *indexHeader) Read(r io.ReaderAt) error {
	return readBytes(r, 0, 44, i)
}
