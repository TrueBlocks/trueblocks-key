package scenario

import (
	"net/http"
)

type Scenario struct {
	// User returns auth headers
	Headers http.Header `toml:"headers"`
	Address string      `toml:"address"`
	PerPage uint        `toml:"perPage"`
}
