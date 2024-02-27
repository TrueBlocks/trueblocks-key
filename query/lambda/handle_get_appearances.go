package main

import (
	"context"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

const defaultAppearancesLimit = 100

func handleGetAppearances(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[[]database.Appearance], err error) {
	limit := rpcRequest.Parameters().PerPage
	if limit == 0 {
		// Just in case we forgot to define the limit in configuration
		limit = defaultAppearancesLimit
	}
	if limit < query.MinLimit {
		limit = query.MinLimit
	}

	if confLimit := cnf.Query.MaxLimit; confLimit > 0 {
		if limit > int(confLimit) {
			limit = int(confLimit)
		}
	}

	offset := rpcRequest.Parameters().Page
	if offset < 0 {
		offset = 0
	}
	offset = offset * limit

	// get status first, so we know max block number
	meta, err := getMeta(ctx, rpcRequest.Address())
	if err != nil {
		return
	}

	items, err := database.FetchAppearances(ctx, dbConn, rpcRequest.Address(), meta.LastIndexedBlock, uint(limit), uint(offset))
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
