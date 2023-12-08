package export

import (
	"context"
	"errors"
	"fmt"
	"log"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/jackc/pgx/v5"
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
		fmt.Sprintf(
			"COPY (SELECT * FROM %s) TO PROGRAM %s (FORMAT 'csv')",
			pgx.Identifier.Sanitize(pgx.Identifier{tableName}),
			pgx.Identifier.Sanitize(pgx.Identifier{"split -b 1G -d - " + destPath}),
		),
	)
	if err != nil {
		return fmt.Errorf("export: copy SQL: %w", err)
	}

	return nil
}
