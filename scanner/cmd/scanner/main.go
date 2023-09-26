package main

import (
	"log"
	"os"

	"trueblocks.io/searcher/pkg/query"
	"trueblocks.io/searcher/pkg/query/chunk"
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
	indexPath := os.Args[2]
	if indexPath == "" {
		log.Fatalln("indexPath required")
	}
	runEnv.IndexPath = indexPath

	results := make(chan chunk.AppearanceRecord, 100)
	go func() {
		if err := query.Find(chain, address, runEnv, results); err != nil {
			log.Fatalln("find error:", err)
		}
	}()

	for app := range results {
		log.Println(app.BlockNumber, app.TransactionId)
	}
}
