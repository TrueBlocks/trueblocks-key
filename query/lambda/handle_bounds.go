package main

import (
	"context"
	"log"
	"net/http"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleBounds(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[database.AppearancesDatasetBounds], err error) {
	rpcParams, err := rpcRequest.BoundsParams()
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

	// get status first, so we know max block number
	meta, err := getMeta(ctx, param.Address)
	if err != nil {
		return
	}

	bounds, err := database.FetchAppearancesDatasetBounds(
		ctx,
		dbConn,
		param.Address,
		meta.LastIndexedBlock,
	)
	if err != nil {
		log.Println("database query (count):", err)
		err = ErrInternal
		return
	}

	response = &query.RpcResponse[database.AppearancesDatasetBounds]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result: query.Result[database.AppearancesDatasetBounds]{
			Data: bounds,
			Meta: meta,
		},
	}
	return
}
