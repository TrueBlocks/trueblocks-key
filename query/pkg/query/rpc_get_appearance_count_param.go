package query

type BoundsParam struct {
	Address string `json:"address"`
}

func (r *BoundsParam) Validate() error {
	return validateAddress(r.Address)
}
