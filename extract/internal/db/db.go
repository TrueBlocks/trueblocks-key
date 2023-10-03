package db

import (
	extractConfig "trueblocks.io/config/pkg"
	"trueblocks.io/uploader/pkg/db"
)

var connection *db.Connection

func Connection(configPath string) (*db.Connection, error) {
	if connection != nil {
		return connection, nil
	}

	config, err := extractConfig.Get(configPath)
	if err != nil {
		return nil, err
	}

	dbConnection := &db.Connection{
		Host:     config.Database.Host,
		Port:     config.Database.Port,
		User:     config.Database.User,
		Password: config.Database.Password,
		Database: config.Database.Database,
	}
	if err := dbConnection.Connect(); err != nil {
		return nil, err
	}

	connection = dbConnection
	return connection, nil
}
