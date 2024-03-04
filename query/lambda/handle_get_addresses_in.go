package main

import (
	"context"
	"net/http"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleGetAddressesInTx(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[[]string], err error) {
	rpcParams, err := rpcRequest.AddressesInParam()
	if err != nil {
		err = NewRpcError(err, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err = rpcParams.Validate(); err != nil {
		// Validate() always returns public errors
		return
	}

	param := rpcParams.Get()
	if err = param.Validate(); err != nil {
		return
	}

	meta, err := getMeta(ctx, "")
	if err != nil {
		return
	}

	addrs, err := database.FetchAddressesInTx(
		ctx,
		dbConn,
		int(param.BlockNumber),
		int(param.TransactionIndex),
	)

	response = &query.RpcResponse[[]string]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result: query.Result[[]string]{
			Data: addrs,
			Meta: meta,
		},
	}
	return
}
