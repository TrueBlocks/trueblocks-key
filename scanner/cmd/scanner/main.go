package main

import (
	"log"
	"os"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"trueblocks.io/searcher/pkg/query"
)

var runEnv *CmdRunEnv

func init() {
	runEnv = &CmdRunEnv{}
}

func main() {
	chain := "mainnet"
	address := os.Args[1]
	if address == "" {
		log.Fatalln("Address required")
	}

	results := make(chan index.AppearanceRecord, 100)
	go func() {
		if err := query.Find(chain, base.HexToAddress(address), runEnv, results); err != nil {
			log.Fatalln("find error:", err)
		}
	}()

	for app := range results {
		log.Println(app.BlockNumber, app.TransactionId)
	}
}
