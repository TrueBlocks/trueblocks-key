package main

import (
	"context"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleLastIndexedBlock(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcLastIndexedBlockResponse, err error) {
	block, err := database.FetchMaxBlockNumber(ctx, dbConn)
	if err != nil {
		log.Println("database query (lastIndexedBlock):", err)
		err = ErrInternal
		return
	}

	response = &query.RpcLastIndexedBlockResponse{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result:  query.Result[int]{Data: block},
	}
	return
}
