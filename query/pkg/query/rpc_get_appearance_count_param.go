package query

type RpcGetAppearanceCountParam struct {
	Address string `json:"address"`
}

func (r *RpcGetAppearanceCountParam) Validate() error {
	return validateAddress(r.Address)
}
