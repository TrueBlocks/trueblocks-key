package query

type RpcGetAddressesInParam struct {
	BlockNumber      uint `json:"blockNumber"`
	TransactionIndex uint `json:"transactionIndex"`
	Page             uint `json:"page"`
	PerPage          uint `json:"perPage"`
}

func (r *RpcGetAddressesInParam) Pagination() (uint, uint) {
	return r.Page, r.PerPage
}

func (r *RpcGetAddressesInParam) Validate() error {
	// Validate pagination
	return validatePagination(r)
}
