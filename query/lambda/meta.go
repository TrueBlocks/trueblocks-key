package main

import (
	"context"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

func getMeta(ctx context.Context, address string) (m *query.Meta, err error) {
	status, err := database.FetchStatus(ctx, dbConn)
	if err != nil {
		log.Println("database status query:", err)
		err = ErrInternal
		return
	}

	if !status.HasLastIndexedBlock() {
		log.Println("last indexed block is 0, returning ErrInternal")
		err = ErrInternal
		return
	}

	m = &query.Meta{
		Address: address,
	}
	m.SetLastIndexedBlock(status.LastIndexedBlock)
	return
}
