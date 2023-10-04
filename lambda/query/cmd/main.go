package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	extractConfig "trueblocks.io/config/pkg"
	database "trueblocks.io/database/pkg"
)

const maxLimit = 1000

var configFilePath string
var dbConfigKey string
var offset int
var limit = 100

func init() {
	flag.StringVar(&configFilePath, "config", "", "configuration file path")
	flag.StringVar(&dbConfigKey, "database", "default", "database to use")
	flag.IntVar(&offset, "offset", 0, "offset")
	flag.IntVar(&limit, "limit", 100, "limit")
}

func main() {
	flag.Parse()

	if configFilePath == "" {
		log.Fatalln("configuration file path required")
	}

	if len(flag.Args()) != 1 {
		log.Fatalln("address required")
	}

	address := strings.ToLower(flag.Arg(0))

	if limit > maxLimit {
		limit = maxLimit
	}

	// TODO: we should have separate DB config and not use extract config here
	config, err := extractConfig.Get(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	dbConnection := &database.Connection{
		Host:     config.Database[dbConfigKey].Host,
		Port:     config.Database[dbConfigKey].Port,
		User:     config.Database[dbConfigKey].User,
		Password: config.Database[dbConfigKey].Password,
		Database: config.Database[dbConfigKey].Database,
	}
	if err := dbConnection.Connect(); err != nil {
		log.Fatalln(err)
	}

	results := make([]database.Appearance, 0, limit)
	dbtx := dbConnection.Db().Where("address like ?", address).Limit(limit).Offset(offset).Find(&results)
	if err = dbtx.Error; err != nil {
		log.Fatalln(err)
	}

	for _, appearance := range results {
		fmt.Println(appearance.Address, appearance.BlockNumber, appearance.TransactionId)
	}
}
