package query

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrAddressIncorrect = errors.New("incorrect address")
var ErrIncorrectPagePerPage = errors.New("incorrect page or perPage")

// MaxSafePerPage is the largest sane value of PerPage that we would allow users to use
const MaxSafePerPage = 10000

// MaxSafePage is the largest sane page number that we would allow users to use
const MaxSafePage = 100000

type RpcRequest struct {
	Id     int              `json:"id"`
	Method string           `json:"method"`
	Params RpcRequestParams `json:"params"`
}

type RpcRequestParams struct {
	Address string `json:"address"`
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
}

func (r *RpcRequest) Address() string {
	return strings.ToLower(r.Params.Address)
}

func (r *RpcRequest) Validate() error {
	// Validate address
	if len(r.Params.Address) != 42 {
		return ErrAddressIncorrect
	}
	if r.Params.Address[:2] != "0x" {
		return ErrAddressIncorrect
	}
	if _, err := hex.DecodeString(r.Params.Address[2:]); err != nil {
		return ErrAddressIncorrect
	}

	// Validate pagination
	if r.Params.Page < 0 || r.Params.PerPage < 0 {
		return ErrIncorrectPagePerPage
	}

	if r.Params.Page > MaxSafePage || r.Params.PerPage > MaxSafePerPage {
		return ErrIncorrectPagePerPage
	}

	return nil
}

func (r *RpcRequest) LambdaPayload() (string, error) {
	encoded, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`{"body": %s}`, strconv.Quote(string(encoded))), nil
}

// PublicAppearance has only members that we want to share with
// the outside world
type PublicAppearance struct {
	Address       string
	BlockNumber   uint32
	TransactionId uint32
}

type RpcResponse struct {
	JsonRpc string             `json:"jsonrpc"`
	Id      int                `json:"id"`
	Result  []PublicAppearance `json:"result"`
}
