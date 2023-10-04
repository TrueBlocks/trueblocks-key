package database

import (
	"fmt"

	"gorm.io/gorm"
)

func TableName[Model any](dbConn *Connection) (string, error) {
	db := dbConn.Db()
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(Model)); err != nil {
		return "", fmt.Errorf("get model table name: %w", err)
	}
	return stmt.Schema.Table, nil
}
