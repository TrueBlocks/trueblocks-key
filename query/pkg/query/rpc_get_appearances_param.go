package query

type RpcGetAppearancesParam struct {
	Address string `json:"address"`
	Page    uint   `json:"page"`
	PerPage uint   `json:"perPage"`
}

func (r *RpcGetAppearancesParam) Pagination() (uint, uint) {
	return r.Page, r.PerPage
}

func (r *RpcGetAppearancesParam) Validate() error {
	if err := validateAddress(r.Address); err != nil {
		return err
	}

	// Validate pagination
	if err := validatePagination(r); err != nil {
		return err
	}

	return nil
}
