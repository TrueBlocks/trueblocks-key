package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
)

var ErrZero = errors.New("zero value is invalid")

func main() {
	var flagHelp bool
	flag.BoolVar(&flagHelp, "help", false, "displays help message")

	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	var directionNextPage bool
	var lastBlock uint32
	var blockNumber uint32
	var transactionIndex uint32
	buildCmd.BoolVar(&directionNextPage, "direction-next", false, "set if pageId is used to ")
	buildCmd.Func("last-block", "the latest block included in the dataset", parseUint32Flag(&lastBlock))
	buildCmd.Func("block-number", "current page appearance block number", parseUint32Flag(&blockNumber))
	buildCmd.Func("tx", "current page appearance transaction index", parseUint32Flag(&transactionIndex))

	inspectCmd := flag.NewFlagSet("inspect", flag.ExitOnError)

	flag.Parse()
	printHelp := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: mode [build mode flags] [global flags]\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Valid mode: inspect, build\n")
		fmt.Fprintf(flag.CommandLine.Output(), "build mode flags:\n")
		buildCmd.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "global flags:\n")
		flag.PrintDefaults()
	}

	if flagHelp {
		printHelp()
		return
	}

	if len(os.Args) < 2 {
		log.Println("mode is required: inspect, build")
		printHelp()
		os.Exit(1)
	}

	mode := os.Args[1]
	restArgs := os.Args[2:]
	switch mode {
	case "inspect":
		inspectCmd.Parse(restArgs)
		pageId := flag.Arg(1)
		if pageId == "" {
			log.Fatalln("pageId required")
		}
		inspect(pageId)
	case "build":
		buildCmd.Parse(restArgs)
		build(directionNextPage, lastBlock, blockNumber, transactionIndex)
	default:
		log.Println("valid modes are: inspect, build")
		printHelp()
		os.Exit(1)
	}
}

func parseUint32Flag(target *uint32) func(string) error {
	return func(s string) error {
		value, err := strconv.ParseUint(s, 0, 32)
		if err != nil {
			return err
		}
		if value == 0 {
			return ErrZero
		}
		*target = uint32(value)
		return nil
	}
}

func inspect(pageIdSrc string) {
	p := query.PageId{}
	if err := p.UnmarshalJSON([]byte(strconv.Quote(pageIdSrc))); err != nil {
		log.Fatalln("decoding pageId:", err)
	}

	// we use MarshalIndent to pretty-print PageId
	indented, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		log.Fatalln("pretty-print:", err)
	}
	fmt.Println(string(indented))
}

func build(directionNext bool, lastBlock uint32, blockNumber uint32, transactionIndex uint32) {
	if lastBlock == 0 ||
		blockNumber == 0 ||
		transactionIndex == 0 {
		log.Fatalln(ErrZero)
	}

	p := query.PageId{
		DirectionNextPage: directionNext,
		LastBlock:         lastBlock,
		BlockNumber:       blockNumber,
		TransactionIndex:  transactionIndex,
	}
	encoded, err := p.MarshalJSON()
	if err != nil {
		log.Fatalln("encoding:", err)
	}
	fmt.Println(string(encoded))
}
