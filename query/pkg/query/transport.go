package query

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

var ErrInvalidMethod = errors.New("invalid method")
var ErrAddressIncorrect = errors.New("incorrect address")
var ErrIncorrectPerPage = errors.New("incorrect perPage")
var ErrWrongNumOfParameters = errors.New("exactly 1 parameter object required")
var ErrInvalidLastBlockSpecial = errors.New("if lastBlock is a string, it has to be 'latest'")
var ErrInvalidLastBlockInvalid = errors.New("lastBlock must be a number or string")

// MaxSafePerPage is the largest sane value of PerPage that we would allow users to use
const MaxSafePerPage = 10000

// If limit is too small the DB can take too long to return results
const MinSafePerPage = 10

const LastBlockLatest = "latest"

type RpcRequest struct {
	Id     any                `json:"id"`
	Method string             `json:"method"`
	Params []RpcRequestParams `json:"params"`
}

type RpcRequestParams struct {
	Address   string           `json:"address"`
	LastBlock *json.RawMessage `json:"lastBlock,omitempty"`
	PageId    *PageId          `json:"pageId,omitempty"`
	PerPage   int              `json:"perPage"`
}

func (r *RpcRequest) Parameters() RpcRequestParams {
	return r.Params[0]
}

func (r *RpcRequest) Address() string {
	return strings.ToLower(r.Params[0].Address)
}

// LastBlock returns nil for latest block, block number otherwise
func (r *RpcRequest) LastBlockNumber() (*uint, error) {
	params := r.Parameters()

	if params.LastBlock == nil {
		return nil, nil
	}

	var special string
	err := json.Unmarshal(*params.LastBlock, &special)
	if err != nil {
		log.Println("last block is not special because of error:", err)
	} else {
		// it is special
		if special == LastBlockLatest {
			return nil, nil
		} else {
			// it's invalid
			return nil, ErrInvalidLastBlockSpecial
		}
	}

	// Try a number
	var blockNumber uint
	if err := json.Unmarshal(*params.LastBlock, &blockNumber); err != nil {
		return nil, ErrInvalidLastBlockInvalid
	}

	return &blockNumber, nil
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
	if r.Parameters().PerPage < 0 || r.Parameters().PerPage > MaxSafePerPage {
		return ErrIncorrectPerPage
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
	Id        any    `json:"id"`
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
	LastIndexedBlock uint    `json:"lastIndexedBlock"`
	Address          string  `json:"address,omitempty"`
	PreviousPageId   *PageId `json:"previousPageId"`
	NextPageId       *PageId `json:"nextPageId"`
}
