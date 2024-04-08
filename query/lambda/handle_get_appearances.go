package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

const defaultAppearancesLimit = 100

func handleGetAppearances(ctx context.Context, rpcRequest *query.RpcRequest) (response *query.RpcResponse[[]database.PublicAppearance], err error) {
	rpcParams, err := rpcRequest.AppearancesParams()
	if err != nil {
		err = NewRpcError(err, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err = rpcParams.Validate(); err != nil {
		err = NewRpcError(err, http.StatusBadRequest, err.Error())
		// Validate() always returns public errors
		return
	}

	param := rpcParams.Get()
	if err = param.Validate(); err != nil {
		err = NewRpcError(err, http.StatusBadRequest, err.Error())
		return
	}

	limit := getValidLimits(param)

	// get status first, so we know max block number
	meta, err := getMeta(ctx, param.Address)
	if err != nil {
		return
	}

	lastBlock, err := param.LastBlockNumber()
	if err != nil {
		log.Println("last block error:", err)
		return
	}
	if lastBlock == nil {
		// nil means "latest"
		lastBlock = &meta.LastIndexedBlock
	}

	var items []database.Appearance
	var fetchBounds bool
	specialPageId, pageId, err := param.PageIdValue()
	if err != nil {
		log.Println("reading page id value:", err)
		err = errors.New("invalid pageId")
		return
	}

	switch specialPageId {
	case query.PageIdLatest, query.PageIdEarliest:
		log.Println("fetching first page")
		items, err = database.FetchAppearancesFirstPage(
			ctx,
			dbConn,
			specialPageId == query.PageIdEarliest,
			param.Address,
			*lastBlock,
			uint(limit),
		)
		fetchBounds = true
	default:
		// pageId.LastBlock takes precedence before query's lastBlock (it shouldn't be there if the user sends pageId)
		bn := uint(pageId.LastBlock)
		lastBlock = &bn

		log.Println("fetching page -- next?", pageId.DirectionNextPage, "last seen:", fmt.Sprint(pageId.LastSeen), "latest in set:", fmt.Sprint(pageId.LatestInSet), "earliest in set:", fmt.Sprint(pageId.EarliestInSet))
		items, err = database.FetchAppearancesPage(ctx, dbConn, pageId.DirectionNextPage, param.Address, *lastBlock, uint(limit), uint(pageId.LastSeen.BlockNumber), uint(pageId.LastSeen.TransactionIndex))
	}

	if err != nil {
		log.Println("database query:", err)
		err = ErrInternal
		return
	}

	hasItems := len(items) > 0
	var bounds database.AppearancesDatasetBounds
	if fetchBounds {
		if hasItems {
			bounds, err = database.FetchAppearancesDatasetBounds(ctx, dbConn, param.Address, *lastBlock)
			if err != nil {
				log.Println("error while getting bounds:", err)
				err = ErrInternal
				return
			}
		}
	} else {
		bounds = database.AppearancesDatasetBounds{
			Latest:   pageId.LatestInSet,
			Earliest: pageId.EarliestInSet,
		}
	}

	if hasItems {
		previousPageId, nextPageId := getPageIds(items, *lastBlock, &bounds)
		meta.PreviousPageId = previousPageId
		meta.NextPageId = nextPageId

		if nextPageId != nil {
			log.Println("next page id: next?", nextPageId.DirectionNextPage, "last seen:", nextPageId.LastSeen, "latest in set:", nextPageId.LatestInSet, "earliest in set:", nextPageId.EarliestInSet)
		}
	}

	publicApps := database.AppearanceSliceToPublicSlice(items)

	response = &query.RpcResponse[[]database.PublicAppearance]{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result: query.Result[[]database.PublicAppearance]{
			Data: publicApps,
			Meta: meta,
		},
	}
	return
}

func getPageIds(items []database.Appearance, lastBlock uint, bounds *database.AppearancesDatasetBounds) (previousPageId *query.PageId, nextPageId *query.PageId) {
	if len(items) == 0 {
		return
	}

	if !bounds.IsLatest(&items[0]) {
		nextPageId = &query.PageId{
			DirectionNextPage: false,
			LastBlock:         uint32(lastBlock),
			LastSeen:          items[0],
			LatestInSet:       bounds.Latest,
			EarliestInSet:     bounds.Earliest,
		}
	}

	lastCurrentAppearance := items[len(items)-1]

	if !bounds.IsEarliest(&lastCurrentAppearance) {
		previousPageId = &query.PageId{
			DirectionNextPage: true,
			LastBlock:         uint32(lastBlock),
			LastSeen:          lastCurrentAppearance,
			LatestInSet:       bounds.Latest,
			EarliestInSet:     bounds.Earliest,
		}
	}
	return
}
