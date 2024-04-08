package query

import "strconv"

type RpcGetAddressesInParam struct {
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex uint   `json:"transactionIndex"`
}

func (r *RpcGetAddressesInParam) Validate() error {
	return nil
}

func (r *RpcGetAddressesInParam) BlockNumberUint() (uint32, error) {
	parsed, err := strconv.ParseUint(r.BlockNumber, 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), err
}
