package query

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type RpcRequest struct {
	Id     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type NoParam struct{}

type rpcRequestParams interface {
	RpcGetAppearancesParam |
		RpcGetAddressesInParam |
		RpcGetAppearanceCountParam |
		NoParam
}

func (r *RpcRequest) AppearancesParams() (RpcParams[RpcGetAppearancesParam], error) {
	return unmarshalParams[RpcGetAppearancesParam](r)
}

func (r *RpcRequest) AppearanceCountParams() (RpcParams[RpcGetAppearanceCountParam], error) {
	return unmarshalParams[RpcGetAppearanceCountParam](r)
}

func (r *RpcRequest) AddressesInParam() (RpcParams[RpcGetAddressesInParam], error) {
	return unmarshalParams[RpcGetAddressesInParam](r)
}

func (r *RpcRequest) LambdaPayload() (string, error) {
	encoded, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`{"body": %s}`, strconv.Quote(string(encoded))), nil
}

func SetParams[T rpcRequestParams](r *RpcRequest, params RpcParams[T]) (err error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	r.Params = raw
	return
}

func unmarshalParams[T rpcRequestParams](r *RpcRequest) (result RpcParams[T], err error) {
	err = json.Unmarshal(r.Params, &result)
	return
}
