package main

import (
	"context"
	"fmt"
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

		log.Println("fetching page -- next?", pageId.DirectionNextPage, "last seen:", fmt.Sprint(pageId.LastSeen), "latest in set:", fmt.Sprint(pageId.LatestInSet), "earliest in set:", fmt.Sprint(pageId.EarliestInSet))
		items, err = database.FetchAppearancesPage(ctx, dbConn, pageId.DirectionNextPage, rpcRequest.Address(), *lastBlock, uint(limit), uint(pageId.LastSeen.BlockNumber), uint(pageId.LastSeen.TransactionIndex))
	}

	if err != nil {
		log.Println("database query:", err)
		err = ErrInternal
		return
	}

	hasItems := len(items) > 0
	var boundaries database.AppearancesDatasetBoundaries
	if firstPage {
		if hasItems {
			boundaries, err = database.FetchAppearancesDatasetBoundaries(ctx, dbConn, rpcRequest.Address(), *lastBlock)
			if err != nil {
				log.Println("error while getting boundaries:", err)
				err = ErrInternal
				return
			}
		}
	} else {
		boundaries = database.AppearancesDatasetBoundaries{
			Latest:   pageId.LatestInSet,
			Earliest: pageId.EarliestInSet,
		}
	}

	if hasItems {
		previousPageId, nextPageId := getPageIds(items, *lastBlock, &boundaries)
		if !firstPage {
			meta.PreviousPageId = previousPageId
		}
		meta.NextPageId = nextPageId

		if nextPageId != nil {
			log.Println("next page id: next?", nextPageId.DirectionNextPage, "last seen:", nextPageId.LastSeen, "latest in set:", nextPageId.LatestInSet, "earliest in set:", nextPageId.EarliestInSet)
		}
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

func getPageIds(items []database.Appearance, lastBlock uint, boundaries *database.AppearancesDatasetBoundaries) (previousPageId *query.PageId, nextPageId *query.PageId) {
	if len(items) == 0 {
		return
	}

	if !boundaries.IsLatest(&items[0]) {
		previousPageId = &query.PageId{
			DirectionNextPage: false,
			LastBlock:         uint32(lastBlock),
			LastSeen:          items[0],
			LatestInSet:       boundaries.Latest,
			EarliestInSet:     boundaries.Earliest,
		}
	}

	lastCurrentAppearance := items[len(items)-1]

	if !boundaries.IsEarliest(&lastCurrentAppearance) {
		nextPageId = &query.PageId{
			DirectionNextPage: true,
			LastBlock:         uint32(lastBlock),
			LastSeen:          lastCurrentAppearance,
			LatestInSet:       boundaries.Latest,
			EarliestInSet:     boundaries.Earliest,
		}
	}
	return
}
