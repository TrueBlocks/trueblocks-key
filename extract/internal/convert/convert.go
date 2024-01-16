package convert

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
)

var realTimeProgress bool

func init() {
	if os.Getenv("KY_REALTIME_PROGRESS") != "false" {
		realTimeProgress = true
	}
}

type AppearanceReceiver interface {
	SendBatch([]queueItem.Appearance) error
}

func Convert(cnf *config.ConfigFile, receiver AppearanceReceiver, indexPath string) error {
	log.Println("Reading Unchained Index from", indexPath)
	log.Println("Saving status info to", statusFile.Name())
	defer CloseStatusFile()

	// Collect chunk file names
	chunkPaths := make(map[string]string)
	lastFile := ""

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
	batchSize := cnf.Convert.BatchSize
	batch := make([]queueItem.Appearance, 0, batchSize)
	var mu sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error)
	step := func(key string, value string) error {
		chunkPath := key
		fileName := path.Base(chunkPath)
		chunk, err := index.NewChunkData(chunkPath)
		if err != nil {
			SaveStatus(fileName, StatusError)
			return err
		}

		_, err = chunk.File.Seek(int64(index.HeaderWidth), io.SeekStart)
		if err != nil {
			SaveStatus(fileName, StatusError)
			return err
		}

		fileRange, err := base.RangeFromFilenameE(chunkPath)
		if err != nil {
			SaveStatus(fileName, StatusError)
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
				SaveStatus(fileName, StatusError)
				return err
			}
			apps, err := chunk.ReadAppearanceRecordsAndResetOffset(&addressRecord)
			if err != nil {
				SaveStatus(fileName, StatusError)
				return err
			}

			for _, app := range apps {
				mu.Lock()
				batch = append(batch, queueItem.Appearance{
					Address:         addressRecord.Address.Hex(),
					BlockNumber:     app.BlockNumber,
					TransactionId:   app.TransactionId,
					BlockRangeStart: fileRange.First,
					BlockRangeEnd:   fileRange.Last,
				})
				if len(batch) >= batchSize {
					if err := receiver.SendBatch(batch); err != nil {
						SaveStatus(fileName, StatusAppError)
						return err
					}
					batch = make([]queueItem.Appearance, 0, batchSize)
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
		if !realTimeProgress {
			fmt.Println("done chunk", fileRange.String(), "appearances total:", appsFound.Load())
		}
		SaveStatus(fileName, StatusDone)
		chunk.Close()

		// Empty buffer
		if chunkPath == lastFile && len(batch) > 0 {
			if err := receiver.SendBatch(batch); err != nil {
				SaveStatus(fileName, StatusAppError)
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
			if realTimeProgress {
				fmt.Printf("\rchunk %d/%d, range %s, total appearances found %d		", p, totalChunks, message.Range, f)
			}
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
