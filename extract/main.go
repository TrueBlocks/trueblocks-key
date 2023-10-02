package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var outputDir = "/tmp/uploader"

func main() {
	startBlock := -1
	endBlock := -1
	flag.IntVar(&startBlock, "start", -1, "start block")
	flag.IntVar(&endBlock, "end", -1, "end block")
	flag.Parse()

	indexPath := flag.Arg(0)
	if indexPath == "" {
		log.Fatalln("index (finalized) path required")
	}

	dsn := "host=localhost user=postgres password=example dbname=index port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		CreateBatchSize: 5000,
		Logger:          logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Appearance{}, &Progress{})

	log.Println("Reading Unchained Index from", indexPath)

	// Collect chunk file names
	flist := make(map[string]string)
	lastFile := ""

	err = filepath.WalkDir(indexPath, func(path string, d fs.DirEntry, derr error) error {
		if derr != nil {
			return derr
		}

		if d.IsDir() {
			return nil
		}

		r := base.RangeFromFilename(path)

		if startBlock > 0 && r.EarlierThanB(uint64(startBlock)) {
			// WalkDir uses lexical order, so we can skip all
			return filepath.SkipAll
		}

		if endBlock > 0 && r.LaterThanB(uint64(endBlock)) {
			return filepath.SkipAll
		}

		fmt.Printf("\rAdding: %s		", path)

		flist[path] = path
		lastFile = path
		return nil
	})
	fmt.Println()

	if err != nil {
		log.Fatalln(err)
	}

	progressChan := make(chan struct {
		Chunk, App int
		Range      string
	}, 10)
	totalChunks := len(flist)
	chunkProgress := atomic.Int32{}
	appsFound := atomic.Int32{}

	// Extract appearances in batches and push them to DB

	batch := make([]*Appearance, 0, 5000)
	var mu sync.Mutex

	ctx := context.Background()
	errs := make(chan error)
	step := func(key string, value string) error {
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

		progressChan <- struct {
			Chunk int
			App   int
			Range string
		}{
			Chunk: 1,
			Range: fileRange.String(),
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
				mu.Lock()
				batch = append(batch, &Appearance{
					Address:         addressRecord.Address.Hex(),
					BlockNumber:     app.BlockNumber,
					TransactionId:   app.TransactionId,
					BlockRangeStart: fileRange.First,
					BlockRangeEnd:   fileRange.Last,
				})
				if len(batch) >= 5000 {
					dbtx := db.Create(batch)
					if err = dbtx.Error; err != nil {
						return err
					}
					batch = make([]*Appearance, 0, 5000)
				}
				mu.Unlock()
				progressChan <- struct {
					Chunk int
					App   int
					Range string
				}{
					App:   1,
					Range: fileRange.String(),
				}
			}
		}
		chunk.Close()

		// Empty buffer
		if path == lastFile && len(batch) > 0 {
			dbtx := db.Create(batch)
			if err = dbtx.Error; err != nil {
				return err
			}
		}

		return nil
	}
	go utils.IterateOverMap[string](ctx, errs, flist, step)

	// Progress reporting
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var p int32
		var f int32
		for message := range progressChan {
			if message.Chunk == 1 {
				p = chunkProgress.Add(1)
			}
			if message.App == 1 {
				f = appsFound.Add(1)
			}
			fmt.Printf("\rchunk %d/%d, range %s, total appearances found %d		", p, totalChunks, message.Range, f)
		}
		fmt.Println()
		wg.Done()
	}()

	if err := <-errs; err != nil {
		log.Fatalln(err)
	}
	close(progressChan)

	wg.Wait()
	log.Println("Wrote", totalChunks, "chunks")
}
