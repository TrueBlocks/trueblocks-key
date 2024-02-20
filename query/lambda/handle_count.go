package main

import (
	"context"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleCount(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[int], err error) {
	// get status first, so we know max block number
	meta, err := getMeta(ctx, rpcRequest.Address())
	if err != nil {
		return
	}

	count, err := database.FetchCount(ctx, dbConn, rpcRequest.Address(), meta.LastIndexedBlock)
	if err != nil {
		log.Println("database query (count):", err)
		err = ErrInternal
		return
	}

	response = &query.RpcResponse[int]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result: query.Result[int]{
			Data: count,
			Meta: meta,
		},
	}
	return
}
