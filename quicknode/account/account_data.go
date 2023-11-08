package qnaccount

import (
	"errors"
	"fmt"
	"strings"

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
		err = fmt.Errorf("parsing account data json: %w", err)
		return
	}
	accountData.Test = (c.GetHeader("X-QN-TESTING") == "true")

	cnf, err := config.Get("")
	if err != nil {
		err = fmt.Errorf("account data: reading config: %w", err)
		return
	}

	err = ValidateChainNetwork(accountData.Chain, accountData.Network, cnf)
	return
}

func ValidateChainNetwork(chain string, network string, cnf *config.ConfigFile) (err error) {
	chain = strings.ToLower(chain)
	network = strings.ToLower(network)
	allowedChain, ok := cnf.Chains.Allowed[chain]
	if !ok {
		err = ErrInvalidChainNetwork
		return
	}
	for _, allowedNetwork := range allowedChain {
		if allowedNetwork == network {
			return
		}
	}

	// Network is not supported
	err = ErrInvalidChainNetwork
	return
}
