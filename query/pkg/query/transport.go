package query

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

var ErrInvalidMethod = errors.New("invalid method")
var ErrAddressIncorrect = errors.New("incorrect address")
var ErrIncorrectPagePerPage = errors.New("incorrect page or perPage")
var ErrWrongNumOfParameters = errors.New("exactly 1 parameter object required")

// MaxSafePerPage is the largest sane value of PerPage that we would allow users to use
const MaxSafePerPage = 10000

// MaxSafePage is the largest sane page number that we would allow users to use
const MaxSafePage = 100000

type RpcRequest struct {
	Id     int                `json:"id"`
	Method string             `json:"method"`
	Params []RpcRequestParams `json:"params"`
}

type RpcRequestParams struct {
	Address string `json:"address"`
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
}

func (r *RpcRequest) Parameters() RpcRequestParams {
	return r.Params[0]
}

func (r *RpcRequest) Address() string {
	return strings.ToLower(r.Params[0].Address)
}

func (r *RpcRequest) Validate() error {
	// Validate method
	if r.Method != MethodGetAppearances && r.Method != MethodGetAppearanceCount && r.Method != MethodLastIndexedBlock {
		return ErrInvalidMethod
	}

	if r.Method == MethodLastIndexedBlock {
		return nil
	}

	if len(r.Params) != 1 {
		return ErrWrongNumOfParameters
	}

	// Validate address
	if len(r.Parameters().Address) != 42 {
		return ErrAddressIncorrect
	}
	if r.Parameters().Address[:2] != "0x" {
		return ErrAddressIncorrect
	}
	if _, err := hex.DecodeString(r.Parameters().Address[2:]); err != nil {
		return ErrAddressIncorrect
	}

	// Validate pagination
	if r.Parameters().Page < 0 || r.Parameters().PerPage < 0 {
		return ErrIncorrectPagePerPage
	}

	if r.Parameters().Page > MaxSafePage || r.Parameters().PerPage > MaxSafePerPage {
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

type RpcResponse[T RpcResponseResult] struct {
	JsonRpc   string `json:"jsonrpc"`
	Id        int    `json:"id"`
	Result[T] `json:"result"`
}

type Result[T RpcResponseResult] struct {
	Data  T `json:"data"`
	*Meta `json:"meta"`
}

type RpcResponseResult interface {
	[]database.Appearance |
		*database.Status |
		*int
}

type Meta struct {
	LastIndexedBlock uint   `json:"lastIndexedBlock"`
	Address          string `json:"address,omitempty"`
}
