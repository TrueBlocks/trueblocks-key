package main

import (
	"flag"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/TrueBlocks/trueblocks-key/test/simulate_session/pkg/config"
	"github.com/TrueBlocks/trueblocks-key/test/simulate_session/pkg/runner"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "path", "", "path to testing scenario config file")
	flag.Parse()

	if configPath == "" {
		log.Fatalln("--path is required")
	}
}

func main() {
	cnf, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	if cnf.BaseUrl == "" {
		log.Fatalln("baseUrl is missing in the configuration file")
	}
	if len(cnf.Scenarios) == 0 {
		log.Fatalln("at least 1 testing scenario required")
	}
	// if cnf.Rate == 0 {
	// 	log.Fatalln("rate required")
	// }
	if cnf.Duration == 0 {
		log.Fatalln("duration required")
	}

	results := make(chan runner.Result, 500)
	var statsMutex sync.Mutex
	errors := make(map[string]int, 0)
	var succeeded int
	durations := make([]int64, 0, 1000)
	go runner.Run(cnf, results)

	for r := range results {
		statsMutex.Lock()
		if r.Ok {
			succeeded++
		} else {
			errors[r.Error.String()]++
		}
		durations = append(durations, r.Duration.Microseconds())
		statsMutex.Unlock()
	}

	log.Println("Done")

	minD := slices.Min(durations)
	maxD := slices.Max(durations)

	fmt.Println("Success:", succeeded)
	fmt.Println("Errors:")
	for err, count := range errors {
		fmt.Printf("	%s: %d\n", err, count)
	}
	fmt.Printf("Duration: min = %dms, max = %dms\n", minD, maxD)
}
