package main

import (
	"context"
	"log"
	"net/http"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

const defaultAppearancesLimit = 100

func handleGetAppearances(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[[]database.Appearance], err error) {
	rpcParams, err := rpcRequest.AppearancesParams()
	if err != nil {
		err = NewRpcError(err, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err = rpcParams.Validate(); err != nil {
		err = NewRpcError(err, http.StatusBadRequest, err.Error())
		// Validate() always returns public errors
		return
	}

	param := rpcParams.Get()
	if err = param.Validate(); err != nil {
		err = NewRpcError(err, http.StatusBadRequest, err.Error())
		return
	}
	limit, offset := getValidLimits(param)

	// get status first, so we know max block number
	meta, err := getMeta(ctx, param.Address)
	if err != nil {
		return
	}

	items, err := database.FetchAppearances(ctx, dbConn, param.Address, meta.LastIndexedBlock, uint(limit), uint(offset))
	if err != nil {
		log.Println("database query:", err)
		err = ErrInternal
		return
	}

	response = &query.RpcResponse[[]database.Appearance]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result: query.Result[[]database.Appearance]{
			Data: items,
			Meta: meta,
		},
	}
	return
}
