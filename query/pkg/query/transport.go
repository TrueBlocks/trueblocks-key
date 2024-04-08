package query

import (
	"errors"
	"strconv"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

var ErrAddressIncorrect = errors.New("incorrect address")
var ErrIncorrectPerPage = errors.New("incorrect perPage")
var ErrWrongNumOfParameters = errors.New("exactly 1 parameter object required")
var ErrInvalidLastBlockSpecial = errors.New("if lastBlock is a string, it has to be 'latest'")
var ErrInvalidLastBlockInvalid = errors.New("lastBlock must be a number or string")

// MaxSafePerPage is the largest sane value of PerPage that we would allow users to use
const MaxSafePerPage = 1000

// If limit is too small the DB can take too long to return results
const MinSafePerPage = 5

const LastBlockLatest = "latest"

type Validator interface {
	Validate() error
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
	[]database.PublicAppearance |
		[]string |
		database.PublicAppearancesDatasetBounds |
		*database.Status |
		*int
}

type Meta struct {
	LastIndexedBlock string  `json:"lastIndexedBlock"`
	Address          string  `json:"address,omitempty"`
	PreviousPageId   *PageId `json:"previousPageId"`
	NextPageId       *PageId `json:"nextPageId"`

	lastIndexedBlock uint
}

func (m *Meta) SetLastIndexedBlock(blockNumber uint) {
	m.LastIndexedBlock = strconv.FormatUint(uint64(blockNumber), 10)
	m.lastIndexedBlock = blockNumber
}

func (m *Meta) LastIndexedBlockUint() uint {
	return m.lastIndexedBlock
}

type RpcAddressesResponse struct {
	JsonRpc string   `json:"jsonrpc"`
	Id      int      `json:"id"`
	Result  []string `json:"result"`
}
