package main

import (
	"context"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func handleLastIndexedBlock(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[*database.Status], err error) {
	meta, err := getMeta(ctx, "")
	if err != nil {
		return
	}

	response = &query.RpcResponse[*database.Status]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result: query.Result[*database.Status]{
			Data: nil,
			Meta: meta,
		},
	}
	return
}
