package main

import (
	"log"
	"os"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"trueblocks.io/searcher/pkg/query"
)

func main() {
	chain := "mainnet"
	address := os.Args[1]
	if address == "" {
		log.Fatalln("Address required")
	}

	if err := query.Find(chain, base.HexToAddress(address)); err != nil {
		log.Fatalln("find error:", err)
	}
}
