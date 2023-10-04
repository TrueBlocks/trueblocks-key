package swap

import (
	"fmt"
	"log"

	"gorm.io/gorm"
	database "trueblocks.io/database/pkg"
)

func Swap(dbConn *database.Connection) error {
	db := dbConn.Db()

	liveTableName, err := database.TableName[database.Appearance](dbConn)
	if err != nil {
		return fmt.Errorf("swap: %w", err)
	}

	tempName := liveTableName + "_temp"
	stagingName := liveTableName + "_staging"

	log.Println("Swapping live table", liveTableName, "and", stagingName)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().RenameTable(liveTableName, tempName); err != nil {
			return err
		}
		if err := tx.Migrator().RenameTable(stagingName, liveTableName); err != nil {
			return err
		}
		return nil
	})
}
