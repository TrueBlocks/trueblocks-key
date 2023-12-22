package convertNew

import "github.com/ethereum/go-ethereum/common"

type addressRecord struct {
	Address common.Address `json:"address"`
	Offset  uint32         `json:"offset"`
	Count   uint32         `json:"count"`
}
