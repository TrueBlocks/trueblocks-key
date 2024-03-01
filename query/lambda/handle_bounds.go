package main

import (
	"context"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleBounds(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[database.AppearancesDatasetBounds], err error) {
	// get status first, so we know max block number
	meta, err := getMeta(ctx, rpcRequest.Address())
	if err != nil {
		return
	}

	bounds, err := database.FetchAppearancesDatasetBounds(
		ctx,
		dbConn,
		rpcRequest.Address(),
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
