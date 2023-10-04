package export

import (
	"errors"
	"fmt"
	"log"

	database "trueblocks.io/database/pkg"
)

func Export(dbConn *database.Connection, destPath string) error {
	if destPath == "" {
		return errors.New("export: destination path required")
	}

	tableName, err := database.TableName[database.Appearance](dbConn)
	if err != nil {
		return fmt.Errorf("export: %w", err)
	}

	log.Println("Exporting", tableName, "table, destination:", destPath)

	dbtx := dbConn.Db().Exec(
		fmt.Sprintf("COPY (SELECT * FROM %s) TO PROGRAM '%s' (FORMAT 'csv')", tableName, "split -b 1G -d - "+destPath),
	)
	if err := dbtx.Error; err != nil {
		return fmt.Errorf("export: copy SQL: %w", err)
	}

	return nil
}
