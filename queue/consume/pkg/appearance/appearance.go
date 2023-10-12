package appearance

import "fmt"

// TODO: this is same as database Appearance model. Can we reuse it there?
type Appearance struct {
	Address         string
	BlockNumber     uint32
	TransactionId   uint32
	BlockRangeStart uint64
	BlockRangeEnd   uint64
	// AppearanceId is unique and has format of "address.block_number.transaction_id"
	AppearanceId string
}

func (a *Appearance) SetAppearanceId() {
	a.AppearanceId = fmt.Sprintf("%s.%d.%d", a.Address, a.BlockNumber, a.TransactionId)
}
