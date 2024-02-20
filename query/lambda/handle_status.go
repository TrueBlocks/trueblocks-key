package main

import (
	"context"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleLastIndexedBlock(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[database.Status], err error) {
	status, err := database.FetchStatus(ctx, dbConn)
	if err != nil {
		log.Println("fetching status:", err)
		err = ErrInternal
		return
	}

	response = &query.RpcResponse[database.Status]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result: query.Result[database.Status]{
			Data: status,
		},
	}
	return
}
