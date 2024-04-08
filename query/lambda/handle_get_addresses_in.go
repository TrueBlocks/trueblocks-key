package main

import (
	"context"
	"net/http"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleGetAddressesIn(ctx context.Context, rpcRequest *query.RpcRequest, inTx bool) (response *query.RpcResponse[[]string], err error) {
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

	blockNumber, err := param.BlockNumberUint()
	if err != nil {
		err = NewRpcError(err, http.StatusBadRequest, "invalid block number")
		return
	}

	var transactionIndex uint32

	if inTx {
		transactionIndex, err = param.TransactionIndexUint()
		if err != nil {
			err = NewRpcError(err, http.StatusBadRequest, "invalid transaction index")
			return
		}
	}

	var addrs []string
	if inTx {
		addrs, err = database.FetchAddressesInTx(
			ctx,
			dbConn,
			int(blockNumber),
			int(transactionIndex),
		)
	} else {
		addrs, err = database.FetchAddressesInBlock(
			ctx,
			dbConn,
			int(blockNumber),
		)
	}

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
