package main

import "github.com/TrueBlocks/trueblocks-key/query/pkg/query"

func getValidLimits(p query.Limiter) (validLimit uint) {
	limit := p.Limit()
	if limit == 0 {
		// Just in case we forgot to define the limit in configuration
		validLimit = defaultAppearancesLimit
	}

	if confLimit := cnf.Query.MaxLimit; confLimit > 0 {
		if validLimit > confLimit {
			validLimit = confLimit
		}
	}

	if validLimit < query.MinSafePerPage {
		validLimit = query.MinSafePerPage
	}

	return
}
