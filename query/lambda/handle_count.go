package main

import (
	"context"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleCount(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[int], err error) {
	count, err := database.FetchCount(ctx, dbConn, rpcRequest.Address())
	if err != nil {
		log.Println("database query (count):", err)
		err = ErrInternal
		return
	}

	response = &query.RpcResponse[int]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result:  query.Result[int]{Data: count},
	}
	return
}
