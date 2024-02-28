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
	if limit < query.MinSafePerPage {
		limit = query.MinSafePerPage
	}

	if confLimit := cnf.Query.MaxLimit; confLimit > 0 {
		if limit > int(confLimit) {
			limit = int(confLimit)
		}
	}

	// get status first, so we know max block number
	meta, err := getMeta(ctx, rpcRequest.Address())
	if err != nil {
		return
	}

	lastBlock, err := rpcRequest.LastBlockNumber()
	if err != nil {
		log.Println("last block error:", err)
		return
	}
	if lastBlock == nil {
		// nil means "latest"
		lastBlock = &meta.LastIndexedBlock
	}

	var items []database.Appearance
	var firstPage bool
	pageId := rpcRequest.Parameters().PageId

	if pageId == nil {
		log.Println("fetching first page")
		items, err = database.FetchAppearancesFirstPage(ctx, dbConn, rpcRequest.Address(), *lastBlock, uint(limit))
		firstPage = true
	} else {
		// pageId.LastBlock takes precedence before query's lastBlock (it shouldn't be there if the user sends pageId)
		bn := uint(pageId.LastBlock)
		lastBlock = &bn

		log.Println("fetching page -- next?", pageId.DirectionNextPage, "blockNum:", pageId.BlockNumber, "tx_id:", pageId.TransactionIndex)
		items, err = database.FetchAppearancesPage(ctx, dbConn, pageId.DirectionNextPage, rpcRequest.Address(), *lastBlock, uint(limit), uint(pageId.BlockNumber), uint(pageId.TransactionIndex))
	}

	if err != nil {
		log.Println("database query:", err)
		err = ErrInternal
		return
	}

	previousPageId, nextPageId := getPageIds(items, *lastBlock)
	if !firstPage {
		meta.PreviousPageId = previousPageId
	}
	meta.NextPageId = nextPageId

	if nextPageId != nil {
		log.Println("next page id: next?", nextPageId.DirectionNextPage, "blockNum:", nextPageId.BlockNumber, "tx_id:", nextPageId.TransactionIndex)
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

func getPageIds(items []database.Appearance, lastBlock uint) (previousPageId *query.PageId, nextPageId *query.PageId) {
	if len(items) == 0 {
		return
	}
	previousPageId = &query.PageId{
		DirectionNextPage: false,
		LastBlock:         uint32(lastBlock),
		BlockNumber:       items[0].BlockNumber,
		TransactionIndex:  items[0].TransactionIndex,
	}

	lastCurrentAppearance := items[len(items)-1]
	nextPageId = &query.PageId{
		DirectionNextPage: true,
		LastBlock:         uint32(lastBlock),
		BlockNumber:       lastCurrentAppearance.BlockNumber,
		TransactionIndex:  lastCurrentAppearance.TransactionIndex,
	}
	return
}
