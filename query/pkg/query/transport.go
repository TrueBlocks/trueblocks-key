package query

import (
	"errors"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

var ErrAddressIncorrect = errors.New("incorrect address")
var ErrIncorrectPagePerPage = errors.New("incorrect page or perPage")
var ErrWrongNumOfParameters = errors.New("exactly 1 parameter object required")

// MaxSafePerPage is the largest sane value of PerPage that we would allow users to use
const MaxSafePerPage = 10000

// MaxSafePage is the largest sane page number that we would allow users to use
const MaxSafePage = 100000

type Validator interface {
	Validate() error
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
		[]string |
		*database.Status |
		*int
}

type Meta struct {
	LastIndexedBlock uint   `json:"lastIndexedBlock"`
	Address          string `json:"address,omitempty"`
}

type RpcAddressesResponse struct {
	JsonRpc string   `json:"jsonrpc"`
	Id      int      `json:"id"`
	Result  []string `json:"result"`
}
