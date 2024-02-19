package main

import "github.com/TrueBlocks/trueblocks-key/query/pkg/query"

func getValidLimits(p query.Paginator) (validLimit uint, validOffset uint) {
	offset, limit := p.Pagination()
	if limit == 0 {
		// Just in case we forgot to define the limit in configuration
		validLimit = defaultAppearancesLimit
	}

	if confLimit := cnf.Query.MaxLimit; confLimit > 0 {
		if validLimit > confLimit {
			validLimit = confLimit
		}
	}

	if offset < 0 {
		validOffset = 0
	}
	validOffset = validOffset * validLimit

	return
}
