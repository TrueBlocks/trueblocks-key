package query

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/walk"
	"trueblocks.io/searcher/bloom"
	"trueblocks.io/searcher/chunk"
)

func Find(chain string, address base.Address) error {
	foundRangesCh := make(chan string, 100)
	results := make(chan index.AppearanceRecord, 100)

	go func() {
		if err := QueryBlooms(chain, address, foundRangesCh); err != nil {
			panic(fmt.Errorf("querying bloom filters: %w", err))
		}
	}()

	go func() {
		defer close(results)
		for fileRange := range foundRangesCh {
			apps, err := Extract(chain, fileRange, address)
			if err != nil {
				panic(err)
			}
			for _, app := range apps {
				results <- app
			}
		}
	}()

	for app := range results {
		log.Println(app.BlockNumber, app.TransactionId)
	}

	return nil
}

func QueryBlooms(chain string, address base.Address, foundRangesCh chan string) error {
	bloomPath := config.GetPathToIndex(chain) + "blooms/"
	files, err := os.ReadDir(bloomPath)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, info := range files {
		if info.IsDir() {
			continue
		}

		rawFileName := info.Name()
		fileName := bloomPath + "/" + rawFileName
		if !walk.IsCacheType(fileName, walk.Index_Bloom, true /* checkExt */) {
			continue // sometimes there are .gz files in this folder, for example
		}
		fileRange, err := base.RangeFromFilenameE(fileName)
		if err != nil {
			// don't respond further -- there may be foreign files in the folder
			fmt.Println(err)
			continue
		}

		// Run a go routine for each index file
		wg.Add(1)
		go func() {
			f, err := os.Open(fileName)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			defer wg.Done()
			b, err := bloom.NewBloom(f, fileRange.String())
			if err != nil {
				panic(err)
			}

			v, err := b.IsMember(address)
			if err != nil {
				panic(err)
			}

			if v {
				foundRangesCh <- rawFileName
				log.Println("Bloom match:", rawFileName)
			}
		}()

	}

	wg.Wait()
	close(foundRangesCh)

	return nil
}

func Extract(chain string, fileName string, address base.Address) (result []index.AppearanceRecord, err error) {
	indexFilename := config.GetPathToIndex(chain) + "finalized/" + index.ToIndexPath(fileName)
	f, err := os.Open(indexFilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	chunk, err := chunk.NewChunkData(f, fileName)
	if err != nil {
		return
	}
	found := chunk.GetAppearanceRecords(address)
	if found == nil {
		return
	}
	if found.AppRecords == nil {
		return
	}
	if found.Err != nil {
		return nil, found.Err
	}
	result = *found.AppRecords
	return
}
