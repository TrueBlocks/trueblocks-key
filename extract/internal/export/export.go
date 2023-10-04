package export

import (
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
	dbPkg "trueblocks.io/uploader/pkg/db"
)

func Export(dbConn *dbPkg.Connection, destPath string) error {
	if destPath == "" {
		return errors.New("export: destination path required")
	}

	db := dbConn.Db()
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&dbPkg.Appearance{}); err != nil {
		return fmt.Errorf("export: get model table name: %w", err)
	}

	log.Println("Exporting", stmt.Schema.Table, "table, destination:", destPath)

	dbtx := dbConn.Db().Exec(
		fmt.Sprintf("COPY (SELECT * FROM %s) TO PROGRAM '%s' (FORMAT 'csv')", stmt.Schema.Table, "split -b 1G -d - "+destPath),
	)
	if err := dbtx.Error; err != nil {
		return fmt.Errorf("export: copy SQL: %w", err)
	}

	return nil
}
