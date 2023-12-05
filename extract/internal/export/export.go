package export

import (
	"context"
	"errors"
	"fmt"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

func ExportAddresses(dbConn *database.Connection, destPath string) error {
	return export(dbConn, destPath, dbConn.AddressesTableName())
}

func ExportAppearances(dbConn *database.Connection, destPath string) error {
	return export(dbConn, destPath, dbConn.AppearancesTableName())
}

func export(dbConn *database.Connection, destPath string, tableName string) error {
	if destPath == "" {
		return errors.New("export: destination path required")
	}

	log.Println("Exporting", tableName, "table, destination:", destPath)

	_, err := dbConn.Db().Exec(
		context.TODO(),
		"COPY (SELECT * FROM $1) TO PROGRAM '$2' (FORMAT 'csv')",
		tableName, "split -b 1G -d - "+destPath,
	)
	if err != nil {
		return fmt.Errorf("export: copy SQL: %w", err)
	}

	return nil
}
