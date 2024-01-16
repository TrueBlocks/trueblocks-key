package queueItem

import "fmt"

type Appearance struct {
	Address         string
	BlockNumber     uint32
	TransactionId   uint32
	BlockRangeStart uint64
	BlockRangeEnd   uint64
}

func (a *Appearance) String() string {
	return fmt.Sprintf("%s.%d.%d", a.Address, a.BlockNumber, a.TransactionId)
}
