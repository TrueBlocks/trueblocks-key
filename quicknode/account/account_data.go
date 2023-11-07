package qnaccount

import (
	"errors"

	config "trueblocks.io/config/pkg"
)

var ErrInvalidChainNetwork = errors.New("network or chain is not supported")

type AccountData struct {
	QuicknodeId string `json:"quicknode-id"`
	Plan        string `json:"plan"`
	EndpointId  string `json:"endpoint-id"`
	WssUrl      string `json:"wss-url"`
	HttpUrl     string `json:"http-url"`
	Chain       string `json:"chain"`
	Network     string `json:"network"`
	// Test does not come with request body, it has to be read from
	// request headers
	Test bool `json:"test"`
}

type requestContext interface {
	GetHeader(key string) string
	BindJSON(obj any) error
}

func NewAccountData(c requestContext) (accountData *AccountData, err error) {
	accountData = &AccountData{}
	if err = c.BindJSON(accountData); err != nil {
		return
	}
	accountData.Test = (c.GetHeader("X-QN-TESTING") == "true")

	cnf, err := config.Get("")
	if err != nil {
		return
	}
	chain, ok := cnf.Chains.Allowed[accountData.Chain]
	if !ok {
		err = ErrInvalidChainNetwork
		return
	}
	for _, network := range chain {
		if network == accountData.Network {
			return
		}
	}

	// Network is not supported
	err = ErrInvalidChainNetwork
	return
}
