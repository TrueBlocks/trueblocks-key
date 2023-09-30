package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var outputDir = "/tmp/uploader"

func main() {
	indexPath := os.Args[1]
	if indexPath == "" {
		log.Fatalln("index (finalized) path required")
	}

	db, err := gorm.Open(sqlite.Open("appearances.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Appearance{})

	log.Println("Reading Unchained Index from", indexPath)

	// Collect chunk file names
	flist := make(map[string]string)

	err = filepath.WalkDir(indexPath, func(path string, d fs.DirEntry, derr error) error {
		if derr != nil {
			return derr
		}

		if d.IsDir() {
			return nil
		}

		fmt.Printf("\rAdding: %s		", path)

		flist[path] = path
		return nil
	})
	fmt.Println()

	if err != nil {
		log.Fatalln(err)
	}

	// Extract appearances in batches and push them to DB

	ctx := context.Background()
	errs := make(chan error)
	step := func(key string, value string) error {
		fmt.Println("Working:", key)
		path := key
		chunk, err := index.NewChunkData(path)
		if err != nil {
			return err
		}

		_, err = chunk.File.Seek(int64(index.HeaderWidth), io.SeekStart)
		if err != nil {
			return err
		}

		fileRange, err := base.RangeFromFilenameE(path)
		if err != nil {
			return err
		}

		for i := 0; i < int(chunk.Header.AddressCount); i++ {
			addressRecord := index.AddressRecord{}
			if err := addressRecord.ReadAddress(chunk.File); err != nil {
				return err
			}
			apps, err := chunk.ReadAppearanceRecordsAndResetOffset(&addressRecord)
			if err != nil {
				return err
			}

			for _, app := range apps {
				log.Println("Found", addressRecord.Address.Hex(), app.BlockNumber, app.TransactionId)
				// TODO: use batch create instead
				dbtx := db.Create(&Appearance{
					Address:         addressRecord.Address.Hex(),
					BlockNumber:     app.BlockNumber,
					TransactionId:   app.TransactionId,
					BlockRangeStart: fileRange.First,
					BlockRangeEnd:   fileRange.Last,
				})
				if err = dbtx.Error; err != nil {
					return err
				}
			}
		}
		chunk.Close()
		return nil
	}
	utils.IterateOverMap[string](ctx, errs, flist, step)

	if err := <-errs; err != nil {
		log.Fatalln(err)
	}

	log.Println("Done")
}
