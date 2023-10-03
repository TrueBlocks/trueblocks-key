package convert

import (
	"context"
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
	dbPkg "trueblocks.io/uploader/pkg/db"
)

func Convert(dbConn *dbPkg.Connection, indexPath string) error {
	log.Println("Reading Unchained Index from", indexPath)

	// Collect chunk file names
	chunkPaths := make(map[string]string)
	lastFile := ""

	if err := dbConn.AutoMigrate(); err != nil {
		return err
	}
	db := dbConn.Db()

	err := filepath.WalkDir(indexPath, func(path string, d fs.DirEntry, derr error) error {
		if derr != nil {
			return derr
		}

		if d.IsDir() {
			return nil
		}

		// It has to be a map so that we can use IterateOverMap
		chunkPaths[path] = path
		lastFile = path
		return nil
	})
	if err != nil {
		return err
	}

	log.Println("Will convert", len(chunkPaths), "chunks")

	progressChan := make(chan struct {
		Chunk, App int
		Range      string
	}, 10)
	totalChunks := len(chunkPaths)
	chunkProgress := atomic.Int32{}
	appsFound := atomic.Int32{}

	// Extract appearances in batches and push them to DB

	batch := make([]*dbPkg.Appearance, 0, dbConn.BatchSize())
	var mu sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
				batch = append(batch, &dbPkg.Appearance{
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
					batch = make([]*dbPkg.Appearance, 0, 5000)
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
	go utils.IterateOverMap[string](ctx, errs, chunkPaths, step)

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
		return err
	}
	close(progressChan)

	wg.Wait()
	log.Println("Wrote", totalChunks, "chunks")

	return nil
}
