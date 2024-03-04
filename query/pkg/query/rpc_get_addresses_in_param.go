package query

type RpcGetAddressesInParam struct {
	BlockNumber      uint `json:"blockNumber"`
	TransactionIndex uint `json:"transactionIndex"`
}

func (r *RpcGetAddressesInParam) Validate() error {
	return nil
}
