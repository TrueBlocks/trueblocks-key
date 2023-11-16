package db

import (
	extractConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

var connection *database.Connection

func Connection(configPath string, dbConfigKey string) (*database.Connection, error) {
	if connection != nil {
		return connection, nil
	}

	config, err := extractConfig.Get(configPath)
	if err != nil {
		return nil, err
	}

	dbConnection := &database.Connection{
		Host:     config.Database[dbConfigKey].Host,
		Port:     config.Database[dbConfigKey].Port,
		User:     config.Database[dbConfigKey].User,
		Password: config.Database[dbConfigKey].Password,
		Database: config.Database[dbConfigKey].Database,
	}
	if err := dbConnection.Connect(); err != nil {
		return nil, err
	}

	connection = dbConnection
	return connection, nil
}
